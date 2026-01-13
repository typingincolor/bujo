package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var (
	lsFrom string
	lsTo   string
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Display entries for the last 7 days",
	Long: `Display entries for the last 7 days, including overdue tasks and monthly goals.

Use --from and --to to specify a custom date range.

Examples:
  bujo ls
  bujo ls --from yesterday
  bujo ls --from "last monday" --to today
  bujo ls --from 2026-01-01 --to 2026-01-07`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()
		todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

		from := todayStart.AddDate(0, 0, -6)
		to := todayStart

		if lsFrom != "" {
			parsed, err := parsePastDate(lsFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if lsTo != "" {
			parsed, err := parsePastDate(lsTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		if err := validateDateRange(from, to); err != nil {
			return err
		}

		agenda, err := bujoService.GetMultiDayAgenda(cmd.Context(), from, to)
		if err != nil {
			return fmt.Errorf("failed to get entries: %w", err)
		}

		fmt.Print(cli.RenderMultiDayAgenda(agenda, todayStart))

		currentMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		goals, err := goalService.GetGoalsForMonth(cmd.Context(), currentMonth)
		if err != nil {
			return fmt.Errorf("failed to get goals: %w", err)
		}

		fmt.Print(cli.RenderGoalsSection(goals, currentMonth))
		return nil
	},
}

func init() {
	lsCmd.Flags().StringVar(&lsFrom, "from", "", "Start date (e.g., 'yesterday', 'last monday', '2026-01-01')")
	lsCmd.Flags().StringVar(&lsTo, "to", "", "End date (e.g., 'today', '2026-01-07')")
	rootCmd.AddCommand(lsCmd)
}
