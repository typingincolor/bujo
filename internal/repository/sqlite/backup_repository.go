package sqlite

import (
	"context"
	"database/sql"
)

type BackupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

func (r *BackupRepository) Backup(ctx context.Context, destPath string) error {
	_, err := r.db.ExecContext(ctx, "VACUUM INTO ?", destPath)
	return err
}
