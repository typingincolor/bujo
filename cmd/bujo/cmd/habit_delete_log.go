package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var habitDeleteLogCmd = &cobra.Command{
	Use:   "delete-log <log-id>",
	Short: "Delete a specific habit log entry",
	Long: `Delete a specific habit log entry by its ID.

Use 'bujo habit inspect <habit>' to see log IDs.

Examples:
  bujo habit inspect Gym    # See log IDs
  bujo habit delete-log 42  # Delete log #42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid log ID: %s", args[0])
		}

		err = habitService.DeleteLog(cmd.Context(), logID)
		if err != nil {
			return fmt.Errorf("failed to delete log: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Deleted log #%d\n", logID)
		return nil
	},
}

func init() {
	habitCmd.AddCommand(habitDeleteLogCmd)
}
