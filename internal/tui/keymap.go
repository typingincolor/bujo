package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Done       key.Binding
	Delete     key.Binding
	Edit       key.Binding
	Add        key.Binding
	AddChild   key.Binding
	AddRoot    key.Binding
	Migrate    key.Binding
	ToggleView key.Binding
	GotoDate   key.Binding
	Confirm    key.Binding
	Cancel     key.Binding
	Quit       key.Binding
	Help       key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Done: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "done"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		AddChild: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "add child"),
		),
		AddRoot: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "add root"),
		),
		Migrate: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "migrate"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "day/week"),
		),
		GotoDate: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "go to date"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "N", "esc"),
			key.WithHelp("n/esc", "cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Done, k.Edit, k.Add, k.Migrate, k.Delete, k.Quit, k.Help}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Done, k.Edit, k.Add, k.AddChild, k.AddRoot, k.Migrate, k.Delete},
		{k.ToggleView, k.GotoDate, k.Quit, k.Help},
	}
}
