package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type MentionRepository struct {
	db *sql.DB
}

func NewMentionRepository(db *sql.DB) *MentionRepository {
	return &MentionRepository{db: db}
}

func (r *MentionRepository) InsertEntryMentions(ctx context.Context, entryID int64, mentions []string) error {
	if len(mentions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, "INSERT OR IGNORE INTO entry_mentions (entry_id, mention) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, mention := range mentions {
		if _, err := stmt.ExecContext(ctx, entryID, mention); err != nil {
			return fmt.Errorf("insert mention %q: %w", mention, err)
		}
	}

	return tx.Commit()
}

func (r *MentionRepository) GetMentionsForEntries(ctx context.Context, entryIDs []int64) (map[int64][]string, error) {
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
		"SELECT entry_id, mention FROM entry_mentions WHERE entry_id IN (%s) ORDER BY entry_id, mention",
		strings.Join(placeholders, ","),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query mentions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var entryID int64
		var mention string
		if err := rows.Scan(&entryID, &mention); err != nil {
			return nil, fmt.Errorf("scan mention: %w", err)
		}
		result[entryID] = append(result[entryID], mention)
	}

	return result, rows.Err()
}

func (r *MentionRepository) GetAllMentions(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT DISTINCT mention FROM entry_mentions ORDER BY mention")
	if err != nil {
		return nil, fmt.Errorf("query all mentions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var mentions []string
	for rows.Next() {
		var mention string
		if err := rows.Scan(&mention); err != nil {
			return nil, fmt.Errorf("scan mention: %w", err)
		}
		mentions = append(mentions, mention)
	}

	return mentions, rows.Err()
}

func (r *MentionRepository) DeleteByEntryID(ctx context.Context, entryID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM entry_mentions WHERE entry_id = ?", entryID)
	if err != nil {
		return fmt.Errorf("delete mentions for entry %d: %w", entryID, err)
	}
	return nil
}
