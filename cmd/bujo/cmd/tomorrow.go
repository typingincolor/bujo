package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var tomorrowCmd = &cobra.Command{
	Use:   "tomorrow",
	Short: "Show tomorrow's entries",
	Long: `Show entries scheduled for tomorrow.

This is a shortcut for: bujo ls --from tomorrow --to tomorrow

Examples:
  bujo tomorrow`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tomorrow := time.Now().AddDate(0, 0, 1)
		tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

		agenda, err := bujoService.GetMultiDayAgenda(cmd.Context(), tomorrow, tomorrow)
		if err != nil {
			return fmt.Errorf("failed to get agenda: %w", err)
		}

		fmt.Print(cli.RenderMultiDayAgenda(agenda))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tomorrowCmd)
}
