package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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

func (r *BackupRepository) VerifyIntegrity(ctx context.Context, backupPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	db, err := sql.Open("sqlite", backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer func() { _ = db.Close() }()

	var result string
	err = db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to verify backup: %w", err)
	}

	if result != "ok" {
		return fmt.Errorf("backup integrity check failed: %s", result)
	}

	return nil
}
