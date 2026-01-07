package cmd

import (
	"testing"
	"time"
)

func TestParsePastDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "ISO date format",
			input:   "2026-01-05",
			wantErr: false,
		},
		{
			name:    "natural language yesterday",
			input:   "yesterday",
			wantErr: false,
		},
		{
			name:    "natural language last week",
			input:   "last week",
			wantErr: false,
		},
		{
			name:    "invalid date",
			input:   "not-a-date-xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePastDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePastDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParsePastDate_ISOFormat(t *testing.T) {
	parsed, err := parsePastDate("2026-01-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Year() != 2026 || parsed.Month() != 1 || parsed.Day() != 5 {
		t.Errorf("expected 2026-01-05, got %v", parsed)
	}
}

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name    string
		from    time.Time
		to      time.Time
		wantErr bool
	}{
		{
			name:    "valid range",
			from:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			to:      time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "same day is valid",
			from:    time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			to:      time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "from after to is invalid",
			from:    time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
			to:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
