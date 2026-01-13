package dateutil

import (
	"testing"
	"time"
)

func TestParsePast_ISOFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantYear int
		wantMon  time.Month
		wantDay  int
	}{
		{
			name:     "ISO format with dashes",
			input:    "2024-12-25",
			wantYear: 2024,
			wantMon:  time.December,
			wantDay:  25,
		},
		{
			name:     "ISO format without dashes",
			input:    "20240625",
			wantYear: 2024,
			wantMon:  time.June,
			wantDay:  25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePast(tt.input)
			if err != nil {
				t.Fatalf("ParsePast() error = %v", err)
			}
			if got.Year() != tt.wantYear || got.Month() != tt.wantMon || got.Day() != tt.wantDay {
				t.Errorf("ParsePast() = %v, want %d-%02d-%02d", got, tt.wantYear, tt.wantMon, tt.wantDay)
			}
		})
	}
}

func TestParsePast_NaturalLanguage(t *testing.T) {
	tests := []struct {
		name string
		input string
	}{
		{
			name:  "yesterday",
			input: "yesterday",
		},
		{
			name:  "last week",
			input: "last week",
		},
		{
			name:  "last month",
			input: "last month",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePast(tt.input)
			if err != nil {
				t.Fatalf("ParsePast() error = %v", err)
			}
			if got.After(time.Now()) {
				t.Errorf("ParsePast() = %v, should be in the past", got)
			}
		})
	}
}

func TestParsePast_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePast(tt.input)
			if err == nil {
				t.Errorf("ParsePast() expected error for input %q", tt.input)
			}
		})
	}
}

func TestParseFuture_ISOFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantYear int
		wantMon  time.Month
		wantDay  int
	}{
		{
			name:     "ISO format with dashes",
			input:    "2025-06-15",
			wantYear: 2025,
			wantMon:  time.June,
			wantDay:  15,
		},
		{
			name:     "ISO format without dashes",
			input:    "20250815",
			wantYear: 2025,
			wantMon:  time.August,
			wantDay:  15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFuture(tt.input)
			if err != nil {
				t.Fatalf("ParseFuture() error = %v", err)
			}
			if got.Year() != tt.wantYear || got.Month() != tt.wantMon || got.Day() != tt.wantDay {
				t.Errorf("ParseFuture() = %v, want %d-%02d-%02d", got, tt.wantYear, tt.wantMon, tt.wantDay)
			}
		})
	}
}

func TestParseFuture_NaturalLanguage(t *testing.T) {
	tests := []struct {
		name string
		input string
	}{
		{
			name:  "tomorrow",
			input: "tomorrow",
		},
		{
			name:  "next week",
			input: "next week",
		},
		{
			name:  "next month",
			input: "next month",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFuture(tt.input)
			if err != nil {
				t.Fatalf("ParseFuture() error = %v", err)
			}
			if got.Before(time.Now().Add(-24 * time.Hour)) {
				t.Errorf("ParseFuture() = %v, should not be in the past", got)
			}
		})
	}
}

func TestParseFuture_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFuture(tt.input)
			if err == nil {
				t.Errorf("ParseFuture() expected error for input %q", tt.input)
			}
		})
	}
}

func TestParsePast_Today(t *testing.T) {
	now := time.Now()
	today := now.Format("2006-01-02")

	got, err := ParsePast(today)
	if err != nil {
		t.Fatalf("ParsePast() error = %v", err)
	}

	if got.Year() != now.Year() || got.Month() != now.Month() || got.Day() != now.Day() {
		t.Errorf("ParsePast() = %v, want today %v", got, now)
	}
}

func TestParseFuture_Today(t *testing.T) {
	now := time.Now()
	today := now.Format("2006-01-02")

	got, err := ParseFuture(today)
	if err != nil {
		t.Fatalf("ParseFuture() error = %v", err)
	}

	if got.Year() != now.Year() || got.Month() != now.Month() || got.Day() != now.Day() {
		t.Errorf("ParseFuture() = %v, want today %v", got, now)
	}
}
