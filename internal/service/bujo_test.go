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
	t.Cleanup(func() { db.Close() })

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
