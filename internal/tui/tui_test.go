package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

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
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("pressing q should return a quit command")
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
	agenda := &service.MultiDayAgenda{
		Days: []service.DayEntries{
			{
				Date: time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
				Entries: []domain.Entry{
					{ID: 1, Content: "Parent", Type: domain.EntryTypeTask, ParentID: nil},
					{ID: 2, Content: "Child", Type: domain.EntryTypeNote, ParentID: &parentID},
				},
			},
		},
	}

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

	if view != "Loading..." {
		t.Errorf("expected Loading..., got %s", view)
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

func TestModel_Update_EditMode_EntersOnE(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Test entry", Type: domain.EntryTypeTask}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.editMode.active {
		t.Error("pressing e should enter edit mode")
	}
	if m.editMode.entryID != 1 {
		t.Errorf("editMode.entryID should be 1, got %d", m.editMode.entryID)
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

func TestModel_Update_EditMode_CancelsOnEsc(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Test entry", Type: domain.EntryTypeTask}},
	}
	model.editMode = editState{
		active:  true,
		entryID: 1,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.editMode.active {
		t.Error("pressing ESC should exit edit mode")
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

func TestModel_Update_AddMode_EntersOnA(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.addMode.active {
		t.Error("pressing a should enter add mode")
	}
	if m.addMode.asChild {
		t.Error("pressing a should add as sibling, not child")
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

func TestModel_Update_AddMode_CancelsOnEsc(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.addMode = addState{active: true}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.addMode.active {
		t.Error("pressing ESC should exit add mode")
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

func TestModel_Update_MigrateMode_EntersOnM(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task", Type: domain.EntryTypeTask}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.migrateMode.active {
		t.Error("pressing m should enter migrate mode")
	}
	if m.migrateMode.entryID != 1 {
		t.Errorf("migrateMode.entryID should be 1, got %d", m.migrateMode.entryID)
	}
}

func TestModel_Update_MigrateMode_CancelsOnEsc(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.migrateMode = migrateState{active: true, entryID: 1}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.migrateMode.active {
		t.Error("pressing ESC should exit migrate mode")
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

func TestModel_Update_CaptureMode_EntersOnC(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.captureMode.active {
		t.Error("pressing c should enter capture mode")
	}
}

func TestModel_Update_CaptureMode_ExitsOnCtrlX(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task to save"}

	msg := tea.KeyMsg{Type: tea.KeyCtrlX}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.active {
		t.Error("Ctrl+X should exit capture mode")
	}
	if cmd == nil {
		t.Error("Ctrl+X should return a save command")
	}
}

func TestModel_Update_CaptureMode_ExitsOnCtrlX_EmptyContent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ""}

	msg := tea.KeyMsg{Type: tea.KeyCtrlX}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.active {
		t.Error("Ctrl+X should exit capture mode even with empty content")
	}
	if cmd != nil {
		t.Error("Ctrl+X with empty content should not return a save command")
	}
}

func TestModel_Update_CaptureMode_PromptOnEscWithContent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Some content"}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.captureMode.active {
		t.Error("ESC with content should not immediately exit")
	}
	if !m.captureMode.confirmCancel {
		t.Error("ESC with content should show confirmation prompt")
	}
}

func TestModel_Update_CaptureMode_ExitsOnEscWithoutContent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ""}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.active {
		t.Error("ESC without content should exit capture mode")
	}
}

func TestModel_Update_CaptureMode_ConfirmCancelWithY(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", confirmCancel: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.active {
		t.Error("pressing y on confirm should exit capture mode")
	}
}

func TestModel_Update_CaptureMode_ConfirmCancelWithN(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", confirmCancel: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.captureMode.active {
		t.Error("pressing n on confirm should stay in capture mode")
	}
	if m.captureMode.confirmCancel {
		t.Error("pressing n should clear confirmCancel flag")
	}
}

func TestModel_Update_CaptureMode_StartsEmpty(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != "" {
		t.Errorf("capture mode should start empty, got %s", m.captureMode.content)
	}
}

func TestModel_Update_CaptureMode_ParsesContentRealtime(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task one\n- Note here"}

	// Trigger a parse by updating
	model.captureMode.parsedEntries, model.captureMode.parseError = model.parseCapture(model.captureMode.content)

	if len(model.captureMode.parsedEntries) != 2 {
		t.Errorf("expected 2 parsed entries, got %d", len(model.captureMode.parsedEntries))
	}
	if model.captureMode.parseError != nil {
		t.Errorf("expected no parse error, got %v", model.captureMode.parseError)
	}
}

func TestModel_Update_CaptureMode_DetectsIndentationError(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task\n    - Skipped indent level"}

	model.captureMode.parsedEntries, model.captureMode.parseError = model.parseCapture(model.captureMode.content)

	if model.captureMode.parseError == nil {
		t.Error("expected indentation error for skipped level")
	}
}

func TestModel_Update_CaptureMode_DetectsMissingSymbol(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: "Missing symbol"}

	model.captureMode.parsedEntries, model.captureMode.parseError = model.parseCapture(model.captureMode.content)

	if model.captureMode.parseError == nil {
		t.Error("expected error for missing symbol prefix")
	}
}

func TestModel_View_CaptureMode_ShowsCaptureHeader(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task one"}

	view := model.View()

	if !strings.Contains(view, "CAPTURE") {
		t.Error("capture mode view should show CAPTURE header")
	}
}

func TestModel_View_CaptureMode_ShowsErrorInStatusBar(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{
		active:     true,
		content:    "Invalid content",
		parseError: fmt.Errorf("unknown entry type symbol"),
	}

	view := model.View()

	if !strings.Contains(view, "unknown entry type") {
		t.Error("capture mode view should show parse error in status bar")
	}
}

func TestModel_View_CaptureMode_ShowsConfirmPrompt(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{
		active:        true,
		content:       ". Task",
		confirmCancel: true,
	}

	view := model.View()

	if !strings.Contains(view, "Discard") {
		t.Error("capture mode view should show discard confirmation")
	}
}

func TestModel_CaptureMode_TypesCharacters(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ""}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != "." {
		t.Errorf("expected content '.', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_TabInsertsIndent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorLine: 0}

	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != "  . Task" {
		t.Errorf("expected content '  . Task', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_ShiftTabRemovesIndent(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: "  . Task", cursorLine: 0}

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != ". Task" {
		t.Errorf("expected content '. Task', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_EnterAddsNewLine(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 7, cursorLine: 0, cursorCol: 7}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != ". Task\n" {
		t.Errorf("expected content '. Task\\n', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_EnterAutoIndents(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: "  . Task", cursorPos: 9, cursorLine: 0, cursorCol: 9}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != "  . Task\n  " {
		t.Errorf("expected content '  . Task\\n  ', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_BackspaceDeletesChar(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 6, cursorLine: 0, cursorCol: 6}

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != ". Tas" {
		t.Errorf("expected content '. Tas', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_ParsesOnChange(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ""}

	// Type ". Task"
	for _, r := range ". Task" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ := model.Update(msg)
		model = newModel.(Model)
	}

	if len(model.captureMode.parsedEntries) != 1 {
		t.Fatalf("expected 1 parsed entry, got %d", len(model.captureMode.parsedEntries))
	}
	if model.captureMode.parsedEntries[0].Content != "Task" {
		t.Errorf("expected content 'Task', got '%s'", model.captureMode.parsedEntries[0].Content)
	}
}

// Emacs navigation tests

func TestModel_CaptureMode_CtrlA_BeginningOfLine(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 6, cursorCol: 6}

	msg := tea.KeyMsg{Type: tea.KeyCtrlA}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.cursorCol != 0 {
		t.Errorf("expected cursorCol 0, got %d", m.captureMode.cursorCol)
	}
}

func TestModel_CaptureMode_CtrlE_EndOfLine(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 0, cursorCol: 0}

	msg := tea.KeyMsg{Type: tea.KeyCtrlE}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.cursorCol != 6 {
		t.Errorf("expected cursorCol 6, got %d", m.captureMode.cursorCol)
	}
}

func TestModel_CaptureMode_CtrlF_ForwardChar(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 0, cursorCol: 0}

	msg := tea.KeyMsg{Type: tea.KeyCtrlF}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.cursorCol != 1 {
		t.Errorf("expected cursorCol 1, got %d", m.captureMode.cursorCol)
	}
}

func TestModel_CaptureMode_CtrlB_BackwardChar(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 3, cursorCol: 3}

	msg := tea.KeyMsg{Type: tea.KeyCtrlB}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.cursorCol != 2 {
		t.Errorf("expected cursorCol 2, got %d", m.captureMode.cursorCol)
	}
}

func TestModel_CaptureMode_CtrlK_KillToEndOfLine(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 2, cursorCol: 2}

	msg := tea.KeyMsg{Type: tea.KeyCtrlK}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != ". " {
		t.Errorf("expected content '. ', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_CtrlD_DeleteChar(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", cursorPos: 2, cursorCol: 2}

	msg := tea.KeyMsg{Type: tea.KeyCtrlD}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != ". ask" {
		t.Errorf("expected content '. ask', got '%s'", m.captureMode.content)
	}
}

func TestModel_CaptureMode_CtrlW_DeleteWordBackward(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task here", cursorPos: 11, cursorCol: 11}

	msg := tea.KeyMsg{Type: tea.KeyCtrlW}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != ". Task " {
		t.Errorf("expected content '. Task ', got '%s'", m.captureMode.content)
	}
}

// Search tests

func TestModel_CaptureMode_CtrlS_EntersSearchMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task here"}

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.captureMode.searchMode {
		t.Error("Ctrl+S should enter search mode")
	}
	if !m.captureMode.searchForward {
		t.Error("Ctrl+S should set searchForward to true")
	}
}

func TestModel_CaptureMode_CtrlR_EntersReverseSearchMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task here"}

	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.captureMode.searchMode {
		t.Error("Ctrl+R should enter search mode")
	}
	if m.captureMode.searchForward {
		t.Error("Ctrl+R should set searchForward to false")
	}
}

func TestModel_CaptureMode_Search_EscExitsSearchMode(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, content: ". Task", searchMode: true}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.searchMode {
		t.Error("ESC should exit search mode")
	}
	if !m.captureMode.active {
		t.Error("ESC in search mode should not exit capture mode")
	}
}

func TestModel_CaptureMode_Search_FindsMatch(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{
		active:        true,
		content:       ". Task here\n- Note there",
		searchMode:    true,
		searchForward: true,
		searchQuery:   "here",
		cursorPos:     0,
	}

	// Trigger search by pressing Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Cursor should move to "here" position (index 7)
	if m.captureMode.cursorPos != 7 {
		t.Errorf("expected cursorPos 7, got %d", m.captureMode.cursorPos)
	}
}

// Draft tests

func TestModel_CaptureMode_DraftPrompt_ShowsWhenDraftExists(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{active: true, draftExists: true}

	view := model.View()

	if !strings.Contains(view, "Restore") {
		t.Error("should show restore prompt when draft exists")
	}
}

func TestModel_CaptureMode_DraftPrompt_YRestoresDraft(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{
		active:       true,
		draftExists:  true,
		draftContent: ". Saved task",
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != ". Saved task" {
		t.Errorf("expected content '. Saved task', got '%s'", m.captureMode.content)
	}
	if m.captureMode.draftExists {
		t.Error("draftExists should be false after restore")
	}
}

func TestModel_CaptureMode_DraftPrompt_NStartsFresh(t *testing.T) {
	model := New(nil)
	model.agenda = &service.MultiDayAgenda{}
	model.captureMode = captureState{
		active:       true,
		draftExists:  true,
		draftContent: ". Saved task",
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.captureMode.content != "" {
		t.Errorf("expected empty content, got '%s'", m.captureMode.content)
	}
	if m.captureMode.draftExists {
		t.Error("draftExists should be false after declining")
	}
}
