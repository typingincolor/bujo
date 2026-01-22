package sqlite

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestEntryRepository_Search_BasicContentSearch(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entries := []domain.Entry{
		{Type: domain.EntryTypeTask, Content: "Buy groceries for dinner", CreatedAt: time.Now()},
		{Type: domain.EntryTypeNote, Content: "Meeting notes from project sync", CreatedAt: time.Now()},
		{Type: domain.EntryTypeEvent, Content: "Doctor appointment", CreatedAt: time.Now()},
		{Type: domain.EntryTypeTask, Content: "Grocery list updated", CreatedAt: time.Now()},
	}
	for _, e := range entries {
		_, err := repo.Insert(ctx, e)
		require.NoError(t, err)
	}

	opts := domain.NewSearchOptions("grocer")
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	for _, r := range results {
		assert.Contains(t, strings.ToLower(r.Content), "grocer", "All results should contain search term (case-insensitive)")
	}
}

func TestEntryRepository_Search_CaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entries := []domain.Entry{
		{Type: domain.EntryTypeTask, Content: "Buy GROCERIES", CreatedAt: time.Now()},
		{Type: domain.EntryTypeNote, Content: "groceries list", CreatedAt: time.Now()},
		{Type: domain.EntryTypeTask, Content: "Groceries shopping", CreatedAt: time.Now()},
	}
	for _, e := range entries {
		_, err := repo.Insert(ctx, e)
		require.NoError(t, err)
	}

	opts := domain.NewSearchOptions("groceries")
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 3, "Search should be case-insensitive")
}

func TestEntryRepository_Search_FilterByType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entries := []domain.Entry{
		{Type: domain.EntryTypeTask, Content: "Task with keyword", CreatedAt: time.Now()},
		{Type: domain.EntryTypeNote, Content: "Note with keyword", CreatedAt: time.Now()},
		{Type: domain.EntryTypeEvent, Content: "Event with keyword", CreatedAt: time.Now()},
	}
	for _, e := range entries {
		_, err := repo.Insert(ctx, e)
		require.NoError(t, err)
	}

	opts := domain.NewSearchOptions("keyword").WithType(domain.EntryTypeNote)
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, domain.EntryTypeNote, results[0].Type)
}

func TestEntryRepository_Search_FilterByDateRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	jan5 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	jan10 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	entries := []domain.Entry{
		{Type: domain.EntryTypeTask, Content: "Early meeting", ScheduledDate: &jan5, CreatedAt: time.Now()},
		{Type: domain.EntryTypeTask, Content: "Middle meeting", ScheduledDate: &jan10, CreatedAt: time.Now()},
		{Type: domain.EntryTypeTask, Content: "Late meeting", ScheduledDate: &jan15, CreatedAt: time.Now()},
	}
	for _, e := range entries {
		_, err := repo.Insert(ctx, e)
		require.NoError(t, err)
	}

	from := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	opts := domain.NewSearchOptions("meeting").WithDateRange(from, to)
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Middle meeting", results[0].Content)
}

func TestEntryRepository_Search_CombinedFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	jan10 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	entries := []domain.Entry{
		{Type: domain.EntryTypeTask, Content: "Project meeting", ScheduledDate: &jan10, CreatedAt: time.Now()},
		{Type: domain.EntryTypeNote, Content: "Project notes", ScheduledDate: &jan10, CreatedAt: time.Now()},
		{Type: domain.EntryTypeTask, Content: "Project deadline", ScheduledDate: &jan15, CreatedAt: time.Now()},
	}
	for _, e := range entries {
		_, err := repo.Insert(ctx, e)
		require.NoError(t, err)
	}

	from := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	opts := domain.NewSearchOptions("project").WithType(domain.EntryTypeTask).WithDateRange(from, to)
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Project meeting", results[0].Content)
	assert.Equal(t, domain.EntryTypeTask, results[0].Type)
}

func TestEntryRepository_Search_NoResults(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Something completely different",
		CreatedAt: time.Now(),
	}
	_, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	opts := domain.NewSearchOptions("nonexistent")
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestEntryRepository_Search_RespectsLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		entry := domain.Entry{
			Type:      domain.EntryTypeTask,
			Content:   "Repeated task content",
			CreatedAt: time.Now(),
		}
		_, err := repo.Insert(ctx, entry)
		require.NoError(t, err)
	}

	opts := domain.NewSearchOptions("repeated").WithLimit(3)
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestEntryRepository_Search_ExcludesDeletedEntries(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry1 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Active searchable entry",
		CreatedAt: time.Now(),
	}
	_, err := repo.Insert(ctx, entry1)
	require.NoError(t, err)

	entry2 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Deleted searchable entry",
		CreatedAt: time.Now(),
	}
	id2, err := repo.Insert(ctx, entry2)
	require.NoError(t, err)

	err = repo.Delete(ctx, id2)
	require.NoError(t, err)

	opts := domain.NewSearchOptions("searchable")
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1, "Search should exclude deleted entries")
	assert.Equal(t, "Active searchable entry", results[0].Content)
}

func TestEntryRepository_Search_ExcludesOldVersions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Original searchable content",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	inserted, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	inserted.Content = "Updated searchable content"
	err = repo.Update(ctx, *inserted)
	require.NoError(t, err)

	opts := domain.NewSearchOptions("searchable")
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 1, "Search should return only current version")
	assert.Equal(t, "Updated searchable content", results[0].Content)
}

func TestEntryRepository_Search_IncludesCompletedEntries(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entries := []domain.Entry{
		{Type: domain.EntryTypeTask, Content: "Active task with keyword", CreatedAt: time.Now()},
		{Type: domain.EntryTypeDone, Content: "Completed task with keyword", CreatedAt: time.Now()},
		{Type: domain.EntryTypeCancelled, Content: "Cancelled task with keyword", CreatedAt: time.Now()},
	}
	for _, e := range entries {
		_, err := repo.Insert(ctx, e)
		require.NoError(t, err)
	}

	opts := domain.NewSearchOptions("keyword")
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 3, "Search should include completed and cancelled entries")
}

func TestEntryRepository_Search_EmptyQuery(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Some content",
		CreatedAt: time.Now(),
	}
	_, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	opts := domain.NewSearchOptions("")
	results, err := repo.Search(ctx, opts)

	require.NoError(t, err)
	assert.Len(t, results, 0, "Empty query should return no results")
}

func TestEntryRepository_Update_PreservesOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	sameTimestamp := time.Date(2026, 1, 6, 10, 30, 0, 0, time.UTC)

	entry1 := domain.Entry{
		Type:          domain.EntryTypeNote,
		Content:       "entry 1",
		ScheduledDate: &today,
		CreatedAt:     sameTimestamp,
	}
	entry2 := domain.Entry{
		Type:          domain.EntryTypeNote,
		Content:       "entry 2",
		ScheduledDate: &today,
		CreatedAt:     sameTimestamp,
	}
	entry3 := domain.Entry{
		Type:          domain.EntryTypeNote,
		Content:       "entry 3",
		ScheduledDate: &today,
		CreatedAt:     sameTimestamp,
	}

	id1, err := repo.Insert(ctx, entry1)
	require.NoError(t, err)
	id2, err := repo.Insert(ctx, entry2)
	require.NoError(t, err)
	id3, err := repo.Insert(ctx, entry3)
	require.NoError(t, err)

	entriesBefore, err := repo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, entriesBefore, 3)
	assert.Equal(t, "entry 1", entriesBefore[0].Content)
	assert.Equal(t, "entry 2", entriesBefore[1].Content)
	assert.Equal(t, "entry 3", entriesBefore[2].Content)

	updatedEntry, err := repo.GetByID(ctx, id1)
	require.NoError(t, err)
	updatedEntry.Content = "entry 4"
	err = repo.Update(ctx, *updatedEntry)
	require.NoError(t, err)

	entriesAfter, err := repo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, entriesAfter, 3)

	assert.Equal(t, "entry 4", entriesAfter[0].Content, "updated entry should maintain position at index 0")
	assert.Equal(t, "entry 2", entriesAfter[1].Content)
	assert.Equal(t, "entry 3", entriesAfter[2].Content)

	_ = id2
	_ = id3
}

func TestEntryRepository_GetLastModified_ReturnsLatestValidFrom(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	baseTime := time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)
	laterTime := time.Date(2026, 1, 20, 11, 0, 0, 0, time.UTC)

	entry1 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "First entry",
		CreatedAt: baseTime,
	}
	_, err := repo.Insert(ctx, entry1)
	require.NoError(t, err)

	lastModified, err := repo.GetLastModified(ctx)
	require.NoError(t, err)
	assert.False(t, lastModified.IsZero(), "should return a non-zero timestamp")

	entry2 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Second entry",
		CreatedAt: laterTime,
	}
	_, err = repo.Insert(ctx, entry2)
	require.NoError(t, err)

	newLastModified, err := repo.GetLastModified(ctx)
	require.NoError(t, err)
	assert.True(t, newLastModified.After(lastModified) || newLastModified.Equal(lastModified),
		"last modified should be >= previous after insert")
}

func TestEntryRepository_GetLastModified_EmptyTable(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	lastModified, err := repo.GetLastModified(ctx)
	require.NoError(t, err)
	assert.True(t, lastModified.IsZero(), "should return zero time for empty table")
}
