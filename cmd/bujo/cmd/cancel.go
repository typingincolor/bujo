package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel an entry (strikethrough)",
	Long: `Cancel an entry to mark it as no longer relevant.

Cancelled entries remain visible with strikethrough styling but are
clearly marked as not active. This is useful when a task becomes
irrelevant rather than completed.

Use 'bujo uncancel' to restore a cancelled entry.

Examples:
  bujo cancel 42
  bujo cancel 15`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = bujoService.CancelEntry(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to cancel: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✗ Cancelled entry %d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cancelCmd)
}
