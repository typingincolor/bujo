package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var goalMoveCmd = &cobra.Command{
	Use:   "move <#id> <YYYY-MM>",
	Short: "Move a goal to a different month",
	Long: `Move a goal to a different month.

Examples:
  bujo goal move #1 2026-02
  bujo goal move 1 2026-03`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseGoalID(args[0])
		if err != nil {
			return err
		}

		newMonth, err := parseGoalMonth(args[1])
		if err != nil {
			return err
		}

		err = goalService.MoveToMonth(cmd.Context(), id, newMonth)
		if err != nil {
			return fmt.Errorf("failed to move goal: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Moved goal #%d to %s\n", id, newMonth.Format("January 2006"))
		return nil
	},
}

func init() {
	goalCmd.AddCommand(goalMoveCmd)
}
