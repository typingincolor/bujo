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
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		days, err := bujoService.GetDayEntries(cmd.Context(), todayStart, todayStart)
		if err != nil {
			return fmt.Errorf("failed to get entries: %w", err)
		}

		overdue, err := bujoService.GetOverdue(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get overdue: %w", err)
		}

		fmt.Print(cli.RenderDaysWithOverdue(days, overdue, todayStart))

		currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		goals, err := goalService.GetGoalsForMonth(cmd.Context(), currentMonth)
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
