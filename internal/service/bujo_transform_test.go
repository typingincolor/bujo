package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestBujoService_ParseEntries_SingleTask(t *testing.T) {
	service, _, _ := setupBujoService(t)

	entries, err := service.ParseEntries(". Buy groceries")

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, domain.EntryTypeTask, entries[0].Type)
	assert.Equal(t, "Buy groceries", entries[0].Content)
}

func TestBujoService_ParseEntries_MultipleTypes(t *testing.T) {
	service, _, _ := setupBujoService(t)

	input := `. Task one
- Note one
o Event one`

	entries, err := service.ParseEntries(input)

	require.NoError(t, err)
	require.Len(t, entries, 3)
	assert.Equal(t, domain.EntryTypeTask, entries[0].Type)
	assert.Equal(t, domain.EntryTypeNote, entries[1].Type)
	assert.Equal(t, domain.EntryTypeEvent, entries[2].Type)
}

func TestBujoService_ParseEntries_WithHierarchy(t *testing.T) {
	service, _, _ := setupBujoService(t)

	input := `. Parent task
  - Child note`

	entries, err := service.ParseEntries(input)

	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, 0, entries[0].Depth)
	assert.Equal(t, 1, entries[1].Depth)
}

func TestBujoService_ParseEntries_EmptyInput(t *testing.T) {
	service, _, _ := setupBujoService(t)

	entries, err := service.ParseEntries("")

	require.NoError(t, err)
	assert.Empty(t, entries)
}

// Cancel Entry Tests

func TestBujoService_CancelEntry_TaskBecomesCancelled(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeCancelled, entry.Type)
}

func TestBujoService_CancelEntry_NotFoundReturnsError(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.CancelEntry(ctx, 9999)

	assert.Error(t, err)
}

func TestBujoService_CancelEntry_AlreadyCancelledNoOp(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeCancelled, entry.Type)
}

func TestBujoService_UncancelEntry_CancelledTaskBecomesTask(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	err = service.UncancelEntry(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, entry.Type)
}

func TestBujoService_UncancelEntry_CancelledNoteBecomesNote(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "- This is a note", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	err = service.UncancelEntry(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeNote, entry.Type)
}

func TestBujoService_UncancelEntry_CancelledEventBecomesEvent(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "o Meeting at 3pm", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	err = service.UncancelEntry(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeEvent, entry.Type)
}

func TestBujoService_UncancelEntry_NotCancelledNoOp(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.UncancelEntry(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, entry.Type)
}

// Retype Entry Tests

func TestBujoService_RetypeEntry_TaskToNote(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeNote)
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeNote, entry.Type)
}

func TestBujoService_RetypeEntry_TaskToEvent(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeEvent)
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeEvent, entry.Type)
}

func TestBujoService_RetypeEntry_NoteToTask(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "- Some note", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeTask)
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, entry.Type)
}

func TestBujoService_RetypeEntry_InvalidTypeReturnsError(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryType("invalid"))

	assert.Error(t, err)
}

func TestBujoService_RetypeEntry_CannotRetypeToDone(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeDone)

	assert.Error(t, err)
}

func TestBujoService_RetypeEntry_CannotRetypeToMigrated(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeMigrated)

	assert.Error(t, err)
}

func TestBujoService_RetypeEntry_NotFoundReturnsError(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.RetypeEntry(ctx, 9999, domain.EntryTypeNote)

	assert.Error(t, err)
}

func TestBujoService_RetypeEntry_CannotRetypeCancelled(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.CancelEntry(ctx, ids[0])
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeNote)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot change type of cancelled entry")
}

func TestBujoService_RetypeEntry_CannotRetypeDone(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkDone(ctx, ids[0])
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeNote)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot change type of completed entry")
}

func TestBujoService_RetypeEntry_CannotRetypeMigrated(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	tomorrow := today.AddDate(0, 0, 1)
	_, err = service.MigrateEntry(ctx, ids[0], tomorrow)
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeNote)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot change type of migrated entry")
}

func TestBujoService_RetypeEntry_CannotRetypeAnswered(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "? What time is the meeting", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkAnswered(ctx, ids[0], "3pm")
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeNote)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot change type of answered entry")
}

func TestBujoService_RetypeEntry_CannotRetypeAnswer(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `? What time is the meeting
	a 3pm`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 2)

	answerID := ids[1]
	err = service.RetypeEntry(ctx, answerID, domain.EntryTypeNote)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot change type of answer entry")
}

func TestBujoService_RetypeEntry_PreservesContent(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.RetypeEntry(ctx, ids[0], domain.EntryTypeNote)
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, "Buy groceries", entry.Content)
}

func TestBujoService_SearchEntries_FindsMatchingContent(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	_, err := service.LogEntries(ctx, `. Buy groceries
- Meeting notes
o Doctor appointment`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	opts := domain.NewSearchOptions("groceries")
	results, err := service.SearchEntries(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Buy groceries", results[0].Content)
}

func TestBujoService_SearchEntries_WithTypeFilter(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	_, err := service.LogEntries(ctx, `. Project task
- Project notes
o Project meeting`, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	opts := domain.NewSearchOptions("project").WithType(domain.EntryTypeNote)
	results, err := service.SearchEntries(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, domain.EntryTypeNote, results[0].Type)
}

func TestBujoService_SearchEntries_WithDateRange(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	jan5 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	jan10 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_, err := service.LogEntries(ctx, ". Early meeting", LogEntriesOptions{Date: jan5})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Middle meeting", LogEntriesOptions{Date: jan10})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Late meeting", LogEntriesOptions{Date: jan15})
	require.NoError(t, err)

	from := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	opts := domain.NewSearchOptions("meeting").WithDateRange(from, to)
	results, err := service.SearchEntries(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Middle meeting", results[0].Content)
}

func TestBujoService_SearchEntries_EmptyQuery(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	_, err := service.LogEntries(ctx, ". Some task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	opts := domain.NewSearchOptions("")
	results, err := service.SearchEntries(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestBujoService_SearchEntries_NoMatches(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	_, err := service.LogEntries(ctx, ". Some task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	opts := domain.NewSearchOptions("nonexistent")
	results, err := service.SearchEntries(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 0)
}
