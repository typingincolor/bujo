package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

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
