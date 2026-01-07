package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

type Model struct {
	bujoService  *service.BujoService
	agenda       *service.MultiDayAgenda
	entries      []EntryItem
	selectedIdx  int
	scrollOffset int
	viewMode     ViewMode
	viewDate     time.Time
	confirmMode  confirmState
	editMode     editState
	addMode      addState
	migrateMode  migrateState
	gotoMode     gotoState
	help         help.Model
	keyMap       KeyMap
	width        int
	height       int
	err          error
}

type confirmState struct {
	active      bool
	entryID     int64
	hasChildren bool
}

type editState struct {
	active  bool
	entryID int64
	input   textinput.Model
}

type addState struct {
	active   bool
	asChild  bool
	parentID *int64
	input    textinput.Model
}

type migrateState struct {
	active   bool
	entryID  int64
	fromDate time.Time
	input    textinput.Model
}

type gotoState struct {
	active bool
	input  textinput.Model
}

type ViewMode int

const (
	ViewModeDay ViewMode = iota
	ViewModeWeek
)

type EntryItem struct {
	Entry     domain.Entry
	DayHeader string
	IsOverdue bool
	Indent    int
}

func New(bujoSvc *service.BujoService) Model {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return Model{
		bujoService: bujoSvc,
		viewMode:    ViewModeDay,
		viewDate:    today,
		help:        help.New(),
		keyMap:      DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadAgendaCmd()
}

func (m Model) availableLines() int {
	// Reserve: 2 for toolbar, 2 for help, 2 for padding
	h := m.height - 6
	if h < 5 {
		h = 5
	}
	return h
}

func (m Model) linesForEntry(idx int) int {
	if idx < 0 || idx >= len(m.entries) {
		return 0
	}
	item := m.entries[idx]
	lines := 1 // entry itself
	if item.DayHeader != "" {
		lines++ // header line
		if idx > 0 && m.entries[idx-1].DayHeader == "" {
			lines++ // blank line before header (unless first visible)
		}
	}
	return lines
}

func (m Model) ensuredVisible() Model {
	if len(m.entries) == 0 {
		return m
	}

	available := m.availableLines()

	// If selected is above visible area, scroll up
	if m.selectedIdx < m.scrollOffset {
		m.scrollOffset = m.selectedIdx
		return m
	}

	// Calculate lines used from scrollOffset to selectedIdx (inclusive)
	linesUsed := 0
	for i := m.scrollOffset; i <= m.selectedIdx; i++ {
		entryLines := m.linesForEntry(i)
		// First visible entry doesn't get blank line before header
		if i == m.scrollOffset && m.entries[i].DayHeader != "" {
			entryLines = 2 // just header + entry, no blank line
		}
		linesUsed += entryLines
	}

	// Account for scroll indicators
	if m.scrollOffset > 0 {
		linesUsed++ // "more above" indicator
	}
	if m.selectedIdx < len(m.entries)-1 {
		linesUsed++ // reserve for "more below" indicator
	}

	// If selected is below visible area, scroll down
	for linesUsed > available && m.scrollOffset < m.selectedIdx {
		// Remove lines for the entry we're scrolling past
		entryLines := m.linesForEntry(m.scrollOffset)
		if m.scrollOffset == 0 && m.entries[0].DayHeader != "" {
			entryLines = 2
		}
		linesUsed -= entryLines
		m.scrollOffset++

		// After scrolling, first visible might change its line count
		if m.scrollOffset < len(m.entries) && m.entries[m.scrollOffset].DayHeader != "" {
			// Recalculate: now this is first visible, so no blank line before
			// But we already counted it with possible blank line, adjust
		}
	}

	return m
}

func (m Model) loadAgendaCmd() tea.Cmd {
	viewMode := m.viewMode
	viewDate := m.viewDate
	return func() tea.Msg {
		ctx := context.Background()
		var from, to time.Time

		switch viewMode {
		case ViewModeDay:
			from = viewDate
			to = viewDate
		case ViewModeWeek:
			from = viewDate.AddDate(0, 0, -6)
			to = viewDate
		}

		agenda, err := m.bujoService.GetMultiDayAgenda(ctx, from, to)
		if err != nil {
			return errMsg{err}
		}
		return agendaLoadedMsg{agenda}
	}
}

func (m Model) flattenAgenda(agenda *service.MultiDayAgenda) []EntryItem {
	if agenda == nil {
		return nil
	}

	var items []EntryItem

	if len(agenda.Overdue) > 0 {
		items = append(items, m.flattenEntries(agenda.Overdue, "⚠️  OVERDUE", true)...)
	}

	for _, day := range agenda.Days {
		if len(day.Entries) == 0 {
			continue
		}

		dayHeader := fmt.Sprintf("📅 %s", day.Date.Format("Monday, Jan 2"))
		if day.Location != nil && *day.Location != "" {
			dayHeader = fmt.Sprintf("%s | 📍 %s", dayHeader, *day.Location)
		}

		items = append(items, m.flattenEntries(day.Entries, dayHeader, false)...)
	}

	return items
}

func (m Model) flattenEntries(entries []domain.Entry, header string, isOverdue bool) []EntryItem {
	var items []EntryItem

	parentMap := make(map[int64][]domain.Entry)
	var roots []domain.Entry

	for _, e := range entries {
		if e.ParentID == nil {
			roots = append(roots, e)
		} else {
			parentMap[*e.ParentID] = append(parentMap[*e.ParentID], e)
		}
	}

	var flatten func(entry domain.Entry, depth int, showHeader bool)
	flatten = func(entry domain.Entry, depth int, showHeader bool) {
		item := EntryItem{
			Entry:     entry,
			IsOverdue: isOverdue,
			Indent:    depth,
		}
		if showHeader {
			item.DayHeader = header
		}
		items = append(items, item)

		for _, child := range parentMap[entry.ID] {
			flatten(child, depth+1, false)
		}
	}

	for i, root := range roots {
		flatten(root, 0, i == 0)
	}

	return items
}
