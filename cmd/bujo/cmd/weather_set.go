package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var weatherSetDate string
var weatherSetYes bool

var weatherSetCmd = &cobra.Command{
	Use:   "set <weather>",
	Short: "Set weather for a day",
	Long: `Set the weather for a day.

Defaults to today if no date specified.

Examples:
  bujo weather set sunny
  bujo weather set "Rainy, 15°C"
  bujo weather set cloudy --date yesterday
  bujo weather set "Snow, -5°C" -d "last monday"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		weather := strings.Join(args, " ")

		targetDate, err := parseDateOrToday(weatherSetDate)
		if err != nil {
			return err
		}

		if weatherSetDate != "" {
			targetDate, err = confirmDate(weatherSetDate, targetDate, weatherSetYes)
			if err != nil {
				return err
			}
		}

		err = services.Bujo.SetWeather(cmd.Context(), targetDate, weather)
		if err != nil {
			return fmt.Errorf("failed to set weather: %w", err)
		}

		if weatherSetDate == "" {
			fmt.Fprintf(os.Stderr, "✓ Weather set to: %s\n", weather)
		} else {
			fmt.Fprintf(os.Stderr, "✓ Weather for %s set to: %s\n", targetDate.Format("Jan 2, 2006"), weather)
		}
		return nil
	},
}

func init() {
	weatherSetCmd.Flags().StringVarP(&weatherSetDate, "date", "d", "", "Date to set weather for (e.g., 'yesterday', '2026-01-05')")
	weatherSetCmd.Flags().BoolVarP(&weatherSetYes, "yes", "y", false, "Skip date confirmation prompt")
	weatherCmd.AddCommand(weatherSetCmd)
}
