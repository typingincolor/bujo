package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var workCmd = &cobra.Command{
	Use:   "work <location>",
	Short: "Set today's work location",
	Long: `Set the location context for today.

This location will be displayed in the daily agenda and can be used
for location-based summaries.

Examples:
  bujo work "Home Office"
  bujo work Manchester`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		location := strings.Join(args, " ")

		err := bujoService.SetLocation(cmd.Context(), time.Now(), location)
		if err != nil {
			return fmt.Errorf("failed to set location: %w", err)
		}

		fmt.Fprintf(os.Stderr, "📍 Location set to: %s\n", location)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(workCmd)
}
