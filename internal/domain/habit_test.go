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
			name:    "valid habit",
			habit:   Habit{Name: "Gym", GoalPerDay: 1},
			wantErr: false,
		},
		{
			name:    "valid habit with higher goal",
			habit:   Habit{Name: "Water", GoalPerDay: 8},
			wantErr: false,
		},
		{
			name:    "empty name is invalid",
			habit:   Habit{Name: "", GoalPerDay: 1},
			wantErr: true,
		},
		{
			name:    "zero goal is invalid",
			habit:   Habit{Name: "Gym", GoalPerDay: 0},
			wantErr: true,
		},
		{
			name:    "negative goal is invalid",
			habit:   Habit{Name: "Gym", GoalPerDay: -1},
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
