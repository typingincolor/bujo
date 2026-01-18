package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var habitSetWeeklyGoalCmd = &cobra.Command{
	Use:   "set-weekly-goal <name|#id> <goal>",
	Short: "Set weekly goal for a habit",
	Long: `Set the weekly goal for a habit.

The weekly goal tracks how many times you want to complete a habit per week.
Weekly progress is calculated based on logs from the last 7 days.

Examples:
  bujo habit set-weekly-goal Gym 5
  bujo habit set-weekly-goal #1 3
  bujo habit set-weekly-goal "Morning Run" 4`,
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
			err = habitService.SetHabitWeeklyGoalByID(cmd.Context(), id, goal)
		} else {
			err = habitService.SetHabitWeeklyGoal(cmd.Context(), name, goal)
		}

		if err != nil {
			return fmt.Errorf("failed to set weekly goal: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Set weekly goal for %s to %d/week\n", displayName, goal)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitSetWeeklyGoalCmd)
}
