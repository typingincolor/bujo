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
	Long:  `Display today's entries, including overdue tasks and the current location.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		agenda, err := bujoService.GetDailyAgenda(cmd.Context(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to get agenda: %w", err)
		}

		fmt.Print(cli.RenderDailyAgenda(agenda))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
