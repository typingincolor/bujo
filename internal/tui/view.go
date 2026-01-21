package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func (m Model) View() string {
	if m.quitConfirmMode.active {
		return m.renderQuitConfirm()
	}

	if m.err != nil {
		return m.renderErrorPopup()
	}

	var sb strings.Builder

	toolbar := m.renderToolbar()
	sb.WriteString(toolbar)
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(m.width, 60)))
	sb.WriteString("\n")

	switch m.currentView {
	case ViewTypeHabits:
		sb.WriteString(m.renderHabitsContent())
	case ViewTypeLists, ViewTypeListItems:
		sb.WriteString(m.renderListsContent())
	case ViewTypeSearch:
		sb.WriteString(m.renderSearchContent())
	case ViewTypeStats:
		sb.WriteString(m.renderStatsContent())
	case ViewTypeGoals:
		sb.WriteString(m.renderGoalsContent())
	case ViewTypeSettings:
		sb.WriteString(m.renderSettingsContent())
	case ViewTypePendingTasks:
		sb.WriteString(m.renderPendingTasksContent())
	case ViewTypeQuestions:
		sb.WriteString(m.renderQuestionsContent())
	default:
		sb.WriteString(m.renderJournalContent())
	}

	if m.editMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderEditInput())
		sb.WriteString("\n")
	} else if m.answerMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderAnswerInput())
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
	} else if m.setLocationMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderSetLocationInput())
		sb.WriteString("\n")
	} else if m.searchMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderSearchInput())
		sb.WriteString("\n")
	} else if m.commandPalette.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderCommandPalette())
		sb.WriteString("\n")
	} else if m.addGoalMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderAddGoalInput())
		sb.WriteString("\n")
	} else if m.editGoalMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderEditGoalInput())
		sb.WriteString("\n")
	} else if m.confirmGoalDeleteMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderConfirmGoalDeleteDialog())
		sb.WriteString("\n")
	} else if m.moveGoalMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderMoveGoalInput())
		sb.WriteString("\n")
	} else if m.migrateToGoalMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderMigrateToGoalInput())
		sb.WriteString("\n")
	} else if m.moveListItemMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderMoveListItemModal())
		sb.WriteString("\n")
	} else if m.createListMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderCreateListInput())
		sb.WriteString("\n")
	} else if m.moveToListMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderMoveToListModal())
		sb.WriteString("\n")
	} else if m.addHabitMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderAddHabitInput())
		sb.WriteString("\n")
	} else if m.confirmHabitDeleteMode.active {
		sb.WriteString("\n")
		sb.WriteString(m.renderConfirmHabitDeleteDialog())
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(HelpStyle.Render(m.renderContextHelp()))

	return sb.String()
}

func (m Model) renderContextHelp() string {
	switch m.currentView {
	case ViewTypeHabits:
		return "j/k: navigate  ←/→: day  [/]: period  space: log  ⌫: remove  a: add  d: delete  w: view  esc: back  q: quit"
	case ViewTypeLists:
		return "j/k: navigate  enter: open  a: add list  esc: back  q: quit"
	case ViewTypeListItems:
		return "j/k: navigate  space: toggle  a: add  e: edit  d: delete  M: move  esc: back  q: quit"
	case ViewTypeGoals:
		return "j/k: navigate  h/l: month  space: toggle  a: add  e: edit  d: delete  >: move  esc: back  q: quit"
	case ViewTypeSearch:
		return "j/k: navigate  /: search  enter: go to  esc: back  q: quit"
	case ViewTypeStats:
		return "esc: back  q: quit"
	case ViewTypePendingTasks:
		return "j/k: navigate  esc: back  q: quit"
	case ViewTypeQuestions:
		return "j/k: navigate  esc: back  q: quit"
	default:
		return m.help.View(m.keyMap)
	}
}

func (m Model) renderJournalContent() string {
	if m.agenda == nil {
		return "Loading..."
	}

	var sb strings.Builder

	if m.isViewingPast() {
		sb.WriteString(m.renderJournalAISummary())
		sb.WriteString("\n")
	}

	if len(m.entries) == 0 {
		emptyMsg := "No entries for today."
		if m.viewMode == ViewModeWeek {
			emptyMsg = "No entries for the last 7 days."
		}
		sb.WriteString(HelpStyle.Render(emptyMsg))
		sb.WriteString("\n\n")
	} else {
		availableLines := m.height - 6 // 2 for toolbar, 2 for help, 2 for padding
		if availableLines < 5 {
			availableLines = 5
		}

		if m.scrollOffset > 0 {
			sb.WriteString(HelpStyle.Render(fmt.Sprintf("  ↑ %d more above", m.scrollOffset)))
			sb.WriteString("\n")
			availableLines--
		}

		reserveForBelow := 1

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

			spaceNeeded := linesNeeded
			if i < len(m.entries)-1 {
				spaceNeeded += reserveForBelow
			}
			if linesUsed+spaceNeeded > availableLines {
				break
			}

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

			line := m.renderEntry(item, i == m.selectedIdx)
			if m.searchMode.active && m.searchMode.query != "" {
				line = m.highlightSearchTerm(line)
			}
			sb.WriteString(line)
			sb.WriteString("\n")
			linesUsed++
			endIdx = i + 1
		}

		if endIdx < len(m.entries) {
			sb.WriteString(HelpStyle.Render(fmt.Sprintf("  ↓ %d more below", len(m.entries)-endIdx)))
			sb.WriteString("\n")
		}
	}

	if len(m.journalGoals) > 0 {
		sb.WriteString("\n")
		now := time.Now()
		monthName := now.Format("January")
		sb.WriteString(fmt.Sprintf("🎯 %s Goals\n", monthName))

		doneCount := 0
		for _, goal := range m.journalGoals {
			var status string
			var content string
			if goal.IsDone() {
				status = DoneStyle.Render("✓")
				content = DoneStyle.Render(goal.Content)
				doneCount++
			} else {
				status = HelpStyle.Render("○")
				content = goal.Content
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", status, content))
		}

		progress := float64(doneCount) / float64(len(m.journalGoals)) * 100
		sb.WriteString(HelpStyle.Render(fmt.Sprintf("  Progress: %.0f%%", progress)))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderHabitsContent() string {
	var sb strings.Builder

	if len(m.habitState.habits) == 0 {
		sb.WriteString(HelpStyle.Render("No habits yet. Press 'a' to add a habit."))
		sb.WriteString("\n\n")
		return sb.String()
	}

	days := HabitDaysWeek
	switch m.habitState.viewMode {
	case HabitViewModeMonth:
		days = HabitDaysMonth
	case HabitViewModeQuarter:
		days = HabitDaysQuarter
	}

	for i, habit := range m.habitState.habits {
		streakText := "day"
		if habit.CurrentStreak != 1 {
			streakText = "days"
		}
		nameLine := fmt.Sprintf("%s (%d %s streak)", habit.Name, habit.CurrentStreak, streakText)
		if i == m.habitState.selectedIdx {
			nameLine = SelectedStyle.Render(nameLine)
		}
		sb.WriteString(nameLine)
		sb.WriteString("\n")

		sparkline := m.renderSparkline(habit.DayHistory, i == m.habitState.selectedIdx)
		sb.WriteString("  " + sparkline)
		sb.WriteString("\n")

		dayLabels := m.renderDayLabels(days)
		sb.WriteString("  " + HelpStyle.Render(dayLabels))
		sb.WriteString("\n")

		todayInfo := fmt.Sprintf("  %d/%d today | %.0f%% completion", habit.TodayCount, habit.GoalPerDay, habit.CompletionPercent)
		sb.WriteString(HelpStyle.Render(todayInfo))
		sb.WriteString("\n")

		if habit.GoalPerWeek > 0 || habit.GoalPerMonth > 0 {
			var progressParts []string
			if habit.GoalPerWeek > 0 {
				progressParts = append(progressParts, fmt.Sprintf("Week: %.0f%%", habit.WeeklyProgress))
			}
			if habit.GoalPerMonth > 0 {
				progressParts = append(progressParts, fmt.Sprintf("Month: %.0f%%", habit.MonthlyProgress))
			}
			progressLine := HelpStyle.Render("  " + strings.Join(progressParts, "  "))
			sb.WriteString(progressLine)
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderSparkline(history []service.DayStatus, isSelected bool) string {
	var parts []string
	days := len(history)

	for i := days - 1; i >= 0; i-- {
		day := history[i]
		displayPos := days - 1 - i
		selected := isSelected && displayPos == m.habitState.selectedDayIdx

		char := "○"
		if day.Completed {
			char = "●"
		}

		if selected {
			char = HabitSelectedStyle.Render(char)
		}
		parts = append(parts, char)
	}
	return strings.Join(parts, " ")
}

func (m Model) renderDayLabels(days int) string {
	referenceDate := m.getHabitReferenceDate()

	switch m.habitState.viewMode {
	case HabitViewModeMonth:
		return m.renderMonthLabels(referenceDate, days)
	case HabitViewModeQuarter:
		return m.renderQuarterLabels(referenceDate, days)
	default:
		return m.renderWeekLabels(referenceDate, days)
	}
}

func (m Model) renderWeekLabels(referenceDate time.Time, days int) string {
	dayNames := []string{"S", "M", "T", "W", "T", "F", "S"}
	var labels []string

	for i := days - 1; i >= 0; i-- {
		date := referenceDate.AddDate(0, 0, -i)
		dayOfWeek := int(date.Weekday())
		labels = append(labels, dayNames[dayOfWeek])
	}
	return strings.Join(labels, " ")
}

func (m Model) renderMonthLabels(referenceDate time.Time, days int) string {
	var labels []string
	var lastShownDate int

	for i := days - 1; i >= 0; i-- {
		date := referenceDate.AddDate(0, 0, -i)
		dayOfMonth := date.Day()
		displayPos := days - 1 - i

		if displayPos == 0 || displayPos == days-1 || (dayOfMonth == 1 && displayPos > 0) {
			labels = append(labels, fmt.Sprintf("%d", dayOfMonth))
			lastShownDate = displayPos
		} else if displayPos-lastShownDate >= 7 && dayOfMonth%7 == 0 {
			labels = append(labels, fmt.Sprintf("%d", dayOfMonth))
			lastShownDate = displayPos
		} else {
			labels = append(labels, "·")
		}
	}
	return strings.Join(labels, " ")
}

func (m Model) renderQuarterLabels(referenceDate time.Time, days int) string {
	var labels []string
	var lastMonth time.Month

	for i := days - 1; i >= 0; i-- {
		date := referenceDate.AddDate(0, 0, -i)
		displayPos := days - 1 - i

		if displayPos == 0 || date.Month() != lastMonth {
			monthAbbrev := date.Format("Jan")
			labels = append(labels, monthAbbrev)
			lastMonth = date.Month()
		} else {
			labels = append(labels, " · ")
		}
	}
	return strings.Join(labels, "")
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

func (m Model) renderEntry(item EntryItem, selected bool) string {
	entry := item.Entry
	indent := strings.Repeat("  ", item.Indent)

	collapseIndicator := ""
	if item.HasChildren {
		if item.HiddenChildCount > 0 {
			collapseIndicator = "▶ "
		} else {
			collapseIndicator = "▼ "
		}
	}

	symbol := entry.Type.Symbol()
	prioritySymbol := entry.Priority.Symbol()
	content := entry.Content

	hiddenSuffix := ""
	if item.HiddenChildCount > 0 {
		hiddenSuffix = fmt.Sprintf(" [%d hidden]", item.HiddenChildCount)
	}

	var base string
	if prioritySymbol != "" {
		base = fmt.Sprintf("%s%s%s %s %s%s", indent, collapseIndicator, symbol, prioritySymbol, content, hiddenSuffix)
	} else {
		base = fmt.Sprintf("%s%s%s %s%s", indent, collapseIndicator, symbol, content, hiddenSuffix)
	}

	if selected {
		return SelectedStyle.Render(base)
	}

	switch entry.Type {
	case domain.EntryTypeDone, domain.EntryTypeAnswered:
		return DoneStyle.Render(base)
	case domain.EntryTypeMigrated:
		return MigratedStyle.Render(base)
	case domain.EntryTypeCancelled:
		return CancelledStyle.Render(base)
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

func (m Model) renderAddHabitInput() string {
	var sb strings.Builder
	sb.WriteString("Add habit:\n")
	sb.WriteString(m.addHabitMode.input.View())
	sb.WriteString("\n\nEnter to add, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderConfirmHabitDeleteDialog() string {
	habitName := ""
	for _, h := range m.habitState.habits {
		if h.ID == m.confirmHabitDeleteMode.habitID {
			habitName = h.Name
			break
		}
	}
	dialog := fmt.Sprintf(`Delete habit "%s"?

  y - Yes, delete
  n - No, cancel`, habitName)

	return ConfirmStyle.Render(dialog)
}

func (m Model) renderEditInput() string {
	var sb strings.Builder
	sb.WriteString("Edit entry:\n")
	sb.WriteString(m.editMode.input.View())
	sb.WriteString("\n\nEnter to save, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderAnswerInput() string {
	var sb strings.Builder
	sb.WriteString("Answer question:\n")
	sb.WriteString(m.answerMode.input.View())
	sb.WriteString("\n\nEnter to submit, Esc to cancel")
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

func (m Model) renderSetLocationInput() string {
	var sb strings.Builder
	sb.WriteString("Set location:\n")
	sb.WriteString(m.setLocationMode.input.View())

	if m.setLocationMode.pickerMode && len(m.setLocationMode.locations) > 0 {
		sb.WriteString("\n\nPrevious locations:\n")
		for i, loc := range m.setLocationMode.locations {
			if i == m.setLocationMode.selectedIdx {
				sb.WriteString(SelectedStyle.Render(fmt.Sprintf("  > %s", loc)))
			} else {
				sb.WriteString(fmt.Sprintf("    %s", loc))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n↑/↓ to select, Enter to pick, Esc to cancel")
	} else {
		sb.WriteString("\n\nEnter to save, Esc to cancel")
	}
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

	if m.selectedIdx >= 0 && m.selectedIdx < len(m.entries) {
		selectedEntry := m.entries[m.selectedIdx].Entry
		ancestors := m.getAncestryChain(selectedEntry.ID)
		if len(ancestors) > 0 {
			const maxAncestors = 3
			const maxContentLen = 40

			start := 0
			prefix := ""
			if len(ancestors) > maxAncestors {
				start = len(ancestors) - maxAncestors
				prefix = "... > "
			}

			var ancestorNames []string
			for i := start; i < len(ancestors); i++ {
				content := ancestors[i].Content
				if len(content) > maxContentLen {
					content = content[:maxContentLen] + "..."
				}
				ancestorNames = append(ancestorNames, content)
			}
			sb.WriteString("\n\n")
			sb.WriteString(HelpStyle.Render("↳ " + prefix + strings.Join(ancestorNames, " > ")))
		}
	}

	return ConfirmStyle.Render(sb.String())
}

func (m Model) highlightSearchTerm(line string) string {
	query := m.searchMode.query
	if query == "" {
		return line
	}

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
	case ViewTypeReview:
		viewTypeStr = "Review"
	case ViewTypePendingTasks:
		viewTypeStr = "Outstanding Tasks"
	case ViewTypeQuestions:
		viewTypeStr = "Open Questions"
	case ViewTypeHabits:
		viewTypeStr = "Habits"
	case ViewTypeLists:
		viewTypeStr = "Lists"
	case ViewTypeListItems:
		viewTypeStr = "List Items"
	case ViewTypeGoals:
		viewTypeStr = "Goals"
	case ViewTypeSearch:
		viewTypeStr = "Search"
	case ViewTypeStats:
		viewTypeStr = "Stats"
	case ViewTypeSettings:
		viewTypeStr = "Settings"
	default:
		viewTypeStr = "Journal"
	}

	var dateStr string
	if m.currentView == ViewTypeReview {
		// For Review view, show the week range
		weekStart := m.viewDate.AddDate(0, 0, -int(m.viewDate.Weekday()))
		weekEnd := weekStart.AddDate(0, 0, 6)
		dateStr = fmt.Sprintf("%s - %s", weekStart.Format("Jan 2"), weekEnd.Format("Jan 2, 2006"))
	} else {
		dateStr = m.viewDate.Format("Mon, Jan 2 2006")
	}

	return ToolbarStyle.Render(fmt.Sprintf("📓 bujo | %s | %s", viewTypeStr, dateStr))
}

func (m Model) renderQuitConfirm() string {
	var sb strings.Builder

	sb.WriteString("\n\n")
	sb.WriteString(TitleStyle.Render("Quit Confirmation"))
	sb.WriteString("\n\n")
	sb.WriteString("Are you sure you want to quit?")
	sb.WriteString("\n\n")
	sb.WriteString(HelpStyle.Render("y = yes, n = no, esc = cancel"))
	sb.WriteString("\n")

	return sb.String()
}

func (m Model) renderErrorPopup() string {
	headerText := "Error"
	message := fmt.Sprintf("%v", m.err)
	footer := "Press any key to dismiss"

	maxLen := len(footer)
	if len(message) > maxLen {
		maxLen = len(message)
	}

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

func (m Model) renderSearchContent() string {
	var sb strings.Builder

	sb.WriteString("🔍 Search\n\n")

	sb.WriteString("  ")
	sb.WriteString(m.searchView.input.View())
	sb.WriteString("\n\n")

	if m.searchView.loading {
		sb.WriteString("  Searching...\n")
	} else if m.searchView.query == "" {
		sb.WriteString(HelpStyle.Render("  Type to search entries"))
		sb.WriteString("\n")
	} else if len(m.searchView.results) == 0 {
		sb.WriteString(fmt.Sprintf("  No results found for %q\n", m.searchView.query))
	} else {
		sb.WriteString(fmt.Sprintf("  Found %d result(s)\n\n", len(m.searchView.results)))
		for i, entry := range m.searchView.results {
			line := m.renderSearchResultLine(entry, i == m.searchView.selectedIdx)
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(HelpStyle.Render("j/k: navigate • enter: view • esc: clear • /: focus search"))
	sb.WriteString("\n")

	return sb.String()
}

func (m Model) renderSearchResultLine(entry domain.Entry, selected bool) string {
	var parts []string

	dateStr := "no date"
	if entry.ScheduledDate != nil {
		dateStr = entry.ScheduledDate.Format("2006-01-02")
	}

	symbol := entry.Type.Symbol()
	content := entry.Content
	idStr := fmt.Sprintf("(%d)", entry.ID)

	switch entry.Type {
	case domain.EntryTypeDone, domain.EntryTypeAnswered:
		symbol = DoneStyle.Render(symbol)
		content = DoneStyle.Render(content)
		dateStr = DoneStyle.Render(dateStr)
		idStr = DoneStyle.Render(idStr)
	case domain.EntryTypeMigrated, domain.EntryTypeCancelled:
		symbol = MigratedStyle.Render(symbol)
		content = MigratedStyle.Render(content)
		dateStr = MigratedStyle.Render(dateStr)
		idStr = MigratedStyle.Render(idStr)
	default:
		dateStr = IDStyle.Render(dateStr)
		idStr = IDStyle.Render(idStr)
	}

	prefix := "  "
	if selected {
		prefix = TitleStyle.Render("> ")
		content = SelectedStyle.Render(entry.Content)
	}

	parts = append(parts, prefix)
	parts = append(parts, fmt.Sprintf("[%s]", dateStr))
	parts = append(parts, symbol)
	parts = append(parts, content)
	parts = append(parts, idStr)

	return strings.Join(parts, " ")
}

func (m Model) renderStatsContent() string {
	var sb strings.Builder

	if m.statsViewState.stats != nil {
		sb.WriteString(fmt.Sprintf("📊 Statistics (%s to %s)\n",
			m.statsViewState.from.Format("Jan 2"),
			m.statsViewState.to.Format("Jan 2, 2006")))
	} else {
		sb.WriteString("📊 Statistics\n")
	}
	sb.WriteString(strings.Repeat("─", 50))
	sb.WriteString("\n\n")

	if m.statsViewState.loading {
		sb.WriteString("Loading statistics...\n")
		return sb.String()
	}

	stats := m.statsViewState.stats
	if stats == nil {
		sb.WriteString(HelpStyle.Render("No statistics available"))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Entries: %d total\n", stats.EntryCounts.Total))
		if stats.EntryCounts.Tasks > 0 {
			pct := float64(stats.EntryCounts.Tasks) / float64(stats.EntryCounts.Total) * 100
			sb.WriteString(fmt.Sprintf("  • Tasks:     %d (%.0f%%)\n", stats.EntryCounts.Tasks, pct))
		}
		if stats.EntryCounts.Notes > 0 {
			pct := float64(stats.EntryCounts.Notes) / float64(stats.EntryCounts.Total) * 100
			sb.WriteString(fmt.Sprintf("  – Notes:     %d (%.0f%%)\n", stats.EntryCounts.Notes, pct))
		}
		if stats.EntryCounts.Events > 0 {
			pct := float64(stats.EntryCounts.Events) / float64(stats.EntryCounts.Total) * 100
			sb.WriteString(fmt.Sprintf("  ○ Events:    %d (%.0f%%)\n", stats.EntryCounts.Events, pct))
		}
		if stats.EntryCounts.Done > 0 {
			pct := float64(stats.EntryCounts.Done) / float64(stats.EntryCounts.Total) * 100
			sb.WriteString(fmt.Sprintf("  ✓ Completed: %d (%.0f%%)\n", stats.EntryCounts.Done, pct))
		}
		sb.WriteString("\n")

		if stats.TaskCompletion.Total > 0 {
			sb.WriteString(fmt.Sprintf("Task completion: %.0f%% (%d/%d)\n",
				stats.TaskCompletion.Rate,
				stats.TaskCompletion.Completed,
				stats.TaskCompletion.Total))
		}

		if stats.Productivity.AveragePerDay > 0 {
			sb.WriteString(fmt.Sprintf("Average entries/day: %.1f\n", stats.Productivity.AveragePerDay))
		}
		if stats.Productivity.MostProductive.Average > 0 {
			sb.WriteString(fmt.Sprintf("\nMost productive: %ss (avg %.1f)\n",
				stats.Productivity.MostProductive.Day.String(),
				stats.Productivity.MostProductive.Average))
		}
		if stats.Productivity.LeastProductive.Average > 0 {
			sb.WriteString(fmt.Sprintf("Least productive: %ss (avg %.1f)\n",
				stats.Productivity.LeastProductive.Day.String(),
				stats.Productivity.LeastProductive.Average))
		}

		if stats.HabitStats.Active > 0 {
			sb.WriteString(fmt.Sprintf("\nHabits: %d active\n", stats.HabitStats.Active))
			if stats.HabitStats.BestStreak.Days > 0 {
				sb.WriteString(fmt.Sprintf("  Best streak: %s (%d days)\n",
					stats.HabitStats.BestStreak.HabitName,
					stats.HabitStats.BestStreak.Days))
			}
			if stats.HabitStats.MostLogged.Count > 0 {
				sb.WriteString(fmt.Sprintf("  Most logged: %s (%d logs)\n",
					stats.HabitStats.MostLogged.HabitName,
					stats.HabitStats.MostLogged.Count))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderGoalsContent() string {
	var sb strings.Builder

	monthName := m.goalState.viewMonth.Format("January 2006")
	sb.WriteString(fmt.Sprintf("🎯 Monthly Goals - %s\n\n", monthName))

	if len(m.goalState.goals) == 0 {
		sb.WriteString(HelpStyle.Render("No goals for this month. Press 'a' to add one."))
		sb.WriteString("\n\n")
	} else {
		for i, goal := range m.goalState.goals {
			status := "  "
			if goal.IsDone() {
				status = "✓ "
			}

			line := fmt.Sprintf("%s#%-3d %s", status, goal.ID, goal.Content)

			if goal.IsDone() {
				line = DoneStyle.Render(line)
			}

			if i == m.goalState.selectedIdx {
				line = SelectedStyle.Render(line)
			}

			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderSettingsContent() string {
	var sb strings.Builder

	sb.WriteString("⚙️  Settings\n\n")

	sb.WriteString("Current configuration:\n\n")
	sb.WriteString(fmt.Sprintf("  Theme:         %s\n", "default"))
	sb.WriteString(fmt.Sprintf("  Default view:  %s\n", "journal"))
	sb.WriteString(fmt.Sprintf("  Date format:   %s\n", "Mon, Jan 2 2006"))
	sb.WriteString("\n")

	sb.WriteString(HelpStyle.Render("Edit ~/.config/bujo/config.yaml to change settings"))
	sb.WriteString("\n\n")

	return sb.String()
}

func (m Model) renderPendingTasksContent() string {
	var sb strings.Builder

	sb.WriteString("📋 Outstanding Tasks\n\n")

	if m.pendingTasksState.loading {
		sb.WriteString("Loading...")
		return sb.String()
	}

	if len(m.pendingTasksState.entries) == 0 {
		sb.WriteString(HelpStyle.Render("No outstanding tasks. All caught up!"))
		sb.WriteString("\n\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d outstanding task(s)\n\n", len(m.pendingTasksState.entries)))

	maxLines := m.pendingTasksVisibleRows()
	startIdx := m.pendingTasksState.scrollOffset
	linesUsed := 0
	endIdx := startIdx

	if startIdx > 0 {
		sb.WriteString(HelpStyle.Render("  ↑ more above"))
		sb.WriteString("\n")
		linesUsed++
	}

	var currentDateStr string
	for i := startIdx; i < len(m.pendingTasksState.entries); i++ {
		entry := m.pendingTasksState.entries[i]
		isSelected := i == m.pendingTasksState.selectedIdx
		isExpanded := entry.ID == m.pendingTasksState.expandedID

		entryDateStr := ""
		if entry.ScheduledDate != nil {
			entryDateStr = entry.ScheduledDate.Format("2006-01-02")
		}

		linesNeeded := 1
		if entryDateStr != currentDateStr {
			linesNeeded += 2
		}
		if isExpanded {
			if chain, ok := m.pendingTasksState.parentChains[entry.ID]; ok && len(chain) > 0 {
				linesNeeded += len(chain)
			}
		}

		if linesUsed+linesNeeded > maxLines && i > startIdx {
			break
		}

		if entryDateStr != currentDateStr {
			if currentDateStr != "" {
				sb.WriteString("\n")
				linesUsed++
			}
			if entry.ScheduledDate != nil {
				sb.WriteString(DateHeaderStyle.Render(entry.ScheduledDate.Format("Mon, Jan 2")))
			} else {
				sb.WriteString(DateHeaderStyle.Render("No Date"))
			}
			sb.WriteString("\n")
			linesUsed++
			currentDateStr = entryDateStr
		}

		if isExpanded {
			if chain, ok := m.pendingTasksState.parentChains[entry.ID]; ok && len(chain) > 0 {
				sb.WriteString(m.renderParentChain(chain))
				linesUsed += len(chain)
			}
		}

		line := m.renderPendingEntryLine(entry, isSelected, isExpanded, m.pendingTasksState.parentChains)
		sb.WriteString(line)
		sb.WriteString("\n")
		linesUsed++
		endIdx = i + 1
	}

	if endIdx < len(m.pendingTasksState.entries) {
		sb.WriteString(HelpStyle.Render("  ↓ more below"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (m Model) renderEntryLine(entry domain.Entry, selected bool) string {
	dateStr := "no date"
	if entry.ScheduledDate != nil {
		dateStr = entry.ScheduledDate.Format("2006-01-02")
	}

	symbol := entry.Type.Symbol()
	content := entry.Content

	prefix := "  "
	if selected {
		prefix = "> "
	}

	line := fmt.Sprintf("%s[%s] %s %s", prefix, dateStr, symbol, content)

	if selected {
		return SelectedStyle.Render(line)
	}

	return line
}

func (m Model) renderPendingEntryLine(entry domain.Entry, selected bool, expanded bool, parentChains map[int64][]domain.Entry) string {
	symbol := entry.Type.Symbol()
	content := entry.Content
	hasParents := entry.ParentID != nil

	prefix := "  "
	if expanded && hasParents {
		prefix = "    "
	} else if selected {
		prefix = "> "
	}

	contextIndicator := ""
	if !expanded && hasParents {
		contextIndicator = "↳ "
	}

	line := fmt.Sprintf("%s%s%s %s", prefix, contextIndicator, symbol, content)

	if selected {
		return SelectedStyle.Render(line)
	}

	return line
}

func (m Model) renderParentChain(chain []domain.Entry) string {
	var sb strings.Builder

	for i := len(chain) - 1; i >= 0; i-- {
		ancestor := chain[i]
		indent := strings.Repeat("  ", len(chain)-1-i)
		sb.WriteString(fmt.Sprintf("  %s> %s %s\n", indent, ancestor.Type.Symbol(), HelpStyle.Render(ancestor.Content)))
	}

	return sb.String()
}

func (m Model) renderQuestionsContent() string {
	var sb strings.Builder

	sb.WriteString("❓ Open Questions\n\n")

	if m.questionsState.loading {
		sb.WriteString("Loading...")
		return sb.String()
	}

	if len(m.questionsState.entries) == 0 {
		sb.WriteString(HelpStyle.Render("No open questions."))
		sb.WriteString("\n\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d open question(s)\n\n", len(m.questionsState.entries)))

	for i, entry := range m.questionsState.entries {
		line := m.renderEntryLine(entry, i == m.questionsState.selectedIdx)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (m Model) renderAddGoalInput() string {
	var sb strings.Builder
	sb.WriteString("Add goal:\n")
	sb.WriteString(m.addGoalMode.input.View())
	sb.WriteString("\n\nEnter to add, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderEditGoalInput() string {
	var sb strings.Builder
	sb.WriteString("Edit goal:\n")
	sb.WriteString(m.editGoalMode.input.View())
	sb.WriteString("\n\nEnter to save, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderConfirmGoalDeleteDialog() string {
	dialog := `Delete this goal?

  y - Yes, delete
  n - No, cancel`

	return ConfirmStyle.Render(dialog)
}

func (m Model) renderMoveGoalInput() string {
	var sb strings.Builder
	sb.WriteString("Move goal to month (YYYY-MM):\n")
	sb.WriteString(m.moveGoalMode.input.View())
	sb.WriteString("\n\nEnter to move, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderMigrateToGoalInput() string {
	var sb strings.Builder
	sb.WriteString("Convert task to goal:\n")
	sb.WriteString(fmt.Sprintf("Task: %s\n\n", m.migrateToGoalMode.content))
	sb.WriteString("Target month (YYYY-MM):\n")
	sb.WriteString(m.migrateToGoalMode.input.View())
	sb.WriteString("\n\nEnter to convert, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderMoveListItemModal() string {
	var sb strings.Builder
	sb.WriteString("Move item to list:\n\n")

	for i, list := range m.moveListItemMode.targetLists {
		prefix := "  "
		if i == m.moveListItemMode.selectedIdx {
			prefix = "> "
		}
		num := i + 1
		if num <= 9 {
			sb.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, num, list.Name))
		} else {
			sb.WriteString(fmt.Sprintf("%s   %s\n", prefix, list.Name))
		}
	}

	sb.WriteString("\n1-9 or Enter to move, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderCreateListInput() string {
	var sb strings.Builder
	sb.WriteString("Create new list:\n")
	sb.WriteString(m.createListMode.input.View())
	sb.WriteString("\n\nEnter to create, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) renderMoveToListModal() string {
	var sb strings.Builder
	sb.WriteString("Move entry to list:\n\n")

	for i, list := range m.moveToListMode.targetLists {
		prefix := "  "
		if i == m.moveToListMode.selectedIdx {
			prefix = "> "
		}
		num := i + 1
		if num <= 9 {
			sb.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, num, list.Name))
		} else {
			sb.WriteString(fmt.Sprintf("%s   %s\n", prefix, list.Name))
		}
	}

	sb.WriteString("\n1-9 or Enter to move, Esc to cancel")
	return ConfirmStyle.Render(sb.String())
}

func (m Model) isViewingPast() bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	viewDate := time.Date(m.viewDate.Year(), m.viewDate.Month(), m.viewDate.Day(), 0, 0, 0, 0, m.viewDate.Location())
	return viewDate.Before(today)
}

func (m Model) renderJournalAISummary() string {
	var sb strings.Builder

	horizonLabel := "Daily"
	if m.summaryState.horizon == "weekly" {
		horizonLabel = "Weekly"
	}

	sb.WriteString(strings.Repeat("─", 50))
	sb.WriteString("\n\n")

	collapseIndicator := "▼"
	if m.summaryCollapsed {
		collapseIndicator = "▶"
	}

	sb.WriteString(fmt.Sprintf("%s 🤖 AI %s Summary", collapseIndicator, horizonLabel))

	if m.summaryCollapsed {
		sb.WriteString(HelpStyle.Render("  (press 's' to expand)"))
		sb.WriteString("\n\n")
		return sb.String()
	}

	sb.WriteString("\n\n")

	if m.summaryState.streaming && m.summaryState.accumulatedText != "" {
		sb.WriteString("⏳ Generating...\n\n")
		rendered, err := m.renderMarkdown(m.summaryState.accumulatedText)
		if err != nil {
			sb.WriteString(m.summaryState.accumulatedText)
		} else {
			sb.WriteString(rendered)
		}
		sb.WriteString("\n")
	} else if m.summaryState.summary != nil {
		rendered, err := m.renderMarkdown(m.summaryState.summary.Content)
		if err != nil {
			sb.WriteString(m.summaryState.summary.Content)
		} else {
			sb.WriteString(rendered)
		}
		sb.WriteString("\n")
	} else if m.summaryService == nil {
		sb.WriteString(HelpStyle.Render("AI summaries unavailable - configure BUJO_MODEL or GEMINI_API_KEY"))
		sb.WriteString("\n")
	} else if m.summaryState.loading {
		sb.WriteString("⏳ Generating AI summary...\n")
	} else if m.summaryState.error != nil {
		if errors.Is(m.summaryState.error, domain.ErrNoEntries) {
			sb.WriteString(HelpStyle.Render("No entries to summarize for this period"))
			sb.WriteString("\n")
		} else {
			sb.WriteString(fmt.Sprintf("❌ Error: %v\n", m.summaryState.error))
		}
	} else {
		sb.WriteString(HelpStyle.Render("No summary generated for this period"))
		sb.WriteString("\n")
	}

	return sb.String()
}
