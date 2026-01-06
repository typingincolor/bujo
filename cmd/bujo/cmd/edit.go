package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id> <new-content>",
	Short: "Edit an entry's content",
	Long: `Edit the content of an existing entry.

Examples:
  bujo edit 42 "Buy milk instead"
  bujo edit 1 "Updated task description"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid entry ID: %s", args[0])
		}

		newContent := strings.Join(args[1:], " ")

		err = bujoService.EditEntry(cmd.Context(), id, newContent)
		if err != nil {
			return fmt.Errorf("failed to edit entry: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Updated entry #%d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
