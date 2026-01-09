package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestHabitRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := domain.Habit{
		Name:       "Gym",
		GoalPerDay: 1,
		CreatedAt:  time.Now(),
	}

	id, err := repo.Insert(ctx, habit)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestHabitRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := domain.Habit{
		Name:       "Meditation",
		GoalPerDay: 1,
		CreatedAt:  time.Now(),
	}
	id, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Meditation", result.Name)
	assert.Equal(t, 1, result.GoalPerDay)
}

func TestHabitRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := domain.Habit{
		Name:       "Water",
		GoalPerDay: 8,
		CreatedAt:  time.Now(),
	}
	_, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	result, err := repo.GetByName(ctx, "Water")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Water", result.Name)
	assert.Equal(t, 8, result.GoalPerDay)
}

func TestHabitRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	result, err := repo.GetByName(ctx, "NonExistent")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHabitRepository_GetOrCreate_Creates(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	result, err := repo.GetOrCreate(ctx, "NewHabit", 3)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "NewHabit", result.Name)
	assert.Equal(t, 3, result.GoalPerDay)
	assert.Greater(t, result.ID, int64(0))
}

func TestHabitRepository_GetOrCreate_GetsExisting(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := domain.Habit{
		Name:       "ExistingHabit",
		GoalPerDay: 5,
		CreatedAt:  time.Now(),
	}
	originalID, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	result, err := repo.GetOrCreate(ctx, "ExistingHabit", 10)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, originalID, result.ID)
	assert.Equal(t, 5, result.GoalPerDay) // Original goal, not new one
}

func TestHabitRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habits := []domain.Habit{
		{Name: "Gym", GoalPerDay: 1, CreatedAt: time.Now()},
		{Name: "Water", GoalPerDay: 8, CreatedAt: time.Now()},
		{Name: "Reading", GoalPerDay: 1, CreatedAt: time.Now()},
	}
	for _, h := range habits {
		_, err := repo.Insert(ctx, h)
		require.NoError(t, err)
	}

	results, err := repo.GetAll(ctx)

	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestHabitRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := domain.Habit{
		Name:       "ToDelete",
		GoalPerDay: 1,
		CreatedAt:  time.Now(),
	}
	id, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHabitRepository_Delete_SoftDeletes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := domain.Habit{
		Name:       "SoftDeleteTest",
		GoalPerDay: 1,
		CreatedAt:  time.Now(),
	}
	id, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Verify the habit is soft deleted (not visible via GetByID)
	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, result)

	// But the data should still exist in the database with a DELETE marker
	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM habits WHERE name = 'SoftDeleteTest'`).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1, "Soft deleted habit should still exist in DB")
}

func TestHabitRepository_Delete_CreatesDeleteMarker(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := domain.Habit{
		Name:       "DeleteMarkerTest",
		GoalPerDay: 2,
		CreatedAt:  time.Now(),
	}
	id, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	// Get the entity_id before delete
	h, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	entityID := h.EntityID

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Verify DELETE marker exists
	var opType string
	err = db.QueryRowContext(ctx, `
		SELECT op_type FROM habits
		WHERE entity_id = ?
		ORDER BY version DESC LIMIT 1
	`, entityID.String()).Scan(&opType)
	require.NoError(t, err)
	assert.Equal(t, "DELETE", opType)
}

func TestHabitRepository_GetDeleted_ReturnsDeletedHabits(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	// Create and delete a habit
	habit := domain.Habit{
		Name:       "DeletedHabit",
		GoalPerDay: 1,
		CreatedAt:  time.Now(),
	}
	id, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Get deleted habits
	deleted, err := repo.GetDeleted(ctx)
	require.NoError(t, err)
	require.Len(t, deleted, 1)
	assert.Equal(t, "DeletedHabit", deleted[0].Name)
}

func TestHabitRepository_Restore_BringsBackDeletedHabit(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHabitRepository(db)
	ctx := context.Background()

	// Create and delete a habit
	habit := domain.Habit{
		Name:       "RestoreTest",
		GoalPerDay: 3,
		CreatedAt:  time.Now(),
	}
	id, err := repo.Insert(ctx, habit)
	require.NoError(t, err)

	h, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	entityID := h.EntityID

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	// Verify it's gone
	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, result)

	// Restore it
	newID, err := repo.Restore(ctx, entityID)
	require.NoError(t, err)
	assert.NotZero(t, newID)

	// Verify it's back
	restored, err := repo.GetByID(ctx, newID)
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, "RestoreTest", restored.Name)
	assert.Equal(t, 3, restored.GoalPerDay)
}
