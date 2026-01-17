package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Manage lists",
	Long: `Manage lists for organizing items outside the daily journal.

Lists are separate from your daily entries and useful for things like
shopping lists, project backlogs, or any collection of items.

Without subcommands, shows all lists with their progress.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lists, err := services.List.GetAllLists(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get lists: %w", err)
		}

		if len(lists) == 0 {
			fmt.Println("No lists yet. Create one with: bujo list create <name>")
			return nil
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		fmt.Println("Lists")
		fmt.Println(gray("---------------------------------------------------------"))

		for _, list := range lists {
			summary, err := services.List.GetListSummary(cmd.Context(), list.ID)
			if err != nil {
				return err
			}

			progress := ""
			if summary.TotalItems > 0 {
				progress = fmt.Sprintf(" %d/%d done", summary.DoneItems, summary.TotalItems)
			}

			fmt.Printf("#%d %s%s\n", list.ID, cyan(list.Name), gray(progress))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
