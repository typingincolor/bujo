package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
)

var listShowCmd = &cobra.Command{
	Use:   "show <list>",
	Short: "Show items in a list",
	Long: `Display all items in a list.

Examples:
  bujo list show Shopping
  bujo list show #1`,
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

		items, err := listService.GetListItems(ctx, listID)
		if err != nil {
			return fmt.Errorf("failed to get list items: %w", err)
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("#%d %s\n", list.ID, cyan(list.Name))
		fmt.Println(gray("---------------------------------------------------------"))

		if len(items) == 0 {
			fmt.Println(gray("No items"))
			return nil
		}

		for _, item := range items {
			symbol := item.Type.Symbol()
			content := item.Content

			if item.Type == domain.EntryTypeDone {
				content = green(content)
			}

			fmt.Printf("%s %s %s\n", gray(fmt.Sprintf("(%d)", item.ID)), symbol, content)
		}

		summary, _ := listService.GetListSummary(ctx, listID)
		if summary != nil && summary.TotalItems > 0 {
			fmt.Println(gray("---------------------------------------------------------"))
			fmt.Printf("%s\n", gray(fmt.Sprintf("%d/%d done", summary.DoneItems, summary.TotalItems)))
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listShowCmd)
}
