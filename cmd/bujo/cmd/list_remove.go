package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listRemoveCmd = &cobra.Command{
	Use:   "remove <item-id>",
	Short: "Remove an item from a list",
	Long: `Remove an item from a list by its entry ID.

Use 'list show' to see item IDs.

Example:
  bujo list remove 42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		itemID, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = listService.RemoveItem(ctx, itemID)
		if err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}

		fmt.Println("Item removed")
		return nil
	},
}

func init() {
	listCmd.AddCommand(listRemoveCmd)
}
