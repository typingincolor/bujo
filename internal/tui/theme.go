package tui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name            string
	Primary         lipgloss.Color
	Secondary       lipgloss.Color
	Background      lipgloss.Color
	Foreground      lipgloss.Color
	Muted           lipgloss.Color
	Accent          lipgloss.Color
	Success         lipgloss.Color
	Warning         lipgloss.Color
	Error           lipgloss.Color
	Done            lipgloss.Color
	Migrated        lipgloss.Color
	Selection       lipgloss.Color
	SelectionFg     lipgloss.Color
}

var DefaultTheme = Theme{
	Name:        "default",
	Primary:     lipgloss.Color("36"),
	Secondary:   lipgloss.Color("105"),
	Background:  lipgloss.Color(""),
	Foreground:  lipgloss.Color(""),
	Muted:       lipgloss.Color("240"),
	Accent:      lipgloss.Color("212"),
	Success:     lipgloss.Color("82"),
	Warning:     lipgloss.Color("214"),
	Error:       lipgloss.Color("196"),
	Done:        lipgloss.Color("242"),
	Migrated:    lipgloss.Color("105"),
	Selection:   lipgloss.Color("62"),
	SelectionFg: lipgloss.Color("255"),
}

var DarkTheme = Theme{
	Name:        "dark",
	Primary:     lipgloss.Color("39"),
	Secondary:   lipgloss.Color("105"),
	Background:  lipgloss.Color("235"),
	Foreground:  lipgloss.Color("252"),
	Muted:       lipgloss.Color("245"),
	Accent:      lipgloss.Color("219"),
	Success:     lipgloss.Color("78"),
	Warning:     lipgloss.Color("220"),
	Error:       lipgloss.Color("203"),
	Done:        lipgloss.Color("245"),
	Migrated:    lipgloss.Color("147"),
	Selection:   lipgloss.Color("24"),
	SelectionFg: lipgloss.Color("255"),
}

var LightTheme = Theme{
	Name:        "light",
	Primary:     lipgloss.Color("25"),
	Secondary:   lipgloss.Color("54"),
	Background:  lipgloss.Color("255"),
	Foreground:  lipgloss.Color("232"),
	Muted:       lipgloss.Color("245"),
	Accent:      lipgloss.Color("127"),
	Success:     lipgloss.Color("28"),
	Warning:     lipgloss.Color("172"),
	Error:       lipgloss.Color("160"),
	Done:        lipgloss.Color("247"),
	Migrated:    lipgloss.Color("61"),
	Selection:   lipgloss.Color("153"),
	SelectionFg: lipgloss.Color("232"),
}

var SolarizedTheme = Theme{
	Name:        "solarized",
	Primary:     lipgloss.Color("37"),
	Secondary:   lipgloss.Color("136"),
	Background:  lipgloss.Color("234"),
	Foreground:  lipgloss.Color("187"),
	Muted:       lipgloss.Color("244"),
	Accent:      lipgloss.Color("166"),
	Success:     lipgloss.Color("64"),
	Warning:     lipgloss.Color("136"),
	Error:       lipgloss.Color("160"),
	Done:        lipgloss.Color("240"),
	Migrated:    lipgloss.Color("33"),
	Selection:   lipgloss.Color("240"),
	SelectionFg: lipgloss.Color("187"),
}

var themes = map[string]Theme{
	"default":   DefaultTheme,
	"dark":      DarkTheme,
	"light":     LightTheme,
	"solarized": SolarizedTheme,
}

func GetTheme(name string) Theme {
	if theme, ok := themes[name]; ok {
		return theme
	}
	return DefaultTheme
}

func AvailableThemes() []string {
	return []string{"default", "dark", "light", "solarized"}
}

func (t Theme) HasAllColors() bool {
	return t.Primary != "" &&
		t.Secondary != "" &&
		t.Muted != "" &&
		t.Accent != "" &&
		t.Success != "" &&
		t.Warning != "" &&
		t.Error != "" &&
		t.Done != "" &&
		t.Migrated != "" &&
		t.Selection != "" &&
		t.SelectionFg != ""
}

type ThemeStyles struct {
	Toolbar   lipgloss.Style
	Header    lipgloss.Style
	Entry     lipgloss.Style
	Done      lipgloss.Style
	Migrated  lipgloss.Style
	Overdue   lipgloss.Style
	Selected  lipgloss.Style
	Help      lipgloss.Style
	Confirm   lipgloss.Style
	Error     lipgloss.Style
	ErrorTitle lipgloss.Style
}

func NewThemeStyles(theme Theme) ThemeStyles {
	return ThemeStyles{
		Toolbar: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Accent),
		Entry: lipgloss.NewStyle().
			Foreground(theme.Foreground),
		Done: lipgloss.NewStyle().
			Foreground(theme.Done).
			Strikethrough(true),
		Migrated: lipgloss.NewStyle().
			Foreground(theme.Migrated),
		Overdue: lipgloss.NewStyle().
			Foreground(theme.Warning),
		Selected: lipgloss.NewStyle().
			Background(theme.Selection).
			Foreground(theme.SelectionFg),
		Help: lipgloss.NewStyle().
			Foreground(theme.Muted),
		Confirm: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary).
			Padding(1, 2),
		Error: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Error).
			Padding(1, 2),
		ErrorTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Error),
	}
}
