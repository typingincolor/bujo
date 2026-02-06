package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
)

func newPendingTasksModel(entries []domain.Entry) Model {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypePendingTasks
	model.pendingTasksState = pendingTasksState{
		entries:      entries,
		selectedIdx:  0,
		parentChains: make(map[int64][]domain.Entry),
	}
	return model
}

func scheduledDate(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &t
}

func TestPendingTasks_SpaceTogglesDone(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if cmd == nil {
		t.Error("space on pending task should produce a command")
	}
	_ = m
}

func TestPendingTasks_SpaceNoOpOnEmptyEntries(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("space with no pending tasks should not produce a command")
	}
}

func TestPendingTasks_CancelEntry(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("x on pending task should produce a cancel command")
	}
}

func TestPendingTasks_CancelNoOpOnCancelledEntry(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeCancelled, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("x on already cancelled entry should be no-op")
	}
}

func TestPendingTasks_UncancelEntry(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeCancelled, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("X on cancelled pending task should produce an uncancel command")
	}
}

func TestPendingTasks_UncancelNoOpOnNonCancelledEntry(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("X on non-cancelled entry should be no-op")
	}
}

func TestPendingTasks_EditMode(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Error("e should enter edit mode for pending task")
	}
	if m.editMode.entryID != 1 {
		t.Errorf("edit mode should target entry 1, got %d", m.editMode.entryID)
	}
	if m.editMode.input.Value() != "Buy milk" {
		t.Errorf("edit input should contain entry content, got %s", m.editMode.input.Value())
	}
}

func TestPendingTasks_EditNoOpOnEmptyEntries(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.active {
		t.Error("e with no pending tasks should not enter edit mode")
	}
}

func TestPendingTasks_DeleteConfirmMode(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.confirmMode.active {
		t.Error("d should enter confirm delete mode for pending task")
	}
	if m.confirmMode.entryID != 1 {
		t.Errorf("confirm mode should target entry 1, got %d", m.confirmMode.entryID)
	}
}

func TestPendingTasks_MigrateMode(t *testing.T) {
	sd := scheduledDate(2026, 1, 1)
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: sd},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.migrateMode.active {
		t.Error("> should enter migrate mode for pending task")
	}
	if m.migrateMode.entryID != 1 {
		t.Errorf("migrate mode should target entry 1, got %d", m.migrateMode.entryID)
	}
}

func TestPendingTasks_MigrateNoOpOnNonTask(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "A note", Type: domain.EntryTypeNote, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.migrateMode.active {
		t.Error("> on non-task should not enter migrate mode")
	}
}

func TestPendingTasks_RetypeMode(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.retypeMode.active {
		t.Error("t should enter retype mode for pending task")
	}
	if m.retypeMode.entryID != 1 {
		t.Errorf("retype mode should target entry 1, got %d", m.retypeMode.entryID)
	}
}

func TestPendingTasks_CyclePriority(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, Priority: domain.PriorityNone, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("! on pending task should produce a priority command")
	}
}

func TestPendingTasks_CyclePriorityNoOpOnEmpty(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("! with no pending tasks should not produce a command")
	}
}

func TestPendingTasks_AnswerMode(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "What time?", Type: domain.EntryTypeQuestion, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.answerMode.active {
		t.Error("R should enter answer mode for question in pending tasks")
	}
	if m.answerMode.questionID != 1 {
		t.Errorf("answer mode should target entry 1, got %d", m.answerMode.questionID)
	}
}

func TestPendingTasks_AnswerNoOpOnNonQuestion(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.answerMode.active {
		t.Error("R on non-question should not enter answer mode")
	}
}

func TestPendingTasks_MoveToList(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("L on pending task should produce a move-to-list command")
	}
}

func TestPendingTasks_MoveToListNoOpOnEmpty(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("L with no pending tasks should not produce a command")
	}
}

func TestPendingTasks_EnterNavigatesToJournal(t *testing.T) {
	sd := scheduledDate(2026, 1, 15)
	model := newPendingTasksModel([]domain.Entry{
		{ID: 1, Content: "Buy milk", Type: domain.EntryTypeTask, ScheduledDate: sd},
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

func TestPendingTasks_EnterNoOpOnEmpty(t *testing.T) {
	model := newPendingTasksModel([]domain.Entry{})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypePendingTasks {
		t.Error("enter with no entries should stay on pending tasks")
	}
}

func TestPendingTasks_ActionsOperateOnSelectedEntry(t *testing.T) {
	entries := []domain.Entry{
		{ID: 1, Content: "First task", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 1)},
		{ID: 2, Content: "Second task", Type: domain.EntryTypeTask, ScheduledDate: scheduledDate(2026, 1, 2)},
	}
	model := newPendingTasksModel(entries)
	model.pendingTasksState.selectedIdx = 1

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
