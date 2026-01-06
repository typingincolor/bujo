package service

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type HabitRepository interface {
	GetOrCreate(ctx context.Context, name string, goalPerDay int) (*domain.Habit, error)
	GetAll(ctx context.Context) ([]domain.Habit, error)
}

type HabitLogRepository interface {
	Insert(ctx context.Context, log domain.HabitLog) (int64, error)
	GetRange(ctx context.Context, habitID int64, start, end time.Time) ([]domain.HabitLog, error)
}

type HabitService struct {
	habitRepo HabitRepository
	logRepo   HabitLogRepository
}

func NewHabitService(habitRepo HabitRepository, logRepo HabitLogRepository) *HabitService {
	return &HabitService{
		habitRepo: habitRepo,
		logRepo:   logRepo,
	}
}

func (s *HabitService) LogHabit(ctx context.Context, name string, count int) error {
	return s.LogHabitForDate(ctx, name, count, time.Now())
}

func (s *HabitService) LogHabitForDate(ctx context.Context, name string, count int, date time.Time) error {
	habit, err := s.habitRepo.GetOrCreate(ctx, name, 1)
	if err != nil {
		return err
	}

	log := domain.HabitLog{
		HabitID:  habit.ID,
		Count:    count,
		LoggedAt: date,
	}

	_, err = s.logRepo.Insert(ctx, log)
	return err
}

type HabitStatus struct {
	Name              string
	GoalPerDay        int
	CurrentStreak     int
	CompletionPercent float64
	TodayCount        int
	DayHistory        []DayStatus
}

type DayStatus struct {
	Date      time.Time
	Completed bool
	Count     int
}

type TrackerStatus struct {
	Habits []HabitStatus
}

func (s *HabitService) GetTrackerStatus(ctx context.Context, today time.Time, days int) (*TrackerStatus, error) {
	habits, err := s.habitRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	status := &TrackerStatus{
		Habits: make([]HabitStatus, 0, len(habits)),
	}

	// Normalize to start of day for consistent range queries
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	startDate := todayStart.AddDate(0, 0, -(days - 1))
	endDate := todayStart.Add(24 * time.Hour)

	for _, habit := range habits {
		logs, err := s.logRepo.GetRange(ctx, habit.ID, startDate, endDate)
		if err != nil {
			return nil, err
		}

		habitStatus := HabitStatus{
			Name:              habit.Name,
			GoalPerDay:        habit.GoalPerDay,
			CurrentStreak:     domain.CalculateStreak(logs, todayStart),
			CompletionPercent: domain.CalculateCompletion(logs, days, todayStart),
			TodayCount:        domain.SumCountForDay(logs, todayStart),
			DayHistory:        buildDayHistory(logs, todayStart, days),
		}

		status.Habits = append(status.Habits, habitStatus)
	}

	return status, nil
}

func buildDayHistory(logs []domain.HabitLog, today time.Time, numDays int) []DayStatus {
	history := make([]DayStatus, numDays)

	for i := 0; i < numDays; i++ {
		day := today.AddDate(0, 0, -i)
		dayLogs := domain.GetLogsForDay(logs, day)

		count := 0
		for _, log := range dayLogs {
			count += log.Count
		}

		history[i] = DayStatus{
			Date:      day,
			Completed: len(dayLogs) > 0,
			Count:     count,
		}
	}

	return history
}
