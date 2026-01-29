package domain

import (
	"testing"
	"time"
)

func TestComputeDiff_NoChanges(t *testing.T) {
	entityID1 := NewEntityID()
	entityID2 := NewEntityID()

	original := []Entry{
		{EntityID: entityID1, Type: EntryTypeTask, Content: "Task one", Depth: 0},
		{EntityID: entityID2, Type: EntryTypeNote, Content: "Note two", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &entityID1, Symbol: EntryTypeTask, Content: "Task one", Depth: 0, IsValid: true},
			{EntityID: &entityID2, Symbol: EntryTypeNote, Content: "Note two", Depth: 0, IsValid: true},
		},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Operations) != 0 {
		t.Errorf("expected no operations, got %d", len(changeset.Operations))
	}
	if len(changeset.Errors) != 0 {
		t.Errorf("expected no errors, got %d", len(changeset.Errors))
	}
}

func TestComputeDiff_Insert(t *testing.T) {
	entityID1 := NewEntityID()

	original := []Entry{
		{EntityID: entityID1, Type: EntryTypeTask, Content: "Existing task", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &entityID1, Symbol: EntryTypeTask, Content: "Existing task", Depth: 0, IsValid: true},
			{EntityID: nil, Symbol: EntryTypeTask, Content: "New task", Depth: 0, IsValid: true, LineNumber: 2},
		},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(changeset.Operations))
	}

	op := changeset.Operations[0]
	if op.Type != DiffOpInsert {
		t.Errorf("expected DiffOpInsert, got %v", op.Type)
	}
	if op.Entry.Content != "New task" {
		t.Errorf("expected content 'New task', got %q", op.Entry.Content)
	}
	if op.Entry.Type != EntryTypeTask {
		t.Errorf("expected EntryTypeTask, got %v", op.Entry.Type)
	}
}

func TestComputeDiff_Update(t *testing.T) {
	tests := []struct {
		name             string
		originalType     EntryType
		originalContent  string
		originalPriority Priority
		parsedType       EntryType
		parsedContent    string
		parsedPriority   Priority
		expectUpdate     bool
	}{
		{
			name:            "content changed",
			originalType:    EntryTypeTask,
			originalContent: "Old content",
			parsedType:      EntryTypeTask,
			parsedContent:   "New content",
			expectUpdate:    true,
		},
		{
			name:            "type changed",
			originalType:    EntryTypeTask,
			originalContent: "Same content",
			parsedType:      EntryTypeDone,
			parsedContent:   "Same content",
			expectUpdate:    true,
		},
		{
			name:             "priority changed",
			originalType:     EntryTypeTask,
			originalContent:  "Same content",
			originalPriority: PriorityNone,
			parsedType:       EntryTypeTask,
			parsedContent:    "Same content",
			parsedPriority:   PriorityHigh,
			expectUpdate:     true,
		},
		{
			name:            "no change",
			originalType:    EntryTypeTask,
			originalContent: "Same content",
			parsedType:      EntryTypeTask,
			parsedContent:   "Same content",
			expectUpdate:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entityID := NewEntityID()

			original := []Entry{
				{EntityID: entityID, Type: tt.originalType, Content: tt.originalContent, Priority: tt.originalPriority, Depth: 0},
			}

			parsed := &EditableDocument{
				Lines: []ParsedLine{
					{EntityID: &entityID, Symbol: tt.parsedType, Content: tt.parsedContent, Priority: tt.parsedPriority, Depth: 0, IsValid: true},
				},
			}

			changeset := ComputeDiff(original, parsed)

			if tt.expectUpdate {
				if len(changeset.Operations) != 1 {
					t.Fatalf("expected 1 operation, got %d", len(changeset.Operations))
				}
				if changeset.Operations[0].Type != DiffOpUpdate {
					t.Errorf("expected DiffOpUpdate, got %v", changeset.Operations[0].Type)
				}
			} else {
				if len(changeset.Operations) != 0 {
					t.Errorf("expected no operations, got %d", len(changeset.Operations))
				}
			}
		})
	}
}

func TestComputeDiff_Delete(t *testing.T) {
	entityID1 := NewEntityID()
	entityID2 := NewEntityID()

	original := []Entry{
		{EntityID: entityID1, Type: EntryTypeTask, Content: "Task one", Depth: 0},
		{EntityID: entityID2, Type: EntryTypeNote, Content: "Task two", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &entityID1, Symbol: EntryTypeTask, Content: "Task one", Depth: 0, IsValid: true},
		},
		PendingDeletes: []EntityID{entityID2},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(changeset.Operations))
	}

	op := changeset.Operations[0]
	if op.Type != DiffOpDelete {
		t.Errorf("expected DiffOpDelete, got %v", op.Type)
	}
	if op.EntityID == nil || *op.EntityID != entityID2 {
		t.Errorf("expected EntityID %s, got %v", entityID2, op.EntityID)
	}
}

func TestComputeDiff_DeleteMissingEntries(t *testing.T) {
	entityID1 := NewEntityID()
	entityID2 := NewEntityID()

	original := []Entry{
		{EntityID: entityID1, Type: EntryTypeTask, Content: "Task one", Depth: 0},
		{EntityID: entityID2, Type: EntryTypeNote, Content: "Task two", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines:          []ParsedLine{},
		PendingDeletes: []EntityID{},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Operations) != 2 {
		t.Fatalf("expected 2 delete operations, got %d", len(changeset.Operations))
	}

	deletedIDs := make(map[EntityID]bool)
	for _, op := range changeset.Operations {
		if op.Type != DiffOpDelete {
			t.Errorf("expected DiffOpDelete, got %v", op.Type)
		}
		if op.EntityID != nil {
			deletedIDs[*op.EntityID] = true
		}
	}

	if !deletedIDs[entityID1] {
		t.Errorf("expected entityID1 %s to be deleted", entityID1)
	}
	if !deletedIDs[entityID2] {
		t.Errorf("expected entityID2 %s to be deleted", entityID2)
	}
}

func TestComputeDiff_Migrate(t *testing.T) {
	entityID := NewEntityID()
	migrateDate := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	original := []Entry{
		{EntityID: entityID, Type: EntryTypeTask, Content: "Task to migrate", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &entityID, Symbol: EntryTypeTask, Content: "Task to migrate", Depth: 0, IsValid: true, MigrateTarget: &migrateDate},
		},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(changeset.Operations))
	}

	op := changeset.Operations[0]
	if op.Type != DiffOpMigrate {
		t.Errorf("expected DiffOpMigrate, got %v", op.Type)
	}
	if op.MigrateDate == nil || !op.MigrateDate.Equal(migrateDate) {
		t.Errorf("expected migrate date %v, got %v", migrateDate, op.MigrateDate)
	}
}

func TestComputeDiff_Reparent(t *testing.T) {
	parentID := NewEntityID()
	childID := NewEntityID()

	original := []Entry{
		{EntityID: parentID, Type: EntryTypeTask, Content: "Parent", Depth: 0},
		{EntityID: childID, Type: EntryTypeTask, Content: "Child", Depth: 0, ParentEntityID: nil},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &parentID, Symbol: EntryTypeTask, Content: "Parent", Depth: 0, IsValid: true, LineNumber: 1},
			{EntityID: &childID, Symbol: EntryTypeTask, Content: "Child", Depth: 1, IsValid: true, LineNumber: 2},
		},
	}

	changeset := ComputeDiff(original, parsed)

	hasReparent := false
	for _, op := range changeset.Operations {
		if op.Type == DiffOpReparent {
			hasReparent = true
			if op.NewParentID == nil || *op.NewParentID != parentID {
				t.Errorf("expected new parent %s, got %v", parentID, op.NewParentID)
			}
		}
	}

	if !hasReparent {
		t.Error("expected DiffOpReparent operation")
	}
}

func TestComputeDiff_OrphanChild(t *testing.T) {
	entityID := NewEntityID()

	original := []Entry{}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &entityID, Symbol: EntryTypeTask, Content: "Orphan child", Depth: 1, IsValid: true, LineNumber: 1},
		},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Errors) == 0 {
		t.Error("expected orphan child error")
	}

	foundOrphanError := false
	for _, err := range changeset.Errors {
		if err.LineNumber == 1 {
			foundOrphanError = true
		}
	}
	if !foundOrphanError {
		t.Error("expected error on line 1 for orphan child")
	}
}

func TestComputeDiff_MultipleOperations(t *testing.T) {
	existingID1 := NewEntityID()
	existingID2 := NewEntityID()
	deletedID := NewEntityID()

	original := []Entry{
		{EntityID: existingID1, Type: EntryTypeTask, Content: "Will update", Depth: 0},
		{EntityID: existingID2, Type: EntryTypeTask, Content: "Unchanged", Depth: 0},
		{EntityID: deletedID, Type: EntryTypeTask, Content: "Will delete", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &existingID1, Symbol: EntryTypeDone, Content: "Will update", Depth: 0, IsValid: true},
			{EntityID: &existingID2, Symbol: EntryTypeTask, Content: "Unchanged", Depth: 0, IsValid: true},
			{EntityID: nil, Symbol: EntryTypeNote, Content: "New note", Depth: 0, IsValid: true},
		},
		PendingDeletes: []EntityID{deletedID},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Errors) != 0 {
		t.Errorf("expected no errors, got %d", len(changeset.Errors))
	}

	insertCount := 0
	updateCount := 0
	deleteCount := 0

	for _, op := range changeset.Operations {
		switch op.Type {
		case DiffOpInsert:
			insertCount++
		case DiffOpUpdate:
			updateCount++
		case DiffOpDelete:
			deleteCount++
		}
	}

	if insertCount != 1 {
		t.Errorf("expected 1 insert, got %d", insertCount)
	}
	if updateCount != 1 {
		t.Errorf("expected 1 update, got %d", updateCount)
	}
	if deleteCount != 1 {
		t.Errorf("expected 1 delete, got %d", deleteCount)
	}
}

func TestComputeDiff_SkipsInvalidLines(t *testing.T) {
	entityID := NewEntityID()

	original := []Entry{
		{EntityID: entityID, Type: EntryTypeTask, Content: "Valid", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{EntityID: &entityID, Symbol: EntryTypeTask, Content: "Valid", Depth: 0, IsValid: true},
			{IsValid: false, ErrorMessage: "Invalid line", LineNumber: 2},
		},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Operations) != 0 {
		t.Errorf("expected no operations (invalid line should be skipped), got %d", len(changeset.Operations))
	}
}

func TestComputeDiff_SkipsHeaders(t *testing.T) {
	entityID := NewEntityID()

	original := []Entry{
		{EntityID: entityID, Type: EntryTypeTask, Content: "Task", Depth: 0},
	}

	parsed := &EditableDocument{
		Lines: []ParsedLine{
			{IsHeader: true, IsValid: true, Raw: "── Monday ──"},
			{EntityID: &entityID, Symbol: EntryTypeTask, Content: "Task", Depth: 0, IsValid: true},
		},
	}

	changeset := ComputeDiff(original, parsed)

	if len(changeset.Operations) != 0 {
		t.Errorf("expected no operations (headers should be skipped), got %d", len(changeset.Operations))
	}
}
