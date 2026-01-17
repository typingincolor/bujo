package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

var exportCmd = &cobra.Command{
	Use:   "export [entry-id]",
	Short: "Export data to JSON or markdown",
	Long: `Export bujo data to JSON format for backup or migration, or export a specific entry tree to markdown.

Examples:
  bujo export > backup.json              # Export all data
  bujo export --from 2026-01-01          # Export from date
  bujo export --from 2026-01-01 --to 2026-01-31  # Export date range
  bujo export 42                         # Export entry 42 and children as markdown
  bujo export 42 -o entry.md             # Export entry 42 to file`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

var (
	exportFrom   string
	exportTo     string
	exportFormat string
	exportOutput string
)

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportFrom, "from", "", "Start date for export (YYYY-MM-DD)")
	exportCmd.Flags().StringVar(&exportTo, "to", "", "End date for export (YYYY-MM-DD)")
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "Export format (json or csv)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (for markdown export)")
}

func runExport(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		return runMarkdownExport(cmd, args[0])
	}

	entryRepo := sqlite.NewEntryRepository(services.DB)
	habitRepo := sqlite.NewHabitRepository(services.DB)
	habitLogRepo := sqlite.NewHabitLogRepository(services.DB)
	dayContextRepo := sqlite.NewDayContextRepository(services.DB)
	summaryRepo := sqlite.NewSummaryRepository(services.DB)
	listRepo := sqlite.NewListRepository(services.DB)
	listItemRepo := sqlite.NewListItemRepository(services.DB)
	goalRepo := sqlite.NewGoalRepository(services.DB)

	exportSvc := service.NewExportService(
		entryRepo, habitRepo, habitLogRepo, dayContextRepo,
		summaryRepo, listRepo, listItemRepo, goalRepo,
	)

	opts := domain.NewExportOptions()

	if exportFrom != "" {
		from, err := time.Parse("2006-01-02", exportFrom)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		to := from
		if exportTo != "" {
			to, err = time.Parse("2006-01-02", exportTo)
			if err != nil {
				return fmt.Errorf("invalid --to date: %w", err)
			}
		}
		opts = opts.WithDateRange(from, to)
	}

	data, err := exportSvc.Export(cmd.Context(), opts)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	if exportFormat == "csv" {
		return exportCSV(data)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func runMarkdownExport(cmd *cobra.Command, entryIDStr string) error {
	entryID, err := parseEntryID(entryIDStr)
	if err != nil {
		return err
	}

	markdown, err := services.Bujo.ExportEntryMarkdown(cmd.Context(), entryID)
	if err != nil {
		return fmt.Errorf("failed to export entry: %w", err)
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, []byte(markdown), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Exported to %s\n", exportOutput)
		return nil
	}

	fmt.Print(markdown)
	return nil
}

func exportCSV(data *domain.ExportData) error {
	fmt.Fprintln(os.Stderr, "CSV export creates separate files for each entity type.")

	if err := writeEntriesCSV(data.Entries); err != nil {
		return err
	}
	if err := writeHabitsCSV(data.Habits); err != nil {
		return err
	}
	if err := writeHabitLogsCSV(data.HabitLogs); err != nil {
		return err
	}
	if err := writeListsCSV(data.Lists); err != nil {
		return err
	}
	if err := writeListItemsCSV(data.ListItems); err != nil {
		return err
	}
	if err := writeGoalsCSV(data.Goals); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Export complete.")
	return nil
}

func writeEntriesCSV(entries []domain.Entry) error {
	f, err := os.Create("entries.csv")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintln(f, "id,entity_id,type,content,priority,parent_id,depth,location,scheduled_date,created_at"); err != nil {
		return err
	}
	for _, e := range entries {
		scheduledDate := ""
		if e.ScheduledDate != nil {
			scheduledDate = e.ScheduledDate.Format("2006-01-02")
		}
		parentID := ""
		if e.ParentID != nil {
			parentID = fmt.Sprintf("%d", *e.ParentID)
		}
		location := ""
		if e.Location != nil {
			location = *e.Location
		}
		if _, err := fmt.Fprintf(f, "%d,%s,%s,%q,%s,%s,%d,%q,%s,%s\n",
			e.ID, e.EntityID, e.Type, e.Content, e.Priority, parentID, e.Depth, location, scheduledDate, e.CreatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "  Created entries.csv")
	return nil
}

func writeHabitsCSV(habits []domain.Habit) error {
	f, err := os.Create("habits.csv")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintln(f, "id,entity_id,name,goal_per_day,goal_per_week,goal_per_month,created_at"); err != nil {
		return err
	}
	for _, h := range habits {
		if _, err := fmt.Fprintf(f, "%d,%s,%q,%d,%d,%d,%s\n",
			h.ID, h.EntityID, h.Name, h.GoalPerDay, h.GoalPerWeek, h.GoalPerMonth, h.CreatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "  Created habits.csv")
	return nil
}

func writeHabitLogsCSV(logs []domain.HabitLog) error {
	f, err := os.Create("habit_logs.csv")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintln(f, "id,entity_id,habit_id,habit_entity_id,count,logged_at"); err != nil {
		return err
	}
	for _, l := range logs {
		if _, err := fmt.Fprintf(f, "%d,%s,%d,%s,%d,%s\n",
			l.ID, l.EntityID, l.HabitID, l.HabitEntityID, l.Count, l.LoggedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "  Created habit_logs.csv")
	return nil
}

func writeListsCSV(lists []domain.List) error {
	f, err := os.Create("lists.csv")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintln(f, "id,entity_id,name,created_at"); err != nil {
		return err
	}
	for _, l := range lists {
		if _, err := fmt.Fprintf(f, "%d,%s,%q,%s\n",
			l.ID, l.EntityID, l.Name, l.CreatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "  Created lists.csv")
	return nil
}

func writeListItemsCSV(items []domain.ListItem) error {
	f, err := os.Create("list_items.csv")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintln(f, "row_id,entity_id,version,list_entity_id,type,content,created_at"); err != nil {
		return err
	}
	for _, i := range items {
		if _, err := fmt.Fprintf(f, "%d,%s,%d,%s,%s,%q,%s\n",
			i.RowID, i.EntityID, i.Version, i.ListEntityID, i.Type, i.Content, i.CreatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "  Created list_items.csv")
	return nil
}

func writeGoalsCSV(goals []domain.Goal) error {
	f, err := os.Create("goals.csv")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintln(f, "id,entity_id,content,month,status,created_at"); err != nil {
		return err
	}
	for _, g := range goals {
		if _, err := fmt.Fprintf(f, "%d,%s,%q,%s,%s,%s\n",
			g.ID, g.EntityID, g.Content, g.Month.Format("2006-01"), g.Status, g.CreatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "  Created goals.csv")
	return nil
}
