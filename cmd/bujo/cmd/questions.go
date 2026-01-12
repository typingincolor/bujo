package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
	"github.com/typingincolor/bujo/internal/domain"
)

var questionsCmd = &cobra.Command{
	Use:   "questions",
	Short: "List all unanswered questions",
	Long: `List all unanswered questions across all dates.

Shows questions that haven't been answered yet, with their scheduled date and ID.
Use 'bujo answer <id> <answer-text>' to answer a question.

Examples:
  bujo questions`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := domain.NewSearchOptions("").WithType(domain.EntryTypeQuestion).WithLimit(100)
		results, err := bujoService.SearchEntries(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list questions: %w", err)
		}

		if len(results) == 0 {
			fmt.Println("No unanswered questions")
			return nil
		}

		for _, entry := range results {
			fmt.Println(formatQuestionEntry(entry))
		}
		fmt.Printf("\n%d unanswered question(s)\n", len(results))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(questionsCmd)
}

func formatQuestionEntry(entry domain.Entry) string {
	var parts []string

	if entry.ScheduledDate != nil {
		parts = append(parts, cli.Dimmed(fmt.Sprintf("[%s]", entry.ScheduledDate.Format("2006-01-02"))))
	} else {
		parts = append(parts, cli.Dimmed("[no date]"))
	}

	parts = append(parts, entry.Type.Symbol())

	if entry.Priority != domain.PriorityNone {
		parts = append(parts, cli.Yellow(entry.Priority.Symbol()))
	}

	parts = append(parts, entry.Content)
	parts = append(parts, cli.Dimmed(fmt.Sprintf("(%d)", entry.ID)))

	return strings.Join(parts, " ")
}
