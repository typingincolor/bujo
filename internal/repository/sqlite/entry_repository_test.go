package sqlite

import (
	"context"
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

	entry.ID = id
	entry.Type = domain.EntryTypeDone
	entry.Content = "Updated content"

	err = repo.Update(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeDone, result.Type)
	assert.Equal(t, "Updated content", result.Content)
}

func TestEntryRepository_Delete(t *testing.T) {
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
