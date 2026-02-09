package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
	"github.com/typingincolor/bujo/internal/domain"
)

var (
	searchFrom     string
	searchTo       string
	searchType     string
	searchLimit    int
	searchTags     string
	searchMentions string
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search through entries",
	Long: `Search through entries by content, tags, or both.

Supports optional filters for date range, entry type, and tags.

Examples:
  bujo search "groceries"                    # Search all entries
  bujo search "meeting" --from "last month"  # With date range
  bujo search "project" --type task          # Filter by type
  bujo search "call" -f "last week" -t today # Date range filter
  bujo search "report" -n 10                 # Limit results
  bujo search --tag shopping,errands         # Search by tags
  bujo search "milk" --tag shopping          # Combined search
  bujo search --mention john                 # Search by @mention
  bujo search "meeting" --mention john.smith # Combined with mention`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var query string
		if len(args) > 0 {
			query = args[0]
		}

		if query == "" && searchTags == "" && searchType == "" && searchMentions == "" {
			return fmt.Errorf("provide a search query, --tag filter, --type filter, or --mention filter")
		}

		opts := domain.NewSearchOptions(query)

		if searchType != "" {
			entryType := domain.EntryType(searchType)
			if !entryType.IsValid() {
				return fmt.Errorf("invalid entry type: %s (valid types: task, note, event, done, migrated, cancelled, question, answered, answer)", searchType)
			}
			opts = opts.WithType(entryType)
		}

		if searchTags != "" {
			tags := strings.Split(searchTags, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			opts = opts.WithTags(tags)
		}

		if searchMentions != "" {
			mentions := strings.Split(searchMentions, ",")
			for i := range mentions {
				mentions[i] = strings.TrimSpace(mentions[i])
			}
			opts = opts.WithMentions(mentions)
		}

		if searchFrom != "" || searchTo != "" {
			var from, to time.Time
			if searchFrom != "" {
				parsed, err := parsePastDate(searchFrom)
				if err != nil {
					return fmt.Errorf("invalid --from date: %w", err)
				}
				from = parsed
			} else {
				from = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			}
			if searchTo != "" {
				parsed, err := parsePastDate(searchTo)
				if err != nil {
					return fmt.Errorf("invalid --to date: %w", err)
				}
				to = parsed
			} else {
				to = time.Now()
			}
			opts = opts.WithDateRange(from, to)
		}

		if searchLimit > 0 {
			opts = opts.WithLimit(searchLimit)
		}

		results, err := bujoService.SearchEntries(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) == 0 {
			if query != "" {
				fmt.Printf("No results found for %q\n", query)
			} else {
				fmt.Println("No results found")
			}
			return nil
		}

		ids := make([]int64, len(results))
		for i, entry := range results {
			ids[i] = entry.ID
		}
		ancestorsMap, err := bujoService.GetEntriesAncestorsMap(cmd.Context(), ids)
		if err != nil {
			return fmt.Errorf("failed to fetch ancestors: %w", err)
		}

		for _, entry := range results {
			ancestors := ancestorsMap[entry.ID]
			fmt.Println(formatSearchResultWithContext(entry, ancestors, query))
		}
		if query != "" {
			fmt.Printf("\nFound %d result(s) for %q\n", len(results), query)
		} else {
			fmt.Printf("\nFound %d result(s)\n", len(results))
		}

		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchFrom, "from", "f", "", "Start date for search (natural language supported)")
	searchCmd.Flags().StringVarP(&searchTo, "to", "t", "", "End date for search (natural language supported)")
	searchCmd.Flags().StringVar(&searchType, "type", "", "Filter by entry type (task, note, event, done, migrated, cancelled)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 50, "Maximum number of results")
	searchCmd.Flags().StringVar(&searchTags, "tag", "", "Filter by tags (comma-separated, e.g. shopping,errands)")
	searchCmd.Flags().StringVar(&searchMentions, "mention", "", "Filter by @mentions (comma-separated, e.g. john,alice.smith)")
	rootCmd.AddCommand(searchCmd)
}

func formatSearchResultWithContext(entry domain.Entry, ancestors []domain.Entry, query string) string {
	var lines []string

	if len(ancestors) > 0 {
		var contextParts []string
		for _, a := range ancestors {
			contextParts = append(contextParts, a.Content)
		}
		contextLine := cli.Dimmed("  ↳ " + strings.Join(contextParts, " > "))
		lines = append(lines, contextLine)
	}

	var parts []string

	if entry.ScheduledDate != nil {
		parts = append(parts, cli.Dimmed(fmt.Sprintf("[%s]", entry.ScheduledDate.Format("2006-01-02"))))
	} else {
		parts = append(parts, cli.Dimmed("[no date]"))
	}

	symbol := entry.Type.Symbol()
	content := entry.Content

	switch entry.Type {
	case domain.EntryTypeDone, domain.EntryTypeAnswered:
		symbol = cli.Green(symbol)
		content = cli.Green(content)
	case domain.EntryTypeMigrated, domain.EntryTypeCancelled:
		symbol = cli.Dimmed(symbol)
		content = cli.Dimmed(content)
	}

	content = highlightQuery(content, query)

	parts = append(parts, symbol)
	parts = append(parts, content)
	parts = append(parts, cli.Dimmed(fmt.Sprintf("(%d)", entry.ID)))

	lines = append(lines, strings.Join(parts, " "))

	return strings.Join(lines, "\n")
}

func highlightQuery(content, query string) string {
	lower := strings.ToLower(content)
	queryLower := strings.ToLower(query)
	idx := strings.Index(lower, queryLower)
	if idx == -1 {
		return content
	}

	before := content[:idx]
	match := content[idx : idx+len(query)]
	after := content[idx+len(query):]

	return before + cli.Highlight(match) + after
}
