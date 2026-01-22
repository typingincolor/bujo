package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func TestModel_ViewSwitch_Key1_SwitchesToJournal(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits // Start in habits view

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeJournal {
		t.Errorf("expected ViewTypeJournal, got %v", m.currentView)
	}
}

func TestModel_New_DefaultsToJournalView(t *testing.T) {
	model := New(nil)

	if model.currentView != ViewTypeJournal {
		t.Errorf("expected default view to be Journal, got %v", model.currentView)
	}
}

func TestModel_View_StatusBar_ShowsCurrentView(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.agenda = &service.MultiDayAgenda{}

	// Test Journal view
	model.currentView = ViewTypeJournal
	view := model.View()
	if !strings.Contains(view, "Journal") {
		t.Error("status bar should show 'Journal' for journal view")
	}

	// Test Habits view
	model.currentView = ViewTypeHabits
	view = model.View()
	if !strings.Contains(view, "Habits") {
		t.Error("status bar should show 'Habits' for habits view")
	}

	// Test Lists view
	model.currentView = ViewTypeLists
	view = model.View()
	if !strings.Contains(view, "Lists") {
		t.Error("status bar should show 'Lists' for lists view")
	}
}

// ============================================================================
// Phase 4: Habits View Tests
// ============================================================================

func TestModel_HabitsView_ShowsStreak(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeHabits
	model.habitState = habitState{
		habits: []service.HabitStatus{
			{ID: 1, Name: "Meditation", CurrentStreak: 5},
		},
	}

	view := model.View()

	if !strings.Contains(view, "5") {
		t.Error("view should contain streak count")
	}
}

func TestModel_HabitsView_Navigation_BoundsCheck(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState = habitState{
		habits: []service.HabitStatus{
			{ID: 1, Name: "Meditation"},
			{ID: 2, Name: "Exercise"},
		},
		selectedIdx: 0,
	}

	// Try to go up from first item - should stay at 0
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.habitState.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 (bounds check), got %d", m.habitState.selectedIdx)
	}

	// Go to last item and try to go down - should stay at last
	m.habitState.selectedIdx = 1
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.habitState.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (bounds check), got %d", m.habitState.selectedIdx)
	}
}

func TestModel_HabitsView_EmptyState(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeHabits
	model.habitState = habitState{
		habits: []service.HabitStatus{},
	}

	view := model.View()

	if !strings.Contains(view, "No habits") {
		t.Error("view should show 'No habits' message when empty")
	}
}

func TestModel_HabitsView_ShowsCompletionRate(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeHabits
	model.habitState = habitState{
		habits: []service.HabitStatus{
			{Name: "Exercise", CompletionPercent: 85.5},
		},
	}

	view := model.View()

	if !strings.Contains(view, "85%") && !strings.Contains(view, "86%") {
		t.Errorf("view should show completion rate, got: %s", view)
	}
}

func TestModel_HabitsView_LogHabit_Space_LogsSelectedHabit(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState = habitState{
		habits: []service.HabitStatus{
			{ID: 1, Name: "Exercise"},
			{ID: 2, Name: "Reading"},
		},
		selectedIdx: 1,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should return a command to log the habit
	if cmd == nil {
		t.Error("expected a command to log the habit")
	}

	// Selected index should remain the same
	if m.habitState.selectedIdx != 1 {
		t.Errorf("expected selectedIdx to remain 1, got %d", m.habitState.selectedIdx)
	}
}

func TestModel_HabitsView_ToggleMonthView_W(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState = habitState{
		habits:   []service.HabitStatus{{Name: "Exercise"}},
		viewMode: HabitViewModeWeek,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}

	// First press: Week -> Month
	newModel, _ := model.Update(msg)
	m := newModel.(Model)
	if m.habitState.viewMode != HabitViewModeMonth {
		t.Errorf("expected HabitViewModeMonth after first 'w', got %v", m.habitState.viewMode)
	}

	// Second press: Month -> Quarter
	newModel, _ = m.Update(msg)
	m = newModel.(Model)
	if m.habitState.viewMode != HabitViewModeQuarter {
		t.Errorf("expected HabitViewModeQuarter after second 'w', got %v", m.habitState.viewMode)
	}

	// Third press: Quarter -> Week (cycle back)
	newModel, _ = m.Update(msg)
	m = newModel.(Model)
	if m.habitState.viewMode != HabitViewModeWeek {
		t.Errorf("expected HabitViewModeWeek after third 'w', got %v", m.habitState.viewMode)
	}
}

// Lists View Tests

func TestModel_ListsView_ShowsItemCount(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeLists
	model.listState = listState{
		lists: []domain.List{
			{ID: 1, Name: "Shopping"},
		},
		summaries: map[int64]*service.ListSummary{
			1: {ID: 1, Name: "Shopping", TotalItems: 5, DoneItems: 2},
		},
	}

	view := model.View()

	// Should show item count like "5 items" or "2/5"
	if !strings.Contains(view, "5") {
		t.Errorf("view should show total items count, got: %s", view)
	}
}

func TestModel_ListsView_ShowsCompletionProgress(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeLists
	model.listState = listState{
		lists: []domain.List{
			{ID: 1, Name: "Shopping"},
		},
		summaries: map[int64]*service.ListSummary{
			1: {ID: 1, Name: "Shopping", TotalItems: 5, DoneItems: 2},
		},
	}

	view := model.View()

	// Should show progress like "2/5" or "40%"
	if !strings.Contains(view, "2/5") && !strings.Contains(view, "40%") {
		t.Errorf("view should show completion progress (2/5 or 40%%), got: %s", view)
	}
}

func TestModel_ListsView_EmptyState(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeLists
	model.listState = listState{
		lists: []domain.List{},
	}

	view := model.View()

	if !strings.Contains(view, "No lists") {
		t.Error("view should show 'No lists' message when empty")
	}
}

func TestModel_ListItemsView_RendersItems(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeListItems
	model.listState = listState{
		lists:         []domain.List{{ID: 1, Name: "Shopping"}},
		currentListID: 1,
		items: []domain.ListItem{
			{VersionInfo: domain.VersionInfo{EntityID: domain.EntityID("item1")}, Content: "Milk", Type: domain.ListItemTypeTask},
			{VersionInfo: domain.VersionInfo{EntityID: domain.EntityID("item2")}, Content: "Bread", Type: domain.ListItemTypeDone},
		},
	}

	view := model.View()

	if !strings.Contains(view, "Milk") {
		t.Error("view should show 'Milk' item")
	}
	if !strings.Contains(view, "Bread") {
		t.Error("view should show 'Bread' item")
	}
}

func TestModel_ListItemsView_Navigation_J_MovesDown(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeListItems
	model.listState = listState{
		items: []domain.ListItem{
			{Content: "Milk"},
			{Content: "Bread"},
		},
		selectedItemIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.listState.selectedItemIdx != 1 {
		t.Errorf("expected selectedItemIdx 1, got %d", m.listState.selectedItemIdx)
	}
}

func TestModel_ListItemsView_Navigation_K_MovesUp(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeListItems
	model.listState = listState{
		items: []domain.ListItem{
			{Content: "Milk"},
			{Content: "Bread"},
		},
		selectedItemIdx: 1,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.listState.selectedItemIdx != 0 {
		t.Errorf("expected selectedItemIdx 0, got %d", m.listState.selectedItemIdx)
	}
}

func TestModel_ListItemsView_ToggleDone_Space(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeListItems
	model.listState = listState{
		lists:         []domain.List{{ID: 1, Name: "Shopping"}},
		currentListID: 1,
		items: []domain.ListItem{
			{VersionInfo: domain.VersionInfo{EntityID: domain.EntityID("item1")}, Content: "Milk", Type: domain.ListItemTypeTask},
		},
		selectedItemIdx: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmd := model.Update(msg)

	// Should return a command to toggle the item
	if cmd == nil {
		t.Error("expected a command to toggle the item")
	}
}

func TestModel_ListItemsView_EmptyState(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeListItems
	model.listState = listState{
		lists:         []domain.List{{ID: 1, Name: "Shopping"}},
		currentListID: 1,
		items:         []domain.ListItem{},
	}

	view := model.View()

	if !strings.Contains(view, "No items") || !strings.Contains(view, "empty") {
		t.Error("view should show empty state message")
	}
}

// Command Palette Tests

func TestModel_CommandPalette_CtrlP_Opens(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal

	msg := tea.KeyMsg{Type: tea.KeyCtrlP}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.commandPalette.active {
		t.Error("expected command palette to be active after Ctrl+P")
	}
}

func TestModel_CommandPalette_Colon_Opens(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.commandPalette.active {
		t.Error("expected command palette to be active after ':'")
	}
}

func TestModel_CommandPalette_Escape_Closes(t *testing.T) {
	model := New(nil)
	model.commandPalette.active = true
	model.commandPalette.filtered = model.commandRegistry.All()

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.commandPalette.active {
		t.Error("expected command palette to be closed after Escape")
	}
}

func TestModel_CommandPalette_RendersOverlay(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.commandPalette.active = true
	model.commandPalette.filtered = model.commandRegistry.All()

	view := model.View()

	if !strings.Contains(view, "Command Palette") && !strings.Contains(view, ">") {
		t.Errorf("view should show command palette, got: %s", view)
	}
}

func TestModel_CommandPalette_ShowsAllCommands(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.commandPalette.active = true
	model.commandPalette.filtered = model.commandRegistry.All()

	view := model.View()

	if !strings.Contains(view, "Journal") {
		t.Error("view should show 'Switch to Journal' command")
	}
	if !strings.Contains(view, "Habits") {
		t.Error("view should show 'Switch to Habits' command")
	}
}

func TestModel_CommandPalette_ShowsKeybindings(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.commandPalette.active = true
	model.commandPalette.filtered = model.commandRegistry.All()

	view := model.View()

	// Should show keybindings like "1", "2", "3"
	if !strings.Contains(view, "1") || !strings.Contains(view, "2") {
		t.Errorf("view should show keybindings, got: %s", view)
	}
}

func TestModel_CommandPalette_FiltersByQuery(t *testing.T) {
	model := New(nil)
	model.commandPalette.active = true
	model.commandPalette.query = "journal"
	model.commandPalette.filtered = model.commandRegistry.Filter("journal")

	if len(model.commandPalette.filtered) == 0 {
		t.Error("expected filtered commands to contain journal-related commands")
	}

	for _, cmd := range model.commandPalette.filtered {
		if !strings.Contains(strings.ToLower(cmd.Name), "journal") &&
			!strings.Contains(strings.ToLower(cmd.Description), "journal") {
			t.Errorf("filtered command '%s' should match 'journal'", cmd.Name)
		}
	}
}

func TestModel_CommandPalette_FuzzyMatch(t *testing.T) {
	model := New(nil)

	// "swj" should fuzzy match "Switch to Journal"
	filtered := model.commandRegistry.Filter("swj")

	found := false
	for _, cmd := range filtered {
		if strings.Contains(cmd.Name, "Journal") {
			found = true
			break
		}
	}

	if !found {
		t.Error("fuzzy search 'swj' should match 'Switch to Journal'")
	}
}

func TestModel_CommandPalette_Navigation_J_MovesDown(t *testing.T) {
	model := New(nil)
	model.commandPalette.active = true
	model.commandPalette.filtered = model.commandRegistry.All()
	model.commandPalette.selectedIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.commandPalette.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1, got %d", m.commandPalette.selectedIdx)
	}
}

func TestModel_CommandPalette_Navigation_K_MovesUp(t *testing.T) {
	model := New(nil)
	model.commandPalette.active = true
	model.commandPalette.filtered = model.commandRegistry.All()
	model.commandPalette.selectedIdx = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.commandPalette.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0, got %d", m.commandPalette.selectedIdx)
	}
}

func TestModel_CommandPalette_Enter_ExecutesCommand(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.commandPalette.active = true
	model.commandPalette.filtered = model.commandRegistry.All()

	// Find the "Switch to Habits" command
	for i, cmd := range model.commandPalette.filtered {
		if strings.Contains(cmd.Name, "Habits") {
			model.commandPalette.selectedIdx = i
			break
		}
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.commandPalette.active {
		t.Error("expected command palette to close after executing command")
	}
	if m.currentView != ViewTypeHabits {
		t.Errorf("expected view to be Habits after executing command, got %v", m.currentView)
	}
}

// Theme Tests
