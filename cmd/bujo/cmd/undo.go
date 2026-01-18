package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo <id>",
	Short: "Mark a completed entry as incomplete",
	Long: `Mark a completed task back to incomplete by its ID.

Use 'bujo ls' to see entry IDs.

Examples:
  bujo undo 42
  bujo undo 15`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = bujoService.Undo(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to undo: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Marked entry %d as incomplete\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
