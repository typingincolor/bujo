package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

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

	// Add today's tasks
	_, err := service.LogEntries(ctx, `. Today's task
- Today's note`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	agenda, err := service.GetDailyAgenda(ctx, today)

	require.NoError(t, err)
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

func TestBujoService_SetLocation_PreservesExistingMoodAndWeather(t *testing.T) {
	service, _, dayCtxRepo := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	err := service.SetMood(ctx, today, "happy")
	require.NoError(t, err)
	err = service.SetWeather(ctx, today, "sunny")
	require.NoError(t, err)

	err = service.SetLocation(ctx, today, "Office")
	require.NoError(t, err)

	dayCtx, err := dayCtxRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.NotNil(t, dayCtx)
	require.NotNil(t, dayCtx.Location, "Location should be set")
	assert.Equal(t, "Office", *dayCtx.Location)
	require.NotNil(t, dayCtx.Mood, "Mood should be preserved after setting location")
	assert.Equal(t, "happy", *dayCtx.Mood)
	require.NotNil(t, dayCtx.Weather, "Weather should be preserved after setting location")
	assert.Equal(t, "sunny", *dayCtx.Weather)
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

func TestBujoService_MarkDone_OnlyTasks(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	// Try to mark a note as done
	noteIDs, err := service.LogEntries(ctx, "- This is a note", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkDone(ctx, noteIDs[0])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only tasks")

	// Try to mark an event as done
	eventIDs, err := service.LogEntries(ctx, "o Meeting at 3pm", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkDone(ctx, eventIDs[0])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only tasks")
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

func TestBujoService_GetEntry(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	entry, err := service.GetEntry(ctx, ids[0])

	require.NoError(t, err)
	assert.Equal(t, "Buy groceries", entry.Content)
	assert.Equal(t, domain.EntryTypeTask, entry.Type)
}

func TestBujoService_GetEntry_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	_, err := service.GetEntry(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_GetEntry_ReturnsCurrentVersionAfterUpdate(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	originalID := ids[0]

	err = service.MarkDone(ctx, originalID)
	require.NoError(t, err)

	entry, err := service.GetEntry(ctx, originalID)

	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeDone, entry.Type)
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

func TestBujoService_EditEntry_CannotEditCancelled(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	err = service.EditEntry(ctx, ids[0], "New content")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot edit cancelled entry")
}

func TestBujoService_EditEntryPriority(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.EditEntryPriority(ctx, ids[0], domain.PriorityHigh)
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.PriorityHigh, entry.Priority)
}

func TestBujoService_CyclePriority(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Default priority is none, cycle to low
	err = service.CyclePriority(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.PriorityLow, entry.Priority)

	// Cycle to medium
	err = service.CyclePriority(ctx, ids[0])
	require.NoError(t, err)

	entry, err = entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.PriorityMedium, entry.Priority)

	// Cycle to high
	err = service.CyclePriority(ctx, ids[0])
	require.NoError(t, err)

	entry, err = entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.PriorityHigh, entry.Priority)

	// Cycle back to none
	err = service.CyclePriority(ctx, ids[0])
	require.NoError(t, err)

	entry, err = entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.PriorityNone, entry.Priority)
}
