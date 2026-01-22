package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHabitService_RenameHabitByID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	// Rename by ID
	err = service.RenameHabitByID(ctx, habitID, "Workout")
	require.NoError(t, err)

	// Verify new name
	today := time.Now()
	from := today.AddDate(0, 0, -30)
	details, err := service.InspectHabitByID(ctx, habitID, from, today, today)
	require.NoError(t, err)
	assert.Equal(t, "Workout", details.Name)
}

func TestHabitService_RenameHabit_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.RenameHabit(ctx, "NonExistent", "NewName")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_SetHabitGoal(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit
	err := service.LogHabit(ctx, "Water", 1)
	require.NoError(t, err)

	// Set goal to 8
	err = service.SetHabitGoal(ctx, "Water", 8)
	require.NoError(t, err)

	// Verify goal via inspect
	today := time.Now()
	from := today.AddDate(0, 0, -30)
	details, err := service.InspectHabit(ctx, "Water", from, today, today)
	require.NoError(t, err)
	assert.Equal(t, 8, details.GoalPerDay)
}

func TestHabitService_SetHabitGoalByID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit
	err := service.LogHabit(ctx, "Water", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	// Set goal by ID
	err = service.SetHabitGoalByID(ctx, habitID, 10)
	require.NoError(t, err)

	// Verify goal
	today := time.Now()
	from := today.AddDate(0, 0, -30)
	details, err := service.InspectHabitByID(ctx, habitID, from, today, today)
	require.NoError(t, err)
	assert.Equal(t, 10, details.GoalPerDay)
}

func TestHabitService_SetHabitGoal_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.SetHabitGoal(ctx, "NonExistent", 5)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_SetHabitGoal_InvalidGoal(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit
	err := service.LogHabit(ctx, "Water", 1)
	require.NoError(t, err)

	// Try to set invalid goal
	err = service.SetHabitGoal(ctx, "Water", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "goal must be at least 1")
}

func TestHabitService_DeleteHabit(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit with logs
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)
	err = service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	// Delete the habit
	err = service.DeleteHabit(ctx, "Gym")
	require.NoError(t, err)

	// Verify habit no longer exists
	today := time.Now()
	from := today.AddDate(0, 0, -30)
	_, err = service.InspectHabit(ctx, "Gym", from, today, today)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_DeleteHabitByID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	// Get the habit ID
	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	// Delete by ID
	err = service.DeleteHabitByID(ctx, habitID)
	require.NoError(t, err)

	// Verify habit no longer exists
	today := time.Now()
	from := today.AddDate(0, 0, -30)
	_, err = service.InspectHabitByID(ctx, habitID, from, today, today)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_DeleteHabit_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.DeleteHabit(ctx, "NonExistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_DeleteHabitByID_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.DeleteHabitByID(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_HabitExists(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Non-existent habit
	exists, err := service.HabitExists(ctx, "Gym")
	require.NoError(t, err)
	assert.False(t, exists)

	// Create habit
	err = service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	// Now it exists
	exists, err = service.HabitExists(ctx, "Gym")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestHabitService_GetTrackerStatus_IncludesWeeklyProgress(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 9, 12, 0, 0, 0, time.UTC)

	// Create habit with weekly goal (LogHabit creates one log today)
	err := service.LogHabitForDate(ctx, "Workout", 1, today)
	require.NoError(t, err)
	err = service.SetHabitWeeklyGoal(ctx, "Workout", 5)
	require.NoError(t, err)

	// Log 2 more times this week (total 3)
	for i := 1; i <= 2; i++ {
		err = service.LogHabitForDate(ctx, "Workout", 1, today.AddDate(0, 0, -i))
		require.NoError(t, err)
	}

	status, err := service.GetTrackerStatus(ctx, today, 7)
	require.NoError(t, err)
	require.Len(t, status.Habits, 1)
	assert.Equal(t, 5, status.Habits[0].GoalPerWeek)
	assert.Equal(t, 60.0, status.Habits[0].WeeklyProgress) // 3/5 = 60%
}

func TestHabitService_GetTrackerStatus_IncludesMonthlyProgress(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	// Create habit with monthly goal by logging first
	err := service.LogHabitForDate(ctx, "Reading", 1, time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	err = service.SetHabitMonthlyGoal(ctx, "Reading", 20)
	require.NoError(t, err)

	// Log 9 more times this month (total 10)
	for i := 1; i < 10; i++ {
		err = service.LogHabitForDate(ctx, "Reading", 1, time.Date(2026, 1, i+1, 10, 0, 0, 0, time.UTC))
		require.NoError(t, err)
	}

	status, err := service.GetTrackerStatus(ctx, today, 7)
	require.NoError(t, err)
	require.Len(t, status.Habits, 1)
	assert.Equal(t, 20, status.Habits[0].GoalPerMonth)
	assert.Equal(t, 50.0, status.Habits[0].MonthlyProgress) // 10/20 = 50%
}

func TestHabitService_SetHabitWeeklyGoal(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create habit
	err := service.LogHabit(ctx, "Exercise", 1)
	require.NoError(t, err)

	// Set weekly goal
	err = service.SetHabitWeeklyGoal(ctx, "Exercise", 4)
	require.NoError(t, err)

	// Verify
	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Equal(t, 4, status.Habits[0].GoalPerWeek)
}

func TestHabitService_SetHabitMonthlyGoal(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create habit
	err := service.LogHabit(ctx, "Meditation", 1)
	require.NoError(t, err)

	// Set monthly goal
	err = service.SetHabitMonthlyGoal(ctx, "Meditation", 15)
	require.NoError(t, err)

	// Verify
	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Equal(t, 15, status.Habits[0].GoalPerMonth)
}

func TestHabitService_RemoveHabitLogForDateByID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Create habit and log multiple times for today
	err := service.LogHabitForDate(ctx, "Gym", 1, today)
	require.NoError(t, err)
	err = service.LogHabitForDate(ctx, "Gym", 1, today)
	require.NoError(t, err)
	// Also log for yesterday
	err = service.LogHabitForDate(ctx, "Gym", 1, yesterday)
	require.NoError(t, err)

	// Get habit ID
	status, err := service.GetTrackerStatus(ctx, today, 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	// Verify we have 3 logs total, 2 for today
	from := today.AddDate(0, 0, -7)
	details, err := service.InspectHabitByID(ctx, habitID, from, today, today)
	require.NoError(t, err)
	require.Len(t, details.Logs, 3)

	// Remove one log for today
	err = service.RemoveHabitLogForDateByID(ctx, habitID, today)
	require.NoError(t, err)

	// Verify only 2 logs remain (1 for today, 1 for yesterday)
	details, err = service.InspectHabitByID(ctx, habitID, from, today, today)
	require.NoError(t, err)
	assert.Len(t, details.Logs, 2)

	// Remove another log for today
	err = service.RemoveHabitLogForDateByID(ctx, habitID, today)
	require.NoError(t, err)

	// Verify only 1 log remains (yesterday's)
	details, err = service.InspectHabitByID(ctx, habitID, from, today, today)
	require.NoError(t, err)
	assert.Len(t, details.Logs, 1)
}

func TestHabitService_RemoveHabitLogForDateByID_NoLogsForDate(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Create habit with log for yesterday only
	err := service.LogHabitForDate(ctx, "Gym", 1, yesterday)
	require.NoError(t, err)

	// Get habit ID
	status, err := service.GetTrackerStatus(ctx, today, 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	// Try to remove log for today (which has no logs)
	err = service.RemoveHabitLogForDateByID(ctx, habitID, today)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no logs")
}

func TestHabitService_RemoveHabitLogForDateByID_HabitNotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.RemoveHabitLogForDateByID(ctx, 99999, time.Now())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_CreateHabit_ReturnsHabitID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	habitID, err := service.CreateHabit(ctx, "Morning Run")

	require.NoError(t, err)
	assert.Greater(t, habitID, int64(0))

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	require.Len(t, status.Habits, 1)
	assert.Equal(t, "Morning Run", status.Habits[0].Name)
	assert.Equal(t, habitID, status.Habits[0].ID)
}

func TestHabitService_CreateHabit_ExistingHabitReturnsID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	firstID, err := service.CreateHabit(ctx, "Workout")
	require.NoError(t, err)

	secondID, err := service.CreateHabit(ctx, "Workout")
	require.NoError(t, err)

	assert.Equal(t, firstID, secondID)
}
