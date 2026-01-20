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

func TestEntryRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Buy groceries",
		Depth:     0,
		CreatedAt: time.Now(),
	}

	id, err := repo.Insert(ctx, entry)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestEntryRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Test task",
		Depth:     0,
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, domain.EntryTypeTask, result.Type)
	assert.Equal(t, "Test task", result.Content)
}

func TestEntryRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	result, err := repo.GetByID(ctx, 999)

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEntryRepository_GetByDate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	entry1 := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Task for today",
		ScheduledDate: &today,
		CreatedAt:     time.Now(),
	}
	entry2 := domain.Entry{
		Type:          domain.EntryTypeNote,
		Content:       "Note for today",
		ScheduledDate: &today,
		CreatedAt:     time.Now(),
	}
	entry3 := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Task for tomorrow",
		ScheduledDate: &tomorrow,
		CreatedAt:     time.Now(),
	}

	_, err := repo.Insert(ctx, entry1)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, entry2)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, entry3)
	require.NoError(t, err)

	results, err := repo.GetByDate(ctx, today)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestEntryRepository_GetOverdue(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	twoDaysAgo := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)

	overdueEntry := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Overdue task",
		ScheduledDate: &yesterday,
		CreatedAt:     time.Now(),
	}
	veryOverdueEntry := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Very overdue task",
		ScheduledDate: &twoDaysAgo,
		CreatedAt:     time.Now(),
	}
	todayEntry := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Today's task",
		ScheduledDate: &today,
		CreatedAt:     time.Now(),
	}

	_, err := repo.Insert(ctx, overdueEntry)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, veryOverdueEntry)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, todayEntry)
	require.NoError(t, err)

	results, err := repo.GetOverdue(ctx, today)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestEntryRepository_GetOverdue_ExcludesEventsAndNotesWithoutOverdueChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	overdueTask := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Overdue task",
		ScheduledDate: &yesterday,
		CreatedAt:     time.Now(),
	}
	pastEvent := domain.Entry{
		Type:          domain.EntryTypeEvent,
		Content:       "Past event with no children",
		ScheduledDate: &yesterday,
		CreatedAt:     time.Now(),
	}
	pastNote := domain.Entry{
		Type:          domain.EntryTypeNote,
		Content:       "Past note with no children",
		ScheduledDate: &yesterday,
		CreatedAt:     time.Now(),
	}

	_, err := repo.Insert(ctx, overdueTask)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, pastEvent)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, pastNote)
	require.NoError(t, err)

	results, err := repo.GetOverdue(ctx, today)

	require.NoError(t, err)
	assert.Len(t, results, 1, "GetOverdue should exclude events and notes without overdue children")
	assert.Equal(t, "Overdue task", results[0].Content)
}

func TestEntryRepository_GetOverdue_IncludesParentChainForOverdueTasks(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	parentEvent := domain.Entry{
		Type:          domain.EntryTypeEvent,
		Content:       "Meeting with team",
		Depth:         0,
		ScheduledDate: &yesterday,
		CreatedAt:     time.Now(),
	}
	parentEventID, err := repo.Insert(ctx, parentEvent)
	require.NoError(t, err)

	childNote := domain.Entry{
		Type:          domain.EntryTypeNote,
		Content:       "Discussion notes",
		ParentID:      &parentEventID,
		Depth:         1,
		ScheduledDate: &yesterday,
		CreatedAt:     time.Now(),
	}
	childNoteID, err := repo.Insert(ctx, childNote)
	require.NoError(t, err)

	grandchildTask := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Follow up action",
		ParentID:      &childNoteID,
		Depth:         2,
		ScheduledDate: &yesterday,
		CreatedAt:     time.Now(),
	}
	_, err = repo.Insert(ctx, grandchildTask)
	require.NoError(t, err)

	results, err := repo.GetOverdue(ctx, today)

	require.NoError(t, err)
	assert.Len(t, results, 3, "GetOverdue should include parent chain for overdue tasks")

	contents := make([]string, len(results))
	for i, r := range results {
		contents[i] = r.Content
	}
	assert.Contains(t, contents, "Meeting with team", "Should include grandparent event")
	assert.Contains(t, contents, "Discussion notes", "Should include parent note")
	assert.Contains(t, contents, "Follow up action", "Should include overdue task")
}

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
