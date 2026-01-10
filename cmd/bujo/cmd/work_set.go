package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var workSetDate string
var workSetYes bool

var workSetCmd = &cobra.Command{
	Use:   "set <location>",
	Short: "Set work location for a day",
	Long: `Set the location context for a day.

If the habit doesn't exist, it will be created automatically.
Defaults to today if no date specified.

Examples:
  bujo work set "Home Office"
  bujo work set Manchester
  bujo work set "Coffee Shop" --date yesterday
  bujo work set "Client Site" -d "last monday"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		location := strings.Join(args, " ")

		targetDate, err := parseDateOrToday(workSetDate)
		if err != nil {
			return err
		}

		if workSetDate != "" {
			targetDate, err = confirmDate(workSetDate, targetDate, workSetYes)
			if err != nil {
				return err
			}
		}

		err = bujoService.SetLocation(cmd.Context(), targetDate, location)
		if err != nil {
			return fmt.Errorf("failed to set location: %w", err)
		}

		if workSetDate == "" {
			fmt.Fprintf(os.Stderr, "✓ Location set to: %s\n", location)
		} else {
			fmt.Fprintf(os.Stderr, "✓ Location for %s set to: %s\n", targetDate.Format("Jan 2, 2006"), location)
		}
		return nil
	},
}

func init() {
	workSetCmd.Flags().StringVarP(&workSetDate, "date", "d", "", "Date to set location for (e.g., 'yesterday', '2026-01-05')")
	workSetCmd.Flags().BoolVarP(&workSetYes, "yes", "y", false, "Skip date confirmation prompt")
	workCmd.AddCommand(workSetCmd)
}
