package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntryType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		et       EntryType
		expected bool
	}{
		{"task is valid", EntryTypeTask, true},
		{"note is valid", EntryTypeNote, true},
		{"event is valid", EntryTypeEvent, true},
		{"done is valid", EntryTypeDone, true},
		{"migrated is valid", EntryTypeMigrated, true},
		{"empty is invalid", EntryType(""), false},
		{"unknown is invalid", EntryType("?"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.et.IsValid())
		})
	}
}

func TestEntryType_Symbol(t *testing.T) {
	tests := []struct {
		name     string
		et       EntryType
		expected string
	}{
		{"task symbol", EntryTypeTask, "."},
		{"note symbol", EntryTypeNote, "-"},
		{"event symbol", EntryTypeEvent, "o"},
		{"done symbol", EntryTypeDone, "x"},
		{"migrated symbol", EntryTypeMigrated, ">"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.et.Symbol())
		})
	}
}

func TestEntry_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		entry    Entry
		expected bool
	}{
		{
			name:     "done entry is complete",
			entry:    Entry{Type: EntryTypeDone},
			expected: true,
		},
		{
			name:     "task entry is not complete",
			entry:    Entry{Type: EntryTypeTask},
			expected: false,
		},
		{
			name:     "note entry is not complete",
			entry:    Entry{Type: EntryTypeNote},
			expected: false,
		},
		{
			name:     "event entry is not complete",
			entry:    Entry{Type: EntryTypeEvent},
			expected: false,
		},
		{
			name:     "migrated entry is not complete",
			entry:    Entry{Type: EntryTypeMigrated},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.entry.IsComplete())
		})
	}
}

func TestEntry_IsOverdue(t *testing.T) {
	today := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		entry    Entry
		expected bool
	}{
		{
			name:     "task scheduled yesterday is overdue",
			entry:    Entry{Type: EntryTypeTask, ScheduledDate: &yesterday},
			expected: true,
		},
		{
			name:     "task scheduled today is not overdue",
			entry:    Entry{Type: EntryTypeTask, ScheduledDate: &today},
			expected: false,
		},
		{
			name:     "task scheduled tomorrow is not overdue",
			entry:    Entry{Type: EntryTypeTask, ScheduledDate: &tomorrow},
			expected: false,
		},
		{
			name:     "task with no scheduled date is not overdue",
			entry:    Entry{Type: EntryTypeTask, ScheduledDate: nil},
			expected: false,
		},
		{
			name:     "completed task scheduled yesterday is not overdue",
			entry:    Entry{Type: EntryTypeDone, ScheduledDate: &yesterday},
			expected: false,
		},
		{
			name:     "note is never overdue",
			entry:    Entry{Type: EntryTypeNote, ScheduledDate: &yesterday},
			expected: false,
		},
		{
			name:     "event is never overdue",
			entry:    Entry{Type: EntryTypeEvent, ScheduledDate: &yesterday},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.entry.IsOverdue(today))
		})
	}
}

func TestEntry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		entry   Entry
		wantErr bool
	}{
		{
			name:    "valid task entry",
			entry:   Entry{Type: EntryTypeTask, Content: "Do something"},
			wantErr: false,
		},
		{
			name:    "empty content is invalid",
			entry:   Entry{Type: EntryTypeTask, Content: ""},
			wantErr: true,
		},
		{
			name:    "invalid type is invalid",
			entry:   Entry{Type: EntryType("?"), Content: "Something"},
			wantErr: true,
		},
		{
			name:    "negative depth is invalid",
			entry:   Entry{Type: EntryTypeTask, Content: "Task", Depth: -1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
