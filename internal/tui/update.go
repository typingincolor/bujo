package tui

import (
	"context"
	"fmt"
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
		if m.selectedIdx >= len(m.entries) {
			m.selectedIdx = max(0, len(m.entries)-1)
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case entryUpdatedMsg, entryDeletedMsg:
		return m, m.loadAgendaCmd()

	case confirmDeleteMsg:
		m.confirmMode = confirmState{
			active:      true,
			entryID:     msg.entryID,
			hasChildren: msg.hasChildren,
		}
		return m, nil

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
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
		return m.handleNormalMode(msg)
	}

	return m, nil
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keyMap.Up):
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Down):
		if m.selectedIdx < len(m.entries)-1 {
			m.selectedIdx++
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Top):
		m.selectedIdx = 0
		return m, nil

	case key.Matches(msg, m.keyMap.Bottom):
		if len(m.entries) > 0 {
			m.selectedIdx = len(m.entries) - 1
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Done):
		return m, m.toggleDoneCmd()

	case key.Matches(msg, m.keyMap.Delete):
		return m, m.initiateDeleteCmd()

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
		m.migrateMode = migrateState{
			active:  true,
			entryID: entry.ID,
			input:   ti,
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Help):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil
	}

	return m, nil
}

func (m Model) handleConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Confirm):
		entryID := m.confirmMode.entryID
		m.confirmMode.active = false
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
		parentID := m.addMode.parentID
		m.addMode.active = false
		return m, m.addEntryCmd(content, parentID)
	}

	var cmd tea.Cmd
	m.addMode.input, cmd = m.addMode.input.Update(msg)
	return m, cmd
}

func (m Model) addEntryCmd(content string, parentID *int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		today := m.getTodayDate()
		opts := service.LogEntriesOptions{Date: today}
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
		m.migrateMode.active = false
		return m, m.migrateEntryCmd(entryID, dateStr)
	}

	var cmd tea.Cmd
	m.migrateMode.input, cmd = m.migrateMode.input.Update(msg)
	return m, cmd
}

func (m Model) migrateEntryCmd(id int64, dateStr string) tea.Cmd {
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

func (m Model) toggleDoneCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.selectedIdx].Entry

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

func (m Model) initiateDeleteCmd() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[m.selectedIdx].Entry

	return func() tea.Msg {
		ctx := context.Background()
		hasChildren, err := m.bujoService.HasChildren(ctx, entry.ID)
		if err != nil {
			return errMsg{err}
		}

		if hasChildren {
			return confirmDeleteMsg{entryID: entry.ID, hasChildren: true}
		}

		if err := m.bujoService.DeleteEntry(ctx, entry.ID); err != nil {
			return errMsg{err}
		}
		return entryDeletedMsg{entry.ID}
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
	now := time.Now()

	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	parsed, err := naturaldate.Parse(s, now, naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}
