package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSummaryHorizon_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		horizon  SummaryHorizon
		expected bool
	}{
		{"daily is valid", SummaryHorizonDaily, true},
		{"weekly is valid", SummaryHorizonWeekly, true},
		{"empty is invalid", SummaryHorizon(""), false},
		{"unknown is invalid", SummaryHorizon("monthly"), false},
		{"quarterly no longer supported", SummaryHorizon("quarterly"), false},
		{"annual no longer supported", SummaryHorizon("annual"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.horizon.IsValid())
		})
	}
}

func TestSummary_PeriodLength(t *testing.T) {
	tests := []struct {
		name     string
		summary  Summary
		expected int
	}{
		{
			name: "single day period",
			summary: Summary{
				StartDate: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
			},
			expected: 1,
		},
		{
			name: "week period",
			summary: Summary{
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
			},
			expected: 7,
		},
		{
			name: "month period",
			summary: Summary{
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			},
			expected: 31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.summary.PeriodLength())
		})
	}
}

func TestSummary_IsRecent(t *testing.T) {
	now := time.Date(2026, 1, 6, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		summary  Summary
		expected bool
	}{
		{
			name: "created today is recent",
			summary: Summary{
				CreatedAt: time.Date(2026, 1, 6, 10, 0, 0, 0, time.UTC),
			},
			expected: true,
		},
		{
			name: "created yesterday is not recent",
			summary: Summary{
				CreatedAt: time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC),
			},
			expected: false,
		},
		{
			name: "created last week is not recent",
			summary: Summary{
				CreatedAt: time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.summary.IsRecent(now))
		})
	}
}

func TestSummary_Validate(t *testing.T) {
	validStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	validEnd := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		summary Summary
		wantErr bool
	}{
		{
			name: "valid summary",
			summary: Summary{
				Horizon:   SummaryHorizonWeekly,
				Content:   "Weekly reflection content",
				StartDate: validStart,
				EndDate:   validEnd,
			},
			wantErr: false,
		},
		{
			name: "invalid horizon",
			summary: Summary{
				Horizon:   SummaryHorizon("invalid"),
				Content:   "Content",
				StartDate: validStart,
				EndDate:   validEnd,
			},
			wantErr: true,
		},
		{
			name: "empty content",
			summary: Summary{
				Horizon:   SummaryHorizonWeekly,
				Content:   "",
				StartDate: validStart,
				EndDate:   validEnd,
			},
			wantErr: true,
		},
		{
			name: "end before start",
			summary: Summary{
				Horizon:   SummaryHorizonWeekly,
				Content:   "Content",
				StartDate: validEnd,
				EndDate:   validStart,
			},
			wantErr: true,
		},
		{
			name: "zero start date",
			summary: Summary{
				Horizon:   SummaryHorizonWeekly,
				Content:   "Content",
				StartDate: time.Time{},
				EndDate:   validEnd,
			},
			wantErr: true,
		},
		{
			name: "zero end date",
			summary: Summary{
				Horizon:   SummaryHorizonWeekly,
				Content:   "Content",
				StartDate: validStart,
				EndDate:   time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.summary.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
