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
	Cyan          = color.New(color.FgCyan).SprintFunc()
	Green         = color.New(color.FgGreen).SprintFunc()
	Red           = color.New(color.FgRed).SprintFunc()
	Yellow        = color.New(color.FgYellow).SprintFunc()
	Bold          = color.New(color.Bold).SprintFunc()
	Dimmed        = color.New(color.Faint).SprintFunc()
	Highlight     = color.New(color.FgYellow, color.Bold).SprintFunc()
	Strikethrough = color.New(color.CrossedOut).SprintFunc()
)

const separator = "---------------------------------------------------------"

func RenderDaysWithOverdue(days []service.DayEntries, overdue []domain.Entry, today time.Time) string {
	var sb strings.Builder

	if len(overdue) > 0 {
		sb.WriteString(fmt.Sprintf("⚠️  %s\n", Red(Bold("OVERDUE"))))
		renderEntryTreeWithOverdue(&sb, overdue, 0, true, today)
		sb.WriteString("\n")
	}

	for _, day := range days {
		dateStr := day.Date.Format("Monday, Jan 2")
		header := fmt.Sprintf("📅 %s", Cyan(Bold(dateStr)))

		if day.Location != nil {
			header += fmt.Sprintf(" | 📍 %s", Yellow(*day.Location))
		}
		if day.Weather != nil {
			header += fmt.Sprintf(" | ☀️  %s", Cyan(*day.Weather))
		}
		if day.Mood != nil {
			header += fmt.Sprintf(" | 😊 %s", Yellow(*day.Mood))
		}

		sb.WriteString(header + "\n")

		if len(day.Entries) > 0 {
			renderEntryTreeWithOverdue(&sb, day.Entries, 0, false, today)
		} else {
			sb.WriteString(Dimmed("  No entries\n"))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderEntryTreeWithOverdue(sb *strings.Builder, entries []domain.Entry, depth int, forceOverdue bool, today time.Time) {
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
		renderEntryWithChildren(sb, root, children, depth, forceOverdue, today)
	}
}

func renderEntryWithChildren(sb *strings.Builder, entry domain.Entry, children map[int64][]domain.Entry, depth int, forceOverdue bool, today time.Time) {
	sb.WriteString(renderEntry(entry, depth, forceOverdue, today))

	for _, child := range children[entry.ID] {
		renderEntryWithChildren(sb, child, children, depth+1, forceOverdue, today)
	}
}

func renderEntry(entry domain.Entry, depth int, forceOverdue bool, today time.Time) string {
	indent := strings.Repeat("  ", depth)

	symbol := entry.Type.Symbol()
	prioritySymbol := entry.Priority.Symbol()
	content := entry.Content
	idStr := fmt.Sprintf("(%d)", entry.ID)

	switch entry.Type {
	case domain.EntryTypeDone, domain.EntryTypeAnswered:
		content = Green(content)
		symbol = Green(symbol)
		idStr = Green(idStr)
	case domain.EntryTypeAnswer:
		symbol = Dimmed(symbol)
	case domain.EntryTypeMigrated:
		content = Dimmed(content)
		symbol = Dimmed(symbol)
		idStr = Dimmed(idStr)
	case domain.EntryTypeCancelled:
		content = Strikethrough(Dimmed(content))
		symbol = Dimmed(symbol)
		idStr = Dimmed(idStr)
	}

	isOverdue := forceOverdue || (!today.IsZero() && entry.IsOverdue(today))
	if isOverdue {
		content = Red(content)
		idStr = Red(idStr)
	}

	if prioritySymbol != "" {
		return fmt.Sprintf("%s%s %s %s %s\n", indent, symbol, prioritySymbol, content, idStr)
	}
	return fmt.Sprintf("%s%s %s %s\n", indent, symbol, content, idStr)
}

func RenderHabitTracker(status *service.TrackerStatus) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔥 %s\n\n", Cyan(Bold("Habit Tracker"))))

	if len(status.Habits) == 0 {
		sb.WriteString(Dimmed("No habits tracked yet\n"))
		return sb.String()
	}

	for _, habit := range status.Habits {
		streakColor := Green
		if habit.CurrentStreak == 0 {
			streakColor = Red
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", Bold(habit.Name), streakColor(fmt.Sprintf("(%d day streak)", habit.CurrentStreak))))

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
		streakColor := Green
		if habit.CurrentStreak == 0 {
			streakColor = Red
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", Bold(habit.Name), streakColor(fmt.Sprintf("(%d day streak)", habit.CurrentStreak))))

		sb.WriteString(renderMonthCalendar(habit.DayHistory))

		sb.WriteString(renderHabitProgress(habit, " (last 30 days)"))
	}

	return sb.String()
}

func renderMonthCalendar(days []service.DayStatus) string {
	var sb strings.Builder

	sb.WriteString("  ")
	for _, d := range []string{"M", "T", "W", "T", "F", "S", "S"} {
		sb.WriteString(Dimmed(d) + " ")
	}
	sb.WriteString("\n")

	completed := make(map[string]bool)
	for _, day := range days {
		key := day.Date.Format("2006-01-02")
		completed[key] = day.Completed
	}

	if len(days) == 0 {
		return sb.String()
	}

	oldest := days[len(days)-1].Date
	newest := days[0].Date

	startDate := oldest
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, -1)
	}

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

	sb.WriteString(fmt.Sprintf("%s\n", Cyan(Bold(details.Name))))

	streakColor := Green
	if details.CurrentStreak == 0 {
		streakColor = Red
	}
	sb.WriteString(fmt.Sprintf("Streak: %s\n", streakColor(fmt.Sprintf("%d days", details.CurrentStreak))))

	sb.WriteString("\nGoals:\n")
	if details.GoalPerDay > 0 {
		sb.WriteString(fmt.Sprintf("  Daily:   %d/day\n", details.GoalPerDay))
	}
	if details.GoalPerWeek > 0 {
		sb.WriteString(fmt.Sprintf("  Weekly:  %d/week  %s\n",
			details.GoalPerWeek,
			formatProgress(details.WeeklyProgress)))
	}
	if details.GoalPerMonth > 0 {
		sb.WriteString(fmt.Sprintf("  Monthly: %d/month %s\n",
			details.GoalPerMonth,
			formatProgress(details.MonthlyProgress)))
	}
	if details.GoalPerDay == 0 && details.GoalPerWeek == 0 && details.GoalPerMonth == 0 {
		sb.WriteString(Dimmed("  No goals set\n"))
	}

	sb.WriteString("\n")
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

func formatProgress(progress float64) string {
	text := fmt.Sprintf("%.0f%%", progress)
	if progress >= 100 {
		return Green(text)
	} else if progress >= 50 {
		return Yellow(text)
	}
	return Red(text)
}

func RenderGoalsSection(goals []domain.Goal, month time.Time) string {
	if len(goals) == 0 {
		return ""
	}

	var sb strings.Builder
	monthName := month.Format("January")

	sb.WriteString(fmt.Sprintf("🎯 %s\n", Cyan(Bold(monthName+" Goals"))))

	doneCount := 0
	for _, goal := range goals {
		var status string
		var content string
		if goal.IsDone() {
			status = Green("✓")
			content = Green(goal.Content)
			doneCount++
		} else {
			status = Dimmed("○")
			content = goal.Content
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", status, content))
	}

	progress := float64(doneCount) / float64(len(goals)) * 100
	sb.WriteString(fmt.Sprintf("  %s\n", Dimmed(fmt.Sprintf("Progress: %s", formatProgress(progress)))))
	sb.WriteString("\n")

	return sb.String()
}
