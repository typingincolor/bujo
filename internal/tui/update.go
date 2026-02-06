package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/dateutil"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case daysLoadedMsg:
		// Preserve the currently selected entry EntityID before reloading.
		var selectedEntityID domain.EntityID
		if m.selectedIdx >= 0 && m.selectedIdx < len(m.entries) {
			selectedEntityID = m.entries[m.selectedIdx].Entry.EntityID
		}

		m.days = msg.days
		m.entries = m.flattenDays(msg.days)

		// Try to restore selection to the same entry
		if selectedEntityID != "" {
			found := false
			for idx, item := range m.entries {
				if item.Entry.EntityID == selectedEntityID {
					m.selectedIdx = idx
					found = true
					break
				}
			}
			if !found && m.selectedIdx >= len(m.entries) {
				if len(m.entries) > 0 {
					m.selectedIdx = len(m.entries) - 1
				} else {
					m.selectedIdx = 0
				}
			}
		} else if m.selectedIdx >= len(m.entries) {
			m.selectedIdx = 0
		}
		m.scrollOffset = 0
		m = m.syncKeyMapToSelection()

		return m.ensuredVisible(), m.loadJournalGoalsCmd()

	case journalGoalsLoadedMsg:
		m.journalGoals = msg.goals
		return m, nil

	case errMsg:
		m.err = msg.err
		m.undoState = undoState{}
		return m, nil

	case entryUpdatedMsg, entryDeletedMsg, entryMovedToListMsg, agendaReloadNeededMsg:
		return m, m.loadDaysCmd()

	case gotoDateMsg:
		m.viewDate = msg.date
		m.selectedIdx = 0
		return m, m.loadDaysCmd()

	case confirmDeleteMsg:
		m.confirmMode = confirmState{
			active:      true,
			entryID:     msg.entryID,
			hasChildren: msg.hasChildren,
		}
		return m, nil

	case habitsLoadedMsg:
		m.habitState.habits = msg.habits
		if m.habitState.selectedIdx >= len(m.habitState.habits) {
			m.habitState.selectedIdx = 0
		}
		if !m.habitState.dayIdxInited {
			days := HabitDaysWeek
			switch m.habitState.viewMode {
			case HabitViewModeMonth:
				days = HabitDaysMonth
			case HabitViewModeQuarter:
				days = HabitDaysQuarter
			}
			m.habitState.selectedDayIdx = days - 1
			m.habitState.dayIdxInited = true
		}
		return m, nil

	case habitLoggedMsg:
		return m, m.loadHabitsCmd()

	case habitLogRemovedMsg:
		return m, m.loadHabitsCmd()

	case habitAddedMsg:
		return m, m.loadHabitsCmd()

	case habitDeletedMsg:
		return m, m.loadHabitsCmd()

	case listsLoadedMsg:
		m.listState.lists = msg.lists
		m.listState.summaries = msg.summaries
		if m.listState.selectedListIdx >= len(m.listState.lists) {
			m.listState.selectedListIdx = 0
		}
		return m, nil

	case listCreatedMsg:
		return m, m.loadListsCmd()

	case listsForMoveLoadedMsg:
		m.moveToListMode = moveToListState{
			active:      true,
			entryID:     msg.entryID,
			targetLists: msg.lists,
			selectedIdx: 0,
		}
		return m, nil

	case listItemsLoadedMsg:
		m.listState.items = msg.items
		if m.listState.selectedItemIdx >= len(m.listState.items) {
			m.listState.selectedItemIdx = 0
		}
		return m, nil

	case listItemToggledMsg:
		return m, m.loadListItemsCmd(m.listState.currentListID)

	case listItemAddedMsg:
		return m, m.loadListItemsCmd(msg.listID)

	case listItemDeletedMsg:
		return m, m.loadListItemsCmd(msg.listID)

	case listItemEditedMsg:
		return m, m.loadListItemsCmd(msg.listID)

	case listItemMovedMsg:
		return m, m.loadListItemsCmd(msg.fromListID)

	case goalsLoadedMsg:
		m.goalState.goals = msg.goals
		if m.goalState.selectedIdx >= len(m.goalState.goals) {
			m.goalState.selectedIdx = 0
		}
		return m, nil

	case goalToggledMsg:
		return m, m.loadGoalsCmd()

	case goalAddedMsg:
		return m, m.loadGoalsCmd()

	case goalEditedMsg:
		return m, m.loadGoalsCmd()

	case goalDeletedMsg:
		m.goalState.selectedIdx = 0
		return m, m.loadGoalsCmd()

	case goalMovedMsg:
		return m, m.loadGoalsCmd()

	case entryMigratedToGoalMsg:
		return m, m.loadDaysCmd()

	case statsLoadedMsg:
		m.statsViewState.loading = false
		m.statsViewState.stats = msg.stats
		return m, nil

	case insightsDashboardLoadedMsg:
		m.insightsState.loading = false
		m.insightsState.dashboard = msg.dashboard
		return m, nil

	case insightsSummaryLoadedMsg:
		m.insightsState.loading = false
		m.insightsState.weekSummary = msg.summary
		m.insightsState.weekTopics = msg.topics
		return m, nil

	case insightsActionsLoadedMsg:
		m.insightsState.loading = false
		m.insightsState.weekActions = msg.actions
		return m, nil

	case searchResultsMsg:
		m.searchView.loading = false
		m.searchView.results = msg.results
		m.searchView.query = msg.query
		m.searchView.selectedIdx = 0
		return m, nil

	case pendingTasksLoadedMsg:
		m.pendingTasksState.loading = false
		m.pendingTasksState.entries = msg.entries
		m.pendingTasksState.selectedIdx = 0
		m.pendingTasksState.parentChains = make(map[int64][]domain.Entry)
		m.pendingTasksState.expandedID = 0
		return m, nil

	case parentChainLoadedMsg:
		if m.pendingTasksState.parentChains == nil {
			m.pendingTasksState.parentChains = make(map[int64][]domain.Entry)
		}
		m.pendingTasksState.parentChains[msg.entryID] = msg.chain
		m.pendingTasksState.expandedID = msg.entryID
		return m, nil

	case questionsLoadedMsg:
		m.questionsState.loading = false
		m.questionsState.entries = msg.entries
		m.questionsState.selectedIdx = 0
		return m, nil

	case locationsLoadedMsg:
		m.setLocationMode.locations = msg.locations
		return m, nil

	case moodsLoadedMsg:
		m.setMoodMode.presets = mergePresets(moodPresets, msg.moods)
		return m, nil

	case weathersLoadedMsg:
		m.setWeatherMode.presets = mergePresets(weatherPresets, msg.weathers)
		return m, nil

	case editorFinishedMsg:
		_ = DeleteDraft(m.draftPath)
		_ = CleanupCaptureTempFile(CaptureTempFilePath())
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		content := strings.TrimSpace(msg.content)
		if content == "" {
			return m, nil
		}
		return m, m.saveCaptureCmd(content)

	case checkChangesMsg:
		return m.handleCheckChanges()

	case dataChangedMsg:
		return m.handleDataChanged()

	case tea.KeyMsg:
		if m.quitConfirmMode.active {
			return m.handleQuitConfirmMode(msg)
		}
		if m.err != nil {
			m.err = nil
			return m, nil
		}
		if m.editMode.active {
			return m.handleEditMode(msg)
		}
		if m.answerMode.active {
			return m.handleAnswerMode(msg)
		}
		if m.addMode.active {
			return m.handleAddMode(msg)
		}
		if m.migrateMode.active {
			return m.handleMigrateMode(msg)
		}
		if m.confirmMode.active {
			return m.handleConfirmMode(msg)
		}
		if m.gotoMode.active {
			return m.handleGotoMode(msg)
		}
		if m.retypeMode.active {
			return m.handleRetypeMode(msg)
		}
		if m.searchMode.active {
			return m.handleSearchMode(msg)
		}
		if m.setLocationMode.active {
			return m.handleSetLocationMode(msg)
		}
		if m.setMoodMode.active {
			return m.handleSetMoodMode(msg)
		}
		if m.setWeatherMode.active {
			return m.handleSetWeatherMode(msg)
		}
		if m.commandPalette.active {
			return m.handleCommandPaletteMode(msg)
		}
		if m.addHabitMode.active {
			return m.handleAddHabitMode(msg)
		}
		if m.confirmHabitDeleteMode.active {
			return m.handleConfirmHabitDeleteMode(msg)
		}
		if m.addGoalMode.active {
			return m.handleAddGoalMode(msg)
		}
		if m.editGoalMode.active {
			return m.handleEditGoalMode(msg)
		}
		if m.confirmGoalDeleteMode.active {
			return m.handleConfirmGoalDeleteMode(msg)
		}
		if m.moveGoalMode.active {
			return m.handleMoveGoalMode(msg)
		}
		if m.migrateToGoalMode.active {
			return m.handleMigrateToGoalMode(msg)
		}
		if m.moveListItemMode.active {
			return m.handleMoveListItemMode(msg)
		}
		if m.createListMode.active {
			return m.handleCreateListMode(msg)
		}
		if m.moveToListMode.active {
			return m.handleMoveToListMode(msg)
		}

		if key.Matches(msg, m.keyMap.CommandPalette) {
			m.commandPalette.active = true
			m.commandPalette.query = ""
			m.commandPalette.selectedIdx = 0
			m.commandPalette.filtered = m.commandRegistry.All()
			return m, nil
		}

		switch m.currentView {
		case ViewTypeHabits:
			return m.handleHabitsMode(msg)
		case ViewTypeLists, ViewTypeListItems:
			return m.handleListsMode(msg)
		case ViewTypeGoals:
			return m.handleGoalsMode(msg)
		case ViewTypeStats:
			return m.handleStatsMode(msg)
		case ViewTypeInsights:
			return m.handleInsightsMode(msg)
		case ViewTypeSearch:
			return m.handleSearchViewMode(msg)
		case ViewTypePendingTasks:
			return m.handlePendingTasksMode(msg)
		case ViewTypeQuestions:
			return m.handleQuestionsMode(msg)
		default:
			return m.handleNormalMode(msg)
		}
	}

	return m, nil
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.ExpandAll):
		m = m.expandAllSiblings()
		return m, nil

	case key.Matches(msg, m.keyMap.CollapseAll):
		m = m.collapseAllSiblings()
		return m, nil

	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case key.Matches(msg, m.keyMap.ToggleOverdueContext):
		m = m.toggleOverdueContext()
		m.entries = m.flattenDays(m.days)
		return m.ensuredVisible(), nil

	case key.Matches(msg, m.keyMap.Up):
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m.syncKeyMapToSelection().ensuredVisible(), nil

	case key.Matches(msg, m.keyMap.Down):
		if m.selectedIdx < len(m.entries)-1 {
			m.selectedIdx++
		}
		return m.syncKeyMapToSelection().ensuredVisible(), nil

	case key.Matches(msg, m.keyMap.Top):
		m.selectedIdx = 0
		m.scrollOffset = 0
		return m.syncKeyMapToSelection(), nil

	case key.Matches(msg, m.keyMap.Bottom):
		if len(m.entries) > 0 {
			m.selectedIdx = len(m.entries) - 1
			m = m.scrollToBottom().syncKeyMapToSelection()
		}
		return m, nil

	case msg.Type == tea.KeyEnter:
		if len(m.entries) > 0 {
			item := m.entries[m.selectedIdx]
			if item.HasChildren {
				entityID := item.Entry.EntityID
				_, hasState := m.collapsed[entityID]
				if hasState {
					m.collapsed[entityID] = !m.collapsed[entityID]
				} else {
					m.collapsed[entityID] = false
				}
				m.entries = m.flattenDays(m.days)
				return m.ensuredVisible(), nil
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		if len(m.entries) > 0 {
			entry := m.entries[m.selectedIdx].Entry
			oldEntry := entry
			switch entry.Type {
			case domain.EntryTypeDone:
				m.undoState = undoState{
					operation: UndoOpMarkDone,
					entryID:   entry.ID,
					entityID:  entry.EntityID,
					oldEntry:  &oldEntry,
				}
			case domain.EntryTypeTask:
				m.undoState = undoState{
					operation: UndoOpMarkUndone,
					entryID:   entry.ID,
					entityID:  entry.EntityID,
					oldEntry:  &oldEntry,
				}
			}
		}
		return m, m.toggleDoneCmd()

	case key.Matches(msg, m.keyMap.Answer):
		if len(m.entries) > 0 {
			entry := m.entries[m.selectedIdx].Entry
			if entry.Type == domain.EntryTypeQuestion {
				ti := textinput.New()
				ti.Placeholder = "Enter your answer..."
				ti.Focus()
				ti.CharLimit = 512
				ti.Width = m.width - 10
				m.answerMode = answerState{
					active:     true,
					questionID: entry.ID,
					input:      ti,
				}
				return m, nil
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.CancelEntry):
		if len(m.entries) > 0 {
			entry := m.entries[m.selectedIdx].Entry
			if !entry.CanCancel() {
				return m, nil
			}
			return m, m.cancelEntryCmd()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.UncancelEntry):
		if len(m.entries) > 0 {
			entry := m.entries[m.selectedIdx].Entry
			if !entry.CanUncancel() {
				return m, nil
			}
			return m, m.uncancelEntryCmd()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Retype):
		if len(m.entries) > 0 {
			entry := m.entries[m.selectedIdx].Entry
			if !entry.CanCycleType() {
				return m, nil
			}
			m.retypeMode = retypeState{
				active:      true,
				entryID:     entry.ID,
				selectedIdx: 0,
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Undo):
		if m.undoState.operation != UndoOpNone {
			cmd := m.undoCmd()
			m.undoState = undoState{}
			return m, cmd
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Delete):
		if len(m.entries) > 0 {
			entry := m.entries[m.selectedIdx].Entry
			m.confirmMode.active = true
			m.confirmMode.entryID = entry.ID
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Edit):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		if !entry.CanEdit() {
			return m, nil
		}
		ti := textinput.New()
		ti.SetValue(entry.Content)
		ti.Focus()
		ti.CharLimit = 256
		ti.Width = m.width - 10
		m.editMode = editState{
			active:  true,
			entryID: entry.ID,
			input:   ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.OpenURL):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		return m, m.openURLCmd(entry.Content)

	case key.Matches(msg, m.keyMap.Add):
		ti := textinput.New()
		ti.Placeholder = ". task, - note, o event"
		ti.Focus()
		ti.CharLimit = 256
		ti.Width = m.width - 10
		var parentID *int64
		if len(m.entries) > 0 {
			parentID = m.entries[m.selectedIdx].Entry.ParentID
		}
		m.addMode = addState{
			active:   true,
			asChild:  false,
			parentID: parentID,
			input:    ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.AddChild):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		if !entry.CanAddChild() {
			return m, nil
		}
		ti := textinput.New()
		ti.Placeholder = ". task, - note, o event"
		ti.Focus()
		ti.CharLimit = 256
		ti.Width = m.width - 10
		parentID := entry.ID
		m.addMode = addState{
			active:   true,
			asChild:  true,
			parentID: &parentID,
			input:    ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.AddRoot):
		ti := textinput.New()
		ti.Placeholder = ". task, - note, o event"
		ti.Focus()
		ti.CharLimit = 256
		ti.Width = m.width - 10
		m.addMode = addState{
			active:   true,
			asChild:  false,
			parentID: nil,
			input:    ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Migrate):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		if entry.Type != domain.EntryTypeTask {
			return m, nil
		}
		ti := textinput.New()
		ti.Placeholder = "tomorrow, next monday, 2026-01-15"
		ti.Focus()
		ti.CharLimit = 64
		ti.Width = m.width - 10
		fromDate := m.viewDate
		if entry.ScheduledDate != nil {
			fromDate = *entry.ScheduledDate
		}
		m.migrateMode = migrateState{
			active:   true,
			entryID:  entry.ID,
			fromDate: fromDate,
			input:    ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.MigrateToGoal):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		if entry.Type != domain.EntryTypeTask {
			return m, nil
		}
		ti := textinput.New()
		ti.Placeholder = "Target month (YYYY-MM)"
		ti.Focus()
		ti.CharLimit = 7
		ti.Width = m.width - 10
		m.migrateToGoalMode = migrateToGoalState{
			active:  true,
			entryID: entry.ID,
			content: entry.Content,
			input:   ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.MoveToList):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		if !entry.CanMoveToList() {
			return m, nil
		}
		return m, m.loadListsForMoveCmd(entry.ID)

	case key.Matches(msg, m.keyMap.MoveToRoot):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		if !entry.CanMoveToRoot() {
			return m, nil
		}
		return m, m.moveToRootCmd(entry.ID)

	case key.Matches(msg, m.keyMap.Priority):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		newPriority := entry.Priority.Cycle()
		return m, m.cyclePriorityCmd(entry.ID, newPriority)

	case key.Matches(msg, m.keyMap.GotoDate):
		ti := textinput.New()
		ti.Placeholder = "today, yesterday, 2026-01-15"
		ti.Focus()
		ti.CharLimit = 64
		ti.Width = m.width - 10
		m.gotoMode = gotoState{
			active: true,
			input:  ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.DayLeft):
		if m.viewMode == ViewModeDay {
			m.viewDate = m.viewDate.AddDate(0, 0, -1)
		} else {
			m.viewDate = m.viewDate.AddDate(0, 0, -7)
		}
		m.selectedIdx = 0
		return m, m.loadDaysCmd()

	case key.Matches(msg, m.keyMap.DayRight):
		if m.viewMode == ViewModeDay {
			m.viewDate = m.viewDate.AddDate(0, 0, 1)
		} else {
			m.viewDate = m.viewDate.AddDate(0, 0, 7)
		}
		m.selectedIdx = 0
		return m, m.loadDaysCmd()

	case key.Matches(msg, m.keyMap.GotoToday):
		m.viewDate = time.Now()
		m.selectedIdx = 0
		return m, m.loadDaysCmd()

	case key.Matches(msg, m.keyMap.Help):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil

	case key.Matches(msg, m.keyMap.SetLocation):
		input := textinput.New()
		input.Placeholder = "Enter location..."
		input.Focus()
		m.setLocationMode = setLocationState{
			active:      true,
			pickerMode:  true,
			date:        m.viewDate,
			input:       input,
			locations:   nil,
			selectedIdx: 0,
		}
		return m, m.loadLocationsCmd()

	case key.Matches(msg, m.keyMap.SetMood):
		input := textinput.New()
		input.Placeholder = "Enter mood..."
		input.Focus()
		m.setMoodMode = setMoodState{
			active:      true,
			pickerMode:  true,
			date:        m.viewDate,
			input:       input,
			presets:     moodPresets,
			selectedIdx: 0,
		}
		return m, m.loadMoodsCmd()

	case key.Matches(msg, m.keyMap.SetWeather):
		input := textinput.New()
		input.Placeholder = "Enter weather..."
		input.Focus()
		m.setWeatherMode = setWeatherState{
			active:      true,
			pickerMode:  true,
			date:        m.viewDate,
			input:       input,
			presets:     weatherPresets,
			selectedIdx: 0,
		}
		return m, m.loadWeathersCmd()

	case key.Matches(msg, m.keyMap.Capture):
		return m, m.launchExternalEditorCmd()
	}

	switch msg.Type {
	case tea.KeyCtrlS:
		m.searchMode = searchState{active: true, forward: true}
		return m, nil
	case tea.KeyCtrlR:
		m.searchMode = searchState{active: true, forward: false}
		return m, nil
	}

	return m, nil
}

func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchMode = searchState{}
		return m, nil

	case tea.KeyEnter:
		m = m.searchEntries()
		m.searchMode.active = false
		return m.ensuredVisible(), nil

	case tea.KeyCtrlS:
		m.searchMode.forward = true
		m = m.searchEntriesNext()
		return m.ensuredVisible(), nil

	case tea.KeyCtrlR:
		m.searchMode.forward = false
		m = m.searchEntriesNext()
		return m.ensuredVisible(), nil

	case tea.KeyBackspace:
		if len(m.searchMode.query) > 0 {
			m.searchMode.query = m.searchMode.query[:len(m.searchMode.query)-1]
			m = m.searchEntries()
		}
		return m.ensuredVisible(), nil

	case tea.KeySpace:
		m.searchMode.query += " "
		m = m.searchEntries()
		return m.ensuredVisible(), nil

	case tea.KeyRunes:
		m.searchMode.query += string(msg.Runes)
		m = m.searchEntries()
		return m.ensuredVisible(), nil
	}

	return m, nil
}

func (m Model) searchEntries() Model {
	if m.searchMode.query == "" || len(m.entries) == 0 {
		return m
	}

	query := strings.ToLower(m.searchMode.query)

	if strings.Contains(strings.ToLower(m.entries[m.selectedIdx].Entry.Content), query) {
		return m
	}

	start := m.selectedIdx

	if m.searchMode.forward {
		for i := 1; i < len(m.entries); i++ {
			idx := (start + i) % len(m.entries)
			if strings.Contains(strings.ToLower(m.entries[idx].Entry.Content), query) {
				m.selectedIdx = idx
				return m
			}
		}
	} else {
		for i := 1; i < len(m.entries); i++ {
			idx := (start - i + len(m.entries)) % len(m.entries)
			if strings.Contains(strings.ToLower(m.entries[idx].Entry.Content), query) {
				m.selectedIdx = idx
				return m
			}
		}
	}

	return m
}

func (m Model) searchEntriesNext() Model {
	if m.searchMode.query == "" || len(m.entries) == 0 {
		return m
	}

	query := strings.ToLower(m.searchMode.query)
	start := m.selectedIdx

	if m.searchMode.forward {
		for i := 1; i <= len(m.entries); i++ {
			idx := (start + i) % len(m.entries)
			if strings.Contains(strings.ToLower(m.entries[idx].Entry.Content), query) {
				m.selectedIdx = idx
				return m
			}
		}
	} else {
		for i := 1; i <= len(m.entries); i++ {
			idx := (start - i + len(m.entries)) % len(m.entries)
			if strings.Contains(strings.ToLower(m.entries[idx].Entry.Content), query) {
				m.selectedIdx = idx
				return m
			}
		}
	}

	return m
}

func (m Model) saveCaptureCmd(content string) tea.Cmd {
	viewDate := m.viewDate
	return func() tea.Msg {
		ctx := context.Background()
		opts := service.LogEntriesOptions{Date: viewDate}
		_, err := m.bujoService.LogEntries(ctx, content, opts)
		if err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{0}
	}
}

func (m Model) handleConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Confirm):
		entryID := m.confirmMode.entryID
		m.confirmMode.active = false

		if m.currentView == ViewTypeListItems {
			return m, m.deleteListItemCmd(entryID)
		}

		return m, m.deleteWithChildrenCmd(entryID)

	case key.Matches(msg, m.keyMap.Cancel):
		m.confirmMode.active = false
		return m, nil
	}

	return m, nil
}

func (m Model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editMode.active = false
		return m, nil

	case tea.KeyEnter:
		entryID := m.editMode.entryID
		newContent := m.editMode.input.Value()
		m.editMode.active = false
		if m.currentView == ViewTypeListItems {
			return m, m.editListItemCmd(entryID, newContent)
		}
		return m, m.editEntryCmd(entryID, newContent)
	}

	var cmd tea.Cmd
	m.editMode.input, cmd = m.editMode.input.Update(msg)
	return m, cmd
}

func (m Model) editEntryCmd(id int64, content string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.EditEntry(ctx, id, content); err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{id}
	}
}

func (m Model) handleAnswerMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.answerMode.active = false
		return m, nil

	case tea.KeyEnter:
		questionID := m.answerMode.questionID
		answerText := m.answerMode.input.Value()
		m.answerMode.active = false
		if answerText == "" {
			return m, nil
		}
		return m, m.answerQuestionCmd(questionID, answerText)
	}

	var cmd tea.Cmd
	m.answerMode.input, cmd = m.answerMode.input.Update(msg)
	return m, cmd
}

func (m Model) answerQuestionCmd(id int64, answerText string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.MarkAnswered(ctx, id, answerText); err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{id}
	}
}

func (m Model) handleAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.addMode.active = false
		return m, nil

	case tea.KeyEnter:
		content := m.addMode.input.Value()
		if content == "" {
			m.addMode.active = false
			return m, nil
		}
		m.addMode.active = false

		if m.currentView == ViewTypeListItems {
			return m, m.addListItemCmd(content)
		}

		parentID := m.addMode.parentID
		return m, m.addEntryCmd(content, parentID)
	}

	var cmd tea.Cmd
	m.addMode.input, cmd = m.addMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleAddHabitMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.addHabitMode.active = false
		return m, nil

	case tea.KeyEnter:
		name := strings.TrimSpace(m.addHabitMode.input.Value())
		if name == "" {
			m.addHabitMode.active = false
			return m, nil
		}
		m.addHabitMode.active = false
		return m, m.addHabitCmd(name)
	}

	var cmd tea.Cmd
	m.addHabitMode.input, cmd = m.addHabitMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleConfirmHabitDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Confirm):
		habitID := m.confirmHabitDeleteMode.habitID
		m.confirmHabitDeleteMode.active = false
		return m, m.deleteHabitCmd(habitID)

	case key.Matches(msg, m.keyMap.Cancel):
		m.confirmHabitDeleteMode.active = false
		return m, nil
	}

	return m, nil
}

func (m Model) addEntryCmd(content string, parentID *int64) tea.Cmd {
	viewDate := m.viewDate
	return func() tea.Msg {
		ctx := context.Background()
		opts := service.LogEntriesOptions{Date: viewDate}
		ids, err := m.bujoService.LogEntries(ctx, content, opts)
		if err != nil {
			return errMsg{err}
		}
		if len(ids) == 0 {
			return entryUpdatedMsg{0}
		}

		if parentID != nil {
			moveOpts := service.MoveOptions{NewParentID: parentID}
			if err := m.bujoService.MoveEntry(ctx, ids[0], moveOpts); err != nil {
				return errMsg{err}
			}
		}
		return entryUpdatedMsg{ids[0]}
	}
}

func (m Model) handleMigrateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.migrateMode.active = false
		return m, nil

	case tea.KeyEnter:
		dateStr := m.migrateMode.input.Value()
		if dateStr == "" {
			m.migrateMode.active = false
			return m, nil
		}
		entryID := m.migrateMode.entryID
		fromDate := m.migrateMode.fromDate
		m.migrateMode.active = false
		return m, m.migrateEntryCmd(entryID, dateStr, fromDate)
	}

	var cmd tea.Cmd
	m.migrateMode.input, cmd = m.migrateMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleGotoMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.gotoMode.active = false
		return m, nil

	case tea.KeyEnter:
		dateStr := m.gotoMode.input.Value()
		if dateStr == "" {
			m.gotoMode.active = false
			return m, nil
		}
		m.gotoMode.active = false
		return m, m.gotoDateCmd(dateStr)
	}

	var cmd tea.Cmd
	m.gotoMode.input, cmd = m.gotoMode.input.Update(msg)
	return m, cmd
}

func (m Model) gotoDateCmd(dateStr string) tea.Cmd {
	return func() tea.Msg {
		toDate, err := parseDate(dateStr)
		if err != nil {
			return errMsg{err}
		}
		return gotoDateMsg{date: toDate}
	}
}

func (m Model) handleSetLocationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.setLocationMode.active = false
		return m, nil

	case tea.KeyUp:
		if m.setLocationMode.pickerMode && len(m.setLocationMode.locations) > 0 {
			if m.setLocationMode.selectedIdx > 0 {
				m.setLocationMode.selectedIdx--
			}
			return m, nil
		}

	case tea.KeyDown:
		if m.setLocationMode.pickerMode && len(m.setLocationMode.locations) > 0 {
			if m.setLocationMode.selectedIdx < len(m.setLocationMode.locations)-1 {
				m.setLocationMode.selectedIdx++
			}
			return m, nil
		}

	case tea.KeyEnter:
		var location string
		inputValue := m.setLocationMode.input.Value()
		if inputValue != "" {
			location = inputValue
		} else if m.setLocationMode.pickerMode && len(m.setLocationMode.locations) > 0 {
			location = m.setLocationMode.locations[m.setLocationMode.selectedIdx]
		}
		m.setLocationMode.active = false
		if location == "" {
			return m, nil
		}
		return m, m.setLocationCmd(m.setLocationMode.date, location)
	}

	var cmd tea.Cmd
	m.setLocationMode.input, cmd = m.setLocationMode.input.Update(msg)
	return m, cmd
}

func (m Model) setLocationCmd(date time.Time, location string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.bujoService.SetLocation(ctx, date, location)
		if err != nil {
			return errMsg{err}
		}
		return agendaReloadNeededMsg{}
	}
}

func (m Model) handleSetMoodMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.setMoodMode.active = false
		return m, nil

	case tea.KeyUp:
		if m.setMoodMode.pickerMode && len(m.setMoodMode.presets) > 0 {
			if m.setMoodMode.selectedIdx > 0 {
				m.setMoodMode.selectedIdx--
			}
			return m, nil
		}

	case tea.KeyDown:
		if m.setMoodMode.pickerMode && len(m.setMoodMode.presets) > 0 {
			if m.setMoodMode.selectedIdx < len(m.setMoodMode.presets)-1 {
				m.setMoodMode.selectedIdx++
			}
			return m, nil
		}

	case tea.KeyEnter:
		var mood string
		inputValue := m.setMoodMode.input.Value()
		if inputValue != "" {
			mood = inputValue
		} else if m.setMoodMode.pickerMode && len(m.setMoodMode.presets) > 0 {
			mood = m.setMoodMode.presets[m.setMoodMode.selectedIdx]
		}
		m.setMoodMode.active = false
		if mood == "" {
			return m, nil
		}
		return m, m.setMoodCmd(m.setMoodMode.date, mood)
	}

	var cmd tea.Cmd
	m.setMoodMode.input, cmd = m.setMoodMode.input.Update(msg)
	return m, cmd
}

func (m Model) setMoodCmd(date time.Time, mood string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if m.bujoService == nil {
			return nil
		}
		err := m.bujoService.SetMood(ctx, date, mood)
		if err != nil {
			return errMsg{err}
		}
		return agendaReloadNeededMsg{}
	}
}

func (m Model) handleSetWeatherMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.setWeatherMode.active = false
		return m, nil

	case tea.KeyUp:
		if m.setWeatherMode.pickerMode && len(m.setWeatherMode.presets) > 0 {
			if m.setWeatherMode.selectedIdx > 0 {
				m.setWeatherMode.selectedIdx--
			}
			return m, nil
		}

	case tea.KeyDown:
		if m.setWeatherMode.pickerMode && len(m.setWeatherMode.presets) > 0 {
			if m.setWeatherMode.selectedIdx < len(m.setWeatherMode.presets)-1 {
				m.setWeatherMode.selectedIdx++
			}
			return m, nil
		}

	case tea.KeyEnter:
		var weather string
		inputValue := m.setWeatherMode.input.Value()
		if inputValue != "" {
			weather = inputValue
		} else if m.setWeatherMode.pickerMode && len(m.setWeatherMode.presets) > 0 {
			weather = m.setWeatherMode.presets[m.setWeatherMode.selectedIdx]
		}
		m.setWeatherMode.active = false
		if weather == "" {
			return m, nil
		}
		return m, m.setWeatherCmd(m.setWeatherMode.date, weather)
	}

	var cmd tea.Cmd
	m.setWeatherMode.input, cmd = m.setWeatherMode.input.Update(msg)
	return m, cmd
}

func (m Model) setWeatherCmd(date time.Time, weather string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if m.bujoService == nil {
			return nil
		}
		err := m.bujoService.SetWeather(ctx, date, weather)
		if err != nil {
			return errMsg{err}
		}
		return agendaReloadNeededMsg{}
	}
}

func (m Model) loadMoodsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return moodsLoadedMsg{moods: nil}
		}
		ctx := context.Background()
		now := time.Now()
		from := now.AddDate(0, -locationHistoryMonths, 0)
		history, err := m.bujoService.GetMoodHistory(ctx, from, now)
		if err != nil {
			return moodsLoadedMsg{moods: nil}
		}
		seen := make(map[string]bool)
		var moods []string
		for _, dayCtx := range history {
			if dayCtx.Mood != nil && *dayCtx.Mood != "" && !seen[*dayCtx.Mood] {
				seen[*dayCtx.Mood] = true
				moods = append(moods, *dayCtx.Mood)
			}
		}
		return moodsLoadedMsg{moods: moods}
	}
}

func (m Model) loadWeathersCmd() tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return weathersLoadedMsg{weathers: nil}
		}
		ctx := context.Background()
		now := time.Now()
		from := now.AddDate(0, -locationHistoryMonths, 0)
		history, err := m.bujoService.GetWeatherHistory(ctx, from, now)
		if err != nil {
			return weathersLoadedMsg{weathers: nil}
		}
		seen := make(map[string]bool)
		var weathers []string
		for _, dayCtx := range history {
			if dayCtx.Weather != nil && *dayCtx.Weather != "" && !seen[*dayCtx.Weather] {
				seen[*dayCtx.Weather] = true
				weathers = append(weathers, *dayCtx.Weather)
			}
		}
		return weathersLoadedMsg{weathers: weathers}
	}
}

func mergePresets(defaults []string, history []string) []string {
	seen := make(map[string]bool)
	result := make([]string, len(defaults))
	copy(result, defaults)
	for _, d := range defaults {
		seen[d] = true
	}
	for _, h := range history {
		if !seen[h] {
			seen[h] = true
			result = append(result, h)
		}
	}
	return result
}

func (m Model) handleRetypeMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	types := []domain.EntryType{domain.EntryTypeTask, domain.EntryTypeNote, domain.EntryTypeEvent}

	switch msg.Type {
	case tea.KeyEsc:
		m.retypeMode.active = false
		return m, nil

	case tea.KeyEnter:
		m.retypeMode.active = false
		newType := types[m.retypeMode.selectedIdx]
		return m, m.retypeEntryCmd(newType)
	}

	switch {
	case key.Matches(msg, m.keyMap.Up):
		if m.retypeMode.selectedIdx > 0 {
			m.retypeMode.selectedIdx--
		}
	case key.Matches(msg, m.keyMap.Down):
		if m.retypeMode.selectedIdx < len(types)-1 {
			m.retypeMode.selectedIdx++
		}
	case msg.Type == tea.KeyRunes:
		switch string(msg.Runes) {
		case ".", "1":
			m.retypeMode.active = false
			return m, m.retypeEntryCmd(domain.EntryTypeTask)
		case "-", "2":
			m.retypeMode.active = false
			return m, m.retypeEntryCmd(domain.EntryTypeNote)
		case "o", "O", "3":
			m.retypeMode.active = false
			return m, m.retypeEntryCmd(domain.EntryTypeEvent)
		}
	}

	return m, nil
}

func (m Model) migrateEntryCmd(id int64, dateStr string, fromDate time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		toDate, err := parseDate(dateStr)
		if err != nil {
			return errMsg{err}
		}
		newID, err := m.bujoService.MigrateEntry(ctx, id, toDate)
		if err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{newID}
	}
}

func (m Model) toggleDoneForEntryCmd(entry domain.Entry) tea.Cmd {
	validTypes := entry.Type == domain.EntryTypeTask ||
		entry.Type == domain.EntryTypeDone ||
		entry.Type == domain.EntryTypeAnswered

	if !validTypes {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()
		var err error

		switch entry.Type {
		case domain.EntryTypeDone:
			err = m.bujoService.Undo(ctx, entry.ID)
		case domain.EntryTypeTask:
			err = m.bujoService.MarkDone(ctx, entry.ID)
		case domain.EntryTypeAnswered:
			err = m.bujoService.ReopenQuestion(ctx, entry.ID)
		}

		if err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entry.ID}
	}
}

func (m Model) toggleDoneCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	return m.toggleDoneForEntryCmd(m.entries[m.selectedIdx].Entry)
}

func (m Model) cyclePriorityCmd(entryID int64, newPriority domain.Priority) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.bujoService.EditEntryPriority(ctx, entryID, newPriority)
		if err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entryID}
	}
}

func (m Model) cancelForEntryCmd(entry domain.Entry) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.CancelEntry(ctx, entry.ID); err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entry.ID}
	}
}

func (m Model) cancelEntryCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	return m.cancelForEntryCmd(m.entries[m.selectedIdx].Entry)
}

func (m Model) uncancelForEntryCmd(entry domain.Entry) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.UncancelEntry(ctx, entry.ID); err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entry.ID}
	}
}

func (m Model) uncancelEntryCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	return m.uncancelForEntryCmd(m.entries[m.selectedIdx].Entry)
}

func (m Model) retypeEntryCmd(newType domain.EntryType) tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.selectedIdx].Entry

	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.RetypeEntry(ctx, entry.ID, newType); err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entry.ID}
	}
}

func (m Model) deleteWithChildrenCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.DeleteEntry(ctx, id); err != nil {
			return errMsg{err}
		}
		return entryDeletedMsg{id}
	}
}

func parseDate(s string) (time.Time, error) {
	return dateutil.ParseFuture(s)
}

func (m Model) handleHabitsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case key.Matches(msg, m.keyMap.Down):
		if m.habitState.selectedIdx < len(m.habitState.habits)-1 {
			m.habitState.selectedIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Up):
		if m.habitState.selectedIdx > 0 {
			m.habitState.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		if len(m.habitState.habits) > 0 && m.habitState.selectedIdx < len(m.habitState.habits) {
			days := HabitDaysWeek
			switch m.habitState.viewMode {
			case HabitViewModeMonth:
				days = HabitDaysMonth
			case HabitViewModeQuarter:
				days = HabitDaysQuarter
			}
			daysAgo := days - 1 - m.habitState.selectedDayIdx
			logDate := m.getHabitReferenceDate().AddDate(0, 0, -daysAgo)
			return m, m.logHabitForDateCmd(m.habitState.habits[m.habitState.selectedIdx].ID, logDate)
		}
		return m, nil

	case key.Matches(msg, m.keyMap.DayLeft):
		if m.habitState.selectedDayIdx > 0 {
			m.habitState.selectedDayIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.DayRight):
		days := HabitDaysWeek
		switch m.habitState.viewMode {
		case HabitViewModeMonth:
			days = HabitDaysMonth
		case HabitViewModeQuarter:
			days = HabitDaysQuarter
		}
		if m.habitState.selectedDayIdx < days-1 {
			m.habitState.selectedDayIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.RemoveHabitLog):
		if len(m.habitState.habits) > 0 && m.habitState.selectedIdx < len(m.habitState.habits) {
			days := HabitDaysWeek
			switch m.habitState.viewMode {
			case HabitViewModeMonth:
				days = HabitDaysMonth
			case HabitViewModeQuarter:
				days = HabitDaysQuarter
			}
			daysAgo := days - 1 - m.habitState.selectedDayIdx
			removeDate := m.getHabitReferenceDate().AddDate(0, 0, -daysAgo)
			return m, m.removeHabitLogForDateCmd(m.habitState.habits[m.habitState.selectedIdx].ID, removeDate)
		}
		return m, nil

	case key.Matches(msg, m.keyMap.PrevPeriod):
		m.habitState.weekOffset++
		return m, m.loadHabitsCmd()

	case key.Matches(msg, m.keyMap.NextPeriod):
		if m.habitState.weekOffset > 0 {
			m.habitState.weekOffset--
			return m, m.loadHabitsCmd()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.ToggleView):
		switch m.habitState.viewMode {
		case HabitViewModeWeek:
			m.habitState.viewMode = HabitViewModeMonth
		case HabitViewModeMonth:
			m.habitState.viewMode = HabitViewModeQuarter
		case HabitViewModeQuarter:
			m.habitState.viewMode = HabitViewModeWeek
		}
		// Reset weekOffset when switching view modes because the offset represents
		// different absolute time spans in each mode (7 days/week vs 30 days/month).
		// Without resetting, users would see unexpected historical data when switching modes.
		m.habitState.weekOffset = 0
		return m, m.loadHabitsCmd()

	case key.Matches(msg, m.keyMap.Add):
		ti := textinput.New()
		ti.Placeholder = "Habit name"
		ti.Focus()
		ti.CharLimit = 100
		ti.Width = m.width - 10
		m.addHabitMode = addHabitState{
			active: true,
			input:  ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Delete):
		if len(m.habitState.habits) > 0 && m.habitState.selectedIdx < len(m.habitState.habits) {
			m.confirmHabitDeleteMode = confirmHabitDeleteState{
				active:  true,
				habitID: m.habitState.habits[m.habitState.selectedIdx].ID,
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handlePendingTasksMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case key.Matches(msg, m.keyMap.Down):
		if m.pendingTasksState.selectedIdx < len(m.pendingTasksState.entries)-1 {
			m.pendingTasksState.selectedIdx++
			m = m.ensurePendingTaskVisible()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Up):
		if m.pendingTasksState.selectedIdx > 0 {
			m.pendingTasksState.selectedIdx--
			m = m.ensurePendingTaskVisible()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Top):
		m.pendingTasksState.selectedIdx = 0
		m.pendingTasksState.scrollOffset = 0
		return m, nil

	case key.Matches(msg, m.keyMap.Bottom):
		if len(m.pendingTasksState.entries) > 0 {
			m.pendingTasksState.selectedIdx = len(m.pendingTasksState.entries) - 1
			m = m.ensurePendingTaskVisible()
		}
		return m, nil

	case msg.Type == tea.KeyEnter:
		if len(m.pendingTasksState.entries) > 0 &&
			m.pendingTasksState.selectedIdx < len(m.pendingTasksState.entries) {
			entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
			if entry.ScheduledDate != nil {
				m.viewStack = append(m.viewStack, m.currentView)
				m.currentView = ViewTypeJournal
				m.viewDate = *entry.ScheduledDate
				m.viewMode = ViewModeDay
				m.selectedIdx = 0
				return m, m.loadDaysCmd()
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		return m, m.toggleDoneForEntryCmd(entry)

	case key.Matches(msg, m.keyMap.CancelEntry):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		if !entry.CanCancel() {
			return m, nil
		}
		return m, m.cancelForEntryCmd(entry)

	case key.Matches(msg, m.keyMap.UncancelEntry):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		if !entry.CanUncancel() {
			return m, nil
		}
		return m, m.uncancelForEntryCmd(entry)

	case key.Matches(msg, m.keyMap.Edit):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		ti := textinput.New()
		ti.SetValue(entry.Content)
		ti.Focus()
		ti.CharLimit = 256
		ti.Width = m.width - 10
		m.editMode = editState{
			active:  true,
			entryID: entry.ID,
			input:   ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Delete):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		m.confirmMode.active = true
		m.confirmMode.entryID = entry.ID
		return m, nil

	case key.Matches(msg, m.keyMap.Migrate):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		if entry.Type != domain.EntryTypeTask {
			return m, nil
		}
		ti := textinput.New()
		ti.Placeholder = "tomorrow, next monday, 2026-01-15"
		ti.Focus()
		ti.CharLimit = 64
		ti.Width = m.width - 10
		fromDate := m.viewDate
		if entry.ScheduledDate != nil {
			fromDate = *entry.ScheduledDate
		}
		m.migrateMode = migrateState{
			active:   true,
			entryID:  entry.ID,
			fromDate: fromDate,
			input:    ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Retype):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		if !entry.CanCycleType() {
			return m, nil
		}
		m.retypeMode = retypeState{
			active:      true,
			entryID:     entry.ID,
			selectedIdx: 0,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Priority):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		newPriority := entry.Priority.Cycle()
		return m, m.cyclePriorityCmd(entry.ID, newPriority)

	case key.Matches(msg, m.keyMap.Answer):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		if entry.Type != domain.EntryTypeQuestion {
			return m, nil
		}
		ti := textinput.New()
		ti.Placeholder = "Enter your answer..."
		ti.Focus()
		ti.CharLimit = 512
		ti.Width = m.width - 10
		m.answerMode = answerState{
			active:     true,
			questionID: entry.ID,
			input:      ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.MoveToList):
		if len(m.pendingTasksState.entries) == 0 {
			return m, nil
		}
		entry := m.pendingTasksState.entries[m.pendingTasksState.selectedIdx]
		if !entry.CanMoveToList() {
			return m, nil
		}
		return m, m.loadListsForMoveCmd(entry.ID)
	}

	return m, nil
}

func (m Model) handleQuestionsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case key.Matches(msg, m.keyMap.Down):
		if m.questionsState.selectedIdx < len(m.questionsState.entries)-1 {
			m.questionsState.selectedIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Up):
		if m.questionsState.selectedIdx > 0 {
			m.questionsState.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Top):
		m.questionsState.selectedIdx = 0
		return m, nil

	case key.Matches(msg, m.keyMap.Bottom):
		if len(m.questionsState.entries) > 0 {
			m.questionsState.selectedIdx = len(m.questionsState.entries) - 1
		}
		return m, nil

	case msg.Type == tea.KeyEnter:
		if len(m.questionsState.entries) > 0 &&
			m.questionsState.selectedIdx < len(m.questionsState.entries) {
			entry := m.questionsState.entries[m.questionsState.selectedIdx]
			if entry.ScheduledDate != nil {
				m.viewStack = append(m.viewStack, m.currentView)
				m.currentView = ViewTypeJournal
				m.viewDate = *entry.ScheduledDate
				m.viewMode = ViewModeDay
				m.selectedIdx = 0
				return m, m.loadDaysCmd()
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		if len(m.questionsState.entries) == 0 {
			return m, nil
		}
		entry := m.questionsState.entries[m.questionsState.selectedIdx]
		return m, m.toggleDoneForEntryCmd(entry)

	case key.Matches(msg, m.keyMap.CancelEntry):
		if len(m.questionsState.entries) == 0 {
			return m, nil
		}
		entry := m.questionsState.entries[m.questionsState.selectedIdx]
		if !entry.CanCancel() {
			return m, nil
		}
		return m, m.cancelForEntryCmd(entry)

	case key.Matches(msg, m.keyMap.UncancelEntry):
		if len(m.questionsState.entries) == 0 {
			return m, nil
		}
		entry := m.questionsState.entries[m.questionsState.selectedIdx]
		if !entry.CanUncancel() {
			return m, nil
		}
		return m, m.uncancelForEntryCmd(entry)

	case key.Matches(msg, m.keyMap.Edit):
		if len(m.questionsState.entries) == 0 {
			return m, nil
		}
		entry := m.questionsState.entries[m.questionsState.selectedIdx]
		ti := textinput.New()
		ti.SetValue(entry.Content)
		ti.Focus()
		ti.CharLimit = 256
		ti.Width = m.width - 10
		m.editMode = editState{
			active:  true,
			entryID: entry.ID,
			input:   ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Delete):
		if len(m.questionsState.entries) == 0 {
			return m, nil
		}
		entry := m.questionsState.entries[m.questionsState.selectedIdx]
		m.confirmMode.active = true
		m.confirmMode.entryID = entry.ID
		return m, nil

	case key.Matches(msg, m.keyMap.Answer):
		if len(m.questionsState.entries) == 0 {
			return m, nil
		}
		entry := m.questionsState.entries[m.questionsState.selectedIdx]
		if entry.Type != domain.EntryTypeQuestion {
			return m, nil
		}
		ti := textinput.New()
		ti.Placeholder = "Enter your answer..."
		ti.Focus()
		ti.CharLimit = 512
		ti.Width = m.width - 10
		m.answerMode = answerState{
			active:     true,
			questionID: entry.ID,
			input:      ti,
		}
		return m, nil
	}

	return m, nil
}

func (m Model) ensurePendingTaskVisible() Model {
	maxLines := m.pendingTasksVisibleRows()
	if maxLines <= 0 {
		return m
	}

	if m.pendingTasksState.selectedIdx < m.pendingTasksState.scrollOffset {
		m.pendingTasksState.scrollOffset = m.pendingTasksState.selectedIdx
	}

	for {
		visibleCount := m.pendingTasksVisibleCount(m.pendingTasksState.scrollOffset, maxLines)
		lastVisible := m.pendingTasksState.scrollOffset + visibleCount - 1
		if m.pendingTasksState.selectedIdx <= lastVisible {
			break
		}
		m.pendingTasksState.scrollOffset++
		if m.pendingTasksState.scrollOffset >= len(m.pendingTasksState.entries) {
			m.pendingTasksState.scrollOffset = len(m.pendingTasksState.entries) - 1
			break
		}
	}

	return m
}

func (m Model) pendingTasksVisibleCount(startIdx int, maxLines int) int {
	if startIdx >= len(m.pendingTasksState.entries) {
		return 0
	}

	linesUsed := 0
	if startIdx > 0 {
		linesUsed++
	}

	var currentDateStr string
	count := 0
	for i := startIdx; i < len(m.pendingTasksState.entries); i++ {
		entry := m.pendingTasksState.entries[i]

		entryDateStr := ""
		if entry.ScheduledDate != nil {
			entryDateStr = entry.ScheduledDate.Format("2006-01-02")
		}

		linesNeeded := 1
		if entryDateStr != currentDateStr {
			linesNeeded += 2
		}

		if linesUsed+linesNeeded > maxLines && count > 0 {
			break
		}

		if entryDateStr != currentDateStr {
			currentDateStr = entryDateStr
		}
		linesUsed += linesNeeded
		count++
	}

	return count
}

func (m Model) pendingTasksVisibleRows() int {
	return m.height - pendingTasksHeaderLines - pendingTasksFooterLines
}

func (m Model) handleGoalsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case key.Matches(msg, m.keyMap.Down):
		if m.goalState.selectedIdx < len(m.goalState.goals)-1 {
			m.goalState.selectedIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Up):
		if m.goalState.selectedIdx > 0 {
			m.goalState.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		if len(m.goalState.goals) > 0 && m.goalState.selectedIdx < len(m.goalState.goals) {
			goal := m.goalState.goals[m.goalState.selectedIdx]
			return m, m.toggleGoalCmd(goal.ID, goal.IsDone())
		}
		return m, nil

	case key.Matches(msg, m.keyMap.DayLeft):
		m.goalState.viewMonth = m.goalState.viewMonth.AddDate(0, -1, 0)
		m.goalState.selectedIdx = 0
		return m, m.loadGoalsCmd()

	case key.Matches(msg, m.keyMap.DayRight):
		m.goalState.viewMonth = m.goalState.viewMonth.AddDate(0, 1, 0)
		m.goalState.selectedIdx = 0
		return m, m.loadGoalsCmd()

	case key.Matches(msg, m.keyMap.Add):
		ti := textinput.New()
		ti.Placeholder = "Goal content"
		ti.Focus()
		ti.CharLimit = 200
		ti.Width = m.width - 10
		m.addGoalMode = addGoalState{
			active: true,
			input:  ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Edit):
		if len(m.goalState.goals) > 0 && m.goalState.selectedIdx < len(m.goalState.goals) {
			goal := m.goalState.goals[m.goalState.selectedIdx]
			ti := textinput.New()
			ti.Placeholder = "Goal content"
			ti.SetValue(goal.Content)
			ti.Focus()
			ti.CharLimit = 200
			ti.Width = m.width - 10
			m.editGoalMode = editGoalState{
				active: true,
				goalID: goal.ID,
				input:  ti,
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Delete):
		if len(m.goalState.goals) > 0 && m.goalState.selectedIdx < len(m.goalState.goals) {
			m.confirmGoalDeleteMode = confirmGoalDeleteState{
				active: true,
				goalID: m.goalState.goals[m.goalState.selectedIdx].ID,
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Migrate):
		if len(m.goalState.goals) > 0 && m.goalState.selectedIdx < len(m.goalState.goals) {
			goal := m.goalState.goals[m.goalState.selectedIdx]
			ti := textinput.New()
			ti.Placeholder = "Target month (YYYY-MM)"
			ti.Focus()
			ti.CharLimit = 7
			ti.Width = m.width - 10
			m.moveGoalMode = moveGoalState{
				active: true,
				goalID: goal.ID,
				input:  ti,
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleAddGoalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.addGoalMode.active = false
		return m, nil

	case tea.KeyEnter:
		content := strings.TrimSpace(m.addGoalMode.input.Value())
		if content == "" {
			m.addGoalMode.active = false
			return m, nil
		}
		m.addGoalMode.active = false
		return m, m.addGoalCmd(content)
	}

	var cmd tea.Cmd
	m.addGoalMode.input, cmd = m.addGoalMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleEditGoalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editGoalMode.active = false
		return m, nil

	case tea.KeyEnter:
		content := strings.TrimSpace(m.editGoalMode.input.Value())
		if content == "" {
			m.editGoalMode.active = false
			return m, nil
		}
		goalID := m.editGoalMode.goalID
		m.editGoalMode.active = false
		return m, m.editGoalCmd(goalID, content)
	}

	var cmd tea.Cmd
	m.editGoalMode.input, cmd = m.editGoalMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleConfirmGoalDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Confirm):
		goalID := m.confirmGoalDeleteMode.goalID
		m.confirmGoalDeleteMode.active = false
		return m, m.deleteGoalCmd(goalID)

	case key.Matches(msg, m.keyMap.Cancel):
		m.confirmGoalDeleteMode.active = false
		return m, nil
	}

	return m, nil
}

func (m Model) handleMoveGoalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.moveGoalMode.active = false
		return m, nil

	case tea.KeyEnter:
		monthStr := strings.TrimSpace(m.moveGoalMode.input.Value())
		targetMonth, err := time.Parse("2006-01", monthStr)
		if err != nil {
			m.moveGoalMode.active = false
			m.err = fmt.Errorf("invalid month format, use YYYY-MM")
			return m, nil
		}
		goalID := m.moveGoalMode.goalID
		m.moveGoalMode.active = false
		return m, m.moveGoalCmd(goalID, targetMonth)
	}

	var cmd tea.Cmd
	m.moveGoalMode.input, cmd = m.moveGoalMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleMigrateToGoalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.migrateToGoalMode.active = false
		return m, nil

	case tea.KeyEnter:
		monthStr := strings.TrimSpace(m.migrateToGoalMode.input.Value())
		targetMonth, err := time.Parse("2006-01", monthStr)
		if err != nil {
			m.migrateToGoalMode.active = false
			m.err = fmt.Errorf("invalid month format, use YYYY-MM")
			return m, nil
		}
		entryID := m.migrateToGoalMode.entryID
		content := m.migrateToGoalMode.content
		m.migrateToGoalMode.active = false
		return m, m.migrateToGoalCmd(entryID, content, targetMonth)
	}

	var cmd tea.Cmd
	m.migrateToGoalMode.input, cmd = m.migrateToGoalMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleListsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.currentView == ViewTypeListItems {
		return m.handleListItemsMode(msg)
	}

	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case key.Matches(msg, m.keyMap.Down):
		if m.listState.selectedListIdx < len(m.listState.lists)-1 {
			m.listState.selectedListIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Up):
		if m.listState.selectedListIdx > 0 {
			m.listState.selectedListIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Add):
		ti := textinput.New()
		ti.Focus()
		ti.CharLimit = 256
		ti.Width = m.width - 10
		m.createListMode = createListState{
			active: true,
			input:  ti,
		}
		return m, nil

	case msg.Type == tea.KeyEnter:
		if len(m.listState.lists) > 0 && m.listState.selectedListIdx < len(m.listState.lists) {
			selectedList := m.listState.lists[m.listState.selectedListIdx]
			m.listState.currentListID = selectedList.ID
			m.listState.selectedItemIdx = 0
			m.currentView = ViewTypeListItems
			return m, m.loadListItemsCmd(selectedList.ID)
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleListItemsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case msg.Type == tea.KeyEscape:
		m.currentView = ViewTypeLists
		m.listState.items = nil
		m.listState.selectedItemIdx = 0
		return m, nil

	case key.Matches(msg, m.keyMap.Down):
		if m.listState.selectedItemIdx < len(m.listState.items)-1 {
			m.listState.selectedItemIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Up):
		if m.listState.selectedItemIdx > 0 {
			m.listState.selectedItemIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		if len(m.listState.items) > 0 && m.listState.selectedItemIdx < len(m.listState.items) {
			item := m.listState.items[m.listState.selectedItemIdx]
			return m, m.toggleListItemCmd(item)
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Add):
		m.addMode.active = true
		m.addMode.input.Reset()
		m.addMode.input.Focus()
		return m, nil

	case key.Matches(msg, m.keyMap.Edit):
		if len(m.listState.items) > 0 && m.listState.selectedItemIdx < len(m.listState.items) {
			item := m.listState.items[m.listState.selectedItemIdx]
			ti := textinput.New()
			ti.SetValue(item.Content)
			ti.Focus()
			ti.CharLimit = 256
			ti.Width = m.width - 10
			m.editMode = editState{
				active:  true,
				entryID: item.RowID,
				input:   ti,
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Delete):
		if len(m.listState.items) > 0 && m.listState.selectedItemIdx < len(m.listState.items) {
			item := m.listState.items[m.listState.selectedItemIdx]
			m.confirmMode.active = true
			m.confirmMode.entryID = item.RowID
		}
		return m, nil

	case key.Matches(msg, m.keyMap.MoveListItem):
		if len(m.listState.items) > 0 && m.listState.selectedItemIdx < len(m.listState.items) {
			item := m.listState.items[m.listState.selectedItemIdx]
			targetLists := make([]domain.List, 0, len(m.listState.lists)-1)
			for _, list := range m.listState.lists {
				if list.ID != m.listState.currentListID {
					targetLists = append(targetLists, list)
				}
			}
			if len(targetLists) > 0 {
				m.moveListItemMode = moveListItemState{
					active:      true,
					itemID:      item.RowID,
					targetLists: targetLists,
					selectedIdx: 0,
				}
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleMoveListItemMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.moveListItemMode.active = false
		return m, nil

	case tea.KeyUp:
		if m.moveListItemMode.selectedIdx > 0 {
			m.moveListItemMode.selectedIdx--
		}
		return m, nil

	case tea.KeyDown:
		if m.moveListItemMode.selectedIdx < len(m.moveListItemMode.targetLists)-1 {
			m.moveListItemMode.selectedIdx++
		}
		return m, nil

	case tea.KeyEnter:
		if len(m.moveListItemMode.targetLists) > 0 && m.moveListItemMode.selectedIdx < len(m.moveListItemMode.targetLists) {
			targetList := m.moveListItemMode.targetLists[m.moveListItemMode.selectedIdx]
			itemID := m.moveListItemMode.itemID
			fromListID := m.listState.currentListID
			m.moveListItemMode.active = false
			return m, m.moveListItemCmd(itemID, targetList.ID, fromListID)
		}
		return m, nil

	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			if r >= '1' && r <= '9' {
				idx := int(r - '1')
				if idx < len(m.moveListItemMode.targetLists) {
					targetList := m.moveListItemMode.targetLists[idx]
					itemID := m.moveListItemMode.itemID
					fromListID := m.listState.currentListID
					m.moveListItemMode.active = false
					return m, m.moveListItemCmd(itemID, targetList.ID, fromListID)
				}
			}
		}
	}

	return m, nil
}

func (m Model) handleCreateListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.createListMode.active = false
		return m, nil

	case tea.KeyEnter:
		listName := m.createListMode.input.Value()
		if listName == "" {
			m.createListMode.active = false
			return m, nil
		}
		m.createListMode.active = false
		return m, m.createListCmd(listName)
	}

	var cmd tea.Cmd
	m.createListMode.input, cmd = m.createListMode.input.Update(msg)
	return m, cmd
}

func (m Model) handleMoveToListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.moveToListMode.active = false
		return m, nil

	case tea.KeyUp:
		if m.moveToListMode.selectedIdx > 0 {
			m.moveToListMode.selectedIdx--
		}
		return m, nil

	case tea.KeyDown:
		if m.moveToListMode.selectedIdx < len(m.moveToListMode.targetLists)-1 {
			m.moveToListMode.selectedIdx++
		}
		return m, nil

	case tea.KeyEnter:
		if len(m.moveToListMode.targetLists) > 0 && m.moveToListMode.selectedIdx < len(m.moveToListMode.targetLists) {
			targetList := m.moveToListMode.targetLists[m.moveToListMode.selectedIdx]
			entryID := m.moveToListMode.entryID
			m.moveToListMode.active = false
			return m, m.moveEntryToListCmd(entryID, targetList.ID)
		}
		return m, nil

	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			if r >= '1' && r <= '9' {
				idx := int(r - '1')
				if idx < len(m.moveToListMode.targetLists) {
					targetList := m.moveToListMode.targetLists[idx]
					entryID := m.moveToListMode.entryID
					m.moveToListMode.active = false
					return m, m.moveEntryToListCmd(entryID, targetList.ID)
				}
			}
		}
	}

	return m, nil
}

func (m Model) handleCommandPaletteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEscape:
		m.commandPalette.active = false
		m.commandPalette.query = ""
		m.commandPalette.selectedIdx = 0
		return m, nil

	case msg.Type == tea.KeyEnter:
		if len(m.commandPalette.filtered) > 0 && m.commandPalette.selectedIdx < len(m.commandPalette.filtered) {
			cmd := m.commandPalette.filtered[m.commandPalette.selectedIdx]
			m.commandPalette.active = false
			m.commandPalette.query = ""
			m.commandPalette.selectedIdx = 0
			return cmd.Action(m)
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Down) || msg.String() == "down":
		if m.commandPalette.selectedIdx < len(m.commandPalette.filtered)-1 {
			m.commandPalette.selectedIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Up) || msg.String() == "up":
		if m.commandPalette.selectedIdx > 0 {
			m.commandPalette.selectedIdx--
		}
		return m, nil

	case msg.Type == tea.KeyBackspace:
		if len(m.commandPalette.query) > 0 {
			m.commandPalette.query = m.commandPalette.query[:len(m.commandPalette.query)-1]
			m.commandPalette.filtered = m.commandRegistry.Filter(m.commandPalette.query)
			m.commandPalette.selectedIdx = 0
		}
		return m, nil

	case msg.Type == tea.KeyRunes:
		m.commandPalette.query += string(msg.Runes)
		m.commandPalette.filtered = m.commandRegistry.Filter(m.commandPalette.query)
		m.commandPalette.selectedIdx = 0
		return m, nil
	}

	return m, nil
}

func (m Model) handleStatsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()
	}

	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	return m, nil
}
func (m Model) handleInsightsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()
	}

	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch msg.Type {
	case tea.KeyTab:
		m.insightsState.activeTab = (m.insightsState.activeTab + 1) % insightsTabCount
		return m, m.loadInsightsTabDataCmd()

	case tea.KeyShiftTab:
		m.insightsState.activeTab = (m.insightsState.activeTab - 1 + insightsTabCount) % insightsTabCount
		return m, m.loadInsightsTabDataCmd()
	}

	switch string(msg.Runes) {
	case "h":
		if m.insightsState.activeTab != InsightsTabDashboard {
			m.insightsState.weekAnchor = m.insightsState.weekAnchor.AddDate(0, 0, -7)
			return m, m.loadInsightsTabDataCmd()
		}
	case "l":
		if m.insightsState.activeTab != InsightsTabDashboard {
			m.insightsState.weekAnchor = m.insightsState.weekAnchor.AddDate(0, 0, 7)
			return m, m.loadInsightsTabDataCmd()
		}
	}

	return m, nil
}

func (m Model) handleSearchViewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m.handleQuit()

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case msg.String() == "enter":
		if len(m.searchView.results) > 0 && m.searchView.selectedIdx < len(m.searchView.results) {
			entry := m.searchView.results[m.searchView.selectedIdx]
			if entry.ScheduledDate != nil {
				m.viewDate = *entry.ScheduledDate
			}
			m.viewStack = append(m.viewStack, m.currentView)
			m.currentView = ViewTypeJournal
			m.viewMode = ViewModeDay
			m.selectedIdx = 0
			return m, m.loadDaysCmd()
		}
		return m, nil

	case msg.String() == "esc":
		m.searchView.input.SetValue("")
		m.searchView.query = ""
		m.searchView.results = nil
		m.searchView.selectedIdx = 0
		return m, nil

	case msg.String() == "/":
		m.searchView.input.Focus()
		return m, nil

	case key.Matches(msg, m.keyMap.Up):
		if m.searchView.selectedIdx > 0 {
			m.searchView.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Down):
		if m.searchView.selectedIdx < len(m.searchView.results)-1 {
			m.searchView.selectedIdx++
		}
		return m, nil
	}

	if m.searchView.input.Focused() {
		var cmd tea.Cmd
		m.searchView.input, cmd = m.searchView.input.Update(msg)
		newQuery := m.searchView.input.Value()
		if newQuery != m.searchView.query {
			m.searchView.loading = true
			return m, tea.Batch(cmd, m.searchEntriesCmd(newQuery))
		}
		return m, cmd
	}

	if len(m.searchView.results) > 0 && m.searchView.selectedIdx < len(m.searchView.results) {
		entry := m.searchView.results[m.searchView.selectedIdx]

		switch {
		case key.Matches(msg, m.keyMap.Done):
			return m, m.toggleDoneForEntryCmd(entry)

		case key.Matches(msg, m.keyMap.CancelEntry):
			if !entry.CanCancel() {
				return m, nil
			}
			return m, m.cancelForEntryCmd(entry)

		case key.Matches(msg, m.keyMap.UncancelEntry):
			if !entry.CanUncancel() {
				return m, nil
			}
			return m, m.uncancelForEntryCmd(entry)

		case key.Matches(msg, m.keyMap.Edit):
			ti := textinput.New()
			ti.SetValue(entry.Content)
			ti.Focus()
			ti.CharLimit = 256
			ti.Width = m.width - 10
			m.editMode = editState{
				active:  true,
				entryID: entry.ID,
				input:   ti,
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Delete):
			m.confirmMode.active = true
			m.confirmMode.entryID = entry.ID
			return m, nil

		case key.Matches(msg, m.keyMap.Priority):
			newPriority := entry.Priority.Cycle()
			return m, m.cyclePriorityCmd(entry.ID, newPriority)

		case key.Matches(msg, m.keyMap.Retype):
			if !entry.CanCycleType() {
				return m, nil
			}
			m.retypeMode = retypeState{
				active:      true,
				entryID:     entry.ID,
				selectedIdx: 0,
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Answer):
			if entry.Type != domain.EntryTypeQuestion {
				return m, nil
			}
			ti := textinput.New()
			ti.Placeholder = "Enter your answer..."
			ti.Focus()
			ti.CharLimit = 512
			ti.Width = m.width - 10
			m.answerMode = answerState{
				active:     true,
				questionID: entry.ID,
				input:      ti,
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) handleViewSwitch(msg tea.KeyMsg) (bool, Model, tea.Cmd) {
	var newView ViewType
	var cmd tea.Cmd
	switched := false

	switch {
	case key.Matches(msg, m.keyMap.ViewJournal):
		newView = ViewTypeJournal
		m.viewMode = ViewModeDay
		cmd = m.loadDaysCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewReview):
		newView = ViewTypeReview
		m.viewMode = ViewModeWeek
		cmd = m.loadDaysCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewPendingTasks):
		newView = ViewTypePendingTasks
		m.pendingTasksState.loading = true
		cmd = m.loadPendingTasksCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewQuestions):
		newView = ViewTypeQuestions
		m.questionsState.loading = true
		cmd = m.loadQuestionsCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewHabits):
		newView = ViewTypeHabits
		cmd = m.loadHabitsCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewLists):
		newView = ViewTypeLists
		cmd = m.loadListsCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewSearch):
		newView = ViewTypeSearch
		m.searchView.input.Focus()
		cmd = m.searchView.input.Cursor.BlinkCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewStats):
		newView = ViewTypeStats
		m.statsViewState.loading = true
		cmd = m.loadStatsCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewGoals):
		newView = ViewTypeGoals
		cmd = m.loadGoalsCmd()
		switched = true

	case key.Matches(msg, m.keyMap.ViewSettings):
		newView = ViewTypeSettings
		switched = true

	case key.Matches(msg, m.keyMap.ViewInsights):
		newView = ViewTypeInsights
		m.insightsState.loading = true
		cmd = m.loadInsightsDashboardCmd()
		switched = true
	}

	if switched && newView != m.currentView {
		m.viewStack = append(m.viewStack, m.currentView)
		m.currentView = newView
		return true, m, cmd
	}

	return false, m, nil
}

func (m Model) handleQuit() (Model, tea.Cmd) {
	m.quitConfirmMode.active = true
	return m, nil
}

func (m Model) handleBack() (Model, tea.Cmd) {
	if len(m.viewStack) > 0 {
		m.currentView = m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]

		var cmd tea.Cmd
		switch m.currentView {
		case ViewTypeJournal:
			cmd = m.loadDaysCmd()
		case ViewTypeHabits:
			cmd = m.loadHabitsCmd()
		case ViewTypeLists:
			cmd = m.loadListsCmd()
		case ViewTypeGoals:
			cmd = m.loadGoalsCmd()
		case ViewTypeStats:
			cmd = m.loadStatsCmd()
		case ViewTypeInsights:
			m.insightsState.loading = true
			cmd = m.loadInsightsDashboardCmd()
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) handleQuitConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m, tea.Quit
	case "n", "N":
		m.quitConfirmMode.active = false
		return m, nil
	case "esc":
		m.quitConfirmMode.active = false
		return m, nil
	}
	return m, nil
}

func (m Model) undoCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var err error

		switch m.undoState.operation {
		case UndoOpMarkDone:
			err = m.bujoService.Undo(ctx, m.undoState.entryID)
		case UndoOpMarkUndone:
			err = m.bujoService.MarkDone(ctx, m.undoState.entryID)
		}

		if err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{m.undoState.entryID}
	}
}
