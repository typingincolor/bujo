package cmd

import (
	"fmt"
	"os"

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
		newName := args[1]

		name, id, isID, err := parseHabitNameOrID(args[0])
		if err != nil {
			return err
		}

		if isID {
			err = habitService.RenameHabitByID(cmd.Context(), id, newName)
		} else {
			err = habitService.RenameHabit(cmd.Context(), name, newName)
		}

		if err != nil {
			return fmt.Errorf("failed to rename habit: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Renamed %s to %s\n", args[0], newName)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitRenameCmd)
}
