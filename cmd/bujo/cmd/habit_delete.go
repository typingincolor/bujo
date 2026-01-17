package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var habitDeleteForce bool

var habitDeleteCmd = &cobra.Command{
	Use:   "delete <habit-name|#id>",
	Short: "Delete a habit and all its logs",
	Long: `Delete a habit and all its log entries.

This is a destructive action - all logs for this habit will be permanently deleted.
You will be prompted to confirm unless --force is used.

Examples:
  bujo habit delete Gym
  bujo habit delete #1
  bujo habit delete "Morning Run" --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, id, isID, err := parseHabitNameOrID(args[0])
		if err != nil {
			return err
		}

		displayName := args[0]

		if !isID && isPureNumber(args[0]) && !habitDeleteForce {
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

		if !habitDeleteForce {
			fmt.Printf("Delete habit '%s' and all its logs? This cannot be undone.\n", displayName)
			fmt.Print("Type 'yes' to confirm: ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			confirm := strings.TrimSpace(input)

			if confirm != "yes" {
				fmt.Fprintln(os.Stderr, "Cancelled")
				return nil
			}
		}

		if isID {
			err = services.Habit.DeleteHabitByID(cmd.Context(), id)
		} else {
			err = services.Habit.DeleteHabit(cmd.Context(), name)
		}

		if err != nil {
			return fmt.Errorf("failed to delete habit: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Deleted habit '%s'\n", displayName)
		return nil
	},
}

func init() {
	habitDeleteCmd.Flags().BoolVarP(&habitDeleteForce, "force", "f", false, "Delete without prompting")
	habitCmd.AddCommand(habitDeleteCmd)
}
