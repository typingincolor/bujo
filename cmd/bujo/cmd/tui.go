package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Long:  `Launch an interactive terminal UI for viewing and managing journal entries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		model := tui.NewWithConfig(tui.Config{
			BujoService:  bujoService,
			HabitService: habitService,
			ListService:  listService,
			GoalService:  goalService,
		})
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
