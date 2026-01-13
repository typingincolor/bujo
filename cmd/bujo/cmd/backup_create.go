package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new backup",
	Long: `Create a backup of the current database.

Backups use SQLite's VACUUM INTO for a consistent snapshot.
The backup file is stored in ~/.bujo/backups/ with a timestamp.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := backupService.CreateBackup(cmd.Context(), backupDir)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		fmt.Printf("Backup created: %s\n", path)
		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
}
