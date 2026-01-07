package sqlite

import (
	"context"
	"database/sql"

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

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO lists (name, created_at) VALUES (?, ?)",
		list.Name, list.CreatedAt,
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

func (r *ListRepository) GetByID(ctx context.Context, id int64) (*domain.List, error) {
	var list domain.List
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, created_at FROM lists WHERE id = ?",
		id,
	).Scan(&list.ID, &list.Name, &list.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *ListRepository) GetByName(ctx context.Context, name string) (*domain.List, error) {
	var list domain.List
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, created_at FROM lists WHERE name = ?",
		name,
	).Scan(&list.ID, &list.Name, &list.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *ListRepository) GetAll(ctx context.Context) ([]domain.List, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, created_at FROM lists ORDER BY name",
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var lists []domain.List
	for rows.Next() {
		var list domain.List
		if err := rows.Scan(&list.ID, &list.Name, &list.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lists, nil
}

func (r *ListRepository) Rename(ctx context.Context, id int64, newName string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE lists SET name = ? WHERE id = ?",
		newName, id,
	)
	return err
}

func (r *ListRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM lists WHERE id = ?",
		id,
	)
	return err
}

func (r *ListRepository) GetItemCount(ctx context.Context, listID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM entries WHERE list_id = ?",
		listID,
	).Scan(&count)
	return count, err
}

func (r *ListRepository) GetDoneCount(ctx context.Context, listID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM entries WHERE list_id = ? AND type = 'done'",
		listID,
	).Scan(&count)
	return count, err
}
