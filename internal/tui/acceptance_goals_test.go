package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/service"
)

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
