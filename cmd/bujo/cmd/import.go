package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import data from a JSON backup file",
	Long: `Import bujo data from a JSON backup file.

Modes:
  merge   - Add new records, skip if entity_id already exists (default)
  replace - Clear all existing data and import fresh (destructive)

Examples:
  bujo import backup.json                    # Merge with existing data
  bujo import backup.json --mode replace     # Replace all data`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

var importMode string

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&importMode, "mode", "merge", "Import mode: merge or replace")
}

func runImport(cmd *cobra.Command, args []string) error {
	filename := args[0]

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var data domain.ExportData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if data.Version != domain.ExportVersion {
		fmt.Fprintf(os.Stderr, "Warning: Export version %s differs from current version %s\n", data.Version, domain.ExportVersion)
	}

	mode := domain.ImportModeMerge
	if importMode == "replace" {
		mode = domain.ImportModeReplace
		fmt.Fprintln(os.Stderr, "Warning: Replace mode will delete all existing data!")
	}

	entryRepo := sqlite.NewEntryRepository(services.DB)
	habitRepo := sqlite.NewHabitRepository(services.DB)
	habitLogRepo := sqlite.NewHabitLogRepository(services.DB)
	dayContextRepo := sqlite.NewDayContextRepository(services.DB)
	summaryRepo := sqlite.NewSummaryRepository(services.DB)
	listRepo := sqlite.NewListRepository(services.DB)
	listItemRepo := sqlite.NewListItemRepository(services.DB)
	goalRepo := sqlite.NewGoalRepository(services.DB)

	importSvc := service.NewImportService(
		entryRepo, habitRepo, habitLogRepo, dayContextRepo,
		summaryRepo, listRepo, listItemRepo, goalRepo,
	)

	opts := domain.NewImportOptions(mode)

	if err := importSvc.Import(cmd.Context(), &data, opts); err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Import complete:\n")
	fmt.Fprintf(os.Stderr, "  Entries:     %d\n", len(data.Entries))
	fmt.Fprintf(os.Stderr, "  Habits:      %d\n", len(data.Habits))
	fmt.Fprintf(os.Stderr, "  Habit Logs:  %d\n", len(data.HabitLogs))
	fmt.Fprintf(os.Stderr, "  Day Contexts: %d\n", len(data.DayContexts))
	fmt.Fprintf(os.Stderr, "  Summaries:   %d\n", len(data.Summaries))
	fmt.Fprintf(os.Stderr, "  Lists:       %d\n", len(data.Lists))
	fmt.Fprintf(os.Stderr, "  List Items:  %d\n", len(data.ListItems))
	fmt.Fprintf(os.Stderr, "  Goals:       %d\n", len(data.Goals))

	return nil
}
