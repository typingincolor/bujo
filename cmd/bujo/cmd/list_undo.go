package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listUndoCmd = &cobra.Command{
	Use:   "undo <item-id>",
	Short: "Mark a list item as not done",
	Long: `Revert a completed item back to a task.

Example:
  bujo list undo 42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		itemID, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = listService.MarkUndone(ctx, itemID)
		if err != nil {
			return fmt.Errorf("failed to mark undone: %w", err)
		}

		fmt.Println("Item marked undone")
		return nil
	},
}

func init() {
	listCmd.AddCommand(listUndoCmd)
}
