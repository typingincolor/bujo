package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new list",
	Long: `Create a new list with the given name.

Names can include spaces if quoted:
  bujo list create "Shopping List"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		list, err := services.List.CreateList(cmd.Context(), name)
		if err != nil {
			return fmt.Errorf("failed to create list: %w", err)
		}

		fmt.Printf("Created list #%d: %s\n", list.ID, list.Name)
		return nil
	},
}

func init() {
	listCmd.AddCommand(listCreateCmd)
}
