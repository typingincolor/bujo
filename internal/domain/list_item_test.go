package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListItemType_Task_IsValid(t *testing.T) {
	assert.True(t, ListItemTypeTask.IsValid())
}

func TestListItemType_Done_IsValid(t *testing.T) {
	assert.True(t, ListItemTypeDone.IsValid())
}

func TestListItemType_Note_IsNotValid(t *testing.T) {
	invalid := ListItemType("note")
	assert.False(t, invalid.IsValid())
}

func TestListItemType_Event_IsNotValid(t *testing.T) {
	invalid := ListItemType("event")
	assert.False(t, invalid.IsValid())
}

func TestListItemType_Symbol_ReturnsCorrect(t *testing.T) {
	assert.Equal(t, ".", ListItemTypeTask.Symbol())
	assert.Equal(t, "x", ListItemTypeDone.Symbol())
}

func TestListItem_Validate_ValidItem_Succeeds(t *testing.T) {
	item := ListItem{
		ListEntityID: NewEntityID(),
		Type:         ListItemTypeTask,
		Content:      "Buy milk",
	}

	err := item.Validate()

	assert.NoError(t, err)
}

func TestListItem_Validate_EmptyContent_Fails(t *testing.T) {
	item := ListItem{
		ListEntityID: NewEntityID(),
		Type:         ListItemTypeTask,
		Content:      "",
	}

	err := item.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content")
}

func TestListItem_Validate_EmptyListEntityID_Fails(t *testing.T) {
	item := ListItem{
		Type:    ListItemTypeTask,
		Content: "Buy milk",
	}

	err := item.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list entity ID")
}

func TestListItem_Validate_InvalidType_Fails(t *testing.T) {
	item := ListItem{
		ListEntityID: NewEntityID(),
		Type:         ListItemType("note"),
		Content:      "Buy milk",
	}

	err := item.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type")
}

func TestListItem_IsComplete_WhenDone_ReturnsTrue(t *testing.T) {
	item := ListItem{Type: ListItemTypeDone}

	assert.True(t, item.IsComplete())
}

func TestListItem_IsComplete_WhenTask_ReturnsFalse(t *testing.T) {
	item := ListItem{Type: ListItemTypeTask}

	assert.False(t, item.IsComplete())
}

func TestNewListItem_SetsEntityIDAndCreatedAt(t *testing.T) {
	listID := NewEntityID()

	item := NewListItem(listID, ListItemTypeTask, "Buy milk")

	assert.False(t, item.EntityID.IsEmpty())
	assert.Equal(t, listID, item.ListEntityID)
	assert.Equal(t, ListItemTypeTask, item.Type)
	assert.Equal(t, "Buy milk", item.Content)
	assert.False(t, item.CreatedAt.IsZero())
	assert.Equal(t, 1, item.Version)
	assert.Equal(t, OpTypeInsert, item.OpType)
}

func TestListItemType_Cancelled_IsValid(t *testing.T) {
	assert.True(t, ListItemTypeCancelled.IsValid())
}

func TestListItemType_Cancelled_Symbol_ReturnsX(t *testing.T) {
	assert.Equal(t, "X", ListItemTypeCancelled.Symbol())
}

func TestListItem_IsCancelled_WhenCancelled_ReturnsTrue(t *testing.T) {
	item := ListItem{Type: ListItemTypeCancelled}

	assert.True(t, item.IsCancelled())
}

func TestListItem_IsCancelled_WhenTask_ReturnsFalse(t *testing.T) {
	item := ListItem{Type: ListItemTypeTask}

	assert.False(t, item.IsCancelled())
}

func TestListItem_IsCancelled_WhenDone_ReturnsFalse(t *testing.T) {
	item := ListItem{Type: ListItemTypeDone}

	assert.False(t, item.IsCancelled())
}
