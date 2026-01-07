package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
)

var (
	tasksFrom string
	tasksTo   string
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Show outstanding tasks",
	Long: `Show outstanding tasks (incomplete tasks only).

By default shows tasks from the last 30 days. Use --from and --to to specify a date range.

Examples:
  bujo tasks
  bujo tasks --from "last week"
  bujo tasks --from 2026-01-01 --to 2026-01-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now()
		today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

		from := today.AddDate(0, 0, -30)
		to := today

		if tasksFrom != "" {
			parsed, err := parsePastDate(tasksFrom)
			if err != nil {
				return err
			}
			from = parsed
		}

		if tasksTo != "" {
			parsed, err := parsePastDate(tasksTo)
			if err != nil {
				return err
			}
			to = parsed
		}

		if err := validateDateRange(from, to); err != nil {
			return err
		}

		tasks, err := bujoService.GetOutstandingTasks(cmd.Context(), from, to)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Println("No outstanding tasks")
			return nil
		}

		bold := color.New(color.Bold).SprintFunc()
		dimmed := color.New(color.Faint).SprintFunc()

		fmt.Printf("%s\n", bold("Outstanding Tasks"))
		fmt.Println(dimmed(strings.Repeat("-", 50)))

		// Group by date
		byDate := make(map[string][]struct {
			Content string
			ID      int64
		})
		var dates []string

		for _, task := range tasks {
			if task.ScheduledDate == nil {
				continue
			}
			dateKey := task.ScheduledDate.Format("2006-01-02")
			if _, exists := byDate[dateKey]; !exists {
				dates = append(dates, dateKey)
			}
			byDate[dateKey] = append(byDate[dateKey], struct {
				Content string
				ID      int64
			}{task.Content, task.ID})
		}

		for _, dateKey := range dates {
			parsed, _ := time.Parse("2006-01-02", dateKey)
			fmt.Printf("%s:\n", cli.Cyan(parsed.Format("Jan 2")))
			for _, task := range byDate[dateKey] {
				fmt.Printf("  . %s %s\n", task.Content, dimmed(fmt.Sprintf("(%d)", task.ID)))
			}
		}

		fmt.Println()
		fmt.Printf("%d task(s) outstanding\n", len(tasks))

		return nil
	},
}

func init() {
	tasksCmd.Flags().StringVar(&tasksFrom, "from", "", "Start date (e.g., '2026-01-01', 'last week')")
	tasksCmd.Flags().StringVar(&tasksTo, "to", "", "End date (e.g., '2026-01-31', 'yesterday')")
	rootCmd.AddCommand(tasksCmd)
}
