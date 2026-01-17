package cmd

import (
	"bufio"
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
			err = services.Habit.UndoLastLogByID(cmd.Context(), id)
		} else {
			err = services.Habit.UndoLastLog(cmd.Context(), name)
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
