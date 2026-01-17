package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var uncancelCmd = &cobra.Command{
	Use:   "uncancel <id>",
	Short: "Restore a cancelled entry",
	Long: `Restore a cancelled entry back to a task.

This reverses the 'bujo cancel' command and makes the entry
active again.

Examples:
  bujo uncancel 42
  bujo uncancel 15`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = services.Bujo.UncancelEntry(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to uncancel: %w", err)
		}

		fmt.Fprintf(os.Stderr, "• Restored entry %d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uncancelCmd)
}
