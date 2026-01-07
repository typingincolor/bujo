package tui

import (
	"fmt"
	"strings"

	"github.com/typingincolor/bujo/internal/domain"
)

func (m Model) View() string {
	if m.err != nil {
		return m.renderErrorPopup()
	}

	if m.agenda == nil {
		return "Loading..."
	}

	var sb strings.Builder

	// Toolbar
	toolbar := m.renderToolbar()
	sb.WriteString(toolbar)
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(m.width, 60)))
	sb.WriteString("\n")

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
	}

	sb.WriteString("\n")
	sb.WriteString(HelpStyle.Render(m.help.View(m.keyMap)))

	return sb.String()
}

func (m Model) renderEntry(item EntryItem) string {
	entry := item.Entry
	indent := strings.Repeat("  ", item.Indent)
	treePrefix := ""
	if item.Indent > 0 {
		treePrefix = "└── "
	}

	symbol := entry.Type.Symbol()
	content := entry.Content

	base := fmt.Sprintf("%s%s%s %s", indent, treePrefix, symbol, content)

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

func (m Model) renderToolbar() string {
	viewModeStr := "Day"
	if m.viewMode == ViewModeWeek {
		viewModeStr = "Week"
	}

	dateStr := m.viewDate.Format("Mon, Jan 2 2006")

	return ToolbarStyle.Render(fmt.Sprintf("📓 bujo | %s | %s", viewModeStr, dateStr))
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
