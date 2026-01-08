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

func TestListRepository_Create_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, "Shopping")
	require.NoError(t, err)

	_, err = repo.Create(ctx, "Shopping")
	require.Error(t, err)
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
	entryRepo := NewEntryRepository(db)
	ctx := context.Background()

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	// Add entries to the list
	entry1 := domain.Entry{Type: domain.EntryTypeTask, Content: "Milk", ListID: &list.ID}
	_, err = entryRepo.Insert(ctx, entry1)
	require.NoError(t, err)

	entry2 := domain.Entry{Type: domain.EntryTypeTask, Content: "Bread", ListID: &list.ID}
	_, err = entryRepo.Insert(ctx, entry2)
	require.NoError(t, err)

	count, err := listRepo.GetItemCount(ctx, list.ID)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestListRepository_GetDoneCount(t *testing.T) {
	db := setupTestDB(t)
	listRepo := NewListRepository(db)
	entryRepo := NewEntryRepository(db)
	ctx := context.Background()

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	// Add entries - one done, one not
	entry1 := domain.Entry{Type: domain.EntryTypeDone, Content: "Milk", ListID: &list.ID}
	_, err = entryRepo.Insert(ctx, entry1)
	require.NoError(t, err)

	entry2 := domain.Entry{Type: domain.EntryTypeTask, Content: "Bread", ListID: &list.ID}
	_, err = entryRepo.Insert(ctx, entry2)
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
