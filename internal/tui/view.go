package tui

import (
	"fmt"
	"strings"

	"github.com/typingincolor/bujo/internal/domain"
)

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.agenda == nil {
		return "Loading..."
	}

	var sb strings.Builder

	if len(m.entries) == 0 {
		sb.WriteString(HelpStyle.Render("No entries for the last 7 days."))
		sb.WriteString("\n\n")
	} else {
		for i, item := range m.entries {
			if item.DayHeader != "" {
				if i > 0 {
					sb.WriteString("\n")
				}
				if item.IsOverdue {
					sb.WriteString(OverdueHeaderStyle.Render(item.DayHeader))
				} else {
					sb.WriteString(DateHeaderStyle.Render(item.DayHeader))
				}
				sb.WriteString("\n")
			}

			line := m.renderEntry(item)

			if i == m.selectedIdx {
				line = SelectedStyle.Render(line)
			}

			sb.WriteString(line)
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
	idStr := fmt.Sprintf("(%d)", entry.ID)

	base := fmt.Sprintf("%s%s%s %s %s", indent, treePrefix, symbol, content, idStr)

	switch entry.Type {
	case domain.EntryTypeDone:
		return DoneStyle.Render(base)
	case domain.EntryTypeMigrated:
		return MigratedStyle.Render(base)
	default:
		if item.IsOverdue {
			return OverdueStyle.Render(base)
		}
		styled := fmt.Sprintf("%s%s%s %s %s", indent, treePrefix, symbol, content, IDStyle.Render(idStr))
		return styled
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
