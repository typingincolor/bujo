package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listMoveCmd = &cobra.Command{
	Use:   "move <item-id> <target-list>",
	Short: "Move an item to another list",
	Long: `Move an item from one list to another.

Examples:
  bujo list move 42 Work
  bujo list move 42 #2`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		itemID, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		targetListID, err := resolveListID(ctx, args[1])
		if err != nil {
			return err
		}

		err = listService.MoveItem(ctx, itemID, targetListID)
		if err != nil {
			return fmt.Errorf("failed to move item: %w", err)
		}

		fmt.Println("Item moved")
		return nil
	},
}

func init() {
	listCmd.AddCommand(listMoveCmd)
}
