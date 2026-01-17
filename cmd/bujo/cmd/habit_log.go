package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var habitLogDate string
var habitLogYes bool

var habitLogCmd = &cobra.Command{
	Use:   "log <habit-name|#id> [count]",
	Short: "Log a habit completion",
	Long: `Log a habit completion for today or a specific date.

If the habit doesn't exist, you will be prompted to create it.
Use --yes to skip the confirmation and create automatically.
Count defaults to 1 if not specified.
Use #<id> to log by habit ID (shown in bujo habit output).

Examples:
  bujo habit log Gym
  bujo habit log Water 8
  bujo habit log "Morning Run"
  bujo habit log #1              (log by ID)
  bujo habit log #2 5            (log by ID with count)
  bujo habit log Gym --date yesterday
  bujo habit log Gym -d 2026-01-05
  bujo habit log NewHabit --yes  (create without prompting)`,
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

		if habitLogDate != "" {
			logDate, err = confirmDate(habitLogDate, logDate, habitLogYes)
			if err != nil {
				return err
			}
		}

		displayName := args[0]

		if !isID && isPureNumber(args[0]) && !habitLogYes {
			fmt.Printf("'%s' looks like an ID. Did you mean to use #%s? [y/N]: ", args[0], args[0])
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			confirm := strings.TrimSpace(strings.ToLower(input))

			if confirm == "y" || confirm == "yes" {
				id, _ = strconv.ParseInt(args[0], 10, 64)
				isID = true
				displayName = "#" + args[0]
			}
		}

		if isID {
			err = services.Habit.LogHabitByIDForDate(cmd.Context(), id, count, logDate)
		} else {
			var exists bool
			exists, err = services.Habit.HabitExists(cmd.Context(), name)
			if err != nil {
				return fmt.Errorf("failed to check habit: %w", err)
			}

			if !exists && !habitLogYes {
				fmt.Printf("Habit '%s' does not exist. Create it? [y/N]: ", name)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				confirm := strings.TrimSpace(strings.ToLower(input))

				if confirm != "y" && confirm != "yes" {
					fmt.Fprintln(os.Stderr, "Cancelled")
					return nil
				}
			}

			err = services.Habit.LogHabitForDate(cmd.Context(), name, count, logDate)
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
	habitLogCmd.Flags().BoolVarP(&habitLogYes, "yes", "y", false, "Create habit without prompting if it doesn't exist")
	habitCmd.AddCommand(habitLogCmd)
}
