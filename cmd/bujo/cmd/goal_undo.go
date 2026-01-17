package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var goalUndoCmd = &cobra.Command{
	Use:   "undo <#id>",
	Short: "Mark a goal as active again",
	Long: `Mark a completed goal as active again.

Examples:
  bujo goal undo #1
  bujo goal undo 1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseGoalID(args[0])
		if err != nil {
			return err
		}

		err = services.Goal.MarkActive(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to mark goal as active: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Marked goal #%d as active\n", id)
		return nil
	},
}

func init() {
	goalCmd.AddCommand(goalUndoCmd)
}
