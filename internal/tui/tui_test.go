package tui

import (
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
	if result[0].DayHeader != "OVERDUE" {
		t.Errorf("expected OVERDUE header, got %s", result[0].DayHeader)
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
