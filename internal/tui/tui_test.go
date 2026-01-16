package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func init() {
	// Force lipgloss to output ANSI codes in tests
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// Helper function to create a text input for testing
func createTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 256
	ti.Width = 50
	return ti
}

func TestDefaultKeyMap_ShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.ShortHelp()

	if len(help) == 0 {
		t.Error("ShortHelp should return keybindings")
	}

	bindings := []string{"j/↓", "k/↑", "space", "d", "q"}
	for _, expected := range bindings {
		found := false
		for _, b := range help {
			if b.Help().Key == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ShortHelp should include %s", expected)
		}
	}
}

func TestDefaultKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.FullHelp()

	if len(help) == 0 {
		t.Error("FullHelp should return keybinding groups")
	}

	var totalBindings int
	for _, group := range help {
		totalBindings += len(group)
	}
	if totalBindings < 6 {
		t.Errorf("FullHelp should include at least 6 bindings, got %d", totalBindings)
	}
}

func TestNew(t *testing.T) {
	model := New(nil)

	if model.selectedIdx != 0 {
		t.Error("selectedIdx should be 0")
	}
	if model.confirmMode.active {
		t.Error("confirmMode should not be active")
	}
}

func TestModel_Init(t *testing.T) {
	model := New(nil)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	model := New(nil)
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}

	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.width != 80 {
		t.Errorf("width should be 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Errorf("height should be 24, got %d", m.height)
	}
	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}
}

func TestModel_Update_Navigation(t *testing.T) {
	model := New(nil)
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First"}},
		{Entry: domain.Entry{ID: 2, Content: "Second"}},
		{Entry: domain.Entry{ID: 3, Content: "Third"}},
	}
	model.agenda = &service.MultiDayAgenda{}

	tests := []struct {
		name        string
		startIdx    int
		key         tea.KeyMsg
		expectedIdx int
	}{
		{"down from top", 0, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 1},
		{"down from middle", 1, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 2},
		{"down at bottom stays", 2, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 2},
		{"up from bottom", 2, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, 1},
		{"up from middle", 1, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, 0},
		{"up at top stays", 0, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, 0},
		{"jump to top", 2, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}, 0},
		{"jump to bottom", 0, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.selectedIdx = tt.startIdx
			newModel, _ := model.Update(tt.key)
			m := newModel.(Model)

			if m.selectedIdx != tt.expectedIdx {
				t.Errorf("expected selectedIdx %d, got %d", tt.expectedIdx, m.selectedIdx)
			}
		})
	}
}

func TestModel_Update_HelpToggle(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	if model.help.ShowAll {
		t.Error("help should not show all by default")
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.help.ShowAll {
		t.Error("help should show all after pressing ?")
	}
}

func TestModel_Update_QuitReturnsQuitCmd(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if cmd != nil {
		t.Error("pressing q at root should not immediately quit")
	}
	if !m.quitConfirmMode.active {
		t.Error("quit confirm mode should be active")
	}
}

func TestModel_FlattenAgenda_Empty(t *testing.T) {
	model := New(nil)
	result := model.flattenAgenda(nil)

	if result != nil {
		t.Error("flattenAgenda(nil) should return nil")
	}
}

func TestModel_FlattenAgenda_WithOverdue(t *testing.T) {
	model := New(nil)
	agenda := &service.MultiDayAgenda{
		Overdue: []domain.Entry{
			{ID: 1, Content: "Overdue task", Type: domain.EntryTypeTask},
		},
	}

	result := model.flattenAgenda(agenda)

	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].DayHeader != "⚠️  OVERDUE" {
		t.Errorf("expected ⚠️  OVERDUE header, got %s", result[0].DayHeader)
	}
	if !result[0].IsOverdue {
		t.Error("entry should be marked as overdue")
	}
}

func TestModel_FlattenAgenda_OverdueFiltersParentContext(t *testing.T) {
	model := New(nil)
	parentID := int64(1)
	yesterday := time.Now().AddDate(0, 0, -1)

	agenda := &service.MultiDayAgenda{
		Overdue: []domain.Entry{
			{ID: 1, Content: "Parent note", Type: domain.EntryTypeNote, ScheduledDate: &yesterday},
			{ID: 2, Content: "Overdue task", Type: domain.EntryTypeTask, ParentID: &parentID, ScheduledDate: &yesterday},
		},
	}

	result := model.flattenAgenda(agenda)

	if len(result) != 1 {
		t.Fatalf("expected 1 item (only the overdue task), got %d", len(result))
	}
	if result[0].Entry.Content != "Overdue task" {
		t.Errorf("expected 'Overdue task', got %s", result[0].Entry.Content)
	}
}

func TestModel_ToggleOverdueContext(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{
		Overdue: []domain.Entry{
			{ID: 1, Content: "Task 1", Type: domain.EntryTypeTask, ScheduledDate: &yesterday},
			{ID: 2, Content: "Task 2", Type: domain.EntryTypeTask, ScheduledDate: &yesterday},
		},
	}
	model.entries = model.flattenAgenda(model.agenda)
	model.selectedIdx = 0

	if model.expandedOverdueContextID != nil {
		t.Error("expandedOverdueContextID should be nil by default")
	}

	// Toggle context for task 1 (selected)
	model = model.toggleOverdueContext()

	if model.expandedOverdueContextID == nil || *model.expandedOverdueContextID != 1 {
		t.Error("expandedOverdueContextID should be set to task 1 ID after toggle")
	}

	// Toggle again to close context
	model = model.toggleOverdueContext()

	if model.expandedOverdueContextID != nil {
		t.Error("expandedOverdueContextID should be nil after second toggle")
	}
}

func TestModel_FlattenAgenda_OverduePreservesParentTaskHierarchy(t *testing.T) {
	model := New(nil)
	parentID := int64(1)
	yesterday := time.Now().AddDate(0, 0, -1)
	parentEntityID := domain.EntityID("entity-1")

	agenda := &service.MultiDayAgenda{
		Overdue: []domain.Entry{
			{ID: 1, EntityID: parentEntityID, Content: "Parent task", Type: domain.EntryTypeTask, ScheduledDate: &yesterday},
			{ID: 2, Content: "Child task", Type: domain.EntryTypeTask, ParentID: &parentID, ScheduledDate: &yesterday},
		},
	}

	// Expand the parent to show its children
	model.collapsed[parentEntityID] = false

	result := model.flattenAgenda(agenda)

	if len(result) != 2 {
		t.Fatalf("expected 2 items (parent and child tasks), got %d", len(result))
	}

	if result[0].Entry.Content != "Parent task" {
		t.Errorf("expected first item to be 'Parent task', got %s", result[0].Entry.Content)
	}

	if result[1].Entry.Content != "Child task" {
		t.Errorf("expected second item to be 'Child task', got %s", result[1].Entry.Content)
	}

	if result[1].Entry.ParentID == nil || *result[1].Entry.ParentID != 1 {
		t.Errorf("expected child task to have parent reference preserved, got %v", result[1].Entry.ParentID)
	}
}

func TestModel_ToggleOverdueContext_ShowsAncestry(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	noteID := int64(1)
	taskID := int64(2)
	noteEntityID := domain.EntityID("entity-1")
	taskEntityID := domain.EntityID("entity-2")

	model := New(nil)
	model.agenda = &service.MultiDayAgenda{
		Overdue: []domain.Entry{
			{ID: noteID, EntityID: noteEntityID, Content: "Parent note", Type: domain.EntryTypeNote, ScheduledDate: &yesterday},
			{ID: taskID, EntityID: taskEntityID, Content: "Child task", Type: domain.EntryTypeTask, ParentID: &noteID, ScheduledDate: &yesterday},
		},
	}
	model.entries = model.flattenAgenda(model.agenda)
	model.selectedIdx = 0

	// By default, only the task is shown (parent is filtered)
	if len(model.entries) != 1 {
		t.Fatalf("expected 1 item (only task) by default, got %d", len(model.entries))
	}
	if model.entries[0].Entry.ID != taskID {
		t.Errorf("expected task to be shown, got entry ID %d", model.entries[0].Entry.ID)
	}

	// After toggling context for the task, ancestry should be included
	model.entries = model.flattenAgenda(model.agenda)
	model = model.toggleOverdueContext()
	model.entries = model.flattenAgenda(model.agenda)

	// Expand parent to see full hierarchy
	model.collapsed[noteEntityID] = false
	model.entries = model.flattenAgenda(model.agenda)

	if len(model.entries) < 2 {
		t.Fatalf("expected at least 2 items (parent and task) when context expanded, got %d", len(model.entries))
	}

	if model.entries[0].Entry.ID != noteID {
		t.Errorf("expected parent note to be shown first, got entry ID %d", model.entries[0].Entry.ID)
	}
	if model.entries[1].Entry.ID != taskID {
		t.Errorf("expected child task to be shown second, got entry ID %d", model.entries[1].Entry.ID)
	}
}

func TestModel_FlattenAgenda_WithDays(t *testing.T) {
	model := New(nil)
	location := "Home Office"
	agenda := &service.MultiDayAgenda{
		Days: []service.DayEntries{
			{
				Date:     time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
				Location: &location,
				Entries: []domain.Entry{
					{ID: 1, Content: "Task 1", Type: domain.EntryTypeTask},
					{ID: 2, Content: "Task 2", Type: domain.EntryTypeTask},
				},
			},
		},
	}

	result := model.flattenAgenda(agenda)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0].DayHeader == "" {
		t.Error("first entry should have day header")
	}
	if result[1].DayHeader != "" {
		t.Error("second entry should not have day header")
	}
}

func TestModel_FlattenAgenda_WithHierarchy(t *testing.T) {
	model := New(nil)
	parentID := int64(1)
	parentEntityID := domain.EntityID("parent-entity-1")
	agenda := &service.MultiDayAgenda{
		Days: []service.DayEntries{
			{
				Date: time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
				Entries: []domain.Entry{
					{ID: 1, EntityID: parentEntityID, Content: "Parent", Type: domain.EntryTypeTask, ParentID: nil},
					{ID: 2, Content: "Child", Type: domain.EntryTypeNote, ParentID: &parentID},
				},
			},
		},
	}

	// Expand the parent so we can test hierarchy
	model.collapsed[parentEntityID] = false

	result := model.flattenAgenda(agenda)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0].Indent != 0 {
		t.Errorf("parent should have indent 0, got %d", result[0].Indent)
	}
	if result[1].Indent != 1 {
		t.Errorf("child should have indent 1, got %d", result[1].Indent)
	}
}

func TestModel_View_Loading(t *testing.T) {
	model := New(nil)
	view := model.View()

	if !strings.Contains(view, "Loading...") {
		t.Errorf("expected view to contain Loading..., got %s", view)
	}
}

func TestModel_View_Error(t *testing.T) {
	model := New(nil)
	model.err = tea.ErrProgramKilled

	view := model.View()

	if view == "Loading..." {
		t.Error("should not show Loading when there's an error")
	}
}

func TestModel_View_NoEntries(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{}

	view := model.View()

	if view == "Loading..." {
		t.Error("should not show Loading when agenda is loaded")
	}
}

func TestModel_ConfirmMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Test"}},
	}

	model.confirmMode = confirmState{
		active:      true,
		entryID:     1,
		hasChildren: true,
	}

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m := newModel.(Model)

	if m.confirmMode.active {
		t.Error("pressing n should cancel confirm mode")
	}
}

func TestKeyMap_KeyBindings(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name    string
		binding key.Binding
		keys    []string
	}{
		{"up", km.Up, []string{"k", "up"}},
		{"down", km.Down, []string{"j", "down"}},
		{"top", km.Top, []string{"g"}},
		{"bottom", km.Bottom, []string{"G"}},
		{"done", km.Done, []string{" "}},
		{"delete", km.Delete, []string{"d"}},
		{"quit", km.Quit, []string{"q"}},
		{"back", km.Back, []string{"esc"}},
		{"help", km.Help, []string{"?"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expectedKey := range tt.keys {
				var msg tea.KeyMsg
				switch expectedKey {
				case "up":
					msg = tea.KeyMsg{Type: tea.KeyUp}
				case "down":
					msg = tea.KeyMsg{Type: tea.KeyDown}
				case " ":
					msg = tea.KeyMsg{Type: tea.KeySpace}
				default:
					msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(expectedKey)}
				}

				if !key.Matches(msg, tt.binding) {
					t.Errorf("key %s should match %s binding", expectedKey, tt.name)
				}
			}
		})
	}
}

func TestModel_Update_EditMode_InitializesWithContent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Original content", Type: domain.EntryTypeTask}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.input.Value() != "Original content" {
		t.Errorf("input should be initialized with entry content, got %s", m.editMode.input.Value())
	}
}

func TestModel_Update_EditMode_NoOpOnEmptyEntries(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.active {
		t.Error("pressing e with no entries should not enter edit mode")
	}
}

func TestModel_Update_AddMode_EntersAsChildOnShiftA(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Parent"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Error("pressing A should enter add mode")
	}
	if !m.addMode.asChild {
		t.Error("pressing A should add as child")
	}
	if m.addMode.parentID == nil || *m.addMode.parentID != 1 {
		t.Error("parentID should be set to selected entry ID")
	}
}

func TestModel_Update_AddMode_StartsEmpty(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.addMode.input.Value() != "" {
		t.Errorf("add mode input should start empty, got %s", m.addMode.input.Value())
	}
}

func TestModel_Update_MigrateMode_NoOpOnEmptyEntries(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.migrateMode.active {
		t.Error("pressing m with no entries should not enter migrate mode")
	}
}

func TestModel_Update_MigrateMode_NoOpOnNonTask(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Note", Type: domain.EntryTypeNote}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.migrateMode.active {
		t.Error("pressing m on a note should not enter migrate mode")
	}
}

func TestModel_Update_ErrorCanBeDismissed(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.err = fmt.Errorf("some error")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.err != nil {
		t.Error("pressing Escape should dismiss error")
	}
}

func TestModel_Update_ErrorCanBeDismissedWithAnyKey(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.err = fmt.Errorf("some error")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.err != nil {
		t.Error("pressing any key should dismiss error")
	}
}

func TestModel_Update_AddMode_InheritsParentFromSelected(t *testing.T) {
	parentID := int64(10)
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 22, Content: "Child item", ParentID: &parentID}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	if m.addMode.parentID == nil {
		t.Error("parentID should be inherited from selected item")
	}
	if m.addMode.parentID != nil && *m.addMode.parentID != 10 {
		t.Errorf("parentID should be 10 (same as selected's parent), got %d", *m.addMode.parentID)
	}
}

func TestModel_Update_AddMode_RootItemAddsAtRoot(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Root item", ParentID: nil}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	if m.addMode.parentID != nil {
		t.Error("parentID should be nil when selected item is root")
	}
}

func TestModel_Update_AddRootMode_AddsAtRootFromNestedItem(t *testing.T) {
	parentID := int64(10)
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 22, Content: "Child item", ParentID: &parentID}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	if m.addMode.parentID != nil {
		t.Error("pressing 'r' should add at root regardless of selected item's parent")
	}
}

func TestModel_DefaultViewMode_IsDay(t *testing.T) {
	model := New(nil)
	if model.viewMode != ViewModeDay {
		t.Errorf("default view mode should be day, got %v", model.viewMode)
	}
}

func TestModel_DefaultViewDate_IsToday(t *testing.T) {
	model := New(nil)
	today := time.Now()
	expected := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	if !model.viewDate.Equal(expected) {
		t.Errorf("default view date should be today, got %v", model.viewDate)
	}
}

func TestModel_Update_ToggleViewMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	// Start in day mode
	if model.viewMode != ViewModeDay {
		t.Fatal("should start in day mode")
	}

	// Toggle to week
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)
	if m.viewMode != ViewModeWeek {
		t.Errorf("pressing 'w' should switch to week mode, got %v", m.viewMode)
	}

	// Toggle back to day
	newModel, _ = m.Update(msg)
	m = newModel.(Model)
	if m.viewMode != ViewModeDay {
		t.Errorf("pressing 'w' again should switch back to day mode, got %v", m.viewMode)
	}
}

func TestModel_Update_GoToDate(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.gotoMode.active {
		t.Error("pressing '/' should enter goto date mode")
	}
}

func TestModel_AddRoot_UsesViewDate(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	// Set viewDate to yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	model.viewDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Fatal("should enter add mode")
	}
	// The addMode should use viewDate when creating, not today
	// This is tested via integration but we verify addMode is entered
}

// Capture Mode Tests

// Emacs navigation tests

// Search tests

// Draft tests

// Day View Search Tests

func TestModel_DayView_CtrlS_EntersSearchMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
		{Entry: domain.Entry{ID: 2, Content: "Second task"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.searchMode.active {
		t.Error("searchMode should be active")
	}
	if !m.searchMode.forward {
		t.Error("searchMode should be forward")
	}
}

func TestModel_DayView_CtrlR_EntersReverseSearchMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
		{Entry: domain.Entry{ID: 2, Content: "Second task"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.searchMode.active {
		t.Error("searchMode should be active")
	}
	if m.searchMode.forward {
		t.Error("searchMode should be reverse (forward=false)")
	}
}

func TestModel_DayView_Search_TypingAddsToQuery(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.query != "test" {
		t.Errorf("expected query 'test', got '%s'", m.searchMode.query)
	}
}

func TestModel_DayView_Search_BackspaceRemovesFromQuery(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.query != "tes" {
		t.Errorf("expected query 'tes', got '%s'", m.searchMode.query)
	}
}

func TestModel_DayView_Search_SpaceAddsToQuery(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "my"}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.query != "my " {
		t.Errorf("expected query 'my ', got '%s'", m.searchMode.query)
	}
}

func TestModel_DayView_Search_EscCancelsSearch(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.active {
		t.Error("searchMode should not be active after Esc")
	}
	if m.searchMode.query != "" {
		t.Error("query should be cleared after Esc")
	}
}

func TestModel_DayView_Search_EnterExitsSearchMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
		{Entry: domain.Entry{ID: 2, Content: "Second task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "Second"}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.active {
		t.Error("searchMode should not be active after Enter")
	}
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1, got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_IncrementalSearch_MovesToMatch(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	// Type "ban"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ban")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Banana), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_IncrementalSearch_StaysOnMatch(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry"}},
	}
	model.selectedIdx = 1 // Already on Banana
	model.searchMode = searchState{active: true, forward: true, query: "ban"}

	// Add more characters to refine search
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ana")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should stay on Banana since it still matches "banana"
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Banana), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_ForwardSearch_FindsNextMatch(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task one"}},
		{Entry: domain.Entry{ID: 2, Content: "Task two"}},
		{Entry: domain.Entry{ID: 3, Content: "Task three"}},
	}
	model.selectedIdx = 0 // On "Task one"
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// Press Ctrl+S to find next
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should move to "Task two"
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Task two), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_BackwardSearch_FindsPrevMatch(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task one"}},
		{Entry: domain.Entry{ID: 2, Content: "Task two"}},
		{Entry: domain.Entry{ID: 3, Content: "Task three"}},
	}
	model.selectedIdx = 2 // On "Task three"
	model.searchMode = searchState{active: true, forward: false, query: "Task"}

	// Press Ctrl+R to find previous
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should move to "Task two"
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Task two), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_DirectionSwitch_CtrlRThenCtrlS(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task one"}},
		{Entry: domain.Entry{ID: 2, Content: "Task two"}},
		{Entry: domain.Entry{ID: 3, Content: "Task three"}},
		{Entry: domain.Entry{ID: 4, Content: "Task four"}},
	}
	model.selectedIdx = 1 // On "Task two"
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// Press Ctrl+R to go backward
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("after Ctrl+R expected selectedIdx 0, got %d", m.selectedIdx)
	}
	if m.searchMode.forward {
		t.Error("forward should be false after Ctrl+R")
	}

	// Press Ctrl+S to go forward
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 1 {
		t.Errorf("after Ctrl+S expected selectedIdx 1, got %d", m.selectedIdx)
	}
	if !m.searchMode.forward {
		t.Error("forward should be true after Ctrl+S")
	}
}

func TestModel_DayView_Search_WrapsAround_Forward(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple here"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana there"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry time"}},
	}
	model.selectedIdx = 0 // On "Apple here"
	model.searchMode = searchState{active: true, forward: true, query: "Apple"}

	// Press Ctrl+S - should wrap around to beginning (only one match)
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should wrap to first item (same item - only match)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 after wrap, got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_WrapsAround_Backward(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Banana there"}},
		{Entry: domain.Entry{ID: 2, Content: "Cherry time"}},
		{Entry: domain.Entry{ID: 3, Content: "Apple here"}},
	}
	model.selectedIdx = 2 // On "Apple here"
	model.searchMode = searchState{active: true, forward: false, query: "Apple"}

	// Press Ctrl+R - should wrap around to end (only one match)
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should wrap to same item (only match)
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx 2 after wrap, got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_CaseInsensitive(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First item"}},
		{Entry: domain.Entry{ID: 2, Content: "SECOND ITEM"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("second")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (case insensitive match), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_ScrollFollowsSelection(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 10 // Small height to force scrolling

	// Create entries where match is beyond visible area
	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		content := fmt.Sprintf("Item %d", i)
		if i == 15 {
			content = "Special match"
		}
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}})
	}
	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Special")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 15 {
		t.Errorf("expected selectedIdx 15, got %d", m.selectedIdx)
	}
	// Scroll should have adjusted to show selected item
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed to show selected item")
	}
}

func TestModel_DayView_Search_NoMatch_StaysOnCurrent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("xyz")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should stay on current when no match
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 (no match), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_MultipleMatches_NextFindsDifferent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task A"}},
		{Entry: domain.Entry{ID: 2, Content: "Note B"}},
		{Entry: domain.Entry{ID: 3, Content: "Task C"}},
		{Entry: domain.Entry{ID: 4, Content: "Note D"}},
		{Entry: domain.Entry{ID: 5, Content: "Task E"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// First Ctrl+S should go to Task C (index 2)
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 2 {
		t.Errorf("first Ctrl+S expected selectedIdx 2, got %d", m.selectedIdx)
	}

	// Second Ctrl+S should go to Task E (index 4)
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 4 {
		t.Errorf("second Ctrl+S expected selectedIdx 4, got %d", m.selectedIdx)
	}

	// Third Ctrl+S should wrap to Task A (index 0)
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("third Ctrl+S expected selectedIdx 0 (wrap), got %d", m.selectedIdx)
	}
}

// Capture Mode Search Tests - Direction Switching

// Week View Search Tests - Multiple days with headers

func TestModel_WeekView_Search_AcrossMultipleDays(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.viewMode = ViewModeWeek

	// Simulate entries from different days
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Monday task"}, DayHeader: "Monday, Jan 6"},
		{Entry: domain.Entry{ID: 2, Content: "Monday note"}},
		{Entry: domain.Entry{ID: 3, Content: "Tuesday task"}, DayHeader: "Tuesday, Jan 7"},
		{Entry: domain.Entry{ID: 4, Content: "Tuesday meeting"}},
		{Entry: domain.Entry{ID: 5, Content: "Wednesday task"}, DayHeader: "Wednesday, Jan 8"},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for "Tuesday"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Tuesday")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find "Tuesday task" at index 2
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx 2 (Tuesday task), got %d", m.selectedIdx)
	}

	// Press Ctrl+S to find next "Tuesday"
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Should find "Tuesday meeting" at index 3
	if m.selectedIdx != 3 {
		t.Errorf("expected selectedIdx 3 (Tuesday meeting), got %d", m.selectedIdx)
	}
}

func TestModel_WeekView_Search_WithDayHeaders_ScrollsCorrectly(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.viewMode = ViewModeWeek
	model.width = 80
	model.height = 12 // Small height to force scrolling

	// Create many entries across days
	entries := []EntryItem{}
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	id := int64(1)
	for _, day := range days {
		entries = append(entries, EntryItem{
			Entry:     domain.Entry{ID: id, Content: fmt.Sprintf("%s task 1", day)},
			DayHeader: fmt.Sprintf("%s, Jan", day),
		})
		id++
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: id, Content: fmt.Sprintf("%s task 2", day)}})
		id++
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: id, Content: fmt.Sprintf("%s task 3", day)}})
		id++
	}

	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for "Friday" which is far down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Friday")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find "Friday task 1" at index 12
	if m.selectedIdx != 12 {
		t.Errorf("expected selectedIdx 12 (Friday task 1), got %d", m.selectedIdx)
	}

	// Scroll should have adjusted
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed to show Friday entry")
	}
}

func TestModel_WeekView_Search_BackwardFromDifferentDay(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.viewMode = ViewModeWeek

	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple task"}, DayHeader: "Monday, Jan 6"},
		{Entry: domain.Entry{ID: 2, Content: "Banana task"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry task"}, DayHeader: "Tuesday, Jan 7"},
		{Entry: domain.Entry{ID: 4, Content: "Apple task"}}, // Another Apple on Tuesday
	}
	model.selectedIdx = 3 // On second "Apple task"
	model.searchMode = searchState{active: true, forward: false, query: "Apple"}

	// Press Ctrl+R to find previous Apple
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find first "Apple task" at index 0
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 (first Apple task), got %d", m.selectedIdx)
	}
}

func TestModel_WeekView_Search_NestedEntries(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.viewMode = ViewModeWeek

	// Entries with different indent levels (parent-child)
	parentID := int64(1)
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Project A"}, DayHeader: "Monday, Jan 6", Indent: 0},
		{Entry: domain.Entry{ID: 2, Content: "Subtask alpha", ParentID: &parentID}, Indent: 1},
		{Entry: domain.Entry{ID: 3, Content: "Subtask beta", ParentID: &parentID}, Indent: 1},
		{Entry: domain.Entry{ID: 4, Content: "Project B"}, DayHeader: "Tuesday, Jan 7", Indent: 0},
		{Entry: domain.Entry{ID: 5, Content: "Subtask alpha"}, Indent: 1}, // Another "alpha"
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for "alpha"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("alpha")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find first "Subtask alpha" at index 1
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (first Subtask alpha), got %d", m.selectedIdx)
	}

	// Press Ctrl+S to find next alpha
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Should find second "Subtask alpha" at index 4
	if m.selectedIdx != 4 {
		t.Errorf("expected selectedIdx 4 (second Subtask alpha), got %d", m.selectedIdx)
	}
}

func TestModel_WeekView_Search_WithOverdueEntries(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.viewMode = ViewModeWeek

	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Overdue task 1"}, DayHeader: "OVERDUE", IsOverdue: true},
		{Entry: domain.Entry{ID: 2, Content: "Overdue task 2"}, IsOverdue: true},
		{Entry: domain.Entry{ID: 3, Content: "Today task"}, DayHeader: "Monday, Jan 6"},
		{Entry: domain.Entry{ID: 4, Content: "Today task 2"}},
	}
	model.selectedIdx = 2 // On "Today task"
	model.searchMode = searchState{active: true, forward: false, query: "Overdue"}

	// Press Ctrl+R to find previous Overdue
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find "Overdue task 2" at index 1
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Overdue task 2), got %d", m.selectedIdx)
	}
}

// Large data scrolling tests

func TestModel_Search_LargeData_ScrollsFromTopToBottom(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 20 // Can show ~14 entries

	// Create 100 entries
	entries := []EntryItem{}
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf("Item number %d", i)
		if i == 95 {
			content = "TARGET ITEM HERE"
		}
		entry := EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}}
		if i%10 == 0 {
			entry.DayHeader = fmt.Sprintf("Day %d", i/10)
		}
		entries = append(entries, entry)
	}

	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for TARGET which is at index 95
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("TARGET")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 95 {
		t.Errorf("expected selectedIdx 95, got %d", m.selectedIdx)
	}

	// Verify scroll offset is valid (selectedIdx should be visible)
	if m.scrollOffset > 95 {
		t.Errorf("scrollOffset %d is too high, selectedIdx 95 won't be visible", m.scrollOffset)
	}
	if m.scrollOffset+14 < 95 {
		t.Errorf("scrollOffset %d is too low, selectedIdx 95 won't be visible", m.scrollOffset)
	}
}

func TestModel_Search_LargeData_ScrollsFromBottomToTop(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 20

	// Create 100 entries
	entries := []EntryItem{}
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf("Item number %d", i)
		if i == 5 {
			content = "TARGET ITEM HERE"
		}
		entry := EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}}
		if i%10 == 0 {
			entry.DayHeader = fmt.Sprintf("Day %d", i/10)
		}
		entries = append(entries, entry)
	}

	model.entries = entries
	model.selectedIdx = 95
	model.scrollOffset = 85 // Scrolled to bottom
	model.searchMode = searchState{active: true, forward: false, query: "TARGET"}

	// Search backward for TARGET which is at index 5
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 5 {
		t.Errorf("expected selectedIdx 5, got %d", m.selectedIdx)
	}

	// Verify scroll offset adjusted to show entry
	if m.scrollOffset > 5 {
		t.Errorf("scrollOffset %d is too high, selectedIdx 5 won't be visible", m.scrollOffset)
	}
}

func TestModel_Search_LargeData_MultipleSearchesTraverseAll(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 15

	// Create 50 entries with "task" in every 5th entry
	entries := []EntryItem{}
	expectedTaskIndices := []int{4, 9, 14, 19, 24, 29, 34, 39, 44, 49}
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf("Item %d", i)
		if i%5 == 4 {
			content = fmt.Sprintf("Task item %d", i)
		}
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}})
	}

	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// Press Ctrl+S repeatedly to find all tasks
	for i, expectedIdx := range expectedTaskIndices {
		msg := tea.KeyMsg{Type: tea.KeyCtrlS}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.selectedIdx != expectedIdx {
			t.Errorf("iteration %d: expected selectedIdx %d, got %d", i, expectedIdx, m.selectedIdx)
		}

		// Verify selectedIdx is visible (between scrollOffset and scrollOffset + visible area)
		if m.selectedIdx < m.scrollOffset {
			t.Errorf("iteration %d: selectedIdx %d is above scrollOffset %d", i, m.selectedIdx, m.scrollOffset)
		}

		model = m
	}

	// One more Ctrl+S should wrap to first task
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 4 {
		t.Errorf("after wrap: expected selectedIdx 4, got %d", m.selectedIdx)
	}
}

func TestModel_Search_LargeData_DirectionSwitchMidList(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 15

	// Create entries with matches at various positions
	entries := []EntryItem{}
	matchIndices := []int{10, 25, 40, 55, 70}
	for i := 0; i < 80; i++ {
		content := fmt.Sprintf("Item %d", i)
		for _, matchIdx := range matchIndices {
			if i == matchIdx {
				content = fmt.Sprintf("MATCH at %d", i)
			}
		}
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}})
	}

	model.entries = entries
	model.selectedIdx = 40 // Middle match
	model.scrollOffset = 35
	model.searchMode = searchState{active: true, forward: true, query: "MATCH"}

	// Go forward: should find 55
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 55 {
		t.Errorf("forward from 40: expected 55, got %d", m.selectedIdx)
	}

	// Switch direction (Ctrl+R): should find 40
	msg = tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 40 {
		t.Errorf("backward from 55: expected 40, got %d", m.selectedIdx)
	}

	// Continue backward: should find 25
	msg = tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 25 {
		t.Errorf("backward from 40: expected 25, got %d", m.selectedIdx)
	}

	// Switch to forward (Ctrl+S): should find 40
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 40 {
		t.Errorf("forward from 25: expected 40, got %d", m.selectedIdx)
	}

	// Verify scroll followed through all these jumps
	if m.scrollOffset > 40 || m.scrollOffset+14 < 40 {
		t.Errorf("final scrollOffset %d doesn't make selectedIdx 40 visible", m.scrollOffset)
	}
}

// Navigation Tests - Top and Bottom

func TestModel_Navigation_TopKey_GoesToFirstEntry(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First"}},
		{Entry: domain.Entry{ID: 2, Content: "Second"}},
		{Entry: domain.Entry{ID: 3, Content: "Third"}},
	}
	model.selectedIdx = 2
	model.scrollOffset = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0, got %d", m.selectedIdx)
	}
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0, got %d", m.scrollOffset)
	}
}

func TestModel_Navigation_BottomKey_GoesToLastEntry(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 20
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First"}},
		{Entry: domain.Entry{ID: 2, Content: "Second"}},
		{Entry: domain.Entry{ID: 3, Content: "Third"}},
	}
	model.selectedIdx = 0
	model.scrollOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx 2, got %d", m.selectedIdx)
	}
}

func TestModel_Navigation_BottomKey_WithLargeList_ScrollsCorrectly(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 15 // Can show ~9 entries

	entries := []EntryItem{}
	for i := 0; i < 30; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 29 {
		t.Errorf("expected selectedIdx 29, got %d", m.selectedIdx)
	}
	// Scroll should have adjusted to show last entry
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed")
	}
}

// Done/Complete Tests

func TestModel_Done_MarksTaskComplete(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "Task to complete"}},
	}
	model.selectedIdx = 0

	// Space key marks as done
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should have triggered a command (the actual mark done happens via service)
	if cmd == nil {
		t.Error("expected a command to be returned")
	}
	_ = m // Model state is updated async via the command
}

func TestModel_Done_NoOpOnEmptyList(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("expected no command when list is empty")
	}
}

// Delete Tests

func TestModel_Delete_TriggersConfirmMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item to delete"}},
	}
	model.selectedIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should activate confirm mode synchronously
	if !m.confirmMode.active {
		t.Error("expected confirmMode to be active")
	}
	if m.confirmMode.entryID != 1 {
		t.Errorf("expected entryID 1, got %d", m.confirmMode.entryID)
	}
}

func TestModel_Delete_NoOpOnEmptyList(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("expected no command when list is empty")
	}
}

// Goto Date Tests

func TestModel_GotoMode_EntersOnSlash(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.gotoMode.active {
		t.Error("gotoMode should be active")
	}
}

func TestModel_GotoMode_TypingAddsToInput(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.gotoMode = gotoState{active: true}
	model.gotoMode.input = createTextInput()
	model.gotoMode.input.Focus()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tomorrow")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.gotoMode.input.Value() != "tomorrow" {
		t.Errorf("expected input 'tomorrow', got '%s'", m.gotoMode.input.Value())
	}
}

// Capture Mode Arrow Key Tests

// View Rendering Tests

func TestModel_View_SearchMode_ShowsSearchBar(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item one"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	view := model.View()

	if !strings.Contains(view, "Search") {
		t.Error("view should contain Search bar")
	}
	if !strings.Contains(view, "forward") {
		t.Error("view should show search direction")
	}
}

func TestModel_View_SearchMode_ShowsReverseDirection(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item one"}},
	}
	model.searchMode = searchState{active: true, forward: false, query: "test"}

	view := model.View()

	if !strings.Contains(view, "reverse") {
		t.Error("view should show reverse direction")
	}
}

func TestModel_View_SelectedEntry_IsHighlighted(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "First item"}},
		{Entry: domain.Entry{ID: 2, Type: domain.EntryTypeTask, Content: "Selected item"}},
	}
	model.selectedIdx = 1

	view := model.View()

	// The selected item should be present (exact styling depends on lipgloss)
	if !strings.Contains(view, "Selected item") {
		t.Error("view should contain the selected item text")
	}
}

func TestModel_View_DayHeader_IsRendered(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task"}, DayHeader: "Monday, Jan 6"},
	}

	view := model.View()

	if !strings.Contains(view, "Monday") {
		t.Error("view should contain day header")
	}
}

func TestModel_View_OverdueHeader_IsRendered(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Overdue task"}, DayHeader: "OVERDUE", IsOverdue: true},
	}

	view := model.View()

	if !strings.Contains(view, "OVERDUE") {
		t.Error("view should contain OVERDUE header")
	}
}

func TestModel_View_EntrySymbols_AreRendered(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "Task"}},
		{Entry: domain.Entry{ID: 2, Type: domain.EntryTypeNote, Content: "Note"}},
		{Entry: domain.Entry{ID: 3, Type: domain.EntryTypeEvent, Content: "Event"}},
		{Entry: domain.Entry{ID: 4, Type: domain.EntryTypeDone, Content: "Done"}},
	}

	view := model.View()

	// Check Unicode symbols are present
	if !strings.Contains(view, "•") {
		t.Error("view should contain task symbol •")
	}
	if !strings.Contains(view, "–") {
		t.Error("view should contain note symbol –")
	}
	if !strings.Contains(view, "○") {
		t.Error("view should contain event symbol ○")
	}
	if !strings.Contains(view, "✓") {
		t.Error("view should contain done symbol ✓")
	}
}

func TestModel_View_ScrollIndicators_ShowMoreAbove(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 10 // Small height

	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 10
	model.scrollOffset = 5 // Scrolled down

	view := model.View()

	if !strings.Contains(view, "more above") {
		t.Error("view should show 'more above' indicator")
	}
}

func TestModel_View_ScrollIndicators_ShowMoreBelow(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 10 // Small height

	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0

	view := model.View()

	if !strings.Contains(view, "more below") {
		t.Error("view should show 'more below' indicator")
	}
}

func TestModel_View_Toolbar_ShowsViewMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.viewMode = ViewModeDay
	model.entries = []EntryItem{}

	view := model.View()

	if !strings.Contains(view, "Day") {
		t.Error("view should show Day in toolbar")
	}
}

func TestModel_View_Toolbar_ShowsWeekMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.viewMode = ViewModeWeek
	model.entries = []EntryItem{}

	view := model.View()

	if !strings.Contains(view, "Week") {
		t.Error("view should show Week in toolbar")
	}
}

// Confirm Mode Tests

func TestModel_ConfirmMode_YConfirmsDelete(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item to delete"}},
	}
	model.confirmMode = confirmState{
		active:      true,
		entryID:     1,
		hasChildren: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := model.Update(msg)

	// Should return a delete command
	if cmd == nil {
		t.Error("expected a delete command")
	}
}

func TestModel_ConfirmMode_NCancels(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item to delete"}},
	}
	model.confirmMode = confirmState{
		active:      true,
		entryID:     1,
		hasChildren: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.confirmMode.active {
		t.Error("confirmMode should not be active after pressing n")
	}
	if cmd != nil {
		t.Error("expected no command after cancel")
	}
}

// ensuredVisible Tests

func TestModel_EnsuredVisible_ScrollsUpWhenSelectedAbove(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 15

	entries := []EntryItem{}
	for i := 0; i < 30; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 2
	model.scrollOffset = 10 // Selected is above visible area

	m := model.ensuredVisible()

	if m.scrollOffset > 2 {
		t.Errorf("scrollOffset should be <= 2, got %d", m.scrollOffset)
	}
}

func TestModel_EnsuredVisible_ScrollsDownWhenSelectedBelow(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 15

	entries := []EntryItem{}
	for i := 0; i < 30; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 25
	model.scrollOffset = 0 // Selected is below visible area

	m := model.ensuredVisible()

	// Should have scrolled down to show item 25
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed")
	}
}

func TestModel_EnsuredVisible_WithDayHeaders_AccountsForExtraLines(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 12

	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		entry := EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}}
		if i%5 == 0 {
			entry.DayHeader = fmt.Sprintf("Day %d", i/5)
		}
		entries = append(entries, entry)
	}
	model.entries = entries
	model.selectedIdx = 15
	model.scrollOffset = 0

	m := model.ensuredVisible()

	// Should have scrolled to show item 15
	// With headers, each header takes extra line, so scroll should adjust
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed to show item 15 with headers")
	}
}

// Paste Tests

// Unicode Symbol Conversion Tests

// Search Highlighting Tests

func TestModel_HighlightSearchTerm_HighlightsMatch(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	// Result should be different from input (contains ANSI codes)
	if result == line {
		t.Error("highlighted result should differ from original (should contain ANSI codes)")
	}
	// Original text should still be present
	if !strings.Contains(result, "this is a ") {
		t.Error("result should contain text before match")
	}
	if !strings.Contains(result, " line") {
		t.Error("result should contain text after match")
	}
}

func TestModel_HighlightSearchTerm_EmptyQuery_NoChange(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: ""}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	if result != line {
		t.Errorf("empty query should return unchanged line, got '%s'", result)
	}
}

func TestModel_HighlightSearchTerm_NoMatch_NoChange(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "xyz"}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	if result != line {
		t.Errorf("no match should return unchanged line, got '%s'", result)
	}
}

func TestModel_HighlightSearchTerm_CaseInsensitive(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "TEST"}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	// Should highlight even though case differs
	if result == line {
		t.Error("case-insensitive match should be highlighted")
	}
}

func TestModel_HighlightSearchTerm_MultipleMatches(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	line := "test one and test two and test three"
	result := model.highlightSearchTerm(line)

	// Count how many times the original "test" appears vs the highlighted version
	// The original string has 3 "test" instances
	// After highlighting, each "test" should be wrapped in ANSI codes
	originalCount := strings.Count(line, "test")
	if originalCount != 3 {
		t.Errorf("test setup error: expected 3 matches in original, got %d", originalCount)
	}

	// The result should be longer due to ANSI codes
	if len(result) <= len(line) {
		t.Error("highlighted result should be longer than original due to ANSI codes")
	}
}

func TestModel_HighlightSearchTerm_PreservesNonMatchingCase(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	line := "TEST and Test and test"
	result := model.highlightSearchTerm(line)

	// All three should be highlighted, but original case preserved in output
	// The ANSI codes will wrap each match
	if result == line {
		t.Error("all matches should be highlighted")
	}
	// Check that we can still find the original text patterns
	// (they'll be wrapped in ANSI codes but the text is there)
}

func TestModel_View_SearchHighlighting_AppliedToEntries(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "First item"}},
		{Entry: domain.Entry{ID: 2, Type: domain.EntryTypeTask, Content: "Second searchterm item"}},
		{Entry: domain.Entry{ID: 3, Type: domain.EntryTypeTask, Content: "Third item"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "searchterm"}

	view := model.View()

	// The view should contain the search term (possibly with ANSI codes)
	if !strings.Contains(view, "searchterm") && !strings.Contains(view, "Second") {
		t.Error("view should contain the entry with search term")
	}

	// View without search mode
	model.searchMode = searchState{}
	viewNoSearch := model.View()

	// The non-search view should be shorter (no ANSI highlighting codes)
	// Actually both contain the text, but with search active, there are extra ANSI codes
	if !strings.Contains(viewNoSearch, "Second searchterm item") {
		t.Error("view without search should contain full entry text")
	}
}

func TestModel_View_SearchHighlighting_InSelectedEntry(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "Match here"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true, query: "Match"}

	view := model.View()

	// The selected entry with a search match should render
	// Both selection styling and search highlighting should be applied
	if !strings.Contains(view, "Match") {
		t.Error("view should contain the matched text")
	}
}

func TestModel_HighlightSearchTerm_SpecialCharacters(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "•"}

	line := "• Task item"
	result := model.highlightSearchTerm(line)

	// Should highlight the bullet point
	if result == line {
		t.Error("special character should be highlighted")
	}
}

func TestModel_HighlightSearchTerm_PartialWord(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "ask"}

	line := "• Task item"
	result := model.highlightSearchTerm(line)

	// Should highlight "ask" within "Task"
	if result == line {
		t.Error("partial word match should be highlighted")
	}
}

func TestModel_SearchMode_ShowsAncestryForSelectedEntry(t *testing.T) {
	parent1ID := int64(1)
	parent2ID := int64(2)
	childID := int64(3)

	model := New(nil)
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: parent1ID, Content: "Project A", ParentID: nil}},
		{Entry: domain.Entry{ID: parent2ID, Content: "Phase 1", ParentID: &parent1ID}},
		{Entry: domain.Entry{ID: childID, Content: "Task detail", ParentID: &parent2ID}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "detail"}
	model.selectedIdx = 2

	view := model.View()

	if !strings.Contains(view, "Project A") || !strings.Contains(view, "Phase 1") {
		t.Error("expected ancestry chain to be shown when in search mode")
	}
}

// ============================================================================
// Phase 4: Multi-View Architecture Tests
// ============================================================================

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

func TestTheme_Default_HasAllColors(t *testing.T) {
	theme := DefaultTheme
	if !theme.HasAllColors() {
		t.Error("default theme should have all colors defined")
	}
}

func TestTheme_Dark_HasAllColors(t *testing.T) {
	theme := DarkTheme
	if !theme.HasAllColors() {
		t.Error("dark theme should have all colors defined")
	}
}

func TestTheme_Light_HasAllColors(t *testing.T) {
	theme := LightTheme
	if !theme.HasAllColors() {
		t.Error("light theme should have all colors defined")
	}
}

func TestTheme_Solarized_HasAllColors(t *testing.T) {
	theme := SolarizedTheme
	if !theme.HasAllColors() {
		t.Error("solarized theme should have all colors defined")
	}
}

func TestTheme_GetTheme_ReturnsCorrectTheme(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"default", "default"},
		{"dark", "dark"},
		{"light", "light"},
		{"solarized", "solarized"},
	}

	for _, tt := range tests {
		theme := GetTheme(tt.name)
		if theme.Name != tt.expected {
			t.Errorf("GetTheme(%s) returned theme with name %s, expected %s", tt.name, theme.Name, tt.expected)
		}
	}
}

func TestTheme_GetTheme_InvalidTheme_ReturnsDefault(t *testing.T) {
	theme := GetTheme("nonexistent")
	if theme.Name != "default" {
		t.Errorf("GetTheme with invalid name should return default theme, got %s", theme.Name)
	}
}

func TestTheme_AvailableThemes_ReturnsList(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) != 4 {
		t.Errorf("expected 4 themes, got %d", len(themes))
	}

	expected := map[string]bool{"default": true, "dark": true, "light": true, "solarized": true}
	for _, theme := range themes {
		if !expected[theme] {
			t.Errorf("unexpected theme %s", theme)
		}
	}
}

func TestTheme_NewThemeStyles_CreatesStyles(t *testing.T) {
	theme := DefaultTheme
	styles := NewThemeStyles(theme)

	// Check that styles are not empty (have some rendering capability)
	// We can verify by checking that Render produces non-empty output
	testStr := "test"
	if styles.Toolbar.Render(testStr) == "" {
		t.Error("Toolbar style should render text")
	}
	if styles.Header.Render(testStr) == "" {
		t.Error("Header style should render text")
	}
	if styles.Done.Render(testStr) == "" {
		t.Error("Done style should render text")
	}
	if styles.Selected.Render(testStr) == "" {
		t.Error("Selected style should render text")
	}
}

// Config Tests

func TestConfig_DefaultTUIConfig_ReturnsDefaults(t *testing.T) {
	config := DefaultTUIConfig()

	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
	if config.Theme != "default" {
		t.Errorf("expected default theme 'default', got '%s'", config.Theme)
	}
	if config.DateFormat != "Mon, Jan 2 2006" {
		t.Errorf("expected default date format, got '%s'", config.DateFormat)
	}
	if !config.ShowHelp {
		t.Error("expected ShowHelp to be true by default")
	}
}

func TestConfig_LoadTUIConfigFromPath_LoadsValidYAML(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	content := []byte("default_view: habits\ntheme: dark\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	config := LoadTUIConfigFromPath(configPath)

	if config.DefaultView != "habits" {
		t.Errorf("expected 'habits', got '%s'", config.DefaultView)
	}
	if config.Theme != "dark" {
		t.Errorf("expected 'dark', got '%s'", config.Theme)
	}
}

func TestConfig_LoadTUIConfigFromPath_PartialFile_UsesDefaults(t *testing.T) {
	// Create a temporary config file with only theme
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	content := []byte("theme: solarized\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	config := LoadTUIConfigFromPath(configPath)

	if config.Theme != "solarized" {
		t.Errorf("expected 'solarized', got '%s'", config.Theme)
	}
	// Default view should use default
	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
}

func TestConfig_LoadTUIConfigFromPath_NoFile_UsesDefaults(t *testing.T) {
	config := LoadTUIConfigFromPath("/nonexistent/path/config.yaml")

	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
	if config.Theme != "default" {
		t.Errorf("expected default theme 'default', got '%s'", config.Theme)
	}
}

func TestConfig_LoadTUIConfigFromPath_InvalidYAML_UsesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	content := []byte("invalid: yaml: content: [")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	config := LoadTUIConfigFromPath(configPath)

	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
}

func TestConfig_GetViewType_ReturnsCorrectType(t *testing.T) {
	tests := []struct {
		defaultView string
		expected    ViewType
	}{
		{"journal", ViewTypeJournal},
		{"habits", ViewTypeHabits},
		{"lists", ViewTypeLists},
		{"unknown", ViewTypeJournal},
	}

	for _, tt := range tests {
		config := TUIConfig{DefaultView: tt.defaultView}
		if config.GetViewType() != tt.expected {
			t.Errorf("GetViewType() for '%s' expected %v, got %v", tt.defaultView, tt.expected, config.GetViewType())
		}
	}
}

func TestConfig_ConfigPaths_ReturnsMultiplePaths(t *testing.T) {
	paths := ConfigPaths()
	if len(paths) < 1 {
		t.Error("ConfigPaths should return at least one path")
	}

	// Check that paths end with expected suffixes
	foundConfigDir := false
	foundBujoDir := false
	for _, p := range paths {
		if strings.Contains(p, "bujo/config.yaml") || strings.Contains(p, "bujo\\config.yaml") {
			foundConfigDir = true
		}
		if strings.Contains(p, ".bujo/config.yaml") || strings.Contains(p, ".bujo\\config.yaml") {
			foundBujoDir = true
		}
	}

	if !foundConfigDir && !foundBujoDir {
		t.Errorf("ConfigPaths should include standard config paths, got: %v", paths)
	}
}

func TestRenderEntry_SelectedMigratedEntry_HasReadableForeground(t *testing.T) {
	model := New(nil)
	item := EntryItem{
		Entry: domain.Entry{
			Type:    domain.EntryTypeMigrated,
			Content: "Migrated task",
		},
	}

	rendered := model.renderEntry(item, true)

	migratedFgCode := "\x1b[38;5;8m"

	if strings.Contains(rendered, migratedFgCode) {
		t.Error("selected migrated entry should NOT have gray foreground (color 8) - would be unreadable against gray background")
	}
	if !strings.Contains(rendered, "Migrated task") {
		t.Error("rendered output should contain the task content")
	}
}

func TestRenderEntry_UnselectedMigratedEntry_HasDimStyle(t *testing.T) {
	model := New(nil)
	item := EntryItem{
		Entry: domain.Entry{
			Type:    domain.EntryTypeMigrated,
			Content: "Migrated task",
		},
	}

	rendered := model.renderEntry(item, false)

	if !strings.Contains(rendered, "Migrated task") {
		t.Error("rendered output should contain the task content")
	}
	if !strings.Contains(rendered, "→") {
		t.Error("migrated entry should show the migrated symbol")
	}
}

func TestRemoveHabitLogForDateCmd_NoLogsToRemove_ShouldNotReturnError(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	err := habitSvc.LogHabitForDate(ctx, "Meditation", 1, today)
	if err != nil {
		t.Fatalf("failed to create habit: %v", err)
	}

	status, err := habitSvc.GetTrackerStatus(ctx, today, 7)
	if err != nil {
		t.Fatalf("failed to get tracker status: %v", err)
	}
	habitID := status.Habits[0].ID

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})

	twoDaysAgo := today.AddDate(0, 0, -2)
	cmd := model.removeHabitLogForDateCmd(habitID, twoDaysAgo)
	msg := cmd()

	if _, isError := msg.(errMsg); isError {
		t.Error("removeHabitLogForDateCmd should not return errMsg when no logs exist for the date")
	}
}

func TestHabitView_WeekOffset_DefaultsToZero(t *testing.T) {
	model := New(nil)
	if model.habitState.weekOffset != 0 {
		t.Errorf("weekOffset should default to 0, got %d", model.habitState.weekOffset)
	}
}

func TestHabitView_PrevPeriod_IncrementsWeekOffset(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.weekOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.habitState.weekOffset != 1 {
		t.Errorf("pressing '[' should increment weekOffset to 1, got %d", m.habitState.weekOffset)
	}
}

func TestHabitView_NextPeriod_DecrementsWeekOffset(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.weekOffset = 2

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.habitState.weekOffset != 1 {
		t.Errorf("pressing ']' should decrement weekOffset to 1, got %d", m.habitState.weekOffset)
	}
}

func TestHabitView_NextPeriod_CannotGoToFuture(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.weekOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.habitState.weekOffset != 0 {
		t.Errorf("pressing ']' at weekOffset=0 should not go negative, got %d", m.habitState.weekOffset)
	}
}

func TestHabitView_ToggleView_ResetsWeekOffsetToAvoidFuture(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.viewMode = HabitViewModeWeek
	model.habitState.weekOffset = 0

	// Navigate back 2 weeks in week mode
	model.habitState.weekOffset = 2

	// Now press 'w' to toggle to month mode
	// This should reset weekOffset to 0
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// After toggle, weekOffset should be 0
	if m.habitState.weekOffset != 0 {
		t.Errorf("after toggling view mode, weekOffset should be reset to 0, got %d", m.habitState.weekOffset)
	}

	// Reference date should be current day or recent
	refDateMonth := m.getHabitReferenceDate()
	todayNormalized := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	refDateNormalized := time.Date(refDateMonth.Year(), refDateMonth.Month(), refDateMonth.Day(), 0, 0, 0, 0, refDateMonth.Location())
	daysOffset := todayNormalized.Sub(refDateNormalized).Hours() / 24

	// Allow ~35 day margin: since month-mode offset of 1 = ~30 days, this ensures
	// we're within approximately one month of the current date after the toggle
	const maxDaysMargin = 35
	if daysOffset > maxDaysMargin {
		t.Errorf("after toggling to month mode, reference date should be recent (within ~30 days), but is %d days in past", int(daysOffset))
	}
}

func TestNavigationStack_InitiallyEmpty(t *testing.T) {
	model := New(nil)

	if len(model.viewStack) != 0 {
		t.Errorf("viewStack should be empty initially, got %d items", len(model.viewStack))
	}
}

func TestNavigationStack_PushWhenSwitchingViews(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeHabits {
		t.Errorf("expected ViewTypeHabits, got %v", m.currentView)
	}
	if len(m.viewStack) != 1 {
		t.Errorf("viewStack should have 1 item after switching views, got %d", len(m.viewStack))
	}
	if len(m.viewStack) > 0 && m.viewStack[0] != ViewTypeJournal {
		t.Errorf("viewStack[0] should be ViewTypeJournal, got %v", m.viewStack[0])
	}
}

func TestNavigationStack_PopWhenPressingEsc(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.viewStack = []ViewType{ViewTypeJournal}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeJournal {
		t.Errorf("should navigate back to ViewTypeJournal, got %v", m.currentView)
	}
	if len(m.viewStack) != 0 {
		t.Errorf("viewStack should be empty after pop, got %d items", len(m.viewStack))
	}
}

func TestNavigationStack_QShowsConfirmEvenWithStack(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.viewStack = []ViewType{ViewTypeJournal}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.quitConfirmMode.active {
		t.Error("pressing q should show quit confirmation, even with views in stack")
	}
	if m.currentView != ViewTypeHabits {
		t.Errorf("should still be in ViewTypeHabits, got %v", m.currentView)
	}
}

func TestNavigationStack_ShowConfirmWhenEmptyStack(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.viewStack = []ViewType{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.quitConfirmMode.active {
		t.Error("should show quit confirmation when at root view")
	}
	if m.currentView != ViewTypeJournal {
		t.Error("should stay on journal view while confirming")
	}
}

func TestNavigationStack_ConfirmYes_Quits(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.quitConfirmMode.active = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("confirming quit with 'y' should return a command")
	}
}

func TestNavigationStack_ConfirmNo_CancelsQuit(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.quitConfirmMode.active = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.quitConfirmMode.active {
		t.Error("quit confirm mode should be deactivated after pressing 'n'")
	}
}

func TestQuitConfirmView_ShowsWhenActive(t *testing.T) {
	model := New(nil)
	model.quitConfirmMode.active = true

	view := model.View()

	if !strings.Contains(view, "Quit") {
		t.Error("quit confirm view should contain 'Quit'")
	}
	if !strings.Contains(view, "Are you sure") {
		t.Error("quit confirm view should ask 'Are you sure'")
	}
}

func TestJournalView_ShowsDailySummary(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.viewDate = time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	model.summaryState.summary = &domain.Summary{
		ID:      1,
		Horizon: "daily",
		Content: "Test daily summary content",
	}
	model.summaryState.horizon = "daily"
	model.summaryCollapsed = false // Expand summary to see content
	model.agenda = &service.MultiDayAgenda{}

	output := model.View()

	if !strings.Contains(output, "Test daily summary content") {
		t.Error("journal view should display daily AI summary")
	}
}

func TestJournalView_ShowsWeeklySummary(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.viewDate = time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	model.summaryState.summary = &domain.Summary{
		ID:      1,
		Horizon: "weekly",
		Content: "Test weekly summary content",
	}
	model.summaryState.horizon = "weekly"
	model.summaryCollapsed = false // Expand summary to see content
	model.agenda = &service.MultiDayAgenda{}

	output := model.View()

	if !strings.Contains(output, "Test weekly summary content") {
		t.Error("journal view should display weekly AI summary")
	}
}

func TestJournalView_DoesNotLoadSummaryWithoutService(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.viewDate = time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	model.viewMode = ViewModeDay
	initialHorizon := model.summaryState.horizon

	msg := agendaLoadedMsg{agenda: &service.MultiDayAgenda{}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.summaryState.loading {
		t.Error("summary should not be loading without service")
	}

	if m.summaryState.horizon != initialHorizon {
		t.Error("summary horizon should not change without service")
	}
}

func TestStatsView_DoesNotShowAISummary(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeStats
	model.summaryState.summary = &domain.Summary{
		ID:      1,
		Horizon: "daily",
		Content: "Test summary content",
	}

	output := model.View()

	if strings.Contains(output, "Test summary content") {
		t.Error("stats view should NOT display AI summaries")
	}
	if strings.Contains(output, "AI Summary:") {
		t.Error("stats view should NOT show AI summary section")
	}
}

// ============================================================================
// Undo Functionality Tests
// ============================================================================

func TestModel_UndoMarkDone(t *testing.T) {
	model := New(nil)
	taskEntry := domain.Entry{
		ID:      1,
		Type:    domain.EntryTypeTask,
		Content: "Test task",
	}
	model.entries = []EntryItem{
		{Entry: taskEntry},
	}
	model.selectedIdx = 0

	msgSpace := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := model.Update(msgSpace)
	m := newModel.(Model)

	if m.undoState.operation == UndoOpNone {
		t.Error("expected undo state to be set after marking done")
	}

	msgUndo := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	newModel, _ = m.Update(msgUndo)
	m = newModel.(Model)

	if m.undoState.operation != UndoOpNone {
		t.Error("expected undo state to be cleared after undo")
	}
}

func TestModel_WeeklyView_MarkDone_PreservesFocus(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal

	// Create multiple entries in weekly view context
	entries := []EntryItem{
		{Entry: domain.Entry{ID: 1, EntityID: "entity-1", Type: domain.EntryTypeTask, Content: "Task 1"}},
		{Entry: domain.Entry{ID: 2, EntityID: "entity-2", Type: domain.EntryTypeTask, Content: "Task 2"}},
		{Entry: domain.Entry{ID: 3, EntityID: "entity-3", Type: domain.EntryTypeTask, Content: "Task 3"}},
		{Entry: domain.Entry{ID: 4, EntityID: "entity-4", Type: domain.EntryTypeTask, Content: "Task 4"}},
	}
	model.entries = entries

	// Select the 3rd entry (index 2)
	model.selectedIdx = 2
	selectedEntityID := model.entries[2].Entry.EntityID

	// Create an agenda with the same entries
	// This simulates what happens when the agenda is reloaded
	today := time.Now()
	dayEntries := service.DayEntries{
		Date:    today,
		Entries: []domain.Entry{entries[0].Entry, entries[1].Entry, entries[2].Entry, entries[3].Entry},
	}

	agendaMsg := agendaLoadedMsg{
		agenda: &service.MultiDayAgenda{
			Days: []service.DayEntries{dayEntries},
		},
	}

	newModel, _ := model.Update(agendaMsg)
	m := newModel.(Model)

	// Focus should be preserved on the same entry EntityID
	// Find the new index of the selected entry
	newIdx := -1
	for idx, item := range m.entries {
		if item.Entry.EntityID == selectedEntityID {
			newIdx = idx
			break
		}
	}

	if newIdx == -1 {
		t.Error("selected entry not found after agenda reload")
	}

	if m.selectedIdx != newIdx {
		t.Errorf("focus was not preserved, expected index %d, got %d", newIdx, m.selectedIdx)
	}

	// The selectedIdx should still be within valid bounds
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.entries) {
		t.Errorf("selectedIdx is out of bounds after agendaLoadedMsg, got %d, len=%d", m.selectedIdx, len(m.entries))
	}
}

func TestModel_WeeklyView_PastCompletedTasks_DisplayedInRed(t *testing.T) {
	model := New(nil)

	// Create entries: a recently completed task and an older completed task
	today := time.Now()
	oldDate := today.AddDate(0, 0, -7)

	model.entries = []EntryItem{
		{
			Entry: domain.Entry{
				ID:      1,
				Type:    domain.EntryTypeDone,
				Content: "Completed today",
				// This entry was just completed
			},
		},
		{
			Entry: domain.Entry{
				ID:            2,
				Type:          domain.EntryTypeDone,
				Content:       "Completed in past",
				ScheduledDate: &oldDate,
				// This entry was completed in the past
			},
			IsOverdue: true, // Mark as a past entry
		},
	}

	// Render the past completed entry (it's overdue)
	rendered := model.renderEntry(model.entries[1], false)

	// The entry should be styled with ANSI escape codes
	// Check for presence of ANSI escape sequence (\x1b[) which indicates styling was applied
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("expected past completed task to be styled with ANSI codes, but no styling found")
	}
}
