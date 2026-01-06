package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	addLocation string
)

var addCmd = &cobra.Command{
	Use:   "add [entries...]",
	Short: "Add entries to today's journal",
	Long: `Add one or more entries to today's journal.

Entries can be provided as arguments or piped via stdin.

Entry types:
  . Task (todo item)
  - Note (information)
  o Event (scheduled occurrence)

Examples:
  bujo add ". Buy groceries"
  bujo add ". Task one" "- Note one"
  echo ". Task from pipe" | bujo add
  bujo add --location "Home Office" ". Work on project"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var input string

		// Check if input is piped
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Reading from pipe
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			input = strings.Join(lines, "\n")
		} else if len(args) > 0 {
			// Reading from arguments
			input = strings.Join(args, "\n")
		} else {
			return fmt.Errorf("no entries provided; use arguments or pipe input")
		}

		opts := service.LogEntriesOptions{
			Date: time.Now(),
		}

		if addLocation != "" {
			opts.Location = &addLocation
		}

		ids, err := bujoService.LogEntries(cmd.Context(), input, opts)
		if err != nil {
			return fmt.Errorf("failed to add entries: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Added %d entry(s)\n", len(ids))
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addLocation, "location", "l", "", "Set location for entries")
	rootCmd.AddCommand(addCmd)
}
