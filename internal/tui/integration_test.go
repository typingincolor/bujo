package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

func setupTestServices(t *testing.T) (*service.BujoService, *service.HabitService, *service.ListService, *service.GoalService) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	habitRepo := sqlite.NewHabitRepository(db)
	habitLogRepo := sqlite.NewHabitLogRepository(db)
	listRepo := sqlite.NewListRepository(db)
	listItemRepo := sqlite.NewListItemRepository(db)
	goalRepo := sqlite.NewGoalRepository(db)
	dayContextRepo := sqlite.NewDayContextRepository(db)
	parser := domain.NewTreeParser()

	bujoService := service.NewBujoService(entryRepo, dayContextRepo, parser)
	habitService := service.NewHabitService(habitRepo, habitLogRepo)
	listService := service.NewListService(listRepo, listItemRepo)
	goalService := service.NewGoalService(goalRepo)

	return bujoService, habitService, listService, goalService
}

func TestIntegration_HabitsView_LoadsDataFromService(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create test habits by logging them (LogHabit creates if not exists)
	if err := habitSvc.LogHabit(ctx, "Meditation", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}
	if err := habitSvc.LogHabit(ctx, "Exercise", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	// Create model with real services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to habits view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if model.currentView != ViewTypeHabits {
		t.Fatalf("expected ViewTypeHabits, got %v", model.currentView)
	}

	// Execute the command to load habits
	if cmd == nil {
		t.Fatal("expected a command to load habits")
	}
	loadMsg := cmd()

	// Process the loaded message
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Verify habits were loaded
	if len(model.habitState.habits) != 2 {
		t.Errorf("expected 2 habits, got %d", len(model.habitState.habits))
	}

	// Verify the view renders the habits
	view := model.View()
	if !strings.Contains(view, "Meditation") {
		t.Error("view should contain habit name 'Meditation'")
	}
	if !strings.Contains(view, "Exercise") {
		t.Error("view should contain habit name 'Exercise'")
	}
}

func TestIntegration_ListsView_LoadsDataFromService(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create test lists
	list1, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	list2, err := listSvc.CreateList(ctx, "Work Tasks")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	// Add items to lists
	if _, err := listSvc.AddItem(ctx, list1.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list2.ID, domain.EntryTypeTask, "Finish report"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	// Create model with real services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to lists view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if model.currentView != ViewTypeLists {
		t.Fatalf("expected ViewTypeLists, got %v", model.currentView)
	}

	// Execute the command to load lists
	if cmd == nil {
		t.Fatal("expected a command to load lists")
	}
	loadMsg := cmd()

	// Process the loaded message
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Verify lists were loaded
	if len(model.listState.lists) != 2 {
		t.Errorf("expected 2 lists, got %d", len(model.listState.lists))
	}

	// Verify the view renders the lists
	view := model.View()
	if !strings.Contains(view, "Shopping") {
		t.Error("view should contain list name 'Shopping'")
	}
	if !strings.Contains(view, "Work Tasks") {
		t.Error("view should contain list name 'Work Tasks'")
	}
}

func TestIntegration_ListItemsView_LoadsItemsFromService(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a list with items
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

	// Create model with real services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to lists view and load
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Press Enter to view items
	model.listState.selectedListIdx = 0
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)

	if model.currentView != ViewTypeListItems {
		t.Fatalf("expected ViewTypeListItems, got %v", model.currentView)
	}

	// Execute the command to load items
	if cmd == nil {
		t.Fatal("expected a command to load list items")
	}
	loadItemsMsg := cmd()

	// Process the loaded message
	newModel, _ = model.Update(loadItemsMsg)
	model = newModel.(Model)

	// Verify items were loaded
	if len(model.listState.items) != 2 {
		t.Errorf("expected 2 items, got %d", len(model.listState.items))
	}

	// Verify the view renders the items
	view := model.View()
	if !strings.Contains(view, "Buy milk") {
		t.Error("view should contain item 'Buy milk'")
	}
	if !strings.Contains(view, "Buy bread") {
		t.Error("view should contain item 'Buy bread'")
	}
}

func TestIntegration_JournalView_LoadsEntriesFromService(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create test entries for today using LogEntries
	today := time.Now()
	opts := service.LogEntriesOptions{Date: today}
	if _, err := bujoSvc.LogEntries(ctx, ". Complete project", opts); err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}
	if _, err := bujoSvc.LogEntries(ctx, "- Meeting notes", opts); err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	// Create model with real services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Init loads journal data
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init should return a command to load agenda")
	}
	loadMsg := cmd()

	// Process the loaded message
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	// Verify entries were loaded
	if len(model.entries) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(model.entries))
	}

	// Verify the view renders the entries
	view := model.View()
	if !strings.Contains(view, "Complete project") {
		t.Error("view should contain entry 'Complete project'")
	}
	if !strings.Contains(view, "Meeting notes") {
		t.Error("view should contain entry 'Meeting notes'")
	}
}

func TestIntegration_SwitchBetweenViews_MaintainsData(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create test data - LogHabit creates the habit if it doesn't exist
	if err := habitSvc.LogHabit(ctx, "Meditation", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	if _, err := listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk"); err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	opts := service.LogEntriesOptions{Date: time.Now()}
	if _, err := bujoSvc.LogEntries(ctx, ". Do stuff", opts); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	// Create model
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Load journal (init)
	cmd := model.Init()
	loadMsg := cmd()
	newModel, _ := model.Update(loadMsg)
	model = newModel.(Model)

	// Switch to habits, load data
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)
	loadMsg = cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	if len(model.habitState.habits) == 0 {
		t.Error("habits should be loaded after switching to habits view")
	}

	// Switch to lists, load data
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)
	loadMsg = cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	if len(model.listState.lists) == 0 {
		t.Error("lists should be loaded after switching to lists view")
	}

	// Switch back to journal
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	newModel, cmd = model.Update(msg)
	model = newModel.(Model)
	loadMsg = cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Journal should reload
	if len(model.entries) == 0 {
		t.Error("entries should be loaded after switching back to journal view")
	}
}

func TestIntegration_GoalsView_LoadsDataFromService(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create test goals for current month
	currentMonth := time.Now()
	if _, err := goalSvc.CreateGoal(ctx, "Learn Go", currentMonth); err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}
	if _, err := goalSvc.CreateGoal(ctx, "Ship new feature", currentMonth); err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	// Create model with real services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if model.currentView != ViewTypeGoals {
		t.Fatalf("expected ViewTypeGoals, got %v", model.currentView)
	}

	// Execute the command to load goals
	if cmd == nil {
		t.Fatal("expected a command to load goals")
	}
	loadMsg := cmd()

	// Process the loaded message
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Verify goals were loaded
	if len(model.goalState.goals) != 2 {
		t.Errorf("expected 2 goals, got %d", len(model.goalState.goals))
	}

	// Verify the view renders the goals
	view := model.View()
	if !strings.Contains(view, "Learn Go") {
		t.Error("view should contain goal 'Learn Go'")
	}
	if !strings.Contains(view, "Ship new feature") {
		t.Error("view should contain goal 'Ship new feature'")
	}
}

func TestIntegration_GoalsView_ToggleGoalDone(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	// Create a test goal
	currentMonth := time.Now()
	goalID, err := goalSvc.CreateGoal(ctx, "Learn Go", currentMonth)
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	// Create model with real services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})
	model.width = 80
	model.height = 24

	// Switch to goals view and load
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	loadMsg := cmd()
	newModel, _ = model.Update(loadMsg)
	model = newModel.(Model)

	// Verify goal is not done
	if len(model.goalState.goals) == 0 {
		t.Fatal("expected at least one goal")
	}
	if model.goalState.goals[0].IsDone() {
		t.Error("goal should not be done initially")
	}

	// Press space to toggle done
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(spaceMsg)
	model = newModel.(Model)

	// Execute the toggle command
	if cmd != nil {
		toggleMsg := cmd()
		newModel, cmd = model.Update(toggleMsg)
		model = newModel.(Model)

		// Execute the reload command
		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify goal is now done
	goal, err := goalSvc.GetGoal(ctx, goalID)
	if err != nil {
		t.Fatalf("failed to get goal: %v", err)
	}
	if !goal.IsDone() {
		t.Error("goal should be marked as done after pressing space")
	}
}

func TestIntegration_ListItems_AddItem(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a list
	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}

	// Create model with services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to lists view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Enter list items view
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify we're in list items view
	if model.currentView != ViewTypeListItems {
		t.Fatalf("expected ViewTypeListItems, got %v", model.currentView)
	}

	// Verify list is empty initially
	if len(model.listState.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(model.listState.items))
	}

	// Press 'a' to add
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ = model.Update(aMsg)
	model = newModel.(Model)

	if !model.addMode.active {
		t.Fatal("add mode should be active after pressing 'a'")
	}

	// Type the item content
	for _, r := range "Buy milk" {
		runeMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = model.Update(runeMsg)
		model = newModel.(Model)
	}

	// Press Enter to submit
	enterMsg = tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)

	// Execute the add command
	if cmd != nil {
		addMsg := cmd()
		newModel, cmd = model.Update(addMsg)
		model = newModel.(Model)

		// Execute the reload command
		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify item was added
	if len(model.listState.items) != 1 {
		t.Errorf("expected 1 item, got %d", len(model.listState.items))
	}

	// Verify item content
	if len(model.listState.items) > 0 && model.listState.items[0].Content != "Buy milk" {
		t.Errorf("expected item content 'Buy milk', got '%s'", model.listState.items[0].Content)
	}

	// Verify the view shows the item
	view := model.View()
	if !strings.Contains(view, "Buy milk") {
		t.Error("view should show the added item 'Buy milk'")
	}

	// Also verify in the database
	items, err := listSvc.GetListItems(ctx, list.ID)
	if err != nil {
		t.Fatalf("failed to get list items: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("database should have 1 item, got %d", len(items))
	}
}

func TestIntegration_ListItems_DeleteItem(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a list with an item
	list, err := listSvc.CreateList(ctx, "Shopping")
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	_, err = listSvc.AddItem(ctx, list.ID, domain.EntryTypeTask, "Buy milk")
	if err != nil {
		t.Fatalf("failed to add item: %v", err)
	}

	// Create model with services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to lists view and load
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Enter list items view
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd = model.Update(enterMsg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify item exists
	if len(model.listState.items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(model.listState.items))
	}

	// Press 'd' to delete
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ = model.Update(dMsg)
	model = newModel.(Model)

	if !model.confirmMode.active {
		t.Fatal("confirm mode should be active after pressing 'd'")
	}

	// Press 'y' to confirm
	yMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, cmd = model.Update(yMsg)
	model = newModel.(Model)

	// Execute the delete command chain
	if cmd != nil {
		deleteMsg := cmd()
		newModel, cmd = model.Update(deleteMsg)
		model = newModel.(Model)

		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify item was deleted
	if len(model.listState.items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(model.listState.items))
	}

	// Verify in database
	items, err := listSvc.GetListItems(ctx, list.ID)
	if err != nil {
		t.Fatalf("failed to get list items: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("database should have 0 items, got %d", len(items))
	}
}

func TestIntegration_HabitsView_ShowsStreakAndCompletion(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit by logging it
	if err := habitSvc.LogHabit(ctx, "Meditation", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	// Create model with services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to habits view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)

	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Verify habits loaded
	if len(model.habitState.habits) != 1 {
		t.Fatalf("expected 1 habit, got %d", len(model.habitState.habits))
	}

	// Verify habit has correct data
	habit := model.habitState.habits[0]
	if habit.Name != "Meditation" {
		t.Errorf("expected habit name 'Meditation', got '%s'", habit.Name)
	}

	// Streak should be at least 1 since we logged today
	if habit.CurrentStreak < 1 {
		t.Errorf("expected streak >= 1, got %d", habit.CurrentStreak)
	}

	// Verify the view shows the habit with streak
	view := model.View()
	if !strings.Contains(view, "Meditation") {
		t.Error("view should show habit name")
	}
	if !strings.Contains(view, "streak") {
		t.Error("view should show streak information")
	}
}

func TestIntegration_HabitsView_LogHabitIncrementsCount(t *testing.T) {
	bujoSvc, habitSvc, listSvc, _ := setupTestServices(t)
	ctx := context.Background()

	// Create a habit by logging it once
	if err := habitSvc.LogHabit(ctx, "Exercise", 1); err != nil {
		t.Fatalf("failed to log habit: %v", err)
	}

	// Create model with services
	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
	})
	model.width = 80
	model.height = 24

	// Navigate to habits view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	newModel, cmd := model.Update(msg)
	model = newModel.(Model)
	if cmd != nil {
		loadMsg := cmd()
		newModel, _ = model.Update(loadMsg)
		model = newModel.(Model)
	}

	// Check initial today count
	initialCount := model.habitState.habits[0].TodayCount

	// Press space to log habit
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = model.Update(spaceMsg)
	model = newModel.(Model)

	// Execute the log command
	if cmd != nil {
		logMsg := cmd()
		newModel, cmd = model.Update(logMsg)
		model = newModel.(Model)

		// Execute the reload command
		if cmd != nil {
			reloadMsg := cmd()
			newModel, _ = model.Update(reloadMsg)
			model = newModel.(Model)
		}
	}

	// Verify count increased
	newCount := model.habitState.habits[0].TodayCount
	if newCount != initialCount+1 {
		t.Errorf("expected count to increase from %d to %d, got %d", initialCount, initialCount+1, newCount)
	}
}
