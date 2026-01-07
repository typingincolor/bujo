package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	weatherInspectFrom string
	weatherInspectTo   string
)

var weatherInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Show weather history",
	Long: `Show weather history for a date range.

By default shows the last 30 days. Use --from and --to to specify a date range.

Examples:
  bujo weather inspect
  bujo weather inspect --from 2025-12-01
  bujo weather inspect --from "last month"
  bujo weather inspect --from 2025-12-01 --to 2025-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()

		from := today.AddDate(0, 0, -30)
		to := today

		if weatherInspectFrom != "" {
			parsed, err := parsePastDate(weatherInspectFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if weatherInspectTo != "" {
			parsed, err := parsePastDate(weatherInspectTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		if err := validateDateRange(from, to); err != nil {
			return err
		}

		history, err := bujoService.GetWeatherHistory(cmd.Context(), from, to)
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
	weatherInspectCmd.Flags().StringVar(&weatherInspectFrom, "from", "", "Start date (e.g., '2025-12-01', 'last month')")
	weatherInspectCmd.Flags().StringVar(&weatherInspectTo, "to", "", "End date (e.g., '2025-12-31', 'yesterday')")
	weatherCmd.AddCommand(weatherInspectCmd)
}
