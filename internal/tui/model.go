package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

const (
	toolbarHeight             = 2
	helpBarHeight             = 2
	verticalPadding           = 2
	minAvailableLines         = 5
	locationHistoryMonths     = 6
	pendingTasksLookbackYears = 1
	pendingTasksHeaderLines   = 4
	pendingTasksFooterLines   = 2
)

type ChangeDetector interface {
	GetLastModified(ctx context.Context) (time.Time, error)
}

type InsightsReader interface {
	IsAvailable() bool
	GetLatestSummary(ctx context.Context) (*domain.InsightsSummary, error)
	GetActiveInitiatives(ctx context.Context, limit int) ([]domain.InsightsInitiative, error)
	GetPendingActions(ctx context.Context) ([]domain.InsightsAction, error)
	GetRecentDecisions(ctx context.Context, limit int) ([]domain.InsightsDecision, error)
	GetDaysSinceLastSummary(ctx context.Context) (int, error)
	GetSummaryForWeek(ctx context.Context, weekStart, nextWeekStart string) (*domain.InsightsSummary, error)
	GetTopicsForSummary(ctx context.Context, summaryID int64) ([]domain.InsightsTopic, error)
	GetActionsForWeek(ctx context.Context, weekStart, nextWeekStart string) ([]domain.InsightsAction, error)
	GetInitiativePortfolio(ctx context.Context) ([]domain.InsightsInitiativePortfolio, error)
	GetInitiativeDetail(ctx context.Context, initiativeID int64) (*domain.InsightsInitiativeDetail, error)
	GetDistinctTopics(ctx context.Context) ([]string, error)
	GetTopicTimeline(ctx context.Context, topic string) ([]domain.InsightsTopicTimeline, error)
	GetDecisionsWithInitiatives(ctx context.Context) ([]domain.InsightsDecisionWithInitiatives, error)
	GetWeeklyReport(ctx context.Context, weekStart, nextWeekStart string) (*domain.InsightsWeeklyReport, error)
}

type Config struct {
	BujoService     *service.BujoService
	HabitService    *service.HabitService
	ListService     *service.ListService
	GoalService     *service.GoalService
	StatsService    *service.StatsService
	InsightsReader  InsightsReader
	ChangeDetection ChangeDetector
	Theme           string
	Version         string
	Commit          string
	Date            string
	DBPath          string
}

type Model struct {
	bujoService              *service.BujoService
	habitService             *service.HabitService
	listService              *service.ListService
	goalService              *service.GoalService
	statsService             *service.StatsService
	changeDetection          ChangeDetector
	lastCheckedModified      time.Time
	days                     []service.DayEntries
	journalGoals             []domain.Goal
	entries                  []EntryItem
	collapsed                map[domain.EntityID]bool
	selectedIdx              int
	scrollOffset             int
	viewMode                 ViewMode
	viewDate                 time.Time
	currentView              ViewType
	viewStack                []ViewType
	confirmMode              confirmState
	quitConfirmMode          confirmQuitState
	editMode                 editState
	answerMode               answerState
	addMode                  addState
	migrateMode              migrateState
	gotoMode                 gotoState
	searchMode               searchState
	searchView               searchViewState
	retypeMode               retypeState
	habitState               habitState
	addHabitMode             addHabitState
	confirmHabitDeleteMode   confirmHabitDeleteState
	listState                listState
	moveListItemMode         moveListItemState
	createListMode           createListState
	moveToListMode           moveToListState
	goalState                goalState
	addGoalMode              addGoalState
	editGoalMode             editGoalState
	confirmGoalDeleteMode    confirmGoalDeleteState
	moveGoalMode             moveGoalState
	migrateToGoalMode        migrateToGoalState
	expandedOverdueContextID *int64
	insightsReader           InsightsReader
	statsViewState           statsState
	insightsState            insightsState
	presetPicker             presetPickerState
	commandPalette           commandPaletteState
	commandRegistry          *CommandRegistry
	pendingTasksState        pendingTasksState
	questionsState           questionsState
	undoState                undoState
	appVersion               string
	appCommit                string
	appDate                  string
	appDBPath                string
	help                     help.Model
	keyMap                   KeyMap
	width                    int
	height                   int
	err                      error
	draftPath                string
}

type searchState struct {
	active  bool
	forward bool
	query   string
}

type searchViewState struct {
	query       string
	results     []domain.Entry
	selectedIdx int
	loading     bool
	input       textinput.Model
}

type confirmState struct {
	active      bool
	entryID     int64
	hasChildren bool
}

type confirmQuitState struct {
	active bool
}

type editState struct {
	active  bool
	entryID int64
	input   textinput.Model
}

type answerState struct {
	active     bool
	questionID int64
	input      textinput.Model
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

type presetPickerKind int

const (
	pickerLocation presetPickerKind = iota
	pickerMood
	pickerWeather
)

type presetPickerState struct {
	active       bool
	pickerMode   bool
	kind         presetPickerKind
	date         time.Time
	input        textinput.Model
	items        []string
	defaultItems []string
	selectedIdx  int
	title        string
	pickerLabel  string
}

var moodPresets = []string{"Happy", "Neutral", "Sad", "Frustrated", "Tired", "Sick", "Anxious", "Grateful"}

var weatherPresets = []string{"Sunny", "Partly-Cloudy", "Cloudy", "Rainy", "Stormy", "Snowy"}

type retypeState struct {
	active      bool
	entryID     int64
	selectedIdx int
}

type habitState struct {
	habits         []service.HabitStatus
	selectedIdx    int
	selectedDayIdx int
	dayIdxInited   bool
	viewMode       HabitViewMode
	weekOffset     int
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

type moveListItemState struct {
	active      bool
	itemID      int64
	targetLists []domain.List
	selectedIdx int
}

type createListState struct {
	active bool
	input  textinput.Model
}

type moveToListState struct {
	active      bool
	entryID     int64
	targetLists []domain.List
	selectedIdx int
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

type statsState struct {
	stats   *domain.Stats
	loading bool
	from    time.Time
	to      time.Time
}

type InsightsTab int

const (
	InsightsTabDashboard InsightsTab = iota
	InsightsTabSummaries
	InsightsTabActions
	InsightsTabInitiatives
	InsightsTabTopics
	InsightsTabDecisions
	insightsTabCount
)

type insightsState struct {
	loading               bool
	activeTab             InsightsTab
	weekAnchor            time.Time
	dashboard             domain.InsightsDashboard
	weekSummary           *domain.InsightsSummary
	weekTopics            []domain.InsightsTopic
	weekActions           []domain.InsightsAction
	initiatives           []domain.InsightsInitiativePortfolio
	initiativeSelectedIdx int
	selectedInitiativeID  *int64
	initiativeDetail      *domain.InsightsInitiativeDetail
	distinctTopics        []string
	topicSelectedIdx      int
	selectedTopic         string
	topicTimeline         []domain.InsightsTopicTimeline
	decisions             []domain.InsightsDecisionWithInitiatives
	weeklyReport          *domain.InsightsWeeklyReport
}

type commandPaletteState struct {
	active      bool
	query       string
	selectedIdx int
	filtered    []Command
}

type pendingTasksState struct {
	entries      []domain.Entry
	selectedIdx  int
	scrollOffset int
	loading      bool
	parentChains map[int64][]domain.Entry
	expandedID   int64
}

type questionsState struct {
	entries     []domain.Entry
	selectedIdx int
	loading     bool
}

type ViewMode int

const (
	ViewModeDay ViewMode = iota
	ViewModeWeek
)

type ViewType int

const (
	ViewTypeJournal      ViewType = iota // key 1
	ViewTypeReview                       // key 2
	ViewTypePendingTasks                 // key 3
	ViewTypeQuestions                    // key 4
	ViewTypeHabits                       // key 5
	ViewTypeLists                        // key 6
	ViewTypeGoals                        // key 7
	ViewTypeSearch                       // key 8
	ViewTypeStats                        // key 9
	ViewTypeSettings                     // key 0
	ViewTypeInsights                     // key i
	ViewTypeListItems                    // internal (accessed via Lists)
)

type HabitViewMode int

const (
	HabitViewModeWeek HabitViewMode = iota
	HabitViewModeMonth
	HabitViewModeQuarter
)

const (
	HabitDaysWeek    = 7
	HabitDaysMonth   = 30
	HabitDaysQuarter = 90
)

type UndoOperation int

const (
	UndoOpNone UndoOperation = iota
	UndoOpMarkDone
	UndoOpMarkUndone
)

type undoState struct {
	operation UndoOperation
	entryID   int64
	entityID  domain.EntityID
	oldEntry  *domain.Entry
}

type EntryItem struct {
	Entry            domain.Entry
	DayHeader        string
	IsOverdue        bool
	Indent           int
	HasChildren      bool
	HiddenChildCount int
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

	searchInput := textinput.New()
	searchInput.Placeholder = "Search entries..."
	searchInput.Focus()

	statsFrom := now.AddDate(0, 0, -29)
	statsTo := now

	weekday := today.Weekday()
	daysFromMonday := (int(weekday) - int(time.Monday) + 7) % 7
	currentMonday := today.AddDate(0, 0, -daysFromMonday)

	return Model{
		bujoService:       cfg.BujoService,
		habitService:      cfg.HabitService,
		listService:       cfg.ListService,
		goalService:       cfg.GoalService,
		statsService:      cfg.StatsService,
		changeDetection:   cfg.ChangeDetection,
		collapsed:         make(map[domain.EntityID]bool),
		viewMode:          ViewModeDay,
		viewDate:          today,
		currentView:       ViewTypeJournal,
		commandRegistry:   DefaultCommands(),
		help:              help.New(),
		keyMap:            DefaultKeyMap(),
		draftPath:         DraftPath(),
		editMode:          editState{input: editInput},
		addMode:           addState{input: addInput},
		migrateMode:       migrateState{input: migrateInput},
		gotoMode:          gotoState{input: gotoInput},
		goalState:         goalState{viewMonth: currentMonth},
		migrateToGoalMode: migrateToGoalState{input: migrateToGoalInput},
		searchView:        searchViewState{input: searchInput},
		statsViewState:    statsState{from: statsFrom, to: statsTo},
		insightsReader:    cfg.InsightsReader,
		insightsState:     insightsState{weekAnchor: currentMonday},
		appVersion:        cfg.Version,
		appCommit:         cfg.Commit,
		appDate:           cfg.Date,
		appDBPath:         cfg.DBPath,
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadDaysCmd()}

	if m.changeDetection != nil {
		cmds = append(cmds, m.checkChangesCmd())
	}

	return tea.Batch(cmds...)
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

	if m.selectedIdx < m.scrollOffset {
		m.scrollOffset = m.selectedIdx
		return m
	}

	linesUsed := 0
	for i := m.scrollOffset; i <= m.selectedIdx; i++ {
		entryLines := m.linesForEntry(i)
		if i == m.scrollOffset && m.entries[i].DayHeader != "" {
			entryLines = 2 // just header + entry, no blank line
		}
		linesUsed += entryLines
	}

	if m.scrollOffset > 0 {
		linesUsed++ // "more above" indicator
	}
	if m.selectedIdx < len(m.entries)-1 {
		linesUsed++ // reserve for "more below" indicator
	}

	for linesUsed > available && m.scrollOffset < m.selectedIdx {
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

	linesNeeded := 0
	startIdx := len(m.entries) - 1

	for i := len(m.entries) - 1; i >= 0; i-- {
		entryLines := m.linesForEntry(i)
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

func (m Model) loadDaysCmd() tea.Cmd {
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

		days, err := m.bujoService.GetDayEntries(ctx, from, to)
		if err != nil {
			return errMsg{err}
		}
		return daysLoadedMsg{days}
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

func (m Model) removeHabitLogForDateCmd(habitID int64, date time.Time) tea.Cmd {
	return func() tea.Msg {
		if m.habitService == nil {
			return errMsg{fmt.Errorf("habit service not available")}
		}
		ctx := context.Background()
		err := m.habitService.RemoveHabitLogForDateByID(ctx, habitID, date)
		if err != nil {
			if err.Error() == "no logs to remove for this date" {
				return habitLogRemovedMsg{habitID}
			}
			return errMsg{err}
		}
		return habitLogRemovedMsg{habitID}
	}
}

func (m Model) getHabitReferenceDate() time.Time {
	days := HabitDaysWeek
	switch m.habitState.viewMode {
	case HabitViewModeMonth:
		days = HabitDaysMonth
	case HabitViewModeQuarter:
		days = HabitDaysQuarter
	}
	return time.Now().AddDate(0, 0, -m.habitState.weekOffset*days)
}

func (m Model) loadHabitsCmd() tea.Cmd {
	days := HabitDaysWeek
	switch m.habitState.viewMode {
	case HabitViewModeMonth:
		days = HabitDaysMonth
	case HabitViewModeQuarter:
		days = HabitDaysQuarter
	}
	referenceDate := m.getHabitReferenceDate()
	return func() tea.Msg {
		if m.habitService == nil {
			return errMsg{fmt.Errorf("habit service not available")}
		}
		ctx := context.Background()
		status, err := m.habitService.GetTrackerStatus(ctx, referenceDate, days)
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

func (m Model) createListCmd(name string) tea.Cmd {
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		_, err := m.listService.CreateList(ctx, name)
		if err != nil {
			return errMsg{err}
		}
		return listCreatedMsg{}
	}
}

func (m Model) loadListsForMoveCmd(entryID int64) tea.Cmd {
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		lists, err := m.listService.GetAllLists(ctx)
		if err != nil {
			return errMsg{err}
		}
		return listsForMoveLoadedMsg{
			entryID: entryID,
			lists:   lists,
		}
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

func (m Model) editListItemCmd(itemID int64, content string) tea.Cmd {
	listID := m.listState.currentListID
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		err := m.listService.EditItem(ctx, itemID, content)
		if err != nil {
			return errMsg{err}
		}
		return listItemEditedMsg{listID}
	}
}

func (m Model) moveListItemCmd(itemID int64, targetListID int64, fromListID int64) tea.Cmd {
	return func() tea.Msg {
		if m.listService == nil {
			return errMsg{fmt.Errorf("list service not available")}
		}
		ctx := context.Background()
		err := m.listService.MoveItem(ctx, itemID, targetListID)
		if err != nil {
			return errMsg{err}
		}
		return listItemMovedMsg{fromListID: fromListID, toListID: targetListID}
	}
}

func (m Model) moveEntryToListCmd(entryID int64, listID int64) tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return errMsg{fmt.Errorf("bujo service not available")}
		}
		ctx := context.Background()

		err := m.bujoService.MoveEntryToList(ctx, entryID, listID)
		if err != nil {
			return errMsg{err}
		}

		return entryMovedToListMsg{entryID: entryID}
	}
}

func (m Model) moveToRootCmd(entryID int64) tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return errMsg{fmt.Errorf("bujo service not available")}
		}
		ctx := context.Background()

		moveToRoot := true
		err := m.bujoService.MoveEntry(ctx, entryID, service.MoveOptions{MoveToRoot: &moveToRoot})
		if err != nil {
			return errMsg{err}
		}

		return entryUpdatedMsg{id: entryID}
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
		err := m.goalService.UpdateGoal(ctx, goalID, content)
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

func (m Model) searchEntriesCmd(query string) tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return errMsg{fmt.Errorf("bujo service not available")}
		}
		if query == "" {
			return searchResultsMsg{results: nil, query: query}
		}
		ctx := context.Background()
		opts := domain.NewSearchOptions(query)
		results, err := m.bujoService.SearchEntries(ctx, opts)
		if err != nil {
			return errMsg{err}
		}
		return searchResultsMsg{results: results, query: query}
	}
}

func (m Model) loadStatsCmd() tea.Cmd {
	from := m.statsViewState.from
	to := m.statsViewState.to
	return func() tea.Msg {
		if m.statsService == nil {
			return errMsg{fmt.Errorf("stats service not available")}
		}
		ctx := context.Background()
		stats, err := m.statsService.GetStats(ctx, from, to)
		if err != nil {
			return errMsg{err}
		}
		return statsLoadedMsg{stats}
	}
}

func (m Model) loadPendingTasksCmd() tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return errMsg{fmt.Errorf("bujo service not available")}
		}
		ctx := context.Background()
		now := time.Now()
		from := now.AddDate(-pendingTasksLookbackYears, 0, 0)
		tasks, err := m.bujoService.GetOutstandingTasks(ctx, from, now)
		if err != nil {
			return errMsg{err}
		}
		return pendingTasksLoadedMsg{entries: tasks}
	}
}

func (m Model) loadQuestionsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return errMsg{fmt.Errorf("bujo service not available")}
		}
		ctx := context.Background()
		opts := domain.NewSearchOptions("").WithType(domain.EntryTypeQuestion)
		questions, err := m.bujoService.SearchEntries(ctx, opts)
		if err != nil {
			return errMsg{err}
		}
		return questionsLoadedMsg{entries: questions}
	}
}

func (m Model) loadInsightsDashboardCmd() tea.Cmd {
	return func() tea.Msg {
		if m.insightsReader == nil || !m.insightsReader.IsAvailable() {
			return insightsDashboardLoadedMsg{dashboard: domain.InsightsDashboard{Status: "unavailable"}}
		}
		ctx := context.Background()
		var errs []error
		summary, err := m.insightsReader.GetLatestSummary(ctx)
		if err != nil {
			errs = append(errs, err)
		}
		initiatives, err := m.insightsReader.GetActiveInitiatives(ctx, 5)
		if err != nil {
			errs = append(errs, err)
		}
		actions, err := m.insightsReader.GetPendingActions(ctx)
		if err != nil {
			errs = append(errs, err)
		}
		decisions, err := m.insightsReader.GetRecentDecisions(ctx, 5)
		if err != nil {
			errs = append(errs, err)
		}
		daysSince, err := m.insightsReader.GetDaysSinceLastSummary(ctx)
		if err != nil {
			errs = append(errs, err)
		}

		var highPriority []domain.InsightsAction
		for _, a := range actions {
			if a.Priority == "high" {
				highPriority = append(highPriority, a)
			}
		}

		status := "ok"
		if len(errs) > 0 {
			status = "error"
		}

		return insightsDashboardLoadedMsg{
			dashboard: domain.InsightsDashboard{
				LatestSummary:        summary,
				ActiveInitiatives:    initiatives,
				HighPriorityActions:  highPriority,
				RecentDecisions:      decisions,
				DaysSinceLastSummary: daysSince,
				Status:               status,
			},
		}
	}
}

func (m Model) loadInsightsTabDataCmd() tea.Cmd {
	tab := m.insightsState.activeTab
	weekStart := m.insightsState.weekAnchor.Format("2006-01-02")
	nextWeek := m.insightsState.weekAnchor.AddDate(0, 0, 7).Format("2006-01-02")

	return func() tea.Msg {
		if m.insightsReader == nil || !m.insightsReader.IsAvailable() {
			switch tab {
			case InsightsTabSummaries:
				return insightsSummaryLoadedMsg{}
			case InsightsTabActions:
				return insightsActionsLoadedMsg{}
			case InsightsTabInitiatives:
				return insightsInitiativesLoadedMsg{}
			case InsightsTabTopics:
				return insightsTopicsListLoadedMsg{}
			case InsightsTabDecisions:
				return insightsDecisionsLoadedMsg{}
			default:
				return insightsDashboardLoadedMsg{dashboard: domain.InsightsDashboard{Status: "unavailable"}}
			}
		}
		ctx := context.Background()

		switch tab {
		case InsightsTabSummaries:
			summary, err := m.insightsReader.GetSummaryForWeek(ctx, weekStart, nextWeek)
			if err != nil {
				return insightsSummaryLoadedMsg{err: err}
			}
			var topics []domain.InsightsTopic
			if summary != nil {
				topics, err = m.insightsReader.GetTopicsForSummary(ctx, summary.ID)
				if err != nil {
					return insightsSummaryLoadedMsg{summary: summary, err: err}
				}
			}
			return insightsSummaryLoadedMsg{summary: summary, topics: topics}

		case InsightsTabActions:
			actions, err := m.insightsReader.GetActionsForWeek(ctx, weekStart, nextWeek)
			return insightsActionsLoadedMsg{actions: actions, err: err}

		case InsightsTabInitiatives:
			initiatives, err := m.insightsReader.GetInitiativePortfolio(ctx)
			return insightsInitiativesLoadedMsg{initiatives: initiatives, err: err}

		case InsightsTabTopics:
			topics, err := m.insightsReader.GetDistinctTopics(ctx)
			return insightsTopicsListLoadedMsg{topics: topics, err: err}

		case InsightsTabDecisions:
			decisions, err := m.insightsReader.GetDecisionsWithInitiatives(ctx)
			return insightsDecisionsLoadedMsg{decisions: decisions, err: err}

		default:
			return nil
		}
	}
}

func (m Model) loadInsightsInitiativeDetailCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		if m.insightsReader == nil || !m.insightsReader.IsAvailable() {
			return insightsInitiativeDetailLoadedMsg{}
		}
		ctx := context.Background()
		detail, err := m.insightsReader.GetInitiativeDetail(ctx, id)
		return insightsInitiativeDetailLoadedMsg{detail: detail, err: err}
	}
}

func (m Model) loadInsightsTopicTimelineCmd(topic string) tea.Cmd {
	return func() tea.Msg {
		if m.insightsReader == nil || !m.insightsReader.IsAvailable() {
			return insightsTopicTimelineLoadedMsg{}
		}
		ctx := context.Background()
		timeline, err := m.insightsReader.GetTopicTimeline(ctx, topic)
		return insightsTopicTimelineLoadedMsg{timeline: timeline, err: err}
	}
}

func (m Model) loadInsightsWeeklyReportCmd() tea.Cmd {
	return func() tea.Msg {
		if m.insightsReader == nil || !m.insightsReader.IsAvailable() {
			return insightsWeeklyReportLoadedMsg{}
		}
		ctx := context.Background()
		weekStart := m.insightsState.weekAnchor.Format("2006-01-02")
		nextWeekStart := m.insightsState.weekAnchor.AddDate(0, 0, 7).Format("2006-01-02")
		report, err := m.insightsReader.GetWeeklyReport(ctx, weekStart, nextWeekStart)
		return insightsWeeklyReportLoadedMsg{report: report, err: err}
	}
}

func (m Model) flattenDays(days []service.DayEntries) []EntryItem {
	if days == nil {
		return nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var items []EntryItem

	for _, day := range days {
		if len(day.Entries) == 0 {
			continue
		}

		dayHeader := fmt.Sprintf("📅 %s", day.Date.Format("Monday, Jan 2"))
		if day.Location != nil && *day.Location != "" {
			dayHeader = fmt.Sprintf("%s | 📍 %s", dayHeader, *day.Location)
		}
		if day.Weather != nil && *day.Weather != "" {
			dayHeader = fmt.Sprintf("%s | ☀️  %s", dayHeader, *day.Weather)
		}
		if day.Mood != nil && *day.Mood != "" {
			dayHeader = fmt.Sprintf("%s | 😊 %s", dayHeader, *day.Mood)
		}

		items = append(items, m.flattenEntries(day.Entries, dayHeader, false, today)...)
	}

	return items
}

func (m Model) flattenEntries(entries []domain.Entry, header string, forceOverdue bool, today time.Time) []EntryItem {
	var items []EntryItem

	entriesToProcess := entries

	if forceOverdue {
		entriesToProcess = m.filterOverdueContext(entries, m.expandedOverdueContextID)
	}

	parentMap := make(map[int64][]domain.Entry)
	var roots []domain.Entry

	for _, e := range entriesToProcess {
		if e.ParentID == nil {
			roots = append(roots, e)
		} else {
			parentMap[*e.ParentID] = append(parentMap[*e.ParentID], e)
		}
	}

	var flatten func(entry domain.Entry, depth int, showHeader bool)
	flatten = func(entry domain.Entry, depth int, showHeader bool) {
		children := parentMap[entry.ID]
		hasChildren := len(children) > 0

		isCollapsed, hasCollapseState := m.collapsed[entry.EntityID]
		if !hasCollapseState && hasChildren {
			if entry.Type == domain.EntryTypeAnswered {
				isCollapsed = false
			} else {
				isCollapsed = true
			}
		}

		hiddenCount := 0
		if isCollapsed && hasChildren {
			hiddenCount = countEntryDescendants(entry.ID, parentMap)
		}

		entryIsOverdue := forceOverdue || entry.IsOverdue(today)
		item := EntryItem{
			Entry:            entry,
			IsOverdue:        entryIsOverdue,
			Indent:           depth,
			HasChildren:      hasChildren,
			HiddenChildCount: hiddenCount,
		}
		if showHeader {
			item.DayHeader = header
		}
		items = append(items, item)

		if !isCollapsed {
			for _, child := range children {
				flatten(child, depth+1, false)
			}
		}
	}

	for i, root := range roots {
		flatten(root, 0, i == 0)
	}

	return items
}

func (m Model) expandAllSiblings() Model {
	if len(m.entries) == 0 || m.days == nil {
		return m
	}

	selectedEntry := m.entries[m.selectedIdx].Entry
	parentID := selectedEntry.ParentID

	for _, day := range m.days {
		for _, entry := range day.Entries {
			if (parentID == nil && entry.ParentID == nil) ||
				(parentID != nil && entry.ParentID != nil && *parentID == *entry.ParentID) {
				m.collapsed[entry.EntityID] = false
			}
		}
	}
	m.entries = m.flattenDays(m.days)
	return m.ensuredVisible()
}

func (m Model) collapseAllSiblings() Model {
	if len(m.entries) == 0 || m.days == nil {
		return m
	}

	selectedEntry := m.entries[m.selectedIdx].Entry
	parentID := selectedEntry.ParentID

	for _, day := range m.days {
		for _, entry := range day.Entries {
			if (parentID == nil && entry.ParentID == nil) ||
				(parentID != nil && entry.ParentID != nil && *parentID == *entry.ParentID) {
				m.collapsed[entry.EntityID] = true
			}
		}
	}
	m.entries = m.flattenDays(m.days)
	return m.ensuredVisible()
}

func (m Model) toggleOverdueContext() Model {
	if len(m.entries) == 0 {
		return m
	}

	selectedEntry := m.entries[m.selectedIdx].Entry

	// Only toggle for overdue tasks
	if !selectedEntry.IsOverdue(time.Now()) {
		return m
	}

	// Toggle context visibility for this entry
	if m.expandedOverdueContextID != nil && *m.expandedOverdueContextID == selectedEntry.ID {
		m.expandedOverdueContextID = nil
	} else {
		id := selectedEntry.ID
		m.expandedOverdueContextID = &id
	}

	return m
}

func (m Model) filterOverdueContext(entries []domain.Entry, expandedOverdueContextID *int64) []domain.Entry {
	// First pass: keep only tasks
	var filtered []domain.Entry
	for _, e := range entries {
		if e.Type == domain.EntryTypeTask {
			filtered = append(filtered, e)
		}
	}

	// Build ID maps for lookups
	taskIDSet := make(map[int64]bool)
	for _, e := range filtered {
		taskIDSet[e.ID] = true
	}

	entryByID := make(map[int64]domain.Entry)
	for _, e := range entries {
		entryByID[e.ID] = e
	}

	// If a task is selected and context is expanded, include its full ancestry path
	var expandedAncestrySet map[int64]bool
	if expandedOverdueContextID != nil {
		expandedAncestrySet = make(map[int64]bool)
		expandedAncestrySet[*expandedOverdueContextID] = true

		// Trace path from expanded entry to root
		current, exists := entryByID[*expandedOverdueContextID]
		for exists && current.ParentID != nil {
			parent, parentExists := entryByID[*current.ParentID]
			if !parentExists {
				break
			}
			expandedAncestrySet[*current.ParentID] = true
			current = parent
		}
	}

	// Include all entries that are either tasks or in the expanded ancestry path
	var result []domain.Entry
	for _, e := range entries {
		isTask := taskIDSet[e.ID]
		inExpandedPath := expandedAncestrySet != nil && expandedAncestrySet[e.ID]

		if isTask || inExpandedPath {
			result = append(result, e)
		}
	}

	// Orphan any non-task parents for entries not in the expanded path
	for i := range result {
		if result[i].ParentID != nil {
			_, parentIsTask := taskIDSet[*result[i].ParentID]
			inExpandedPath := expandedAncestrySet != nil && expandedAncestrySet[*result[i].ParentID]

			if !parentIsTask && !inExpandedPath {
				result[i].ParentID = nil
			}
		}
	}

	return result
}

func (m Model) ensureSelectedAndAncestorsExpanded() Model {
	if len(m.entries) == 0 || m.days == nil {
		return m
	}

	selectedEntry := m.entries[m.selectedIdx].Entry

	entryByID := make(map[int64]domain.Entry)
	for _, day := range m.days {
		for _, entry := range day.Entries {
			entryByID[entry.ID] = entry
		}
	}
	current := selectedEntry
	for current.ParentID != nil {
		parent, exists := entryByID[*current.ParentID]
		if !exists {
			break
		}
		m.collapsed[parent.EntityID] = false
		current = parent
	}

	m.entries = m.flattenDays(m.days)
	return m
}

func (m Model) openURLCmd(content string) tea.Cmd {
	return func() tea.Msg {
		urls := domain.ExtractURLs(content)
		if len(urls) == 0 {
			return errMsg{err: fmt.Errorf("no URL found in entry")}
		}

		url := urls[0]
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default:
			return errMsg{err: fmt.Errorf("unsupported platform: %s", runtime.GOOS)}
		}

		if err := cmd.Start(); err != nil {
			return errMsg{err: fmt.Errorf("failed to open URL: %w", err)}
		}

		return nil
	}
}

func (m Model) loadLocationsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.bujoService == nil {
			return locationsLoadedMsg{locations: nil}
		}
		ctx := context.Background()
		now := time.Now()
		from := now.AddDate(0, -locationHistoryMonths, 0)
		history, err := m.bujoService.GetLocationHistory(ctx, from, now)
		if err != nil {
			return locationsLoadedMsg{locations: nil}
		}
		seen := make(map[string]bool)
		var locations []string
		for _, dayCtx := range history {
			if dayCtx.Location != nil && *dayCtx.Location != "" {
				lower := strings.ToLower(*dayCtx.Location)
				if !seen[lower] {
					seen[lower] = true
					locations = append(locations, *dayCtx.Location)
				}
			}
		}
		return locationsLoadedMsg{locations: locations}
	}
}

func (m Model) getAncestryChain(entryID int64) []domain.Entry {
	var ancestors []domain.Entry
	entryMap := make(map[int64]domain.Entry)

	for _, item := range m.entries {
		entryMap[item.Entry.ID] = item.Entry
	}

	currentID := entryID
	for {
		entry, exists := entryMap[currentID]
		if !exists {
			break
		}
		if entry.ParentID == nil {
			break
		}
		parent, parentExists := entryMap[*entry.ParentID]
		if !parentExists {
			break
		}
		ancestors = append([]domain.Entry{parent}, ancestors...)
		currentID = *entry.ParentID
	}

	return ancestors
}

func (m Model) launchExternalEditorCmd() tea.Cmd {
	tempFile := CaptureTempFilePath()
	draftContent := ""
	if content, exists := LoadDraft(m.draftPath); exists {
		draftContent = content
	}

	if err := PrepareEditorFile(tempFile, draftContent); err != nil {
		return func() tea.Msg {
			return errMsg{err: fmt.Errorf("failed to prepare editor file: %w", err)}
		}
	}

	editorCmd := GetEditorCommand()
	cmd := BuildEditorCmd(editorCmd, tempFile)
	cmd.Stdin = nil

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return editorFinishedMsg{content: "", err: err}
		}
		content, readErr := ReadEditorResult(tempFile)
		return editorFinishedMsg{content: content, err: readErr}
	})
}
