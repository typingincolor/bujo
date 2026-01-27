package wails

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func TestApp_Search_ReturnsEmptyForNoMatch(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	_, err = services.Bujo.LogEntries(ctx, ". Buy groceries", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)

	results, err := wailsApp.Search("xyz123")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestApp_EditEntry_UpdatesEntryContent(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Original content", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = wailsApp.EditEntry(ids[0], "Updated content")
	require.NoError(t, err)

	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	require.Len(t, days, 1)
	require.Len(t, days[0].Entries, 1)
	assert.Equal(t, "Updated content", days[0].Entries[0].Content)
}

func TestApp_DeleteEntry_RemovesEntry(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Task to delete", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = wailsApp.DeleteEntry(ids[0])
	require.NoError(t, err)

	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	require.Len(t, days, 1)
	assert.Empty(t, days[0].Entries)
}

func TestApp_HasChildren_ReturnsTrueForParent(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Parent\n  . Child", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 2)

	hasChildren, err := wailsApp.HasChildren(ids[0])
	require.NoError(t, err)
	assert.True(t, hasChildren)
}

func TestApp_HasChildren_ReturnsFalseForLeaf(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Leaf entry", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	hasChildren, err := wailsApp.HasChildren(ids[0])
	require.NoError(t, err)
	assert.False(t, hasChildren)
}

func TestApp_CreateHabit_CreatesNewHabit(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	habitID, err := wailsApp.CreateHabit("Morning Run")
	require.NoError(t, err)
	assert.Greater(t, habitID, int64(0))

	status, err := wailsApp.GetHabits(7)
	require.NoError(t, err)
	require.Len(t, status.Habits, 1)
	assert.Equal(t, "Morning Run", status.Habits[0].Name)
}

func TestApp_DeleteHabit_RemovesHabit(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	habitID, err := wailsApp.CreateHabit("To Delete")
	require.NoError(t, err)

	err = wailsApp.DeleteHabit(habitID)
	require.NoError(t, err)

	status, err := wailsApp.GetHabits(7)
	require.NoError(t, err)
	assert.Empty(t, status.Habits)
}

func TestApp_CancelEntry_CancelsTask(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Test task", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = wailsApp.CancelEntry(ids[0])
	require.NoError(t, err)

	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	require.Len(t, days, 1)
	require.Len(t, days[0].Entries, 1)
	assert.Equal(t, "✗", days[0].Entries[0].Type.Symbol())
}

func TestApp_UncancelEntry_RevertsToTask(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Test task", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = wailsApp.CancelEntry(ids[0])
	require.NoError(t, err)

	err = wailsApp.UncancelEntry(ids[0])
	require.NoError(t, err)

	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	require.Len(t, days, 1)
	require.Len(t, days[0].Entries, 1)
	assert.Equal(t, "•", days[0].Entries[0].Type.Symbol())
}

func TestApp_SetPriority_SetsPriorityOnEntry(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Test task", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = wailsApp.SetPriority(ids[0], "high")
	require.NoError(t, err)

	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	require.Len(t, days, 1)
	require.Len(t, days[0].Entries, 1)
	assert.Equal(t, domain.PriorityHigh, days[0].Entries[0].Priority)
}

func TestApp_CyclePriority_CyclesThroughPriorities(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := services.Bujo.LogEntries(ctx, ". Test task", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	// Initial priority is none, cycle to low
	err = wailsApp.CyclePriority(ids[0])
	require.NoError(t, err)

	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	assert.Equal(t, domain.PriorityLow, days[0].Entries[0].Priority)

	// Cycle to medium
	err = wailsApp.CyclePriority(ids[0])
	require.NoError(t, err)

	days, err = wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	assert.Equal(t, domain.PriorityMedium, days[0].Entries[0].Priority)
}

func TestApp_MigrateEntry_MovesTaskToFutureDate(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)
	ids, err := services.Bujo.LogEntries(ctx, ". Test task", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	newID, err := wailsApp.MigrateEntry(ids[0], tomorrow)
	require.NoError(t, err)
	assert.Greater(t, newID, int64(0))

	// Original entry should be marked as migrated
	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	require.Len(t, days, 1)
	require.Len(t, days[0].Entries, 1)
	assert.Equal(t, "→", days[0].Entries[0].Type.Symbol())

	// New entry should exist on tomorrow
	days, err = wailsApp.GetDayEntries(tomorrow, tomorrow)
	require.NoError(t, err)
	require.Len(t, days, 1)
	require.Len(t, days[0].Entries, 1)
	assert.Equal(t, "Test task", days[0].Entries[0].Content)
}

func TestApp_SetLocation_SetsLocationForDate(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	err = wailsApp.SetLocation(today, "Manchester Office")
	require.NoError(t, err)

	// Verify location is set by checking the days
	days, err := wailsApp.GetDayEntries(today, today)
	require.NoError(t, err)
	require.Len(t, days, 1)
	require.NotNil(t, days[0].Location)
	assert.Equal(t, "Manchester Office", *days[0].Location)
}

func TestApp_GetLocationHistory_ReturnsUniqueLocations(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	dayBefore := today.AddDate(0, 0, -2)

	// Set locations for different days, with a duplicate
	err = wailsApp.SetLocation(today, "Manchester Office")
	require.NoError(t, err)
	err = wailsApp.SetLocation(yesterday, "Home")
	require.NoError(t, err)
	err = wailsApp.SetLocation(dayBefore, "Manchester Office") // Duplicate
	require.NoError(t, err)

	locations, err := wailsApp.GetLocationHistory()
	require.NoError(t, err)

	// Should have 2 unique locations
	assert.Len(t, locations, 2)
	assert.Contains(t, locations, "Manchester Office")
	assert.Contains(t, locations, "Home")
}
