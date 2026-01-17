package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	workShowFrom string
	workShowTo   string
)

var workShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show location history",
	Long: `Show location history for a date range.

By default shows the last 30 days. Use --from and --to to specify a date range.

Examples:
  bujo work show
  bujo work show --from 2025-12-01
  bujo work show --from "last month"
  bujo work show --from 2025-12-01 --to 2025-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()

		from := today.AddDate(0, 0, -30)
		to := today

		if workShowFrom != "" {
			parsed, err := parsePastDate(workShowFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if workShowTo != "" {
			parsed, err := parsePastDate(workShowTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		if err := validateDateRange(from, to); err != nil {
			return err
		}

		history, err := bujoService.GetLocationHistory(cmd.Context(), from, to)
		if err != nil {
			return fmt.Errorf("failed to get location history: %w", err)
		}

		if len(history) == 0 {
			fmt.Println("No locations recorded in this period")
			return nil
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf("Location History (%s to %s)\n", from.Format("Jan 2"), to.Format("Jan 2, 2006"))
		fmt.Println(strings.Repeat("-", 50))

		for _, ctx := range history {
			location := "(no location)"
			if ctx.Location != nil {
				location = *ctx.Location
			}
			fmt.Printf("  %s  %s\n", cyan(ctx.Date.Format("Mon, Jan 2")), yellow(location))
		}

		return nil
	},
}

func init() {
	workShowCmd.Flags().StringVar(&workShowFrom, "from", "", "Start date (e.g., '2025-12-01', 'last month')")
	workShowCmd.Flags().StringVar(&workShowTo, "to", "", "End date (e.g., '2025-12-31', 'yesterday')")
	workCmd.AddCommand(workShowCmd)
}
