package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	addLocation string
	addDate     string
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
  bujo add --at "Home Office" ". Work on project"
  bujo add --date yesterday ". Forgot to log this"
  bujo add -d "last monday" ". Backfill task"

`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Manually parse flags to allow entries starting with '-'
		var entries []string
		for i := 0; i < len(args); i++ {
			arg := args[i]
			switch {
			case arg == "-a" || arg == "--at":
				if i+1 < len(args) {
					addLocation = args[i+1]
					i++
				}
			case arg == "-d" || arg == "--date":
				if i+1 < len(args) {
					addDate = args[i+1]
					i++
				}
			case strings.HasPrefix(arg, "-a="):
				addLocation = arg[3:]
			case strings.HasPrefix(arg, "--at="):
				addLocation = arg[5:]
			case strings.HasPrefix(arg, "-d="):
				addDate = arg[3:]
			case strings.HasPrefix(arg, "--date="):
				addDate = arg[7:]
			case arg == "-h" || arg == "--help":
				return cmd.Help()
			case arg == "--":
				entries = append(entries, args[i+1:]...)
				i = len(args) // break loop
			// Skip global flags (handled by parent)
			case arg == "--db-path":
				if i+1 < len(args) {
					i++ // skip value
				}
			case strings.HasPrefix(arg, "--db-path="):
				// skip
			case arg == "-v" || arg == "--verbose":
				// skip
			default:
				entries = append(entries, arg)
			}
		}

		var input string

		// Prefer args over stdin if args were provided
		if len(entries) > 0 {
			input = strings.Join(entries, "\n")
		} else {
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
			}
		}

		if input == "" {
			return fmt.Errorf("no entries provided; use arguments or pipe input")
		}

		date, err := parseDateOrToday(addDate)
		if err != nil {
			return err
		}

		opts := service.LogEntriesOptions{
			Date: date,
		}

		if addLocation != "" {
			opts.Location = &addLocation
		}

		ids, err := bujoService.LogEntries(cmd.Context(), input, opts)
		if err != nil {
			return fmt.Errorf("failed to add entries: %w", err)
		}

		// Print IDs to stdout for scripting
		for _, id := range ids {
			fmt.Println(id)
		}

		fmt.Fprintf(os.Stderr, "Added %d entry(s)\n", len(ids))
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addLocation, "at", "a", "", "Set location for entries")
	addCmd.Flags().StringVarP(&addDate, "date", "d", "", "Date to add entries (e.g., 'yesterday', '2026-01-01')")
	rootCmd.AddCommand(addCmd)
}
