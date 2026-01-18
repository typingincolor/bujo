package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupGoalService(t *testing.T) *GoalService {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	goalRepo := sqlite.NewGoalRepository(db)

	return NewGoalService(goalRepo)
}

func TestGoalService_CreateGoal(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	id, err := service.CreateGoal(ctx, "Learn Go", month)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestGoalService_GetGoal(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	id, _ := service.CreateGoal(ctx, "Read 12 books", month)

	goal, err := service.GetGoal(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, goal)
	assert.Equal(t, "Read 12 books", goal.Content)
	assert.Equal(t, domain.GoalStatusActive, goal.Status)
}

func TestGoalService_GetGoalsForMonth(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	_, _ = service.CreateGoal(ctx, "Jan Goal 1", jan)
	_, _ = service.CreateGoal(ctx, "Jan Goal 2", jan)
	_, _ = service.CreateGoal(ctx, "Feb Goal", feb)

	goals, err := service.GetGoalsForMonth(ctx, jan)

	require.NoError(t, err)
	assert.Len(t, goals, 2)
}

func TestGoalService_MarkDone(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	id, _ := service.CreateGoal(ctx, "Complete task", month)

	err := service.MarkDone(ctx, id)

	require.NoError(t, err)

	goal, _ := service.GetGoal(ctx, id)
	assert.True(t, goal.IsDone())
}

func TestGoalService_MarkActive(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	id, _ := service.CreateGoal(ctx, "Undo task", month)
	_ = service.MarkDone(ctx, id)

	err := service.MarkActive(ctx, id)

	require.NoError(t, err)

	goal, _ := service.GetGoal(ctx, id)
	assert.False(t, goal.IsDone())
}

func TestGoalService_MoveToMonth(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	id, _ := service.CreateGoal(ctx, "Move me", jan)

	err := service.MoveToMonth(ctx, id, feb)

	require.NoError(t, err)

	goal, _ := service.GetGoal(ctx, id)
	assert.Equal(t, "2026-02", goal.MonthKey())
}

func TestGoalService_DeleteGoal(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	id, _ := service.CreateGoal(ctx, "Delete me", month)

	err := service.DeleteGoal(ctx, id)

	require.NoError(t, err)

	goal, _ := service.GetGoal(ctx, id)
	assert.Nil(t, goal)
}

func TestGoalService_GetCurrentMonthGoals(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()

	// Get current month
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	_, _ = service.CreateGoal(ctx, "Current month goal", currentMonth)

	goals, err := service.GetCurrentMonthGoals(ctx)

	require.NoError(t, err)
	assert.Len(t, goals, 1)
}

func TestGoalService_MigrateGoal(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	id, _ := service.CreateGoal(ctx, "Migrate me", jan)

	newID, err := service.MigrateGoal(ctx, id, feb)

	require.NoError(t, err)
	assert.Greater(t, newID, int64(0))
	assert.NotEqual(t, id, newID)

	// Original goal should be marked as migrated
	original, _ := service.GetGoal(ctx, id)
	assert.True(t, original.IsMigrated())
	assert.NotNil(t, original.MigratedTo)
	assert.Equal(t, "2026-02", original.MigratedTo.Format("2006-01"))

	// New goal should exist in target month
	newGoal, _ := service.GetGoal(ctx, newID)
	assert.Equal(t, "Migrate me", newGoal.Content)
	assert.Equal(t, "2026-02", newGoal.MonthKey())
	assert.Equal(t, domain.GoalStatusActive, newGoal.Status)
}

func TestGoalService_UpdateGoal(t *testing.T) {
	service := setupGoalService(t)
	ctx := context.Background()
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	id, _ := service.CreateGoal(ctx, "Original content", month)

	err := service.UpdateGoal(ctx, id, "Updated content")

	require.NoError(t, err)

	goal, _ := service.GetGoal(ctx, id)
	assert.Equal(t, "Updated content", goal.Content)
}
