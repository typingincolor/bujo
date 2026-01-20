package service

import (
	"context"
	"fmt"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type HabitService struct {
	habitRepo domain.HabitRepository
	logRepo   domain.HabitLogRepository
}

func NewHabitService(habitRepo domain.HabitRepository, logRepo domain.HabitLogRepository) *HabitService {
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

func (s *HabitService) HabitExists(ctx context.Context, name string) (bool, error) {
	habit, err := s.habitRepo.GetByName(ctx, name)
	if err != nil {
		return false, err
	}
	return habit != nil, nil
}

func (s *HabitService) CreateHabit(ctx context.Context, name string) (int64, error) {
	habit, err := s.habitRepo.GetOrCreate(ctx, name, 1)
	if err != nil {
		return 0, err
	}
	return habit.ID, nil
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
		HabitID:       habit.ID,
		HabitEntityID: habit.EntityID,
		Count:         count,
		LoggedAt:      date,
	}

	_, err = s.logRepo.Insert(ctx, log)
	return err
}

func (s *HabitService) LogHabitByID(ctx context.Context, habitID int64, count int) error {
	return s.LogHabitByIDForDate(ctx, habitID, count, time.Now())
}

func (s *HabitService) LogHabitByIDForDate(ctx context.Context, habitID int64, count int, date time.Time) error {
	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return err
	}

	log := domain.HabitLog{
		HabitID:       habit.ID,
		HabitEntityID: habit.EntityID,
		Count:         count,
		LoggedAt:      date,
	}

	_, err = s.logRepo.Insert(ctx, log)
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

func (s *HabitService) RemoveHabitLogForDateByID(ctx context.Context, habitID int64, date time.Time) error {
	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return err
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	logs, err := s.logRepo.GetRange(ctx, habit.ID, dayStart, dayEnd)
	if err != nil {
		return err
	}

	dayLogs := domain.GetLogsForDay(logs, date)
	if len(dayLogs) == 0 {
		return fmt.Errorf("no logs to remove for this date")
	}

	latestLog := dayLogs[0]
	for _, log := range dayLogs[1:] {
		if log.LoggedAt.After(latestLog.LoggedAt) {
			latestLog = log
		}
	}

	return s.logRepo.Delete(ctx, latestLog.ID)
}

func (s *HabitService) DeleteHabit(ctx context.Context, name string) error {
	habit, err := s.getHabitByName(ctx, name)
	if err != nil {
		return err
	}

	return s.habitRepo.Delete(ctx, habit.ID)
}

func (s *HabitService) DeleteHabitByID(ctx context.Context, habitID int64) error {
	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return err
	}

	return s.habitRepo.Delete(ctx, habit.ID)
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

func (s *HabitService) SetHabitGoal(ctx context.Context, name string, goal int) error {
	if goal < 1 {
		return fmt.Errorf("goal must be at least 1")
	}

	habit, err := s.getHabitByName(ctx, name)
	if err != nil {
		return err
	}

	habit.GoalPerDay = goal
	return s.habitRepo.Update(ctx, *habit)
}

func (s *HabitService) SetHabitGoalByID(ctx context.Context, habitID int64, goal int) error {
	if goal < 1 {
		return fmt.Errorf("goal must be at least 1")
	}

	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return err
	}

	habit.GoalPerDay = goal
	return s.habitRepo.Update(ctx, *habit)
}

func (s *HabitService) SetHabitWeeklyGoal(ctx context.Context, name string, goal int) error {
	if goal < 0 {
		return fmt.Errorf("weekly goal cannot be negative")
	}

	habit, err := s.getHabitByName(ctx, name)
	if err != nil {
		return err
	}

	habit.GoalPerWeek = goal
	return s.habitRepo.Update(ctx, *habit)
}

func (s *HabitService) SetHabitWeeklyGoalByID(ctx context.Context, habitID int64, goal int) error {
	if goal < 0 {
		return fmt.Errorf("weekly goal cannot be negative")
	}

	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return err
	}

	habit.GoalPerWeek = goal
	return s.habitRepo.Update(ctx, *habit)
}

func (s *HabitService) SetHabitMonthlyGoal(ctx context.Context, name string, goal int) error {
	if goal < 0 {
		return fmt.Errorf("monthly goal cannot be negative")
	}

	habit, err := s.getHabitByName(ctx, name)
	if err != nil {
		return err
	}

	habit.GoalPerMonth = goal
	return s.habitRepo.Update(ctx, *habit)
}

func (s *HabitService) SetHabitMonthlyGoalByID(ctx context.Context, habitID int64, goal int) error {
	if goal < 0 {
		return fmt.Errorf("monthly goal cannot be negative")
	}

	habit, err := s.getHabitByID(ctx, habitID)
	if err != nil {
		return err
	}

	habit.GoalPerMonth = goal
	return s.habitRepo.Update(ctx, *habit)
}

type HabitDetails struct {
	ID                int64
	Name              string
	GoalPerDay        int
	GoalPerWeek       int
	GoalPerMonth      int
	CurrentStreak     int
	CompletionPercent float64
	WeeklyProgress    float64
	MonthlyProgress   float64
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
	logs, err := s.logRepo.GetRangeByEntityID(ctx, habit.EntityID, from, to.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	streakLogs, err := s.logRepo.GetRangeByEntityID(ctx, habit.EntityID, todayStart.AddDate(0, 0, -365), todayStart.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	year, month, _ := todayStart.Date()
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, todayStart.Location())
	weekStart := todayStart.AddDate(0, 0, -6)
	progressStart := monthStart
	if weekStart.Before(progressStart) {
		progressStart = weekStart
	}
	progressLogs, err := s.logRepo.GetRange(ctx, habit.ID, progressStart, todayStart.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	return &HabitDetails{
		ID:                habit.ID,
		Name:              habit.Name,
		GoalPerDay:        habit.GoalPerDay,
		GoalPerWeek:       habit.GoalPerWeek,
		GoalPerMonth:      habit.GoalPerMonth,
		CurrentStreak:     domain.CalculateStreak(streakLogs, todayStart),
		CompletionPercent: domain.CalculateCompletion(logs, int(to.Sub(from).Hours()/24)+1, todayStart),
		WeeklyProgress:    domain.CalculateWeeklyProgress(progressLogs, habit.GoalPerWeek, todayStart),
		MonthlyProgress:   domain.CalculateMonthlyProgress(progressLogs, habit.GoalPerMonth, todayStart),
		Logs:              logs,
	}, nil
}

type HabitStatus struct {
	ID                int64
	Name              string
	GoalPerDay        int
	GoalPerWeek       int
	GoalPerMonth      int
	CurrentStreak     int
	CompletionPercent float64
	WeeklyProgress    float64
	MonthlyProgress   float64
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

	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	startDate := todayStart.AddDate(0, 0, -(days - 1))
	endDate := todayStart.Add(24 * time.Hour)

	weekStart := todayStart.AddDate(0, 0, -6)
	year, month, _ := todayStart.Date()
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, todayStart.Location())

	logStartDate := startDate
	if weekStart.Before(logStartDate) {
		logStartDate = weekStart
	}
	if monthStart.Before(logStartDate) {
		logStartDate = monthStart
	}

	for _, habit := range habits {
		logs, err := s.logRepo.GetRangeByEntityID(ctx, habit.EntityID, logStartDate, endDate)
		if err != nil {
			return nil, err
		}

		habitStatus := HabitStatus{
			ID:                habit.ID,
			Name:              habit.Name,
			GoalPerDay:        habit.GoalPerDay,
			GoalPerWeek:       habit.GoalPerWeek,
			GoalPerMonth:      habit.GoalPerMonth,
			CurrentStreak:     domain.CalculateStreak(logs, todayStart),
			CompletionPercent: domain.CalculateCompletion(logs, days, todayStart),
			WeeklyProgress:    domain.CalculateWeeklyProgress(logs, habit.GoalPerWeek, todayStart),
			MonthlyProgress:   domain.CalculateMonthlyProgress(logs, habit.GoalPerMonth, todayStart),
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
