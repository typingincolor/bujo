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

func setupTestServices(t *testing.T) (*service.BujoService, *service.HabitService, *service.ListService) {
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
	dayContextRepo := sqlite.NewDayContextRepository(db)
	parser := domain.NewTreeParser()

	bujoService := service.NewBujoService(entryRepo, dayContextRepo, parser)
	habitService := service.NewHabitService(habitRepo, habitLogRepo)
	listService := service.NewListService(listRepo, listItemRepo)

	return bujoService, habitService, listService
}

func TestIntegration_HabitsView_LoadsDataFromService(t *testing.T) {
	bujoSvc, habitSvc, listSvc := setupTestServices(t)
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
	bujoSvc, habitSvc, listSvc := setupTestServices(t)
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
	bujoSvc, habitSvc, listSvc := setupTestServices(t)
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
	bujoSvc, habitSvc, listSvc := setupTestServices(t)
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
	bujoSvc, habitSvc, listSvc := setupTestServices(t)
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
