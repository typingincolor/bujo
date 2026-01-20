package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupRepository_VerifyIntegrity_ValidBackup_ReturnsNil(t *testing.T) {
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	backupPath := filepath.Join(tempDir, "test-backup.db")

	repo := NewBackupRepository(db)
	ctx := context.Background()

	err = repo.Backup(ctx, backupPath)
	require.NoError(t, err)

	err = repo.VerifyIntegrity(ctx, backupPath)

	assert.NoError(t, err)
}

func TestBackupRepository_VerifyIntegrity_CorruptFile_ReturnsError(t *testing.T) {
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tempDir := t.TempDir()
	corruptPath := filepath.Join(tempDir, "corrupt.db")
	err = os.WriteFile(corruptPath, []byte("not a valid sqlite file"), 0644)
	require.NoError(t, err)

	repo := NewBackupRepository(db)
	ctx := context.Background()

	err = repo.VerifyIntegrity(ctx, corruptPath)

	assert.Error(t, err)
}

func TestBackupRepository_VerifyIntegrity_MissingFile_ReturnsError(t *testing.T) {
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewBackupRepository(db)
	ctx := context.Background()

	err = repo.VerifyIntegrity(ctx, "/nonexistent/path.db")

	assert.Error(t, err)
}
