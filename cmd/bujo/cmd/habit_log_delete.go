package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var habitLogDeleteCmd = &cobra.Command{
	Use:   "delete <log-id>",
	Short: "Delete a specific habit log entry",
	Long: `Delete a specific habit log entry by its ID.

Use 'bujo habit show <habit>' to see log IDs.

Examples:
  bujo habit show Gym       # See log IDs
  bujo habit log delete 42  # Delete log #42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid log ID: %s", args[0])
		}

		err = services.Habit.DeleteLog(cmd.Context(), logID)
		if err != nil {
			return fmt.Errorf("failed to delete log: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Deleted log #%d\n", logID)
		return nil
	},
}

func init() {
	habitLogCmd.AddCommand(habitLogDeleteCmd)
}
