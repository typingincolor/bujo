package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/service"
)

var backupService *service.BackupService

func getDefaultBackupDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "backups"
	}
	return filepath.Join(home, ".bujo", "backups")
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage database backups",
	Long: `Manage database backups for your bujo data.

Without subcommands, shows all available backups.

Backups are stored in ~/.bujo/backups/ by default.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}
		backupService = service.NewBackupService(db, getDefaultBackupDir())
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		backups, err := backupService.ListBackups(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found. Create one with: bujo backup create")
			return nil
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		fmt.Println("Backups")
		fmt.Println(gray("---------------------------------------------------------"))

		for _, b := range backups {
			sizeStr := formatSize(b.Size)
			fmt.Printf("%s %s %s\n",
				cyan(b.Filename),
				gray(b.CreatedAt.Format("2006-01-02 15:04:05")),
				gray(sizeStr),
			)
		}

		return nil
	},
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
