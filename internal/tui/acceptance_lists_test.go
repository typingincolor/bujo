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
