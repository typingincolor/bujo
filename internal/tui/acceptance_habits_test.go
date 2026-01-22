package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
