package cmd

import (
	"fmt"
	"os"
	"strconv"

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

		if isID {
			err = habitService.SetHabitGoalByID(cmd.Context(), id, goal)
		} else {
			err = habitService.SetHabitGoal(cmd.Context(), name, goal)
		}

		if err != nil {
			return fmt.Errorf("failed to set goal: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Set goal for %s to %d/day\n", args[0], goal)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitSetGoalCmd)
}
