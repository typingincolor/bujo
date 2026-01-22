package wails

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/service"
)

func TestNewApp_AcceptsServices(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)

	assert.NotNil(t, wailsApp)
	assert.NotNil(t, wailsApp.services)
}

func TestApp_Startup_StoresContext(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	assert.NotNil(t, wailsApp.ctx)
}

func TestApp_GetAgenda_ReturnsMultiDayAgenda(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	agenda, err := wailsApp.GetAgenda(today, today.AddDate(0, 0, 7))

	require.NoError(t, err)
	assert.NotNil(t, agenda)
	assert.NotNil(t, agenda.Days)
}

func TestApp_GetHabits_ReturnsTrackerStatus(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	status, err := wailsApp.GetHabits(7)

	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.NotNil(t, status.Habits)
}

func TestApp_GetLists_ReturnsAllLists(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	lists, err := wailsApp.GetLists()

	require.NoError(t, err)
	assert.NotNil(t, lists)
}

func TestApp_GetGoals_ReturnsGoalsForMonth(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	goals, err := wailsApp.GetGoals(month)

	require.NoError(t, err)
	assert.Empty(t, goals)
}

func TestApp_MarkEntryDone_MarksTaskAsDone(t *testing.T) {
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

	err = wailsApp.MarkEntryDone(ids[0])
	require.NoError(t, err)

	agenda, err := wailsApp.GetAgenda(today, today)
	require.NoError(t, err)
	require.Len(t, agenda.Days, 1)
	require.Len(t, agenda.Days[0].Entries, 1)
	assert.Equal(t, "✓", agenda.Days[0].Entries[0].Type.Symbol())
}

func TestApp_MarkEntryUndone_RevertsToTask(t *testing.T) {
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

	err = wailsApp.MarkEntryDone(ids[0])
	require.NoError(t, err)

	err = wailsApp.MarkEntryUndone(ids[0])
	require.NoError(t, err)

	agenda, err := wailsApp.GetAgenda(today, today)
	require.NoError(t, err)
	require.Len(t, agenda.Days, 1)
	require.Len(t, agenda.Days[0].Entries, 1)
	assert.Equal(t, "•", agenda.Days[0].Entries[0].Type.Symbol())
}

func TestApp_AddEntry_CreatesNewEntry(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := wailsApp.AddEntry(". New task from desktop", today)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	agenda, err := wailsApp.GetAgenda(today, today)
	require.NoError(t, err)
	require.Len(t, agenda.Days, 1)
	require.Len(t, agenda.Days[0].Entries, 1)
	assert.Equal(t, "New task from desktop", agenda.Days[0].Entries[0].Content)
}

func TestApp_LogHabit_LogsHabitByID(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	err = services.Habit.LogHabit(ctx, "exercise", 1)
	require.NoError(t, err)

	status, err := wailsApp.GetHabits(7)
	require.NoError(t, err)
	require.Len(t, status.Habits, 1)
	habitID := status.Habits[0].ID

	err = wailsApp.LogHabit(habitID, 1)
	require.NoError(t, err)

	status, err = wailsApp.GetHabits(7)
	require.NoError(t, err)
	require.Len(t, status.Habits, 1)
	assert.Equal(t, 2, status.Habits[0].TodayCount)
}
