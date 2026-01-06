package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark an entry as complete",
	Long: `Mark a task or event as complete by its ID.

Use 'bujo ls' to see entry IDs.

Examples:
  bujo done 42
  bujo done 15`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = bujoService.MarkDone(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to mark done: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Marked entry %d as done\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
