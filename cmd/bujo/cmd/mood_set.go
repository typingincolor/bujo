package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var moodSetDate string
var moodSetYes bool

var moodSetCmd = &cobra.Command{
	Use:   "set <mood>",
	Short: "Set mood for a day",
	Long: `Set the mood for a day.

Defaults to today if no date specified.

Examples:
  bujo mood set happy
  bujo mood set "tired but productive"
  bujo mood set energetic --date yesterday
  bujo mood set focused -d "last monday"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mood := strings.Join(args, " ")

		targetDate, err := parseDateOrToday(moodSetDate)
		if err != nil {
			return err
		}

		if moodSetDate != "" {
			targetDate, err = confirmDate(moodSetDate, targetDate, moodSetYes)
			if err != nil {
				return err
			}
		}

		err = services.Bujo.SetMood(cmd.Context(), targetDate, mood)
		if err != nil {
			return fmt.Errorf("failed to set mood: %w", err)
		}

		if moodSetDate == "" {
			fmt.Fprintf(os.Stderr, "✓ Mood set to: %s\n", mood)
		} else {
			fmt.Fprintf(os.Stderr, "✓ Mood for %s set to: %s\n", targetDate.Format("Jan 2, 2006"), mood)
		}
		return nil
	},
}

func init() {
	moodSetCmd.Flags().StringVarP(&moodSetDate, "date", "d", "", "Date to set mood for (e.g., 'yesterday', '2026-01-05')")
	moodSetCmd.Flags().BoolVarP(&moodSetYes, "yes", "y", false, "Skip date confirmation prompt")
	moodCmd.AddCommand(moodSetCmd)
}
