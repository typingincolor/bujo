package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupListService(t *testing.T) *ListService {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	listRepo := sqlite.NewListRepository(db)
	listItemRepo := sqlite.NewListItemRepository(db)
	return NewListService(listRepo, listItemRepo)
}

func TestListService_CreateList(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")

	require.NoError(t, err)
	assert.Equal(t, "Shopping", list.Name)
	assert.Greater(t, list.ID, int64(0))
}

func TestListService_CreateList_WithSpaces(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping List")

	require.NoError(t, err)
	assert.Equal(t, "Shopping List", list.Name)
}

func TestListService_CreateList_EmptyName(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	_, err := svc.CreateList(ctx, "")

	require.Error(t, err)
}

func TestListService_GetListByID(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	created, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	found, err := svc.GetListByID(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Shopping", found.Name)
}

func TestListService_GetListByID_NotFound(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	_, err := svc.GetListByID(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list not found")
}

func TestListService_GetListByName(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	created, err := svc.CreateList(ctx, "Shopping List")
	require.NoError(t, err)

	found, err := svc.GetListByName(ctx, "Shopping List")

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestListService_GetListByName_NotFound(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	_, err := svc.GetListByName(ctx, "Nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list not found")
}

func TestListService_GetAllLists(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	_, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)
	_, err = svc.CreateList(ctx, "Work")
	require.NoError(t, err)

	lists, err := svc.GetAllLists(ctx)

	require.NoError(t, err)
	assert.Len(t, lists, 2)
}

func TestListService_RenameList(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	err = svc.RenameList(ctx, list.ID, "Groceries")
	require.NoError(t, err)

	found, err := svc.GetListByID(ctx, list.ID)
	require.NoError(t, err)
	assert.Equal(t, "Groceries", found.Name)
}

func TestListService_RenameList_NotFound(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	err := svc.RenameList(ctx, 99999, "New Name")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list not found")
}

func TestListService_DeleteList_Empty(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	err = svc.DeleteList(ctx, list.ID, false)
	require.NoError(t, err)

	_, err = svc.GetListByID(ctx, list.ID)
	require.Error(t, err)
}

func TestListService_DeleteList_WithItems_NoForce(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.DeleteList(ctx, list.ID, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "has items")
}

func TestListService_DeleteList_WithItems_Force(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.DeleteList(ctx, list.ID, true)
	require.NoError(t, err)

	_, err = svc.GetListByID(ctx, list.ID)
	require.Error(t, err)
}

func TestListService_AddItem(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	id, err := svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestListService_AddItem_ListNotFound(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	_, err := svc.AddItem(ctx, 99999, domain.EntryTypeTask, "Milk")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list not found")
}

func TestListService_AddItem_RejectsNotes(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.AddItem(ctx, list.ID, domain.EntryTypeNote, "This is a note")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only tasks can be added to lists")
}

func TestListService_AddItem_RejectsEvents(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.AddItem(ctx, list.ID, domain.EntryTypeEvent, "Meeting at 3pm")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only tasks can be added to lists")
}

func TestListService_GetListItems(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)
	_, err = svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Bread")
	require.NoError(t, err)

	items, err := svc.GetListItems(ctx, list.ID)

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Milk", items[0].Content)
	assert.Equal(t, "Bread", items[1].Content)
}

func TestListService_RemoveItem(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	itemID, err := svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.RemoveItem(ctx, itemID)
	require.NoError(t, err)

	items, err := svc.GetListItems(ctx, list.ID)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestListService_RemoveItem_NotFound(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	err := svc.RemoveItem(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "item not found")
}

func TestListService_MarkDone(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	itemID, err := svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.MarkDone(ctx, itemID)
	require.NoError(t, err)

	items, err := svc.GetListItems(ctx, list.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ListItemTypeDone, items[0].Type)
}

func TestListService_MarkUndone(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	itemID, err := svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.MarkDone(ctx, itemID)
	require.NoError(t, err)

	err = svc.MarkUndone(ctx, itemID)
	require.NoError(t, err)

	items, err := svc.GetListItems(ctx, list.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ListItemTypeTask, items[0].Type)
}

func TestListService_MoveItem(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list1, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)
	list2, err := svc.CreateList(ctx, "Work")
	require.NoError(t, err)

	itemID, err := svc.AddItem(ctx, list1.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.MoveItem(ctx, itemID, list2.ID)
	require.NoError(t, err)

	items1, err := svc.GetListItems(ctx, list1.ID)
	require.NoError(t, err)
	assert.Len(t, items1, 0)

	items2, err := svc.GetListItems(ctx, list2.ID)
	require.NoError(t, err)
	assert.Len(t, items2, 1)
	assert.Equal(t, "Milk", items2[0].Content)
}

func TestListService_GetListSummary(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	id1, err := svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)
	_, err = svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Bread")
	require.NoError(t, err)

	err = svc.MarkDone(ctx, id1)
	require.NoError(t, err)

	summary, err := svc.GetListSummary(ctx, list.ID)

	require.NoError(t, err)
	assert.Equal(t, "Shopping", summary.Name)
	assert.Equal(t, 2, summary.TotalItems)
	assert.Equal(t, 1, summary.DoneItems)
}

func TestListService_EditItem(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	itemID, err := svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.EditItem(ctx, itemID, "Oat Milk")
	require.NoError(t, err)

	items, err := svc.GetListItems(ctx, list.ID)
	require.NoError(t, err)
	assert.Equal(t, "Oat Milk", items[0].Content)
}

func TestListService_EditItem_NotFound(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	err := svc.EditItem(ctx, 99999, "New Content")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "item not found")
}

func TestListService_EditItem_PreservesType(t *testing.T) {
	svc := setupListService(t)
	ctx := context.Background()

	list, err := svc.CreateList(ctx, "Shopping")
	require.NoError(t, err)

	itemID, err := svc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Milk")
	require.NoError(t, err)

	err = svc.MarkDone(ctx, itemID)
	require.NoError(t, err)

	err = svc.EditItem(ctx, itemID, "Oat Milk")
	require.NoError(t, err)

	items, err := svc.GetListItems(ctx, list.ID)
	require.NoError(t, err)
	assert.Equal(t, "Oat Milk", items[0].Content)
	assert.Equal(t, domain.ListItemTypeDone, items[0].Type)
}
