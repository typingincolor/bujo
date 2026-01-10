package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestGoalRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGoalRepository(db)
	ctx := context.Background()

	goal := domain.Goal{
		Content:   "Learn Go",
		Month:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:    domain.GoalStatusActive,
		CreatedAt: time.Now(),
	}

	id, err := repo.Insert(ctx, goal)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestGoalRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGoalRepository(db)
	ctx := context.Background()

	goal := domain.Goal{
		Content:   "Read 12 books",
		Month:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:    domain.GoalStatusActive,
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, goal)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Read 12 books", result.Content)
	assert.Equal(t, domain.GoalStatusActive, result.Status)
}

func TestGoalRepository_GetByMonth(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGoalRepository(db)
	ctx := context.Background()

	// Create goals for different months
	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	_, err := repo.Insert(ctx, domain.Goal{Content: "Jan Goal 1", Month: jan, Status: domain.GoalStatusActive, CreatedAt: time.Now()})
	require.NoError(t, err)
	_, err = repo.Insert(ctx, domain.Goal{Content: "Jan Goal 2", Month: jan, Status: domain.GoalStatusActive, CreatedAt: time.Now()})
	require.NoError(t, err)
	_, err = repo.Insert(ctx, domain.Goal{Content: "Feb Goal", Month: feb, Status: domain.GoalStatusActive, CreatedAt: time.Now()})
	require.NoError(t, err)

	goals, err := repo.GetByMonth(ctx, jan)

	require.NoError(t, err)
	assert.Len(t, goals, 2)
}

func TestGoalRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGoalRepository(db)
	ctx := context.Background()

	goal := domain.Goal{
		Content:   "Original",
		Month:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:    domain.GoalStatusActive,
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, goal)
	require.NoError(t, err)

	// Update status to done
	result, _ := repo.GetByID(ctx, id)
	result.Status = domain.GoalStatusDone

	err = repo.Update(ctx, *result)
	require.NoError(t, err)

	// Verify update
	updated, _ := repo.GetByID(ctx, id)
	assert.Equal(t, domain.GoalStatusDone, updated.Status)
}

func TestGoalRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGoalRepository(db)
	ctx := context.Background()

	goal := domain.Goal{
		Content:   "To Delete",
		Month:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:    domain.GoalStatusActive,
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, goal)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Should not be found
	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestGoalRepository_MoveToMonth(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGoalRepository(db)
	ctx := context.Background()

	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	goal := domain.Goal{
		Content:   "Move Me",
		Month:     jan,
		Status:    domain.GoalStatusActive,
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, goal)
	require.NoError(t, err)

	err = repo.MoveToMonth(ctx, id, feb)
	require.NoError(t, err)

	// Verify move
	result, _ := repo.GetByID(ctx, id)
	assert.Equal(t, "2026-02", result.MonthKey())
}
