package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupArchiveService(t *testing.T) (*ArchiveService, *sqlite.ListItemRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	listItemRepo := sqlite.NewListItemRepository(db)
	return NewArchiveService(listItemRepo), listItemRepo
}

func TestArchiveService_GetArchivableCount_ReturnsOldVersionCount(t *testing.T) {
	svc, repo := setupArchiveService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Test item")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Updated content"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	loaded, err = repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Updated again"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	cutoff := time.Now().Add(time.Hour)
	count, err := svc.GetArchivableCount(ctx, cutoff)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestArchiveService_GetArchivableCount_ExcludesCurrentVersions(t *testing.T) {
	svc, repo := setupArchiveService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Test item")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	cutoff := time.Now().Add(time.Hour)
	count, err := svc.GetArchivableCount(ctx, cutoff)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestArchiveService_GetArchivableCount_ExcludesRecentVersions(t *testing.T) {
	svc, repo := setupArchiveService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Test item")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Updated content"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	cutoff := time.Now().Add(-time.Hour)
	count, err := svc.GetArchivableCount(ctx, cutoff)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestArchiveService_Archive_DeletesOldVersions(t *testing.T) {
	svc, repo := setupArchiveService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Test item")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Updated content"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	cutoff := time.Now().Add(time.Hour)
	deleted, err := svc.Archive(ctx, cutoff)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	history, err := repo.GetHistory(ctx, item.EntityID)
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestArchiveService_Archive_PreservesCurrentVersion(t *testing.T) {
	svc, repo := setupArchiveService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Test item")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Current content"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	cutoff := time.Now().Add(time.Hour)
	_, err = svc.Archive(ctx, cutoff)
	require.NoError(t, err)

	current, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	assert.Equal(t, "Current content", current.Content)
}

func TestArchiveService_DryRun_DoesNotDelete(t *testing.T) {
	svc, repo := setupArchiveService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Test item")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Updated content"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	cutoff := time.Now().Add(time.Hour)
	count, err := svc.DryRun(ctx, cutoff)

	require.NoError(t, err)
	assert.Equal(t, 1, count)

	history, err := repo.GetHistory(ctx, item.EntityID)
	require.NoError(t, err)
	assert.Len(t, history, 2)
}
