package tui

import (
	"testing"

	"github.com/typingincolor/bujo/internal/domain"
)

func TestUpdateKeyMapForEntry_Priority_AlwaysEnabled(t *testing.T) {
	testCases := []struct {
		name        string
		entryType   domain.EntryType
		wantEnabled bool
	}{
		{"task", domain.EntryTypeTask, true},
		{"note", domain.EntryTypeNote, true},
		{"event", domain.EntryTypeEvent, true},
		{"done", domain.EntryTypeDone, true},
		{"migrated", domain.EntryTypeMigrated, true},
		{"cancelled", domain.EntryTypeCancelled, true},
		{"question", domain.EntryTypeQuestion, true},
		{"answered", domain.EntryTypeAnswered, true},
		{"answer", domain.EntryTypeAnswer, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			km := DefaultKeyMap()
			km.Priority.SetEnabled(false)

			entry := domain.NewEntry(tc.entryType, "test", nil)
			UpdateKeyMapForEntry(&km, entry)

			if km.Priority.Enabled() != tc.wantEnabled {
				t.Errorf("Priority.Enabled() = %v, want %v for entry type %s",
					km.Priority.Enabled(), tc.wantEnabled, tc.name)
			}
		})
	}
}

func TestUpdateKeyMapForEntry_Delete_AlwaysEnabled(t *testing.T) {
	testCases := []struct {
		name        string
		entryType   domain.EntryType
		wantEnabled bool
	}{
		{"task", domain.EntryTypeTask, true},
		{"note", domain.EntryTypeNote, true},
		{"event", domain.EntryTypeEvent, true},
		{"done", domain.EntryTypeDone, true},
		{"migrated", domain.EntryTypeMigrated, true},
		{"cancelled", domain.EntryTypeCancelled, true},
		{"question", domain.EntryTypeQuestion, true},
		{"answered", domain.EntryTypeAnswered, true},
		{"answer", domain.EntryTypeAnswer, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			km := DefaultKeyMap()
			km.Delete.SetEnabled(false)

			entry := domain.NewEntry(tc.entryType, "test", nil)
			UpdateKeyMapForEntry(&km, entry)

			if km.Delete.Enabled() != tc.wantEnabled {
				t.Errorf("Delete.Enabled() = %v, want %v for entry type %s",
					km.Delete.Enabled(), tc.wantEnabled, tc.name)
			}
		})
	}
}

func TestResetKeyMapEnabled_IncludesPriorityAndDelete(t *testing.T) {
	km := DefaultKeyMap()

	km.Priority.SetEnabled(false)
	km.Delete.SetEnabled(false)

	ResetKeyMapEnabled(&km)

	if !km.Priority.Enabled() {
		t.Error("ResetKeyMapEnabled should enable Priority")
	}
	if !km.Delete.Enabled() {
		t.Error("ResetKeyMapEnabled should enable Delete")
	}
}

func TestUpdateKeyMapForEntry_MoveToRoot_ContextDependent(t *testing.T) {
	parentID := int64(1)

	testCases := []struct {
		name        string
		parentID    *int64
		wantEnabled bool
	}{
		{"with parent", &parentID, true},
		{"without parent (root)", nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			km := DefaultKeyMap()
			km.MoveToRoot.SetEnabled(!tc.wantEnabled)

			entry := domain.NewEntry(domain.EntryTypeTask, "test", nil)
			entry.ParentID = tc.parentID
			UpdateKeyMapForEntry(&km, entry)

			if km.MoveToRoot.Enabled() != tc.wantEnabled {
				t.Errorf("MoveToRoot.Enabled() = %v, want %v for parent=%v",
					km.MoveToRoot.Enabled(), tc.wantEnabled, tc.parentID)
			}
		})
	}
}

func TestResetKeyMapEnabled_IncludesMoveToRoot(t *testing.T) {
	km := DefaultKeyMap()

	km.MoveToRoot.SetEnabled(false)

	ResetKeyMapEnabled(&km)

	if !km.MoveToRoot.Enabled() {
		t.Error("ResetKeyMapEnabled should enable MoveToRoot")
	}
}
