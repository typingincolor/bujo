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
		{"cancelled is valid", EntryTypeCancelled, true},
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
		{"task symbol", EntryTypeTask, "•"},
		{"note symbol", EntryTypeNote, "–"},
		{"event symbol", EntryTypeEvent, "○"},
		{"done symbol", EntryTypeDone, "✓"},
		{"migrated symbol", EntryTypeMigrated, "→"},
		{"cancelled symbol", EntryTypeCancelled, "✗"},
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
		{
			name:     "cancelled entry is complete",
			entry:    Entry{Type: EntryTypeCancelled},
			expected: true,
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
		{
			name:     "cancelled entry is never overdue",
			entry:    Entry{Type: EntryTypeCancelled, ScheduledDate: &yesterday},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.entry.IsOverdue(today))
		})
	}
}

func TestEntry_IsOverdue_ComparesDateOnly(t *testing.T) {
	todayMidnight := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	todayAfternoon := time.Date(2026, 1, 6, 14, 30, 0, 0, time.UTC)

	entry := Entry{Type: EntryTypeTask, ScheduledDate: &todayMidnight}

	assert.False(t, entry.IsOverdue(todayAfternoon),
		"task scheduled for today should NOT be overdue even when checked later in the day")
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

func TestEntry_WithEntityID_Validates(t *testing.T) {
	entry := Entry{
		EntityID: NewEntityID(),
		Type:     EntryTypeTask,
		Content:  "Do something",
	}

	err := entry.Validate()

	assert.NoError(t, err)
	assert.False(t, entry.EntityID.IsEmpty())
}

func TestNewEntry_GeneratesEntityID(t *testing.T) {
	scheduledDate := time.Now()

	entry := NewEntry(EntryTypeTask, "Do something", &scheduledDate)

	assert.False(t, entry.EntityID.IsEmpty())
	assert.Equal(t, EntryTypeTask, entry.Type)
	assert.Equal(t, "Do something", entry.Content)
	assert.Equal(t, &scheduledDate, entry.ScheduledDate)
	assert.False(t, entry.CreatedAt.IsZero())
}

func TestEntry_HasParent_WhenParentEntityIDSet_ReturnsTrue(t *testing.T) {
	parentID := NewEntityID()
	entry := Entry{
		Type:           EntryTypeTask,
		Content:        "Child task",
		ParentEntityID: &parentID,
	}

	assert.True(t, entry.HasParent())
}

func TestEntry_HasParent_WhenParentEntityIDNil_ReturnsFalse(t *testing.T) {
	entry := Entry{
		Type:    EntryTypeTask,
		Content: "Root task",
	}

	assert.False(t, entry.HasParent())
}

func TestPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected bool
	}{
		{"none is valid", PriorityNone, true},
		{"low is valid", PriorityLow, true},
		{"medium is valid", PriorityMedium, true},
		{"high is valid", PriorityHigh, true},
		{"empty string is invalid", Priority(""), false},
		{"unknown priority is invalid", Priority("urgent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.IsValid())
		})
	}
}

func TestPriority_Symbol(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected string
	}{
		{"none has no symbol", PriorityNone, ""},
		{"low has single exclamation", PriorityLow, "!"},
		{"medium has double exclamation", PriorityMedium, "!!"},
		{"high has triple exclamation", PriorityHigh, "!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.Symbol())
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Priority
		wantErr  bool
	}{
		{"none", "none", PriorityNone, false},
		{"low", "low", PriorityLow, false},
		{"medium", "medium", PriorityMedium, false},
		{"high", "high", PriorityHigh, false},
		{"empty string defaults to none", "", PriorityNone, false},
		{"invalid returns error", "urgent", PriorityNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePriority(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPriority_Cycle(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected Priority
	}{
		{"none cycles to low", PriorityNone, PriorityLow},
		{"low cycles to medium", PriorityLow, PriorityMedium},
		{"medium cycles to high", PriorityMedium, PriorityHigh},
		{"high cycles to none", PriorityHigh, PriorityNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.Cycle())
		})
	}
}
