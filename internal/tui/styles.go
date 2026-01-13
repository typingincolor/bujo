package tui

import "github.com/charmbracelet/lipgloss"

var (
	CyanColor   = lipgloss.Color("6")
	GreenColor  = lipgloss.Color("2")
	RedColor    = lipgloss.Color("1")
	YellowColor = lipgloss.Color("3")
	DimColor    = lipgloss.Color("8")

	TitleStyle = lipgloss.NewStyle().
			Foreground(CyanColor).
			Bold(true)

	DateHeaderStyle = lipgloss.NewStyle().
			Foreground(CyanColor).
			Bold(true)

	OverdueHeaderStyle = lipgloss.NewStyle().
				Foreground(RedColor).
				Bold(true)

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15"))

	DoneStyle = lipgloss.NewStyle().
			Foreground(GreenColor)

	OverdueStyle = lipgloss.NewStyle().
			Foreground(RedColor)

	MigratedStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	CancelledStyle = lipgloss.NewStyle().
			Foreground(DimColor).
			Strikethrough(true)

	IDStyle = lipgloss.NewStyle().
		Foreground(DimColor)

	HelpStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	ConfirmStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(YellowColor).
			Padding(1, 2)

	LocationStyle = lipgloss.NewStyle().
			Foreground(YellowColor)

	ToolbarStyle = lipgloss.NewStyle().
			Foreground(CyanColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(RedColor).
			Padding(1, 2)

	ErrorTitleStyle = lipgloss.NewStyle().
			Foreground(RedColor).
			Bold(true)

	SearchHighlightStyle = lipgloss.NewStyle().
				Background(YellowColor).
				Foreground(lipgloss.Color("0"))

	HabitSelectedStyle = lipgloss.NewStyle().
				Reverse(true)
)
