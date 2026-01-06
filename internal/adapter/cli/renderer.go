package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

var (
	cyan    = color.New(color.FgCyan).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	bold    = color.New(color.Bold).SprintFunc()
	dimmed  = color.New(color.Faint).SprintFunc()
)

func RenderDailyAgenda(agenda *service.DailyAgenda) string {
	var sb strings.Builder

	// Header
	dateStr := agenda.Date.Format("Monday, January 2, 2006")
	sb.WriteString(fmt.Sprintf("📅 %s\n", cyan(bold(dateStr))))

	// Location
	if agenda.Location != nil {
		sb.WriteString(fmt.Sprintf("📍 %s\n", yellow(*agenda.Location)))
	}

	sb.WriteString("\n")

	// Overdue section
	if len(agenda.Overdue) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", red(bold("⚠️  Overdue"))))
		for _, entry := range agenda.Overdue {
			sb.WriteString(renderEntry(entry, 0, true))
		}
		sb.WriteString("\n")
	}

	// Today section
	if len(agenda.Today) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", bold("Today")))
		renderEntryTree(&sb, agenda.Today, 0)
	} else if len(agenda.Overdue) == 0 {
		sb.WriteString(dimmed("No entries for today\n"))
	}

	return sb.String()
}

func renderEntryTree(sb *strings.Builder, entries []domain.Entry, depth int) {
	// Build parent-child map
	children := make(map[int64][]domain.Entry)
	var roots []domain.Entry

	for _, e := range entries {
		if e.ParentID == nil {
			roots = append(roots, e)
		} else {
			children[*e.ParentID] = append(children[*e.ParentID], e)
		}
	}

	for _, root := range roots {
		renderEntryWithChildren(sb, root, children, depth)
	}
}

func renderEntryWithChildren(sb *strings.Builder, entry domain.Entry, children map[int64][]domain.Entry, depth int) {
	sb.WriteString(renderEntry(entry, depth, false))

	for _, child := range children[entry.ID] {
		renderEntryWithChildren(sb, child, children, depth+1)
	}
}

func renderEntry(entry domain.Entry, depth int, overdue bool) string {
	indent := strings.Repeat("  ", depth)
	prefix := "├── "
	if depth == 0 {
		prefix = ""
	}

	symbol := getEntrySymbol(entry.Type)
	content := entry.Content

	// Color based on type
	switch entry.Type {
	case domain.EntryTypeDone:
		content = green(content)
		symbol = green(symbol)
	case domain.EntryTypeMigrated:
		content = dimmed(content)
		symbol = dimmed(symbol)
	}

	if overdue {
		content = red(content)
	}

	return fmt.Sprintf("%s%s%s %s\n", indent, prefix, symbol, content)
}

func getEntrySymbol(t domain.EntryType) string {
	switch t {
	case domain.EntryTypeTask:
		return "."
	case domain.EntryTypeNote:
		return "-"
	case domain.EntryTypeEvent:
		return "o"
	case domain.EntryTypeDone:
		return "x"
	case domain.EntryTypeMigrated:
		return ">"
	default:
		return "."
	}
}

func RenderHabitTracker(status *service.TrackerStatus) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔥 %s\n\n", cyan(bold("Habit Tracker"))))

	if len(status.Habits) == 0 {
		sb.WriteString(dimmed("No habits tracked yet\n"))
		return sb.String()
	}

	for _, habit := range status.Habits {
		// Habit name and streak
		streakColor := green
		if habit.CurrentStreak == 0 {
			streakColor = red
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", bold(habit.Name), streakColor(fmt.Sprintf("(%d day streak)", habit.CurrentStreak))))

		// Sparkline for last 7 days
		sparkline := renderSparkline(habit.Last7Days)
		sb.WriteString(fmt.Sprintf("  %s\n", sparkline))

		// Completion percentage
		completionColor := green
		if habit.CompletionPercent < 50 {
			completionColor = red
		} else if habit.CompletionPercent < 80 {
			completionColor = yellow
		}
		sb.WriteString(fmt.Sprintf("  %s completion\n\n", completionColor(fmt.Sprintf("%.0f%%", habit.CompletionPercent))))
	}

	return sb.String()
}

func renderSparkline(days []service.DayStatus) string {
	var sb strings.Builder

	// Reverse to show oldest first
	for i := len(days) - 1; i >= 0; i-- {
		day := days[i]
		if day.Completed {
			sb.WriteString(green("●"))
		} else {
			sb.WriteString(dimmed("○"))
		}
		sb.WriteString(" ")
	}

	// Add day labels
	sb.WriteString("\n  ")
	for i := len(days) - 1; i >= 0; i-- {
		day := days[i]
		label := day.Date.Format("Mon")[:1]
		sb.WriteString(dimmed(label))
		sb.WriteString(" ")
	}

	return sb.String()
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
