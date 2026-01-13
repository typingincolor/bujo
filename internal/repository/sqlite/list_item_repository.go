package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type ListItemRepository struct {
	db *sql.DB
}

func NewListItemRepository(db *sql.DB) *ListItemRepository {
	return &ListItemRepository{db: db}
}

func (r *ListItemRepository) Insert(ctx context.Context, item domain.ListItem) (int64, error) {
	now := time.Now().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO list_items (entity_id, version, valid_from, op_type, list_entity_id, type, content, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.EntityID.String(), item.Version, now, domain.OpTypeInsert.String(),
		item.ListEntityID.String(), string(item.Type), item.Content, item.CreatedAt.Format(time.RFC3339))

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *ListItemRepository) GetByID(ctx context.Context, id int64) (*domain.ListItem, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT row_id, entity_id, version, valid_from, valid_to, op_type, list_entity_id, type, content, created_at
		FROM list_items
		WHERE row_id = ? AND valid_to IS NULL
	`, id)

	item, err := r.scanItem(row)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}

	var entityID string
	err = r.db.QueryRowContext(ctx, `SELECT entity_id FROM list_items WHERE row_id = ?`, id).Scan(&entityID)
	if err != nil {
		return nil, nil // Row doesn't exist at all
	}

	return r.GetByEntityID(ctx, domain.EntityID(entityID))
}

func (r *ListItemRepository) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.ListItem, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT row_id, entity_id, version, valid_from, valid_to, op_type, list_entity_id, type, content, created_at
		FROM list_items
		WHERE entity_id = ? AND valid_to IS NULL AND op_type != 'DELETE'
	`, entityID.String())

	return r.scanItem(row)
}

func (r *ListItemRepository) GetByListEntityID(ctx context.Context, listEntityID domain.EntityID) ([]domain.ListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT row_id, entity_id, version, valid_from, valid_to, op_type, list_entity_id, type, content, created_at
		FROM list_items
		WHERE list_entity_id = ? AND valid_to IS NULL AND op_type != 'DELETE'
		ORDER BY row_id
	`, listEntityID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanItems(rows)
}

func (r *ListItemRepository) GetByListID(ctx context.Context, listID int64) ([]domain.ListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT li.row_id, li.entity_id, li.version, li.valid_from, li.valid_to, li.op_type, li.list_entity_id, li.type, li.content, li.created_at
		FROM list_items li
		JOIN lists l ON li.list_entity_id = l.entity_id
		WHERE l.id = ? AND li.valid_to IS NULL AND li.op_type != 'DELETE'
		ORDER BY li.row_id
	`, listID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanItems(rows)
}

func (r *ListItemRepository) GetAll(ctx context.Context) ([]domain.ListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT row_id, entity_id, version, valid_from, valid_to, op_type, list_entity_id, type, content, created_at
		FROM list_items
		WHERE valid_to IS NULL AND op_type != 'DELETE'
		ORDER BY row_id
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanItems(rows)
}

func (r *ListItemRepository) Update(ctx context.Context, item domain.ListItem) error {
	now := time.Now().Format(time.RFC3339)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE list_items SET valid_to = ? WHERE entity_id = ? AND valid_to IS NULL
	`, now, item.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM list_items WHERE entity_id = ?
	`, item.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO list_items (entity_id, version, valid_from, op_type, list_entity_id, type, content, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.EntityID.String(), maxVersion+1, now, domain.OpTypeUpdate.String(),
		item.ListEntityID.String(), string(item.Type), item.Content, item.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ListItemRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().Format(time.RFC3339)

	item, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return nil // Already deleted or doesn't exist
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE list_items SET valid_to = ? WHERE entity_id = ? AND valid_to IS NULL
	`, now, item.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM list_items WHERE entity_id = ?
	`, item.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO list_items (entity_id, version, valid_from, op_type, list_entity_id, type, content, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.EntityID.String(), maxVersion+1, now, domain.OpTypeDelete.String(),
		item.ListEntityID.String(), string(item.Type), item.Content, item.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ListItemRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM list_items")
	return err
}

func (r *ListItemRepository) GetHistory(ctx context.Context, entityID domain.EntityID) ([]domain.ListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT row_id, entity_id, version, valid_from, valid_to, op_type, list_entity_id, type, content, created_at
		FROM list_items
		WHERE entity_id = ?
		ORDER BY version
	`, entityID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanItems(rows)
}

func (r *ListItemRepository) scanItem(row *sql.Row) (*domain.ListItem, error) {
	var item domain.ListItem
	var entityID, listEntityID, opType, itemType string
	var validFrom, createdAt string
	var validTo sql.NullString

	err := row.Scan(
		&item.RowID, &entityID, &item.Version, &validFrom, &validTo, &opType,
		&listEntityID, &itemType, &item.Content, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	item.EntityID = domain.EntityID(entityID)
	item.ListEntityID = domain.EntityID(listEntityID)
	item.OpType = domain.OpType(opType)
	item.Type = domain.ListItemType(itemType)
	item.ValidFrom, _ = time.Parse(time.RFC3339, validFrom)
	if validTo.Valid {
		t, _ := time.Parse(time.RFC3339, validTo.String)
		item.ValidTo = &t
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &item, nil
}

func (r *ListItemRepository) scanItems(rows *sql.Rows) ([]domain.ListItem, error) {
	var items []domain.ListItem

	for rows.Next() {
		var item domain.ListItem
		var entityID, listEntityID, opType, itemType string
		var validFrom, createdAt string
		var validTo sql.NullString

		err := rows.Scan(
			&item.RowID, &entityID, &item.Version, &validFrom, &validTo, &opType,
			&listEntityID, &itemType, &item.Content, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		item.EntityID = domain.EntityID(entityID)
		item.ListEntityID = domain.EntityID(listEntityID)
		item.OpType = domain.OpType(opType)
		item.Type = domain.ListItemType(itemType)
		item.ValidFrom, _ = time.Parse(time.RFC3339, validFrom)
		if validTo.Valid {
			t, _ := time.Parse(time.RFC3339, validTo.String)
			item.ValidTo = &t
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ListItemRepository) GetAtVersion(ctx context.Context, entityID domain.EntityID, version int) (*domain.ListItem, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT row_id, entity_id, version, valid_from, valid_to, op_type, list_entity_id, type, content, created_at
		FROM list_items
		WHERE entity_id = ? AND version = ?
	`, entityID.String(), version)

	return r.scanItem(row)
}

func (r *ListItemRepository) CountArchivable(ctx context.Context, olderThan time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM list_items
		WHERE valid_to IS NOT NULL AND valid_to < ?
	`, olderThan.Format(time.RFC3339)).Scan(&count)
	return count, err
}

func (r *ListItemRepository) DeleteArchivable(ctx context.Context, olderThan time.Time) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM list_items
		WHERE valid_to IS NOT NULL AND valid_to < ?
	`, olderThan.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}
