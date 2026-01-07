package tui

import "github.com/typingincolor/bujo/internal/service"

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
