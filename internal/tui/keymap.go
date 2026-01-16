package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up             key.Binding
	Down           key.Binding
	Top            key.Binding
	Bottom         key.Binding
	Done           key.Binding
	Answer         key.Binding
	CancelEntry    key.Binding
	Delete         key.Binding
	Edit           key.Binding
	Add            key.Binding
	AddChild       key.Binding
	AddRoot        key.Binding
	Migrate        key.Binding
	MigrateToGoal  key.Binding
	MoveListItem   key.Binding
	MoveToList     key.Binding
	Priority       key.Binding
	ToggleView     key.Binding
	GotoDate       key.Binding
	Capture        key.Binding
	Confirm        key.Binding
	Cancel         key.Binding
	UncancelEntry  key.Binding
	Retype         key.Binding
	Undo           key.Binding
	Quit           key.Binding
	Back           key.Binding
	Help           key.Binding
	ViewJournal    key.Binding
	ViewHabits     key.Binding
	ViewLists      key.Binding
	ViewSearch     key.Binding
	ViewStats      key.Binding
	ViewGoals      key.Binding
	ViewSettings   key.Binding
	CommandPalette key.Binding
	LogHabit       key.Binding
	RemoveHabitLog key.Binding
	DayLeft        key.Binding
	DayRight       key.Binding
	PrevPeriod     key.Binding
	NextPeriod     key.Binding
	ExpandAll      key.Binding
	CollapseAll    key.Binding
	OpenURL        key.Binding
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
		Answer: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "answer"),
		),
		CancelEntry: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "cancel"),
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
		MigrateToGoal: key.NewBinding(
			key.WithKeys("M"),
			key.WithHelp("M", "to goal"),
		),
		MoveListItem: key.NewBinding(
			key.WithKeys("M"),
			key.WithHelp("M", "move"),
		),
		MoveToList: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "to list"),
		),
		Priority: key.NewBinding(
			key.WithKeys("!"),
			key.WithHelp("!", "priority"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "view"),
		),
		GotoDate: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "go to date"),
		),
		Capture: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "capture"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "N", "esc"),
			key.WithHelp("n/esc", "cancel"),
		),
		UncancelEntry: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("X", "uncancel"),
		),
		Retype: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "change type"),
		),
		Undo: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "undo"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		ViewJournal: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "journal"),
		),
		ViewHabits: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "habits"),
		),
		ViewLists: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "lists"),
		),
		ViewSearch: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "search"),
		),
		ViewStats: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "stats"),
		),
		ViewGoals: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "goals"),
		),
		ViewSettings: key.NewBinding(
			key.WithKeys("7"),
			key.WithHelp("7", "settings"),
		),
		CommandPalette: key.NewBinding(
			key.WithKeys("ctrl+p", ":"),
			key.WithHelp("ctrl+p/:", "commands"),
		),
		LogHabit: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "log"),
		),
		RemoveHabitLog: key.NewBinding(
			key.WithKeys("backspace", "delete"),
			key.WithHelp("⌫/del", "remove"),
		),
		DayLeft: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "prev day"),
		),
		DayRight: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "next day"),
		),
		PrevPeriod: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "prev period"),
		),
		NextPeriod: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next period"),
		),
		ExpandAll: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "expand all"),
		),
		CollapseAll: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "collapse all"),
		),
		OpenURL: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open link"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Done, k.CancelEntry, k.Edit, k.Add, k.Capture, k.Delete, k.Quit, k.Help}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Done, k.CancelEntry, k.Edit, k.Add, k.AddChild, k.AddRoot, k.Migrate, k.Priority, k.Capture, k.Delete},
		{k.ToggleView, k.GotoDate, k.Quit, k.Help},
	}
}
