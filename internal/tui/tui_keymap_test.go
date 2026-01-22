package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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

func TestDefaultKeyMap_MigrateKey(t *testing.T) {
	km := DefaultKeyMap()

	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}}, km.Migrate) {
		t.Error("Migrate key should be bound to '>'")
	}

	if km.Migrate.Help().Key != ">" {
		t.Errorf("Migrate help key should be '>', got %s", km.Migrate.Help().Key)
	}
}

func TestDefaultKeyMap_FullHelp_IncludesAllActions(t *testing.T) {
	km := DefaultKeyMap()
	help := km.FullHelp()

	requiredKeys := []string{
		"t", // Retype (change type)
		"L", // MoveToList
		"X", // UncancelEntry
		"u", // Undo
		"R", // Answer
	}

	allHelpKeys := make(map[string]bool)
	for _, group := range help {
		for _, binding := range group {
			allHelpKeys[binding.Help().Key] = true
		}
	}

	for _, required := range requiredKeys {
		if !allHelpKeys[required] {
			t.Errorf("FullHelp should include key '%s'", required)
		}
	}
}

func TestDefaultKeyMap_ViewKeyBindings(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name    string
		key     string
		binding key.Binding
	}{
		{"Journal", "1", km.ViewJournal},
		{"Review", "2", km.ViewReview},
		{"Pending Tasks", "3", km.ViewPendingTasks},
		{"Questions", "4", km.ViewQuestions},
		{"Habits", "5", km.ViewHabits},
		{"Lists", "6", km.ViewLists},
		{"Goals", "7", km.ViewGoals},
		{"Search", "8", km.ViewSearch},
		{"Stats", "9", km.ViewStats},
		{"Settings", "0", km.ViewSettings},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if !key.Matches(msg, tt.binding) {
				t.Errorf("%s view should be bound to key '%s'", tt.name, tt.key)
			}
		})
	}
}

func TestDefaultKeyMap_GotoTodayKey(t *testing.T) {
	km := DefaultKeyMap()

	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}}, km.GotoToday) {
		t.Error("GotoToday key should be bound to 'T'")
	}

	if km.GotoToday.Help().Key != "T" {
		t.Errorf("GotoToday help key should be 'T', got %s", km.GotoToday.Help().Key)
	}
}

func TestModel_GotoToday_ResetsViewDateToToday(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.viewMode = ViewModeDay
	// Set view date to 5 days ago
	model.viewDate = time.Now().AddDate(0, 0, -5)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}}
	m, _ := model.Update(msg)
	updated := m.(Model)

	today := time.Now()
	if updated.viewDate.Year() != today.Year() ||
		updated.viewDate.Month() != today.Month() ||
		updated.viewDate.Day() != today.Day() {
		t.Errorf("GotoToday should reset viewDate to today, got %v", updated.viewDate)
	}
}

func TestModel_GotoToday_WorksInReviewView(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeReview
	model.viewMode = ViewModeWeek
	// Set view date to 2 weeks ago
	model.viewDate = time.Now().AddDate(0, 0, -14)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}}
	m, _ := model.Update(msg)
	updated := m.(Model)

	today := time.Now()
	if updated.viewDate.Year() != today.Year() ||
		updated.viewDate.Month() != today.Month() ||
		updated.viewDate.Day() != today.Day() {
		t.Errorf("GotoToday should reset viewDate to today in Review view, got %v", updated.viewDate)
	}
}
