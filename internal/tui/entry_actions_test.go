package tui

import (
	"testing"

	"github.com/typingincolor/bujo/internal/domain"
)

func TestCanCancel(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can be cancelled", domain.EntryTypeTask, true},
		{"note can be cancelled", domain.EntryTypeNote, true},
		{"event can be cancelled", domain.EntryTypeEvent, true},
		{"done can be cancelled", domain.EntryTypeDone, true},
		{"migrated can be cancelled", domain.EntryTypeMigrated, true},
		{"cancelled cannot be cancelled", domain.EntryTypeCancelled, false},
		{"question can be cancelled", domain.EntryTypeQuestion, true},
		{"answered can be cancelled", domain.EntryTypeAnswered, true},
		{"answer can be cancelled", domain.EntryTypeAnswer, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanCancel(entry); got != tt.want {
				t.Errorf("CanCancel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanUncancel(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task cannot be uncancelled", domain.EntryTypeTask, false},
		{"note cannot be uncancelled", domain.EntryTypeNote, false},
		{"event cannot be uncancelled", domain.EntryTypeEvent, false},
		{"done cannot be uncancelled", domain.EntryTypeDone, false},
		{"migrated cannot be uncancelled", domain.EntryTypeMigrated, false},
		{"cancelled can be uncancelled", domain.EntryTypeCancelled, true},
		{"question cannot be uncancelled", domain.EntryTypeQuestion, false},
		{"answered cannot be uncancelled", domain.EntryTypeAnswered, false},
		{"answer cannot be uncancelled", domain.EntryTypeAnswer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanUncancel(entry); got != tt.want {
				t.Errorf("CanUncancel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanCycleType(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can cycle type", domain.EntryTypeTask, true},
		{"note can cycle type", domain.EntryTypeNote, true},
		{"event can cycle type", domain.EntryTypeEvent, true},
		{"done cannot cycle type", domain.EntryTypeDone, false},
		{"migrated cannot cycle type", domain.EntryTypeMigrated, false},
		{"cancelled cannot cycle type", domain.EntryTypeCancelled, false},
		{"question can cycle type", domain.EntryTypeQuestion, true},
		{"answered cannot cycle type", domain.EntryTypeAnswered, false},
		{"answer cannot cycle type", domain.EntryTypeAnswer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanCycleType(entry); got != tt.want {
				t.Errorf("CanCycleType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanEdit(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can be edited", domain.EntryTypeTask, true},
		{"note can be edited", domain.EntryTypeNote, true},
		{"event can be edited", domain.EntryTypeEvent, true},
		{"done can be edited", domain.EntryTypeDone, true},
		{"migrated can be edited", domain.EntryTypeMigrated, true},
		{"cancelled cannot be edited", domain.EntryTypeCancelled, false},
		{"question can be edited", domain.EntryTypeQuestion, true},
		{"answered can be edited", domain.EntryTypeAnswered, true},
		{"answer can be edited", domain.EntryTypeAnswer, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanEdit(entry); got != tt.want {
				t.Errorf("CanEdit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanMigrate(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can be migrated", domain.EntryTypeTask, true},
		{"note cannot be migrated", domain.EntryTypeNote, false},
		{"event cannot be migrated", domain.EntryTypeEvent, false},
		{"done cannot be migrated", domain.EntryTypeDone, false},
		{"migrated cannot be migrated", domain.EntryTypeMigrated, false},
		{"cancelled cannot be migrated", domain.EntryTypeCancelled, false},
		{"question cannot be migrated", domain.EntryTypeQuestion, false},
		{"answered cannot be migrated", domain.EntryTypeAnswered, false},
		{"answer cannot be migrated", domain.EntryTypeAnswer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanMigrate(entry); got != tt.want {
				t.Errorf("CanMigrate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanAnswer(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task cannot be answered", domain.EntryTypeTask, false},
		{"note cannot be answered", domain.EntryTypeNote, false},
		{"event cannot be answered", domain.EntryTypeEvent, false},
		{"done cannot be answered", domain.EntryTypeDone, false},
		{"migrated cannot be answered", domain.EntryTypeMigrated, false},
		{"cancelled cannot be answered", domain.EntryTypeCancelled, false},
		{"question can be answered", domain.EntryTypeQuestion, true},
		{"answered cannot be answered", domain.EntryTypeAnswered, false},
		{"answer cannot be answered", domain.EntryTypeAnswer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanAnswer(entry); got != tt.want {
				t.Errorf("CanAnswer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanAddChild(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can add child", domain.EntryTypeTask, true},
		{"note can add child", domain.EntryTypeNote, true},
		{"event can add child", domain.EntryTypeEvent, true},
		{"done can add child", domain.EntryTypeDone, true},
		{"migrated can add child", domain.EntryTypeMigrated, true},
		{"cancelled can add child", domain.EntryTypeCancelled, true},
		{"question cannot add child", domain.EntryTypeQuestion, false},
		{"answered can add child", domain.EntryTypeAnswered, true},
		{"answer can add child", domain.EntryTypeAnswer, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanAddChild(entry); got != tt.want {
				t.Errorf("CanAddChild() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanMoveToList(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can move to list", domain.EntryTypeTask, true},
		{"note cannot move to list", domain.EntryTypeNote, false},
		{"event cannot move to list", domain.EntryTypeEvent, false},
		{"done cannot move to list", domain.EntryTypeDone, false},
		{"migrated cannot move to list", domain.EntryTypeMigrated, false},
		{"cancelled cannot move to list", domain.EntryTypeCancelled, false},
		{"question cannot move to list", domain.EntryTypeQuestion, false},
		{"answered cannot move to list", domain.EntryTypeAnswered, false},
		{"answer cannot move to list", domain.EntryTypeAnswer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanMoveToList(entry); got != tt.want {
				t.Errorf("CanMoveToList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanMoveToRoot(t *testing.T) {
	tests := []struct {
		name      string
		hasParent bool
		want      bool
	}{
		{"entry with parent can move to root", true, true},
		{"entry without parent cannot move to root", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parentID *int64
			if tt.hasParent {
				id := int64(1)
				parentID = &id
			}
			entry := domain.Entry{Type: domain.EntryTypeTask, ParentID: parentID}
			if got := CanMoveToRoot(entry); got != tt.want {
				t.Errorf("CanMoveToRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanCyclePriority(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can cycle priority", domain.EntryTypeTask, true},
		{"note can cycle priority", domain.EntryTypeNote, true},
		{"event can cycle priority", domain.EntryTypeEvent, true},
		{"done can cycle priority", domain.EntryTypeDone, true},
		{"migrated can cycle priority", domain.EntryTypeMigrated, true},
		{"cancelled can cycle priority", domain.EntryTypeCancelled, true},
		{"question can cycle priority", domain.EntryTypeQuestion, true},
		{"answered can cycle priority", domain.EntryTypeAnswered, true},
		{"answer can cycle priority", domain.EntryTypeAnswer, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanCyclePriority(entry); got != tt.want {
				t.Errorf("CanCyclePriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanDelete(t *testing.T) {
	tests := []struct {
		name      string
		entryType domain.EntryType
		want      bool
	}{
		{"task can be deleted", domain.EntryTypeTask, true},
		{"note can be deleted", domain.EntryTypeNote, true},
		{"event can be deleted", domain.EntryTypeEvent, true},
		{"done can be deleted", domain.EntryTypeDone, true},
		{"migrated can be deleted", domain.EntryTypeMigrated, true},
		{"cancelled can be deleted", domain.EntryTypeCancelled, true},
		{"question can be deleted", domain.EntryTypeQuestion, true},
		{"answered can be deleted", domain.EntryTypeAnswered, true},
		{"answer can be deleted", domain.EntryTypeAnswer, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := domain.Entry{Type: tt.entryType}
			if got := CanDelete(entry); got != tt.want {
				t.Errorf("CanDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateKeyMapForEntry(t *testing.T) {
	tests := []struct {
		name           string
		entryType      domain.EntryType
		hasParent      bool
		wantCancel     bool
		wantUncancel   bool
		wantEdit       bool
		wantRetype     bool
		wantAddChild   bool
		wantMigrate    bool
		wantMoveToList bool
		wantAnswer     bool
	}{
		{
			name:           "task enables most actions",
			entryType:      domain.EntryTypeTask,
			hasParent:      false,
			wantCancel:     true,
			wantUncancel:   false,
			wantEdit:       true,
			wantRetype:     true,
			wantAddChild:   true,
			wantMigrate:    true,
			wantMoveToList: true,
			wantAnswer:     false,
		},
		{
			name:           "cancelled entry disables edit, cancel, retype",
			entryType:      domain.EntryTypeCancelled,
			hasParent:      false,
			wantCancel:     false,
			wantUncancel:   true,
			wantEdit:       false,
			wantRetype:     false,
			wantAddChild:   true,
			wantMigrate:    false,
			wantMoveToList: false,
			wantAnswer:     false,
		},
		{
			name:           "question enables answer, disables addChild",
			entryType:      domain.EntryTypeQuestion,
			hasParent:      false,
			wantCancel:     true,
			wantUncancel:   false,
			wantEdit:       true,
			wantRetype:     true,
			wantAddChild:   false,
			wantMigrate:    false,
			wantMoveToList: false,
			wantAnswer:     true,
		},
		{
			name:           "note disables migrate and moveToList",
			entryType:      domain.EntryTypeNote,
			hasParent:      false,
			wantCancel:     true,
			wantUncancel:   false,
			wantEdit:       true,
			wantRetype:     true,
			wantAddChild:   true,
			wantMigrate:    false,
			wantMoveToList: false,
			wantAnswer:     false,
		},
		{
			name:           "done entry disables retype",
			entryType:      domain.EntryTypeDone,
			hasParent:      false,
			wantCancel:     true,
			wantUncancel:   false,
			wantEdit:       true,
			wantRetype:     false,
			wantAddChild:   true,
			wantMigrate:    false,
			wantMoveToList: false,
			wantAnswer:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km := DefaultKeyMap()
			var parentID *int64
			if tt.hasParent {
				id := int64(1)
				parentID = &id
			}
			entry := domain.Entry{Type: tt.entryType, ParentID: parentID}

			UpdateKeyMapForEntry(&km, entry)

			if km.CancelEntry.Enabled() != tt.wantCancel {
				t.Errorf("CancelEntry.Enabled() = %v, want %v", km.CancelEntry.Enabled(), tt.wantCancel)
			}
			if km.UncancelEntry.Enabled() != tt.wantUncancel {
				t.Errorf("UncancelEntry.Enabled() = %v, want %v", km.UncancelEntry.Enabled(), tt.wantUncancel)
			}
			if km.Edit.Enabled() != tt.wantEdit {
				t.Errorf("Edit.Enabled() = %v, want %v", km.Edit.Enabled(), tt.wantEdit)
			}
			if km.Retype.Enabled() != tt.wantRetype {
				t.Errorf("Retype.Enabled() = %v, want %v", km.Retype.Enabled(), tt.wantRetype)
			}
			if km.AddChild.Enabled() != tt.wantAddChild {
				t.Errorf("AddChild.Enabled() = %v, want %v", km.AddChild.Enabled(), tt.wantAddChild)
			}
			if km.Migrate.Enabled() != tt.wantMigrate {
				t.Errorf("Migrate.Enabled() = %v, want %v", km.Migrate.Enabled(), tt.wantMigrate)
			}
			if km.MoveToList.Enabled() != tt.wantMoveToList {
				t.Errorf("MoveToList.Enabled() = %v, want %v", km.MoveToList.Enabled(), tt.wantMoveToList)
			}
			if km.Answer.Enabled() != tt.wantAnswer {
				t.Errorf("Answer.Enabled() = %v, want %v", km.Answer.Enabled(), tt.wantAnswer)
			}
		})
	}
}
