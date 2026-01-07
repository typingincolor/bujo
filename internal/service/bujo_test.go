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

func setupBujoService(t *testing.T) (*BujoService, *sqlite.EntryRepository, *sqlite.DayContextRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	parser := domain.NewTreeParser()

	service := NewBujoService(entryRepo, dayCtxRepo, parser)
	return service, entryRepo, dayCtxRepo
}

func TestBujoService_LogEntries_SingleEntry(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	input := ". Buy groceries"
	opts := LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	}

	ids, err := service.LogEntries(ctx, input, opts)

	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func TestBujoService_LogEntries_MultipleEntries(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	input := `. Task one
- Note one
o Event one`
	opts := LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	}

	ids, err := service.LogEntries(ctx, input, opts)

	require.NoError(t, err)
	assert.Len(t, ids, 3)
}

func TestBujoService_LogEntries_WithLocation(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	input := ". Task with location"
	location := "Home Office"
	opts := LogEntriesOptions{
		Date:     time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
		Location: &location,
	}

	ids, err := service.LogEntries(ctx, input, opts)
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	require.NotNil(t, entry.Location)
	assert.Equal(t, "Home Office", *entry.Location)
}

func TestBujoService_LogEntries_NestedEntries(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	input := `o Meeting
  - Attendees: Alice, Bob
  . Follow up`
	opts := LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	}

	ids, err := service.LogEntries(ctx, input, opts)
	require.NoError(t, err)
	assert.Len(t, ids, 3)

	// Verify parent-child relationships
	child, err := entryRepo.GetByID(ctx, ids[1])
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, ids[0], *child.ParentID)
}

func TestBujoService_GetDailyAgenda(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	// Add overdue task
	_, err := service.LogEntries(ctx, ". Overdue task", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)

	// Add today's tasks
	_, err = service.LogEntries(ctx, `. Today's task
- Today's note`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	agenda, err := service.GetDailyAgenda(ctx, today)

	require.NoError(t, err)
	assert.Len(t, agenda.Overdue, 1)
	assert.Len(t, agenda.Today, 2)
}

func TestBujoService_SetLocation(t *testing.T) {
	service, _, dayCtxRepo := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, today, "Manchester Office")

	require.NoError(t, err)

	dayCtx, err := dayCtxRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.NotNil(t, dayCtx)
	assert.Equal(t, "Manchester Office", *dayCtx.Location)
}

func TestBujoService_GetDailyAgenda_WithLocation(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, today, "Home")
	require.NoError(t, err)

	agenda, err := service.GetDailyAgenda(ctx, today)

	require.NoError(t, err)
	require.NotNil(t, agenda.Location)
	assert.Equal(t, "Home", *agenda.Location)
}

func TestBujoService_GetLocationHistory(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, day1, "Home")
	require.NoError(t, err)
	err = service.SetLocation(ctx, day2, "Office")
	require.NoError(t, err)
	err = service.SetLocation(ctx, day3, "Client Site")
	require.NoError(t, err)

	history, err := service.GetLocationHistory(ctx, day1, day3)
	require.NoError(t, err)
	require.Len(t, history, 3)
	assert.Equal(t, "Home", *history[0].Location)
	assert.Equal(t, "Office", *history[1].Location)
	assert.Equal(t, "Client Site", *history[2].Location)

	// Test partial range
	history, err = service.GetLocationHistory(ctx, day2, day2)
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, "Office", *history[0].Location)
}

func TestBujoService_GetLocation(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	// No location set
	loc, err := service.GetLocation(ctx, today)
	require.NoError(t, err)
	assert.Nil(t, loc)

	// Set location
	err = service.SetLocation(ctx, today, "Office")
	require.NoError(t, err)

	loc, err = service.GetLocation(ctx, today)
	require.NoError(t, err)
	require.NotNil(t, loc)
	assert.Equal(t, "Office", *loc)
}

func TestBujoService_ClearLocation(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	// Set then clear
	err := service.SetLocation(ctx, today, "Office")
	require.NoError(t, err)

	err = service.ClearLocation(ctx, today)
	require.NoError(t, err)

	// Should be gone
	loc, err := service.GetLocation(ctx, today)
	require.NoError(t, err)
	assert.Nil(t, loc)
}

func TestBujoService_MarkDone(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	// Add a task
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	// Mark it done
	err = service.MarkDone(ctx, ids[0])
	require.NoError(t, err)

	// Verify it's marked as done
	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeDone, entry.Type)
}

func TestBujoService_MarkDone_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.MarkDone(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_Undo(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	// Add a task and mark it done
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	err = service.MarkDone(ctx, ids[0])
	require.NoError(t, err)

	// Undo it
	err = service.Undo(ctx, ids[0])
	require.NoError(t, err)

	// Verify it's back to task
	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, entry.Type)
}

func TestBujoService_Undo_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.Undo(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_GetEntryContext_RootEntry(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `. Parent
  - Child 1
  - Child 2`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// View the parent - should show parent and its children
	entries, err := service.GetEntryContext(ctx, ids[0], 0)

	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Equal(t, "Parent", entries[0].Content)
}

func TestBujoService_GetEntryContext_ChildEntry(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `. Parent
  - Child 1
  - Child 2`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// View a child - should show parent and all siblings
	entries, err := service.GetEntryContext(ctx, ids[1], 0)

	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Equal(t, "Parent", entries[0].Content)
}

func TestBujoService_GetEntryContext_WithAncestors(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `. Grandparent
  - Parent
    . Grandchild`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// View grandchild with default (0) - shows from parent down
	entries, err := service.GetEntryContext(ctx, ids[2], 0)
	require.NoError(t, err)
	assert.Len(t, entries, 2) // Parent + Grandchild

	// View grandchild with 1 additional ancestor level - shows from grandparent down
	entries, err = service.GetEntryContext(ctx, ids[2], 1)
	require.NoError(t, err)
	assert.Len(t, entries, 3) // All three
}

func TestBujoService_GetEntryContext_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	_, err := service.GetEntryContext(ctx, 99999, 0)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_EditEntry(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.EditEntry(ctx, ids[0], "Buy milk")
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, "Buy milk", entry.Content)
}

func TestBujoService_EditEntry_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.EditEntry(ctx, 99999, "New content")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

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

func TestBujoService_GetMultiDayAgenda_ReturnsEntriesGroupedByDate(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	_, err := service.LogEntries(ctx, ". Task on day 1", LogEntriesOptions{Date: day1})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Task on day 2", LogEntriesOptions{Date: day2})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Task on day 3", LogEntriesOptions{Date: day3})
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day3)

	require.NoError(t, err)
	require.Len(t, agenda.Days, 3)
	assert.Len(t, agenda.Days[0].Entries, 1)
	assert.Len(t, agenda.Days[1].Entries, 1)
	assert.Len(t, agenda.Days[2].Entries, 1)
}

func TestBujoService_GetMultiDayAgenda_IncludesOverdueOnFirstDay(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	oldDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	// Overdue task from before the range
	_, err := service.LogEntries(ctx, ". Overdue task", LogEntriesOptions{Date: oldDate})
	require.NoError(t, err)

	_, err = service.LogEntries(ctx, ". Task on day 1", LogEntriesOptions{Date: day1})
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day2)

	require.NoError(t, err)
	assert.Len(t, agenda.Overdue, 1)
	assert.Equal(t, "Overdue task", agenda.Overdue[0].Content)
}

func TestBujoService_GetMultiDayAgenda_IncludesLocations(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, day1, "Home")
	require.NoError(t, err)
	err = service.SetLocation(ctx, day2, "Office")
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day2)

	require.NoError(t, err)
	require.Len(t, agenda.Days, 2)
	require.NotNil(t, agenda.Days[0].Location)
	assert.Equal(t, "Home", *agenda.Days[0].Location)
	require.NotNil(t, agenda.Days[1].Location)
	assert.Equal(t, "Office", *agenda.Days[1].Location)
}

func TestBujoService_GetMultiDayAgenda_EmptyRange(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day2)

	require.NoError(t, err)
	assert.Len(t, agenda.Days, 2)
	assert.Empty(t, agenda.Days[0].Entries)
	assert.Empty(t, agenda.Days[1].Entries)
}
