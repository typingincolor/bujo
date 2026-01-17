package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/service"
)

var addCmd = &cobra.Command{
	Use:   "add [entries...]",
	Short: "Add entries to today's journal",
	Long: `Add one or more entries to today's journal (or a specific date).

Entries can be provided as arguments or piped via stdin.

Entry types:
  . Task (todo item)
  - Note (information)
  o Event (scheduled occurrence)

Examples:
  bujo add ". Buy groceries"
  bujo add ". Task one" "- Note one"
  echo ". Task from pipe" | bujo add
  bujo add --file tasks.txt
  bujo add -f tasks.txt --at "Home Office"
  bujo add --at "Home Office" ". Work on project"
  bujo add --date yesterday ". Forgot to log this"
  bujo add -d "last monday" ". Backfill task"
  bujo add -d "next friday" -y ". Skip confirmation"
  bujo add --parent 123 ". Add as child of entry 123"

`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, addLocation, addDate, addFile, addParent, showHelp, skipConfirm := parseAddArgs(args)
		if showHelp {
			return cmd.Help()
		}

		var input string

		if addFile != "" {
			content, err := os.ReadFile(addFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			input = string(content)
		} else if len(entries) > 0 {
			input = strings.Join(entries, "\n")
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				scanner := bufio.NewScanner(os.Stdin)
				var lines []string
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
				}
				input = strings.Join(lines, "\n")
			}
		}

		if input == "" {
			return fmt.Errorf("no entries provided; use arguments, --file, or pipe input")
		}

		date, err := parseDateOrToday(addDate)
		if err != nil {
			return err
		}

		if addDate != "" {
			date, err = confirmDate(addDate, date, skipConfirm)
			if err != nil {
				return err
			}
		}

		opts := service.LogEntriesOptions{
			Date: date,
		}

		if addLocation != "" {
			opts.Location = &addLocation
		}

		if addParent != "" {
			parentID, err := parseEntryID(addParent)
			if err != nil {
				return fmt.Errorf("invalid parent ID: %w", err)
			}
			opts.ParentID = &parentID
		}

		ids, err := services.Bujo.LogEntries(cmd.Context(), input, opts)
		if err != nil {
			return fmt.Errorf("failed to add entries: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Added %d entry(s)\n", len(ids))
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("at", "a", "", "Set location for entries")
	addCmd.Flags().StringP("date", "d", "", "Date to add entries (e.g., 'yesterday', '2026-01-01')")
	addCmd.Flags().StringP("file", "f", "", "Read entries from file")
	addCmd.Flags().StringP("parent", "p", "", "Add entries as children of specified entry ID")
	addCmd.Flags().BoolP("yes", "y", false, "Skip date confirmation prompt")
	rootCmd.AddCommand(addCmd)
}
