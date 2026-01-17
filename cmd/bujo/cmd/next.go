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
	Long: `Show entries for the upcoming 7 days (starting from tomorrow).

This is a shortcut for viewing your upcoming week.
Use 'bujo today' to see today's entries.

Examples:
  bujo next`,
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		tomorrow := today.AddDate(0, 0, 1)
		endDate := tomorrow.AddDate(0, 0, 6)

		agenda, err := bujoService.GetMultiDayAgenda(cmd.Context(), tomorrow, endDate)
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
