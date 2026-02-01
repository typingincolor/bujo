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
		DELETE FROM entries WHERE id = ?
	`, entry.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
