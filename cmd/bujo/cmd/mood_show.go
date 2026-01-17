package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	moodShowFrom string
	moodShowTo   string
)

var moodShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show mood history",
	Long: `Show mood history for a date range.

By default shows the last 30 days. Use --from and --to to specify a date range.

Examples:
  bujo mood show
  bujo mood show --from 2025-12-01
  bujo mood show --from "last month"
  bujo mood show --from 2025-12-01 --to 2025-12-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()

		from := today.AddDate(0, 0, -30)
		to := today

		if moodShowFrom != "" {
			parsed, err := parsePastDate(moodShowFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if moodShowTo != "" {
			parsed, err := parsePastDate(moodShowTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		if err := validateDateRange(from, to); err != nil {
			return err
		}

		history, err := services.Bujo.GetMoodHistory(cmd.Context(), from, to)
		if err != nil {
			return fmt.Errorf("failed to get mood history: %w", err)
		}

		if len(history) == 0 {
			fmt.Println("No moods recorded in this period")
			return nil
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf("Mood History (%s to %s)\n", from.Format("Jan 2"), to.Format("Jan 2, 2006"))
		fmt.Println(strings.Repeat("-", 50))

		for _, ctx := range history {
			mood := "(no mood)"
			if ctx.Mood != nil {
				mood = *ctx.Mood
			}
			fmt.Printf("  %s  %s\n", cyan(ctx.Date.Format("Mon, Jan 2")), yellow(mood))
		}

		return nil
	},
}

func init() {
	moodShowCmd.Flags().StringVar(&moodShowFrom, "from", "", "Start date (e.g., '2025-12-01', 'last month')")
	moodShowCmd.Flags().StringVar(&moodShowTo, "to", "", "End date (e.g., '2025-12-31', 'yesterday')")
	moodCmd.AddCommand(moodShowCmd)
}
