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

type listsLoadedMsg struct {
	lists     []domain.List
	summaries map[int64]*service.ListSummary
}

type listItemsLoadedMsg struct {
	listID int64
	items  []domain.ListItem
}

type listItemToggledMsg struct {
	itemID int64
}
