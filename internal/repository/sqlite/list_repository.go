package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type ListRepository struct {
	db *sql.DB
}

func NewListRepository(db *sql.DB) *ListRepository {
	return &ListRepository{db: db}
}

func (r *ListRepository) Create(ctx context.Context, name string) (*domain.List, error) {
	list := domain.NewList(name)
	if err := list.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO lists (name, entity_id, created_at, version, valid_from, op_type) VALUES (?, ?, ?, ?, ?, ?)",
		list.Name, list.EntityID.String(), list.CreatedAt, 1, now, domain.OpTypeInsert.String(),
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	list.ID = id
	return &list, nil
}

func (r *ListRepository) InsertWithEntityID(ctx context.Context, list domain.List) (int64, error) {
	now := time.Now().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO lists (name, entity_id, created_at, version, valid_from, op_type) VALUES (?, ?, ?, ?, ?, ?)",
		list.Name, list.EntityID.String(), list.CreatedAt.Format(time.RFC3339), 1, now, domain.OpTypeInsert.String(),
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *ListRepository) GetByID(ctx context.Context, id int64) (*domain.List, error) {
	// First, get the entity_id for this ID (may be from a closed version)
	var entityID string
	err := r.db.QueryRowContext(ctx,
		"SELECT entity_id FROM lists WHERE id = ?", id,
	).Scan(&entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Then get the current version for that entity
	var list domain.List
	var eid sql.NullString
	var createdAt string
	err = r.db.QueryRowContext(ctx,
		"SELECT id, entity_id, name, created_at FROM lists WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'",
		entityID,
	).Scan(&list.ID, &eid, &list.Name, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	list.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if eid.Valid {
		list.EntityID = domain.EntityID(eid.String)
	}
	return &list, nil
}

func (r *ListRepository) GetByName(ctx context.Context, name string) (*domain.List, error) {
	var list domain.List
	var entityID sql.NullString
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		"SELECT id, entity_id, name, created_at FROM lists WHERE name = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'",
		name,
	).Scan(&list.ID, &entityID, &list.Name, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	list.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if entityID.Valid {
		list.EntityID = domain.EntityID(entityID.String)
	}
	return &list, nil
}

func (r *ListRepository) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.List, error) {
	var list domain.List
	var eid sql.NullString
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		"SELECT id, entity_id, name, created_at FROM lists WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'",
		entityID.String(),
	).Scan(&list.ID, &eid, &list.Name, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	list.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if eid.Valid {
		list.EntityID = domain.EntityID(eid.String)
	}
	return &list, nil
}

func (r *ListRepository) GetAll(ctx context.Context) ([]domain.List, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, entity_id, name, created_at FROM lists WHERE (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE' ORDER BY name",
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var lists []domain.List
	for rows.Next() {
		var list domain.List
		var entityID sql.NullString
		var createdAt string
		if err := rows.Scan(&list.ID, &entityID, &list.Name, &createdAt); err != nil {
			return nil, err
		}
		list.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if entityID.Valid {
			list.EntityID = domain.EntityID(entityID.String)
		}
		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lists, nil
}

func (r *ListRepository) Rename(ctx context.Context, id int64, newName string) error {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("list not found: %d", id)
	}

	now := time.Now().Format(time.RFC3339)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Close current version
	_, err = tx.ExecContext(ctx, `
		UPDATE lists SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, current.EntityID.String())
	if err != nil {
		return err
	}

	// Get next version number
	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM lists WHERE entity_id = ?
	`, current.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	// Insert new version with UPDATE op_type
	_, err = tx.ExecContext(ctx, `
		INSERT INTO lists (name, entity_id, created_at, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?)
	`, newName, current.EntityID.String(), current.CreatedAt.Format(time.RFC3339),
		maxVersion+1, now, domain.OpTypeUpdate.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ListRepository) Delete(ctx context.Context, id int64) error {
	list, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if list == nil {
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
		UPDATE lists SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, list.EntityID.String())
	if err != nil {
		return err
	}

	// Get next version number
	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM lists WHERE entity_id = ?
	`, list.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	// Insert delete marker
	_, err = tx.ExecContext(ctx, `
		INSERT INTO lists (name, entity_id, created_at, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?)
	`, list.Name, list.EntityID.String(), list.CreatedAt.Format(time.RFC3339),
		maxVersion+1, now, domain.OpTypeDelete.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ListRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM lists")
	return err
}

func (r *ListRepository) GetDeleted(ctx context.Context) ([]domain.List, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, entity_id, name, created_at
		FROM lists
		WHERE op_type = 'DELETE'
		AND valid_to IS NULL
		ORDER BY valid_from DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var lists []domain.List
	for rows.Next() {
		var list domain.List
		var entityID sql.NullString
		var createdAt string
		if err := rows.Scan(&list.ID, &entityID, &list.Name, &createdAt); err != nil {
			return nil, err
		}
		list.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if entityID.Valid {
			list.EntityID = domain.EntityID(entityID.String)
		}
		lists = append(lists, list)
	}

	return lists, rows.Err()
}

func (r *ListRepository) Restore(ctx context.Context, entityID domain.EntityID) (int64, error) {
	now := time.Now().Format(time.RFC3339)

	// Get the most recent version (which should be a DELETE marker)
	var lastList struct {
		Name      string
		CreatedAt string
		Version   int
		OpType    string
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT name, created_at, version, op_type
		FROM lists WHERE entity_id = ?
		ORDER BY version DESC LIMIT 1
	`, entityID.String()).Scan(
		&lastList.Name, &lastList.CreatedAt, &lastList.Version, &lastList.OpType)
	if err != nil {
		return 0, err
	}

	if lastList.OpType != domain.OpTypeDelete.String() {
		return 0, nil // Not deleted, nothing to restore
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	// Close the DELETE marker
	_, err = tx.ExecContext(ctx, `
		UPDATE lists SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, entityID.String())
	if err != nil {
		return 0, err
	}

	// Insert a new version with INSERT op_type to restore
	result, err := tx.ExecContext(ctx, `
		INSERT INTO lists (name, entity_id, created_at, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?)
	`, lastList.Name, entityID.String(), lastList.CreatedAt,
		lastList.Version+1, now, domain.OpTypeInsert.String())
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *ListRepository) GetItemCount(ctx context.Context, listID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM list_items li
		JOIN lists l ON li.list_entity_id = l.entity_id
		WHERE l.id = ? AND li.valid_to IS NULL AND li.op_type != 'DELETE'
	`, listID).Scan(&count)
	return count, err
}

func (r *ListRepository) GetDoneCount(ctx context.Context, listID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM list_items li
		JOIN lists l ON li.list_entity_id = l.entity_id
		WHERE l.id = ? AND li.type = 'done' AND li.valid_to IS NULL AND li.op_type != 'DELETE'
	`, listID).Scan(&count)
	return count, err
}
