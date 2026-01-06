package service

import (
	"context"
	"fmt"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type HabitRepository interface {
	GetOrCreate(ctx context.Context, name string, goalPerDay int) (*domain.Habit, error)
	GetByID(ctx context.Context, id int64) (*domain.Habit, error)
	GetByName(ctx context.Context, name string) (*domain.Habit, error)
	GetAll(ctx context.Context) ([]domain.Habit, error)
	Update(ctx context.Context, habit domain.Habit) error
}

type HabitLogRepository interface {
	Insert(ctx context.Context, log domain.HabitLog) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.HabitLog, error)
	GetRange(ctx context.Context, habitID int64, start, end time.Time) ([]domain.HabitLog, error)
	GetLastByHabitID(ctx context.Context, habitID int64) (*domain.HabitLog, error)
	Delete(ctx context.Context, id int64) error
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

func (s *HabitService) getHabitByID(ctx context.Context, id int64) (*domain.Habit, error) {
	habit, err := s.habitRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if habit == nil {
		return nil, fmt.Errorf("habit not found: %d", id)
	}
	return habit, nil
}

func (s *HabitService) getHabitByName(ctx context.Context, name string) (*domain.Habit, error) {
	habit, err := s.habitRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if habit == nil {
		return nil, fmt.Errorf("habit not found: %s", name)
	}
	return habit, nil
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

func (s *HabitService) LogHabitByID(ctx context.Context, habitID int64, count int) error {
	return s.LogHabitByIDForDate(ctx, habitID, count, time.Now())
}

func (s *HabitService) LogHabitByIDForDate(ctx context.Context, habitID int64, count int, date time.Time) error {
	if _, err := s.getHabitByID(ctx, habitID); err != nil {
		return err
	}

	log := domain.HabitLog{
		HabitID:  habitID,
		Count:    count,
		LoggedAt: date,
	}

	_, err := s.logRepo.Insert(ctx, log)
	return err
}

func (s *HabitService) UndoLastLog(ctx context.Context, name string) error {
	habit, err := s.getHabitByName(ctx, name)
	if err != nil {
		return err
	}

	return s.undoLastLogForHabit(ctx, habit.ID)
}

func (s *HabitService) UndoLastLogByID(ctx context.Context, habitID int64) error {
	if _, err := s.getHabitByID(ctx, habitID); err != nil {
		return err
	}

	return s.undoLastLogForHabit(ctx, habitID)
}

func (s *HabitService) undoLastLogForHabit(ctx context.Context, habitID int64) error {
	lastLog, err := s.logRepo.GetLastByHabitID(ctx, habitID)
	if err != nil {
		return err
	}
	if lastLog == nil {
		return fmt.Errorf("no logs to undo")
	}

	return s.logRepo.Delete(ctx, lastLog.ID)
}

func (s *HabitService) DeleteLog(ctx context.Context, logID int64) error {
	log, err := s.logRepo.GetByID(ctx, logID)
	if err != nil {
		return err
	}
	if log == nil {
		return fmt.Errorf("log not found: %d", logID)
	}

	return s.logRepo.Delete(ctx, logID)
}

func (s *HabitService) RenameHabit(ctx context.Context, oldName, newName string) error {
	habit, err := s.getHabitByName(ctx, oldName)
	if err != nil {
		return err
	}

	habit.Name = newName
	return s.habitRepo.Update(ctx, *habit)
}

func (s *HabitService) RenameHabitByID(ctx context.Context, habitID int64, newName string) error {
	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return err
	}

	habit.Name = newName
	return s.habitRepo.Update(ctx, *habit)
}

type HabitDetails struct {
	ID                int64
	Name              string
	GoalPerDay        int
	CurrentStreak     int
	CompletionPercent float64
	Logs              []domain.HabitLog
}

func (s *HabitService) InspectHabit(ctx context.Context, name string, from, to, today time.Time) (*HabitDetails, error) {
	habit, err := s.getHabitByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return s.inspectHabitByHabit(ctx, habit, from, to, today)
}

func (s *HabitService) InspectHabitByID(ctx context.Context, habitID int64, from, to, today time.Time) (*HabitDetails, error) {
	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	return s.inspectHabitByHabit(ctx, habit, from, to, today)
}

func (s *HabitService) inspectHabitByHabit(ctx context.Context, habit *domain.Habit, from, to, today time.Time) (*HabitDetails, error) {
	// Get logs for the date range
	logs, err := s.logRepo.GetRange(ctx, habit.ID, from, to.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	// Calculate streak using all recent logs (not just the range)
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	streakLogs, err := s.logRepo.GetRange(ctx, habit.ID, todayStart.AddDate(0, 0, -365), todayStart.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	return &HabitDetails{
		ID:                habit.ID,
		Name:              habit.Name,
		GoalPerDay:        habit.GoalPerDay,
		CurrentStreak:     domain.CalculateStreak(streakLogs, todayStart),
		CompletionPercent: domain.CalculateCompletion(logs, int(to.Sub(from).Hours()/24)+1, todayStart),
		Logs:              logs,
	}, nil
}

type HabitStatus struct {
	ID                int64
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
			ID:                habit.ID,
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
