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
