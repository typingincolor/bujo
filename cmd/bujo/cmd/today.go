package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Display today's entries",
	Long:  `Display today's entries, including overdue tasks, current location, and monthly goals.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		agenda, err := services.Bujo.GetDailyAgenda(cmd.Context(), now)
		if err != nil {
			return fmt.Errorf("failed to get agenda: %w", err)
		}

		fmt.Print(cli.RenderDailyAgenda(agenda))

		currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		goals, err := services.Goal.GetGoalsForMonth(cmd.Context(), currentMonth)
		if err != nil {
			return fmt.Errorf("failed to get goals: %w", err)
		}

		fmt.Print(cli.RenderGoalsSection(goals, currentMonth))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
