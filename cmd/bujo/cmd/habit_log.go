package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var habitLogCmd = &cobra.Command{
	Use:   "log <habit-name> [count]",
	Short: "Log a habit completion",
	Long: `Log a habit completion for today.

If the habit doesn't exist, it will be created automatically.
Count defaults to 1 if not specified.

Examples:
  bujo habit log Gym
  bujo habit log Water 8
  bujo habit log "Morning Run"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		count := 1

		if len(args) > 1 {
			var err error
			count, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid count: %s", args[1])
			}
		}

		err := habitService.LogHabit(cmd.Context(), name, count)
		if err != nil {
			return fmt.Errorf("failed to log habit: %w", err)
		}

		if count == 1 {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s\n", name)
		} else {
			fmt.Fprintf(os.Stderr, "✓ Logged: %s (x%d)\n", name, count)
		}

		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitLogCmd)
}
