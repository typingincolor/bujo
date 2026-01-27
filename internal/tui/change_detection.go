package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const changePollingInterval = 2 * time.Second

func (m Model) checkChangesCmd() tea.Cmd {
	return tea.Tick(changePollingInterval, func(time.Time) tea.Msg {
		return checkChangesMsg{}
	})
}

func (m Model) handleCheckChanges() (tea.Model, tea.Cmd) {
	if m.changeDetection == nil {
		return m, m.checkChangesCmd()
	}

	ctx := context.Background()
	currentMod, err := m.changeDetection.GetLastModified(ctx)
	if err != nil {
		return m, m.checkChangesCmd()
	}

	// First call: initialize the timestamp without triggering a reload
	if m.lastCheckedModified.IsZero() {
		m.lastCheckedModified = currentMod
		return m, m.checkChangesCmd()
	}

	if currentMod.After(m.lastCheckedModified) {
		m.lastCheckedModified = currentMod
		return m, tea.Batch(m.reloadCurrentViewCmd(), m.checkChangesCmd())
	}

	return m, m.checkChangesCmd()
}

func (m Model) handleDataChanged() (tea.Model, tea.Cmd) {
	return m, m.reloadCurrentViewCmd()
}

func (m Model) reloadCurrentViewCmd() tea.Cmd {
	switch m.currentView {
	case ViewTypeJournal, ViewTypeReview:
		return m.loadDaysCmd()
	case ViewTypeHabits:
		return m.loadHabitsCmd()
	case ViewTypeLists:
		return m.loadListsCmd()
	case ViewTypeListItems:
		return m.loadListItemsCmd(m.listState.currentListID)
	case ViewTypeGoals:
		return m.loadGoalsCmd()
	case ViewTypePendingTasks:
		return m.loadPendingTasksCmd()
	case ViewTypeQuestions:
		return m.loadQuestionsCmd()
	case ViewTypeStats:
		return m.loadStatsCmd()
	default:
		return nil
	}
}
