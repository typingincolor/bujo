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

	days, _ := bujoSvc.GetDayEntries(ctx, todayDate, todayDate)
	model.days = days
	model.entries = model.flattenDays(days)
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

	days, _ := bujoSvc.GetDayEntries(ctx, todayDate, todayDate)
	model.days = days
	model.entries = model.flattenDays(days)

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

	days, _ := bujoSvc.GetDayEntries(ctx, todayDate, todayDate)
	model.days = days
	model.entries = model.flattenDays(days)
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

	days, _ := bujoSvc.GetDayEntries(ctx, todayDate, todayDate)
	model.days = days
	model.entries = model.flattenDays(days)
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

	days, _ := bujoSvc.GetDayEntries(ctx, todayDate, todayDate)
	model.days = days
	model.entries = model.flattenDays(days)

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

	days, _ := bujoSvc.GetDayEntries(ctx, todayDate, todayDate)
	model.days = days
	model.entries = model.flattenDays(days)

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

	days, _ := bujoSvc.GetDayEntries(ctx, todayDate, todayDate)
	model.days = days
	model.entries = model.flattenDays(days)

	view := model.View()

	if strings.Contains(view, "Delete this entry") {
		t.Error("deleted entry should not appear in journal view")
	}

	if !strings.Contains(view, "Keep this entry") {
		t.Error("non-deleted entry should appear in journal view")
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

	// Manually expand all entries from days to see the full tree
	for _, day := range model.days {
		for _, entry := range day.Entries {
			model.collapsed[entry.EntityID] = false
		}
	}
	model.entries = model.flattenDays(model.days)

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

	// Simulate days reload (this happens after editing an entry)
	reloadCmd := model.loadDaysCmd()
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
	model.entries = model.flattenDays(model.days)

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
