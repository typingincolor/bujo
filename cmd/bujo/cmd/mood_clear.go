package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var moodClearDate string
var moodClearYes bool

var moodClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear mood for a day",
	Long: `Clear the mood for a specific day.

Defaults to today if no date specified.

Examples:
  bujo mood clear
  bujo mood clear --date yesterday
  bujo mood clear -d "last monday"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDate, err := parseDateOrToday(moodClearDate)
		if err != nil {
			return err
		}

		if moodClearDate != "" {
			targetDate, err = confirmDate(moodClearDate, targetDate, moodClearYes)
			if err != nil {
				return err
			}
		}

		err = services.Bujo.ClearMood(cmd.Context(), targetDate)
		if err != nil {
			return fmt.Errorf("failed to clear mood: %w", err)
		}

		if moodClearDate == "" {
			fmt.Fprintf(os.Stderr, "✓ Mood cleared for today\n")
		} else {
			fmt.Fprintf(os.Stderr, "✓ Mood cleared for %s\n", targetDate.Format("Jan 2, 2006"))
		}
		return nil
	},
}

func init() {
	moodClearCmd.Flags().StringVarP(&moodClearDate, "date", "d", "", "Date to clear mood for (e.g., 'yesterday', '2026-01-05')")
	moodClearCmd.Flags().BoolVarP(&moodClearYes, "yes", "y", false, "Skip date confirmation prompt")
	moodCmd.AddCommand(moodClearCmd)
}
