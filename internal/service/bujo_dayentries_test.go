package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBujoService_GetDayEntries_ReturnsEntriesGroupedByDate(t *testing.T) {
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

	days, err := service.GetDayEntries(ctx, day1, day3)

	require.NoError(t, err)
	require.Len(t, days, 3)
	assert.Len(t, days[0].Entries, 1)
	assert.Len(t, days[1].Entries, 1)
	assert.Len(t, days[2].Entries, 1)
}

func TestBujoService_GetDayEntries_IncludesLocations(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, day1, "Home")
	require.NoError(t, err)
	err = service.SetLocation(ctx, day2, "Office")
	require.NoError(t, err)

	days, err := service.GetDayEntries(ctx, day1, day2)

	require.NoError(t, err)
	require.Len(t, days, 2)
	require.NotNil(t, days[0].Location)
	assert.Equal(t, "Home", *days[0].Location)
	require.NotNil(t, days[1].Location)
	assert.Equal(t, "Office", *days[1].Location)
}

func TestBujoService_GetDayEntries_EmptyRange(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	days, err := service.GetDayEntries(ctx, day1, day2)

	require.NoError(t, err)
	assert.Len(t, days, 2)
	assert.Empty(t, days[0].Entries)
	assert.Empty(t, days[1].Entries)
}

func TestBujoService_GetDayEntries_SingleDay(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	_, err := service.LogEntries(ctx, ". Task for today", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	days, err := service.GetDayEntries(ctx, today, today)

	require.NoError(t, err)
	require.Len(t, days, 1)
	assert.Len(t, days[0].Entries, 1)
	assert.Equal(t, "Task for today", days[0].Entries[0].Content)
}

func TestBujoService_GetOverdue_ReturnsOverdueEntries(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	_, err := service.LogEntries(ctx, ". Overdue task 1", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Overdue task 2", LogEntriesOptions{Date: twoDaysAgo})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Current task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	overdue, err := service.GetOverdue(ctx)

	require.NoError(t, err)
	assert.Len(t, overdue, 2, "Should return only past-due tasks, not today's")
}

func TestBujoService_GetOverdue_ExcludesCompletedTasks(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	ids, err := service.LogEntries(ctx, ". Task to complete", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)
	_, err = service.LogEntries(ctx, ". Incomplete task", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)

	err = service.MarkDone(ctx, ids[0])
	require.NoError(t, err)

	overdue, err := service.GetOverdue(ctx)

	require.NoError(t, err)
	assert.Len(t, overdue, 1, "Should exclude completed tasks")
	assert.Equal(t, "Incomplete task", overdue[0].Content)
}

func TestBujoService_GetOverdue_Empty(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	overdue, err := service.GetOverdue(ctx)

	require.NoError(t, err)
	assert.Empty(t, overdue)
}
