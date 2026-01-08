package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show entries for the next 7 days",
	Long: `Show entries for today and the next 6 days.

This is a shortcut for viewing your upcoming week.

Examples:
  bujo next`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()
		today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
		endDate := today.AddDate(0, 0, 6)

		agenda, err := bujoService.GetMultiDayAgenda(cmd.Context(), today, endDate)
		if err != nil {
			return fmt.Errorf("failed to get agenda: %w", err)
		}

		fmt.Print(cli.RenderMultiDayAgenda(agenda, today))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nextCmd)
}
