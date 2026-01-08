package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func setupListItemRepo(t *testing.T) (*ListItemRepository, *ListRepository) {
	t.Helper()
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewListItemRepository(db), NewListRepository(db)
}

func createTestList(t *testing.T, listRepo *ListRepository) *domain.List {
	t.Helper()
	list, err := listRepo.Create(context.Background(), "Test List")
	require.NoError(t, err)
	return list
}

func TestListItemRepository_Insert_Success(t *testing.T) {
	repo, listRepo := setupListItemRepo(t)
	ctx := context.Background()
	list := createTestList(t, listRepo)

	item := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy milk")

	id, err := repo.Insert(ctx, item)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestListItemRepository_GetByID_Found(t *testing.T) {
	repo, listRepo := setupListItemRepo(t)
	ctx := context.Background()
	list := createTestList(t, listRepo)

	item := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy milk")
	id, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, id, found.RowID)
	assert.Equal(t, "Buy milk", found.Content)
	assert.Equal(t, domain.ListItemTypeTask, found.Type)
}

func TestListItemRepository_GetByID_NotFound(t *testing.T) {
	repo, _ := setupListItemRepo(t)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, 99999)

	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestListItemRepository_GetByEntityID_Found(t *testing.T) {
	repo, listRepo := setupListItemRepo(t)
	ctx := context.Background()
	list := createTestList(t, listRepo)

	item := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy milk")
	_, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	found, err := repo.GetByEntityID(ctx, item.EntityID)

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, item.EntityID, found.EntityID)
	assert.Equal(t, "Buy milk", found.Content)
}

func TestListItemRepository_GetByListEntityID_ReturnsItems(t *testing.T) {
	repo, listRepo := setupListItemRepo(t)
	ctx := context.Background()
	list := createTestList(t, listRepo)

	item1 := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy milk")
	item2 := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy bread")
	_, err := repo.Insert(ctx, item1)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, item2)
	require.NoError(t, err)

	items, err := repo.GetByListEntityID(ctx, list.EntityID)

	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestListItemRepository_Update_Success(t *testing.T) {
	repo, listRepo := setupListItemRepo(t)
	ctx := context.Background()
	list := createTestList(t, listRepo)

	item := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy milk")
	id, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	found.Type = domain.ListItemTypeDone
	err = repo.Update(ctx, *found)
	require.NoError(t, err)

	updated, err := repo.GetByEntityID(ctx, found.EntityID)
	require.NoError(t, err)
	assert.Equal(t, domain.ListItemTypeDone, updated.Type)
}

func TestListItemRepository_Delete_SoftDeletes(t *testing.T) {
	repo, listRepo := setupListItemRepo(t)
	ctx := context.Background()
	list := createTestList(t, listRepo)

	item := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy milk")
	id, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Should not be found via GetByID (current state)
	found, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, found)

	// But history should show the delete
	history, err := repo.GetHistory(ctx, item.EntityID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 1)
}

func TestListItemRepository_GetHistory_ReturnsAllVersions(t *testing.T) {
	repo, listRepo := setupListItemRepo(t)
	ctx := context.Background()
	list := createTestList(t, listRepo)

	item := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Buy milk")
	id, err := repo.Insert(ctx, item)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	found.Content = "Buy almond milk"
	err = repo.Update(ctx, *found)
	require.NoError(t, err)

	history, err := repo.GetHistory(ctx, item.EntityID)

	require.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, 1, history[0].Version)
	assert.Equal(t, 2, history[1].Version)
}
