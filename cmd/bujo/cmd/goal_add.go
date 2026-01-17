package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var goalAddMonthFlag string

var goalAddCmd = &cobra.Command{
	Use:   "add <content>",
	Short: "Add a new monthly goal",
	Long: `Add a new goal to the specified month (default: current month).

Examples:
  bujo goal add "Learn Go"
  bujo goal add "Read 12 books" --month 2026-02
  bujo goal add "Ship new feature"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content := strings.Join(args, " ")

		month, err := parseGoalMonth(goalAddMonthFlag)
		if err != nil {
			return err
		}

		id, err := services.Goal.CreateGoal(cmd.Context(), content, month)
		if err != nil {
			return fmt.Errorf("failed to create goal: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Created goal #%d for %s\n", id, month.Format("January 2006"))
		return nil
	},
}

func init() {
	goalAddCmd.Flags().StringVar(&goalAddMonthFlag, "month", "", "Month in YYYY-MM format (default: current month)")
	goalCmd.AddCommand(goalAddCmd)
}
