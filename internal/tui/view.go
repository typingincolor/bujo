package tui

import (
	"fmt"
	"strings"

	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func (m Model) View() string {
	if m.err != nil {
		return m.renderErrorPopup()
	}

	if m.captureMode.active {
		return m.renderCaptureMode()
	}

	var sb strings.Builder

	// Toolbar
	toolbar := m.renderToolbar()
	sb.WriteString(toolbar)
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(m.width, 60)))
	sb.WriteString("\n")

	// View-specific content
	switch m.currentView {
	case ViewTypeHabits:
		sb.WriteString(m.renderHabitsContent())
	case ViewTypeLists, ViewTypeListItems:
		sb.WriteString(m.renderListsContent())
	default:
		sb.WriteString(m.renderJournalContent())
	}

	// Modal overlays (shared across all views)
	if m.editMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderEditInput())
		sb.WriteString("\n")
	} else if m.addMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderAddInput())
		sb.WriteString("\n")
	} else if m.migrateMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderMigrateInput())
		sb.WriteString("\n")
	} else if m.confirmMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderConfirmDialog())
		sb.WriteString("\n")
	} else if m.gotoMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderGotoInput())
		sb.WriteString("\n")
	} else if m.searchMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderSearchInput())
		sb.WriteString("\n")
	} else if m.commandPalette.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderCommandPalette())
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(HelpStyle.Render(m.help.View(m.keyMap)))

	return sb.String()
}

func (m Model) renderJournalContent() string {
	if m.agenda == nil {
		return "Loading..."
	}

	var sb strings.Builder

	if len(m.entries) == 0 {
		sb.WriteString(HelpStyle.Render("No entries for the last 7 days."))
		sb.WriteString("\n\n")
	} else {
		// Calculate available lines (reserve for toolbar, help bar and padding)
		availableLines := m.height - 6 // 2 for toolbar, 2 for help, 2 for padding
		if availableLines < 5 {
			availableLines = 5
		}

		// Show scroll indicator if there's content above
		if m.scrollOffset > 0 {
			sb.WriteString(HelpStyle.Render(fmt.Sprintf("  ↑ %d more above", m.scrollOffset)))
			sb.WriteString("\n")
			availableLines--
		}

		// Reserve line for "more below" indicator
		reserveForBelow := 1

		// Render entries, counting lines used
		linesUsed := 0
		endIdx := m.scrollOffset
		for i := m.scrollOffset; i < len(m.entries); i++ {
			item := m.entries[i]
			linesNeeded := 1 // entry line
			if item.DayHeader != "" {
				linesNeeded += 2 // header + blank line before (except first)
				if i == m.scrollOffset {
					linesNeeded = 2 // no blank line before first header
				}
			}

			// Check if we have room (leave space for "more below" if not at end)
			spaceNeeded := linesNeeded
			if i < len(m.entries)-1 {
				spaceNeeded += reserveForBelow
			}
			if linesUsed+spaceNeeded > availableLines {
				break
			}

			// Render this entry
			if item.DayHeader != "" {
				if i > m.scrollOffset {
					sb.WriteString("\n")
					linesUsed++
				}
				if item.IsOverdue {
					sb.WriteString(OverdueHeaderStyle.Render(item.DayHeader))
				} else {
					sb.WriteString(DateHeaderStyle.Render(item.DayHeader))
				}
				sb.WriteString("\n")
				linesUsed++
			}

			line := m.renderEntry(item)
			// Highlight search matches
			if m.searchMode.active && m.searchMode.query != "" {
				line = m.highlightSearchTerm(line)
			}
			if i == m.selectedIdx {
				line = SelectedStyle.Render(line)
			}
			sb.WriteString(line)
			sb.WriteString("\n")
			linesUsed++
			endIdx = i + 1
		}

		// Show scroll indicator if there's content below
		if endIdx < len(m.entries) {
			sb.WriteString(HelpStyle.Render(fmt.Sprintf("  ↓ %d more below", len(m.entries)-endIdx)))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (m Model) renderHabitsContent() string {
	var sb strings.Builder

	if len(m.habitState.habits) == 0 {
		sb.WriteString(HelpStyle.Render("No habits yet. Use 'bujo habit log <name>' to create one."))
		sb.WriteString("\n\n")
		return sb.String()
	}

	for i, habit := range m.habitState.habits {
		// Build sparkline from day history
		sparkline := m.renderSparkline(habit.DayHistory)

		// Format: Name | Sparkline | Streak | Completion%
		line := fmt.Sprintf("%-20s %s  %d day streak  %.0f%%",
			habit.Name,
			sparkline,
			habit.CurrentStreak,
			habit.CompletionPercent,
		)

		if i == m.habitState.selectedIdx {
			line = SelectedStyle.Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (m Model) renderSparkline(history []service.DayStatus) string {
	var sb strings.Builder
	for _, day := range history {
		if day.Completed {
			sb.WriteString("▓")
		} else {
			sb.WriteString("░")
		}
	}
	return sb.String()
}

func (m Model) renderListsContent() string {
	if m.currentView == ViewTypeListItems {
		return m.renderListItemsContent()
	}
	return m.renderListsOverview()
}

func (m Model) renderListsOverview() string {
	var sb strings.Builder

	if len(m.listState.lists) == 0 {
		sb.WriteString(HelpStyle.Render("No lists yet. Use 'bujo list create <name>' to create one."))
		sb.WriteString("\n\n")
		return sb.String()
	}

	for i, list := range m.listState.lists {
		summary := m.listState.summaries[list.ID]
		var progress string
		if summary != nil {
			progress = fmt.Sprintf("%d/%d", summary.DoneItems, summary.TotalItems)
		} else {
			progress = "0/0"
		}

		line := fmt.Sprintf("📋 %-20s  %s", list.Name, progress)

		if i == m.listState.selectedListIdx {
			line = SelectedStyle.Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (m Model) renderListItemsContent() string {
	var sb strings.Builder

	var listName string
	for _, list := range m.listState.lists {
		if list.ID == m.listState.currentListID {
			listName = list.Name
			break
		}
	}
	sb.WriteString(fmt.Sprintf("📋 %s\n", listName))
	sb.WriteString("────────────────────────────────────────\n")

	if len(m.listState.items) == 0 {
		sb.WriteString(HelpStyle.Render("No items yet. List is empty. Press 'a' to add an item."))
		sb.WriteString("\n\n")
		return sb.String()
	}

	for i, item := range m.listState.items {
		symbol := item.Type.Symbol()
		line := fmt.Sprintf("%s %s", symbol, item.Content)

		if item.Type == domain.ListItemTypeDone {
			line = DoneStyle.Render(line)
		}

		if i == m.listState.selectedItemIdx {
			line = SelectedStyle.Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (m Model) renderEntry(item EntryItem) string {
	entry := item.Entry
	indent := strings.Repeat("  ", item.Indent)

	symbol := entry.Type.Symbol()
	content := entry.Content

	base := fmt.Sprintf("%s%s %s", indent, symbol, content)

	switch entry.Type {
	case domain.EntryTypeDone:
		return DoneStyle.Render(base)
	case domain.EntryTypeMigrated:
		return MigratedStyle.Render(base)
	default:
		if item.IsOverdue {
			return OverdueStyle.Render(base)
		}
		return base
	}
}

func (m Model) renderConfirmDialog() string {
	dialog := `Delete entry with children?

  y - Yes, delete all
  n - No, cancel`

	return ConfirmStyle.Render(dialog)
}

func (m Model) renderEditInput() string {
	var sb strings.Builder
	sb.WriteString("Edit entry:\n")
	sb.WriteString(m.editMode.input.View())
	sb.WriteString("\n\nEnter to save, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderAddInput() string {
	var sb strings.Builder
	if m.addMode.asChild {
		sb.WriteString("Add child entry:\n")
	} else {
		sb.WriteString("Add entry:\n")
	}
	sb.WriteString(m.addMode.input.View())
	sb.WriteString("\n\nEnter to add, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderMigrateInput() string {
	var sb strings.Builder
	sb.WriteString("Migrate to date:\n")
	sb.WriteString(m.migrateMode.input.View())
	sb.WriteString("\n\nEnter to migrate, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderGotoInput() string {
	var sb strings.Builder
	sb.WriteString("Go to date:\n")
	sb.WriteString(m.gotoMode.input.View())
	sb.WriteString("\n\nEnter to go, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderSearchInput() string {
	direction := "forward"
	if !m.searchMode.forward {
		direction = "reverse"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search (%s): %s█", direction, m.searchMode.query))
	sb.WriteString("\n\nEnter to find, Ctrl+S/R to find next/prev, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) highlightSearchTerm(line string) string {
	query := m.searchMode.query
	if query == "" {
		return line
	}

	// Case-insensitive search and highlight
	lowerLine := strings.ToLower(line)
	lowerQuery := strings.ToLower(query)

	var result strings.Builder
	remaining := line
	lowerRemaining := lowerLine

	for {
		idx := strings.Index(lowerRemaining, lowerQuery)
		if idx < 0 {
			result.WriteString(remaining)
			break
		}
		result.WriteString(remaining[:idx])
		result.WriteString(SearchHighlightStyle.Render(remaining[idx : idx+len(query)]))
		remaining = remaining[idx+len(query):]
		lowerRemaining = lowerRemaining[idx+len(query):]
	}

	return result.String()
}

func (m Model) renderToolbar() string {
	var viewTypeStr string
	switch m.currentView {
	case ViewTypeJournal:
		viewTypeStr = "Journal"
	case ViewTypeHabits:
		viewTypeStr = "Habits"
	case ViewTypeLists:
		viewTypeStr = "Lists"
	case ViewTypeListItems:
		viewTypeStr = "List Items"
	default:
		viewTypeStr = "Journal"
	}

	viewModeStr := "Day"
	if m.viewMode == ViewModeWeek {
		viewModeStr = "Week"
	}

	dateStr := m.viewDate.Format("Mon, Jan 2 2006")

	return ToolbarStyle.Render(fmt.Sprintf("📓 bujo | %s | %s | %s", viewTypeStr, viewModeStr, dateStr))
}

func (m Model) renderCaptureMode() string {
	var sb strings.Builder

	// Header
	header := "CAPTURE MODE"
	dateStr := m.viewDate.Format("Mon, Jan 2 2006")
	sb.WriteString(ToolbarStyle.Render(fmt.Sprintf("📝 %s | %s", header, dateStr)))
	sb.WriteString("\n")

	maxWidth := m.width
	if maxWidth > 80 {
		maxWidth = 80
	}
	if maxWidth < 20 {
		maxWidth = 20
	}

	sb.WriteString(strings.Repeat("─", maxWidth))
	sb.WriteString("\n")

	// Calculate editor height
	editorHeight := m.height - 8 // Reserve for header, status, help
	if editorHeight < 5 {
		editorHeight = 5
	}

	editorLines := strings.Split(m.captureMode.content, "\n")
	if m.captureMode.content == "" {
		editorLines = []string{""}
	}

	// Calculate scroll offset to keep cursor visible
	scrollOffset := m.captureMode.scrollOffset
	if m.captureMode.cursorLine < scrollOffset {
		scrollOffset = m.captureMode.cursorLine
	}
	if m.captureMode.cursorLine >= scrollOffset+editorHeight {
		scrollOffset = m.captureMode.cursorLine - editorHeight + 1
	}

	// Show scroll indicator if needed
	if scrollOffset > 0 {
		sb.WriteString(HelpStyle.Render(fmt.Sprintf("  ↑ %d more lines above", scrollOffset)))
		sb.WriteString("\n")
		editorHeight--
	}

	// Show editor lines with cursor and search highlighting
	searchQuery := m.captureMode.searchQuery
	linesShown := 0
	for i := scrollOffset; i < len(editorLines) && linesShown < editorHeight; i++ {
		origLine := editorLines[i]
		line := origLine

		// Insert cursor on current line first (before highlighting)
		cursorCol := -1
		if i == m.captureMode.cursorLine {
			cursorCol = m.captureMode.cursorCol
			if cursorCol > len(origLine) {
				cursorCol = len(origLine)
			}
		}

		// Apply search highlighting to the original line content
		if m.captureMode.searchMode && searchQuery != "" && strings.Contains(origLine, searchQuery) {
			var highlighted strings.Builder
			pos := 0
			remaining := origLine
			for {
				idx := strings.Index(remaining, searchQuery)
				if idx < 0 {
					// No more matches - add remaining content with cursor if needed
					if cursorCol >= 0 && cursorCol >= pos {
						relCol := cursorCol - pos
						if relCol < len(remaining) {
							highlighted.WriteString(remaining[:relCol])
							highlighted.WriteString("█")
							highlighted.WriteString(remaining[relCol+1:])
						} else {
							highlighted.WriteString(remaining)
							highlighted.WriteString("█")
						}
					} else {
						highlighted.WriteString(remaining)
					}
					break
				}

				matchStart := pos + idx
				matchEnd := matchStart + len(searchQuery)

				// Add text before match, possibly with cursor
				if cursorCol >= 0 && cursorCol >= pos && cursorCol < matchStart {
					relCol := cursorCol - pos
					highlighted.WriteString(remaining[:relCol])
					highlighted.WriteString("█")
					highlighted.WriteString(remaining[relCol+1 : idx])
					cursorCol = -1
				} else {
					highlighted.WriteString(remaining[:idx])
				}

				// Add highlighted match, possibly with cursor inside
				if cursorCol >= 0 && cursorCol >= matchStart && cursorCol < matchEnd {
					relCol := cursorCol - matchStart
					matchText := searchQuery[:relCol] + "█" + searchQuery[relCol+1:]
					highlighted.WriteString(SearchHighlightStyle.Render(matchText))
					cursorCol = -1
				} else {
					highlighted.WriteString(SearchHighlightStyle.Render(searchQuery))
				}

				pos = matchEnd
				remaining = remaining[idx+len(searchQuery):]
			}
			line = highlighted.String()
		} else if cursorCol >= 0 {
			// No search highlighting, just add cursor
			if cursorCol < len(origLine) {
				line = origLine[:cursorCol] + "█" + origLine[cursorCol+1:]
			} else {
				line = origLine + "█"
			}
		}

		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteString("\n")
		linesShown++
	}

	// Pad remaining lines
	for linesShown < editorHeight {
		sb.WriteString("\n")
		linesShown++
	}

	// Show scroll indicator if more below
	if scrollOffset+editorHeight < len(editorLines) {
		sb.WriteString(HelpStyle.Render(fmt.Sprintf("  ↓ %d more lines below", len(editorLines)-scrollOffset-editorHeight)))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Status bar with error or entry count
	if m.captureMode.draftExists {
		sb.WriteString(ErrorStyle.Render("Restore previous draft? (y/n)"))
	} else if m.captureMode.searchMode {
		direction := "forward"
		if !m.captureMode.searchForward {
			direction = "reverse"
		}
		sb.WriteString(HelpStyle.Render(fmt.Sprintf("Search (%s): %s", direction, m.captureMode.searchQuery)))
	} else if m.captureMode.confirmCancel {
		sb.WriteString(ErrorStyle.Render("Discard changes? (y/n)"))
	} else if m.captureMode.parseError != nil {
		sb.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.captureMode.parseError)))
	} else {
		count := len(m.captureMode.parsedEntries)
		switch count {
		case 0:
			sb.WriteString(HelpStyle.Render("No entries"))
		case 1:
			sb.WriteString(HelpStyle.Render("1 entry"))
		default:
			sb.WriteString(HelpStyle.Render(fmt.Sprintf("%d entries", count)))
		}
	}
	sb.WriteString("\n\n")

	// Help
	if m.captureMode.searchMode {
		sb.WriteString(HelpStyle.Render("Enter/Ctrl+S: next | Ctrl+R: prev | ESC: exit search"))
	} else {
		sb.WriteString(HelpStyle.Render("Ctrl+X: save | ESC: cancel | Tab: indent | Ctrl+S: search"))
	}

	return sb.String()
}

func (m Model) renderErrorPopup() string {
	headerText := "Error"
	message := fmt.Sprintf("%v", m.err)
	footer := "Press any key to dismiss"

	// Find the longest line
	maxLen := len(footer)
	if len(message) > maxLen {
		maxLen = len(message)
	}

	// Pad header to match longest line
	headerPad := maxLen - len(headerText)
	if headerPad < 0 {
		headerPad = 0
	}

	var sb strings.Builder
	sb.WriteString(ErrorTitleStyle.Render(headerText))
	sb.WriteString(strings.Repeat(" ", headerPad))
	sb.WriteString("\n\n")
	sb.WriteString(message)
	sb.WriteString("\n\n")
	sb.WriteString(footer)

	return ErrorStyle.Render(sb.String())
}

func (m Model) renderCommandPalette() string {
	var sb strings.Builder

	sb.WriteString("Command Palette\n")
	sb.WriteString("────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("> %s█\n\n", m.commandPalette.query))

	for i, cmd := range m.commandPalette.filtered {
		prefix := "  "
		if i == m.commandPalette.selectedIdx {
			prefix = "> "
		}

		line := fmt.Sprintf("%s%-25s  [%s]  %s", prefix, cmd.Name, cmd.Keybinding, cmd.Description)

		if i == m.commandPalette.selectedIdx {
			line = SelectedStyle.Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n↑/↓ navigate • Enter select • Esc cancel")
	return ConfirmStyle.Render(sb.String())
}
