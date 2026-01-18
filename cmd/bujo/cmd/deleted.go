package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var deletedCmd = &cobra.Command{
	Use:   "deleted",
	Short: "List deleted entries that can be restored",
	Long: `List entries that have been deleted but can still be restored.

Each entry shows its entity ID which can be used with 'bujo restore' to bring it back.

Examples:
  bujo deleted
  bujo restore abc123  # restore using entity ID from list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := bujoService.GetDeletedEntries(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get deleted entries: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No deleted entries found.")
			return nil
		}

		gray := color.New(color.FgHiBlack).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()

		fmt.Printf("Deleted entries (%d):\n\n", len(entries))
		for _, entry := range entries {
			dateStr := ""
			if entry.ScheduledDate != nil {
				dateStr = entry.ScheduledDate.Format("2006-01-02")
			}
			fmt.Printf("  %s %s  %s %s\n",
				cyan(entry.EntityID.String()),
				string(entry.Type),
				entry.Content,
				gray(dateStr),
			)
		}

		fmt.Printf("\n%s\n", gray("Use 'bujo restore <entity-id>' to restore an entry."))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deletedCmd)
}
