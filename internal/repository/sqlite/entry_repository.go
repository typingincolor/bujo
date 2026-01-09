package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type EntryRepository struct {
	db *sql.DB
}

func NewEntryRepository(db *sql.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

func (r *EntryRepository) Insert(ctx context.Context, entry domain.Entry) (int64, error) {
	var scheduledDate *string
	if entry.ScheduledDate != nil {
		s := entry.ScheduledDate.Format("2006-01-02")
		scheduledDate = &s
	}

	entityID := entry.EntityID
	if entityID.IsEmpty() {
		entityID = domain.NewEntityID()
	}
	now := time.Now().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO entries (type, content, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.Type, entry.Content, entry.ParentID, entry.Depth, entry.Location, scheduledDate, entry.CreatedAt.Format(time.RFC3339),
		entityID.String(), 1, now, domain.OpTypeInsert.String())

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *EntryRepository) GetByID(ctx context.Context, id int64) (*domain.Entry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM entries WHERE id = ? AND (valid_to IS NULL OR valid_to = '')
	`, id)

	return r.scanEntry(row)
}

func (r *EntryRepository) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.Entry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM entries WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
	`, entityID.String())

	return r.scanEntry(row)
}

func (r *EntryRepository) GetByDate(ctx context.Context, date time.Time) ([]domain.Entry, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM entries WHERE scheduled_date = ? AND (valid_to IS NULL OR valid_to = '')
		ORDER BY created_at
	`, dateStr)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM entries WHERE scheduled_date >= ? AND scheduled_date <= ? AND (valid_to IS NULL OR valid_to = '')
		ORDER BY scheduled_date, created_at
	`, fromStr, toStr)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) GetOverdue(ctx context.Context, date time.Time) ([]domain.Entry, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		WITH RECURSIVE
		overdue_tasks AS (
			SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
			FROM entries
			WHERE scheduled_date < ? AND type = 'task' AND (valid_to IS NULL OR valid_to = '')
		),
		parent_chain AS (
			SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
			FROM overdue_tasks
			UNION
			SELECT e.id, e.type, e.content, e.parent_id, e.depth, e.location, e.scheduled_date, e.created_at, e.entity_id
			FROM entries e
			INNER JOIN parent_chain pc ON e.id = pc.parent_id
			WHERE (e.valid_to IS NULL OR e.valid_to = '')
		)
		SELECT DISTINCT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM parent_chain
		ORDER BY scheduled_date, depth, created_at
	`, dateStr)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) GetWithChildren(ctx context.Context, id int64) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH RECURSIVE tree AS (
			SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
			FROM entries WHERE id = ? AND (valid_to IS NULL OR valid_to = '')
			UNION ALL
			SELECT e.id, e.type, e.content, e.parent_id, e.depth, e.location, e.scheduled_date, e.created_at, e.entity_id
			FROM entries e
			JOIN tree t ON e.parent_id = t.id
			WHERE (e.valid_to IS NULL OR e.valid_to = '')
		)
		SELECT * FROM tree ORDER BY depth, id
	`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) Update(ctx context.Context, entry domain.Entry) error {
	var scheduledDate *string
	if entry.ScheduledDate != nil {
		s := entry.ScheduledDate.Format("2006-01-02")
		scheduledDate = &s
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE entries
		SET type = ?, content = ?, parent_id = ?, depth = ?, location = ?, scheduled_date = ?
		WHERE id = ?
	`, entry.Type, entry.Content, entry.ParentID, entry.Depth, entry.Location, scheduledDate, entry.ID)

	return err
}

func (r *EntryRepository) Delete(ctx context.Context, id int64) error {
	entry, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if entry == nil {
		return nil // Already deleted or doesn't exist
	}

	now := time.Now().Format(time.RFC3339)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Close current version
	_, err = tx.ExecContext(ctx, `
		UPDATE entries SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, entry.EntityID.String())
	if err != nil {
		return err
	}

	// Get next version number
	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM entries WHERE entity_id = ?
	`, entry.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	// Insert delete marker
	var scheduledDate *string
	if entry.ScheduledDate != nil {
		s := entry.ScheduledDate.Format("2006-01-02")
		scheduledDate = &s
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO entries (type, content, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.Type, entry.Content, entry.ParentID, entry.Depth, entry.Location, scheduledDate,
		entry.CreatedAt.Format(time.RFC3339), entry.EntityID.String(), maxVersion+1, now, domain.OpTypeDelete.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *EntryRepository) GetChildren(ctx context.Context, parentID int64) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM entries WHERE parent_id = ? AND (valid_to IS NULL OR valid_to = '')
		ORDER BY id
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) DeleteWithChildren(ctx context.Context, id int64) error {
	// Get all entries in the tree
	entries, err := r.GetWithChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	// Soft delete each entry
	for _, entry := range entries {
		if err := r.Delete(ctx, entry.ID); err != nil {
			return err
		}
	}

	return nil
}

func (r *EntryRepository) GetDeleted(ctx context.Context) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.type, e.content, e.parent_id, e.depth, e.location, e.scheduled_date, e.created_at, e.entity_id
		FROM entries e
		WHERE e.op_type = 'DELETE'
		AND e.valid_to IS NULL
		ORDER BY e.valid_from DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) Restore(ctx context.Context, entityID domain.EntityID) (int64, error) {
	now := time.Now().Format(time.RFC3339)

	// Get the most recent version (which should be a DELETE marker)
	var lastEntry struct {
		Type          string
		Content       string
		ParentID      sql.NullInt64
		Depth         int
		Location      sql.NullString
		ScheduledDate sql.NullString
		CreatedAt     string
		Version       int
		OpType        string
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT type, content, parent_id, depth, location, scheduled_date, created_at, version, op_type
		FROM entries WHERE entity_id = ?
		ORDER BY version DESC LIMIT 1
	`, entityID.String()).Scan(
		&lastEntry.Type, &lastEntry.Content, &lastEntry.ParentID, &lastEntry.Depth,
		&lastEntry.Location, &lastEntry.ScheduledDate, &lastEntry.CreatedAt, &lastEntry.Version, &lastEntry.OpType)
	if err != nil {
		return 0, err
	}

	if lastEntry.OpType != domain.OpTypeDelete.String() {
		return 0, nil // Not deleted, nothing to restore
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	// Close the DELETE marker
	_, err = tx.ExecContext(ctx, `
		UPDATE entries SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, entityID.String())
	if err != nil {
		return 0, err
	}

	// Insert a new version with INSERT op_type to restore
	var parentID *int64
	if lastEntry.ParentID.Valid {
		parentID = &lastEntry.ParentID.Int64
	}
	var location *string
	if lastEntry.Location.Valid {
		location = &lastEntry.Location.String
	}
	var scheduledDate *string
	if lastEntry.ScheduledDate.Valid {
		scheduledDate = &lastEntry.ScheduledDate.String
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO entries (type, content, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, lastEntry.Type, lastEntry.Content, parentID, lastEntry.Depth, location, scheduledDate,
		lastEntry.CreatedAt, entityID.String(), lastEntry.Version+1, now, domain.OpTypeInsert.String())
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *EntryRepository) scanEntry(row *sql.Row) (*domain.Entry, error) {
	var entry domain.Entry
	var typeStr string
	var scheduledDate, location, createdAt, entityID sql.NullString
	var parentID sql.NullInt64

	err := row.Scan(&entry.ID, &typeStr, &entry.Content, &parentID, &entry.Depth, &location, &scheduledDate, &createdAt, &entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	entry.Type = domain.EntryType(typeStr)

	if parentID.Valid {
		entry.ParentID = &parentID.Int64
	}
	if location.Valid {
		entry.Location = &location.String
	}
	if scheduledDate.Valid {
		t, _ := time.Parse("2006-01-02", scheduledDate.String)
		entry.ScheduledDate = &t
	}
	if createdAt.Valid {
		entry.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}
	if entityID.Valid {
		entry.EntityID = domain.EntityID(entityID.String)
	}

	return &entry, nil
}

func (r *EntryRepository) scanEntries(rows *sql.Rows) ([]domain.Entry, error) {
	var entries []domain.Entry

	for rows.Next() {
		var entry domain.Entry
		var typeStr string
		var scheduledDate, location, createdAt, entityID sql.NullString
		var parentID sql.NullInt64

		err := rows.Scan(&entry.ID, &typeStr, &entry.Content, &parentID, &entry.Depth, &location, &scheduledDate, &createdAt, &entityID)
		if err != nil {
			return nil, err
		}

		entry.Type = domain.EntryType(typeStr)

		if parentID.Valid {
			entry.ParentID = &parentID.Int64
		}
		if location.Valid {
			entry.Location = &location.String
		}
		if scheduledDate.Valid {
			t, _ := time.Parse("2006-01-02", scheduledDate.String)
			entry.ScheduledDate = &t
		}
		if createdAt.Valid {
			entry.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		}
		if entityID.Valid {
			entry.EntityID = domain.EntityID(entityID.String)
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *EntryRepository) GetHistory(ctx context.Context, entityID domain.EntityID) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM entries WHERE entity_id = ?
		ORDER BY version
	`, entityID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) GetAsOf(ctx context.Context, entityID domain.EntityID, asOf time.Time) (*domain.Entry, error) {
	asOfStr := asOf.Format(time.RFC3339)
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id
		FROM entries
		WHERE entity_id = ?
		AND valid_from <= ?
		AND (valid_to IS NULL OR valid_to = '' OR valid_to > ?)
		ORDER BY version DESC
		LIMIT 1
	`, entityID.String(), asOfStr, asOfStr)

	return r.scanEntry(row)
}
