package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/domain"
)

var retypeCmd = &cobra.Command{
	Use:   "retype <id> <type>",
	Short: "Change an entry's type",
	Long: `Change an entry's type (task, note, or event).

This is useful when you create an entry with the wrong type
and want to fix it without recreating it.

Valid types:
  task  (•) - A task to be done
  note  (–) - A note or observation
  event (○) - An event or appointment

Examples:
  bujo retype 42 note     # Change to note
  bujo retype 15 task     # Change to task
  bujo retype 23 event    # Change to event`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		newType, err := parseEntryType(args[1])
		if err != nil {
			return err
		}

		err = services.Bujo.RetypeEntry(cmd.Context(), id, newType)
		if err != nil {
			return fmt.Errorf("failed to retype: %w", err)
		}

		fmt.Fprintf(os.Stderr, "%s Changed entry %d to %s\n", newType.Symbol(), id, newType)
		return nil
	},
}

func parseEntryType(s string) (domain.EntryType, error) {
	return domain.ParseEntryTypeFromString(strings.ToLower(s))
}

func init() {
	rootCmd.AddCommand(retypeCmd)
}
