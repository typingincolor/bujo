package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
)

var editPriority string

var editCmd = &cobra.Command{
	Use:   "edit <id> [new-content]",
	Short: "Edit an entry's content or priority",
	Long: `Edit the content or priority of an existing entry.

Priority levels:
  none    - No priority (default)
  low     - Low priority (!)
  medium  - Medium priority (!!)
  high    - High priority (!!!)

Examples:
  bujo edit 42 "Buy milk instead"
  bujo edit 1 --priority high
  bujo edit 5 "New content" --priority medium`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		hasContent := len(args) > 1
		hasPriority := editPriority != ""

		if !hasContent && !hasPriority {
			return fmt.Errorf("provide new content or --priority flag")
		}

		if hasContent {
			newContent := strings.Join(args[1:], " ")
			err = services.Bujo.EditEntry(cmd.Context(), id, newContent)
			if err != nil {
				return fmt.Errorf("failed to edit entry: %w", err)
			}
		}

		if hasPriority {
			priority, err := domain.ParsePriority(editPriority)
			if err != nil {
				return err
			}
			err = services.Bujo.EditEntryPriority(cmd.Context(), id, priority)
			if err != nil {
				return fmt.Errorf("failed to update priority: %w", err)
			}
		}

		fmt.Fprintf(os.Stderr, "✓ Updated entry #%d\n", id)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&editPriority, "priority", "p", "", "Set priority (none, low, medium, high)")
	rootCmd.AddCommand(editCmd)
}
