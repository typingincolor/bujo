package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	var scheduledDateStr string
	if entry.ScheduledDate != nil {
		scheduledDateStr = entry.ScheduledDate.Format("2006-01-02")
	} else {
		scheduledDateStr = entry.CreatedAt.Format("2006-01-02")
	}

	entityID := entry.EntityID
	if entityID.IsEmpty() {
		entityID = domain.NewEntityID()
	}
	now := time.Now().Format(time.RFC3339)

	priority := entry.Priority
	if priority == "" {
		priority = domain.PriorityNone
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO entries (type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, op_type, sort_order, migration_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, 'INSERT', ?, ?)
	`, entry.Type, entry.Content, priority, entry.ParentID, entry.Depth, entry.Location, scheduledDateStr, entry.CreatedAt.Format(time.RFC3339),
		entityID.String(), now, entry.SortOrder, entry.MigrationCount)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *EntryRepository) GetByID(ctx context.Context, id int64) (*domain.Entry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
		FROM entries WHERE id = ?
	`, id)

	return r.scanEntry(row)
}

func (r *EntryRepository) GetByDate(ctx context.Context, date time.Time) ([]domain.Entry, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
		FROM entries WHERE scheduled_date = ?
		ORDER BY created_at, id
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
		SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
		FROM entries WHERE scheduled_date >= ? AND scheduled_date <= ?
		ORDER BY scheduled_date, created_at, id
	`, fromStr, toStr)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) GetOverdue(ctx context.Context) ([]domain.Entry, error) {
	dateStr := time.Now().Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		WITH RECURSIVE
		overdue_tasks AS (
			SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
			FROM entries
			WHERE scheduled_date < ? AND type = 'task'
		),
		parent_chain AS (
			SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
			FROM overdue_tasks
			UNION
			SELECT e.id, e.type, e.content, e.priority, e.parent_id, e.depth, e.location, e.scheduled_date, e.created_at, e.entity_id, e.sort_order, e.migration_count
			FROM entries e
			INNER JOIN parent_chain pc ON e.id = pc.parent_id
		)
		SELECT DISTINCT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
		FROM parent_chain
		ORDER BY scheduled_date, depth, created_at, id
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
			SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
			FROM entries WHERE id = ?
			UNION ALL
			SELECT e.id, e.type, e.content, e.priority, e.parent_id, e.depth, e.location, e.scheduled_date, e.created_at, e.entity_id, e.sort_order, e.migration_count
			FROM entries e
			JOIN tree t ON e.parent_id = t.id
		)
		SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count FROM tree ORDER BY depth, id
	`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) Update(ctx context.Context, entry domain.Entry) error {
	var scheduledDateStr string
	if entry.ScheduledDate != nil {
		scheduledDateStr = entry.ScheduledDate.Format("2006-01-02")
	} else {
		scheduledDateStr = entry.CreatedAt.Format("2006-01-02")
	}

	priority := entry.Priority
	if priority == "" {
		priority = domain.PriorityNone
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE entries SET type = ?, content = ?, priority = ?, parent_id = ?, depth = ?, location = ?, scheduled_date = ?, sort_order = ?, migration_count = ?
		WHERE id = ?
	`, entry.Type, entry.Content, priority, entry.ParentID, entry.Depth, entry.Location, scheduledDateStr, entry.SortOrder, entry.MigrationCount, entry.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("entry not found: %d", entry.ID)
	}

	return nil
}

func (r *EntryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM entries WHERE id = ?`, id)
	return err
}

func (r *EntryRepository) DeleteByDate(ctx context.Context, date time.Time) error {
	dateStr := date.Format("2006-01-02")
	_, err := r.db.ExecContext(ctx, "DELETE FROM entries WHERE scheduled_date = ?", dateStr)
	return err
}

func (r *EntryRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM entries")
	return err
}

func (r *EntryRepository) GetChildren(ctx context.Context, parentID int64) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
		FROM entries WHERE parent_id = ?
		ORDER BY id
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) DeleteWithChildren(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `
		WITH RECURSIVE tree AS (
			SELECT id FROM entries WHERE id = ?
			UNION ALL
			SELECT e.id FROM entries e JOIN tree t ON e.parent_id = t.id
		)
		DELETE FROM entries WHERE id IN (SELECT id FROM tree)
	`, id)
	return err
}

func (r *EntryRepository) scanEntry(row *sql.Row) (*domain.Entry, error) {
	var entry domain.Entry
	var typeStr, priorityStr string
	var scheduledDate, location, createdAt, entityID sql.NullString
	var parentID sql.NullInt64

	err := row.Scan(&entry.ID, &typeStr, &entry.Content, &priorityStr, &parentID, &entry.Depth, &location, &scheduledDate, &createdAt, &entityID, &entry.SortOrder, &entry.MigrationCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	entry.Type = domain.EntryType(typeStr)
	entry.Priority = domain.Priority(priorityStr)

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
		var typeStr, priorityStr string
		var scheduledDate, location, createdAt, entityID sql.NullString
		var parentID sql.NullInt64

		err := rows.Scan(&entry.ID, &typeStr, &entry.Content, &priorityStr, &parentID, &entry.Depth, &location, &scheduledDate, &createdAt, &entityID, &entry.SortOrder, &entry.MigrationCount)
		if err != nil {
			return nil, err
		}

		entry.Type = domain.EntryType(typeStr)
		entry.Priority = domain.Priority(priorityStr)

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

func (r *EntryRepository) GetAll(ctx context.Context) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, sort_order, migration_count
		FROM entries
		ORDER BY scheduled_date, created_at, id
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}

func (r *EntryRepository) GetLastModified(ctx context.Context) (time.Time, error) {
	var validFrom sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT MAX(valid_from) FROM entries
	`).Scan(&validFrom)
	if err != nil {
		return time.Time{}, err
	}
	if !validFrom.Valid {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, validFrom.String)
}

func (r *EntryRepository) Search(ctx context.Context, opts domain.SearchOptions) ([]domain.Entry, error) {
	if opts.Query == "" && opts.Type == nil && len(opts.Tags) == 0 && len(opts.Mentions) == 0 {
		return []domain.Entry{}, nil
	}

	query := `
		SELECT DISTINCT e.id, e.type, e.content, e.priority, e.parent_id, e.depth, e.location, e.scheduled_date, e.created_at, e.entity_id, e.sort_order, e.migration_count
		FROM entries e
	`
	var args []any

	if len(opts.Tags) > 0 {
		placeholders := make([]string, len(opts.Tags))
		for i, tag := range opts.Tags {
			placeholders[i] = "?"
			args = append(args, tag)
		}
		query += fmt.Sprintf(` JOIN entry_tags et ON e.id = et.entry_id AND et.tag IN (%s)`, strings.Join(placeholders, ","))
	}

	if len(opts.Mentions) > 0 {
		placeholders := make([]string, len(opts.Mentions))
		for i, mention := range opts.Mentions {
			placeholders[i] = "?"
			args = append(args, mention)
		}
		query += fmt.Sprintf(` JOIN entry_mentions em ON e.id = em.entry_id AND em.mention IN (%s)`, strings.Join(placeholders, ","))
	}

	query += ` WHERE 1=1`

	if opts.Query != "" {
		query += ` AND e.content LIKE '%' || ? || '%' COLLATE NOCASE`
		args = append(args, opts.Query)
	}

	if opts.Type != nil {
		query += ` AND e.type = ?`
		args = append(args, string(*opts.Type))
	}

	if opts.DateFrom != nil && opts.DateTo != nil {
		query += ` AND e.scheduled_date >= ? AND e.scheduled_date <= ?`
		args = append(args, opts.DateFrom.Format("2006-01-02"), opts.DateTo.Format("2006-01-02"))
	}

	query += ` ORDER BY e.scheduled_date DESC, e.created_at DESC, e.id DESC`

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	query += ` LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEntries(rows)
}
