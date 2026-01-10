package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
)

var (
	summaryWeekly    bool
	summaryQuarterly bool
	summaryAnnual    bool
	summaryDate      string
	summaryRefresh   bool
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Generate AI-powered summary",
	Long: `Generate AI-powered reflections and summaries of your journal entries.

Requires GEMINI_API_KEY environment variable to be set.

Examples:
  bujo summary              # Today's summary
  bujo summary --weekly     # Weekly reflection
  bujo summary --quarterly  # Quarterly review
  bujo summary --annual     # Annual review`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY environment variable is required")
		}

		if summaryService == nil {
			return fmt.Errorf("summary service not initialized - check GEMINI_API_KEY")
		}

		horizon := domain.SummaryHorizonDaily
		if summaryWeekly {
			horizon = domain.SummaryHorizonWeekly
		} else if summaryQuarterly {
			horizon = domain.SummaryHorizonQuarterly
		} else if summaryAnnual {
			horizon = domain.SummaryHorizonAnnual
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
	summaryCmd.Flags().BoolVar(&summaryQuarterly, "quarterly", false, "Generate quarterly review")
	summaryCmd.Flags().BoolVar(&summaryAnnual, "annual", false, "Generate annual review")
	summaryCmd.Flags().StringVarP(&summaryDate, "date", "d", "", "Reference date (e.g., yesterday, 2026-01-05)")
	summaryCmd.Flags().BoolVar(&summaryRefresh, "refresh", false, "Force regenerate even for completed periods")
	rootCmd.AddCommand(summaryCmd)
}
