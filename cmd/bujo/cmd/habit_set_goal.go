package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var habitSetGoalCmd = &cobra.Command{
	Use:   "set-goal <name|#id> <goal>",
	Short: "Set daily goal for a habit",
	Long: `Set the daily goal for a habit.

The goal is used to calculate completion percentage and displayed in the tracker.

Examples:
  bujo habit set-goal Water 8
  bujo habit set-goal #1 10
  bujo habit set-goal Gym 1`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		goal, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid goal: %s (must be a number)", args[1])
		}

		name, id, isID, err := parseHabitNameOrID(args[0])
		if err != nil {
			return err
		}

		displayName := args[0]

		if !isID && isPureNumber(args[0]) {
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
			err = services.Habit.SetHabitGoalByID(cmd.Context(), id, goal)
		} else {
			err = services.Habit.SetHabitGoal(cmd.Context(), name, goal)
		}

		if err != nil {
			return fmt.Errorf("failed to set goal: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Set goal for %s to %d/day\n", displayName, goal)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitSetGoalCmd)
}
