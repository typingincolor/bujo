package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
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
