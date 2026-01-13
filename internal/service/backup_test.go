package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func TestBackupService_CreateBackup_CreatesFile(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	backupPath, err := svc.CreateBackup(ctx, tempDir)

	require.NoError(t, err)
	assert.FileExists(t, backupPath)
	assert.Contains(t, backupPath, tempDir)
	assert.Contains(t, backupPath, "bujo-")
	assert.Contains(t, backupPath, ".db")
}

func TestBackupService_CreateBackup_ValidSQLite(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	backupPath, err := svc.CreateBackup(ctx, tempDir)
	require.NoError(t, err)

	// Verify we can open the backup as a valid SQLite database
	backupDB, err := sqlite.OpenAndMigrate(backupPath)
	require.NoError(t, err)
	defer func() { _ = backupDB.Close() }()

	// Simple query to verify it works
	var result int
	err = backupDB.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestBackupService_ListBackups_ReturnsFiles(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	// Create a few backups
	_, err = svc.CreateBackup(ctx, tempDir)
	require.NoError(t, err)
	_, err = svc.CreateBackup(ctx, tempDir)
	require.NoError(t, err)

	backups, err := svc.ListBackups(ctx, tempDir)

	require.NoError(t, err)
	assert.Len(t, backups, 2)
}

func TestBackupService_VerifyBackup_ValidFile_Succeeds(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	backupPath, err := svc.CreateBackup(ctx, tempDir)
	require.NoError(t, err)

	err = svc.VerifyBackup(ctx, backupPath)

	require.NoError(t, err)
}

func TestBackupService_VerifyBackup_CorruptFile_Fails(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	// Create a corrupt file
	corruptPath := filepath.Join(tempDir, "corrupt.db")
	err = os.WriteFile(corruptPath, []byte("not a valid sqlite file"), 0644)
	require.NoError(t, err)

	err = svc.VerifyBackup(ctx, corruptPath)

	require.Error(t, err)
}

func TestBackupService_VerifyBackup_MissingFile_Fails(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	err = svc.VerifyBackup(ctx, "/nonexistent/path.db")

	require.Error(t, err)
}

func TestBackupService_EnsureRecentBackup_NoBackups_CreatesOne(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	created, path, err := svc.EnsureRecentBackup(ctx, tempDir, 7)

	require.NoError(t, err)
	assert.True(t, created, "should create backup when none exists")
	assert.NotEmpty(t, path)
	assert.FileExists(t, path)
}

func TestBackupService_EnsureRecentBackup_RecentBackup_DoesNothing(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	// Create a backup first
	_, err = svc.CreateBackup(ctx, tempDir)
	require.NoError(t, err)

	created, path, err := svc.EnsureRecentBackup(ctx, tempDir, 7)

	require.NoError(t, err)
	assert.False(t, created, "should not create backup when recent one exists")
	assert.Empty(t, path)
}

func TestBackupService_EnsureRecentBackup_OldBackup_CreatesNew(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	repo := sqlite.NewBackupRepository(db)
	svc := NewBackupService(repo)
	ctx := context.Background()

	// Create a backup and make it old by changing its modification time
	backupPath, err := svc.CreateBackup(ctx, tempDir)
	require.NoError(t, err)

	oldTime := time.Now().Add(-8 * 24 * time.Hour) // 8 days ago
	err = os.Chtimes(backupPath, oldTime, oldTime)
	require.NoError(t, err)

	created, path, err := svc.EnsureRecentBackup(ctx, tempDir, 7)

	require.NoError(t, err)
	assert.True(t, created, "should create backup when existing is too old")
	assert.NotEmpty(t, path)
	assert.NotEqual(t, backupPath, path, "should create a new backup file")
}
