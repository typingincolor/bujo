package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tj/go-naturaldate"
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

	case agendaLoadedMsg:
		m.agenda = msg.agenda
		m.entries = m.flattenAgenda(msg.agenda)
		// Keep selection if valid, otherwise reset to start
		if m.selectedIdx >= len(m.entries) {
			m.selectedIdx = 0
		}
		m.scrollOffset = 0
		return m.ensuredVisible(), m.loadJournalGoalsCmd()

	case journalGoalsLoadedMsg:
		m.journalGoals = msg.goals
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case entryUpdatedMsg, entryDeletedMsg:
		return m, m.loadAgendaCmd()

	case gotoDateMsg:
		m.viewDate = msg.date
		m.selectedIdx = 0
		return m, m.loadAgendaCmd()

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
		// Initialize selected day to today (rightmost in 7-day view)
		days := 7
		if m.habitState.monthView {
			days = 30
		}
		m.habitState.selectedDayIdx = days - 1
		return m, nil

	case habitLoggedMsg:
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
		return m, m.loadAgendaCmd()

	case summaryLoadedMsg:
		m.summaryState.loading = false
		m.summaryState.error = nil
		m.summaryState.summary = msg.summary
		return m, nil

	case summaryErrorMsg:
		m.summaryState.loading = false
		m.summaryState.error = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}
		if m.captureMode.active {
			return m.handleCaptureMode(msg)
		}
		if m.editMode.active {
			return m.handleEditMode(msg)
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

		// Check for command palette activation
		if key.Matches(msg, m.keyMap.CommandPalette) {
			m.commandPalette.active = true
			m.commandPalette.query = ""
			m.commandPalette.selectedIdx = 0
			m.commandPalette.filtered = m.commandRegistry.All()
			return m, nil
		}

		// View-specific handling
		switch m.currentView {
		case ViewTypeHabits:
			return m.handleHabitsMode(msg)
		case ViewTypeLists, ViewTypeListItems:
			return m.handleListsMode(msg)
		case ViewTypeGoals:
			return m.handleGoalsMode(msg)
		case ViewTypeStats:
			return m.handleStatsMode(msg)
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
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keyMap.Up):
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m.ensuredVisible(), nil

	case key.Matches(msg, m.keyMap.Down):
		if m.selectedIdx < len(m.entries)-1 {
			m.selectedIdx++
		}
		return m.ensuredVisible(), nil

	case key.Matches(msg, m.keyMap.Top):
		m.selectedIdx = 0
		m.scrollOffset = 0
		return m, nil

	case key.Matches(msg, m.keyMap.Bottom):
		if len(m.entries) > 0 {
			m.selectedIdx = len(m.entries) - 1
			// Scroll to show the bottom entry
			m = m.scrollToBottom()
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
				m.entries = m.flattenAgenda(m.agenda)
				return m.ensuredVisible(), nil
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		return m, m.toggleDoneCmd()

	case key.Matches(msg, m.keyMap.CancelEntry):
		if len(m.entries) > 0 {
			return m, m.cancelEntryCmd()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.UncancelEntry):
		if len(m.entries) > 0 {
			return m, m.uncancelEntryCmd()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Retype):
		if len(m.entries) > 0 {
			entry := m.entries[m.selectedIdx].Entry
			m.retypeMode = retypeState{
				active:      true,
				entryID:     entry.ID,
				selectedIdx: 0,
			}
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
		// Use entry's scheduled date as reference for natural date parsing
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

	case key.Matches(msg, m.keyMap.Priority):
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.selectedIdx].Entry
		newPriority := entry.Priority.Cycle()
		return m, m.cyclePriorityCmd(entry.ID, newPriority)

	case key.Matches(msg, m.keyMap.ToggleView):
		if m.viewMode == ViewModeDay {
			m.viewMode = ViewModeWeek
		} else {
			m.viewMode = ViewModeDay
		}
		return m, m.loadAgendaCmd()

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

	case key.Matches(msg, m.keyMap.Help):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil

	case key.Matches(msg, m.keyMap.Capture):
		m.captureMode = captureState{active: true}
		if content, exists := LoadDraft(m.draftPath); exists {
			m.captureMode.draftExists = true
			m.captureMode.draftContent = content
		}
		return m, nil
	}

	// Search mode triggers (not in keyMap to avoid conflicts)
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

	// For incremental search, start from current position (include current)
	// Check current entry first
	if strings.Contains(strings.ToLower(m.entries[m.selectedIdx].Entry.Content), query) {
		return m
	}

	start := m.selectedIdx

	if m.searchMode.forward {
		// Search forward
		for i := 1; i < len(m.entries); i++ {
			idx := (start + i) % len(m.entries)
			if strings.Contains(strings.ToLower(m.entries[idx].Entry.Content), query) {
				m.selectedIdx = idx
				return m
			}
		}
	} else {
		// Search backward
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
		// Search forward from next position
		for i := 1; i <= len(m.entries); i++ {
			idx := (start + i) % len(m.entries)
			if strings.Contains(strings.ToLower(m.entries[idx].Entry.Content), query) {
				m.selectedIdx = idx
				return m
			}
		}
	} else {
		// Search backward from prev position
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

func (m Model) handleCaptureMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.captureMode.confirmCancel {
		switch msg.String() {
		case "y", "Y":
			_ = DeleteDraft(m.draftPath)
			m.captureMode = captureState{}
			return m, nil
		case "n", "N":
			m.captureMode.confirmCancel = false
			return m, nil
		}
		return m, nil
	}

	if m.captureMode.draftExists {
		switch msg.String() {
		case "y", "Y":
			m.captureMode.content = m.captureMode.draftContent
			m.captureMode.draftExists = false
			m.captureMode.draftContent = ""
			m.captureMode.cursorPos = len(m.captureMode.content)
			m = m.captureReparse()
			return m, nil
		case "n", "N":
			_ = DeleteDraft(m.draftPath)
			m.captureMode.draftExists = false
			m.captureMode.draftContent = ""
			return m, nil
		}
		return m, nil
	}

	if m.captureMode.searchMode {
		return m.handleCaptureSearchMode(msg)
	}

	// Handle paste events first (bracketed paste)
	if msg.Paste && len(msg.Runes) > 0 {
		// Convert \r to \n for cross-platform paste compatibility
		pasteContent := strings.ReplaceAll(string(msg.Runes), "\r\n", "\n")
		pasteContent = strings.ReplaceAll(pasteContent, "\r", "\n")
		m = m.captureInsertRunes([]rune(pasteContent))
		m = m.captureEnsureCursorVisible()
		m = m.captureReparse()
		return m, nil
	}

	switch msg.Type {
	case tea.KeyCtrlX:
		content := m.captureMode.content
		_ = DeleteDraft(m.draftPath)
		m.captureMode = captureState{}
		if content == "" {
			return m, nil
		}
		return m, m.saveCaptureCmd(content)

	case tea.KeyEsc:
		if m.captureMode.showHelp {
			m.captureMode.showHelp = false
			return m, nil
		}
		if m.captureMode.content == "" {
			_ = DeleteDraft(m.draftPath)
			m.captureMode = captureState{}
			return m, nil
		}
		m.captureMode.confirmCancel = true
		return m, nil

	case tea.KeyEnter:
		m = m.captureInsertNewline()
		m = m.captureReparse()
		return m, nil

	case tea.KeyBackspace:
		m = m.captureBackspace()
		m = m.captureReparse()
		return m, nil

	case tea.KeyTab:
		m = m.captureIndentLine()
		m = m.captureReparse()
		return m, nil

	case tea.KeyShiftTab:
		m = m.captureOutdentLine()
		m = m.captureReparse()
		return m, nil

	case tea.KeyUp:
		m = m.captureMoveUp()
		m = m.captureEnsureCursorVisible()
		return m, nil

	case tea.KeyDown:
		m = m.captureMoveDown()
		m = m.captureEnsureCursorVisible()
		return m, nil

	case tea.KeyLeft:
		m = m.captureBackwardChar()
		return m, nil

	case tea.KeyRight:
		m = m.captureForwardChar()
		return m, nil

	case tea.KeySpace:
		m = m.captureInsertRunes([]rune{' '})
		m = m.captureReparse()
		return m, nil

	case tea.KeyF1:
		m.captureMode.showHelp = !m.captureMode.showHelp
		return m, nil

	case tea.KeyRunes:
		m = m.captureInsertRunes(msg.Runes)
		m = m.captureReparse()
		return m, nil

	// Emacs navigation
	case tea.KeyCtrlA:
		m = m.captureBeginningOfLine()
		return m, nil

	case tea.KeyCtrlE:
		m = m.captureEndOfLine()
		return m, nil

	case tea.KeyCtrlF:
		m = m.captureForwardChar()
		return m, nil

	case tea.KeyCtrlB:
		m = m.captureBackwardChar()
		return m, nil

	case tea.KeyCtrlK:
		m = m.captureKillToEndOfLine()
		m = m.captureReparse()
		return m, nil

	case tea.KeyCtrlU:
		m = m.captureKillToBeginningOfLine()
		m = m.captureReparse()
		return m, nil

	case tea.KeyCtrlD:
		m = m.captureDeleteChar()
		m = m.captureReparse()
		return m, nil

	case tea.KeyCtrlW:
		m = m.captureDeleteWordBackward()
		m = m.captureReparse()
		return m, nil

	case tea.KeyHome:
		m = m.captureBeginningOfLine()
		return m, nil

	case tea.KeyEnd:
		m = m.captureEndOfLine()
		return m, nil

	case tea.KeyCtrlHome:
		m.captureMode.cursorPos = 0
		m = m.captureUpdateCursorLineCol()
		m = m.captureEnsureCursorVisible()
		return m, nil

	case tea.KeyCtrlEnd:
		m.captureMode.cursorPos = len(m.captureMode.content)
		m = m.captureUpdateCursorLineCol()
		m = m.captureEnsureCursorVisible()
		return m, nil

	case tea.KeyCtrlS:
		m.captureMode.searchMode = true
		m.captureMode.searchForward = true
		m.captureMode.searchQuery = ""
		return m, nil

	case tea.KeyCtrlR:
		m.captureMode.searchMode = true
		m.captureMode.searchForward = false
		m.captureMode.searchQuery = ""
		return m, nil
	}

	return m, nil
}

func (m Model) handleCaptureSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.captureMode.searchMode = false
		m.captureMode.searchQuery = ""
		return m, nil

	case tea.KeyEnter:
		m = m.captureSearch()
		m.captureMode.searchMode = false
		return m, nil

	case tea.KeyCtrlS:
		m.captureMode.searchForward = true
		m = m.captureSearchNext()
		m = m.captureEnsureCursorVisible()
		return m, nil

	case tea.KeyCtrlR:
		m.captureMode.searchForward = false
		m = m.captureSearchNext()
		m = m.captureEnsureCursorVisible()
		return m, nil

	case tea.KeyBackspace:
		if len(m.captureMode.searchQuery) > 0 {
			m.captureMode.searchQuery = m.captureMode.searchQuery[:len(m.captureMode.searchQuery)-1]
			m = m.captureSearch()
			m = m.captureEnsureCursorVisible()
		}
		return m, nil

	case tea.KeySpace:
		m.captureMode.searchQuery += " "
		m = m.captureSearch()
		m = m.captureEnsureCursorVisible()
		return m, nil

	case tea.KeyRunes:
		m.captureMode.searchQuery += string(msg.Runes)
		m = m.captureSearch()
		m = m.captureEnsureCursorVisible()
		return m, nil
	}

	return m, nil
}

func (m Model) captureInsertRunes(runes []rune) Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos > len(content) {
		pos = len(content)
	}

	toInsert := string(runes)

	// Auto-convert ASCII entry symbols to Unicode and add space at line start
	if len(runes) == 1 && m.isAtLineStart() {
		if unicode := asciiToUnicodeSymbol(runes[0]); unicode != "" {
			toInsert = unicode + " "
		}
	}

	newContent := content[:pos] + toInsert + content[pos:]
	m.captureMode.content = newContent
	m.captureMode.cursorPos = pos + len(toInsert)

	// Recalculate cursor line and column for multi-line pastes
	newlineCount := strings.Count(toInsert, "\n")
	if newlineCount > 0 {
		m.captureMode.cursorLine += newlineCount
		// Find column position on the last line
		lastNewline := strings.LastIndex(toInsert, "\n")
		m.captureMode.cursorCol = len(toInsert) - lastNewline - 1
	} else {
		m.captureMode.cursorCol += len(toInsert)
	}
	return m
}

func asciiToUnicodeSymbol(r rune) string {
	switch r {
	case '.':
		return "•"
	case '-':
		return "–"
	case 'o':
		return "○"
	case 'x':
		return "✓"
	case '>':
		return "→"
	default:
		return ""
	}
}

func (m Model) isAtLineStart() bool {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos

	// Find start of current line
	lineStart := pos
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}

	// Check if only whitespace between line start and cursor
	for i := lineStart; i < pos; i++ {
		if content[i] != ' ' && content[i] != '\t' {
			return false
		}
	}
	return true
}

func (m Model) captureBackspace() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos <= 0 || len(content) == 0 {
		return m
	}
	if pos > len(content) {
		pos = len(content)
	}
	newContent := content[:pos-1] + content[pos:]
	m.captureMode.content = newContent
	m.captureMode.cursorPos = pos - 1
	if m.captureMode.cursorCol > 0 {
		m.captureMode.cursorCol--
	} else if m.captureMode.cursorLine > 0 {
		m.captureMode.cursorLine--
		lines := strings.Split(newContent, "\n")
		if m.captureMode.cursorLine < len(lines) {
			m.captureMode.cursorCol = len(lines[m.captureMode.cursorLine])
		}
	}
	return m
}

func (m Model) captureInsertNewline() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos > len(content) {
		pos = len(content)
	}

	// Get current line's indentation
	lines := strings.Split(content[:pos], "\n")
	currentLine := ""
	if len(lines) > 0 {
		currentLine = lines[len(lines)-1]
	}
	indent := ""
	for _, ch := range currentLine {
		if ch == ' ' {
			indent += " "
		} else {
			break
		}
	}

	newContent := content[:pos] + "\n" + indent + content[pos:]
	m.captureMode.content = newContent
	m.captureMode.cursorPos = pos + 1 + len(indent)
	m.captureMode.cursorLine++
	m.captureMode.cursorCol = len(indent)
	return m
}

func (m Model) captureEnsureCursorVisible() Model {
	// Calculate editor height (same as in view)
	editorHeight := m.height - 8
	if editorHeight < 5 {
		editorHeight = 5
	}

	// Account for scroll indicators taking up lines
	effectiveHeight := editorHeight
	if m.captureMode.scrollOffset > 0 {
		effectiveHeight-- // "more above" indicator
	}
	// Reserve space for "more below" indicator
	lines := strings.Split(m.captureMode.content, "\n")
	if m.captureMode.scrollOffset+effectiveHeight < len(lines) {
		effectiveHeight--
	}

	// Adjust scroll offset to keep cursor visible
	if m.captureMode.cursorLine < m.captureMode.scrollOffset {
		m.captureMode.scrollOffset = m.captureMode.cursorLine
	}
	if m.captureMode.cursorLine >= m.captureMode.scrollOffset+effectiveHeight {
		m.captureMode.scrollOffset = m.captureMode.cursorLine - effectiveHeight + 1
	}
	return m
}

func (m Model) captureIndentLine() Model {
	content := m.captureMode.content
	lines := strings.Split(content, "\n")
	lineIdx := m.captureMode.cursorLine
	if lineIdx >= len(lines) {
		return m
	}

	lines[lineIdx] = "  " + lines[lineIdx]
	m.captureMode.content = strings.Join(lines, "\n")
	m.captureMode.cursorPos += 2
	m.captureMode.cursorCol += 2
	return m
}

func (m Model) captureOutdentLine() Model {
	content := m.captureMode.content
	lines := strings.Split(content, "\n")
	lineIdx := m.captureMode.cursorLine
	if lineIdx >= len(lines) {
		return m
	}

	line := lines[lineIdx]
	if strings.HasPrefix(line, "  ") {
		lines[lineIdx] = line[2:]
		m.captureMode.content = strings.Join(lines, "\n")
		m.captureMode.cursorPos -= 2
		if m.captureMode.cursorCol >= 2 {
			m.captureMode.cursorCol -= 2
		} else {
			m.captureMode.cursorCol = 0
		}
	}
	return m
}

func (m Model) captureReparse() Model {
	m.captureMode.parsedEntries, m.captureMode.parseError = m.parseCapture(m.captureMode.content)
	_ = SaveDraft(m.draftPath, m.captureMode.content)
	return m
}

func (m Model) captureMoveUp() Model {
	if m.captureMode.cursorLine <= 0 {
		return m
	}

	lines := strings.Split(m.captureMode.content, "\n")

	// Move to previous line
	m.captureMode.cursorLine--

	// Adjust column if new line is shorter
	if m.captureMode.cursorLine < len(lines) {
		lineLen := len(lines[m.captureMode.cursorLine])
		if m.captureMode.cursorCol > lineLen {
			m.captureMode.cursorCol = lineLen
		}
	}

	// Recalculate absolute position
	pos := 0
	for i := 0; i < m.captureMode.cursorLine; i++ {
		pos += len(lines[i]) + 1
	}
	pos += m.captureMode.cursorCol
	m.captureMode.cursorPos = pos

	// Adjust scroll offset
	if m.captureMode.cursorLine < m.captureMode.scrollOffset {
		m.captureMode.scrollOffset = m.captureMode.cursorLine
	}

	return m
}

func (m Model) captureMoveDown() Model {
	lines := strings.Split(m.captureMode.content, "\n")

	if m.captureMode.cursorLine >= len(lines)-1 {
		return m
	}

	// Move to next line
	m.captureMode.cursorLine++

	// Adjust column if new line is shorter
	if m.captureMode.cursorLine < len(lines) {
		lineLen := len(lines[m.captureMode.cursorLine])
		if m.captureMode.cursorCol > lineLen {
			m.captureMode.cursorCol = lineLen
		}
	}

	// Recalculate absolute position
	pos := 0
	for i := 0; i < m.captureMode.cursorLine; i++ {
		pos += len(lines[i]) + 1
	}
	pos += m.captureMode.cursorCol
	m.captureMode.cursorPos = pos

	// Scroll is adjusted in view rendering based on cursor position
	return m
}

func (m Model) captureBeginningOfLine() Model {
	lines := strings.Split(m.captureMode.content, "\n")
	lineIdx := m.captureMode.cursorLine
	if lineIdx >= len(lines) {
		return m
	}

	// Calculate position at beginning of current line
	pos := 0
	for i := 0; i < lineIdx; i++ {
		pos += len(lines[i]) + 1 // +1 for newline
	}

	m.captureMode.cursorPos = pos
	m.captureMode.cursorCol = 0
	return m
}

func (m Model) captureEndOfLine() Model {
	lines := strings.Split(m.captureMode.content, "\n")
	lineIdx := m.captureMode.cursorLine
	if lineIdx >= len(lines) {
		return m
	}

	// Calculate position at end of current line
	pos := 0
	for i := 0; i < lineIdx; i++ {
		pos += len(lines[i]) + 1
	}
	pos += len(lines[lineIdx])

	m.captureMode.cursorPos = pos
	m.captureMode.cursorCol = len(lines[lineIdx])
	return m
}

func (m Model) captureUpdateCursorLineCol() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos < 0 {
		pos = 0
	}
	if pos > len(content) {
		pos = len(content)
	}

	line := 0
	col := 0
	for i := 0; i < pos; i++ {
		if content[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	m.captureMode.cursorLine = line
	m.captureMode.cursorCol = col
	return m
}

func (m Model) captureForwardChar() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos >= len(content) {
		return m
	}

	m.captureMode.cursorPos++
	if content[pos] == '\n' {
		m.captureMode.cursorLine++
		m.captureMode.cursorCol = 0
	} else {
		m.captureMode.cursorCol++
	}
	return m
}

func (m Model) captureBackwardChar() Model {
	pos := m.captureMode.cursorPos
	if pos <= 0 {
		return m
	}

	content := m.captureMode.content
	m.captureMode.cursorPos--
	if pos > 0 && content[pos-1] == '\n' {
		m.captureMode.cursorLine--
		lines := strings.Split(content, "\n")
		if m.captureMode.cursorLine < len(lines) {
			m.captureMode.cursorCol = len(lines[m.captureMode.cursorLine])
		}
	} else {
		m.captureMode.cursorCol--
	}
	return m
}

func (m Model) captureKillToEndOfLine() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos > len(content) {
		pos = len(content)
	}

	// Find end of current line
	endPos := pos
	for endPos < len(content) && content[endPos] != '\n' {
		endPos++
	}

	m.captureMode.content = content[:pos] + content[endPos:]
	return m
}

func (m Model) captureKillToBeginningOfLine() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos <= 0 {
		return m
	}

	// Find start of current line
	startPos := pos
	for startPos > 0 && content[startPos-1] != '\n' {
		startPos--
	}

	m.captureMode.content = content[:startPos] + content[pos:]
	m.captureMode.cursorPos = startPos
	m = m.captureUpdateCursorLineCol()
	return m
}

func (m Model) captureDeleteChar() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos >= len(content) {
		return m
	}

	m.captureMode.content = content[:pos] + content[pos+1:]
	return m
}

func (m Model) captureSearch() Model {
	if m.captureMode.searchQuery == "" {
		return m
	}

	return m.captureSearchFrom(m.captureMode.cursorPos)
}

func (m Model) captureSearchNext() Model {
	if m.captureMode.searchQuery == "" {
		return m
	}

	// Start from next position
	startPos := m.captureMode.cursorPos + 1
	if !m.captureMode.searchForward {
		startPos = m.captureMode.cursorPos - 1
	}
	return m.captureSearchFrom(startPos)
}

func (m Model) captureSearchFrom(pos int) Model {
	content := m.captureMode.content
	query := m.captureMode.searchQuery
	if query == "" {
		return m
	}

	foundPos := -1

	if m.captureMode.searchForward {
		// Search forward from position
		searchStart := pos
		if searchStart < 0 {
			searchStart = 0
		}
		if searchStart >= len(content) {
			searchStart = 0
		}
		idx := strings.Index(content[searchStart:], query)
		if idx >= 0 {
			foundPos = searchStart + idx
		} else if searchStart > 0 {
			// Wrap around
			idx = strings.Index(content[:searchStart], query)
			if idx >= 0 {
				foundPos = idx
			}
		}
	} else {
		// Search backward from position
		searchEnd := pos
		if searchEnd < 0 {
			searchEnd = len(content)
		}
		if searchEnd > len(content) {
			searchEnd = len(content)
		}
		idx := strings.LastIndex(content[:searchEnd], query)
		if idx >= 0 {
			foundPos = idx
		} else if searchEnd < len(content) {
			// Wrap around
			idx = strings.LastIndex(content[searchEnd:], query)
			if idx >= 0 {
				foundPos = searchEnd + idx
			}
		}
	}

	if foundPos >= 0 {
		m.captureMode.cursorPos = foundPos
		// Update cursor line and column
		lines := strings.Split(content[:foundPos], "\n")
		m.captureMode.cursorLine = len(lines) - 1
		if len(lines) > 0 {
			m.captureMode.cursorCol = len(lines[len(lines)-1])
		} else {
			m.captureMode.cursorCol = 0
		}
	}

	return m
}

func (m Model) captureDeleteWordBackward() Model {
	content := m.captureMode.content
	pos := m.captureMode.cursorPos
	if pos <= 0 {
		return m
	}
	if pos > len(content) {
		pos = len(content)
	}

	// Skip any trailing spaces
	startPos := pos
	for startPos > 0 && content[startPos-1] == ' ' {
		startPos--
	}

	// Skip word characters
	for startPos > 0 && content[startPos-1] != ' ' && content[startPos-1] != '\n' {
		startPos--
	}

	m.captureMode.content = content[:startPos] + content[pos:]
	m.captureMode.cursorPos = startPos
	m.captureMode.cursorCol -= (pos - startPos)
	if m.captureMode.cursorCol < 0 {
		m.captureMode.cursorCol = 0
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

		// Handle list items differently
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

		// Handle list items differently
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
		toDate, err := parseDateFrom(dateStr, fromDate)
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

func (m Model) toggleDoneCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.selectedIdx].Entry

	// Only tasks and done entries can be toggled
	if entry.Type != domain.EntryTypeTask && entry.Type != domain.EntryTypeDone {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()
		var err error

		if entry.Type == domain.EntryTypeDone {
			err = m.bujoService.Undo(ctx, entry.ID)
		} else {
			err = m.bujoService.MarkDone(ctx, entry.ID)
		}

		if err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entry.ID}
	}
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

func (m Model) cancelEntryCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.selectedIdx].Entry

	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.CancelEntry(ctx, entry.ID); err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entry.ID}
	}
}

func (m Model) uncancelEntryCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.selectedIdx].Entry

	return func() tea.Msg {
		ctx := context.Background()
		if err := m.bujoService.UncancelEntry(ctx, entry.ID); err != nil {
			return errMsg{err}
		}
		return entryUpdatedMsg{entry.ID}
	}
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
	return parseDateFrom(s, time.Now())
}

func parseDateFrom(s string, reference time.Time) (time.Time, error) {
	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	parsed, err := naturaldate.Parse(s, reference, naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}

func (m Model) handleHabitsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

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
			days := 7
			if m.habitState.monthView {
				days = 30
			}
			daysAgo := days - 1 - m.habitState.selectedDayIdx
			logDate := time.Now().AddDate(0, 0, -daysAgo)
			return m, m.logHabitForDateCmd(m.habitState.habits[m.habitState.selectedIdx].ID, logDate)
		}
		return m, nil

	case key.Matches(msg, m.keyMap.DayLeft):
		if m.habitState.selectedDayIdx > 0 {
			m.habitState.selectedDayIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.DayRight):
		days := 7
		if m.habitState.monthView {
			days = 30
		}
		if m.habitState.selectedDayIdx < days-1 {
			m.habitState.selectedDayIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.ToggleView):
		m.habitState.monthView = !m.habitState.monthView
		return m, nil

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

func (m Model) handleGoalsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

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
	// Handle ViewTypeListItems separately
	if m.currentView == ViewTypeListItems {
		return m.handleListItemsMode(msg)
	}

	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	// ViewTypeLists handling
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

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
		return m, tea.Quit

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
	if handled, newModel, cmd := m.handleViewSwitch(msg); handled {
		return newModel, cmd
	}

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

	case msg.String() == "r":
		if m.summaryService != nil && !m.summaryState.loading {
			m.summaryState.loading = true
			m.summaryState.error = nil
			return m, m.loadSummaryCmd()
		}
		return m, nil

	case msg.String() == "1":
		m.summaryState.horizon = "daily"
		m.summaryState.summary = nil
		return m, nil

	case msg.String() == "2":
		m.summaryState.horizon = "weekly"
		m.summaryState.summary = nil
		return m, nil

	case msg.String() == "3":
		m.summaryState.horizon = "quarterly"
		m.summaryState.summary = nil
		return m, nil

	case msg.String() == "4":
		m.summaryState.horizon = "annual"
		m.summaryState.summary = nil
		return m, nil

	case msg.String() == "h":
		m.summaryState.refDate = m.navigateSummaryPeriod(-1)
		m.summaryState.summary = nil
		return m, nil

	case msg.String() == "l":
		m.summaryState.refDate = m.navigateSummaryPeriod(1)
		m.summaryState.summary = nil
		return m, nil
	}

	return m, nil
}

func (m Model) navigateSummaryPeriod(direction int) time.Time {
	refDate := m.summaryState.refDate
	switch m.summaryState.horizon {
	case domain.SummaryHorizonDaily:
		return refDate.AddDate(0, 0, direction)
	case domain.SummaryHorizonWeekly:
		return refDate.AddDate(0, 0, direction*7)
	case domain.SummaryHorizonQuarterly:
		return refDate.AddDate(0, direction*3, 0)
	case domain.SummaryHorizonAnnual:
		return refDate.AddDate(direction, 0, 0)
	default:
		return refDate.AddDate(0, 0, direction)
	}
}

func (m Model) handleViewSwitch(msg tea.KeyMsg) (bool, Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.ViewJournal):
		m.currentView = ViewTypeJournal
		return true, m, m.loadAgendaCmd()

	case key.Matches(msg, m.keyMap.ViewHabits):
		m.currentView = ViewTypeHabits
		return true, m, m.loadHabitsCmd()

	case key.Matches(msg, m.keyMap.ViewLists):
		m.currentView = ViewTypeLists
		return true, m, m.loadListsCmd()

	case key.Matches(msg, m.keyMap.ViewSearch):
		m.currentView = ViewTypeSearch
		return true, m, nil

	case key.Matches(msg, m.keyMap.ViewStats):
		m.currentView = ViewTypeStats
		return true, m, nil

	case key.Matches(msg, m.keyMap.ViewGoals):
		m.currentView = ViewTypeGoals
		return true, m, m.loadGoalsCmd()

	case key.Matches(msg, m.keyMap.ViewSettings):
		m.currentView = ViewTypeSettings
		return true, m, nil
	}

	return false, m, nil
}
