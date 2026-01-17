package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	weatherShowFrom string
	weatherShowTo   string
)

var weatherShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show weather history",
	Long: `Show weather history for a date range.

By default shows the last 30 days. Use --from and --to to specify a date range.

Examples:
  bujo weather show
  bujo weather show --from 2025-12-01
  bujo weather show --from "last month"
  bujo weather show --from 2025-12-01 --to 2025-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()

		from := today.AddDate(0, 0, -30)
		to := today

		if weatherShowFrom != "" {
			parsed, err := parsePastDate(weatherShowFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if weatherShowTo != "" {
			parsed, err := parsePastDate(weatherShowTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		if err := validateDateRange(from, to); err != nil {
			return err
		}

		history, err := services.Bujo.GetWeatherHistory(cmd.Context(), from, to)
		if err != nil {
			return fmt.Errorf("failed to get weather history: %w", err)
		}

		if len(history) == 0 {
			fmt.Println("No weather recorded in this period")
			return nil
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf("Weather History (%s to %s)\n", from.Format("Jan 2"), to.Format("Jan 2, 2006"))
		fmt.Println(strings.Repeat("-", 50))

		for _, ctx := range history {
			weather := "(no weather)"
			if ctx.Weather != nil {
				weather = *ctx.Weather
			}
			fmt.Printf("  %s  %s\n", cyan(ctx.Date.Format("Mon, Jan 2")), yellow(weather))
		}

		return nil
	},
}

func init() {
	weatherShowCmd.Flags().StringVar(&weatherShowFrom, "from", "", "Start date (e.g., '2025-12-01', 'last month')")
	weatherShowCmd.Flags().StringVar(&weatherShowTo, "to", "", "End date (e.g., '2025-12-31', 'yesterday')")
	weatherCmd.AddCommand(weatherShowCmd)
}
