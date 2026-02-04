package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsightsDashboard_StatusReady(t *testing.T) {
	d := InsightsDashboard{
		LatestSummary: &InsightsSummary{ID: 1, WeekStart: "2026-01-27"},
		Status:        "ready",
	}
	assert.Equal(t, "ready", d.Status)
}

func TestInsightsDashboard_StatusEmpty(t *testing.T) {
	d := InsightsDashboard{
		Status: "empty",
	}
	assert.Equal(t, "empty", d.Status)
}

func TestInsightsAction_IsOverdue(t *testing.T) {
	tests := []struct {
		name   string
		action InsightsAction
		today  string
		want   bool
	}{
		{
			name:   "overdue action",
			action: InsightsAction{DueDate: "2026-01-15", Status: "pending"},
			today:  "2026-02-04",
			want:   true,
		},
		{
			name:   "future action",
			action: InsightsAction{DueDate: "2026-03-01", Status: "pending"},
			today:  "2026-02-04",
			want:   false,
		},
		{
			name:   "no due date",
			action: InsightsAction{DueDate: "", Status: "pending"},
			today:  "2026-02-04",
			want:   false,
		},
		{
			name:   "completed action not overdue",
			action: InsightsAction{DueDate: "2026-01-15", Status: "completed"},
			today:  "2026-02-04",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.action.IsOverdue(tt.today))
		})
	}
}
