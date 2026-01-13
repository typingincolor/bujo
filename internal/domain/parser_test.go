package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEntryType(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected EntryType
	}{
		{"task from dot", ". Do something", EntryTypeTask},
		{"note from dash", "- Some note", EntryTypeNote},
		{"event from o", "o Meeting @ 10am", EntryTypeEvent},
		{"done from x", "x Completed task", EntryTypeDone},
		{"migrated from >", "> Moved to tomorrow", EntryTypeMigrated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEntryType(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseIndentation(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		expectedDepth int
		expectedRest  string
	}{
		{"no indentation", ". Task", 0, ". Task"},
		{"two spaces", "  . Child task", 1, ". Child task"},
		{"four spaces", "    . Grandchild", 2, ". Grandchild"},
		{"one tab", "\t. Child task", 1, ". Child task"},
		{"two tabs", "\t\t. Grandchild", 2, ". Grandchild"},
		{"mixed spaces", "      . Third level", 3, ". Third level"},
		{"empty string", "", 0, ""},
		{"only spaces", "    ", 2, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth, rest := ParseIndentation(tt.line)
			assert.Equal(t, tt.expectedDepth, depth)
			assert.Equal(t, tt.expectedRest, rest)
		})
	}
}

func TestParseContent(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"task content", ". Do something", "Do something"},
		{"note content", "- Some note here", "Some note here"},
		{"event content", "o Meeting @ 10am", "Meeting @ 10am"},
		{"content with extra spaces", ".   Multiple spaces", "Multiple spaces"},
		{"empty string", "", ""},
		{"single character", ".", ""},
		{"no content after symbol", "- ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseContent(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTreeParser_Parse_SingleEntry(t *testing.T) {
	parser := NewTreeParser()
	input := ". Buy groceries"

	entries, err := parser.Parse(input)

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, EntryTypeTask, entries[0].Type)
	assert.Equal(t, "Buy groceries", entries[0].Content)
	assert.Equal(t, 0, entries[0].Depth)
	assert.Nil(t, entries[0].ParentID)
}

func TestTreeParser_Parse_MultipleRootEntries(t *testing.T) {
	parser := NewTreeParser()
	input := `. Task one
- Note one
o Event one`

	entries, err := parser.Parse(input)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, EntryTypeTask, entries[0].Type)
	assert.Equal(t, "Task one", entries[0].Content)
	assert.Nil(t, entries[0].ParentID)

	assert.Equal(t, EntryTypeNote, entries[1].Type)
	assert.Equal(t, "Note one", entries[1].Content)
	assert.Nil(t, entries[1].ParentID)

	assert.Equal(t, EntryTypeEvent, entries[2].Type)
	assert.Equal(t, "Event one", entries[2].Content)
	assert.Nil(t, entries[2].ParentID)
}

func TestTreeParser_Parse_NestedEntries(t *testing.T) {
	parser := NewTreeParser()
	input := `o Project Kickoff
  - Attendees: Alice, Bob
  . Send follow-up email
    - Include PDF attachment`

	entries, err := parser.Parse(input)

	require.NoError(t, err)
	require.Len(t, entries, 4)

	// Root entry
	assert.Equal(t, EntryTypeEvent, entries[0].Type)
	assert.Equal(t, "Project Kickoff", entries[0].Content)
	assert.Equal(t, 0, entries[0].Depth)
	assert.Nil(t, entries[0].ParentID)

	// First child - note
	assert.Equal(t, EntryTypeNote, entries[1].Type)
	assert.Equal(t, "Attendees: Alice, Bob", entries[1].Content)
	assert.Equal(t, 1, entries[1].Depth)
	require.NotNil(t, entries[1].ParentID)
	assert.Equal(t, int64(0), *entries[1].ParentID)

	// Second child - task
	assert.Equal(t, EntryTypeTask, entries[2].Type)
	assert.Equal(t, "Send follow-up email", entries[2].Content)
	assert.Equal(t, 1, entries[2].Depth)
	require.NotNil(t, entries[2].ParentID)
	assert.Equal(t, int64(0), *entries[2].ParentID)

	// Grandchild - note under task
	assert.Equal(t, EntryTypeNote, entries[3].Type)
	assert.Equal(t, "Include PDF attachment", entries[3].Content)
	assert.Equal(t, 2, entries[3].Depth)
	require.NotNil(t, entries[3].ParentID)
	assert.Equal(t, int64(2), *entries[3].ParentID)
}

func TestTreeParser_Parse_TabIndentation(t *testing.T) {
	parser := NewTreeParser()
	input := `. Parent task
	. Child task
		. Grandchild task`

	entries, err := parser.Parse(input)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, 0, entries[0].Depth)
	assert.Nil(t, entries[0].ParentID)

	assert.Equal(t, 1, entries[1].Depth)
	require.NotNil(t, entries[1].ParentID)
	assert.Equal(t, int64(0), *entries[1].ParentID)

	assert.Equal(t, 2, entries[2].Depth)
	require.NotNil(t, entries[2].ParentID)
	assert.Equal(t, int64(1), *entries[2].ParentID)
}

func TestTreeParser_Parse_EmptyInput(t *testing.T) {
	parser := NewTreeParser()

	entries, err := parser.Parse("")

	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestTreeParser_Parse_SkipsEmptyLines(t *testing.T) {
	parser := NewTreeParser()
	input := `. Task one

. Task two`

	entries, err := parser.Parse(input)

	require.NoError(t, err)
	require.Len(t, entries, 2)
}

func TestTreeParser_Parse_InvalidIndentation(t *testing.T) {
	parser := NewTreeParser()
	// Child without parent (jumps to depth 2 without depth 1)
	input := `. Root task
      . Invalid grandchild`

	_, err := parser.Parse(input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid indentation")
}

func TestTreeParser_Parse_UnknownSymbol_DefaultsToNote(t *testing.T) {
	parser := NewTreeParser()

	tests := []struct {
		name    string
		input   string
		content string
	}{
		{"number prefix", "1 Numbered item", "Numbered item"},
		{"hash prefix", "# Heading style", "Heading style"},
		{"asterisk prefix", "* Bullet point", "Bullet point"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := parser.Parse(tt.input)

			require.NoError(t, err, "Unknown symbols should not cause an error")
			require.Len(t, entries, 1)
			assert.Equal(t, EntryTypeNote, entries[0].Type, "Unknown symbols should default to note")
			assert.Equal(t, tt.content, entries[0].Content)
		})
	}
}

func TestTreeParser_Parse_UnknownSymbol_PreservesHierarchy(t *testing.T) {
	parser := NewTreeParser()
	input := `. Valid task
  # Unknown child symbol
  - Valid note`

	entries, err := parser.Parse(input)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, EntryTypeTask, entries[0].Type)
	assert.Equal(t, EntryTypeNote, entries[1].Type, "Unknown symbol should default to note")
	assert.Equal(t, "Unknown child symbol", entries[1].Content)
	assert.NotNil(t, entries[1].ParentID)
	assert.Equal(t, EntryTypeNote, entries[2].Type)
}

func TestTreeParser_Parse_WithPriority(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedContent  string
		expectedPriority Priority
	}{
		{
			name:             "low priority single exclamation",
			input:            ". ! Buy groceries",
			expectedContent:  "Buy groceries",
			expectedPriority: PriorityLow,
		},
		{
			name:             "medium priority double exclamation",
			input:            ". !! Urgent task",
			expectedContent:  "Urgent task",
			expectedPriority: PriorityMedium,
		},
		{
			name:             "high priority triple exclamation",
			input:            ". !!! Critical task",
			expectedContent:  "Critical task",
			expectedPriority: PriorityHigh,
		},
		{
			name:             "no priority",
			input:            ". Regular task",
			expectedContent:  "Regular task",
			expectedPriority: PriorityNone,
		},
		{
			name:             "note with priority",
			input:            "- ! Important note",
			expectedContent:  "Important note",
			expectedPriority: PriorityLow,
		},
		{
			name:             "event with priority",
			input:            "o !! Urgent meeting",
			expectedContent:  "Urgent meeting",
			expectedPriority: PriorityMedium,
		},
	}

	parser := NewTreeParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := parser.Parse(tt.input)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			assert.Equal(t, tt.expectedContent, entries[0].Content)
			assert.Equal(t, tt.expectedPriority, entries[0].Priority)
		})
	}
}

func TestParsePriorityAndContent(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedContent  string
		expectedPriority Priority
	}{
		{"no priority", "Buy groceries", "Buy groceries", PriorityNone},
		{"low priority", "! Buy groceries", "Buy groceries", PriorityLow},
		{"medium priority", "!! Buy groceries", "Buy groceries", PriorityMedium},
		{"high priority", "!!! Buy groceries", "Buy groceries", PriorityHigh},
		{"exclamation in content", "Say hello!", "Say hello!", PriorityNone},
		{"exclamation not at start", "Hello ! World", "Hello ! World", PriorityNone},
		{"four exclamations treated as high", "!!!! Too many", "! Too many", PriorityHigh},
		{"empty string", "", "", PriorityNone},
		{"only exclamations", "!!!", "!!!", PriorityNone},
		{"exclamation with only whitespace", "!   ", "!", PriorityNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, priority := ParsePriorityAndContent(tt.input)
			assert.Equal(t, tt.expectedContent, content)
			assert.Equal(t, tt.expectedPriority, priority)
		})
	}
}

func TestParseEntryType_QuestionAndAnswered(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected EntryType
	}{
		{"question from ?", "? What is the deadline", EntryTypeQuestion},
		{"answered from star", "★ This is answered", EntryTypeAnswered},
		{"answer from arrow", "↳ This is the answer", EntryTypeAnswer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEntryType(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEntryType_UnicodeSymbols(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected EntryType
	}{
		{"bullet task", "• Unicode task", EntryTypeTask},
		{"en-dash note", "– Unicode note", EntryTypeNote},
		{"circle event", "○ Unicode event", EntryTypeEvent},
		{"checkmark done", "✓ Unicode done", EntryTypeDone},
		{"arrow migrated", "→ Unicode migrated", EntryTypeMigrated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEntryType(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEntryType_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected EntryType
	}{
		{"empty string", "", EntryType("")},
		{"unknown symbol", "@ Unknown", EntryType("")},
		{"number", "1. Numbered item", EntryType("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEntryType(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTreeParser_Parse_QuestionWithPriority(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedType     EntryType
		expectedContent  string
		expectedPriority Priority
	}{
		{
			name:             "question with low priority",
			input:            "? ! What is the deadline",
			expectedType:     EntryTypeQuestion,
			expectedContent:  "What is the deadline",
			expectedPriority: PriorityLow,
		},
		{
			name:             "question with medium priority",
			input:            "? !! Need clarification on requirements",
			expectedType:     EntryTypeQuestion,
			expectedContent:  "Need clarification on requirements",
			expectedPriority: PriorityMedium,
		},
		{
			name:             "question with high priority",
			input:            "? !!! Critical question about deployment",
			expectedType:     EntryTypeQuestion,
			expectedContent:  "Critical question about deployment",
			expectedPriority: PriorityHigh,
		},
		{
			name:             "question without priority",
			input:            "? Simple question",
			expectedType:     EntryTypeQuestion,
			expectedContent:  "Simple question",
			expectedPriority: PriorityNone,
		},
	}

	parser := NewTreeParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := parser.Parse(tt.input)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			assert.Equal(t, tt.expectedType, entries[0].Type)
			assert.Equal(t, tt.expectedContent, entries[0].Content)
			assert.Equal(t, tt.expectedPriority, entries[0].Priority)
		})
	}
}
