package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func createTestHabit(t *testing.T, repo *HabitRepository, name string) int64 {
	t.Helper()
	id, err := repo.Insert(context.Background(), domain.Habit{
		Name:       name,
		GoalPerDay: 1,
		CreatedAt:  time.Now(),
	})
	require.NoError(t, err)
	return id
}

func TestHabitLogRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "Gym")

	log := domain.HabitLog{
		HabitID:  habitID,
		Count:    1,
		LoggedAt: time.Now(),
	}

	id, err := repo.Insert(ctx, log)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestHabitLogRepository_GetByHabitID(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "Water")

	logs := []domain.HabitLog{
		{HabitID: habitID, Count: 2, LoggedAt: time.Now()},
		{HabitID: habitID, Count: 3, LoggedAt: time.Now().Add(time.Hour)},
	}
	for _, log := range logs {
		_, err := repo.Insert(ctx, log)
		require.NoError(t, err)
	}

	results, err := repo.GetByHabitID(ctx, habitID)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestHabitLogRepository_GetRange(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "Meditation")

	now := time.Now()
	logs := []domain.HabitLog{
		{HabitID: habitID, Count: 1, LoggedAt: now.AddDate(0, 0, -5)},
		{HabitID: habitID, Count: 1, LoggedAt: now.AddDate(0, 0, -3)},
		{HabitID: habitID, Count: 1, LoggedAt: now.AddDate(0, 0, -1)},
		{HabitID: habitID, Count: 1, LoggedAt: now},
	}
	for _, log := range logs {
		_, err := repo.Insert(ctx, log)
		require.NoError(t, err)
	}

	start := now.AddDate(0, 0, -4)
	end := now

	results, err := repo.GetRange(ctx, habitID, start, end)

	require.NoError(t, err)
	assert.Len(t, results, 3) // -3, -1, and today
}

func TestHabitLogRepository_GetAllRange(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habit1ID := createTestHabit(t, habitRepo, "Gym")
	habit2ID := createTestHabit(t, habitRepo, "Water")

	now := time.Now()
	logs := []domain.HabitLog{
		{HabitID: habit1ID, Count: 1, LoggedAt: now},
		{HabitID: habit2ID, Count: 8, LoggedAt: now},
		{HabitID: habit1ID, Count: 1, LoggedAt: now.AddDate(0, 0, -10)}, // Outside range
	}
	for _, log := range logs {
		_, err := repo.Insert(ctx, log)
		require.NoError(t, err)
	}

	start := now.AddDate(0, 0, -7)
	end := now

	results, err := repo.GetAllRange(ctx, start, end)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestHabitLogRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "ToDelete")

	log := domain.HabitLog{
		HabitID:  habitID,
		Count:    1,
		LoggedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, log)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	results, err := repo.GetByHabitID(ctx, habitID)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestHabitLogRepository_Delete_SoftDeletes(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "SoftDeleteLogTest")

	log := domain.HabitLog{
		HabitID:  habitID,
		Count:    1,
		LoggedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, log)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Verify log is soft deleted (not visible via GetByHabitID)
	results, err := repo.GetByHabitID(ctx, habitID)
	require.NoError(t, err)
	assert.Len(t, results, 0)

	// But data should still exist in database
	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM habit_logs WHERE habit_id = ?`, habitID).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1, "Soft deleted log should still exist in DB")
}

func TestHabitLogRepository_GetDeleted_ReturnsDeletedLogs(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "DeletedLogTest")

	log := domain.HabitLog{
		HabitID:  habitID,
		Count:    5,
		LoggedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, log)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Get deleted logs
	deleted, err := repo.GetDeleted(ctx)
	require.NoError(t, err)
	require.Len(t, deleted, 1)
	assert.Equal(t, 5, deleted[0].Count)
}

func TestHabitLogRepository_Restore_BringsBackDeletedLog(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "RestoreLogTest")

	log := domain.HabitLog{
		HabitID:  habitID,
		Count:    3,
		LoggedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, log)
	require.NoError(t, err)

	l, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	entityID := l.EntityID

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Verify it's gone
	results, err := repo.GetByHabitID(ctx, habitID)
	require.NoError(t, err)
	assert.Len(t, results, 0)

	// Restore it
	newID, err := repo.Restore(ctx, entityID)
	require.NoError(t, err)
	assert.NotZero(t, newID)

	// Verify it's back
	restored, err := repo.GetByID(ctx, newID)
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, 3, restored.Count)
}

func TestHabitLogRepository_GetRangeByEntityID(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "EntityIDTest")

	habit, err := habitRepo.GetByID(ctx, habitID)
	require.NoError(t, err)
	habitEntityID := habit.EntityID

	now := time.Now()
	logs := []domain.HabitLog{
		{HabitID: habitID, HabitEntityID: habitEntityID, Count: 1, LoggedAt: now.AddDate(0, 0, -5)},
		{HabitID: habitID, HabitEntityID: habitEntityID, Count: 2, LoggedAt: now.AddDate(0, 0, -3)},
		{HabitID: habitID, HabitEntityID: habitEntityID, Count: 3, LoggedAt: now.AddDate(0, 0, -1)},
		{HabitID: habitID, HabitEntityID: habitEntityID, Count: 4, LoggedAt: now},
	}
	for _, log := range logs {
		_, err := repo.Insert(ctx, log)
		require.NoError(t, err)
	}

	start := now.AddDate(0, 0, -4)
	end := now

	results, err := repo.GetRangeByEntityID(ctx, habitEntityID, start, end)

	require.NoError(t, err)
	assert.Len(t, results, 3) // -3, -1, and today
}

func TestHabitLogRepository_GetByID_ReturnsCurrentVersionFromOldRowID(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	habitID := createTestHabit(t, habitRepo, "GetByIDTest")

	log := domain.HabitLog{
		HabitID:  habitID,
		Count:    5,
		LoggedAt: time.Now(),
	}
	originalID, err := repo.Insert(ctx, log)
	require.NoError(t, err)

	original, err := repo.GetByID(ctx, originalID)
	require.NoError(t, err)
	entityID := original.EntityID

	err = repo.Delete(ctx, originalID)
	require.NoError(t, err)

	newID, err := repo.Restore(ctx, entityID)
	require.NoError(t, err)
	require.NotEqual(t, originalID, newID, "Restored log should have a new row ID")

	result, err := repo.GetByID(ctx, originalID)
	require.NoError(t, err)
	require.NotNil(t, result, "GetByID with old row ID should return current version")
	assert.Equal(t, newID, result.ID, "Should return the current version's row ID")
	assert.Equal(t, 5, result.Count)
	assert.Equal(t, entityID, result.EntityID)
}

func TestHabitLogRepository_GetRangeByEntityID_AfterHabitRename(t *testing.T) {
	db := setupTestDB(t)
	habitRepo := NewHabitRepository(db)
	repo := NewHabitLogRepository(db)
	ctx := context.Background()

	originalID := createTestHabit(t, habitRepo, "OriginalName")

	habit, err := habitRepo.GetByID(ctx, originalID)
	require.NoError(t, err)
	habitEntityID := habit.EntityID

	now := time.Now()
	log := domain.HabitLog{
		HabitID:       originalID,
		HabitEntityID: habitEntityID,
		Count:         5,
		LoggedAt:      now,
	}
	_, err = repo.Insert(ctx, log)
	require.NoError(t, err)

	habit.Name = "RenamedHabit"
	err = habitRepo.Update(ctx, *habit)
	require.NoError(t, err)

	renamedHabit, err := habitRepo.GetByID(ctx, originalID)
	require.NoError(t, err)
	assert.Equal(t, "RenamedHabit", renamedHabit.Name)
	assert.NotEqual(t, originalID, renamedHabit.ID, "ID should change after rename")
	assert.Equal(t, habitEntityID, renamedHabit.EntityID, "EntityID should remain stable")

	start := now.AddDate(0, 0, -1)
	end := now.AddDate(0, 0, 1)

	results, err := repo.GetRangeByEntityID(ctx, habitEntityID, start, end)
	require.NoError(t, err)
	assert.Len(t, results, 1, "Logs should be retrievable via entity_id after habit rename")
	assert.Equal(t, 5, results[0].Count)
}
