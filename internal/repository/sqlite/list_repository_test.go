package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestListRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list, err := repo.Create(ctx, "Shopping")

	require.NoError(t, err)
	assert.Equal(t, "Shopping", list.Name)
	assert.Greater(t, list.ID, int64(0))
	assert.False(t, list.CreatedAt.IsZero())
}

func TestListRepository_Create_WithSpaces(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list, err := repo.Create(ctx, "Shopping List")

	require.NoError(t, err)
	assert.Equal(t, "Shopping List", list.Name)
}

func TestListRepository_Create_DuplicateName_AllowedWithEventSourcing(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list1, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)

	list2, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)

	// With event sourcing, duplicate names are allowed (different entity IDs)
	assert.NotEqual(t, list1.EntityID, list2.EntityID)
	assert.Equal(t, list1.Name, list2.Name)
}

func TestListRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, created.ID)

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Shopping", found.Name)
}

func TestListRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, 99999)

	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestListRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, "Shopping List")
	require.NoError(t, err)

	found, err := repo.GetByName(ctx, "Shopping List")

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
}

func TestListRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	found, err := repo.GetByName(ctx, "Nonexistent")

	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestListRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)
	_, err = repo.Create(ctx, "Work")
	require.NoError(t, err)

	lists, err := repo.GetAll(ctx)

	require.NoError(t, err)
	assert.Len(t, lists, 2)
}

func TestListRepository_GetAll_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	lists, err := repo.GetAll(ctx)

	require.NoError(t, err)
	assert.Empty(t, lists)
}

func TestListRepository_Rename(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)

	err = repo.Rename(ctx, list.ID, "Groceries")
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, list.ID)
	require.NoError(t, err)
	assert.Equal(t, "Groceries", found.Name)
}

func TestListRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)

	err = repo.Delete(ctx, list.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, list.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestListRepository_GetItemCount(t *testing.T) {
	db := setupTestDB(t)
	listRepo := NewListRepository(db)
	listItemRepo := NewListItemRepository(db)
	ctx := context.Background()

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	// Add items to the list
	item1 := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Milk")
	_, err = listItemRepo.Insert(ctx, item1)
	require.NoError(t, err)

	item2 := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Bread")
	_, err = listItemRepo.Insert(ctx, item2)
	require.NoError(t, err)

	count, err := listRepo.GetItemCount(ctx, list.ID)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestListRepository_GetDoneCount(t *testing.T) {
	db := setupTestDB(t)
	listRepo := NewListRepository(db)
	listItemRepo := NewListItemRepository(db)
	ctx := context.Background()

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	// Add items - one done, one not
	item1 := domain.NewListItem(list.EntityID, domain.ListItemTypeDone, "Milk")
	_, err = listItemRepo.Insert(ctx, item1)
	require.NoError(t, err)

	item2 := domain.NewListItem(list.EntityID, domain.ListItemTypeTask, "Bread")
	_, err = listItemRepo.Insert(ctx, item2)
	require.NoError(t, err)

	count, err := listRepo.GetDoneCount(ctx, list.ID)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestListRepository_GetByEntityID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)

	found, err := repo.GetByEntityID(ctx, created.EntityID)

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.EntityID, found.EntityID)
	assert.Equal(t, "Shopping", found.Name)
}

func TestListRepository_GetByEntityID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	found, err := repo.GetByEntityID(ctx, domain.NewEntityID())

	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestListRepository_Delete_SoftDeletes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list, err := repo.Create(ctx, "SoftDeleteTest")
	require.NoError(t, err)

	err = repo.Delete(ctx, list.ID)
	require.NoError(t, err)

	// Verify list is soft deleted (not visible via GetByID)
	found, err := repo.GetByID(ctx, list.ID)
	require.NoError(t, err)
	assert.Nil(t, found)

	// But data should still exist in database
	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM lists WHERE name = 'SoftDeleteTest'`).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1, "Soft deleted list should still exist in DB")
}

func TestListRepository_GetDeleted_ReturnsDeletedLists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list, err := repo.Create(ctx, "DeletedList")
	require.NoError(t, err)

	err = repo.Delete(ctx, list.ID)
	require.NoError(t, err)

	deleted, err := repo.GetDeleted(ctx)
	require.NoError(t, err)
	require.Len(t, deleted, 1)
	assert.Equal(t, "DeletedList", deleted[0].Name)
}

func TestListRepository_Restore_BringsBackDeletedList(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	list, err := repo.Create(ctx, "RestoreListTest")
	require.NoError(t, err)
	entityID := list.EntityID

	err = repo.Delete(ctx, list.ID)
	require.NoError(t, err)

	// Verify it's gone
	found, err := repo.GetByID(ctx, list.ID)
	require.NoError(t, err)
	assert.Nil(t, found)

	// Restore it
	newID, err := repo.Restore(ctx, entityID)
	require.NoError(t, err)
	assert.NotZero(t, newID)

	// Verify it's back
	restored, err := repo.GetByID(ctx, newID)
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, "RestoreListTest", restored.Name)
}
