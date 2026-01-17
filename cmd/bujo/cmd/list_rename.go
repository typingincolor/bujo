package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listRenameCmd = &cobra.Command{
	Use:   "rename <list> <new-name>",
	Short: "Rename a list",
	Long: `Rename a list.

Examples:
  bujo list rename Shopping Groceries
  bujo list rename #1 "New Name"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		listID, err := resolveListID(ctx, args[0])
		if err != nil {
			return err
		}

		newName := strings.Join(args[1:], " ")

		err = services.List.RenameList(ctx, listID, newName)
		if err != nil {
			return fmt.Errorf("failed to rename list: %w", err)
		}

		fmt.Printf("Renamed list to: %s\n", newName)
		return nil
	},
}

func init() {
	listCmd.AddCommand(listRenameCmd)
}
