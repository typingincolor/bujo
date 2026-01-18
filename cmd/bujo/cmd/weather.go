package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var weatherCmd = &cobra.Command{
	Use:   "weather",
	Short: "Manage daily weather",
	Long: `Manage daily weather tracking.

When called without a subcommand, shows today's weather.

Examples:
  bujo weather                         # Show today's weather
  bujo weather set sunny               # Set today's weather
  bujo weather set "Rainy, 15°C" -d yesterday  # Set weather for a past date
  bujo weather show --from "last week"     # View weather history
  bujo weather clear -d yesterday      # Clear a day's weather`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()
		weather, err := bujoService.GetWeather(cmd.Context(), today)
		if err != nil {
			return fmt.Errorf("failed to get weather: %w", err)
		}

		if weather == nil {
			fmt.Println("No weather set for today")
		} else {
			fmt.Printf("Today's weather: %s\n", *weather)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(weatherCmd)
}
