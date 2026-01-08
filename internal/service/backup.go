package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type BackupService struct {
	db        *sql.DB
	backupDir string
}

func NewBackupService(db *sql.DB, backupDir string) *BackupService {
	return &BackupService{
		db:        db,
		backupDir: backupDir,
	}
}

func (s *BackupService) CreateBackup(ctx context.Context) (string, error) {
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02-150405.000000000")
	filename := fmt.Sprintf("bujo-%s.db", timestamp)
	destPath := filepath.Join(s.backupDir, filename)

	_, err := s.db.ExecContext(ctx, "VACUUM INTO ?", destPath)
	if err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}

	return destPath, nil
}

type BackupInfo struct {
	Path      string
	Filename  string
	CreatedAt time.Time
	Size      int64
}

func (s *BackupService) ListBackups(ctx context.Context) ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "bujo-") || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      filepath.Join(s.backupDir, entry.Name()),
			Filename:  entry.Name(),
			CreatedAt: info.ModTime(),
			Size:      info.Size(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

func (s *BackupService) VerifyBackup(ctx context.Context, backupPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	db, err := sql.Open("sqlite3", backupPath)
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
