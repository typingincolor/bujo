package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

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
