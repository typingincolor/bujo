package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var habitRenameCmd = &cobra.Command{
	Use:   "rename <old-name|#id> <new-name>",
	Short: "Rename a habit",
	Long: `Rename a habit to a new name.

All existing logs are preserved under the new name.

Examples:
  bujo habit rename Gym Workout
  bujo habit rename #1 "Morning Workout"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldNameOrID := args[0]
		newName := args[1]

		var err error
		var displayOld string

		if strings.HasPrefix(oldNameOrID, "#") {
			habitID, parseErr := strconv.ParseInt(oldNameOrID[1:], 10, 64)
			if parseErr != nil {
				return fmt.Errorf("invalid habit ID: %s", oldNameOrID)
			}
			err = habitService.RenameHabitByID(cmd.Context(), habitID, newName)
			displayOld = oldNameOrID
		} else {
			err = habitService.RenameHabit(cmd.Context(), oldNameOrID, newName)
			displayOld = oldNameOrID
		}

		if err != nil {
			return fmt.Errorf("failed to rename habit: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Renamed %s to %s\n", displayOld, newName)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitRenameCmd)
}
