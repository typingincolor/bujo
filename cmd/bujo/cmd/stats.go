package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var (
	statsFrom string
	statsTo   string
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show summary statistics",
	Long: `Show summary statistics about your journal usage.

Displays entry counts by type, task completion rate, productivity patterns,
and habit tracking overview.

Examples:
  bujo stats                       # Stats for the last 30 days
  bujo stats --from "last month"   # Custom date range
  bujo stats --from "2026-01-01" --to "2026-01-31"  # Specific range`,
	RunE: runStats,
}

func init() {
	statsCmd.Flags().StringVarP(&statsFrom, "from", "f", "", "Start date (natural language or YYYY-MM-DD)")
	statsCmd.Flags().StringVarP(&statsTo, "to", "t", "", "End date (natural language or YYYY-MM-DD)")
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	to := time.Now()
	from := to.AddDate(0, 0, -29) // Default: last 30 days

	if statsFrom != "" {
		parsed, err := parsePastDate(statsFrom)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		from = parsed
	}

	if statsTo != "" {
		parsed, err := parsePastDate(statsTo)
		if err != nil {
			return fmt.Errorf("invalid --to date: %w", err)
		}
		to = parsed
	}

	stats, err := services.Stats.GetStats(cmd.Context(), from, to)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	days := int(to.Sub(from).Hours()/24) + 1
	fmt.Printf("%s Statistics (%s to %s)\n", cli.Bold("📊"), from.Format("Jan 2"), to.Format("Jan 2, 2006"))
	fmt.Println(cli.Dimmed("───────────────────────────────"))

	fmt.Printf("\n%s %d total\n", cli.Bold("Entries:"), stats.EntryCounts.Total)
	if stats.EntryCounts.Tasks > 0 {
		pct := float64(stats.EntryCounts.Tasks) / float64(stats.EntryCounts.Total) * 100
		fmt.Printf("  • Tasks:     %d (%.0f%%)\n", stats.EntryCounts.Tasks, pct)
	}
	if stats.EntryCounts.Notes > 0 {
		pct := float64(stats.EntryCounts.Notes) / float64(stats.EntryCounts.Total) * 100
		fmt.Printf("  – Notes:     %d (%.0f%%)\n", stats.EntryCounts.Notes, pct)
	}
	if stats.EntryCounts.Events > 0 {
		pct := float64(stats.EntryCounts.Events) / float64(stats.EntryCounts.Total) * 100
		fmt.Printf("  ○ Events:    %d (%.0f%%)\n", stats.EntryCounts.Events, pct)
	}
	if stats.EntryCounts.Done > 0 {
		pct := float64(stats.EntryCounts.Done) / float64(stats.EntryCounts.Total) * 100
		fmt.Printf("  ✓ Completed: %d (%.0f%%)\n", stats.EntryCounts.Done, pct)
	}
	if stats.EntryCounts.Migrated > 0 {
		pct := float64(stats.EntryCounts.Migrated) / float64(stats.EntryCounts.Total) * 100
		fmt.Printf("  → Migrated:  %d (%.0f%%)\n", stats.EntryCounts.Migrated, pct)
	}

	if stats.TaskCompletion.Total > 0 {
		fmt.Printf("\n%s %.0f%% (%d/%d)\n",
			cli.Bold("Task completion:"),
			stats.TaskCompletion.Rate,
			stats.TaskCompletion.Completed,
			stats.TaskCompletion.Total,
		)
	}

	if stats.Productivity.AveragePerDay > 0 {
		fmt.Printf("%s %.1f\n", cli.Bold("Average entries/day:"), stats.Productivity.AveragePerDay)
	}

	if stats.Productivity.MostProductive.Average > 0 {
		fmt.Printf("\n%s %s (avg %.1f)\n",
			cli.Bold("Most productive:"),
			stats.Productivity.MostProductive.Day.String()+"s",
			stats.Productivity.MostProductive.Average,
		)
	}
	if stats.Productivity.LeastProductive.Average > 0 && days > 7 {
		fmt.Printf("%s %s (avg %.1f)\n",
			cli.Bold("Least productive:"),
			stats.Productivity.LeastProductive.Day.String()+"s",
			stats.Productivity.LeastProductive.Average,
		)
	}

	if stats.HabitStats.Active > 0 {
		fmt.Printf("\n%s %d active\n", cli.Bold("Habits:"), stats.HabitStats.Active)
		if stats.HabitStats.BestStreak.Days > 0 {
			fmt.Printf("  Best streak: %s (%d days)\n",
				stats.HabitStats.BestStreak.HabitName,
				stats.HabitStats.BestStreak.Days,
			)
		}
		if stats.HabitStats.MostLogged.Count > 0 {
			fmt.Printf("  Most logged: %s (%d logs)\n",
				stats.HabitStats.MostLogged.HabitName,
				stats.HabitStats.MostLogged.Count,
			)
		}
	}

	return nil
}
