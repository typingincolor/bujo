package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var moodCmd = &cobra.Command{
	Use:   "mood",
	Short: "Manage daily mood",
	Long: `Manage daily mood tracking.

When called without a subcommand, shows today's mood.

Examples:
  bujo mood                         # Show today's mood
  bujo mood set happy               # Set today's mood
  bujo mood set tired -d yesterday  # Set mood for a past date
  bujo mood show --from "last week"     # View mood history
  bujo mood clear -d yesterday      # Clear a day's mood`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()
		mood, err := bujoService.GetMood(cmd.Context(), today)
		if err != nil {
			return fmt.Errorf("failed to get mood: %w", err)
		}

		if mood == nil {
			fmt.Println("No mood set for today")
		} else {
			fmt.Printf("Today's mood: %s\n", *mood)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moodCmd)
}
