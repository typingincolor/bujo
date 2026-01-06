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

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO entries (type, content, parent_id, depth, location, scheduled_date, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, entry.Type, entry.Content, entry.ParentID, entry.Depth, entry.Location, scheduledDate, entry.CreatedAt.Format(time.RFC3339))

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *EntryRepository) GetByID(ctx context.Context, id int64) (*domain.Entry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at
		FROM entries WHERE id = ?
	`, id)

	return r.scanEntry(row)
}

func (r *EntryRepository) GetByDate(ctx context.Context, date time.Time) ([]domain.Entry, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at
		FROM entries WHERE scheduled_date = ?
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
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at
		FROM entries WHERE scheduled_date >= ? AND scheduled_date <= ?
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
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at
		FROM entries
		WHERE scheduled_date < ? AND type NOT IN ('done', 'migrated')
		ORDER BY scheduled_date, created_at
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
			SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at
			FROM entries WHERE id = ?
			UNION ALL
			SELECT e.id, e.type, e.content, e.parent_id, e.depth, e.location, e.scheduled_date, e.created_at
			FROM entries e
			JOIN tree t ON e.parent_id = t.id
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
	_, err := r.db.ExecContext(ctx, "DELETE FROM entries WHERE id = ?", id)
	return err
}

func (r *EntryRepository) GetChildren(ctx context.Context, parentID int64) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at
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
	// Use recursive CTE to find all descendants, then delete
	_, err := r.db.ExecContext(ctx, `
		WITH RECURSIVE tree AS (
			SELECT id FROM entries WHERE id = ?
			UNION ALL
			SELECT e.id FROM entries e
			JOIN tree t ON e.parent_id = t.id
		)
		DELETE FROM entries WHERE id IN (SELECT id FROM tree)
	`, id)
	return err
}

func (r *EntryRepository) scanEntry(row *sql.Row) (*domain.Entry, error) {
	var entry domain.Entry
	var typeStr string
	var scheduledDate, location, createdAt sql.NullString
	var parentID sql.NullInt64

	err := row.Scan(&entry.ID, &typeStr, &entry.Content, &parentID, &entry.Depth, &location, &scheduledDate, &createdAt)
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

	return &entry, nil
}

func (r *EntryRepository) scanEntries(rows *sql.Rows) ([]domain.Entry, error) {
	var entries []domain.Entry

	for rows.Next() {
		var entry domain.Entry
		var typeStr string
		var scheduledDate, location, createdAt sql.NullString
		var parentID sql.NullInt64

		err := rows.Scan(&entry.ID, &typeStr, &entry.Content, &parentID, &entry.Depth, &location, &scheduledDate, &createdAt)
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

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
