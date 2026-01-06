package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj/go-naturaldate"
)

var habitLogDate string

var habitLogCmd = &cobra.Command{
	Use:   "log <habit-name|#id> [count]",
	Short: "Log a habit completion",
	Long: `Log a habit completion for today or a specific date.

If the habit doesn't exist, it will be created automatically.
Count defaults to 1 if not specified.
Use #<id> to log by habit ID (shown in bujo habit output).

Examples:
  bujo habit log Gym
  bujo habit log Water 8
  bujo habit log "Morning Run"
  bujo habit log #1              (log by ID)
  bujo habit log #2 5            (log by ID with count)
  bujo habit log Gym --date yesterday
  bujo habit log Gym -d 2026-01-05`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]
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

		var err error
		var displayName string

		if strings.HasPrefix(nameOrID, "#") {
			habitID, parseErr := strconv.ParseInt(nameOrID[1:], 10, 64)
			if parseErr != nil {
				return fmt.Errorf("invalid habit ID: %s", nameOrID)
			}
			err = habitService.LogHabitByIDForDate(cmd.Context(), habitID, count, logDate)
			displayName = nameOrID
		} else {
			err = habitService.LogHabitForDate(cmd.Context(), nameOrID, count, logDate)
			displayName = nameOrID
		}

		if err != nil {
			return fmt.Errorf("failed to log habit: %w", err)
		}

		if habitLogDate != "" {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s for %s\n", displayName, habitLogDate)
		} else if count == 1 {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s\n", displayName)
		} else {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s (x%d)\n", displayName, count)
		}

		return nil
	},
}

func init() {
	habitLogCmd.Flags().StringVarP(&habitLogDate, "date", "d", "", "Date to log for (e.g., 'yesterday', 'last monday', '2 weeks ago')")
	habitCmd.AddCommand(habitLogCmd)
}

func parseDate(s string) (time.Time, error) {
	now := time.Now()

	// Try standard date format first
	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	// Try natural language parsing
	parsed, err := naturaldate.Parse(s, now, naturaldate.WithDirection(naturaldate.Past))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}
