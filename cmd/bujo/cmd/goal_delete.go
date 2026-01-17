package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var goalDeleteCmd = &cobra.Command{
	Use:   "delete <#id>",
	Short: "Delete a goal",
	Long: `Delete a goal permanently.

Examples:
  bujo goal delete #1
  bujo goal delete 1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseGoalID(args[0])
		if err != nil {
			return err
		}

		err = services.Goal.DeleteGoal(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to delete goal: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Deleted goal #%d\n", id)
		return nil
	},
}

func init() {
	goalCmd.AddCommand(goalDeleteCmd)
}
