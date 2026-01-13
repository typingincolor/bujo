package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
)

var listAddCmd = &cobra.Command{
	Use:   "add <list> <content>",
	Short: "Add an item to a list",
	Long: `Add a new item to a list.

Items are added as tasks by default. Use entry type symbols to specify type:
  . Task (default)
  - Note
  o Event

Examples:
  bujo list add Shopping "Buy milk"
  bujo list add #1 ". Buy bread"
  bujo list add Work "- Important note"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		listID, err := resolveListID(ctx, args[0])
		if err != nil {
			return err
		}

		content := strings.Join(args[1:], " ")
		entryType := domain.EntryTypeTask

		if len(content) >= 2 {
			prefix := content[:2]
			switch prefix {
			case ". ":
				entryType = domain.EntryTypeTask
				content = content[2:]
			case "- ":
				entryType = domain.EntryTypeNote
				content = content[2:]
			case "o ":
				entryType = domain.EntryTypeEvent
				content = content[2:]
			case "x ":
				entryType = domain.EntryTypeDone
				content = content[2:]
			}
		}

		id, err := listService.AddItem(ctx, listID, entryType, content)
		if err != nil {
			return fmt.Errorf("failed to add item: %w", err)
		}

		fmt.Println(id)
		return nil
	},
}

func init() {
	listCmd.AddCommand(listAddCmd)
}
