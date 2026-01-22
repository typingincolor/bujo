package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/service"
)

func TestUAT_Navigation_NumberKeys_SwitchViews(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	tests := []struct {
		key      rune
		expected ViewType
		name     string
	}{
		{'1', ViewTypeJournal, "Journal"},
		{'2', ViewTypeReview, "Review"},
		{'3', ViewTypePendingTasks, "PendingTasks"},
		{'4', ViewTypeQuestions, "Questions"},
		{'5', ViewTypeHabits, "Habits"},
		{'6', ViewTypeLists, "Lists"},
		{'7', ViewTypeGoals, "Goals"},
		{'8', ViewTypeSearch, "Search"},
		{'9', ViewTypeStats, "Stats"},
		{'0', ViewTypeSettings, "Settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			newModel, _ := model.Update(msg)
			m := newModel.(Model)

			if m.currentView != tt.expected {
				t.Errorf("pressing '%c' should switch to %s view, got %v", tt.key, tt.name, m.currentView)
			}
		})
	}
}

func TestUAT_Navigation_CommandPalette_Opens(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Test Ctrl+P opens command palette
	msg := tea.KeyMsg{Type: tea.KeyCtrlP}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.commandPalette.active {
		t.Error("Ctrl+P should open command palette")
	}
}

func TestUAT_Navigation_CommandPalette_Colon_Opens(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.commandPalette.active {
		t.Error("':' should open command palette")
	}
}

func TestUAT_Navigation_CommandPalette_ShowsCommands(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Open command palette
	msg := tea.KeyMsg{Type: tea.KeyCtrlP}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	view := m.View()

	// Should show available commands
	if !strings.Contains(view, "Journal") {
		t.Error("command palette should show Journal command")
	}
	if !strings.Contains(view, "Habits") {
		t.Error("command palette should show Habits command")
	}
}

func TestUAT_Navigation_CommandPalette_FuzzyFilter(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Open command palette
	model.commandPalette.active = true
	model.commandPalette.query = "hab"

	filtered := model.commandRegistry.Filter("hab")

	found := false
	for _, cmd := range filtered {
		if strings.Contains(strings.ToLower(cmd.Name), "habit") {
			found = true
			break
		}
	}

	if !found {
		t.Error("filtering by 'hab' should show Habits command")
	}
}

func TestUAT_Navigation_Quit(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if cmd != nil {
		t.Error("pressing q at root should not immediately quit")
	}
	if !m.quitConfirmMode.active {
		t.Fatal("quit confirm mode should be active")
	}

	msgYes := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmdQuit := m.Update(msgYes)
	if cmdQuit == nil {
		t.Fatal("confirming with 'y' should return quit command")
	}

	result := cmdQuit()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Error("confirming should trigger quit")
	}
}

func TestUAT_Navigation_Esc_GoesBackFromNestedView(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Start in journal, navigate to habits (adds to view stack)
	model.currentView = ViewTypeJournal
	model.viewStack = []ViewType{} // empty stack

	// Navigate to habits (simulates pressing 5)
	msg5 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, _ := model.Update(msg5)
	m := newModel.(Model)

	if m.currentView != ViewTypeHabits {
		t.Fatalf("expected habits view, got %v", m.currentView)
	}
	if len(m.viewStack) != 1 || m.viewStack[0] != ViewTypeJournal {
		t.Fatalf("expected journal in view stack, got %v", m.viewStack)
	}

	// Press ESC - should go back to journal
	msgEsc := tea.KeyMsg{Type: tea.KeyEsc}
	newModel2, _ := m.Update(msgEsc)
	m2 := newModel2.(Model)

	if m2.currentView != ViewTypeJournal {
		t.Errorf("expected ESC to go back to journal, got %v", m2.currentView)
	}
	if len(m2.viewStack) != 0 {
		t.Errorf("expected empty view stack after going back, got %v", m2.viewStack)
	}
}

func TestUAT_JournalView_LocationPicker_OpensWithAt(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal
	model.agenda = &service.MultiDayAgenda{}

	// Press '@' to open location picker
	msgAt := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}}
	newModel, _ := model.Update(msgAt)
	m := newModel.(Model)

	if !m.setLocationMode.active {
		t.Error("pressing @ should open location picker")
	}
	if !m.setLocationMode.pickerMode {
		t.Error("location mode should be in picker mode")
	}
}

func TestUAT_JournalView_LocationPicker_SelectsPreviousLocation(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal
	model.agenda = &service.MultiDayAgenda{}

	// Simulate having previous locations loaded
	model.setLocationMode = setLocationState{
		active:      true,
		pickerMode:  true,
		date:        model.viewDate,
		locations:   []string{"Home", "Office", "Coffee Shop"},
		selectedIdx: 0,
	}

	// Navigate down
	msgDown := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msgDown)
	m := newModel.(Model)

	if m.setLocationMode.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1, got %d", m.setLocationMode.selectedIdx)
	}
}

func TestUAT_JournalView_LocationPicker_ClosesOnEsc(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal
	model.agenda = &service.MultiDayAgenda{}

	// Open location picker
	model.setLocationMode = setLocationState{
		active:     true,
		pickerMode: true,
		date:       model.viewDate,
		locations:  []string{"Home", "Office"},
	}

	// Press ESC to close
	msgEsc := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msgEsc)
	m := newModel.(Model)

	if m.setLocationMode.active {
		t.Error("ESC should close the location picker")
	}
}

func TestUAT_JournalView_LocationPicker_EnterSelectsFromList(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal
	model.agenda = &service.MultiDayAgenda{}

	// Open location picker with locations
	ti := textinput.New()
	model.setLocationMode = setLocationState{
		active:      true,
		pickerMode:  true,
		date:        model.viewDate,
		input:       ti,
		locations:   []string{"Home", "Office", "Coffee Shop"},
		selectedIdx: 1, // "Office" selected
	}

	// Press Enter to select
	msgEnter := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msgEnter)
	m := newModel.(Model)

	if m.setLocationMode.active {
		t.Error("Enter should close the location picker")
	}
	if cmd == nil {
		t.Error("Enter should return a command to set location")
	}
}

func TestUAT_JournalView_AISummary_CollapsedByDefault(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Summary should be collapsed by default
	if !model.summaryCollapsed {
		t.Error("AI summary should be collapsed by default")
	}
}

func TestUAT_JournalView_AISummary_ToggleWithS(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal
	model.agenda = &service.MultiDayAgenda{}

	// Initially collapsed
	if !model.summaryCollapsed {
		t.Fatal("AI summary should be collapsed by default")
	}

	// Press 's' to toggle
	msgS := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	newModel, _ := model.Update(msgS)
	m := newModel.(Model)

	if m.summaryCollapsed {
		t.Error("AI summary should be expanded after pressing 's'")
	}

	// Press 's' again to collapse
	newModel2, _ := m.Update(msgS)
	m2 := newModel2.(Model)

	if !m2.summaryCollapsed {
		t.Error("AI summary should be collapsed after pressing 's' again")
	}
}

func TestUAT_Navigation_Q_AlwaysShowsQuitConfirm(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to habits first (to have something in the stack)
	model.currentView = ViewTypeHabits
	model.viewStack = []ViewType{ViewTypeJournal}

	// Press Q - should show quit confirmation (not go back)
	msgQ := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msgQ)
	m := newModel.(Model)

	if !m.quitConfirmMode.active {
		t.Error("pressing q should show quit confirmation, even with views in stack")
	}
	// View should not have changed - still in habits
	if m.currentView != ViewTypeHabits {
		t.Errorf("expected to still be in habits view, got %v", m.currentView)
	}
}

// =============================================================================
// Additional Context Tests
// =============================================================================

func TestUAT_HelpSystem_BottomBarShowsCommands(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal

	view := model.View()

	// Bottom bar should show some key hints
	hasHelpHints := strings.Contains(view, "q") || // quit
		strings.Contains(view, "?") || // help
		strings.Contains(view, "1") || // view switch
		strings.Contains(view, "j") || // navigation
		strings.Contains(view, "k")

	if !hasHelpHints {
		t.Error("view should show help hints in bottom bar")
	}
}
