package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var reopenCmd = &cobra.Command{
	Use:   "reopen <id>",
	Short: "Reopen an answered question",
	Long: `Reopen a previously answered question, changing its status back to unanswered.

The answer entry (child note) will remain attached to the question.

Use 'bujo ls' or 'bujo questions' to see entry IDs.

Examples:
  bujo reopen 42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = services.Bujo.ReopenQuestion(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to reopen question: %w", err)
		}

		fmt.Fprintf(os.Stderr, "? Reopened question %d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reopenCmd)
}
