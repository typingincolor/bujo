package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	archiveService   *service.ArchiveService
	archiveOlderThan string
	archiveExecute   bool
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive old data versions",
	Long: `Archive old versions of data to reduce database size.

By default, shows how many records would be archived (dry run).
Use --execute to actually perform the archive operation.

Old versions are rows with valid_to set, meaning they've been superseded.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}
		listItemRepo := sqlite.NewListItemRepository(services.DB)
		archiveService = service.NewArchiveService(listItemRepo)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cutoff, err := parseArchiveCutoff(archiveOlderThan)
		if err != nil {
			return err
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		if archiveExecute {
			deleted, err := archiveService.Archive(cmd.Context(), cutoff)
			if err != nil {
				return fmt.Errorf("archive failed: %w", err)
			}

			fmt.Printf("%s Archived %s old version(s)\n", green("OK"), cyan(fmt.Sprintf("%d", deleted)))
			return nil
		}

		count, err := archiveService.GetArchivableCount(cmd.Context(), cutoff)
		if err != nil {
			return fmt.Errorf("failed to count archivable records: %w", err)
		}

		if count == 0 {
			fmt.Println("No records to archive.")
			return nil
		}

		fmt.Printf("Found %s old version(s) to archive %s\n",
			cyan(fmt.Sprintf("%d", count)),
			gray(fmt.Sprintf("(older than %s)", cutoff.Format("2006-01-02"))),
		)
		fmt.Println(gray("Use --execute to perform the archive operation."))

		return nil
	},
}

func parseArchiveCutoff(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}

	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", s)
	}
	return parsed, nil
}

func init() {
	archiveCmd.Flags().StringVar(&archiveOlderThan, "older-than", "", "Archive versions older than this date (YYYY-MM-DD)")
	archiveCmd.Flags().BoolVar(&archiveExecute, "execute", false, "Actually perform the archive (default is dry run)")
	rootCmd.AddCommand(archiveCmd)
}
