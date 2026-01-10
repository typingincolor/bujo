package service

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type StatsEntryRepository interface {
	GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error)
}

type StatsHabitRepository interface {
	GetAll(ctx context.Context) ([]domain.Habit, error)
}

type StatsHabitLogRepository interface {
	GetAllRange(ctx context.Context, start, end time.Time) ([]domain.HabitLog, error)
	GetByHabitID(ctx context.Context, habitID int64) ([]domain.HabitLog, error)
}

type StatsService struct {
	entryRepo    StatsEntryRepository
	habitRepo    StatsHabitRepository
	habitLogRepo StatsHabitLogRepository
}

func NewStatsService(
	entryRepo StatsEntryRepository,
	habitRepo StatsHabitRepository,
	habitLogRepo StatsHabitLogRepository,
) *StatsService {
	return &StatsService{
		entryRepo:    entryRepo,
		habitRepo:    habitRepo,
		habitLogRepo: habitLogRepo,
	}
}

func (s *StatsService) GetStats(ctx context.Context, from, to time.Time) (*domain.Stats, error) {
	entries, err := s.entryRepo.GetByDateRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	habits, err := s.habitRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	logs, err := s.habitLogRepo.GetAllRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	stats := &domain.Stats{
		Period: domain.StatsPeriod{
			From: from,
			To:   to,
		},
		TotalDays: int(to.Sub(from).Hours()/24) + 1,
	}

	stats.EntryCounts = s.countEntries(entries)
	stats.TaskCompletion = s.calculateTaskCompletion(entries)
	stats.Productivity = s.calculateProductivity(entries)
	stats.HabitStats = s.calculateHabitStats(ctx, habits, logs, to)

	return stats, nil
}

func (s *StatsService) countEntries(entries []domain.Entry) domain.EntryCounts {
	counts := domain.EntryCounts{
		Total: len(entries),
	}

	for _, e := range entries {
		switch e.Type {
		case domain.EntryTypeTask:
			counts.Tasks++
		case domain.EntryTypeNote:
			counts.Notes++
		case domain.EntryTypeEvent:
			counts.Events++
		case domain.EntryTypeDone:
			counts.Done++
		case domain.EntryTypeMigrated:
			counts.Migrated++
		case domain.EntryTypeCancelled:
			counts.Cancelled++
		}
	}

	return counts
}

func (s *StatsService) calculateTaskCompletion(entries []domain.Entry) domain.TaskCompletion {
	var total, completed int

	for _, e := range entries {
		switch e.Type {
		case domain.EntryTypeTask:
			total++
		case domain.EntryTypeDone, domain.EntryTypeMigrated:
			total++
			completed++
		}
	}

	rate := 0.0
	if total > 0 {
		rate = float64(completed) / float64(total) * 100
	}

	return domain.TaskCompletion{
		Total:     total,
		Completed: completed,
		Rate:      rate,
	}
}

func (s *StatsService) calculateProductivity(entries []domain.Entry) domain.Productivity {
	prod := domain.Productivity{
		EntriesByDay: make(map[time.Weekday]int),
	}

	dayCount := make(map[time.Weekday]int)

	for _, e := range entries {
		if e.ScheduledDate == nil {
			continue
		}
		day := e.ScheduledDate.Weekday()
		prod.EntriesByDay[day]++

		dayKey := e.ScheduledDate.Format("2006-01-02")
		if _, seen := dayCount[day]; !seen {
			dayCount[day] = 0
		}
		_ = dayKey
	}

	// Calculate unique days per weekday
	uniqueDays := make(map[time.Weekday]map[string]bool)
	for _, e := range entries {
		if e.ScheduledDate == nil {
			continue
		}
		day := e.ScheduledDate.Weekday()
		if uniqueDays[day] == nil {
			uniqueDays[day] = make(map[string]bool)
		}
		uniqueDays[day][e.ScheduledDate.Format("2006-01-02")] = true
	}

	// Calculate averages
	var totalEntries, totalDays int
	var maxAvg, minAvg float64 = -1, -1
	var mostDay, leastDay time.Weekday

	for day, count := range prod.EntriesByDay {
		numDays := len(uniqueDays[day])
		if numDays == 0 {
			continue
		}
		avg := float64(count) / float64(numDays)
		totalEntries += count
		totalDays += numDays

		if maxAvg < 0 || avg > maxAvg {
			maxAvg = avg
			mostDay = day
		}
		if minAvg < 0 || avg < minAvg {
			minAvg = avg
			leastDay = day
		}
	}

	if totalDays > 0 {
		prod.AveragePerDay = float64(totalEntries) / float64(totalDays)
	}

	if maxAvg >= 0 {
		prod.MostProductive = domain.Weekday{Day: mostDay, Average: maxAvg}
	}
	if minAvg >= 0 {
		prod.LeastProductive = domain.Weekday{Day: leastDay, Average: minAvg}
	}

	return prod
}

func (s *StatsService) calculateHabitStats(ctx context.Context, habits []domain.Habit, logs []domain.HabitLog, today time.Time) domain.HabitStats {
	stats := domain.HabitStats{
		Active:    len(habits),
		TotalLogs: len(logs),
	}

	if len(habits) == 0 {
		return stats
	}

	// Calculate log counts per habit
	logCounts := make(map[int64]int)
	for _, l := range logs {
		logCounts[l.HabitID] += l.Count
	}

	// Find most logged and best streak
	var maxCount int
	var maxStreak int

	for _, h := range habits {
		count := logCounts[h.ID]
		if count > maxCount {
			maxCount = count
			stats.MostLogged = domain.HabitLogCount{
				HabitName: h.Name,
				Count:     count,
			}
		}

		// Get all logs for this habit to calculate streak
		habitLogs, err := s.habitLogRepo.GetByHabitID(ctx, h.ID)
		if err != nil {
			continue
		}

		streak := domain.CalculateStreak(habitLogs, today)
		if streak > maxStreak {
			maxStreak = streak
			stats.BestStreak = domain.HabitStreak{
				HabitName: h.Name,
				Days:      streak,
			}
		}
	}

	return stats
}
