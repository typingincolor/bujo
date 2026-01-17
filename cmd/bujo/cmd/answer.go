package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var answerCmd = &cobra.Command{
	Use:   "answer <id> <answer-text>",
	Short: "Mark a question as answered with the answer text",
	Long: `Mark a question entry as answered by providing the answer text.

The answer is stored as a child note entry under the question.

Use 'bujo ls' to see entry IDs.

Examples:
  bujo answer 42 "The answer is 42"
  bujo answer 15 "Use the --parent flag"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		answerText := strings.Join(args[1:], " ")

		err = services.Bujo.MarkAnswered(cmd.Context(), id, answerText)
		if err != nil {
			return fmt.Errorf("failed to mark answered: %w", err)
		}

		fmt.Fprintf(os.Stderr, "★ Marked question %d as answered\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(answerCmd)
}
