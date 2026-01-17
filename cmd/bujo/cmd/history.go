package cmd

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	historyService *service.HistoryService
	listItemRepo   *sqlite.ListItemRepository
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View and manage version history",
	Long: `View and manage version history for list items.

The event sourcing system keeps track of all changes to list items,
allowing you to see previous versions and restore them if needed.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}
		listItemRepo = sqlite.NewListItemRepository(services.DB)
		historyService = service.NewHistoryService(listItemRepo)
		return nil
	},
}

var historyShowCmd = &cobra.Command{
	Use:   "show <entity-id>",
	Short: "Show version history for an item",
	Long: `Display all versions of a list item.

The entity-id is the UUID that identifies the item across versions.
You can find entity IDs in the database or from list item details.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := domain.EntityID(args[0])

		history, err := historyService.GetItemHistory(cmd.Context(), entityID)
		if err != nil {
			return fmt.Errorf("failed to get history: %w", err)
		}

		if len(history) == 0 {
			fmt.Println("No history found for this entity.")
			return nil
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf("History for %s\n", cyan(string(entityID)))
		fmt.Println(gray("---------------------------------------------------------"))

		for _, item := range history {
			status := ""
			if item.IsCurrent() {
				status = green(" (current)")
			} else if item.IsDeleted() {
				status = yellow(" (deleted)")
			}

			fmt.Printf("v%d %s %s%s\n",
				item.Version,
				gray(item.ValidFrom.Format("2006-01-02 15:04:05")),
				item.Content,
				status,
			)
		}

		return nil
	},
}

var historyRestoreCmd = &cobra.Command{
	Use:   "restore <entity-id> <version>",
	Short: "Restore an item to a previous version",
	Long: `Restore a list item to a previous version.

This creates a new version with the content from the specified version.
The original history is preserved.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := domain.EntityID(args[0])

		version, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid version number: %s", args[1])
		}

		err = historyService.RestoreItem(cmd.Context(), entityID, version)
		if err != nil {
			return fmt.Errorf("failed to restore: %w", err)
		}

		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Restored to version %d\n", green("OK"), version)
		return nil
	},
}

func init() {
	historyCmd.AddCommand(historyShowCmd)
	historyCmd.AddCommand(historyRestoreCmd)
	rootCmd.AddCommand(historyCmd)
}
