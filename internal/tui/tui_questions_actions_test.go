package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
)

func newQuestionsModel(entries []domain.Entry) Model {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeQuestions
	model.questionsState = questionsState{
		entries:     entries,
		selectedIdx: 0,
	}
	return model
}

func TestQuestions_SpaceTogglesDoneOnTask(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("space on task in questions view should produce a command")
	}
}

func TestQuestions_SpaceNoOpOnQuestion(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("space on question type should not produce a command (use R to answer)")
	}
}

func TestQuestions_SpaceNoOpOnEmpty(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("space with no questions should not produce a command")
	}
}

func TestQuestions_CancelEntry(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("x on question should produce a cancel command")
	}
}

func TestQuestions_CancelNoOpOnCancelled(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeCancelled, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("x on already cancelled entry should be no-op")
	}
}

func TestQuestions_UncancelEntry(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeCancelled, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("X on cancelled question should produce an uncancel command")
	}
}

func TestQuestions_UncancelNoOpOnNonCancelled(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("X on non-cancelled entry should be no-op")
	}
}

func TestQuestions_EditMode(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Error("e should enter edit mode for question")
	}
	if m.editMode.entryID != 1 {
		t.Errorf("edit mode should target entry 1, got %d", m.editMode.entryID)
	}
	if m.editMode.input.Value() != "What time?" {
		t.Errorf("edit input should contain entry content, got %s", m.editMode.input.Value())
	}
}

func TestQuestions_EditNoOpOnEmpty(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.active {
		t.Error("e with no questions should not enter edit mode")
	}
}

func TestQuestions_DeleteConfirmMode(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.confirmMode.active {
		t.Error("d should enter confirm delete mode for question")
	}
	if m.confirmMode.entryID != 1 {
		t.Errorf("confirm mode should target entry 1, got %d", m.confirmMode.entryID)
	}
}

func TestQuestions_AnswerMode(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.answerMode.active {
		t.Error("R should enter answer mode for question")
	}
	if m.answerMode.questionID != 1 {
		t.Errorf("answer mode should target entry 1, got %d", m.answerMode.questionID)
	}
}

func TestQuestions_AnswerNoOpOnNonQuestion(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.answerMode.active {
		t.Error("R on non-question should not enter answer mode")
	}
}

func TestQuestions_EnterNavigatesToJournal(t *testing.T) {
	sd := scheduledDate(2026, 1, 15)
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: sd},
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeJournal {
		t.Errorf("enter should navigate to journal view, got %v", m.currentView)
	}
	if !m.viewDate.Equal(*sd) {
		t.Errorf("viewDate should be set to entry's scheduled date, got %v", m.viewDate)
	}
	if cmd == nil {
		t.Error("enter should produce a load command for the journal")
	}
}

func TestQuestions_EnterNoOpOnEmpty(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeQuestions {
		t.Error("enter with no entries should stay on questions view")
	}
}

func TestQuestions_NavigationJK(t *testing.T) {
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "Q1", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
		{ID: 2, Content: "Q2", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 2)},
		{ID: 3, Content: "Q3", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 3)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.questionsState.selectedIdx != 1 {
		t.Errorf("j should move selection down, got %d", m.questionsState.selectedIdx)
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.questionsState.selectedIdx != 0 {
		t.Errorf("k should move selection up, got %d", m.questionsState.selectedIdx)
	}
}

func TestQuestions_ActionsOperateOnSelectedEntry(t *testing.T) {
	entries := []domain.Entry{
		{ID: 1, Content: "First question", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
		{ID: 2, Content: "Second question", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 2)},
	}
	model := newQuestionsModel(entries)
	model.questionsState.selectedIdx = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Fatal("e should enter edit mode")
	}
	if m.editMode.entryID != 2 {
		t.Errorf("edit should target selected entry (ID=2), got %d", m.editMode.entryID)
	}
	if m.editMode.input.Value() != "Second question" {
		t.Errorf("edit input should contain selected entry content, got %s", m.editMode.input.Value())
	}
}

func TestQuestions_ViewStackPushedOnNavigate(t *testing.T) {
	sd := scheduledDate(2026, 1, 15)
	model := newQuestionsModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: sd},
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if len(m.viewStack) == 0 {
		t.Fatal("navigating from questions should push to view stack")
	}
	if m.viewStack[len(m.viewStack)-1] != ViewTypeQuestions {
		t.Errorf("view stack should contain questions view, got %v", m.viewStack[len(m.viewStack)-1])
	}
}
