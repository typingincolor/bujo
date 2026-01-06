package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var habitUndoCmd = &cobra.Command{
	Use:   "undo <habit-name|#id>",
	Short: "Undo the last habit log",
	Long: `Undo (delete) the most recent log entry for a habit.

Use this to correct mistakes when you accidentally log a habit.
Use #<id> to specify the habit by ID (shown in bujo habit output).

Examples:
  bujo habit undo Gym
  bujo habit undo #1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, id, isID, err := parseHabitNameOrID(args[0])
		if err != nil {
			return err
		}

		if isID {
			err = habitService.UndoLastLogByID(cmd.Context(), id)
		} else {
			err = habitService.UndoLastLog(cmd.Context(), name)
		}

		if err != nil {
			return fmt.Errorf("failed to undo habit: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Undid last log for: %s\n", args[0])
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitUndoCmd)
}
