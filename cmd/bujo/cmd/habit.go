package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var habitCmd = &cobra.Command{
	Use:   "habit",
	Short: "Display habit tracker",
	Long:  `Display the habit tracker with streaks, completion rates, and a 7-day sparkline.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := habitService.GetTrackerStatus(cmd.Context(), time.Now(), 7)
		if err != nil {
			return fmt.Errorf("failed to get habit status: %w", err)
		}

		fmt.Print(cli.RenderHabitTracker(status))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(habitCmd)
}
