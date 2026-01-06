package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/cli"
	"github.com/typingincolor/bujo/internal/domain"
)

var viewAncestors int

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "View an entry with context",
	Long: `View an entry with its parent and siblings for context.

By default shows the parent entry and all its children.
Use --up to go further up the hierarchy.

Examples:
  bujo view 42
  bujo view 42 --up 1    # Show grandparent context
  bujo view 42 -u 2      # Show great-grandparent context`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseEntryID(args[0])
		if err != nil {
			return err
		}

		entries, err := bujoService.GetEntryContext(cmd.Context(), id, viewAncestors)
		if err != nil {
			return fmt.Errorf("failed to get entry: %w", err)
		}

		fmt.Print(renderViewTree(entries, id))
		return nil
	},
}

func init() {
	viewCmd.Flags().IntVarP(&viewAncestors, "up", "u", 0, "Number of additional ancestor levels to show")
	rootCmd.AddCommand(viewCmd)
}

func renderViewTree(entries []domain.Entry, highlightID int64) string {
	var sb strings.Builder

	// Build set of IDs in result and parent-child map
	inResult := make(map[int64]bool)
	for _, e := range entries {
		inResult[e.ID] = true
	}

	children := make(map[int64][]domain.Entry)
	var roots []domain.Entry

	for _, e := range entries {
		// Entry is a root if it has no parent OR its parent isn't in result set
		if e.ParentID == nil || !inResult[*e.ParentID] {
			roots = append(roots, e)
		} else {
			children[*e.ParentID] = append(children[*e.ParentID], e)
		}
	}

	for _, root := range roots {
		renderViewEntry(&sb, root, children, 0, highlightID)
	}

	return sb.String()
}

func renderViewEntry(sb *strings.Builder, entry domain.Entry, children map[int64][]domain.Entry, depth int, highlightID int64) {
	indent := strings.Repeat("  ", depth)
	prefix := ""
	if depth > 0 {
		prefix = "└── "
	}

	symbol := entry.Type.Symbol()
	idStr := fmt.Sprintf("(%d)", entry.ID)
	content := entry.Content

	// Highlight the requested entry
	if entry.ID == highlightID {
		idStr = cli.Highlight(idStr)
		content = cli.Highlight(content)
		symbol = cli.Highlight(symbol)
	} else {
		switch entry.Type {
		case domain.EntryTypeDone:
			content = cli.Green(content)
			symbol = cli.Green(symbol)
			idStr = cli.Green(idStr)
		case domain.EntryTypeMigrated:
			content = cli.Dimmed(content)
			symbol = cli.Dimmed(symbol)
			idStr = cli.Dimmed(idStr)
		}
	}

	fmt.Fprintf(sb, "%s%s%s %s %s\n", indent, prefix, symbol, content, idStr)

	for _, child := range children[entry.ID] {
		renderViewEntry(sb, child, children, depth+1, highlightID)
	}
}
