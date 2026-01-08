package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var backupVerifyCmd = &cobra.Command{
	Use:   "verify <path>",
	Short: "Verify a backup file",
	Long: `Verify the integrity of a backup file.

Runs SQLite's integrity_check on the backup to ensure it's valid.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		err := backupService.VerifyBackup(cmd.Context(), path)
		if err != nil {
			red := color.New(color.FgRed).SprintFunc()
			return fmt.Errorf("%s: %w", red("verification failed"), err)
		}

		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s: %s\n", green("OK"), path)
		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupVerifyCmd)
}
