package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

// =============================================================================
// UAT Section 6: General Navigation
// =============================================================================

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
// UAT Section 7: Journal View
// =============================================================================

func TestUAT_JournalView_ShowsTodaysEntries(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create entries for today
	opts := service.LogEntriesOptions{Date: time.Now()}
	if _, err := bujoSvc.LogEntries(ctx, ". Complete project", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}
	if _, err := bujoSvc.LogEntries(ctx, "- Meeting notes", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Init loads journal
	cmd := model.Init()
	loadMsg := cmd()
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	if !strings.Contains(view, "Complete project") {
		t.Error("journal should show today's task")
	}
	if !strings.Contains(view, "Meeting notes") {
		t.Error("journal should show today's note")
	}
}

func TestUAT_JournalView_Navigation_UpDown(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create multiple entries
	opts := service.LogEntriesOptions{Date: time.Now()}
	if _, err := bujoSvc.LogEntries(ctx, ". Task 1", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}
	if _, err := bujoSvc.LogEntries(ctx, ". Task 2", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}
	if _, err := bujoSvc.LogEntries(ctx, ". Task 3", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	cmd := model.Init()
	loadMsg := cmd()
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	initialIdx := model.selectedIdx

	// Move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.selectedIdx != initialIdx+1 {
		t.Errorf("'j' should move selection down, expected %d got %d", initialIdx+1, model.selectedIdx)
	}

	// Move up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.selectedIdx != initialIdx {
		t.Errorf("'k' should move selection up, expected %d got %d", initialIdx, model.selectedIdx)
	}
}

func TestUAT_JournalView_TimeNavigation_H_GoesToPreviousPeriod(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load journal view
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		newModel, cmd := model.Update(msg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// In day mode, record current date
	initialDate := model.viewDate

	// Press 'h' to go to previous day
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	// Execute the command to load the new agenda
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// viewDate should be one day before
	expectedDate := initialDate.AddDate(0, 0, -1)
	if !model.viewDate.Equal(expectedDate) {
		t.Errorf("after pressing 'h', viewDate should be %v, got %v", expectedDate, model.viewDate)
	}
}

func TestUAT_JournalView_TimeNavigation_L_GoesToNextPeriod(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Start with yesterday's date so we can go forward
	model.viewDate = time.Now().AddDate(0, 0, -1)

	// Load journal view
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		newModel, cmd := model.Update(msg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	initialDate := model.viewDate

	// Press 'l' to go to next day
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	// Execute the command to load the new agenda
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// viewDate should be one day after
	expectedDate := initialDate.AddDate(0, 0, 1)
	if !model.viewDate.Equal(expectedDate) {
		t.Errorf("after pressing 'l', viewDate should be %v, got %v", expectedDate, model.viewDate)
	}
}

func TestUAT_JournalView_NavigateToFutureDates(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create an entry for tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1)
	opts := service.LogEntriesOptions{Date: tomorrow}
	entries, err := bujoSvc.LogEntries(ctx, ". Future task", opts)
	if err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no entries created")
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Start at today
	now := time.Now()
	model.viewDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Load journal view
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		newModel, cmd := model.Update(msg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Press 'l' to go to tomorrow
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	// Execute the command to load tomorrow's agenda
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// viewDate should be tomorrow
	tomorrowDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
	if !model.viewDate.Equal(tomorrowDate) {
		t.Errorf("after pressing 'l' from today, viewDate should be %v, got %v", tomorrowDate, model.viewDate)
	}

	// The future task should be shown in the view
	view := model.View()
	if !strings.Contains(view, "Future task") {
		t.Error("future task should be visible when navigating to tomorrow")
	}
}

func TestUAT_JournalView_PastDays_NoOverdueSection(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a task for a past date that would normally be overdue
	pastDate := time.Now().AddDate(0, 0, -3)
	opts := service.LogEntriesOptions{Date: pastDate}
	_, err := bujoSvc.LogEntries(ctx, ". Old task", opts)
	if err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Set view to a past date (2 days ago)
	model.viewDate = time.Now().AddDate(0, 0, -2)

	// Load journal view
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		newModel, cmd := model.Update(msg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	view := model.View()

	// When viewing past dates, should NOT show overdue section
	if strings.Contains(view, "OVERDUE") {
		t.Error("viewing past dates should not show OVERDUE section")
	}
}

func TestUAT_JournalView_PastDays_ShowsAISummaryPrompt(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Set view to a past date
	model.viewDate = time.Now().AddDate(0, 0, -2)

	// Load journal view
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		newModel, cmd := model.Update(msg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	view := model.View()

	// When viewing past dates, should show AI summary section
	if !strings.Contains(view, "🤖 AI") || !strings.Contains(view, "Summary") {
		t.Error("viewing past dates should show AI summary section")
	}
}

func TestUAT_JournalView_MarkDone(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	opts := service.LogEntriesOptions{Date: time.Now()}
	if _, err := bujoSvc.LogEntries(ctx, ". Task to complete", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	cmd := model.Init()
	loadMsg := cmd()
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	// Press space or x to mark done
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	if cmd == nil {
		t.Fatal("marking done should return a command")
	}

	// Execute the command
	doneMsg := cmd()
	newModel, cmd = model.Update(doneMsg)
	model = newModel.(Model)

	// Reload to see updated state
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Check entry is now done
	if len(model.entries) == 0 {
		t.Fatal("should have entries after reload")
	}

	found := false
	for _, e := range model.entries {
		if e.Entry.Content == "Task to complete" && e.Entry.Type == domain.EntryTypeDone {
			found = true
			break
		}
	}

	if !found {
		t.Error("task should be marked as done")
	}
}

func TestUAT_JournalView_CaptureMode_LaunchesExternalEditor(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal

	// Press 'c' to launch external editor
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	_, cmd := model.Update(msg)

	// Should return a command to launch external editor
	if cmd == nil {
		t.Error("'c' in journal view should return a command to launch external editor")
	}
}

func TestUAT_JournalView_CaptureMode_NotInOtherViews(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Test that capture mode doesn't activate in habits view
	model.currentView = ViewTypeHabits
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("capture should NOT be available in habits view")
	}

	// Test that capture mode doesn't activate in lists view
	model.currentView = ViewTypeLists
	_, cmd = model.Update(msg)

	if cmd != nil {
		t.Error("capture should NOT be available in lists view")
	}
}

func TestUAT_JournalView_Edit(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	opts := service.LogEntriesOptions{Date: time.Now()}
	if _, err := bujoSvc.LogEntries(ctx, ". Original content", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	cmd := model.Init()
	loadMsg := cmd()
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	// Press 'e' to edit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ = model.Update(msg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Error("'e' should activate edit mode")
	}
}

func TestUAT_JournalView_Delete(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	opts := service.LogEntriesOptions{Date: time.Now()}
	if _, err := bujoSvc.LogEntries(ctx, ". Task to delete", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	cmd := model.Init()
	loadMsg := cmd()
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	// Press 'd' to delete
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ = model.Update(msg)
	m := newModel.(Model)

	if !m.confirmMode.active {
		t.Error("'d' should show confirmation dialog")
	}
}

// =============================================================================
// UAT Section 8: Habits View
// =============================================================================

func TestUAT_HabitsView_ShowsHabitDetails(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit with a streak
	for i := 0; i < 5; i++ {
		date := time.Now().AddDate(0, 0, -i)
		if err := habitSvc.LogHabitForDate(ctx, "Gym", 1, date); err != nil {
			t.Fatalf("failed to log habit: %v", err)
		}
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	// Should show habit name
	if !strings.Contains(view, "Gym") {
		t.Error("habits view should show habit name")
	}

	// Should show streak
	if !strings.Contains(view, "5") {
		t.Error("habits view should show streak count")
	}
}

func TestUAT_HabitsView_DeletedHabitsNotShown(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create and log habits
	if err := habitSvc.LogHabit(ctx, "Active Habit", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}
	if err := habitSvc.LogHabit(ctx, "Deleted Habit", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	// Delete one habit
	if err := habitSvc.DeleteHabit(ctx, "Deleted Habit"); err != nil {
		t.Fatalf("failed to delete habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	if strings.Contains(view, "Deleted Habit") {
		t.Error("deleted habits should NOT appear in view")
	}
	if !strings.Contains(view, "Active Habit") {
		t.Error("active habits should appear in view")
	}
}

func TestUAT_HabitsView_LogHabitFromTUI(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	if err := habitSvc.LogHabit(ctx, "Water", 1); err != nil {
		t.Fatalf("failed to create habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view and load (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	initialCount := 0
	if len(model.habitState.habits) > 0 {
		initialCount = model.habitState.habits[0].TodayCount
	}

	// Press Space to log habit
	msg = tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	if cmd == nil {
		t.Fatal("logging habit should return a command")
	}

	logMsg := cmd()
	newModel, cmd = model.Update(logMsg)
	model = newModel.(Model)

	// Process reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	if len(model.habitState.habits) == 0 {
		t.Fatal("should have habits after logging")
	}

	newCount := model.habitState.habits[0].TodayCount
	if newCount != initialCount+1 {
		t.Errorf("today's count should increase from %d to %d, got %d", initialCount, initialCount+1, newCount)
	}
}

func TestUAT_HabitsView_MonthlyHistoryShown(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create habit with history over multiple days
	for i := 0; i < 10; i++ {
		date := time.Now().AddDate(0, 0, -i)
		if err := habitSvc.LogHabitForDate(ctx, "Meditation", 1, date); err != nil {
			t.Fatalf("failed to log habit: %v", err)
		}
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.habitState.viewMode = HabitViewModeMonth // Monthly view by default per UAT

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Check that we have day history
	if len(model.habitState.habits) == 0 {
		t.Fatal("should have habits")
	}

	habit := model.habitState.habits[0]
	if len(habit.DayHistory) < 10 {
		t.Errorf("monthly view should show at least 10 days of history, got %d", len(habit.DayHistory))
	}
}

func TestUAT_HabitsView_Navigation(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	if err := habitSvc.LogHabit(ctx, "Habit 1", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}
	if err := habitSvc.LogHabit(ctx, "Habit 2", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	initialIdx := model.habitState.selectedIdx

	// Move down
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.habitState.selectedIdx != initialIdx+1 {
		t.Error("'j' should move selection down in habits view")
	}
}

func TestUAT_HabitsView_ContextAppropriateCommands(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeHabits

	view := model.View()

	// Should NOT show capture command
	if strings.Contains(view, "[c]apture") || strings.Contains(view, "c:capture") {
		t.Error("habits view should NOT show capture command")
	}
}

// =============================================================================
// UAT Section 9: Lists View
// =============================================================================

func TestUAT_ListsView_ShowsAllLists(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	if _, err := listSvc.CreateList(ctx, "Shopping"); err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.CreateList(ctx, "Work Tasks"); err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	if !strings.Contains(view, "Shopping") {
		t.Error("lists view should show Shopping list")
	}
	if !strings.Contains(view, "Work Tasks") {
		t.Error("lists view should show Work Tasks list")
	}
}

func TestUAT_ListsView_ShowsAccurateCompletionCounts(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	// Add 5 items
	for i := 0; i < 5; i++ {
		if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Item"); err != nil {
			t.Fatalf("failed to add item: %v", err)
		}
	}

	// Mark 2 as done
	items, err := listSvc.GetListItems(ctx, list.ID)
	if err != nil {
		t.Fatalf("failed to get items: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := listSvc.MarkDone(ctx, items[i].RowID); err != nil {
			t.Fatalf("failed to mark done: %v", err)
		}
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	// Should show "2/5" somewhere
	if !strings.Contains(view, "2/5") && !strings.Contains(view, "2 / 5") {
		t.Error("lists view should show accurate completion count (2/5)")
	}
}

func TestUAT_ListsView_DeletedListsNotShown(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	if _, err := listSvc.CreateList(ctx, "Active List"); err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	list2, err := listSvc.CreateList(ctx, "Deleted List")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	if err := listSvc.DeleteList(ctx, list2.ID, false); err != nil {
		t.Fatalf("failed to delete list: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	if strings.Contains(view, "Deleted List") {
		t.Error("deleted lists should NOT appear in view")
	}
	if !strings.Contains(view, "Active List") {
		t.Error("active lists should appear in view")
	}
}

func TestUAT_ListsView_EnterOpensItems(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Go to lists view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Press Enter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = model.Update(enterMsg)
	model = newModel.(Model)

	if model.currentView != ViewTypeListItems {
		t.Errorf("Enter should open list items view, got %v", model.currentView)
	}
}

func TestUAT_ListsView_CreateListWithAddKey(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to lists view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'a' to create a new list
	addMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ = model.Update(addMsg)
	model = newModel.(Model)

	if !model.createListMode.active {
		t.Fatal("pressing 'a' in lists view should activate create list mode")
	}

	// Type list name
	for _, r := range "My New List" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = model.Update(charMsg)
		model = newModel.(Model)
	}

	// Press Enter to create the list
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)

	if cmd == nil {
		t.Fatal("submitting list name should return a command to create the list")
	}

	// Execute the create command
	createMsg := cmd()
	newModel, cmd = model.Update(createMsg)
	model = newModel.(Model)

	// Reload lists
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify list was created and is shown
	view := model.View()
	if !strings.Contains(view, "My New List") {
		t.Error("newly created list should appear in the lists view")
	}

	if model.createListMode.active {
		t.Error("create list mode should be deactivated after submission")
	}
}

func TestUAT_JournalView_MoveEntryToList(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a journal entry for today
	opts := service.LogEntriesOptions{Date: time.Now()}
	entries, err := bujoSvc.LogEntries(ctx, ". Task to move", opts)
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no entries created")
	}

	// Create a list to move to
	list, err := listSvc.CreateList(ctx, "My List")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Load journal view
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		newModel, _ := model.Update(msg)
		model = newModel.(Model)
	}

	// Press 'L' to move entry to list
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	// Execute the load lists command
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	if !model.moveToListMode.active {
		t.Fatal("pressing 'L' should activate move to list mode")
	}

	if len(model.moveToListMode.targetLists) == 0 {
		t.Fatal("move to list mode should have target lists")
	}

	// Press Enter to select the first list
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)

	if cmd == nil {
		t.Fatal("selecting a list should return a command to move the entry")
	}

	// Execute the move command
	moveMsg := cmd()
	newModel, _ = model.Update(moveMsg)
	model = newModel.(Model)

	// Verify the entry was moved to the list
	items, err := listSvc.GetListItems(ctx, list.ID)
	if err != nil {
		t.Fatalf("failed to get list items: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item in list, got %d", len(items))
	}

	if items[0].Content != "Task to move" {
		t.Errorf("expected content 'Task to move', got '%s'", items[0].Content)
	}
}

// =============================================================================
// UAT Section 10: List Items View
// =============================================================================

func TestUAT_ListItemsView_ShowsAllItems(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy bread"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	view := model.View()

	if !strings.Contains(view, "Buy milk") {
		t.Error("list items view should show 'Buy milk'")
	}
	if !strings.Contains(view, "Buy bread") {
		t.Error("list items view should show 'Buy bread'")
	}
}

func TestUAT_ListItemsView_ToggleDone(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Toggle done with space
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(spaceMsg)
	model = newModel.(Model)

	if cmd == nil {
		t.Fatal("toggling done should return a command")
	}

	toggleMsg := cmd()
	newModel, cmd = model.Update(toggleMsg)
	model = newModel.(Model)

	// Reload items
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Check item is now done
	found := false
	for _, item := range model.listState.items {
		if item.Content == "Buy milk" && item.Type == domain.ListItemTypeDone {
			found = true
			break
		}
	}

	if !found {
		t.Error("item should be marked as done after toggle")
	}
}

func TestUAT_ListItemsView_AddItem(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	_, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'a' to add
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ = model.Update(aMsg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Error("'a' should activate add mode in list items view")
	}
}

func TestUAT_ListItemsView_DeleteItem(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'd' to delete
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ = model.Update(dMsg)
	m := newModel.(Model)

	if !m.confirmMode.active {
		t.Error("'d' should show confirmation dialog for delete")
	}
}

func TestUAT_ListItemsView_EscapeReturnsToLists(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	_, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = model.Update(enterMsg)
	model = newModel.(Model)

	// Press Escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ = model.Update(escMsg)
	model = newModel.(Model)

	if model.currentView != ViewTypeLists {
		t.Errorf("Escape should return to lists view, got %v", model.currentView)
	}
}

func TestUAT_ListItemsView_EditItem(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	itemID, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk")
	if err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'e' to edit
	eMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ = model.Update(eMsg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Error("'e' should activate edit mode in list items view")
	}

	if m.editMode.entryID != itemID {
		t.Errorf("editMode.entryID should be %d, got %d", itemID, m.editMode.entryID)
	}

	if m.editMode.input.Value() != "Buy milk" {
		t.Errorf("editMode.input should contain 'Buy milk', got '%s'", m.editMode.input.Value())
	}
}

func TestUAT_ListItemsView_EditItem_PersistsChange(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'e' to edit
	eMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ = model.Update(eMsg)
	model = newModel.(Model)

	// Type new content
	model.editMode.input.SetValue("Buy oat milk")

	// Press Enter to confirm
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		editMsg := cmd()
		newModel, cmd = model.Update(editMsg)
		model = newModel.(Model)
		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify the item was updated
	items, err := listSvc.GetListItems(ctx, list.ID)
	if err != nil {
		t.Fatalf("failed to get list items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Content != "Buy oat milk" {
		t.Errorf("item content should be 'Buy oat milk', got '%s'", items[0].Content)
	}
}

func TestUAT_ListItemsView_MoveItem(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list1, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	list2, err := listSvc.CreateList(ctx, "Work")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list1.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items (Shopping list)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'M' to move item
	shiftMMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}}
	newModel, _ = model.Update(shiftMMsg)
	model = newModel.(Model)

	if !model.moveListItemMode.active {
		t.Error("'M' should activate move list item mode")
	}

	// Press '1' to select Work list (first in target list since Shopping is filtered out)
	oneMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	newModel, cmd = model.Update(oneMsg)
	model = newModel.(Model)

	// Process the move command
	if cmd != nil {
		moveMsg := cmd()
		newModel, cmd = model.Update(moveMsg)
		model = newModel.(Model)
		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify item was moved
	items1, _ := listSvc.GetListItems(ctx, list1.ID)
	items2, _ := listSvc.GetListItems(ctx, list2.ID)

	if len(items1) != 0 {
		t.Errorf("Shopping list should be empty, has %d items", len(items1))
	}
	if len(items2) != 1 {
		t.Errorf("Work list should have 1 item, has %d items", len(items2))
	}
	if len(items2) == 1 && items2[0].Content != "Buy milk" {
		t.Errorf("Work list item should be 'Buy milk', got '%s'", items2[0].Content)
	}
}

// =============================================================================
// UAT Section 11: Search View
// =============================================================================

func TestUAT_SearchView_Accessible(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'8'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeSearch {
		t.Errorf("'8' should switch to search view, got %v", m.currentView)
	}
}

func TestUAT_SearchView_ShowsSearchInput(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeSearch

	view := model.View()

	// Should show search prompt or input field
	if !strings.Contains(strings.ToLower(view), "search") {
		t.Error("search view should show search indicator")
	}
}

// =============================================================================
// UAT Section 12: Summary/Stats View
// =============================================================================

func TestUAT_StatsView_Accessible(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeStats {
		t.Errorf("'9' should switch to stats view, got %v", m.currentView)
	}
}

func TestUAT_StatsView_ShowsProductivityInfo(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create some data
	opts := service.LogEntriesOptions{Date: time.Now()}
	if _, err := bujoSvc.LogEntries(ctx, ". Task 1", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}
	if err := habitSvc.LogHabit(ctx, "Gym", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeStats

	view := model.View()

	// Should show some stats-related content
	viewLower := strings.ToLower(view)
	hasStatsContent := strings.Contains(viewLower, "stat") ||
		strings.Contains(viewLower, "summary") ||
		strings.Contains(viewLower, "task") ||
		strings.Contains(viewLower, "habit")

	if !hasStatsContent {
		t.Error("stats view should show statistics or summary information")
	}
}

// =============================================================================
// UAT Section 13: Settings View
// =============================================================================

func TestUAT_SettingsView_Accessible(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeSettings {
		t.Errorf("'0' should switch to settings view, got %v", m.currentView)
	}
}

func TestUAT_SettingsView_ShowsCurrentSettings(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeSettings

	view := model.View()

	viewLower := strings.ToLower(view)
	hasSettingsContent := strings.Contains(viewLower, "theme") ||
		strings.Contains(viewLower, "setting") ||
		strings.Contains(viewLower, "default")

	if !hasSettingsContent {
		t.Error("settings view should show configuration options")
	}
}

// =============================================================================
// UAT Section 14: Error Handling
// =============================================================================

func TestUAT_ErrorHandling_EmptyStates(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Test empty journal
	model.currentView = ViewTypeJournal
	model.entries = nil
	view := model.View()
	if view == "" {
		t.Error("empty journal should still render something")
	}

	// Test empty habits
	model.currentView = ViewTypeHabits
	model.habitState.habits = nil
	view = model.View()
	if view == "" {
		t.Error("empty habits view should still render something")
	}

	// Test empty lists
	model.currentView = ViewTypeLists
	model.listState.lists = nil
	view = model.View()
	if view == "" {
		t.Error("empty lists view should still render something")
	}
}

// =============================================================================
// UAT Section 15: Data Accuracy
// =============================================================================

func TestUAT_DataAccuracy_DeletedItemsNeverAppear(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create and delete entries
	opts := service.LogEntriesOptions{Date: time.Now()}
	entries, err := bujoSvc.LogEntries(ctx, ". Task to delete", opts)
	if err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}
	if err := bujoSvc.DeleteEntry(ctx, entries[0]); err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}

	// Create and delete habit
	if err := habitSvc.LogHabit(ctx, "Deleted Habit", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}
	if err := habitSvc.DeleteHabit(ctx, "Deleted Habit"); err != nil {
		t.Fatalf("failed to delete habit: %v", err)
	}

	// Create and delete list
	list, err := listSvc.CreateList(ctx, "Deleted List")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if err := listSvc.DeleteList(ctx, list.ID, false); err != nil {
		t.Fatalf("failed to delete list: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Check journal
	cmd := model.Init()
	loadMsg := cmd()
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()
	if strings.Contains(view, "Task to delete") {
		t.Error("deleted entry should not appear in journal")
	}

	// Check habits (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)
	loadMsg = cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view = model.View()
	if strings.Contains(view, "Deleted Habit") {
		t.Error("deleted habit should not appear in habits view")
	}

	// Check lists (key 6)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)
	loadMsg = cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view = model.View()
	if strings.Contains(view, "Deleted List") {
		t.Error("deleted list should not appear in lists view")
	}
}

func TestUAT_DataAccuracy_CountsAreAccurate(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Test List")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	// Add 3 items
	for i := 0; i < 3; i++ {
		if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Item"); err != nil {
			t.Fatalf("failed to add item: %v", err)
		}
	}

	// Mark 1 done
	items, err := listSvc.GetListItems(ctx, list.ID)
	if err != nil {
		t.Fatalf("failed to get items: %v", err)
	}
	if err := listSvc.MarkDone(ctx, items[0].RowID); err != nil {
		t.Fatalf("failed to mark done: %v", err)
	}

	// Delete 1 item
	if err := listSvc.RemoveItem(ctx, items[1].RowID); err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	// Should show 1/2 (1 done out of 2 remaining active items)
	if !strings.Contains(view, "1/2") && !strings.Contains(view, "1 / 2") {
		t.Errorf("list should show accurate count 1/2, view: %s", view)
	}
}

func TestUAT_DataAccuracy_ChangesAppearImmediately(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to list items
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg = cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Toggle done
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(spaceMsg)
	model = newModel.(Model)

	if cmd != nil {
		toggleMsg := cmd()
		newModel, cmd = model.Update(toggleMsg)
		model = newModel.(Model)

		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify item shows as done immediately in state
	found := false
	for _, item := range model.listState.items {
		if item.Content == "Buy milk" && item.Type == domain.ListItemTypeDone {
			found = true
			break
		}
	}

	if !found {
		t.Error("changes should appear in state immediately after toggle")
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

// =============================================================================
// UAT: Cancel/Strikethrough Entries (#77)
// =============================================================================

func TestUAT_JournalView_CancelEntry(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	_, err := bujoSvc.LogEntries(ctx, ". Task to cancel", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	model.agenda = agenda
	model.entries = model.flattenAgenda(agenda)
	model.selectedIdx = 0

	// Press 'x' to cancel the entry
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should return a command to cancel the entry
	if cmd == nil {
		t.Error("pressing 'x' should return a cancel command")
	}
	_ = m
}

func TestUAT_JournalView_CancelledEntryShowsStrikethrough(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	ids, err := bujoSvc.LogEntries(ctx, ". Task to cancel", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Cancel the entry via service
	err = bujoSvc.CancelEntry(ctx, ids[0])
	if err != nil {
		t.Fatalf("failed to cancel entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	model.agenda = agenda
	model.entries = model.flattenAgenda(agenda)

	view := model.View()

	// Cancelled entries should show the cancelled symbol (✗)
	if !strings.Contains(view, "✗") {
		t.Error("cancelled entry should show ✗ symbol")
	}
}

func TestUAT_JournalView_UncancelEntry(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	ids, err := bujoSvc.LogEntries(ctx, ". Task to cancel", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Cancel the entry first
	err = bujoSvc.CancelEntry(ctx, ids[0])
	if err != nil {
		t.Fatalf("failed to cancel entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	model.agenda = agenda
	model.entries = model.flattenAgenda(agenda)
	model.selectedIdx = 0

	// Press 'X' (shift+x) to uncancel the entry
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should return a command to uncancel the entry
	if cmd == nil {
		t.Error("pressing 'X' should return an uncancel command")
	}
	_ = m
}

// =============================================================================
// UAT: Change Entry Type (#78)
// =============================================================================

func TestUAT_JournalView_RetypeEntry_OpensTypePicker(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	_, err := bujoSvc.LogEntries(ctx, ". Task to retype", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	model.agenda = agenda
	model.entries = model.flattenAgenda(agenda)
	model.selectedIdx = 0

	// Press 't' to open type picker
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should activate retype mode
	if !m.retypeMode.active {
		t.Error("pressing 't' should activate retype mode")
	}
}

func TestUAT_JournalView_RetypeEntry_ChangesType(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	ids, err := bujoSvc.LogEntries(ctx, ". Task to retype", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Retype via service
	err = bujoSvc.RetypeEntry(ctx, ids[0], domain.EntryTypeNote)
	if err != nil {
		t.Fatalf("failed to retype entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	model.agenda = agenda
	model.entries = model.flattenAgenda(agenda)

	// Entry should now show as note (–)
	if len(model.entries) == 0 {
		t.Fatal("expected entries")
	}

	if model.entries[0].Entry.Type != domain.EntryTypeNote {
		t.Errorf("entry type should be note, got %v", model.entries[0].Entry.Type)
	}
}

func TestUAT_JournalView_RetypeEntry_PreservesContent(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	ids, err := bujoSvc.LogEntries(ctx, ". Original content here", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Retype via service
	err = bujoSvc.RetypeEntry(ctx, ids[0], domain.EntryTypeEvent)
	if err != nil {
		t.Fatalf("failed to retype entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	model.agenda = agenda
	model.entries = model.flattenAgenda(agenda)

	if len(model.entries) == 0 {
		t.Fatal("expected entries")
	}

	if model.entries[0].Entry.Content != "Original content here" {
		t.Errorf("content should be preserved, got %q", model.entries[0].Entry.Content)
	}
}

// =============================================================================
// UAT: View and Restore Deleted Entries (#79)
// =============================================================================

func TestUAT_DeletedEntries_CanBeViewed(t *testing.T) {
	bujoSvc, _, _, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	// Create and delete an entry
	ids, err := bujoSvc.LogEntries(ctx, ". Entry to delete", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	err = bujoSvc.DeleteEntry(ctx, ids[0])
	if err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}

	// Deleted entries should be viewable
	deleted, err := bujoSvc.GetDeletedEntries(ctx)
	if err != nil {
		t.Fatalf("failed to get deleted entries: %v", err)
	}

	if len(deleted) == 0 {
		t.Error("deleted entry should be viewable via GetDeletedEntries")
	}

	found := false
	for _, e := range deleted {
		if e.Content == "Entry to delete" {
			found = true
			break
		}
	}

	if !found {
		t.Error("deleted entry should appear in deleted entries list")
	}
}

func TestUAT_DeletedEntries_CanBeRestored(t *testing.T) {
	bujoSvc, _, _, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	// Create and delete an entry
	ids, err := bujoSvc.LogEntries(ctx, ". Entry to restore", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Get the entity ID before deleting
	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	var entityID domain.EntityID
	for _, day := range agenda.Days {
		for _, e := range day.Entries {
			if e.ID == ids[0] {
				entityID = e.EntityID
				break
			}
		}
	}

	err = bujoSvc.DeleteEntry(ctx, ids[0])
	if err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}

	// Entry should not appear in agenda
	agenda, _ = bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	for _, day := range agenda.Days {
		for _, e := range day.Entries {
			if e.Content == "Entry to restore" {
				t.Error("deleted entry should not appear in agenda")
			}
		}
	}

	// Restore the entry
	newID, err := bujoSvc.RestoreEntry(ctx, entityID)
	if err != nil {
		t.Fatalf("failed to restore entry: %v", err)
	}

	if newID == 0 {
		t.Error("restore should return new entry ID")
	}

	// Entry should reappear in agenda
	agenda, _ = bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	found := false
	for _, day := range agenda.Days {
		for _, e := range day.Entries {
			if e.Content == "Entry to restore" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("restored entry should appear in agenda")
	}
}

func TestUAT_DeletedEntries_NotShownInJournal(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	// Create two entries
	_, err := bujoSvc.LogEntries(ctx, ". Keep this entry", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	ids, err := bujoSvc.LogEntries(ctx, ". Delete this entry", service.LogEntriesOptions{Date: todayDate})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Delete one entry
	err = bujoSvc.DeleteEntry(ctx, ids[0])
	if err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	agenda, _ := bujoSvc.GetMultiDayAgenda(ctx, todayDate, todayDate)
	model.agenda = agenda
	model.entries = model.flattenAgenda(agenda)

	view := model.View()

	if strings.Contains(view, "Delete this entry") {
		t.Error("deleted entry should not appear in journal view")
	}

	if !strings.Contains(view, "Keep this entry") {
		t.Error("non-deleted entry should appear in journal view")
	}
}

// =============================================================================
// UAT Section: Habit Weekly/Monthly Goals
// =============================================================================

func TestUAT_HabitsView_ShowsWeeklyProgress(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()
	today := time.Now()

	// Create a habit with weekly goal
	if err := habitSvc.LogHabitForDate(ctx, "Workout", 1, today); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}
	if err := habitSvc.SetHabitWeeklyGoal(ctx, "Workout", 5); err != nil {
		t.Fatalf("failed to set weekly goal: %v", err)
	}

	// Log 2 more times this week (total 3)
	for i := 1; i <= 2; i++ {
		if err := habitSvc.LogHabitForDate(ctx, "Workout", 1, today.AddDate(0, 0, -i)); err != nil {
			t.Fatalf("failed to log habit: %v", err)
		}
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	// Should show habit name
	if !strings.Contains(view, "Workout") {
		t.Error("habits view should show habit name")
	}

	// Should show weekly progress
	if !strings.Contains(view, "Week:") {
		t.Error("habits view should show weekly progress for habits with weekly goals")
	}
}

func TestUAT_HabitsView_ShowsMonthlyProgress(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit with monthly goal
	if err := habitSvc.LogHabit(ctx, "Reading", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}
	if err := habitSvc.SetHabitMonthlyGoal(ctx, "Reading", 20); err != nil {
		t.Fatalf("failed to set monthly goal: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	// Should show habit name
	if !strings.Contains(view, "Reading") {
		t.Error("habits view should show habit name")
	}

	// Should show monthly progress
	if !strings.Contains(view, "Month:") {
		t.Error("habits view should show monthly progress for habits with monthly goals")
	}
}

func TestUAT_HabitsView_HidesProgressWhenNoGoalsSet(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit without weekly/monthly goals (only daily)
	if err := habitSvc.LogHabit(ctx, "Simple", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	// Should show habit name
	if !strings.Contains(view, "Simple") {
		t.Error("habits view should show habit name")
	}

	// Should NOT show weekly/monthly progress lines
	if strings.Contains(view, "Week:") || strings.Contains(view, "Month:") {
		t.Error("habits view should NOT show weekly/monthly progress for habits without those goals")
	}
}

func TestUAT_HabitsView_MatchesCLIStyle(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit and log it today
	if err := habitSvc.LogHabit(ctx, "CrossFit", 1); err != nil {
		t.Fatalf("failed to create habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	view := model.View()

	// Should show habit name with streak
	if !strings.Contains(view, "CrossFit") {
		t.Error("view should show habit name")
	}
	if !strings.Contains(view, "streak") {
		t.Error("view should show streak info")
	}

	// Should use circle characters like CLI (● for completed, ○ for empty)
	if !strings.Contains(view, "●") && !strings.Contains(view, "○") {
		t.Error("view should use circle characters (● and ○) like CLI")
	}

	// Should show day labels
	if !strings.Contains(view, "S") && !strings.Contains(view, "M") {
		t.Error("view should show day labels")
	}

	// Should show completion percentage
	if !strings.Contains(view, "%") {
		t.Error("view should show completion percentage")
	}
}

func TestUAT_HabitsView_DayNavigation(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit
	if err := habitSvc.LogHabit(ctx, "Exercise", 1); err != nil {
		t.Fatalf("failed to create habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Initially selected day should be 6 (rightmost = today) for 7-day view
	if model.habitState.selectedDayIdx != 6 {
		t.Fatalf("expected selectedDayIdx to be 6 (today), got %d", model.habitState.selectedDayIdx)
	}

	// Press 'h' to move left (to yesterday)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.habitState.selectedDayIdx != 5 {
		t.Errorf("expected selectedDayIdx to be 5 after 'h', got %d", model.habitState.selectedDayIdx)
	}

	// Press 'l' to move right (back to today)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.habitState.selectedDayIdx != 6 {
		t.Errorf("expected selectedDayIdx to be 6 after 'l', got %d", model.habitState.selectedDayIdx)
	}

	// Can't go past today
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.habitState.selectedDayIdx != 6 {
		t.Errorf("expected selectedDayIdx to stay at 6 (can't go past today), got %d", model.habitState.selectedDayIdx)
	}
}

func TestUAT_HabitsView_LogOnSelectedDay(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit
	if err := habitSvc.LogHabit(ctx, "Workout", 1); err != nil {
		t.Fatalf("failed to create habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Move to yesterday (index 5)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// Log habit on yesterday
	msg = tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	if cmd == nil {
		t.Fatal("logging habit should return a command")
	}

	logMsg := cmd()
	newModel, cmd = model.Update(logMsg)
	model = newModel.(Model)

	// Process reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Check that yesterday's status shows completed
	if len(model.habitState.habits) == 0 {
		t.Fatal("should have habits")
	}

	habit := model.habitState.habits[0]
	// DayHistory is ordered: [0]=today, [1]=yesterday, etc.
	if len(habit.DayHistory) < 2 {
		t.Fatalf("expected at least 2 days of history, got %d", len(habit.DayHistory))
	}

	// Yesterday is DayHistory[1] (0=today, 1=yesterday)
	if !habit.DayHistory[1].Completed {
		t.Error("yesterday should be marked as completed after logging")
	}
}

func TestUAT_HabitsView_DeleteHabitFromTUI(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit first
	if err := habitSvc.LogHabit(ctx, "To Delete", 1); err != nil {
		t.Fatalf("failed to create habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Verify habit exists
	if len(model.habitState.habits) != 1 {
		t.Fatalf("expected 1 habit, got %d", len(model.habitState.habits))
	}

	// Press 'd' to start delete
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// Should be in confirm delete mode
	if !model.confirmHabitDeleteMode.active {
		t.Fatal("pressing 'd' should activate confirm habit delete mode")
	}

	// Press 'y' to confirm
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Should have exited confirm mode
	if model.confirmHabitDeleteMode.active {
		t.Fatal("pressing 'y' should deactivate confirm mode")
	}

	// Execute the delete command
	if cmd == nil {
		t.Fatal("should return a command to delete the habit")
	}
	deleteMsg := cmd()
	newModel, cmd = model.Update(deleteMsg)
	model = newModel.(Model)

	// Execute reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify habit was deleted
	if len(model.habitState.habits) != 0 {
		t.Fatalf("expected 0 habits after deleting, got %d", len(model.habitState.habits))
	}
}

func TestUAT_HabitsView_AddHabitFromTUI(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Verify no habits initially
	if len(model.habitState.habits) != 0 {
		t.Fatalf("expected 0 habits initially, got %d", len(model.habitState.habits))
	}

	// Press 'a' to start adding a habit
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// Should be in add habit mode
	if !model.addHabitMode.active {
		t.Fatal("pressing 'a' should activate add habit mode")
	}

	// Type habit name
	for _, r := range "Morning Run" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = model.Update(charMsg)
		model = newModel.(Model)
	}

	// Press Enter to confirm
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Should have exited add mode
	if model.addHabitMode.active {
		t.Fatal("pressing Enter should deactivate add habit mode")
	}

	// Execute the add command
	if cmd == nil {
		t.Fatal("should return a command to add the habit")
	}
	addMsg := cmd()
	newModel, cmd = model.Update(addMsg)
	model = newModel.(Model)

	// Execute reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify habit was added
	if len(model.habitState.habits) != 1 {
		t.Fatalf("expected 1 habit after adding, got %d", len(model.habitState.habits))
	}

	if model.habitState.habits[0].Name != "Morning Run" {
		t.Errorf("expected habit name 'Morning Run', got '%s'", model.habitState.habits[0].Name)
	}
}

// =============================================================================
// UAT Section 16: Goals View
// =============================================================================

func TestUAT_GoalsView_Accessible(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeGoals {
		t.Errorf("'7' should switch to goals view, got %v", m.currentView)
	}
}

func TestUAT_GoalsView_ShowsEmptyState(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeGoals

	view := model.View()

	if !strings.Contains(view, "No goals") {
		t.Error("goals view should show empty state message when no goals exist")
	}
}

func TestUAT_GoalsView_ShowsGoalContent(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a test goal
	currentMonth := time.Now()
	_, err := goalSvc.CreateGoal(ctx, "Learn Go programming", currentMonth)
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	view := model.View()

	if !strings.Contains(view, "Learn Go programming") {
		t.Error("goals view should display goal content")
	}
}

func TestUAT_GoalsView_ShowsMonthlyGoalsHeader(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeGoals

	view := model.View()

	if !strings.Contains(view, "Monthly Goals") {
		t.Error("goals view should show 'Monthly Goals' header")
	}
}

func TestUAT_GoalsView_NavigationUpDown(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create multiple goals
	currentMonth := time.Now()
	_, _ = goalSvc.CreateGoal(ctx, "Goal One", currentMonth)
	_, _ = goalSvc.CreateGoal(ctx, "Goal Two", currentMonth)
	_, _ = goalSvc.CreateGoal(ctx, "Goal Three", currentMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Initial selection should be 0
	if model.goalState.selectedIdx != 0 {
		t.Errorf("initial selection should be 0, got %d", model.goalState.selectedIdx)
	}

	// Press j to move down
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ = model.Update(jMsg)
	model = newModel.(Model)

	if model.goalState.selectedIdx != 1 {
		t.Errorf("after pressing j, selection should be 1, got %d", model.goalState.selectedIdx)
	}

	// Press k to move up
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = model.Update(kMsg)
	model = newModel.(Model)

	if model.goalState.selectedIdx != 0 {
		t.Errorf("after pressing k, selection should be 0, got %d", model.goalState.selectedIdx)
	}
}

func TestUAT_GoalsView_ToggleDoneWithSpace(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a goal
	currentMonth := time.Now()
	goalID, _ := goalSvc.CreateGoal(ctx, "Complete task", currentMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify goal is not done initially
	goal, _ := goalSvc.GetGoal(ctx, goalID)
	if goal.IsDone() {
		t.Error("goal should not be done initially")
	}

	// Press space to toggle
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(spaceMsg)
	model = newModel.(Model)

	// Execute the command chain
	if cmd != nil {
		toggleMsg := cmd()
		newModel, cmd = model.Update(toggleMsg)
		model = newModel.(Model)
		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify goal is now done
	goal, _ = goalSvc.GetGoal(ctx, goalID)
	if !goal.IsDone() {
		t.Error("goal should be marked as done after pressing space")
	}
}

func TestUAT_GoalsView_ShowsDoneStatus(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create and complete a goal
	currentMonth := time.Now()
	goalID, _ := goalSvc.CreateGoal(ctx, "Completed goal", currentMonth)
	_ = goalSvc.MarkDone(ctx, goalID)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	view := model.View()

	// Should show checkmark for done goals
	if !strings.Contains(view, "✓") {
		t.Error("goals view should show checkmark for completed goals")
	}
}

func TestUAT_GoalsView_ShowsGoalID(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a goal
	currentMonth := time.Now()
	_, _ = goalSvc.CreateGoal(ctx, "Test goal", currentMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	view := model.View()

	// Should show goal ID with # prefix
	if !strings.Contains(view, "#") {
		t.Error("goals view should show goal ID with # prefix")
	}
}

func TestUAT_GoalsView_ShowsHelpText(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeGoals

	view := model.View()

	// Should show help text for navigation
	hasHelp := strings.Contains(view, "space") || strings.Contains(view, "j/k")
	if !hasHelp {
		t.Error("goals view should show help text for navigation")
	}
}

func TestUAT_GoalsView_ToolbarShowsGoals(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeGoals

	view := model.View()

	// Toolbar should indicate Goals view
	if !strings.Contains(view, "Goals") {
		t.Error("toolbar should show 'Goals' when in goals view")
	}
}

func TestUAT_GoalsView_MonthNavigation_PreviousMonth(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create goals in current and previous month
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonth := currentMonth.AddDate(0, -1, 0)

	_, _ = goalSvc.CreateGoal(ctx, "Current month goal", currentMonth)
	_, _ = goalSvc.CreateGoal(ctx, "Previous month goal", prevMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify current month goal is shown
	view := model.View()
	if !strings.Contains(view, "Current month goal") {
		t.Error("should show current month goal initially")
	}

	// Press 'h' to go to previous month
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Execute load command
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Should now show previous month goal
	view = model.View()
	if !strings.Contains(view, "Previous month goal") {
		t.Error("pressing 'h' should navigate to previous month and show its goals")
	}
}

func TestUAT_GoalsView_MonthNavigation_NextMonth(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create goals in current and next month
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := currentMonth.AddDate(0, 1, 0)

	_, _ = goalSvc.CreateGoal(ctx, "Current month goal", currentMonth)
	_, _ = goalSvc.CreateGoal(ctx, "Next month goal", nextMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'l' to go to next month
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Execute load command
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Should now show next month goal
	view := model.View()
	if !strings.Contains(view, "Next month goal") {
		t.Error("pressing 'l' should navigate to next month and show its goals")
	}
}

func TestUAT_GoalsView_AddGoalFromTUI(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify no goals initially
	if len(model.goalState.goals) != 0 {
		t.Fatalf("expected 0 goals initially, got %d", len(model.goalState.goals))
	}

	// Press 'a' to start adding a goal
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// Should be in add goal mode
	if !model.addGoalMode.active {
		t.Fatal("pressing 'a' should activate add goal mode")
	}

	// Type goal content
	for _, r := range "Learn Go" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = model.Update(charMsg)
		model = newModel.(Model)
	}

	// Press Enter to confirm
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Should have exited add mode
	if model.addGoalMode.active {
		t.Fatal("pressing Enter should deactivate add goal mode")
	}

	// Execute the add command
	if cmd == nil {
		t.Fatal("should return a command to add the goal")
	}
	addMsg := cmd()
	newModel, cmd = model.Update(addMsg)
	model = newModel.(Model)

	// Execute reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify goal was added
	if len(model.goalState.goals) != 1 {
		t.Fatalf("expected 1 goal after adding, got %d", len(model.goalState.goals))
	}

	if model.goalState.goals[0].Content != "Learn Go" {
		t.Errorf("expected goal content 'Learn Go', got '%s'", model.goalState.goals[0].Content)
	}
}

func TestUAT_GoalsView_EditGoalFromTUI(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a goal first
	currentMonth := time.Now()
	_, _ = goalSvc.CreateGoal(ctx, "Original content", currentMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Press 'e' to start editing
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// Should be in edit goal mode
	if !model.editGoalMode.active {
		t.Fatal("pressing 'e' should activate edit goal mode")
	}

	// Clear input and type new content
	// First select all and delete (Ctrl+U clears line in most inputs)
	msg = tea.KeyMsg{Type: tea.KeyCtrlU}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	for _, r := range "Updated content" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = model.Update(charMsg)
		model = newModel.(Model)
	}

	// Press Enter to confirm
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Should have exited edit mode
	if model.editGoalMode.active {
		t.Fatal("pressing Enter should deactivate edit goal mode")
	}

	// Execute the edit command
	if cmd == nil {
		t.Fatal("should return a command to edit the goal")
	}
	editMsg := cmd()
	newModel, cmd = model.Update(editMsg)
	model = newModel.(Model)

	// Execute reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify goal was updated
	if model.goalState.goals[0].Content != "Updated content" {
		t.Errorf("expected goal content 'Updated content', got '%s'", model.goalState.goals[0].Content)
	}
}

func TestUAT_GoalsView_DeleteGoalFromTUI(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a goal first
	currentMonth := time.Now()
	_, _ = goalSvc.CreateGoal(ctx, "To be deleted", currentMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify goal exists
	if len(model.goalState.goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(model.goalState.goals))
	}

	// Press 'd' to delete
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// Should be in confirm delete mode
	if !model.confirmGoalDeleteMode.active {
		t.Fatal("pressing 'd' should activate confirm goal delete mode")
	}

	// Press 'y' to confirm
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Should have exited confirm mode
	if model.confirmGoalDeleteMode.active {
		t.Fatal("pressing 'y' should deactivate confirm mode")
	}

	// Execute the delete command
	if cmd == nil {
		t.Fatal("should return a command to delete the goal")
	}
	deleteMsg := cmd()
	newModel, cmd = model.Update(deleteMsg)
	model = newModel.(Model)

	// Execute reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify goal was deleted
	if len(model.goalState.goals) != 0 {
		t.Fatalf("expected 0 goals after deleting, got %d", len(model.goalState.goals))
	}
}

func TestUAT_GoalsView_MoveGoalToMonth(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a goal in current month
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	_, _ = goalSvc.CreateGoal(ctx, "Goal to move", currentMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify goal exists
	if len(model.goalState.goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(model.goalState.goals))
	}

	// Press '>' to move (migrate key)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// Should be in move goal mode
	if !model.moveGoalMode.active {
		t.Fatal("pressing '>' should activate move goal mode")
	}

	// Type target month (next month)
	nextMonth := currentMonth.AddDate(0, 1, 0)
	targetMonthStr := nextMonth.Format("2006-01")
	for _, r := range targetMonthStr {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = model.Update(charMsg)
		model = newModel.(Model)
	}

	// Press Enter to confirm
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Should have exited move mode
	if model.moveGoalMode.active {
		t.Fatal("pressing Enter should deactivate move goal mode")
	}

	// Execute the move command
	if cmd == nil {
		t.Fatal("should return a command to move the goal")
	}
	moveMsg := cmd()
	newModel, cmd = model.Update(moveMsg)
	model = newModel.(Model)

	// Execute reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Goal should no longer be in current month
	if len(model.goalState.goals) != 0 {
		t.Fatalf("expected 0 goals in current month after moving, got %d", len(model.goalState.goals))
	}

	// Navigate to next month to verify goal is there
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	if len(model.goalState.goals) != 1 {
		t.Fatalf("expected 1 goal in next month after moving, got %d", len(model.goalState.goals))
	}
}

func TestUAT_GoalsView_ShowsMonthInHeader(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view (key 7)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	view := model.View()

	// Should show current month name
	now := time.Now()
	monthName := now.Format("January 2006")
	if !strings.Contains(view, monthName) {
		t.Errorf("goals view should show current month name '%s'", monthName)
	}
}

func TestUAT_JournalView_ShowsGoalsSection(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a goal for current month
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	_, _ = goalSvc.CreateGoal(ctx, "Test journal goal", currentMonth)

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view (it's the default view)
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)

		// Process the goals load command
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	view := model.View()

	// Should show the goal in journal view
	if !strings.Contains(view, "Test journal goal") {
		t.Error("journal view should show current month goals")
	}

	// Should show the month name
	monthName := now.Format("January")
	if !strings.Contains(view, monthName+" Goals") {
		t.Errorf("journal view should show '%s Goals' header", monthName)
	}
}

func TestUAT_JournalView_MigrateTaskToGoal(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a task for today
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	opts := service.LogEntriesOptions{Date: today}
	_, err := bujoSvc.LogEntries(ctx, ". Task to migrate", opts)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Verify task exists in journal
	if len(model.entries) == 0 {
		t.Fatal("expected at least one entry in journal")
	}

	// Press 'M' to migrate to goal
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	// Should be in migrate to goal mode
	if !model.migrateToGoalMode.active {
		t.Fatal("pressing 'M' should activate migrate to goal mode")
	}

	// Type target month (next month)
	nextMonth := today.AddDate(0, 1, 0)
	targetMonthStr := nextMonth.Format("2006-01")
	for _, r := range targetMonthStr {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = model.Update(charMsg)
		model = newModel.(Model)
	}

	// Press Enter to confirm
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	// Should have exited migrate mode
	if model.migrateToGoalMode.active {
		t.Fatal("pressing Enter should deactivate migrate to goal mode")
	}

	// Execute the migrate command
	if cmd == nil {
		t.Fatal("should return a command to migrate the task")
	}
	migrateMsg := cmd()
	newModel, cmd = model.Update(migrateMsg)
	model = newModel.(Model)

	// Process any reload commands
	for cmd != nil {
		reloadMsg := cmd()
		newModel, cmd = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify task was removed from journal (soft deleted via event sourcing)
	foundTask := false
	for _, entry := range model.entries {
		if entry.Entry.Content == "Task to migrate" {
			foundTask = true
			break
		}
	}
	if foundTask {
		t.Error("task should be removed from journal after migration")
	}

	// Verify goal was created in the target month
	targetMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, now.Location())
	goals, err := goalSvc.GetGoalsForMonth(ctx, targetMonth)
	if err != nil {
		t.Fatalf("failed to get goals: %v", err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 goal in target month, got %d", len(goals))
	}

	if goals[0].Content != "Task to migrate" {
		t.Errorf("expected goal content 'Task to migrate', got '%s'", goals[0].Content)
	}
}

func TestUAT_JournalView_MigrateToGoal_OnlyWorksOnIncompleteTasks(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a completed task
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	opts := service.LogEntriesOptions{Date: today}
	ids, err := bujoSvc.LogEntries(ctx, ". Completed task", opts)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Mark it as done
	err = bujoSvc.MarkDone(ctx, ids[0])
	if err != nil {
		t.Fatalf("failed to mark task done: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Press 'M' to try to migrate - should NOT activate migrate mode for done tasks
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	// Should NOT be in migrate to goal mode for completed tasks
	if model.migrateToGoalMode.active {
		t.Error("migrate to goal should not activate for completed tasks")
	}
}

// =============================================================================
// UAT Section 14: Collapse/Expand Parent Entries
// =============================================================================

func TestUAT_Collapse_EnterTogglesCollapseState(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a parent entry with child
	opts := service.LogEntriesOptions{Date: time.Now()}
	ids, err := bujoSvc.LogEntries(ctx, ". Parent task\n  . Child task", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}
	if len(ids) < 2 {
		t.Fatalf("expected at least 2 entries created, got %d", len(ids))
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Verify initial state - parent should be collapsed by default (only 1 visible)
	if len(model.entries) != 1 {
		t.Fatalf("expected 1 visible entry (collapsed parent), got %d", len(model.entries))
	}

	// Select parent entry
	model.selectedIdx = 0
	parentEntry := model.entries[0].Entry

	// Press Enter to expand
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	// After expanding, should see 2 entries (parent + child)
	if len(model.entries) != 2 {
		t.Errorf("after expanding, expected 2 entries, got %d", len(model.entries))
	}

	// Parent should now be expanded (not collapsed)
	if model.collapsed[parentEntry.EntityID] {
		t.Error("after pressing Enter, parent should be expanded (not collapsed)")
	}

	// Press Enter again to collapse
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	// After collapsing, should see only 1 entry
	if len(model.entries) != 1 {
		t.Errorf("after collapsing, expected 1 entry, got %d", len(model.entries))
	}

	// Parent should now be collapsed
	if !model.collapsed[parentEntry.EntityID] {
		t.Error("after pressing Enter again, parent should be collapsed")
	}
}

func TestUAT_Collapse_DefaultCollapsedState(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a parent entry with children
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Parent task\n  . Child 1\n  . Child 2", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// All parents should start collapsed by default
	// So we should only see 1 entry (the parent), not all 3
	if len(model.entries) != 1 {
		t.Errorf("expected 1 visible entry (collapsed parent), got %d", len(model.entries))
	}
}

func TestUAT_Collapse_ShowsHiddenCount(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a parent entry with 3 children
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Parent task\n  . Child 1\n  . Child 2\n  . Child 3", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	view := model.View()

	// Should show collapse indicator with hidden count
	if !strings.Contains(view, "▶") {
		t.Error("collapsed parent should show ▶ indicator")
	}
	if !strings.Contains(view, "[3 hidden]") {
		t.Error("collapsed parent should show hidden count [3 hidden]")
	}
}

func TestUAT_Collapse_ExpandedShowsDownArrow(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a parent entry with child
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Parent task\n  . Child task", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Press Enter to expand
	model.selectedIdx = 0
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	view := model.View()

	// Should show expanded indicator
	if !strings.Contains(view, "▼") {
		t.Error("expanded parent should show ▼ indicator")
	}
}

func TestUAT_Collapse_LeafEntryNoIndicator(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a standalone entry with no children
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Single task", opts)
	if err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	view := model.View()

	// Leaf entries should NOT show collapse indicators
	if strings.Contains(view, "▶") || strings.Contains(view, "▼") {
		t.Error("leaf entry (no children) should not show collapse indicator")
	}
}

// =============================================================================
// UAT Section 14a: Expand/Collapse All Siblings
// =============================================================================

func TestUAT_Collapse_CtrlE_ExpandsAllSiblings(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create two parent entries, each with children
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Parent 1\n  . Child 1a\n  . Child 1b\n. Parent 2\n  . Child 2a", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Both parents start collapsed, so we see 2 entries
	if len(model.entries) != 2 {
		t.Fatalf("expected 2 visible entries (both parents collapsed), got %d", len(model.entries))
	}

	// Press Ctrl+E to expand all siblings (both parent entries at root level)
	msg := tea.KeyMsg{Type: tea.KeyCtrlE}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	// After Ctrl+E, all siblings should be expanded - we should see all 5 entries
	if len(model.entries) != 5 {
		t.Errorf("after Ctrl+E, expected 5 entries (all expanded), got %d", len(model.entries))
	}
}

func TestUAT_Collapse_SelectedItemAndAncestorsStayExpanded(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a parent with a child that has a grandchild
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Parent\n  . Child\n    . Grandchild", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Manually expand all entries from agenda to see the full tree
	for _, day := range model.agenda.Days {
		for _, entry := range day.Entries {
			model.collapsed[entry.EntityID] = false
		}
	}
	model.entries = model.flattenAgenda(model.agenda)

	// Should now see all 3 entries
	if len(model.entries) != 3 {
		t.Fatalf("expected 3 entries after expanding all, got %d", len(model.entries))
	}

	// Navigate to the grandchild (index 2)
	model.selectedIdx = 2
	grandchild := model.entries[2].Entry
	child := model.entries[1].Entry
	parent := model.entries[0].Entry

	// Collapse all entries
	model.collapsed[parent.EntityID] = true
	model.collapsed[child.EntityID] = true

	// Now call ensureSelectedAndAncestorsExpanded - should re-expand ancestors
	model = model.ensureSelectedAndAncestorsExpanded()

	// The child and parent should be expanded (not collapsed)
	if model.collapsed[parent.EntityID] {
		t.Error("parent should be expanded when grandchild is selected")
	}
	if model.collapsed[child.EntityID] {
		t.Error("child should be expanded when grandchild is selected")
	}
	// Grandchild has no children, so it doesn't matter if it's in the collapsed map
	_ = grandchild
}

func TestUAT_Collapse_AncestorsStayExpandedAfterReload(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a parent with a child
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Parent\n  . Child", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Expand the parent (press Enter)
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	// Should now see 2 entries
	if len(model.entries) != 2 {
		t.Fatalf("expected 2 entries after expand, got %d", len(model.entries))
	}

	// Navigate to child
	model.selectedIdx = 1

	// Simulate agenda reload (this happens after editing an entry)
	reloadCmd := model.loadAgendaCmd()
	reloadMsg := reloadCmd()
	newModel, _ = model.Update(reloadMsg)
	model = newModel.(Model)

	// After reload, child's parent should still be expanded
	if len(model.entries) != 2 {
		t.Errorf("after reload, expected 2 entries (parent still expanded), got %d", len(model.entries))
	}
}

func TestUAT_Collapse_CtrlC_CollapsesAllSiblings(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create two parent entries, each with children
	opts := service.LogEntriesOptions{Date: time.Now()}
	_, err := bujoSvc.LogEntries(ctx, ". Parent 1\n  . Child 1a\n. Parent 2\n  . Child 2a", opts)
	if err != nil {
		t.Fatalf("failed to log entries: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Load the journal view
	cmd := model.Init()
	if cmd != nil {
		agendaMsg := cmd()
		newModel, cmd := model.Update(agendaMsg)
		model = newModel.(Model)
		if cmd != nil {
			goalsMsg := cmd()
			newModel, _ = model.Update(goalsMsg)
			model = newModel.(Model)
		}
	}

	// Manually expand both parents
	for i, item := range model.entries {
		if item.HasChildren {
			model.collapsed[item.Entry.EntityID] = false
			_ = i
		}
	}
	model.entries = model.flattenAgenda(model.agenda)

	// After manual expansion, should see all 4 entries
	if len(model.entries) != 4 {
		t.Fatalf("expected 4 visible entries after expansion, got %d", len(model.entries))
	}

	// Press Ctrl+C to collapse all siblings
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	// After Ctrl+C, all siblings should be collapsed - we should see only 2 parent entries
	if len(model.entries) != 2 {
		t.Errorf("after Ctrl+C, expected 2 entries (all collapsed), got %d", len(model.entries))
	}
}

func TestUAT_HabitsView_BackspaceRemovesOccurrence(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit with 2 logs for today
	if err := habitSvc.LogHabit(ctx, "Exercise", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}
	if err := habitSvc.LogHabit(ctx, "Exercise", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view and load (key 5)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Navigate to rightmost day (today) - selectedDayIdx should be days-1
	days := HabitDaysWeek
	model.habitState.selectedDayIdx = days - 1

	if len(model.habitState.habits) == 0 {
		t.Fatal("should have habits")
	}

	initialCount := model.habitState.habits[0].TodayCount
	if initialCount != 2 {
		t.Fatalf("expected 2 logs for today, got %d", initialCount)
	}

	// Press Backspace to remove one occurrence
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)

	if cmd == nil {
		t.Fatal("removing habit log should return a command")
	}

	// Process the remove message
	removeMsg := cmd()
	newModel, cmd = model.Update(removeMsg)
	model = newModel.(Model)

	// Process reload command
	if cmd != nil {
		reloadMsg := cmd()
		newModel, _ = model.Update(reloadMsg)
		model = newModel.(Model)
	}

	// Verify count decreased by 1
	newCount := model.habitState.habits[0].TodayCount
	if newCount != initialCount-1 {
		t.Errorf("today's count should decrease from %d to %d, got %d", initialCount, initialCount-1, newCount)
	}
}

// =============================================================================
// UAT Section: Markdown Rendering in AI Summaries (#132)
// =============================================================================
