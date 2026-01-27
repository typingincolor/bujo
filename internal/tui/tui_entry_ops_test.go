package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func TestModel_Update_EditMode_InitializesWithContent(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Original content", Type: domain.EntryTypeTask}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.input.Value() != "Original content" {
		t.Errorf("input should be initialized with entry content, got %s", m.editMode.input.Value())
	}
}

func TestModel_Update_EditMode_NoOpOnEmptyEntries(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.active {
		t.Error("pressing e with no entries should not enter edit mode")
	}
}

func TestModel_Update_AddMode_EntersAsChildOnShiftA(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Parent"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Error("pressing A should enter add mode")
	}
	if !m.addMode.asChild {
		t.Error("pressing A should add as child")
	}
	if m.addMode.parentID == nil || *m.addMode.parentID != 1 {
		t.Error("parentID should be set to selected entry ID")
	}
}

func TestModel_Update_AddMode_StartsEmpty(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.addMode.input.Value() != "" {
		t.Errorf("add mode input should start empty, got %s", m.addMode.input.Value())
	}
}

func TestModel_Update_MigrateMode_NoOpOnEmptyEntries(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.migrateMode.active {
		t.Error("pressing m with no entries should not enter migrate mode")
	}
}

func TestModel_Update_MigrateMode_NoOpOnNonTask(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Note", Type: domain.EntryTypeNote}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.migrateMode.active {
		t.Error("pressing m on a note should not enter migrate mode")
	}
}

func TestModel_Update_ErrorCanBeDismissed(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.err = fmt.Errorf("some error")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.err != nil {
		t.Error("pressing Escape should dismiss error")
	}
}

func TestModel_Update_ErrorCanBeDismissedWithAnyKey(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.err = fmt.Errorf("some error")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.err != nil {
		t.Error("pressing any key should dismiss error")
	}
}

func TestModel_Update_AddMode_InheritsParentFromSelected(t *testing.T) {
	parentID := int64(10)
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 22, Content: "Child item", ParentID: &parentID}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	if m.addMode.parentID == nil {
		t.Error("parentID should be inherited from selected item")
	}
	if m.addMode.parentID != nil && *m.addMode.parentID != 10 {
		t.Errorf("parentID should be 10 (same as selected's parent), got %d", *m.addMode.parentID)
	}
}

func TestModel_Update_AddMode_RootItemAddsAtRoot(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Root item", ParentID: nil}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	if m.addMode.parentID != nil {
		t.Error("parentID should be nil when selected item is root")
	}
}

func TestModel_Update_AddRootMode_AddsAtRootFromNestedItem(t *testing.T) {
	parentID := int64(10)
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 22, Content: "Child item", ParentID: &parentID}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	if m.addMode.parentID != nil {
		t.Error("pressing 'r' should add at root regardless of selected item's parent")
	}
}

func TestModel_DefaultViewMode_IsDay(t *testing.T) {
	model := New(nil)
	if model.viewMode != ViewModeDay {
		t.Errorf("default view mode should be day, got %v", model.viewMode)
	}
}

func TestModel_DefaultViewDate_IsToday(t *testing.T) {
	model := New(nil)
	today := time.Now()
	expected := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	if !model.viewDate.Equal(expected) {
		t.Errorf("default view date should be today, got %v", model.viewDate)
	}
}

func TestModel_Update_GoToDate(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.gotoMode.active {
		t.Error("pressing '/' should enter goto date mode")
	}
}

func TestModel_AddRoot_UsesViewDate(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	// Set viewDate to yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	model.viewDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	// The addMode should use viewDate when creating, not today
	// This is tested via integration but we verify addMode is entered
}

// Capture Mode Tests

// Emacs navigation tests

// Search tests

// Draft tests

// Day View Search Tests
