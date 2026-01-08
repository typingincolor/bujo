package domain

import (
	"errors"
	"time"
)

type Habit struct {
	ID         int64
	EntityID   EntityID
	Name       string
	GoalPerDay int
	CreatedAt  time.Time
}

func (h Habit) Validate() error {
	if h.Name == "" {
		return errors.New("habit name cannot be empty")
	}
	if h.GoalPerDay <= 0 {
		return errors.New("goal per day must be positive")
	}
	return nil
}

type HabitLog struct {
	ID            int64
	EntityID      EntityID
	HabitID       int64
	HabitEntityID EntityID
	Count         int
	LoggedAt      time.Time
}

func (l HabitLog) Validate() error {
	if l.HabitID <= 0 {
		return errors.New("habit ID must be positive")
	}
	if l.Count <= 0 {
		return errors.New("count must be positive")
	}
	if l.LoggedAt.IsZero() {
		return errors.New("logged at time is required")
	}
	return nil
}

func CalculateStreak(logs []HabitLog, today time.Time) int {
	if len(logs) == 0 {
		return 0
	}

	loggedDays := make(map[string]bool)
	for _, log := range logs {
		dayKey := log.LoggedAt.Format("2006-01-02")
		loggedDays[dayKey] = true
	}

	todayKey := today.Format("2006-01-02")
	yesterdayKey := today.AddDate(0, 0, -1).Format("2006-01-02")

	if !loggedDays[todayKey] && !loggedDays[yesterdayKey] {
		return 0
	}

	streak := 0
	checkDay := today

	if !loggedDays[todayKey] {
		checkDay = today.AddDate(0, 0, -1)
	}

	for {
		dayKey := checkDay.Format("2006-01-02")
		if !loggedDays[dayKey] {
			break
		}
		streak++
		checkDay = checkDay.AddDate(0, 0, -1)
	}

	return streak
}

func CalculateCompletion(logs []HabitLog, days int, today time.Time) float64 {
	if days <= 0 {
		return 0.0
	}

	// Normalize to start of day for consistent date comparisons
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)
	startDate := todayStart.AddDate(0, 0, -(days - 1))

	loggedDays := make(map[string]bool)
	for _, log := range logs {
		if !log.LoggedAt.Before(startDate) && log.LoggedAt.Before(todayEnd) {
			dayKey := log.LoggedAt.Format("2006-01-02")
			loggedDays[dayKey] = true
		}
	}

	completedDays := len(loggedDays)
	return (float64(completedDays) / float64(days)) * 100.0
}

func GetLogsForDay(logs []HabitLog, day time.Time) []HabitLog {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	result := make([]HabitLog, 0)
	for _, log := range logs {
		if !log.LoggedAt.Before(dayStart) && log.LoggedAt.Before(dayEnd) {
			result = append(result, log)
		}
	}
	return result
}

func SumCountForDay(logs []HabitLog, day time.Time) int {
	dayLogs := GetLogsForDay(logs, day)
	sum := 0
	for _, log := range dayLogs {
		sum += log.Count
	}
	return sum
}
