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

const separator = "---------------------------------------------------------"

func RenderDailyAgenda(agenda *service.DailyAgenda) string {
	var sb strings.Builder

	// Header line with date and location
	dateStr := agenda.Date.Format("Monday, Jan 2, 2006")
	if agenda.Location != nil {
		sb.WriteString(fmt.Sprintf("📅 %s | 📍 %s\n", cyan(bold(dateStr)), yellow(*agenda.Location)))
	} else {
		sb.WriteString(fmt.Sprintf("📅 %s\n", cyan(bold(dateStr))))
	}
	sb.WriteString(dimmed(separator) + "\n")

	// Overdue section
	if len(agenda.Overdue) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", red(bold("OVERDUE"))))
		for _, entry := range agenda.Overdue {
			sb.WriteString(renderEntry(entry, 0, true))
		}
		sb.WriteString("\n")
	}

	// Today section
	if len(agenda.Today) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", bold("TODAY")))
		renderEntryTree(&sb, agenda.Today, 0)
	} else if len(agenda.Overdue) == 0 {
		sb.WriteString(dimmed("No entries for today\n"))
	}

	sb.WriteString(dimmed(separator) + "\n")

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
	treePrefix := ""
	if depth > 0 {
		treePrefix = "└── "
	}

	// Checkbox for tasks
	checkbox := getCheckbox(entry.Type)
	symbol := getEntrySymbol(entry.Type)
	content := entry.Content
	idStr := fmt.Sprintf("%3d", entry.ID)

	// Color based on type
	switch entry.Type {
	case domain.EntryTypeDone:
		content = green(content)
		symbol = green(symbol)
		checkbox = green(checkbox)
		idStr = green(idStr)
	case domain.EntryTypeMigrated:
		content = dimmed(content)
		symbol = dimmed(symbol)
		checkbox = dimmed(checkbox)
		idStr = dimmed(idStr)
	}

	if overdue {
		content = red(content)
		checkbox = red(checkbox)
		idStr = red(idStr)
	}

	return fmt.Sprintf("%s%s%s %s %s %s\n", indent, treePrefix, checkbox, idStr, symbol, content)
}

func getCheckbox(t domain.EntryType) string {
	switch t {
	case domain.EntryTypeTask:
		return "[ ]"
	case domain.EntryTypeDone:
		return "[x]"
	case domain.EntryTypeMigrated:
		return "[>]"
	default:
		return "   " // Notes and events don't have checkboxes
	}
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
		sparkline := renderSparkline(habit.DayHistory)
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

	// Reverse to show oldest first (only show last 7)
	start := len(days) - 1
	if start > 6 {
		start = 6
	}
	for i := start; i >= 0; i-- {
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
	for i := start; i >= 0; i-- {
		day := days[i]
		label := day.Date.Format("Mon")[:1]
		sb.WriteString(dimmed(label))
		sb.WriteString(" ")
	}

	return sb.String()
}

func RenderHabitMonth(status *service.TrackerStatus) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔥 %s\n\n", cyan(bold("Habit Tracker - Month View"))))

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

		// Month calendar
		sb.WriteString(renderMonthCalendar(habit.DayHistory))

		// Completion percentage
		completionColor := green
		if habit.CompletionPercent < 50 {
			completionColor = red
		} else if habit.CompletionPercent < 80 {
			completionColor = yellow
		}
		sb.WriteString(fmt.Sprintf("  %s completion (last 30 days)\n\n", completionColor(fmt.Sprintf("%.0f%%", habit.CompletionPercent))))
	}

	return sb.String()
}

func renderMonthCalendar(days []service.DayStatus) string {
	var sb strings.Builder

	// Header with week days
	sb.WriteString("  ")
	for _, d := range []string{"M", "T", "W", "T", "F", "S", "S"} {
		sb.WriteString(dimmed(d) + " ")
	}
	sb.WriteString("\n")

	// Build a map of date -> completed
	completed := make(map[string]bool)
	for _, day := range days {
		key := day.Date.Format("2006-01-02")
		completed[key] = day.Completed
	}

	// Find the start of the calendar (go back to find a Monday)
	if len(days) == 0 {
		return sb.String()
	}

	oldest := days[len(days)-1].Date
	newest := days[0].Date

	// Start from oldest, find the Monday of that week
	startDate := oldest
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	// Render weeks
	sb.WriteString("  ")
	current := startDate
	for !current.After(newest) {
		key := current.Format("2006-01-02")

		if current.Before(oldest) || current.After(newest) {
			sb.WriteString(dimmed("·") + " ")
		} else if completed[key] {
			sb.WriteString(green("●") + " ")
		} else {
			sb.WriteString(dimmed("○") + " ")
		}

		// New line on Sunday
		if current.Weekday() == time.Sunday {
			sb.WriteString("\n  ")
		}

		current = current.AddDate(0, 0, 1)
	}
	sb.WriteString("\n")

	return sb.String()
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
