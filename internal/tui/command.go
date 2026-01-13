package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Command struct {
	Name        string
	Description string
	Keybinding  string
	Action      func(m Model) (Model, tea.Cmd)
}

type CommandRegistry struct {
	commands []Command
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: []Command{},
	}
}

func (r *CommandRegistry) Register(cmd Command) {
	r.commands = append(r.commands, cmd)
}

func (r *CommandRegistry) All() []Command {
	return r.commands
}

func (r *CommandRegistry) Filter(query string) []Command {
	if query == "" {
		return r.commands
	}

	query = strings.ToLower(query)
	var filtered []Command

	for _, cmd := range r.commands {
		if r.fuzzyMatch(strings.ToLower(cmd.Name), query) ||
			r.fuzzyMatch(strings.ToLower(cmd.Description), query) {
			filtered = append(filtered, cmd)
		}
	}

	return filtered
}

func (r *CommandRegistry) fuzzyMatch(s, pattern string) bool {
	if strings.Contains(s, pattern) {
		return true
	}

	patternIdx := 0
	for i := 0; i < len(s) && patternIdx < len(pattern); i++ {
		if s[i] == pattern[patternIdx] {
			patternIdx++
		}
	}
	return patternIdx == len(pattern)
}

func DefaultCommands() *CommandRegistry {
	registry := NewCommandRegistry()

	registry.Register(Command{
		Name:        "Switch to Journal",
		Description: "View the journal entries",
		Keybinding:  "1",
		Action: func(m Model) (Model, tea.Cmd) {
			m.currentView = ViewTypeJournal
			return m, m.loadAgendaCmd()
		},
	})

	registry.Register(Command{
		Name:        "Switch to Habits",
		Description: "View and log habits",
		Keybinding:  "2",
		Action: func(m Model) (Model, tea.Cmd) {
			m.currentView = ViewTypeHabits
			return m, m.loadHabitsCmd()
		},
	})

	registry.Register(Command{
		Name:        "Switch to Lists",
		Description: "View and manage lists",
		Keybinding:  "3",
		Action: func(m Model) (Model, tea.Cmd) {
			m.currentView = ViewTypeLists
			return m, m.loadListsCmd()
		},
	})

	registry.Register(Command{
		Name:        "Toggle Day/Week View",
		Description: "Switch between day and week view",
		Keybinding:  "w",
		Action: func(m Model) (Model, tea.Cmd) {
			if m.viewMode == ViewModeDay {
				m.viewMode = ViewModeWeek
			} else {
				m.viewMode = ViewModeDay
			}
			return m, m.loadAgendaCmd()
		},
	})

	registry.Register(Command{
		Name:        "Capture Mode",
		Description: "Enter capture mode for quick entry",
		Keybinding:  "c",
		Action: func(m Model) (Model, tea.Cmd) {
			m.captureMode.active = true
			return m, nil
		},
	})

	registry.Register(Command{
		Name:        "Go to Date",
		Description: "Navigate to a specific date",
		Keybinding:  "/",
		Action: func(m Model) (Model, tea.Cmd) {
			m.gotoMode.active = true
			return m, nil
		},
	})

	registry.Register(Command{
		Name:        "Set Location",
		Description: "Set location for current day",
		Keybinding:  "",
		Action: func(m Model) (Model, tea.Cmd) {
			ti := textinput.New()
			ti.Placeholder = "Enter location..."
			ti.Focus()
			ti.CharLimit = 100
			ti.Width = m.width - 10
			m.setLocationMode = setLocationState{
				active: true,
				date:   m.viewDate,
				input:  ti,
			}
			return m, nil
		},
	})

	registry.Register(Command{
		Name:        "Quit",
		Description: "Exit the application",
		Keybinding:  "q",
		Action: func(m Model) (Model, tea.Cmd) {
			return m, tea.Quit
		},
	})

	return registry
}
