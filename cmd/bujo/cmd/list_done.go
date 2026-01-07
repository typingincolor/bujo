package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listDoneCmd = &cobra.Command{
	Use:   "done <item-id>",
	Short: "Mark a list item as done",
	Long: `Mark a list item as completed.

Example:
  bujo list done 42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		itemID, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		err = listService.MarkDone(ctx, itemID)
		if err != nil {
			return fmt.Errorf("failed to mark done: %w", err)
		}

		fmt.Println("Item marked done")
		return nil
	},
}

func init() {
	listCmd.AddCommand(listDoneCmd)
}
