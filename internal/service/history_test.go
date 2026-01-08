package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupHistoryService(t *testing.T) (*HistoryService, *sqlite.ListItemRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	listItemRepo := sqlite.NewListItemRepository(db)
	return NewHistoryService(listItemRepo), listItemRepo
}

func TestHistoryService_GetItemHistory_ReturnsAllVersions(t *testing.T) {
	svc, repo := setupHistoryService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Version 1")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Version 2"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	loaded, err = repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Version 3"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	history, err := svc.GetItemHistory(ctx, item.EntityID)

	require.NoError(t, err)
	assert.Len(t, history, 3)
	assert.Equal(t, "Version 1", history[0].Content)
	assert.Equal(t, "Version 2", history[1].Content)
	assert.Equal(t, "Version 3", history[2].Content)
}

func TestHistoryService_GetItemHistory_ReturnsEmptyForUnknownEntity(t *testing.T) {
	svc, _ := setupHistoryService(t)
	ctx := context.Background()

	unknownEntityID := domain.NewEntityID()
	history, err := svc.GetItemHistory(ctx, unknownEntityID)

	require.NoError(t, err)
	assert.Len(t, history, 0)
}

func TestHistoryService_GetItemAtVersion_ReturnsCorrectVersion(t *testing.T) {
	svc, repo := setupHistoryService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Version 1")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Version 2"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	result, err := svc.GetItemAtVersion(ctx, item.EntityID, 1)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Version 1", result.Content)
	assert.Equal(t, 1, result.Version)
}

func TestHistoryService_GetItemAtVersion_ReturnsNilForInvalidVersion(t *testing.T) {
	svc, repo := setupHistoryService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Version 1")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	result, err := svc.GetItemAtVersion(ctx, item.EntityID, 99)

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHistoryService_RestoreItem_CreatesNewVersionWithOldContent(t *testing.T) {
	svc, repo := setupHistoryService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Original content")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	loaded, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	loaded.Content = "Changed content"
	err = repo.Update(ctx, *loaded)
	require.NoError(t, err)

	err = svc.RestoreItem(ctx, item.EntityID, 1)
	require.NoError(t, err)

	current, err := repo.GetByEntityID(ctx, item.EntityID)
	require.NoError(t, err)
	assert.Equal(t, "Original content", current.Content)
	assert.Equal(t, 3, current.Version)
}

func TestHistoryService_RestoreItem_FailsForInvalidVersion(t *testing.T) {
	svc, repo := setupHistoryService(t)
	ctx := context.Background()

	listEntityID := domain.NewEntityID()
	item := domain.NewListItem(listEntityID, domain.ListItemTypeTask, "Content")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	err = svc.RestoreItem(ctx, item.EntityID, 99)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "version not found")
}
