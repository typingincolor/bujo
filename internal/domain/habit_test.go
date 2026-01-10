package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHabit_Validate(t *testing.T) {
	tests := []struct {
		name    string
		habit   Habit
		wantErr bool
	}{
		{
			name:    "valid habit with daily goal",
			habit:   Habit{Name: "Gym", GoalPerDay: 1},
			wantErr: false,
		},
		{
			name:    "valid habit with higher daily goal",
			habit:   Habit{Name: "Water", GoalPerDay: 8},
			wantErr: false,
		},
		{
			name:    "valid habit with weekly goal only",
			habit:   Habit{Name: "Gym", GoalPerWeek: 3},
			wantErr: false,
		},
		{
			name:    "valid habit with monthly goal only",
			habit:   Habit{Name: "Reading", GoalPerMonth: 4},
			wantErr: false,
		},
		{
			name:    "valid habit with all goal types",
			habit:   Habit{Name: "Workout", GoalPerDay: 1, GoalPerWeek: 5, GoalPerMonth: 20},
			wantErr: false,
		},
		{
			name:    "empty name is invalid",
			habit:   Habit{Name: "", GoalPerDay: 1},
			wantErr: true,
		},
		{
			name:    "no goals set is invalid",
			habit:   Habit{Name: "Gym"},
			wantErr: true,
		},
		{
			name:    "negative daily goal is invalid",
			habit:   Habit{Name: "Gym", GoalPerDay: -1},
			wantErr: true,
		},
		{
			name:    "negative weekly goal is invalid",
			habit:   Habit{Name: "Gym", GoalPerWeek: -1},
			wantErr: true,
		},
		{
			name:    "negative monthly goal is invalid",
			habit:   Habit{Name: "Gym", GoalPerMonth: -1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.habit.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHabitLog_Validate(t *testing.T) {
	tests := []struct {
		name    string
		log     HabitLog
		wantErr bool
	}{
		{
			name:    "valid log with count 1",
			log:     HabitLog{HabitID: 1, Count: 1, LoggedAt: time.Now()},
			wantErr: false,
		},
		{
			name:    "valid log with higher count",
			log:     HabitLog{HabitID: 1, Count: 8, LoggedAt: time.Now()},
			wantErr: false,
		},
		{
			name:    "zero habit ID is invalid",
			log:     HabitLog{HabitID: 0, Count: 1, LoggedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "zero count is invalid",
			log:     HabitLog{HabitID: 1, Count: 0, LoggedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "zero time is invalid",
			log:     HabitLog{HabitID: 1, Count: 1, LoggedAt: time.Time{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.log.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateStreak(t *testing.T) {
	today := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		logs     []HabitLog
		today    time.Time
		expected int
	}{
		{
			name:     "no logs means zero streak",
			logs:     []HabitLog{},
			today:    today,
			expected: 0,
		},
		{
			name: "single log today is streak of 1",
			logs: []HabitLog{
				{LoggedAt: today},
			},
			today:    today,
			expected: 1,
		},
		{
			name: "consecutive days streak",
			logs: []HabitLog{
				{LoggedAt: today},
				{LoggedAt: today.AddDate(0, 0, -1)},
				{LoggedAt: today.AddDate(0, 0, -2)},
			},
			today:    today,
			expected: 3,
		},
		{
			name: "broken streak resets count",
			logs: []HabitLog{
				{LoggedAt: today},
				{LoggedAt: today.AddDate(0, 0, -1)},
				// gap on day -2
				{LoggedAt: today.AddDate(0, 0, -3)},
			},
			today:    today,
			expected: 2,
		},
		{
			name: "streak requires today or yesterday to start",
			logs: []HabitLog{
				{LoggedAt: today.AddDate(0, 0, -2)},
				{LoggedAt: today.AddDate(0, 0, -3)},
			},
			today:    today,
			expected: 0,
		},
		{
			name: "multiple logs same day count as one",
			logs: []HabitLog{
				{LoggedAt: today.Add(time.Hour)},
				{LoggedAt: today.Add(2 * time.Hour)},
				{LoggedAt: today.AddDate(0, 0, -1)},
			},
			today:    today,
			expected: 2,
		},
		{
			name: "streak can start from yesterday",
			logs: []HabitLog{
				{LoggedAt: today.AddDate(0, 0, -1)},
				{LoggedAt: today.AddDate(0, 0, -2)},
				{LoggedAt: today.AddDate(0, 0, -3)},
			},
			today:    today,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateStreak(tt.logs, tt.today)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCompletion(t *testing.T) {
	today := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		logs     []HabitLog
		days     int
		today    time.Time
		expected float64
	}{
		{
			name:     "no logs means 0% completion",
			logs:     []HabitLog{},
			days:     7,
			today:    today,
			expected: 0.0,
		},
		{
			name: "all days completed is 100%",
			logs: []HabitLog{
				{LoggedAt: today},
				{LoggedAt: today.AddDate(0, 0, -1)},
				{LoggedAt: today.AddDate(0, 0, -2)},
				{LoggedAt: today.AddDate(0, 0, -3)},
				{LoggedAt: today.AddDate(0, 0, -4)},
				{LoggedAt: today.AddDate(0, 0, -5)},
				{LoggedAt: today.AddDate(0, 0, -6)},
			},
			days:     7,
			today:    today,
			expected: 100.0,
		},
		{
			name: "half days completed is ~57%",
			logs: []HabitLog{
				{LoggedAt: today},
				{LoggedAt: today.AddDate(0, 0, -2)},
				{LoggedAt: today.AddDate(0, 0, -4)},
				{LoggedAt: today.AddDate(0, 0, -6)},
			},
			days:     7,
			today:    today,
			expected: 57.14285714285714,
		},
		{
			name: "multiple logs same day count as one",
			logs: []HabitLog{
				{LoggedAt: today},
				{LoggedAt: today.Add(time.Hour)},
				{LoggedAt: today.Add(2 * time.Hour)},
			},
			days:     7,
			today:    today,
			expected: 14.285714285714285,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCompletion(tt.logs, tt.days, tt.today)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestGetLogsForDay(t *testing.T) {
	day := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	logs := []HabitLog{
		{ID: 1, LoggedAt: day.Add(8 * time.Hour)},
		{ID: 2, LoggedAt: day.Add(12 * time.Hour)},
		{ID: 3, LoggedAt: day.AddDate(0, 0, -1).Add(10 * time.Hour)},
	}

	result := GetLogsForDay(logs, day)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, int64(2), result[1].ID)
}

func TestSumCountForDay(t *testing.T) {
	day := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	logs := []HabitLog{
		{Count: 3, LoggedAt: day.Add(8 * time.Hour)},
		{Count: 2, LoggedAt: day.Add(12 * time.Hour)},
		{Count: 5, LoggedAt: day.AddDate(0, 0, -1)},
	}

	result := SumCountForDay(logs, day)

	assert.Equal(t, 5, result)
}

func TestSumCountForWeek(t *testing.T) {
	// Week ending on Jan 6, 2026 (Monday)
	weekEnd := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		logs     []HabitLog
		weekEnd  time.Time
		expected int
	}{
		{
			name:     "no logs means zero count",
			logs:     []HabitLog{},
			weekEnd:  weekEnd,
			expected: 0,
		},
		{
			name: "single log in week",
			logs: []HabitLog{
				{Count: 3, LoggedAt: weekEnd},
			},
			weekEnd:  weekEnd,
			expected: 3,
		},
		{
			name: "multiple logs in week",
			logs: []HabitLog{
				{Count: 2, LoggedAt: weekEnd},
				{Count: 3, LoggedAt: weekEnd.AddDate(0, 0, -1)},
				{Count: 1, LoggedAt: weekEnd.AddDate(0, 0, -6)},
			},
			weekEnd:  weekEnd,
			expected: 6,
		},
		{
			name: "logs outside week are excluded",
			logs: []HabitLog{
				{Count: 2, LoggedAt: weekEnd},
				{Count: 5, LoggedAt: weekEnd.AddDate(0, 0, -7)}, // outside week
			},
			weekEnd:  weekEnd,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SumCountForWeek(tt.logs, tt.weekEnd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSumCountForMonth(t *testing.T) {
	// January 2026
	monthDate := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		logs     []HabitLog
		date     time.Time
		expected int
	}{
		{
			name:     "no logs means zero count",
			logs:     []HabitLog{},
			date:     monthDate,
			expected: 0,
		},
		{
			name: "logs in same month",
			logs: []HabitLog{
				{Count: 2, LoggedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
				{Count: 3, LoggedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)},
				{Count: 1, LoggedAt: time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC)},
			},
			date:     monthDate,
			expected: 6,
		},
		{
			name: "logs in different months are excluded",
			logs: []HabitLog{
				{Count: 2, LoggedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)},
				{Count: 5, LoggedAt: time.Date(2025, 12, 31, 10, 0, 0, 0, time.UTC)}, // December
				{Count: 3, LoggedAt: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)},   // February
			},
			date:     monthDate,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SumCountForMonth(tt.logs, tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateWeeklyProgress(t *testing.T) {
	weekEnd := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		logs        []HabitLog
		goalPerWeek int
		weekEnd     time.Time
		expected    float64
	}{
		{
			name:        "no logs means 0%",
			logs:        []HabitLog{},
			goalPerWeek: 5,
			weekEnd:     weekEnd,
			expected:    0.0,
		},
		{
			name: "goal met is 100%",
			logs: []HabitLog{
				{Count: 1, LoggedAt: weekEnd},
				{Count: 1, LoggedAt: weekEnd.AddDate(0, 0, -1)},
				{Count: 1, LoggedAt: weekEnd.AddDate(0, 0, -2)},
				{Count: 1, LoggedAt: weekEnd.AddDate(0, 0, -3)},
				{Count: 1, LoggedAt: weekEnd.AddDate(0, 0, -4)},
			},
			goalPerWeek: 5,
			weekEnd:     weekEnd,
			expected:    100.0,
		},
		{
			name: "partial progress",
			logs: []HabitLog{
				{Count: 2, LoggedAt: weekEnd},
				{Count: 1, LoggedAt: weekEnd.AddDate(0, 0, -1)},
			},
			goalPerWeek: 5,
			weekEnd:     weekEnd,
			expected:    60.0,
		},
		{
			name: "over 100% is capped",
			logs: []HabitLog{
				{Count: 10, LoggedAt: weekEnd},
			},
			goalPerWeek: 5,
			weekEnd:     weekEnd,
			expected:    200.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateWeeklyProgress(tt.logs, tt.goalPerWeek, tt.weekEnd)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestCalculateMonthlyProgress(t *testing.T) {
	monthDate := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		logs         []HabitLog
		goalPerMonth int
		date         time.Time
		expected     float64
	}{
		{
			name:         "no logs means 0%",
			logs:         []HabitLog{},
			goalPerMonth: 20,
			date:         monthDate,
			expected:     0.0,
		},
		{
			name: "goal met is 100%",
			logs: []HabitLog{
				{Count: 10, LoggedAt: time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC)},
				{Count: 10, LoggedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)},
			},
			goalPerMonth: 20,
			date:         monthDate,
			expected:     100.0,
		},
		{
			name: "partial progress",
			logs: []HabitLog{
				{Count: 5, LoggedAt: time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC)},
			},
			goalPerMonth: 20,
			date:         monthDate,
			expected:     25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMonthlyProgress(tt.logs, tt.goalPerMonth, tt.date)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestHabit_WithEntityID(t *testing.T) {
	habit := Habit{
		EntityID:   NewEntityID(),
		Name:       "Gym",
		GoalPerDay: 1,
	}

	err := habit.Validate()

	assert.NoError(t, err)
	assert.False(t, habit.EntityID.IsEmpty())
}

func TestHabitLog_WithEntityID(t *testing.T) {
	log := HabitLog{
		EntityID:      NewEntityID(),
		HabitEntityID: NewEntityID(),
		HabitID:       1,
		Count:         1,
		LoggedAt:      time.Now(),
	}

	err := log.Validate()

	assert.NoError(t, err)
	assert.False(t, log.EntityID.IsEmpty())
	assert.False(t, log.HabitEntityID.IsEmpty())
}
