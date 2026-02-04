package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestBujoService_DeleteEntry(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Task to delete", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.DeleteEntry(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Nil(t, entry)
}

func TestBujoService_DeleteEntry_WithChildren_Cascade(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `. Parent
  - Child 1
  - Child 2`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 3)

	// Delete parent - should delete children too
	err = service.DeleteEntry(ctx, ids[0])
	require.NoError(t, err)

	// All entries should be gone
	for _, id := range ids {
		entry, err := entryRepo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, entry)
	}
}

func TestBujoService_DeleteEntryAndReparent(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `. Grandparent
  - Parent
    . Grandchild`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 3)

	// Delete parent, reparent grandchild to grandparent
	err = service.DeleteEntryAndReparent(ctx, ids[1])
	require.NoError(t, err)

	// Parent should be gone
	parent, err := entryRepo.GetByID(ctx, ids[1])
	require.NoError(t, err)
	assert.Nil(t, parent)

	// Grandchild should now have grandparent as parent
	grandchild, err := entryRepo.GetByID(ctx, ids[2])
	require.NoError(t, err)
	require.NotNil(t, grandchild)
	require.NotNil(t, grandchild.ParentID)
	assert.Equal(t, ids[0], *grandchild.ParentID)
}

func TestBujoService_HasChildren(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `. Parent
  - Child`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	hasChildren, err := service.HasChildren(ctx, ids[0])
	require.NoError(t, err)
	assert.True(t, hasChildren)

	hasChildren, err = service.HasChildren(ctx, ids[1])
	require.NoError(t, err)
	assert.False(t, hasChildren)
}

func TestBujoService_DeleteEntry_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.DeleteEntry(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_MigrateEntry(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	ids, err := service.LogEntries(ctx, ". Call dentist", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	newID, err := service.MigrateEntry(ctx, ids[0], tomorrow)
	require.NoError(t, err)

	// Old entry should be marked as migrated
	oldEntry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeMigrated, oldEntry.Type)

	// New entry should be a task on tomorrow
	newEntry, err := entryRepo.GetByID(ctx, newID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, newEntry.Type)
	assert.Equal(t, "Call dentist", newEntry.Content)
	assert.Equal(t, tomorrow.Format("2006-01-02"), newEntry.ScheduledDate.Format("2006-01-02"))
}

func TestBujoService_MigrateEntry_OnlyTasks(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	// Try to migrate a note
	ids, err := service.LogEntries(ctx, "- This is a note", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	_, err = service.MigrateEntry(ctx, ids[0], tomorrow)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only tasks")
}

func TestBujoService_MigrateEntry_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	_, err := service.MigrateEntry(ctx, 99999, time.Now())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_MigrateEntry_WithChildren(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	// Create parent with children
	ids, err := service.LogEntries(ctx, `. Parent task
  - Child note
  . Child task`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 3)

	parentID := ids[0]
	childNoteID := ids[1]
	childTaskID := ids[2]

	// Migrate parent
	newParentID, err := service.MigrateEntry(ctx, parentID, tomorrow)
	require.NoError(t, err)

	// Old parent should be marked as migrated
	oldParent, err := entryRepo.GetByID(ctx, parentID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeMigrated, oldParent.Type)

	// Old children should also be marked as migrated
	oldChildNote, err := entryRepo.GetByID(ctx, childNoteID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeMigrated, oldChildNote.Type)

	oldChildTask, err := entryRepo.GetByID(ctx, childTaskID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeMigrated, oldChildTask.Type)

	// New parent should exist on tomorrow
	newParent, err := entryRepo.GetByID(ctx, newParentID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, newParent.Type)
	assert.Equal(t, "Parent task", newParent.Content)

	// New children should exist and be linked to new parent
	children, err := entryRepo.GetChildren(ctx, newParentID)
	require.NoError(t, err)
	assert.Len(t, children, 2)

	// Verify children types preserved
	childTypes := make(map[string]domain.EntryType)
	for _, c := range children {
		childTypes[c.Content] = c.Type
	}
	assert.Equal(t, domain.EntryTypeNote, childTypes["Child note"])
	assert.Equal(t, domain.EntryTypeTask, childTypes["Child task"])
}

func TestBujoService_MigrateEntry_PreservesChildDepth(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	// Create parent with child at depth 1
	ids, err := service.LogEntries(ctx, `. Parent task
  - Child note`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 2)

	parentID := ids[0]
	childNoteID := ids[1]

	// Verify original child has depth 1
	originalChild, err := entryRepo.GetByID(ctx, childNoteID)
	require.NoError(t, err)
	assert.Equal(t, 1, originalChild.Depth, "Original child should have depth 1")

	// Migrate parent
	newParentID, err := service.MigrateEntry(ctx, parentID, tomorrow)
	require.NoError(t, err)

	// Get new children
	newChildren, err := entryRepo.GetChildren(ctx, newParentID)
	require.NoError(t, err)
	require.Len(t, newChildren, 1)

	// Verify new child has depth 1 (this will FAIL if bug exists)
	newChild := newChildren[0]
	assert.Equal(t, "Child note", newChild.Content)
	assert.Equal(t, 1, newChild.Depth, "Migrated child must preserve depth 1, not default to 0")
}

func TestBujoService_MigrateEntry_WithGrandchildren(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	// Create parent → child → grandchild hierarchy
	ids, err := service.LogEntries(ctx, `. Parent task
  - Child note
    - Grandchild note`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 3)

	parentID := ids[0]
	childID := ids[1]
	grandchildID := ids[2]

	// Verify original structure
	originalChild, err := entryRepo.GetByID(ctx, childID)
	require.NoError(t, err)
	assert.Equal(t, 1, originalChild.Depth)

	originalGrandchild, err := entryRepo.GetByID(ctx, grandchildID)
	require.NoError(t, err)
	assert.Equal(t, 2, originalGrandchild.Depth)

	// Migrate parent
	newParentID, err := service.MigrateEntry(ctx, parentID, tomorrow)
	require.NoError(t, err)

	// Old entries should all be marked as migrated
	oldParent, err := entryRepo.GetByID(ctx, parentID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeMigrated, oldParent.Type)

	oldChild, err := entryRepo.GetByID(ctx, childID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeMigrated, oldChild.Type)

	oldGrandchild, err := entryRepo.GetByID(ctx, grandchildID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeMigrated, oldGrandchild.Type)

	// New parent should exist on tomorrow
	newParent, err := entryRepo.GetByID(ctx, newParentID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, newParent.Type)
	assert.Equal(t, "Parent task", newParent.Content)

	// Get full tree under new parent
	newTree, err := entryRepo.GetWithChildren(ctx, newParentID)
	require.NoError(t, err)
	// Tree includes parent + child + grandchild = 3
	assert.Len(t, newTree, 3, "migrated tree should include parent, child, and grandchild")

	// Build a map by content for easier assertions
	byContent := make(map[string]domain.Entry)
	for _, e := range newTree {
		byContent[e.Content] = e
	}

	// Verify child exists with correct depth and parent
	newChild := byContent["Child note"]
	assert.Equal(t, domain.EntryTypeNote, newChild.Type)
	assert.Equal(t, 1, newChild.Depth)
	require.NotNil(t, newChild.ParentID)
	assert.Equal(t, newParentID, *newChild.ParentID)

	// Verify grandchild exists with correct depth and parent
	newGrandchild := byContent["Grandchild note"]
	assert.Equal(t, domain.EntryTypeNote, newGrandchild.Type)
	assert.Equal(t, 2, newGrandchild.Depth)
	require.NotNil(t, newGrandchild.ParentID)
	assert.Equal(t, newChild.ID, *newGrandchild.ParentID, "grandchild should be parented to new child, not new root")

	// Verify all new entries are scheduled on tomorrow
	for _, e := range newTree {
		require.NotNil(t, e.ScheduledDate)
		assert.Equal(t, tomorrow.Format("2006-01-02"), e.ScheduledDate.Format("2006-01-02"),
			"entry %q should be scheduled on target date", e.Content)
	}
}

func TestBujoService_MoveEntry_ChangeParent(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	// Create two separate entries
	ids1, err := service.LogEntries(ctx, `. Parent A
  - Child of A`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	ids2, err := service.LogEntries(ctx, ". Parent B", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Move "Child of A" to be under "Parent B"
	err = service.MoveEntry(ctx, ids1[1], MoveOptions{NewParentID: &ids2[0]})
	require.NoError(t, err)

	// Verify the child now has Parent B as parent
	child, err := entryRepo.GetByID(ctx, ids1[1])
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, ids2[0], *child.ParentID)
}

func TestBujoService_MoveEntry_ChangeLoggedDate(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	ids, err := service.LogEntries(ctx, ". Task logged today", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Move to yesterday
	err = service.MoveEntry(ctx, ids[0], MoveOptions{NewLoggedDate: &yesterday})
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, yesterday.Format("2006-01-02"), entry.ScheduledDate.Format("2006-01-02"))
}

func TestBujoService_MoveEntry_ChangeLoggedDateMovesChildren(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	monday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	ids, err := service.LogEntries(ctx, `- Parent with children
  - Child note
  . Grandchild task`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 3)

	// Move parent to Monday - children should follow
	err = service.MoveEntry(ctx, ids[0], MoveOptions{NewLoggedDate: &monday})
	require.NoError(t, err)

	// Parent should be on Monday
	parent, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, monday.Format("2006-01-02"), parent.ScheduledDate.Format("2006-01-02"))

	// Child should also be on Monday
	child, err := entryRepo.GetByID(ctx, ids[1])
	require.NoError(t, err)
	assert.Equal(t, monday.Format("2006-01-02"), child.ScheduledDate.Format("2006-01-02"))

	// Grandchild should also be on Monday
	grandchild, err := entryRepo.GetByID(ctx, ids[2])
	require.NoError(t, err)
	assert.Equal(t, monday.Format("2006-01-02"), grandchild.ScheduledDate.Format("2006-01-02"))
}

func TestBujoService_MoveEntry_MoveToRoot(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	ids, err := service.LogEntries(ctx, `. Parent
  - Child to become root`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Move child to root (no parent)
	moveToRoot := true
	err = service.MoveEntry(ctx, ids[1], MoveOptions{MoveToRoot: &moveToRoot})
	require.NoError(t, err)

	child, err := entryRepo.GetByID(ctx, ids[1])
	require.NoError(t, err)
	assert.Nil(t, child.ParentID)
	assert.Equal(t, 0, child.Depth)
}

func TestBujoService_MoveEntry_WithChildren(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	ids, err := service.LogEntries(ctx, `. Parent A
  - Child
    . Grandchild`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	ids2, err := service.LogEntries(ctx, ". Parent B", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Move "Child" (with Grandchild) under "Parent B"
	err = service.MoveEntry(ctx, ids[1], MoveOptions{NewParentID: &ids2[0]})
	require.NoError(t, err)

	// Child should be under Parent B with depth 1
	child, err := entryRepo.GetByID(ctx, ids[1])
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, ids2[0], *child.ParentID)
	assert.Equal(t, 1, child.Depth)

	// Grandchild should have updated depth (2)
	grandchild, err := entryRepo.GetByID(ctx, ids[2])
	require.NoError(t, err)
	assert.Equal(t, 2, grandchild.Depth)
}

func TestBujoService_MoveEntry_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.MoveEntry(ctx, 99999, MoveOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_MoveEntry_ParentNotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	invalidParent := int64(99999)
	err = service.MoveEntry(ctx, ids[0], MoveOptions{NewParentID: &invalidParent})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent")
}
