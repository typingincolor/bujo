package domain

import (
	"testing"
)

func TestParseLine_SymbolTypes(t *testing.T) {
	parser := NewEditableDocumentParser()

	tests := []struct {
		name     string
		input    string
		wantType EntryType
		wantText string
	}{
		{
			name:     "task with dot",
			input:    ". Buy groceries",
			wantType: EntryTypeTask,
			wantText: "Buy groceries",
		},
		{
			name:     "note with dash",
			input:    "- Meeting went well",
			wantType: EntryTypeNote,
			wantText: "Meeting went well",
		},
		{
			name:     "event with o",
			input:    "o Team standup at 10am",
			wantType: EntryTypeEvent,
			wantText: "Team standup at 10am",
		},
		{
			name:     "done with x",
			input:    "x Finished report",
			wantType: EntryTypeDone,
			wantText: "Finished report",
		},
		{
			name:     "cancelled with tilde",
			input:    "~ No longer needed",
			wantType: EntryTypeCancelled,
			wantText: "No longer needed",
		},
		{
			name:     "question with question mark",
			input:    "? How does auth work",
			wantType: EntryTypeQuestion,
			wantText: "How does auth work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := parser.ParseLine(tt.input, 1)

			if line.Symbol != tt.wantType {
				t.Errorf("got symbol %v, want %v", line.Symbol, tt.wantType)
			}
			if line.Content != tt.wantText {
				t.Errorf("got content %q, want %q", line.Content, tt.wantText)
			}
			if !line.IsValid {
				t.Errorf("expected line to be valid, got error: %s", line.ErrorMessage)
			}
		})
	}
}

func TestParseLine_Priority(t *testing.T) {
	parser := NewEditableDocumentParser()

	tests := []struct {
		name         string
		input        string
		wantPriority Priority
		wantContent  string
	}{
		{
			name:         "high priority with triple bang",
			input:        ". !!! Urgent task",
			wantPriority: PriorityHigh,
			wantContent:  "Urgent task",
		},
		{
			name:         "medium priority with double bang",
			input:        ". !! Important task",
			wantPriority: PriorityMedium,
			wantContent:  "Important task",
		},
		{
			name:         "low priority with single bang",
			input:        ". ! Minor task",
			wantPriority: PriorityLow,
			wantContent:  "Minor task",
		},
		{
			name:         "no priority",
			input:        ". Normal task",
			wantPriority: PriorityNone,
			wantContent:  "Normal task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := parser.ParseLine(tt.input, 1)

			if line.Priority != tt.wantPriority {
				t.Errorf("got priority %v, want %v", line.Priority, tt.wantPriority)
			}
			if line.Content != tt.wantContent {
				t.Errorf("got content %q, want %q", line.Content, tt.wantContent)
			}
		})
	}
}

func TestParseLine_Indentation(t *testing.T) {
	parser := NewEditableDocumentParser()

	tests := []struct {
		name      string
		input     string
		wantDepth int
	}{
		{
			name:      "no indentation",
			input:     ". Root task",
			wantDepth: 0,
		},
		{
			name:      "one level (2 spaces)",
			input:     "  . Child task",
			wantDepth: 1,
		},
		{
			name:      "two levels (4 spaces)",
			input:     "    . Grandchild task",
			wantDepth: 2,
		},
		{
			name:      "three levels (6 spaces)",
			input:     "      . Great-grandchild",
			wantDepth: 3,
		},
		{
			name:      "tab indentation normalized",
			input:     "\t. Tab indented",
			wantDepth: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := parser.ParseLine(tt.input, 1)

			if line.Depth != tt.wantDepth {
				t.Errorf("got depth %d, want %d", line.Depth, tt.wantDepth)
			}
		})
	}
}

func TestParseLine_InvalidLines(t *testing.T) {
	parser := NewEditableDocumentParser()

	tests := []struct {
		name         string
		input        string
		wantValid    bool
		wantErrorMsg string
	}{
		{
			name:         "unknown symbol",
			input:        "^ Invalid entry",
			wantValid:    false,
			wantErrorMsg: "Unknown entry type",
		},
		{
			name:         "missing content",
			input:        ".",
			wantValid:    false,
			wantErrorMsg: "Entry content required",
		},
		{
			name:         "only whitespace content",
			input:        ".   ",
			wantValid:    false,
			wantErrorMsg: "Entry content required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := parser.ParseLine(tt.input, 1)

			if line.IsValid != tt.wantValid {
				t.Errorf("got IsValid=%v, want %v", line.IsValid, tt.wantValid)
			}
			if tt.wantErrorMsg != "" && line.ErrorMessage != tt.wantErrorMsg {
				t.Errorf("got error %q, want %q", line.ErrorMessage, tt.wantErrorMsg)
			}
		})
	}
}

func TestParseLine_HeaderLines(t *testing.T) {
	parser := NewEditableDocumentParser()

	tests := []struct {
		name       string
		input      string
		wantHeader bool
	}{
		{
			name:       "header line with dashes",
			input:      "── Monday, Jan 27 ──────────────────",
			wantHeader: true,
		},
		{
			name:       "regular entry",
			input:      ". Not a header",
			wantHeader: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := parser.ParseLine(tt.input, 1)

			if line.IsHeader != tt.wantHeader {
				t.Errorf("got IsHeader=%v, want %v", line.IsHeader, tt.wantHeader)
			}
		})
	}
}

func TestParse_FullDocument(t *testing.T) {
	parser := NewEditableDocumentParser()

	input := `── Monday, Jan 27 ──────────────────
. Buy groceries
  . Milk
  . Eggs
- Meeting went well
. !! Urgent: fix prod bug
x Deployed hotfix`

	doc, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Lines) != 7 {
		t.Fatalf("expected 7 lines, got %d", len(doc.Lines))
	}

	if !doc.Lines[0].IsHeader {
		t.Error("expected first line to be header")
	}

	if doc.Lines[1].Content != "Buy groceries" || doc.Lines[1].Depth != 0 {
		t.Errorf("line 1: got content=%q depth=%d, want content='Buy groceries' depth=0",
			doc.Lines[1].Content, doc.Lines[1].Depth)
	}

	if doc.Lines[2].Content != "Milk" || doc.Lines[2].Depth != 1 {
		t.Errorf("line 2: got content=%q depth=%d, want content='Milk' depth=1",
			doc.Lines[2].Content, doc.Lines[2].Depth)
	}

	if doc.Lines[5].Priority != PriorityMedium {
		t.Errorf("line 5: got priority %v, want PriorityMedium", doc.Lines[5].Priority)
	}

	if doc.Lines[6].Symbol != EntryTypeDone {
		t.Errorf("line 6: got symbol %v, want EntryTypeDone", doc.Lines[6].Symbol)
	}
}

func TestParse_EmptyLines(t *testing.T) {
	parser := NewEditableDocumentParser()

	input := `. Task one

. Task two`

	doc, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validLines := 0
	for _, line := range doc.Lines {
		if line.IsValid && !line.IsHeader {
			validLines++
		}
	}

	if validLines != 2 {
		t.Errorf("expected 2 valid entry lines, got %d", validLines)
	}
}

func TestSerialize_BasicEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		want    string
	}{
		{
			name: "task entry",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Buy groceries", Depth: 0},
			},
			want: ". Buy groceries",
		},
		{
			name: "note entry",
			entries: []Entry{
				{Type: EntryTypeNote, Content: "Meeting went well", Depth: 0},
			},
			want: "- Meeting went well",
		},
		{
			name: "event entry",
			entries: []Entry{
				{Type: EntryTypeEvent, Content: "Team standup at 10am", Depth: 0},
			},
			want: "o Team standup at 10am",
		},
		{
			name: "done entry",
			entries: []Entry{
				{Type: EntryTypeDone, Content: "Finished report", Depth: 0},
			},
			want: "x Finished report",
		},
		{
			name: "cancelled entry",
			entries: []Entry{
				{Type: EntryTypeCancelled, Content: "No longer needed", Depth: 0},
			},
			want: "~ No longer needed",
		},
		{
			name: "question entry",
			entries: []Entry{
				{Type: EntryTypeQuestion, Content: "How does auth work", Depth: 0},
			},
			want: "? How does auth work",
		},
		{
			name: "migrated entry",
			entries: []Entry{
				{Type: EntryTypeMigrated, Content: "Moved to next week", Depth: 0},
			},
			want: "> Moved to next week",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Serialize(tt.entries)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSerialize_WithPriority(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		want    string
	}{
		{
			name: "high priority",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Urgent task", Priority: PriorityHigh, Depth: 0},
			},
			want: ". !!! Urgent task",
		},
		{
			name: "medium priority",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Important task", Priority: PriorityMedium, Depth: 0},
			},
			want: ". !! Important task",
		},
		{
			name: "low priority",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Minor task", Priority: PriorityLow, Depth: 0},
			},
			want: ". ! Minor task",
		},
		{
			name: "no priority",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Normal task", Priority: PriorityNone, Depth: 0},
			},
			want: ". Normal task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Serialize(tt.entries)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSerialize_WithIndentation(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		want    string
	}{
		{
			name: "depth 1",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Child task", Depth: 1},
			},
			want: "  . Child task",
		},
		{
			name: "depth 2",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Grandchild task", Depth: 2},
			},
			want: "    . Grandchild task",
		},
		{
			name: "depth 3",
			entries: []Entry{
				{Type: EntryTypeTask, Content: "Great-grandchild", Depth: 3},
			},
			want: "      . Great-grandchild",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Serialize(tt.entries)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSerialize_MultipleEntries(t *testing.T) {
	entries := []Entry{
		{Type: EntryTypeTask, Content: "Buy groceries", Depth: 0},
		{Type: EntryTypeTask, Content: "Milk", Depth: 1},
		{Type: EntryTypeTask, Content: "Eggs", Depth: 1},
		{Type: EntryTypeNote, Content: "Meeting went well", Depth: 0},
		{Type: EntryTypeTask, Content: "Urgent fix", Priority: PriorityMedium, Depth: 0},
		{Type: EntryTypeDone, Content: "Deployed hotfix", Depth: 0},
	}

	want := `. Buy groceries
  . Milk
  . Eggs
- Meeting went well
. !! Urgent fix
x Deployed hotfix`

	got := Serialize(entries)
	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestSerialize_BuildsHierarchyFromParentID(t *testing.T) {
	parentID := int64(1)
	childID := int64(2)

	entries := []Entry{
		{ID: 1, Type: EntryTypeNote, Content: "note 5", Depth: 0},
		{ID: 2, Type: EntryTypeEvent, Content: "event 1", Depth: 0},
		{ID: 3, Type: EntryTypeNote, Content: "note 1x", Depth: 1, ParentID: &childID},
		{ID: 4, Type: EntryTypeNote, Content: "note 6x", Depth: 1, ParentID: &parentID},
	}

	got := Serialize(entries)
	want := `- note 5
  - note 6x
o event 1
  - note 1x`

	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestSerialize_NestedHierarchyFromParentID(t *testing.T) {
	parentID := int64(10)
	childID := int64(20)

	entries := []Entry{
		{ID: 10, Type: EntryTypeTask, Content: "parent", Depth: 0},
		{ID: 20, Type: EntryTypeNote, Content: "child", Depth: 1, ParentID: &parentID},
		{ID: 30, Type: EntryTypeNote, Content: "grandchild", Depth: 2, ParentID: &childID},
	}

	got := Serialize(entries)
	want := `. parent
  - child
    - grandchild`

	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestSerialize_EmptyEntries(t *testing.T) {
	got := Serialize(nil)
	if got != "" {
		t.Errorf("expected empty string for nil entries, got %q", got)
	}

	got = Serialize([]Entry{})
	if got != "" {
		t.Errorf("expected empty string for empty entries, got %q", got)
	}
}
