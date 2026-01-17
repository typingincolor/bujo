package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var habitMonth bool

var habitCmd = &cobra.Command{
	Use:   "habit",
	Short: "Display habit tracker",
	Long: `Display the habit tracker with streaks, completion rates, and history.

By default shows a 7-day sparkline. Use --month for a calendar view.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		days := 7
		if habitMonth {
			days = 30
		}

		status, err := services.Habit.GetTrackerStatus(cmd.Context(), time.Now(), days)
		if err != nil {
			return fmt.Errorf("failed to get habit status: %w", err)
		}

		if habitMonth {
			fmt.Print(cli.RenderHabitMonth(status))
		} else {
			fmt.Print(cli.RenderHabitTracker(status))
		}
		return nil
	},
}

func init() {
	habitCmd.Flags().BoolVarP(&habitMonth, "month", "m", false, "Show month calendar view")
	rootCmd.AddCommand(habitCmd)
}
