package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var habitSetMonthlyGoalCmd = &cobra.Command{
	Use:   "set-monthly-goal <name|#id> <goal>",
	Short: "Set monthly goal for a habit",
	Long: `Set the monthly goal for a habit.

The monthly goal tracks how many times you want to complete a habit per month.
Monthly progress is calculated based on logs from the current calendar month.

Examples:
  bujo habit set-monthly-goal Reading 20
  bujo habit set-monthly-goal #1 15
  bujo habit set-monthly-goal "Deep Work" 10`,
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
			err = habitService.SetHabitMonthlyGoalByID(cmd.Context(), id, goal)
		} else {
			err = habitService.SetHabitMonthlyGoal(cmd.Context(), name, goal)
		}

		if err != nil {
			return fmt.Errorf("failed to set monthly goal: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Set monthly goal for %s to %d/month\n", displayName, goal)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitSetMonthlyGoalCmd)
}
