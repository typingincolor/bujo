package domain

import (
	"testing"
	"time"
)

func mockDateParser(s string) (time.Time, error) {
	switch s {
	case "2026-01-29":
		return time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC), nil
	case "tomorrow":
		return time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC), nil
	case "next monday":
		return time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC), nil
	default:
		return time.Time{}, &time.ParseError{Value: s}
	}
}

func TestParseMigrationSyntax_ValidFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantDate    time.Time
	}{
		{
			name:        "ISO date format",
			input:       ">[2026-01-29] Call dentist",
			wantContent: "Call dentist",
			wantDate:    time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "natural language tomorrow",
			input:       ">[tomorrow] Review PR",
			wantContent: "Review PR",
			wantDate:    time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "natural language next monday",
			input:       ">[next monday] Submit report",
			wantContent: "Submit report",
			wantDate:    time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "preserves priority markers",
			input:       ">[tomorrow] !!! Urgent migration",
			wantContent: "!!! Urgent migration",
			wantDate:    time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, date, err := ParseMigrationSyntax(tt.input, mockDateParser)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if content != tt.wantContent {
				t.Errorf("got content %q, want %q", content, tt.wantContent)
			}
			if date == nil {
				t.Fatal("expected date, got nil")
			}
			if !date.Equal(tt.wantDate) {
				t.Errorf("got date %v, want %v", date, tt.wantDate)
			}
		})
	}
}

func TestParseMigrationSyntax_InvalidFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "invalid date",
			input:   ">[never] Task",
			wantErr: true,
		},
		{
			name:    "missing closing bracket",
			input:   ">[tomorrow Task",
			wantErr: true,
		},
		{
			name:    "empty date",
			input:   ">[] Task",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseMigrationSyntax(tt.input, mockDateParser)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseMigrationSyntax_NotMigration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "regular task",
			input: ". Buy groceries",
		},
		{
			name:  "migrated display (no date)",
			input: "> Already migrated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, date, err := ParseMigrationSyntax(tt.input, mockDateParser)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if date != nil {
				t.Errorf("expected no date for non-migration, got %v", date)
			}
			if content != tt.input {
				t.Errorf("expected original content %q, got %q", tt.input, content)
			}
		})
	}
}
