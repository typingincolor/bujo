package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var migrateTo string
var migrateYes bool

var migrateCmd = &cobra.Command{
	Use:   "migrate <id> --to <date>",
	Short: "Migrate a task to a future date",
	Long: `Migrate a task to a future date.

The original entry is marked as migrated (→) and a new task
is created on the target date.

Only tasks can be migrated (not notes or events).

Examples:
  bujo migrate 42 --to tomorrow
  bujo migrate 1 --to "next monday"
  bujo migrate 5 --to 2026-01-15`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrateTo == "" {
			return fmt.Errorf("--to flag is required")
		}

		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		toDate, err := parseFutureDate(migrateTo)
		if err != nil {
			return err
		}

		toDate, err = confirmDate(migrateTo, toDate, migrateYes)
		if err != nil {
			return err
		}

		newID, err := services.Bujo.MigrateEntry(cmd.Context(), id, toDate)
		if err != nil {
			return fmt.Errorf("failed to migrate entry: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Migrated entry #%d → #%d (scheduled for %s)\n",
			id, newID, toDate.Format("Jan 2, 2006"))
		return nil
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrateTo, "to", "", "Target date (e.g., 'tomorrow', 'next monday', '2026-01-15')")
	migrateCmd.Flags().BoolVarP(&migrateYes, "yes", "y", false, "Skip date confirmation prompt")
	rootCmd.AddCommand(migrateCmd)
}
