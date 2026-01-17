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

func TestApp_AddListItem_AddsItemToList(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	list, err := services.List.CreateList(ctx, "Test List")
	require.NoError(t, err)

	itemID, err := wailsApp.AddListItem(list.ID, "New item")
	require.NoError(t, err)
	assert.Greater(t, itemID, int64(0))

	lists, err := wailsApp.GetLists()
	require.NoError(t, err)
	require.Len(t, lists, 1)
	require.Len(t, lists[0].Items, 1)
	assert.Equal(t, "New item", lists[0].Items[0].Content)
}

func TestApp_MarkListItemDone_MarksItemAsDone(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	list, err := services.List.CreateList(ctx, "Test List")
	require.NoError(t, err)

	itemID, err := wailsApp.AddListItem(list.ID, "Task item")
	require.NoError(t, err)

	err = wailsApp.MarkListItemDone(itemID)
	require.NoError(t, err)

	lists, err := wailsApp.GetLists()
	require.NoError(t, err)
	require.Len(t, lists[0].Items, 1)
	assert.Equal(t, "done", string(lists[0].Items[0].Type))
}

func TestApp_MarkListItemUndone_MarksItemAsUndone(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	list, err := services.List.CreateList(ctx, "Test List")
	require.NoError(t, err)

	itemID, err := wailsApp.AddListItem(list.ID, "Task item")
	require.NoError(t, err)

	err = wailsApp.MarkListItemDone(itemID)
	require.NoError(t, err)

	err = wailsApp.MarkListItemUndone(itemID)
	require.NoError(t, err)

	lists, err := wailsApp.GetLists()
	require.NoError(t, err)
	require.Len(t, lists[0].Items, 1)
	assert.Equal(t, "task", string(lists[0].Items[0].Type))
}

func TestApp_RemoveListItem_RemovesItem(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	list, err := services.List.CreateList(ctx, "Test List")
	require.NoError(t, err)

	itemID, err := wailsApp.AddListItem(list.ID, "Task item")
	require.NoError(t, err)

	err = wailsApp.RemoveListItem(itemID)
	require.NoError(t, err)

	lists, err := wailsApp.GetLists()
	require.NoError(t, err)
	require.Len(t, lists[0].Items, 0)
}

func TestApp_CreateGoal_CreatesGoal(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	goalID, err := wailsApp.CreateGoal("Learn Go", month)
	require.NoError(t, err)
	assert.Greater(t, goalID, int64(0))

	goals, err := wailsApp.GetGoals(month)
	require.NoError(t, err)
	require.Len(t, goals, 1)
	assert.Equal(t, "Learn Go", goals[0].Content)
}

func TestApp_MarkGoalDone_MarksGoalAsDone(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	goalID, err := wailsApp.CreateGoal("Complete project", month)
	require.NoError(t, err)

	err = wailsApp.MarkGoalDone(goalID)
	require.NoError(t, err)

	goals, err := wailsApp.GetGoals(month)
	require.NoError(t, err)
	require.Len(t, goals, 1)
	assert.Equal(t, "done", string(goals[0].Status))
}

func TestApp_MarkGoalActive_ReactivatesGoal(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	goalID, err := wailsApp.CreateGoal("Complete project", month)
	require.NoError(t, err)

	err = wailsApp.MarkGoalDone(goalID)
	require.NoError(t, err)

	err = wailsApp.MarkGoalActive(goalID)
	require.NoError(t, err)

	goals, err := wailsApp.GetGoals(month)
	require.NoError(t, err)
	require.Len(t, goals, 1)
	assert.Equal(t, "active", string(goals[0].Status))
}

func TestApp_DeleteGoal_DeletesGoal(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	goalID, err := wailsApp.CreateGoal("Complete project", month)
	require.NoError(t, err)

	err = wailsApp.DeleteGoal(goalID)
	require.NoError(t, err)

	goals, err := wailsApp.GetGoals(month)
	require.NoError(t, err)
	assert.Empty(t, goals)
}

func TestApp_Search_ReturnsMatchingEntries(t *testing.T) {
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
	_, err = services.Bujo.LogEntries(ctx, ". Call dentist", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)
	_, err = services.Bujo.LogEntries(ctx, "- Meeting notes", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)

	results, err := wailsApp.Search("groceries")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Buy groceries", results[0].Content)
}

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
