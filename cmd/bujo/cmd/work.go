package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Manage work locations",
	Long: `Manage work locations for days.

When called without a subcommand, shows today's location.

Examples:
  bujo work                         # Show today's location
  bujo work set "Home Office"       # Set today's location
  bujo work set "Office" -d monday  # Set location for a past date
  bujo work show --from "last week"     # View location history
  bujo work clear -d yesterday      # Clear a day's location`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()
		loc, err := services.Bujo.GetLocation(cmd.Context(), today)
		if err != nil {
			return fmt.Errorf("failed to get location: %w", err)
		}

		if loc == nil {
			fmt.Println("No location set for today")
		} else {
			fmt.Printf("Today's location: %s\n", *loc)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(workCmd)
}
