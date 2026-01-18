package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <entity-id>",
	Short: "Restore a deleted entry",
	Long: `Restore a previously deleted entry by its entity ID.

Use 'bujo deleted' to see the list of deleted entries and their entity IDs.

Examples:
  bujo deleted              # see deleted entries
  bujo restore abc123def    # restore entry with entity ID abc123def`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := domain.EntityID(args[0])

		newID, err := bujoService.RestoreEntry(cmd.Context(), entityID)
		if err != nil {
			return fmt.Errorf("failed to restore entry: %w", err)
		}

		if newID == 0 {
			fmt.Fprintln(os.Stderr, "Entry is not deleted, nothing to restore.")
			return nil
		}

		fmt.Fprintf(os.Stderr, "Restored entry #%d\n", newID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
