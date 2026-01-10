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
	t.Cleanup(func() { _ = db.Close() })

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

func TestHabitService_GetTrackerStatus_IncludesHabitID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)

	require.NoError(t, err)
	assert.Greater(t, status.Habits[0].ID, int64(0))
}

func TestHabitService_LogHabitByID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit first
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	// Log by ID
	err = service.LogHabitByID(ctx, habitID, 1)
	require.NoError(t, err)

	// Verify count increased
	status, err = service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Equal(t, 2, status.Habits[0].TodayCount)
}

func TestHabitService_LogHabitByID_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.LogHabitByID(ctx, 99999, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_UndoLastLog(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Log a habit twice
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)
	err = service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Equal(t, 2, status.Habits[0].TodayCount)

	// Undo the last log
	err = service.UndoLastLog(ctx, "Gym")
	require.NoError(t, err)

	// Verify count decreased
	status, err = service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Equal(t, 1, status.Habits[0].TodayCount)
}

func TestHabitService_UndoLastLog_ByID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Log a habit
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	// Undo by ID
	err = service.UndoLastLogByID(ctx, habitID)
	require.NoError(t, err)

	// Verify count is 0
	status, err = service.GetTrackerStatus(ctx, time.Now(), 7)
	require.NoError(t, err)
	assert.Equal(t, 0, status.Habits[0].TodayCount)
}

func TestHabitService_UndoLastLog_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.UndoLastLog(ctx, "NonExistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_UndoLastLog_NoLogs(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create habit but don't log
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	// Undo the only log
	err = service.UndoLastLog(ctx, "Gym")
	require.NoError(t, err)

	// Try to undo again - should fail
	err = service.UndoLastLog(ctx, "Gym")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no logs")
}

func TestHabitService_InspectHabit(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	// Log habit for multiple days
	for i := 0; i < 5; i++ {
		err := service.LogHabitForDate(ctx, "Gym", 1, today.AddDate(0, 0, -i))
		require.NoError(t, err)
	}

	from := today.AddDate(0, 0, -30)
	to := today

	details, err := service.InspectHabit(ctx, "Gym", from, to, today)

	require.NoError(t, err)
	assert.Equal(t, "Gym", details.Name)
	assert.Greater(t, details.ID, int64(0))
	assert.Len(t, details.Logs, 5)
	assert.Equal(t, 5, details.CurrentStreak)
}

func TestHabitService_InspectHabitByID(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Now()

	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	status, err := service.GetTrackerStatus(ctx, today, 7)
	require.NoError(t, err)
	habitID := status.Habits[0].ID

	from := today.AddDate(0, 0, -30)
	to := today

	details, err := service.InspectHabitByID(ctx, habitID, from, to, today)

	require.NoError(t, err)
	assert.Equal(t, "Gym", details.Name)
	assert.Len(t, details.Logs, 1)
}

func TestHabitService_InspectHabit_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Now()
	from := today.AddDate(0, 0, -30)

	_, err := service.InspectHabit(ctx, "NonExistent", from, today, today)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_DeleteLog(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	today := time.Now()

	// Create logs
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)
	err = service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	// Get the log IDs via inspect
	from := today.AddDate(0, 0, -30)
	details, err := service.InspectHabit(ctx, "Gym", from, today, today)
	require.NoError(t, err)
	require.Len(t, details.Logs, 2)

	firstLogID := details.Logs[0].ID

	// Delete the first log
	err = service.DeleteLog(ctx, firstLogID)
	require.NoError(t, err)

	// Verify only one log remains
	details, err = service.InspectHabit(ctx, "Gym", from, today, today)
	require.NoError(t, err)
	assert.Len(t, details.Logs, 1)
}

func TestHabitService_DeleteLog_NotFound(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	err := service.DeleteLog(ctx, 99999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHabitService_RenameHabit(t *testing.T) {
	service := setupHabitService(t)
	ctx := context.Background()

	// Create a habit
	err := service.LogHabit(ctx, "Gym", 1)
	require.NoError(t, err)

	// Rename it
	err = service.RenameHabit(ctx, "Gym", "Workout")
	require.NoError(t, err)

	// Verify old name doesn't exist
	today := time.Now()
	from := today.AddDate(0, 0, -30)
	_, err = service.InspectHabit(ctx, "Gym", from, today, today)
	require.Error(t, err)

	// Verify new name exists with the log
	details, err := service.InspectHabit(ctx, "Workout", from, today, today)
	require.NoError(t, err)
	assert.Equal(t, "Workout", details.Name)
	assert.Len(t, details.Logs, 1)
}

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
