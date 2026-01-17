package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceFactory_Create_ReturnsAllServices(t *testing.T) {
	ctx := context.Background()

	factory := NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	assert.NotNil(t, services.DB, "DB should be available")
	assert.NotNil(t, services.Bujo, "BujoService should be created")
	assert.NotNil(t, services.Habit, "HabitService should be created")
	assert.NotNil(t, services.List, "ListService should be created")
	assert.NotNil(t, services.Goal, "GoalService should be created")
	assert.NotNil(t, services.Stats, "StatsService should be created")
}

func TestServiceFactory_Create_WithInvalidPath_ReturnsError(t *testing.T) {
	ctx := context.Background()

	factory := NewServiceFactory()
	_, _, err := factory.Create(ctx, "/nonexistent/path/to/db.db")
	assert.Error(t, err)
}

func TestServiceFactory_Create_Cleanup_ClosesDatabase(t *testing.T) {
	ctx := context.Background()

	factory := NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	require.NotNil(t, services)

	cleanup()
}
