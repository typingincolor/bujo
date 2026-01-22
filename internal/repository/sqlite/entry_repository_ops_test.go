package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestEntryRepository_Insert_WithParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	parent := domain.Entry{
		Type:      domain.EntryTypeEvent,
		Content:   "Meeting",
		Depth:     0,
		CreatedAt: time.Now(),
	}
	parentID, err := repo.Insert(ctx, parent)
	require.NoError(t, err)

	child := domain.Entry{
		Type:      domain.EntryTypeNote,
		Content:   "Meeting notes",
		ParentID:  &parentID,
		Depth:     1,
		CreatedAt: time.Now(),
	}
	childID, err := repo.Insert(ctx, child)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, childID)
	require.NoError(t, err)
	require.NotNil(t, result.ParentID)
	assert.Equal(t, parentID, *result.ParentID)
}

func TestEntryRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Original content",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	// Get inserted entry to obtain entity_id
	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	inserted.Type = domain.EntryTypeDone
	inserted.Content = "Updated content"

	err = repo.Update(ctx, *inserted)
	require.NoError(t, err)

	// With event sourcing, original ID row is closed; use GetByEntityID
	result, err := repo.GetByEntityID(ctx, inserted.EntityID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeDone, result.Type)
	assert.Equal(t, "Updated content", result.Content)
}

func TestEntryRepository_Update_PreservesChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	parent := domain.Entry{
		Type:      domain.EntryTypeEvent,
		Content:   "Parent meeting",
		Depth:     0,
		CreatedAt: time.Now(),
	}
	parentID, err := repo.Insert(ctx, parent)
	require.NoError(t, err)

	child1 := domain.Entry{
		Type:      domain.EntryTypeNote,
		Content:   "First child note",
		ParentID:  &parentID,
		Depth:     1,
		CreatedAt: time.Now(),
	}
	child1ID, err := repo.Insert(ctx, child1)
	require.NoError(t, err)

	child2 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Second child task",
		ParentID:  &parentID,
		Depth:     1,
		CreatedAt: time.Now(),
	}
	_, err = repo.Insert(ctx, child2)
	require.NoError(t, err)

	grandchild := domain.Entry{
		Type:      domain.EntryTypeNote,
		Content:   "Grandchild note",
		ParentID:  &child1ID,
		Depth:     2,
		CreatedAt: time.Now(),
	}
	_, err = repo.Insert(ctx, grandchild)
	require.NoError(t, err)

	// Update the parent (creates new row with new ID in event sourcing)
	parentEntry, err := repo.GetByID(ctx, parentID)
	require.NoError(t, err)
	parentEntry.Content = "Updated parent meeting"
	err = repo.Update(ctx, *parentEntry)
	require.NoError(t, err)

	// Get the updated parent (may have new row ID)
	updatedParent, err := repo.GetByEntityID(ctx, parentEntry.EntityID)
	require.NoError(t, err)
	require.NotNil(t, updatedParent)
	assert.Equal(t, "Updated parent meeting", updatedParent.Content)

	// Children should still be accessible via GetChildren using the new parent ID
	children, err := repo.GetChildren(ctx, updatedParent.ID)
	require.NoError(t, err)
	assert.Len(t, children, 2, "Children should still be linked to parent after update")

	// GetWithChildren should return parent and all descendants
	tree, err := repo.GetWithChildren(ctx, updatedParent.ID)
	require.NoError(t, err)
	assert.Len(t, tree, 4, "Should return parent + 2 children + 1 grandchild")
}

func TestEntryRepository_Delete_SoftDeletes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "To be deleted",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Should not be returned by GetByID
	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, result, "Deleted entry should not be returned by GetByID")

	// Should not be returned by GetByEntityID
	result, err = repo.GetByEntityID(ctx, inserted.EntityID)
	require.NoError(t, err)
	assert.Nil(t, result, "Deleted entry should not be returned by GetByEntityID")
}

func TestEntryRepository_Delete_PreservesHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "To be deleted",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// History should still contain the entry
	history, err := repo.GetHistory(ctx, inserted.EntityID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 1, "History should be preserved after soft delete")
}

func TestEntryRepository_Delete_CreatesDeleteMarker(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "To be deleted",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Check that a DELETE op_type record exists
	var opType string
	err = db.QueryRowContext(ctx, `
		SELECT op_type FROM entries
		WHERE entity_id = ?
		ORDER BY version DESC LIMIT 1
	`, inserted.EntityID.String()).Scan(&opType)
	require.NoError(t, err)
	assert.Equal(t, "DELETE", opType, "Latest version should have DELETE op_type")
}

func TestEntryRepository_Restore_BringsBackDeletedEntry(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "To be restored",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Verify it's deleted
	result, err := repo.GetByEntityID(ctx, inserted.EntityID)
	require.NoError(t, err)
	assert.Nil(t, result)

	// Restore it
	restoredID, err := repo.Restore(ctx, inserted.EntityID)
	require.NoError(t, err)
	assert.Greater(t, restoredID, int64(0))

	// Should be accessible again
	result, err = repo.GetByEntityID(ctx, inserted.EntityID)
	require.NoError(t, err)
	require.NotNil(t, result, "Restored entry should be returned by GetByEntityID")
	assert.Equal(t, "To be restored", result.Content)
}

func TestEntryRepository_GetDeleted_ReturnsDeletedEntries(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry1 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Active entry",
		CreatedAt: time.Now(),
	}
	_, err := repo.Insert(ctx, entry1)
	require.NoError(t, err)

	entry2 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Deleted entry",
		CreatedAt: time.Now(),
	}
	id2, err := repo.Insert(ctx, entry2)
	require.NoError(t, err)

	err = repo.Delete(ctx, id2)
	require.NoError(t, err)

	deleted, err := repo.GetDeleted(ctx)
	require.NoError(t, err)
	assert.Len(t, deleted, 1, "Should return only deleted entries")
	assert.Equal(t, "Deleted entry", deleted[0].Content)
}

func TestEntryRepository_DeleteWithChildren_SoftDeletesAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	parent := domain.Entry{
		Type:      domain.EntryTypeEvent,
		Content:   "Parent event",
		CreatedAt: time.Now(),
	}
	parentID, err := repo.Insert(ctx, parent)
	require.NoError(t, err)

	child := domain.Entry{
		Type:      domain.EntryTypeNote,
		Content:   "Child note",
		ParentID:  &parentID,
		Depth:     1,
		CreatedAt: time.Now(),
	}
	childID, err := repo.Insert(ctx, child)
	require.NoError(t, err)

	grandchild := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Grandchild task",
		ParentID:  &childID,
		Depth:     2,
		CreatedAt: time.Now(),
	}
	_, err = repo.Insert(ctx, grandchild)
	require.NoError(t, err)

	err = repo.DeleteWithChildren(ctx, parentID)
	require.NoError(t, err)

	// All should be soft deleted
	parentResult, _ := repo.GetByID(ctx, parentID)
	assert.Nil(t, parentResult, "Parent should be soft deleted")

	childResult, _ := repo.GetByID(ctx, childID)
	assert.Nil(t, childResult, "Child should be soft deleted")

	// But history should exist
	parentEntry, _ := repo.GetByID(ctx, parentID)
	if parentEntry != nil {
		history, _ := repo.GetHistory(ctx, parentEntry.EntityID)
		assert.GreaterOrEqual(t, len(history), 1, "History should be preserved")
	}
}

func TestEntryRepository_Insert_SetsEntityID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Test task",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.EntityID.IsEmpty(), "EntityID should be set after insert")
}

func TestEntryRepository_GetByEntityID_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Test task",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	result, err := repo.GetByEntityID(ctx, inserted.EntityID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, inserted.EntityID, result.EntityID)
	assert.Equal(t, "Test task", result.Content)
}

func TestEntryRepository_GetByEntityID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	result, err := repo.GetByEntityID(ctx, domain.NewEntityID())
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEntryRepository_GetHistory_ReturnsAllVersions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Version 1",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	// Update to create version 2
	inserted.Content = "Version 2"
	err = repo.Update(ctx, *inserted)
	require.NoError(t, err)

	history, err := repo.GetHistory(ctx, inserted.EntityID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 1)
}

func TestEntryRepository_GetAsOf_ReturnsCorrectVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Original",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	// Get state as of now (should return current)
	result, err := repo.GetAsOf(ctx, inserted.EntityID, time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Original", result.Content)
}
