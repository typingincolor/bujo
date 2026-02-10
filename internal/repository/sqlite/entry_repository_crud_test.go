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

func TestEntryRepository_GetByDate_ExcludesDeleted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	entry1 := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Task to keep",
		ScheduledDate: &today,
		CreatedAt:     time.Now(),
	}
	entry2 := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Task to delete",
		ScheduledDate: &today,
		CreatedAt:     time.Now(),
	}

	id1, err := repo.Insert(ctx, entry1)
	require.NoError(t, err)
	id2, err := repo.Insert(ctx, entry2)
	require.NoError(t, err)

	err = repo.Delete(ctx, id2)
	require.NoError(t, err)

	results, err := repo.GetByDate(ctx, today)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, id1, results[0].ID)
	assert.Equal(t, "Task to keep", results[0].Content)
}

func TestEntryRepository_GetOverdue(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)
	tomorrow := today.AddDate(0, 0, 1)

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
	futureEntry := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Future task",
		ScheduledDate: &tomorrow,
		CreatedAt:     time.Now(),
	}

	_, err := repo.Insert(ctx, overdueEntry)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, veryOverdueEntry)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, todayEntry)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, futureEntry)
	require.NoError(t, err)

	results, err := repo.GetOverdue(ctx)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestEntryRepository_GetOverdue_ExcludesEventsAndNotesWithoutOverdueChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

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

	results, err := repo.GetOverdue(ctx)

	require.NoError(t, err)
	assert.Len(t, results, 1, "GetOverdue should exclude events and notes without overdue children")
	assert.Equal(t, "Overdue task", results[0].Content)
}

func TestEntryRepository_GetOverdue_IncludesParentChainForOverdueTasks(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

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

	results, err := repo.GetOverdue(ctx)

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

func TestEntryRepository_GetByDate_OrdersByCreatedAtThenID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	entry1 := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Created last",
		Depth:     0,
		CreatedAt: time.Date(2026, 1, 28, 10, 0, 0, 0, time.UTC),
	}
	entry2 := domain.Entry{
		Type:      domain.EntryTypeNote,
		Content:   "Created first",
		Depth:     0,
		CreatedAt: time.Date(2026, 1, 28, 8, 0, 0, 0, time.UTC),
	}
	entry3 := domain.Entry{
		Type:      domain.EntryTypeEvent,
		Content:   "Created middle",
		Depth:     0,
		CreatedAt: time.Date(2026, 1, 28, 9, 0, 0, 0, time.UTC),
	}

	_, err := repo.Insert(ctx, entry1)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, entry2)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, entry3)
	require.NoError(t, err)

	entries, err := repo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "Created first", entries[0].Content)
	assert.Equal(t, "Created middle", entries[1].Content)
	assert.Equal(t, "Created last", entries[2].Content)
}

func TestEntryRepository_DeleteByDate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	targetDate := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	otherDate := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	_, err := repo.Insert(ctx, domain.Entry{
		Type: domain.EntryTypeTask, Content: "Task on target date", Depth: 0, CreatedAt: targetDate,
	})
	require.NoError(t, err)

	_, err = repo.Insert(ctx, domain.Entry{
		Type: domain.EntryTypeNote, Content: "Note on target date", Depth: 0, CreatedAt: targetDate,
	})
	require.NoError(t, err)

	otherID, err := repo.Insert(ctx, domain.Entry{
		Type: domain.EntryTypeTask, Content: "Task on other date", Depth: 0, CreatedAt: otherDate,
	})
	require.NoError(t, err)

	err = repo.DeleteByDate(ctx, targetDate)
	require.NoError(t, err)

	targetEntries, err := repo.GetByDate(ctx, targetDate)
	require.NoError(t, err)
	assert.Empty(t, targetEntries)

	otherEntry, err := repo.GetByID(ctx, otherID)
	require.NoError(t, err)
	require.NotNil(t, otherEntry)
	assert.Equal(t, "Task on other date", otherEntry.Content)
}

func TestEntryRepository_DeleteByDate_IncludesAllVersions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	targetDate := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	id, err := repo.Insert(ctx, domain.Entry{
		Type: domain.EntryTypeTask, Content: "Original", Depth: 0, CreatedAt: targetDate,
	})
	require.NoError(t, err)

	entry, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	entry.Content = "Updated"
	err = repo.Update(ctx, *entry)
	require.NoError(t, err)

	err = repo.DeleteByDate(ctx, targetDate)
	require.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM entries WHERE scheduled_date = ?", targetDate.Format("2006-01-02")).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestEntryRepository_MigrationCount_InsertAndRetrieve(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:           domain.EntryTypeTask,
		Content:        "Migrated task",
		Depth:          0,
		CreatedAt:      time.Now(),
		MigrationCount: 2,
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.MigrationCount)
}

func TestEntryRepository_MigrationCount_DefaultsToZero(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Regular task",
		Depth:     0,
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.MigrationCount)
}

func TestEntryRepository_MigrationCount_GetByDate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now()
	entry := domain.Entry{
		Type:           domain.EntryTypeTask,
		Content:        "Migrated task",
		Depth:          0,
		CreatedAt:      now,
		MigrationCount: 3,
	}
	_, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	entries, err := repo.GetByDate(ctx, now)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, 3, entries[0].MigrationCount)
}

func TestEntryRepository_CompletedAt_RoundTrips(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	completedAt := now.Add(3 * 24 * time.Hour).Truncate(time.Second)

	entry := domain.Entry{
		Type:        domain.EntryTypeTask,
		Content:     "Task to complete",
		CreatedAt:   now,
		CompletedAt: &completedAt,
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, result.CompletedAt)
	assert.Equal(t, completedAt.Format(time.RFC3339), result.CompletedAt.Format(time.RFC3339))
}

func TestEntryRepository_OriginalCreatedAt_RoundTrips(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	originalCreatedAt := now.Add(-5 * 24 * time.Hour).Truncate(time.Second)

	entry := domain.Entry{
		Type:              domain.EntryTypeTask,
		Content:           "Migrated task",
		CreatedAt:         now,
		OriginalCreatedAt: &originalCreatedAt,
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, result.OriginalCreatedAt)
	assert.Equal(t, originalCreatedAt.Format(time.RFC3339), result.OriginalCreatedAt.Format(time.RFC3339))
}

func TestEntryRepository_Update_SetsCompletedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Task",
		CreatedAt: now,
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	completedAt := now.Add(2 * 24 * time.Hour).Truncate(time.Second)
	entry.ID = id
	entry.Type = domain.EntryTypeDone
	entry.CompletedAt = &completedAt
	err = repo.Update(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeDone, result.Type)
	require.NotNil(t, result.CompletedAt)
	assert.Equal(t, completedAt.Format(time.RFC3339), result.CompletedAt.Format(time.RFC3339))
}

func TestEntryRepository_Update_SetsOriginalCreatedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Task to migrate",
		CreatedAt: now,
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	originalCreatedAt := now.Add(-10 * 24 * time.Hour).Truncate(time.Second)
	entry.ID = id
	entry.Type = domain.EntryTypeMigrated
	entry.OriginalCreatedAt = &originalCreatedAt
	err = repo.Update(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, result.OriginalCreatedAt, "original_created_at should be set via Update")
	assert.Equal(t, originalCreatedAt.Format(time.RFC3339), result.OriginalCreatedAt.Format(time.RFC3339))
}

func TestEntryRepository_NilCompletedAt_RoundTrips(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEntryRepository(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Incomplete task",
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, result.CompletedAt)
	assert.Nil(t, result.OriginalCreatedAt)
}
