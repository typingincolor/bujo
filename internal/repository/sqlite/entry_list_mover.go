package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type EntryToListMover struct {
	db *sql.DB
}

func NewEntryToListMover(db *sql.DB) *EntryToListMover {
	return &EntryToListMover{db: db}
}

func (m *EntryToListMover) MoveEntryToList(ctx context.Context, entry domain.Entry, listEntityID domain.EntityID) error {
	now := time.Now().Format(time.RFC3339)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	listItem := domain.NewListItem(listEntityID, domain.ListItemTypeTask, entry.Content)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO list_items (entity_id, version, valid_from, op_type, list_entity_id, type, content, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, listItem.EntityID.String(), listItem.Version, now, domain.OpTypeInsert.String(),
		listItem.ListEntityID.String(), string(listItem.Type), listItem.Content, listItem.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE entries SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, entry.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM entries WHERE entity_id = ?
	`, entry.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

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

	_, err = tx.ExecContext(ctx, `
		INSERT INTO entries (type, content, priority, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.Type, entry.Content, priority, entry.ParentID, entry.Depth, entry.Location, scheduledDateStr,
		entry.CreatedAt.Format(time.RFC3339), entry.EntityID.String(), maxVersion+1, now, domain.OpTypeDelete.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}
