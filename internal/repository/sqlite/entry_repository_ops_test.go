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

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	inserted.Type = domain.EntryTypeDone
	inserted.Content = "Updated content"

	err = repo.Update(ctx, *inserted)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
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

	parentEntry, err := repo.GetByID(ctx, parentID)
	require.NoError(t, err)
	parentEntry.Content = "Updated parent meeting"
	err = repo.Update(ctx, *parentEntry)
	require.NoError(t, err)

	updatedParent, err := repo.GetByID(ctx, parentID)
	require.NoError(t, err)
	require.NotNil(t, updatedParent)
	assert.Equal(t, "Updated parent meeting", updatedParent.Content)

	children, err := repo.GetChildren(ctx, parentID)
	require.NoError(t, err)
	assert.Len(t, children, 2)

	tree, err := repo.GetWithChildren(ctx, parentID)
	require.NoError(t, err)
	assert.Len(t, tree, 4)
}

func TestEntryRepository_Delete_RemovesEntry(t *testing.T) {
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

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEntryRepository_DeleteWithChildren_RemovesAll(t *testing.T) {
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
	grandchildID, err := repo.Insert(ctx, grandchild)
	require.NoError(t, err)

	err = repo.DeleteWithChildren(ctx, parentID)
	require.NoError(t, err)

	parentResult, _ := repo.GetByID(ctx, parentID)
	assert.Nil(t, parentResult)

	childResult, _ := repo.GetByID(ctx, childID)
	assert.Nil(t, childResult)

	grandchildResult, _ := repo.GetByID(ctx, grandchildID)
	assert.Nil(t, grandchildResult)
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
	assert.False(t, result.EntityID.IsEmpty())
}
