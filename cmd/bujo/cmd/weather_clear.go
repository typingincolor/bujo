package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var weatherClearDate string
var weatherClearYes bool

var weatherClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear weather for a day",
	Long: `Clear the weather for a day.

Defaults to today if no date specified.

Examples:
  bujo weather clear
  bujo weather clear --date yesterday
  bujo weather clear -d "last monday"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDate, err := parseDateOrToday(weatherClearDate)
		if err != nil {
			return err
		}

		if weatherClearDate != "" {
			targetDate, err = confirmDate(weatherClearDate, targetDate, weatherClearYes)
			if err != nil {
				return err
			}
		}

		err = bujoService.ClearWeather(cmd.Context(), targetDate)
		if err != nil {
			return fmt.Errorf("failed to clear weather: %w", err)
		}

		if weatherClearDate == "" {
			fmt.Fprintln(os.Stderr, "✓ Weather cleared for today")
		} else {
			fmt.Fprintf(os.Stderr, "✓ Weather cleared for %s\n", targetDate.Format("Jan 2, 2006"))
		}
		return nil
	},
}

func init() {
	weatherClearCmd.Flags().StringVarP(&weatherClearDate, "date", "d", "", "Date to clear weather for (e.g., 'yesterday', '2026-01-05')")
	weatherClearCmd.Flags().BoolVarP(&weatherClearYes, "yes", "y", false, "Skip date confirmation prompt")
	weatherCmd.AddCommand(weatherClearCmd)
}
