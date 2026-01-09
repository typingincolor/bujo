package cmd

import (
	"bufio"
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
		newName := args[1]

		name, id, isID, err := parseHabitNameOrID(args[0])
		if err != nil {
			return err
		}

		displayName := args[0]

		if !isID && isPureNumber(args[0]) {
			fmt.Printf("'%s' looks like an ID. Did you mean to use #%s? [y/N]: ", args[0], args[0])
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			confirm := strings.TrimSpace(strings.ToLower(input))

			if confirm == "y" || confirm == "yes" {
				id, _ = strconv.ParseInt(args[0], 10, 64)
				isID = true
				displayName = "#" + args[0]
			}
		}

		if isID {
			err = habitService.RenameHabitByID(cmd.Context(), id, newName)
		} else {
			err = habitService.RenameHabit(cmd.Context(), name, newName)
		}

		if err != nil {
			return fmt.Errorf("failed to rename habit: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Renamed %s to %s\n", displayName, newName)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitRenameCmd)
}
