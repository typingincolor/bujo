package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/typingincolor/bujo/internal/domain"
)

func newSearchViewModel(entries []domain.Entry) Model {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeSearch
	input := textinput.New()
	input.Blur()
	model.searchView = searchViewState{
		query:       "test",
		results:     entries,
		selectedIdx: 0,
		input:       input,
	}
	return model
}

func TestSearch_SpaceTogglesDone(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("space on task in search view should produce a command")
	}
}

func TestSearch_SpaceNoOpOnEmpty(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("space with no results should not produce a command")
	}
}

func TestSearch_CancelEntry(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("x on task in search view should produce a cancel command")
	}
}

func TestSearch_CancelNoOpOnCancelled(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeCancelled, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("x on already cancelled entry should be no-op")
	}
}

func TestSearch_UncancelEntry(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeCancelled, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("X on cancelled entry in search view should produce an uncancel command")
	}
}

func TestSearch_UncancelNoOpOnNonCancelled(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("X on non-cancelled entry should be no-op")
	}
}

func TestSearch_EditMode(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Error("e should enter edit mode for search result")
	}
	if m.editMode.entryID != 1 {
		t.Errorf("edit mode should target entry 1, got %d", m.editMode.entryID)
	}
	if m.editMode.input.Value() != "Buy milk" {
		t.Errorf("edit input should contain entry content, got %s", m.editMode.input.Value())
	}
}

func TestSearch_EditNoOpOnEmpty(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.active {
		t.Error("e with no results should not enter edit mode")
	}
}

func TestSearch_DeleteConfirmMode(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.confirmMode.active {
		t.Error("d should enter confirm delete mode for search result")
	}
	if m.confirmMode.entryID != 1 {
		t.Errorf("confirm mode should target entry 1, got %d", m.confirmMode.entryID)
	}
}

func TestSearch_CyclePriority(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, Priority: domain.PriorityNone, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("! on search result should produce a priority command")
	}
}

func TestSearch_CyclePriorityNoOpOnEmpty(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("! with no results should not produce a command")
	}
}

func TestSearch_RetypeMode(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.retypeMode.active {
		t.Error("t should enter retype mode for search result")
	}
	if m.retypeMode.entryID != 1 {
		t.Errorf("retype mode should target entry 1, got %d", m.retypeMode.entryID)
	}
}

func TestSearch_RetypeNoOpOnNonCycleable(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Done task", Type: domain.EntryTypeDone, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.retypeMode.active {
		t.Error("t on non-cycleable entry should not enter retype mode")
	}
}

func TestSearch_AnswerMode(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.answerMode.active {
		t.Error("R should enter answer mode for question in search view")
	}
	if m.answerMode.questionID != 1 {
		t.Errorf("answer mode should target entry 1, got %d", m.answerMode.questionID)
	}
}

func TestSearch_AnswerNoOpOnNonQuestion(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.answerMode.active {
		t.Error("R on non-question should not enter answer mode")
	}
}

func TestSearch_EnterPushesViewStack(t *testing.T) {
	sd := scheduledDate(2026, 1, 15)
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: sd},
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if len(m.viewStack) == 0 {
		t.Fatal("navigating from search should push to view stack")
	}
	if m.viewStack[len(m.viewStack)-1] != ViewTypeSearch {
		t.Errorf("view stack should contain search view, got %v", m.viewStack[len(m.viewStack)-1])
	}
}

func TestSearch_ActionsOperateOnSelectedEntry(t *testing.T) {
	entries := []domain.Entry{
		{ID: 1, Content: "First task", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
		{ID: 2, Content: "Second task", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 2)},
	}
	model := newSearchViewModel(entries)
	model.searchView.selectedIdx = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Fatal("e should enter edit mode")
	}
	if m.editMode.entryID != 2 {
		t.Errorf("edit should target selected entry (ID=2), got %d", m.editMode.entryID)
	}
	if m.editMode.input.Value() != "Second task" {
		t.Errorf("edit input should contain selected entry content, got %s", m.editMode.input.Value())
	}
}

func TestSearch_ActionsNoOpWhenInputFocused(t *testing.T) {
	model := newSearchViewModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})
	model.searchView.input.Focus()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.active {
		t.Error("actions should not fire when search input is focused")
	}
}
