package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupHabitService(t *testing.T) *HabitService {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	habitRepo := sqlite.NewHabitRepository(db)
	logRepo := sqlite.NewHabitLogRepository(db)

	return NewHabitService(habitRepo, logRepo)
}

func TestHabitService_LogHabit_CreatesNewHabit(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.LogHabit(ctx, "Gym", 1)

	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Len(t, status.Habits, 1)
	assert.Equal(t, "Gym", status.Habits[0].Name)
}

func TestHabitService_LogHabit_UsesExistingHabit(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.LogHabit(ctx, "Water", 3)
	require.NoError(t, err)
	err = service.LogHabit(ctx, "Water", 2)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Len(t, status.Habits, 1) // Still only one habit
}

func TestHabitService_LogHabit_WithCount(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.LogHabit(ctx, "Water", 8)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Equal(t, 8, status.Habits[0].TodayCount)
}

func TestHabitService_GetTrackerStatus(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Log some habits
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)
	err = service.LogHabit(ctx, "Meditation", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)

	require.NoError(t, err)
	assert.Len(t, status.Habits, 2)
}

func TestHabitService_GetTrackerStatus_CalculatesStreak(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Now()

	// Log habit for 3 consecutive days
	err := service.LogHabitForDate(ctx, "Gym", 1, today)
	require.NoError(t, err)
	err = service.LogHabitForDate(ctx, "Gym", 1, today.AddDate(0, 0, -1))
	require.NoError(t, err)
	err = service.LogHabitForDate(ctx, "Gym", 1, today.AddDate(0, 0, -2))
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, today, 7)

	require.NoError(t, err)
	assert.Equal(t, 3, status.Habits[0].CurrentStreak)
}

func TestHabitService_GetTrackerStatus_CalculatesCompletion(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	// Log habit for 4 out of 7 days (days 0, -1, -2, -3)
	for i := 0; i < 4; i++ {
		err := service.LogHabitForDate(ctx, "Gym", 1, today.AddDate(0, 0, -i))
		require.NoError(t, err)
	}

	status, err := service.GetTrackerStatus(ctx, today, 7)

	require.NoError(t, err)
	assert.InDelta(t, 57.14, status.Habits[0].CompletionPercent, 1.0)
}

func TestHabitService_GetTrackerStatus_DayHistory(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	// Log for all 7 days
	for i := 0; i < 7; i++ {
		err := service.LogHabitForDate(ctx, "Meditation", 1, today.AddDate(0, 0, -i))
		require.NoError(t, err)
	}

	status, err := service.GetTrackerStatus(ctx, today, 7)

	require.NoError(t, err)
	assert.Len(t, status.Habits[0].DayHistory, 7)
	assert.InDelta(t, 100.0, status.Habits[0].CompletionPercent, 0.1)
}
