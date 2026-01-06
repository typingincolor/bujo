package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var habitLogDate string

var habitLogCmd = &cobra.Command{
	Use:   "log <habit-name> [count]",
	Short: "Log a habit completion",
	Long: `Log a habit completion for today or a specific date.

If the habit doesn't exist, it will be created automatically.
Count defaults to 1 if not specified.

Examples:
  bujo habit log Gym
  bujo habit log Water 8
  bujo habit log "Morning Run"
  bujo habit log Gym --date yesterday
  bujo habit log Gym -d 2026-01-05`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		count := 1

		if len(args) > 1 {
			var err error
			count, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid count: %s", args[1])
			}
		}

		logDate := time.Now()
		if habitLogDate != "" {
			parsed, err := parseDate(habitLogDate)
			if err != nil {
				return err
			}
			logDate = parsed
		}

		err := habitService.LogHabitForDate(cmd.Context(), name, count, logDate)
		if err != nil {
			return fmt.Errorf("failed to log habit: %w", err)
		}

		if habitLogDate != "" {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s for %s\n", name, habitLogDate)
		} else if count == 1 {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s\n", name)
		} else {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s (x%d)\n", name, count)
		}

		return nil
	},
}

func init() {
	habitLogCmd.Flags().StringVarP(&habitLogDate, "date", "d", "", "Date to log for (YYYY-MM-DD or 'yesterday')")
	habitCmd.AddCommand(habitLogCmd)
}

func parseDate(s string) (time.Time, error) {
	today := time.Now()

	switch s {
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	case "today":
		return today, nil
	default:
		parsed, err := time.Parse("2006-01-02", s)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date format (use YYYY-MM-DD or 'yesterday'): %s", s)
		}
		return parsed, nil
	}
}
