package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var goalDoneCmd = &cobra.Command{
	Use:   "done <#id>",
	Short: "Mark a goal as done",
	Long: `Mark a goal as completed.

Examples:
  bujo goal done #1
  bujo goal done 1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseGoalID(args[0])
		if err != nil {
			return err
		}

		err = services.Goal.MarkDone(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to mark goal as done: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Marked goal #%d as done\n", id)
		return nil
	},
}

func init() {
	goalCmd.AddCommand(goalDoneCmd)
}
