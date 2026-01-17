package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
	"github.com/typingincolor/bujo/internal/domain"
)

var (
	questionsAll   bool
	questionsLimit int
)

var questionsCmd = &cobra.Command{
	Use:   "questions",
	Short: "List all unanswered questions",
	Long: `List questions across all dates.

By default, shows only unanswered questions. Use --all to include answered questions.
Use --limit to control the maximum number of results (default: 100).

Use 'bujo answer <id> <answer-text>' to answer a question.
Use 'bujo reopen <id>' to reopen an answered question.

Examples:
  bujo questions
  bujo questions --all
  bujo questions --limit 50`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := domain.NewSearchOptions("").WithLimit(questionsLimit)

		if !questionsAll {
			opts = opts.WithType(domain.EntryTypeQuestion)
		}

		results, err := services.Bujo.SearchEntries(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list questions: %w", err)
		}

		if questionsAll {
			filtered := make([]domain.Entry, 0)
			for _, entry := range results {
				if entry.Type == domain.EntryTypeQuestion || entry.Type == domain.EntryTypeAnswered {
					filtered = append(filtered, entry)
				}
			}
			results = filtered
		}

		if len(results) == 0 {
			if questionsAll {
				fmt.Println("No questions found")
			} else {
				fmt.Println("No unanswered questions")
			}
			return nil
		}

		for _, entry := range results {
			fmt.Println(formatQuestionEntry(entry))
		}

		if questionsAll {
			fmt.Printf("\n%d question(s)\n", len(results))
		} else {
			fmt.Printf("\n%d unanswered question(s)\n", len(results))
		}

		return nil
	},
}

func init() {
	questionsCmd.Flags().BoolVar(&questionsAll, "all", false, "Show both answered and unanswered questions")
	questionsCmd.Flags().IntVar(&questionsLimit, "limit", 100, "Maximum number of questions to show")
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
