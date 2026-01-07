package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listDeleteForce bool

var listDeleteCmd = &cobra.Command{
	Use:   "delete <list>",
	Short: "Delete a list",
	Long: `Delete a list by name or ID (#1).

If the list has items, use --force to delete them too:
  bujo list delete Shopping --force
  bujo list delete #1 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		listID, err := resolveListID(ctx, args[0])
		if err != nil {
			return err
		}

		list, err := listService.GetListByID(ctx, listID)
		if err != nil {
			return err
		}

		err = listService.DeleteList(ctx, listID, listDeleteForce)
		if err != nil {
			return fmt.Errorf("failed to delete list: %w", err)
		}

		fmt.Printf("Deleted list: %s\n", list.Name)
		return nil
	},
}

func init() {
	listDeleteCmd.Flags().BoolVarP(&listDeleteForce, "force", "f", false, "Delete list and all its items")
	listCmd.AddCommand(listDeleteCmd)
}
