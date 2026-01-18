package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var goalMonthFlag string

var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Manage monthly goals",
	Long: `Manage monthly goals - higher level objectives tracked by month.

By default, shows goals for the current month. Use --month to specify a different month.

Examples:
  bujo goal                        # List current month's goals
  bujo goal --month 2026-02        # List February 2026 goals
  bujo goal add "Learn Go"         # Add goal to current month
  bujo goal done #1                # Mark goal #1 as done
  bujo goal undo #1                # Mark goal #1 as active again
  bujo goal move #1 2026-02        # Move goal #1 to February`,
	RunE: func(cmd *cobra.Command, args []string) error {
		month, err := parseGoalMonth(goalMonthFlag)
		if err != nil {
			return err
		}

		goals, err := goalService.GetGoalsForMonth(cmd.Context(), month)
		if err != nil {
			return fmt.Errorf("failed to get goals: %w", err)
		}

		if len(goals) == 0 {
			fmt.Printf("No goals for %s\n", month.Format("January 2006"))
			return nil
		}

		fmt.Printf("Goals for %s:\n\n", month.Format("January 2006"))
		for _, goal := range goals {
			status := "  "
			if goal.IsDone() {
				status = "x "
			}
			fmt.Printf("  %s#%-3d %s\n", status, goal.ID, goal.Content)
		}
		return nil
	},
}

func parseGoalMonth(monthStr string) (time.Time, error) {
	if monthStr == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	month, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month format: %s (use YYYY-MM)", monthStr)
	}
	return month, nil
}

func parseGoalID(arg string) (int64, error) {
	if len(arg) > 0 && arg[0] == '#' {
		arg = arg[1:]
	}

	var id int64
	_, err := fmt.Sscanf(arg, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid goal ID: %s", arg)
	}
	return id, nil
}

func init() {
	goalCmd.Flags().StringVar(&goalMonthFlag, "month", "", "Month in YYYY-MM format (default: current month)")
	rootCmd.AddCommand(goalCmd)
}
