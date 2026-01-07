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
	bujoService *service.BujoService
	agenda      *service.MultiDayAgenda
	entries     []EntryItem
	selectedIdx int
	confirmMode confirmState
	editMode    editState
	addMode     addState
	migrateMode migrateState
	help        help.Model
	keyMap      KeyMap
	width       int
	height      int
	err         error
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
	active  bool
	entryID int64
	input   textinput.Model
}

type EntryItem struct {
	Entry     domain.Entry
	DayHeader string
	IsOverdue bool
	Indent    int
}

func New(bujoSvc *service.BujoService) Model {
	return Model{
		bujoService: bujoSvc,
		help:        help.New(),
		keyMap:      DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadAgendaCmd()
}

func (m Model) loadAgendaCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		today := time.Now()
		from := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()).AddDate(0, 0, -6)
		to := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

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
		items = append(items, m.flattenEntries(agenda.Overdue, "OVERDUE", true)...)
	}

	for _, day := range agenda.Days {
		if len(day.Entries) == 0 {
			continue
		}

		dayHeader := day.Date.Format("Monday, Jan 2")
		if day.Location != nil && *day.Location != "" {
			dayHeader = fmt.Sprintf("%s | %s", dayHeader, *day.Location)
		}

		items = append(items, m.flattenEntries(day.Entries, dayHeader, false)...)
	}

	return items
}

func (m Model) getTodayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
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
