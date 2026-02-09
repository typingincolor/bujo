package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceFactory_Create_ReturnsAllServices(t *testing.T) {
	ctx := context.Background()

	factory := NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	assert.NotNil(t, services.DB, "DB should be available")
	assert.NotNil(t, services.Bujo, "BujoService should be created")
	assert.NotNil(t, services.Habit, "HabitService should be created")
	assert.NotNil(t, services.List, "ListService should be created")
	assert.NotNil(t, services.Goal, "GoalService should be created")
	assert.NotNil(t, services.Stats, "StatsService should be created")
}

func TestDefaultBackupDir_ReturnsExpectedPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, ".bujo", "backups")
	assert.Equal(t, expected, DefaultBackupDir())
}

func TestServiceFactory_Create_ReturnsBackupService(t *testing.T) {
	ctx := context.Background()

	factory := NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	assert.NotNil(t, services.Backup, "BackupService should be created")
}

func TestServiceFactory_Create_EnsuresRecentBackup(t *testing.T) {
	ctx := context.Background()
	backupDir := t.TempDir()

	factory := NewServiceFactory()
	_, cleanup, err := factory.Create(ctx, ":memory:", WithBackupDir(backupDir))
	require.NoError(t, err)
	defer cleanup()

	entries, err := os.ReadDir(backupDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "should have created a backup file")
}

func TestServiceFactory_Create_WithInvalidPath_ReturnsError(t *testing.T) {
	ctx := context.Background()

	factory := NewServiceFactory()
	_, _, err := factory.Create(ctx, "/nonexistent/path/to/db.db")
	assert.Error(t, err)
}

func TestServiceFactory_Create_Cleanup_ClosesDatabase(t *testing.T) {
	ctx := context.Background()

	factory := NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	require.NotNil(t, services)

	cleanup()
}
