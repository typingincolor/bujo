package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestBujoService_GetMultiDayAgenda_ReturnsEntriesGroupedByDate(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	_, err := service.LogEntries(ctx, ". Task on day 1", LogEntriesOptions{Date: day1})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Task on day 2", LogEntriesOptions{Date: day2})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Task on day 3", LogEntriesOptions{Date: day3})
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day3)

	require.NoError(t, err)
	require.Len(t, agenda.Days, 3)
	assert.Len(t, agenda.Days[0].Entries, 1)
	assert.Len(t, agenda.Days[1].Entries, 1)
	assert.Len(t, agenda.Days[2].Entries, 1)
}

func TestBujoService_GetMultiDayAgenda_DoesNotIncludeOverdue(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	_, err := service.LogEntries(ctx, ". Task on day 1", LogEntriesOptions{Date: day1})
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day2)

	require.NoError(t, err)
	require.Len(t, agenda.Days, 2)
	assert.Len(t, agenda.Days[0].Entries, 1)
}

func TestBujoService_GetMultiDayAgenda_IncludesLocations(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, day1, "Home")
	require.NoError(t, err)
	err = service.SetLocation(ctx, day2, "Office")
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day2)

	require.NoError(t, err)
	require.Len(t, agenda.Days, 2)
	require.NotNil(t, agenda.Days[0].Location)
	assert.Equal(t, "Home", *agenda.Days[0].Location)
	require.NotNil(t, agenda.Days[1].Location)
	assert.Equal(t, "Office", *agenda.Days[1].Location)
}

func TestBujoService_GetMultiDayAgenda_EmptyRange(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day2)

	require.NoError(t, err)
	assert.Len(t, agenda.Days, 2)
	assert.Empty(t, agenda.Days[0].Entries)
	assert.Empty(t, agenda.Days[1].Entries)
}

func TestBujoService_GetMultiDayAgenda_IncludesOverdueEntries(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	_, err := service.LogEntries(ctx, ". Overdue task", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Current task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, today, tomorrow)

	require.NoError(t, err)
	require.Len(t, agenda.Days, 2)
	assert.Len(t, agenda.Days[0].Entries, 1, "Current task should be in Days")
	assert.Len(t, agenda.Overdue, 1, "Overdue task should be in Overdue field")
	assert.Equal(t, "Overdue task", agenda.Overdue[0].Content)
}

// Mood tracking tests

func TestBujoService_SetMood(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	err := service.SetMood(ctx, today, "happy")

	require.NoError(t, err)
}

func TestBujoService_GetMood(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	err := service.SetMood(ctx, today, "energetic")
	require.NoError(t, err)

	mood, err := service.GetMood(ctx, today)

	require.NoError(t, err)
	require.NotNil(t, mood)
	assert.Equal(t, "energetic", *mood)
}

func TestBujoService_GetMood_NotSet(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	mood, err := service.GetMood(ctx, today)

	require.NoError(t, err)
	assert.Nil(t, mood)
}

func TestBujoService_GetMoodHistory(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	err := service.SetMood(ctx, day1, "happy")
	require.NoError(t, err)
	err = service.SetMood(ctx, day2, "tired")
	require.NoError(t, err)
	err = service.SetMood(ctx, day3, "focused")
	require.NoError(t, err)

	history, err := service.GetMoodHistory(ctx, day1, day3)

	require.NoError(t, err)
	assert.Len(t, history, 3)
}

func TestBujoService_ClearMood(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	err := service.SetMood(ctx, today, "happy")
	require.NoError(t, err)

	err = service.ClearMood(ctx, today)
	require.NoError(t, err)

	mood, err := service.GetMood(ctx, today)
	require.NoError(t, err)
	assert.Nil(t, mood)
}

// Weather tracking tests

func TestBujoService_SetWeather(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	err := service.SetWeather(ctx, today, "sunny")

	require.NoError(t, err)
}

func TestBujoService_GetWeather(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	err := service.SetWeather(ctx, today, "rainy, 15°C")
	require.NoError(t, err)

	weather, err := service.GetWeather(ctx, today)

	require.NoError(t, err)
	require.NotNil(t, weather)
	assert.Equal(t, "rainy, 15°C", *weather)
}

func TestBujoService_GetWeather_NotSet(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	weather, err := service.GetWeather(ctx, today)

	require.NoError(t, err)
	assert.Nil(t, weather)
}

func TestBujoService_GetWeatherHistory(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	err := service.SetWeather(ctx, day1, "sunny")
	require.NoError(t, err)
	err = service.SetWeather(ctx, day2, "cloudy")
	require.NoError(t, err)
	err = service.SetWeather(ctx, day3, "rainy")
	require.NoError(t, err)

	history, err := service.GetWeatherHistory(ctx, day1, day3)

	require.NoError(t, err)
	assert.Len(t, history, 3)
}

func TestBujoService_ClearWeather(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	err := service.SetWeather(ctx, today, "sunny")
	require.NoError(t, err)

	err = service.ClearWeather(ctx, today)
	require.NoError(t, err)

	weather, err := service.GetWeather(ctx, today)
	require.NoError(t, err)
	assert.Nil(t, weather)
}

func TestBujoService_GetOutstandingTasks(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	// Add a mix of entry types
	input := `. Task 1
- Note (not a task)
o Event (not a task)
x Done task (completed)
> Migrated task`

	_, err := service.LogEntries(ctx, input, LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Add task from yesterday
	_, err = service.LogEntries(ctx, ". Yesterday task", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)

	// Get outstanding tasks for today only
	tasks, err := service.GetOutstandingTasks(ctx, today, today)
	require.NoError(t, err)

	// Should only get "Task 1" (not note, event, done, or migrated)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "Task 1", tasks[0].Content)
	assert.Equal(t, domain.EntryTypeTask, tasks[0].Type)
}

func TestBujoService_GetOutstandingTasks_DateRange(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	// Add tasks on different days
	_, err := service.LogEntries(ctx, ". Task today", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Task yesterday", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Task old", LogEntriesOptions{Date: twoDaysAgo})
	require.NoError(t, err)

	// Get tasks from yesterday to today (should exclude 2 days ago)
	tasks, err := service.GetOutstandingTasks(ctx, yesterday, today)
	require.NoError(t, err)

	assert.Len(t, tasks, 2)
}

func TestBujoService_GetOutstandingTasks_Empty(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	// No entries at all
	tasks, err := service.GetOutstandingTasks(ctx, today, today)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

