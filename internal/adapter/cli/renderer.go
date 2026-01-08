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
	Cyan      = color.New(color.FgCyan).SprintFunc()
	Green     = color.New(color.FgGreen).SprintFunc()
	Red       = color.New(color.FgRed).SprintFunc()
	Yellow    = color.New(color.FgYellow).SprintFunc()
	Bold      = color.New(color.Bold).SprintFunc()
	Dimmed    = color.New(color.Faint).SprintFunc()
	Highlight = color.New(color.FgYellow, color.Bold).SprintFunc()
)

const separator = "---------------------------------------------------------"

func RenderDailyAgenda(agenda *service.DailyAgenda) string {
	var sb strings.Builder

	// Header line with date and location
	dateStr := agenda.Date.Format("Monday, Jan 2, 2006")
	if agenda.Location != nil {
		sb.WriteString(fmt.Sprintf("📅 %s | 📍 %s\n", Cyan(Bold(dateStr)), Yellow(*agenda.Location)))
	} else {
		sb.WriteString(fmt.Sprintf("📅 %s\n", Cyan(Bold(dateStr))))
	}
	sb.WriteString(Dimmed(separator) + "\n")

	// Overdue section
	if len(agenda.Overdue) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", Red(Bold("OVERDUE"))))
		renderEntryTreeWithOverdue(&sb, agenda.Overdue, 0, true)
		sb.WriteString("\n")
	}

	// Today section
	if len(agenda.Today) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", Bold("TODAY")))
		renderEntryTreeWithOverdue(&sb, agenda.Today, 0, false)
	} else if len(agenda.Overdue) == 0 {
		sb.WriteString(Dimmed("No entries for today\n"))
	}

	sb.WriteString(Dimmed(separator) + "\n")

	return sb.String()
}

func RenderMultiDayAgenda(agenda *service.MultiDayAgenda) string {
	var sb strings.Builder

	// Overdue section
	if len(agenda.Overdue) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", Red(Bold("OVERDUE"))))
		renderEntryTreeWithOverdue(&sb, agenda.Overdue, 0, true)
		sb.WriteString("\n")
	}

	// Each day
	for _, day := range agenda.Days {
		dateStr := day.Date.Format("Monday, Jan 2")
		if day.Location != nil {
			sb.WriteString(fmt.Sprintf("📅 %s | 📍 %s\n", Cyan(Bold(dateStr)), Yellow(*day.Location)))
		} else {
			sb.WriteString(fmt.Sprintf("📅 %s\n", Cyan(Bold(dateStr))))
		}

		if len(day.Entries) > 0 {
			renderEntryTreeWithOverdue(&sb, day.Entries, 0, false)
		} else {
			sb.WriteString(Dimmed("  No entries\n"))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderEntryTree(sb *strings.Builder, entries []domain.Entry, depth int) {
	renderEntryTreeWithOverdue(sb, entries, depth, false)
}

func renderEntryTreeWithOverdue(sb *strings.Builder, entries []domain.Entry, depth int, overdue bool) {
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
		renderEntryWithChildren(sb, root, children, depth, overdue)
	}
}

func renderEntryWithChildren(sb *strings.Builder, entry domain.Entry, children map[int64][]domain.Entry, depth int, overdue bool) {
	sb.WriteString(renderEntry(entry, depth, overdue))

	for _, child := range children[entry.ID] {
		renderEntryWithChildren(sb, child, children, depth+1, overdue)
	}
}

func renderEntry(entry domain.Entry, depth int, overdue bool) string {
	indent := strings.Repeat("  ", depth)
	treePrefix := ""
	if depth > 0 {
		treePrefix = "└── "
	}

	symbol := entry.Type.Symbol()
	content := entry.Content
	idStr := fmt.Sprintf("(%d)", entry.ID)

	// Color based on type
	switch entry.Type {
	case domain.EntryTypeDone:
		content = Green(content)
		symbol = Green(symbol)
		idStr = Green(idStr)
	case domain.EntryTypeMigrated:
		content = Dimmed(content)
		symbol = Dimmed(symbol)
		idStr = Dimmed(idStr)
	}

	if overdue {
		content = Red(content)
		idStr = Red(idStr)
	}

	return fmt.Sprintf("%s%s%s %s %s\n", indent, treePrefix, symbol, content, idStr)
}

func RenderHabitTracker(status *service.TrackerStatus) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔥 %s\n\n", Cyan(Bold("Habit Tracker"))))

	if len(status.Habits) == 0 {
		sb.WriteString(Dimmed("No habits tracked yet\n"))
		return sb.String()
	}

	for _, habit := range status.Habits {
		// Habit name and streak
		streakColor := Green
		if habit.CurrentStreak == 0 {
			streakColor = Red
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", Bold(habit.Name), streakColor(fmt.Sprintf("(%d day streak)", habit.CurrentStreak))))

		// Sparkline for last 7 days
		sparkline := renderSparkline(habit.DayHistory)
		sb.WriteString(fmt.Sprintf("  %s\n", sparkline))

		sb.WriteString(renderHabitProgress(habit, ""))
	}

	return sb.String()
}

func renderHabitProgress(habit service.HabitStatus, suffix string) string {
	todayColor := Green
	if habit.TodayCount < habit.GoalPerDay {
		todayColor = Yellow
	}
	completionColor := Green
	if habit.CompletionPercent < 50 {
		completionColor = Red
	} else if habit.CompletionPercent < 80 {
		completionColor = Yellow
	}
	return fmt.Sprintf("  %s today | %s completion%s\n\n",
		todayColor(fmt.Sprintf("%d/%d", habit.TodayCount, habit.GoalPerDay)),
		completionColor(fmt.Sprintf("%.0f%%", habit.CompletionPercent)),
		suffix)
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
			sb.WriteString(Green("●"))
		} else {
			sb.WriteString(Dimmed("○"))
		}
		sb.WriteString(" ")
	}

	// Add day labels
	sb.WriteString("\n  ")
	for i := start; i >= 0; i-- {
		day := days[i]
		label := day.Date.Format("Mon")[:1]
		sb.WriteString(Dimmed(label))
		sb.WriteString(" ")
	}

	return sb.String()
}

func RenderHabitMonth(status *service.TrackerStatus) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔥 %s\n\n", Cyan(Bold("Habit Tracker - Month View"))))

	if len(status.Habits) == 0 {
		sb.WriteString(Dimmed("No habits tracked yet\n"))
		return sb.String()
	}

	for _, habit := range status.Habits {
		// Habit name and streak
		streakColor := Green
		if habit.CurrentStreak == 0 {
			streakColor = Red
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", Bold(habit.Name), streakColor(fmt.Sprintf("(%d day streak)", habit.CurrentStreak))))

		// Month calendar
		sb.WriteString(renderMonthCalendar(habit.DayHistory))

		sb.WriteString(renderHabitProgress(habit, " (last 30 days)"))
	}

	return sb.String()
}

func renderMonthCalendar(days []service.DayStatus) string {
	var sb strings.Builder

	// Header with week days
	sb.WriteString("  ")
	for _, d := range []string{"M", "T", "W", "T", "F", "S", "S"} {
		sb.WriteString(Dimmed(d) + " ")
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
			sb.WriteString(Dimmed("·") + " ")
		} else if completed[key] {
			sb.WriteString(Green("●") + " ")
		} else {
			sb.WriteString(Dimmed("○") + " ")
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

func RenderHabitInspect(details *service.HabitDetails) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("📋 %s\n", Cyan(Bold(details.Name))))

	// Stats line
	streakColor := Green
	if details.CurrentStreak == 0 {
		streakColor = Red
	}
	sb.WriteString(fmt.Sprintf("Streak: %s | Goal: %d/day\n\n",
		streakColor(fmt.Sprintf("%d days", details.CurrentStreak)),
		details.GoalPerDay))

	// Logs table
	if len(details.Logs) == 0 {
		sb.WriteString(Dimmed("No logs in this period\n"))
	} else {
		sb.WriteString(Bold("Logs:\n"))
		sb.WriteString(Dimmed("  ID      Date         Count\n"))
		for _, log := range details.Logs {
			sb.WriteString(fmt.Sprintf("  %-6d  %-11s  %d\n",
				log.ID,
				log.LoggedAt.Format("Jan 2, 2006"),
				log.Count))
		}
	}

	sb.WriteString(fmt.Sprintf("\n%s %d\n", Dimmed("Habit ID:"), details.ID))

	return sb.String()
}
