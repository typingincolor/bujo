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
			name:    "compact date format YYYYMMDD",
			input:   "20260106",
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

func TestParsePastDate_CompactFormat(t *testing.T) {
	parsed, err := parsePastDate("20260106")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Year() != 2026 || parsed.Month() != 1 || parsed.Day() != 6 {
		t.Errorf("expected 2026-01-06, got %v", parsed)
	}
}

func TestParseDateOrToday(t *testing.T) {
	t.Run("empty string returns today", func(t *testing.T) {
		result, err := parseDateOrToday("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		today := time.Now()
		if result.Year() != today.Year() || result.Month() != today.Month() || result.Day() != today.Day() {
			t.Errorf("expected today, got %v", result)
		}
	})

	t.Run("valid date string is parsed", func(t *testing.T) {
		result, err := parseDateOrToday("2026-01-05")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Year() != 2026 || result.Month() != 1 || result.Day() != 5 {
			t.Errorf("expected 2026-01-05, got %v", result)
		}
	})

	t.Run("invalid date returns error", func(t *testing.T) {
		_, err := parseDateOrToday("not-a-date")
		if err == nil {
			t.Error("expected error for invalid date")
		}
	})
}

func TestParseAddArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantEntries  []string
		wantLocation string
		wantDate     string
		wantFile     string
		wantHelp     bool
		wantYes      bool
	}{
		{
			name:        "simple task entry",
			args:        []string{". Buy groceries"},
			wantEntries: []string{". Buy groceries"},
		},
		{
			name:        "note entry starting with dash",
			args:        []string{"- working on objectives"},
			wantEntries: []string{"- working on objectives"},
		},
		{
			name:        "multiple entries",
			args:        []string{". Task", "- Note", "o Event"},
			wantEntries: []string{". Task", "- Note", "o Event"},
		},
		{
			name:         "with location flag",
			args:         []string{"--at", "Home", ". Task"},
			wantEntries:  []string{". Task"},
			wantLocation: "Home",
		},
		{
			name:         "with short location flag",
			args:         []string{"-a", "Office", "- Note"},
			wantEntries:  []string{"- Note"},
			wantLocation: "Office",
		},
		{
			name:        "with date flag",
			args:        []string{"--date", "yesterday", ". Task"},
			wantEntries: []string{". Task"},
			wantDate:    "yesterday",
		},
		{
			name:        "with short date flag",
			args:        []string{"-d", "last week", "- Note"},
			wantEntries: []string{"- Note"},
			wantDate:    "last week",
		},
		{
			name:         "with equals syntax",
			args:         []string{"--at=Home", "-d=yesterday", ". Task"},
			wantEntries:  []string{". Task"},
			wantLocation: "Home",
			wantDate:     "yesterday",
		},
		{
			name:        "skips global db-path flag",
			args:        []string{"--db-path", "/tmp/test.db", ". Task"},
			wantEntries: []string{". Task"},
		},
		{
			name:        "skips verbose flag",
			args:        []string{"-v", ". Task"},
			wantEntries: []string{". Task"},
		},
		{
			name:     "help flag",
			args:     []string{"-h"},
			wantHelp: true,
		},
		{
			name:        "double dash stops flag parsing",
			args:        []string{"--", "-a", "not a flag"},
			wantEntries: []string{"-a", "not a flag"},
		},
		{
			name:     "with file flag",
			args:     []string{"--file", "tasks.txt"},
			wantFile: "tasks.txt",
		},
		{
			name:     "with short file flag",
			args:     []string{"-f", "entries.txt"},
			wantFile: "entries.txt",
		},
		{
			name:         "file flag with other options",
			args:         []string{"-f", "tasks.txt", "--at", "Home", "-d", "yesterday"},
			wantFile:     "tasks.txt",
			wantLocation: "Home",
			wantDate:     "yesterday",
		},
		{
			name:     "file flag with equals syntax",
			args:     []string{"--file=/path/to/file.txt"},
			wantFile: "/path/to/file.txt",
		},
		{
			name:     "short file flag with equals syntax",
			args:     []string{"-f=file.txt"},
			wantFile: "file.txt",
		},
		{
			name:        "with yes flag",
			args:        []string{"-y", "-d", "yesterday", ". Task"},
			wantEntries: []string{". Task"},
			wantDate:    "yesterday",
			wantYes:     true,
		},
		{
			name:        "with long yes flag",
			args:        []string{"--yes", ". Task"},
			wantEntries: []string{". Task"},
			wantYes:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, location, date, file, help, yes := parseAddArgs(tt.args)

			if help != tt.wantHelp {
				t.Errorf("parseAddArgs() help = %v, want %v", help, tt.wantHelp)
			}
			if yes != tt.wantYes {
				t.Errorf("parseAddArgs() yes = %v, want %v", yes, tt.wantYes)
			}
			if location != tt.wantLocation {
				t.Errorf("parseAddArgs() location = %q, want %q", location, tt.wantLocation)
			}
			if date != tt.wantDate {
				t.Errorf("parseAddArgs() date = %q, want %q", date, tt.wantDate)
			}
			if file != tt.wantFile {
				t.Errorf("parseAddArgs() file = %q, want %q", file, tt.wantFile)
			}
			if len(entries) != len(tt.wantEntries) {
				t.Errorf("parseAddArgs() entries = %v, want %v", entries, tt.wantEntries)
				return
			}
			for i, e := range entries {
				if e != tt.wantEntries[i] {
					t.Errorf("parseAddArgs() entries[%d] = %q, want %q", i, e, tt.wantEntries[i])
				}
			}
		})
	}
}

func TestIsNaturalLanguageDate(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"2026-01-05", false},         // ISO format
		{"20260105", false},           // Compact format
		{"yesterday", true},           // Natural language
		{"last week", true},           // Natural language
		{"next monday", true},         // Natural language
		{"tomorrow", true},            // Natural language
		{"2 days ago", true},          // Natural language
		{"", false},                   // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isNaturalLanguageDate(tt.input)
			if result != tt.expected {
				t.Errorf("isNaturalLanguageDate(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfirmDate_SkipsPrompt(t *testing.T) {
	testDate := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		dateStr     string
		skipConfirm bool
	}{
		{
			name:        "skip when ISO format",
			dateStr:     "2026-01-05",
			skipConfirm: false,
		},
		{
			name:        "skip when compact format",
			dateStr:     "20260105",
			skipConfirm: false,
		},
		{
			name:        "skip when --yes flag",
			dateStr:     "yesterday",
			skipConfirm: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := confirmDate(tt.dateStr, testDate, tt.skipConfirm)
			if err != nil {
				t.Errorf("confirmDate() unexpected error: %v", err)
			}
			if result != testDate {
				t.Errorf("confirmDate() = %v, want %v", result, testDate)
			}
		})
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
