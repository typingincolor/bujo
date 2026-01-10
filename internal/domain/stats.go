package domain

import "time"

type Stats struct {
	Period         StatsPeriod
	TotalDays      int
	EntryCounts    EntryCounts
	TaskCompletion TaskCompletion
	Productivity   Productivity
	HabitStats     HabitStats
}

type StatsPeriod struct {
	From time.Time
	To   time.Time
}

type EntryCounts struct {
	Total     int
	Tasks     int
	Notes     int
	Events    int
	Done      int
	Migrated  int
	Cancelled int
}

type TaskCompletion struct {
	Total     int
	Completed int
	Rate      float64
}

type Productivity struct {
	AveragePerDay  float64
	MostProductive Weekday
	LeastProductive Weekday
	EntriesByDay   map[time.Weekday]int
}

type Weekday struct {
	Day     time.Weekday
	Average float64
}

type HabitStats struct {
	Active      int
	BestStreak  HabitStreak
	MostLogged  HabitLogCount
	TotalLogs   int
}

type HabitStreak struct {
	HabitName string
	Days      int
}

type HabitLogCount struct {
	HabitName string
	Count     int
}
