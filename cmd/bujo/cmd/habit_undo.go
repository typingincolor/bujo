package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
		nameOrID := args[0]

		var err error
		var displayName string

		if strings.HasPrefix(nameOrID, "#") {
			habitID, parseErr := strconv.ParseInt(nameOrID[1:], 10, 64)
			if parseErr != nil {
				return fmt.Errorf("invalid habit ID: %s", nameOrID)
			}
			err = habitService.UndoLastLogByID(cmd.Context(), habitID)
			displayName = nameOrID
		} else {
			err = habitService.UndoLastLog(cmd.Context(), nameOrID)
			displayName = nameOrID
		}

		if err != nil {
			return fmt.Errorf("failed to undo habit: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Undid last log for: %s\n", displayName)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitUndoCmd)
}
