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

const (
	toolbarHeight     = 2
	helpBarHeight     = 2
	verticalPadding   = 2
	minAvailableLines = 5
)

type Config struct {
	BujoService  *service.BujoService
	HabitService *service.HabitService
	ListService  *service.ListService
	GoalService  *service.GoalService
	Theme        string
}

type Model struct {
	bujoService     *service.BujoService
	habitService    *service.HabitService
	listService     *service.ListService
	goalService     *service.GoalService
	agenda          *service.MultiDayAgenda
	journalGoals    []domain.Goal
	entries         []EntryItem
	selectedIdx     int
	scrollOffset    int
	viewMode        ViewMode
	viewDate        time.Time
	currentView     ViewType
	confirmMode     confirmState
	editMode        editState
	addMode         addState
	migrateMode     migrateState
	gotoMode        gotoState
	captureMode     captureState
	searchMode      searchState
	retypeMode      retypeState
	habitState      habitState
	addHabitMode           addHabitState
	confirmHabitDeleteMode confirmHabitDeleteState
	listState              listState
	goalState              goalState
	addGoalMode            addGoalState
	editGoalMode           editGoalState
	confirmGoalDeleteMode  confirmGoalDeleteState
	moveGoalMode           moveGoalState
	migrateToGoalMode      migrateToGoalState
	commandPalette         commandPaletteState
	commandRegistry *CommandRegistry
	help            help.Model
	keyMap          KeyMap
	width           int
	height          int
	err             error
	draftPath       string
}

type searchState struct {
	active  bool
	forward bool
	query   string
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

type retypeState struct {
	active      bool
	entryID     int64
	selectedIdx int
}

type captureState struct {
	active        bool
	content       string
	cursorPos     int
	cursorLine    int
	cursorCol     int
	scrollOffset  int
	parsedEntries []domain.Entry
	parseError    error
	confirmCancel bool
	searchMode    bool
	searchForward bool
	searchQuery   string
	draftExists   bool
	draftContent  string
	showHelp      bool
}

type habitState struct {
	habits         []service.HabitStatus
	selectedIdx    int
	selectedDayIdx int
	monthView      bool
}

type addHabitState struct {
	active bool
	input  textinput.Model
}

type confirmHabitDeleteState struct {
	active  bool
	habitID int64
}

type listState struct {
	lists           []domain.List
	items           []domain.ListItem
	summaries       map[int64]*service.ListSummary
	selectedListIdx int
	selectedItemIdx int
	currentListID   int64
}

type goalState struct {
	goals       []domain.Goal
	selectedIdx int
	viewMonth   time.Time
}

type addGoalState struct {
	active bool
	input  textinput.Model
}

type editGoalState struct {
	active bool
	goalID int64
	input  textinput.Model
}

type confirmGoalDeleteState struct {
	active bool
	goalID int64
}

type moveGoalState struct {
	active bool
	goalID int64
	input  textinput.Model
}

type migrateToGoalState struct {
	active  bool
	entryID int64
	content string
	input   textinput.Model
}

type commandPaletteState struct {
	active      bool
	query       string
	selectedIdx int
	filtered    []Command
}

type ViewMode int

const (
	ViewModeDay ViewMode = iota
	ViewModeWeek
)

type ViewType int

const (
	ViewTypeJournal ViewType = iota
	ViewTypeHabits
	ViewTypeLists
	ViewTypeListItems
	ViewTypeSearch
	ViewTypeStats
	ViewTypeGoals
	ViewTypeSettings
)

type EntryItem struct {
	Entry     domain.Entry
	DayHeader string
	IsOverdue bool
	Indent    int
}

func New(bujoSvc *service.BujoService) Model {
	return NewWithConfig(Config{BujoService: bujoSvc})
}

func NewWithConfig(cfg Config) Model {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	editInput := textinput.New()
	editInput.Placeholder = "Edit content..."

	addInput := textinput.New()
	addInput.Placeholder = "New entry..."

	migrateInput := textinput.New()
	migrateInput.Placeholder = "Enter date..."

	gotoInput := textinput.New()
	gotoInput.Placeholder = "Enter date..."

	migrateToGoalInput := textinput.New()
	migrateToGoalInput.Placeholder = "Target month (YYYY-MM)..."

	return Model{
		bujoService:     cfg.BujoService,
		habitService:    cfg.HabitService,
		listService:     cfg.ListService,
		goalService:     cfg.GoalService,
		viewMode:        ViewModeDay,
		viewDate:        today,
		currentView:     ViewTypeJournal,
		commandRegistry: DefaultCommands(),
		help:            help.New(),
		keyMap:          DefaultKeyMap(),
		draftPath:       DraftPath(),
		editMode:        editState{input: editInput},
		addMode:         addState{input: addInput},
		migrateMode:       migrateState{input: migrateInput},
		gotoMode:          gotoState{input: gotoInput},
		goalState:         goalState{viewMonth: currentMonth},
		migrateToGoalMode: migrateToGoalState{input: migrateToGoalInput},
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadAgendaCmd()
}

func (m Model) availableLines() int {
	reservedHeight := toolbarHeight + helpBarHeight + verticalPadding
	available := m.height - reservedHeight
	if available < minAvailableLines {
		return minAvailableLines
	}
	return available
}

func (m Model) linesForEntry(idx int) int {
	if idx < 0 || idx >= len(m.entries) {
		return 0
	}
	item := m.entries[idx]
	entryLine := 1
	headerLine := m.headerLineCount(item)
	blankBeforeHeader := m.blankLineBeforeHeader(idx)
	return entryLine + headerLine + blankBeforeHeader
}

func (m Model) headerLineCount(item EntryItem) int {
	if item.DayHeader != "" {
		return 1
	}
	return 0
}

func (m Model) blankLineBeforeHeader(idx int) int {
	if idx <= 0 {
		return 0
	}
	currentHasHeader := m.entries[idx].DayHeader != ""
	previousHasNoHeader := m.entries[idx-1].DayHeader == ""
	if currentHasHeader && previousHasNoHeader {
		return 1
	}
	return 0
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
	}

	return m
}

func (m Model) scrollToBottom() Model {
	if len(m.entries) == 0 {
		return m
	}

	available := m.availableLines()

	// Start from the end and work backwards to find the right scroll offset
	linesNeeded := 0
	startIdx := len(m.entries) - 1

	for i := len(m.entries) - 1; i >= 0; i-- {
		entryLines := m.linesForEntry(i)
		// Account for headers properly
		if m.entries[i].DayHeader != "" && i > 0 {
			entryLines = 3 // blank + header + entry
		} else if m.entries[i].DayHeader != "" {
			entryLines = 2 // header + entry (no blank before first)
		}

		if linesNeeded+entryLines > available-1 { // -1 for "more above" indicator
			break
		}
		linesNeeded += entryLines
		startIdx = i
	}

	m.scrollOffset = startIdx
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

func (m Model) loadJournalGoalsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.goalService == nil {
			return journalGoalsLoadedMsg{goals: nil}
		}
		ctx := context.Background()
		now := time.Now()
		currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		goals, err := m.goalService.GetGoalsForMonth(ctx, currentMonth)
		if err != nil {
			return journalGoalsLoadedMsg{goals: nil}
		}
		return journalGoalsLoadedMsg{goals: goals}
	}
}

func (m Model) logHabitForDateCmd(habitID int64, date time.Time) tea.Cmd {
	return func() tea.Msg {
		if m.habitService == nil {
			return errMsg{fmt.Errorf("habit service not available")}
		}
		ctx := context.Background()
		err := m.habitService.LogHabitByIDForDate(ctx, habitID, 1, date)
		if err != nil {
			return errMsg{err}
		}
		return habitLoggedMsg{habitID}
	}
}

func (m Model) loadHabitsCmd() tea.Cmd {
	days := 7
	if m.habitState.monthView {
		days = 30
	}
	return func() tea.Msg {
		if m.habitService == nil {
			return errMsg{fmt.Errorf("habit service not available")}
		}
		ctx := context.Background()
		status, err := m.habitService.GetTrackerStatus(ctx, time.Now(), days)
		if err != nil {
			return errMsg{err}
		}
		return habitsLoadedMsg{status.Habits}
	}
}

func (m Model) addHabitCmd(name string) tea.Cmd {
	return func() tea.Msg {
		if m.habitService == nil {
			return errMsg{fmt.Errorf("habit service not available")}
		}
		ctx := context.Background()
		err := m.habitService.LogHabit(ctx, name, 1)
		if err != nil {
			return errMsg{err}
		}
		return habitAddedMsg{name}
	}
}

func (m Model) deleteHabitCmd(habitID int64) tea.Cmd {
	return func() tea.Msg {
		if m.habitService == nil {
			return errMsg{fmt.Errorf("habit service not available")}
		}
		ctx := context.Background()
		err := m.habitService.DeleteHabitByID(ctx, habitID)
		if err != nil {
			return errMsg{err}
		}
		return habitDeletedMsg{habitID}
	}
}

func (m Model) loadListsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		lists, err := m.listService.GetAllLists(ctx)
		if err != nil {
			return errMsg{err}
		}

		summaries := make(map[int64]*service.ListSummary)
		for _, list := range lists {
			summary, err := m.listService.GetListSummary(ctx, list.ID)
			if err == nil {
				summaries[list.ID] = summary
			}
		}

		return listsLoadedMsg{lists: lists, summaries: summaries}
	}
}

func (m Model) loadListItemsCmd(listID int64) tea.Cmd {
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		items, err := m.listService.GetListItems(ctx, listID)
		if err != nil {
			return errMsg{err}
		}
		return listItemsLoadedMsg{listID, items}
	}
}

func (m Model) toggleListItemCmd(item domain.ListItem) tea.Cmd {
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		var err error
		if item.Type == domain.ListItemTypeDone {
			err = m.listService.MarkUndone(ctx, item.RowID)
		} else {
			err = m.listService.MarkDone(ctx, item.RowID)
		}
		if err != nil {
			return errMsg{err}
		}
		return listItemToggledMsg{item.RowID}
	}
}

func (m Model) addListItemCmd(content string) tea.Cmd {
	listID := m.listState.currentListID
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		_, err := m.listService.AddItem(ctx, listID, domain.EntryTypeTask, content)
		if err != nil {
			return errMsg{err}
		}
		return listItemAddedMsg{listID}
	}
}

func (m Model) deleteListItemCmd(itemID int64) tea.Cmd {
	listID := m.listState.currentListID
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		err := m.listService.RemoveItem(ctx, itemID)
		if err != nil {
			return errMsg{err}
		}
		return listItemDeletedMsg{listID}
	}
}

func (m Model) loadGoalsCmd() tea.Cmd {
	viewMonth := m.goalState.viewMonth
	return func() tea.Msg {
		if m.goalService == nil {
			return errMsg{fmt.Errorf("goal service not available")}
		}
		ctx := context.Background()
		goals, err := m.goalService.GetGoalsForMonth(ctx, viewMonth)
		if err != nil {
			return errMsg{err}
		}
		return goalsLoadedMsg{goals}
	}
}

func (m Model) addGoalCmd(content string) tea.Cmd {
	viewMonth := m.goalState.viewMonth
	return func() tea.Msg {
		if m.goalService == nil {
			return errMsg{fmt.Errorf("goal service not available")}
		}
		ctx := context.Background()
		_, err := m.goalService.CreateGoal(ctx, content, viewMonth)
		if err != nil {
			return errMsg{err}
		}
		return goalAddedMsg{}
	}
}

func (m Model) editGoalCmd(goalID int64, content string) tea.Cmd {
	return func() tea.Msg {
		if m.goalService == nil {
			return errMsg{fmt.Errorf("goal service not available")}
		}
		ctx := context.Background()
		err := m.goalService.UpdateContent(ctx, goalID, content)
		if err != nil {
			return errMsg{err}
		}
		return goalEditedMsg{goalID}
	}
}

func (m Model) deleteGoalCmd(goalID int64) tea.Cmd {
	return func() tea.Msg {
		if m.goalService == nil {
			return errMsg{fmt.Errorf("goal service not available")}
		}
		ctx := context.Background()
		err := m.goalService.DeleteGoal(ctx, goalID)
		if err != nil {
			return errMsg{err}
		}
		return goalDeletedMsg{goalID}
	}
}

func (m Model) moveGoalCmd(goalID int64, targetMonth time.Time) tea.Cmd {
	return func() tea.Msg {
		if m.goalService == nil {
			return errMsg{fmt.Errorf("goal service not available")}
		}
		ctx := context.Background()
		err := m.goalService.MoveToMonth(ctx, goalID, targetMonth)
		if err != nil {
			return errMsg{err}
		}
		return goalMovedMsg{goalID}
	}
}

func (m Model) toggleGoalCmd(goalID int64, isDone bool) tea.Cmd {
	return func() tea.Msg {
		if m.goalService == nil {
			return errMsg{fmt.Errorf("goal service not available")}
		}
		ctx := context.Background()
		var err error
		if isDone {
			err = m.goalService.MarkActive(ctx, goalID)
		} else {
			err = m.goalService.MarkDone(ctx, goalID)
		}
		if err != nil {
			return errMsg{err}
		}
		return goalToggledMsg{goalID}
	}
}

func (m Model) migrateToGoalCmd(entryID int64, content string, targetMonth time.Time) tea.Cmd {
	return func() tea.Msg {
		if m.goalService == nil {
			return errMsg{fmt.Errorf("goal service not available")}
		}
		if m.bujoService == nil {
			return errMsg{fmt.Errorf("bujo service not available")}
		}
		ctx := context.Background()

		goalID, err := m.goalService.CreateGoal(ctx, content, targetMonth)
		if err != nil {
			return errMsg{err}
		}

		err = m.bujoService.DeleteEntry(ctx, entryID)
		if err != nil {
			return errMsg{err}
		}

		return entryMigratedToGoalMsg{entryID: entryID, goalID: goalID}
	}
}

func (m Model) flattenAgenda(agenda *service.MultiDayAgenda) []EntryItem {
	if agenda == nil {
		return nil
	}

	// Calculate today for per-entry overdue checks
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var items []EntryItem

	if len(agenda.Overdue) > 0 {
		items = append(items, m.flattenEntries(agenda.Overdue, "⚠️  OVERDUE", true, today)...)
	}

	for _, day := range agenda.Days {
		if len(day.Entries) == 0 {
			continue
		}

		dayHeader := fmt.Sprintf("📅 %s", day.Date.Format("Monday, Jan 2"))
		if day.Location != nil && *day.Location != "" {
			dayHeader = fmt.Sprintf("%s | 📍 %s", dayHeader, *day.Location)
		}

		items = append(items, m.flattenEntries(day.Entries, dayHeader, false, today)...)
	}

	return items
}

func (m Model) parseCapture(content string) ([]domain.Entry, error) {
	if m.bujoService == nil {
		return nil, fmt.Errorf("bujo service not configured")
	}
	return m.bujoService.ParseEntries(content)
}

func (m Model) flattenEntries(entries []domain.Entry, header string, forceOverdue bool, today time.Time) []EntryItem {
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
		// Check if entry is overdue: either forced (in OVERDUE section) or per-entry check
		entryIsOverdue := forceOverdue || entry.IsOverdue(today)
		item := EntryItem{
			Entry:     entry,
			IsOverdue: entryIsOverdue,
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
