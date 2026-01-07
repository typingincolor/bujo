package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
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
		name, id, isID, err := parseHabitNameOrID(args[0])
		if err != nil {
			return err
		}

		count := 1
		if len(args) > 1 {
			count, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid count: %s", args[1])
			}
		}

		logDate, err := parseDateOrToday(habitLogDate)
		if err != nil {
			return err
		}

		if isID {
			err = habitService.LogHabitByIDForDate(cmd.Context(), id, count, logDate)
		} else {
			err = habitService.LogHabitForDate(cmd.Context(), name, count, logDate)
		}

		if err != nil {
			return fmt.Errorf("failed to log habit: %w", err)
		}

		displayName := args[0]
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
