package tui

import (
	"time"

	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

type errMsg struct {
	err error
}

type agendaLoadedMsg struct {
	agenda *service.MultiDayAgenda
}

type entryUpdatedMsg struct {
	id int64
}

type entryDeletedMsg struct {
	id int64
}

type entryMovedToListMsg struct {
	entryID int64
}

type confirmDeleteMsg struct {
	entryID     int64
	hasChildren bool
}

type gotoDateMsg struct {
	date time.Time
}

type habitsLoadedMsg struct {
	habits []service.HabitStatus
}

type habitLoggedMsg struct {
	habitID int64
}

type habitLogRemovedMsg struct {
	habitID int64
}

type habitAddedMsg struct {
	name string
}

type habitDeletedMsg struct {
	habitID int64
}

type listsLoadedMsg struct {
	lists     []domain.List
	summaries map[int64]*service.ListSummary
}

type listCreatedMsg struct{}

type listsForMoveLoadedMsg struct {
	entryID      int64
	entryType    domain.EntryType
	entryContent string
	lists        []domain.List
}

type listItemsLoadedMsg struct {
	listID int64
	items  []domain.ListItem
}

type listItemToggledMsg struct {
	itemID int64
}

type listItemAddedMsg struct {
	listID int64
}

type listItemDeletedMsg struct {
	listID int64
}

type listItemEditedMsg struct {
	listID int64
}

type listItemMovedMsg struct {
	fromListID int64
	toListID   int64
}

type goalsLoadedMsg struct {
	goals []domain.Goal
}

type goalToggledMsg struct {
	goalID int64
}

type goalAddedMsg struct{}

type goalEditedMsg struct {
	goalID int64
}

type goalDeletedMsg struct {
	goalID int64
}

type goalMovedMsg struct {
	goalID int64
}

type journalGoalsLoadedMsg struct {
	goals []domain.Goal
}

type entryMigratedToGoalMsg struct {
	entryID int64
	goalID  int64
}

type summaryLoadedMsg struct {
	summary *domain.Summary
}

type summaryErrorMsg struct {
	err error
}

type summaryTokenMsg struct {
	token string
}

type searchResultsMsg struct {
	results []domain.Entry
	query   string
}

type statsLoadedMsg struct {
	stats *domain.Stats
}

type agendaReloadNeededMsg struct{}
