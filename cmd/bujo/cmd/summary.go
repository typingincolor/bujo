package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
)

var (
	summaryWeekly  bool
	summaryDate    string
	summaryRefresh bool
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Generate AI-powered summary",
	Long: `Generate AI-powered reflections and summaries of your journal entries.

Supports both local AI (private, offline) and Gemini API (cloud-based).
Configure with BUJO_AI_PROVIDER, BUJO_MODEL, or GEMINI_API_KEY.

Examples:
  bujo summary           # Daily summary
  bujo summary --weekly  # Weekly reflection`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if summaryService == nil {
			return fmt.Errorf("AI not configured. Set BUJO_MODEL or GEMINI_API_KEY")
		}

		horizon := domain.SummaryHorizonDaily
		if summaryWeekly {
			horizon = domain.SummaryHorizonWeekly
		}

		refDate, err := parseDateOrToday(summaryDate)
		if err != nil {
			return fmt.Errorf("invalid date: %w", err)
		}

		bold := color.New(color.Bold).SprintFunc()
		dimmed := color.New(color.Faint).SprintFunc()

		fmt.Printf("%s\n", bold(fmt.Sprintf("Generating %s summary...", horizon)))
		fmt.Println(dimmed(strings.Repeat("-", 50)))

		summary, err := summaryService.GetSummaryWithRefresh(cmd.Context(), horizon, refDate, summaryRefresh)
		if err != nil {
			return fmt.Errorf("failed to generate summary: %w", err)
		}

		fmt.Println()
		fmt.Println(summary.Content)
		fmt.Println()
		fmt.Printf("%s %s to %s\n",
			dimmed("Period:"),
			summary.StartDate.Format("Jan 2, 2006"),
			summary.EndDate.Format("Jan 2, 2006"))

		return nil
	},
}

func init() {
	summaryCmd.Flags().BoolVar(&summaryWeekly, "weekly", false, "Generate weekly reflection")
	summaryCmd.Flags().StringVarP(&summaryDate, "date", "d", "", "Reference date (e.g., yesterday, 2026-01-05)")
	summaryCmd.Flags().BoolVar(&summaryRefresh, "refresh", false, "Force regenerate even for completed periods")
	rootCmd.AddCommand(summaryCmd)
}
