package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) InsertEntryTags(ctx context.Context, entryID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, "INSERT OR IGNORE INTO entry_tags (entry_id, tag) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, tag := range tags {
		if _, err := stmt.ExecContext(ctx, entryID, tag); err != nil {
			return fmt.Errorf("insert tag %q: %w", tag, err)
		}
	}

	return tx.Commit()
}

func (r *TagRepository) GetTagsForEntries(ctx context.Context, entryIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string)
	if len(entryIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(entryIDs))
	args := make([]interface{}, len(entryIDs))
	for i, id := range entryIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		"SELECT entry_id, tag FROM entry_tags WHERE entry_id IN (%s) ORDER BY entry_id, tag",
		strings.Join(placeholders, ","),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var entryID int64
		var tag string
		if err := rows.Scan(&entryID, &tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		result[entryID] = append(result[entryID], tag)
	}

	return result, rows.Err()
}

func (r *TagRepository) GetAllTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT DISTINCT tag FROM entry_tags ORDER BY tag")
	if err != nil {
		return nil, fmt.Errorf("query all tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

func (r *TagRepository) DeleteByEntryID(ctx context.Context, entryID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM entry_tags WHERE entry_id = ?", entryID)
	if err != nil {
		return fmt.Errorf("delete tags for entry %d: %w", entryID, err)
	}
	return nil
}
