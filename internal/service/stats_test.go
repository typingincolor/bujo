package service

import (
	"context"
	"testing"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type mockStatsEntryRepo struct {
	entries []domain.Entry
}

func (m *mockStatsEntryRepo) GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
	var result []domain.Entry
	for _, e := range m.entries {
		if e.ScheduledDate != nil {
			if !e.ScheduledDate.Before(from) && !e.ScheduledDate.After(to) {
				result = append(result, e)
			}
		}
	}
	return result, nil
}

type mockStatsHabitRepo struct {
	habits []domain.Habit
}

func (m *mockStatsHabitRepo) GetAll(ctx context.Context) ([]domain.Habit, error) {
	return m.habits, nil
}

type mockStatsHabitLogRepo struct {
	logs []domain.HabitLog
}

func (m *mockStatsHabitLogRepo) GetAllRange(ctx context.Context, start, end time.Time) ([]domain.HabitLog, error) {
	var result []domain.HabitLog
	for _, l := range m.logs {
		if !l.LoggedAt.Before(start) && !l.LoggedAt.After(end) {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockStatsHabitLogRepo) GetByHabitID(ctx context.Context, habitID int64) ([]domain.HabitLog, error) {
	var result []domain.HabitLog
	for _, l := range m.logs {
		if l.HabitID == habitID {
			result = append(result, l)
		}
	}
	return result, nil
}

func TestStatsService_GetStats_EntryCounts(t *testing.T) {
	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	from := today.AddDate(0, 0, -29)

	entries := []domain.Entry{
		{ID: 1, Type: domain.EntryTypeTask, Content: "Task 1", ScheduledDate: &today},
		{ID: 2, Type: domain.EntryTypeTask, Content: "Task 2", ScheduledDate: &today},
		{ID: 3, Type: domain.EntryTypeNote, Content: "Note 1", ScheduledDate: &today},
		{ID: 4, Type: domain.EntryTypeEvent, Content: "Event 1", ScheduledDate: &today},
		{ID: 5, Type: domain.EntryTypeDone, Content: "Done 1", ScheduledDate: &today},
	}

	svc := NewStatsService(
		&mockStatsEntryRepo{entries: entries},
		&mockStatsHabitRepo{},
		&mockStatsHabitLogRepo{},
	)

	stats, err := svc.GetStats(context.Background(), from, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.EntryCounts.Total != 5 {
		t.Errorf("expected total 5, got %d", stats.EntryCounts.Total)
	}
	if stats.EntryCounts.Tasks != 2 {
		t.Errorf("expected 2 tasks, got %d", stats.EntryCounts.Tasks)
	}
	if stats.EntryCounts.Notes != 1 {
		t.Errorf("expected 1 note, got %d", stats.EntryCounts.Notes)
	}
	if stats.EntryCounts.Events != 1 {
		t.Errorf("expected 1 event, got %d", stats.EntryCounts.Events)
	}
	if stats.EntryCounts.Done != 1 {
		t.Errorf("expected 1 done, got %d", stats.EntryCounts.Done)
	}
}

func TestStatsService_GetStats_TaskCompletion(t *testing.T) {
	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	from := today.AddDate(0, 0, -29)

	entries := []domain.Entry{
		{ID: 1, Type: domain.EntryTypeTask, Content: "Task 1", ScheduledDate: &today},
		{ID: 2, Type: domain.EntryTypeTask, Content: "Task 2", ScheduledDate: &today},
		{ID: 3, Type: domain.EntryTypeTask, Content: "Task 3", ScheduledDate: &today},
		{ID: 4, Type: domain.EntryTypeDone, Content: "Done 1", ScheduledDate: &today},
		{ID: 5, Type: domain.EntryTypeDone, Content: "Done 2", ScheduledDate: &today},
		{ID: 6, Type: domain.EntryTypeMigrated, Content: "Migrated 1", ScheduledDate: &today},
	}

	svc := NewStatsService(
		&mockStatsEntryRepo{entries: entries},
		&mockStatsHabitRepo{},
		&mockStatsHabitLogRepo{},
	)

	stats, err := svc.GetStats(context.Background(), from, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total tasks = task + done + migrated = 3 + 2 + 1 = 6
	// Completed = done + migrated = 2 + 1 = 3
	if stats.TaskCompletion.Total != 6 {
		t.Errorf("expected total tasks 6, got %d", stats.TaskCompletion.Total)
	}
	if stats.TaskCompletion.Completed != 3 {
		t.Errorf("expected completed 3, got %d", stats.TaskCompletion.Completed)
	}
	if stats.TaskCompletion.Rate != 50.0 {
		t.Errorf("expected rate 50.0, got %.1f", stats.TaskCompletion.Rate)
	}
}

func TestStatsService_GetStats_Productivity(t *testing.T) {
	monday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC) // Monday
	tuesday := monday.AddDate(0, 0, 1)
	wednesday := monday.AddDate(0, 0, 2)
	from := monday.AddDate(0, 0, -1)
	to := wednesday

	entries := []domain.Entry{
		{ID: 1, Type: domain.EntryTypeTask, Content: "Mon 1", ScheduledDate: &monday},
		{ID: 2, Type: domain.EntryTypeTask, Content: "Mon 2", ScheduledDate: &monday},
		{ID: 3, Type: domain.EntryTypeTask, Content: "Mon 3", ScheduledDate: &monday},
		{ID: 4, Type: domain.EntryTypeTask, Content: "Tue 1", ScheduledDate: &tuesday},
		{ID: 5, Type: domain.EntryTypeTask, Content: "Wed 1", ScheduledDate: &wednesday},
	}

	svc := NewStatsService(
		&mockStatsEntryRepo{entries: entries},
		&mockStatsHabitRepo{},
		&mockStatsHabitLogRepo{},
	)

	stats, err := svc.GetStats(context.Background(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Productivity.MostProductive.Day != time.Monday {
		t.Errorf("expected most productive Monday, got %v", stats.Productivity.MostProductive.Day)
	}
	if stats.Productivity.MostProductive.Average != 3.0 {
		t.Errorf("expected avg 3.0 for Monday, got %.1f", stats.Productivity.MostProductive.Average)
	}
}

func TestStatsService_GetStats_HabitStats(t *testing.T) {
	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	from := today.AddDate(0, 0, -29)

	habits := []domain.Habit{
		{ID: 1, EntityID: "h1", Name: "Exercise"},
		{ID: 2, EntityID: "h2", Name: "Read"},
		{ID: 3, EntityID: "h3", Name: "Meditate"},
	}

	logs := []domain.HabitLog{
		{ID: 1, HabitID: 1, Count: 1, LoggedAt: today},
		{ID: 2, HabitID: 1, Count: 1, LoggedAt: today.AddDate(0, 0, -1)},
		{ID: 3, HabitID: 1, Count: 1, LoggedAt: today.AddDate(0, 0, -2)},
		{ID: 4, HabitID: 2, Count: 1, LoggedAt: today},
		{ID: 5, HabitID: 2, Count: 2, LoggedAt: today.AddDate(0, 0, -1)},
		{ID: 6, HabitID: 2, Count: 1, LoggedAt: today.AddDate(0, 0, -2)},
		{ID: 7, HabitID: 2, Count: 1, LoggedAt: today.AddDate(0, 0, -3)},
		{ID: 8, HabitID: 2, Count: 1, LoggedAt: today.AddDate(0, 0, -4)},
	}

	svc := NewStatsService(
		&mockStatsEntryRepo{},
		&mockStatsHabitRepo{habits: habits},
		&mockStatsHabitLogRepo{logs: logs},
	)

	stats, err := svc.GetStats(context.Background(), from, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.HabitStats.Active != 3 {
		t.Errorf("expected 3 active habits, got %d", stats.HabitStats.Active)
	}
	if stats.HabitStats.TotalLogs != 8 {
		t.Errorf("expected 8 total logs, got %d", stats.HabitStats.TotalLogs)
	}
	// Read has 5 log entries (most logged)
	if stats.HabitStats.MostLogged.HabitName != "Read" {
		t.Errorf("expected most logged habit 'Read', got '%s'", stats.HabitStats.MostLogged.HabitName)
	}
	if stats.HabitStats.MostLogged.Count != 6 {
		t.Errorf("expected most logged count 6, got %d", stats.HabitStats.MostLogged.Count)
	}
	// Read has 5 day streak (longest)
	if stats.HabitStats.BestStreak.HabitName != "Read" {
		t.Errorf("expected best streak habit 'Read', got '%s'", stats.HabitStats.BestStreak.HabitName)
	}
	if stats.HabitStats.BestStreak.Days != 5 {
		t.Errorf("expected streak 5 days, got %d", stats.HabitStats.BestStreak.Days)
	}
}

func TestStatsService_GetStats_EmptyData(t *testing.T) {
	today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	from := today.AddDate(0, 0, -29)

	svc := NewStatsService(
		&mockStatsEntryRepo{},
		&mockStatsHabitRepo{},
		&mockStatsHabitLogRepo{},
	)

	stats, err := svc.GetStats(context.Background(), from, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.EntryCounts.Total != 0 {
		t.Errorf("expected 0 entries, got %d", stats.EntryCounts.Total)
	}
	if stats.HabitStats.Active != 0 {
		t.Errorf("expected 0 habits, got %d", stats.HabitStats.Active)
	}
}
