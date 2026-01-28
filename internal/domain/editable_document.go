package domain

import "time"

type ParsedLine struct {
	LineNumber    int
	Raw           string
	Depth         int
	Symbol        EntryType
	Priority      Priority
	Content       string
	EntityID      *EntityID
	MigrateTarget *time.Time
	IsValid       bool
	IsHeader      bool
	ErrorMessage  string
}

type EditableDocument struct {
	Date            time.Time
	Lines           []ParsedLine
	PendingDeletes  []EntityID
	OriginalMapping map[EntityID]int
}

type DiffOperation struct {
	Type        DiffOpType
	EntityID    *EntityID
	Entry       Entry
	MigrateDate *time.Time
	NewParentID *EntityID
	LineNumber  int
}

type DiffOpType int

const (
	DiffOpInsert DiffOpType = iota
	DiffOpUpdate
	DiffOpDelete
	DiffOpMigrate
	DiffOpReparent
)

type Changeset struct {
	Operations []DiffOperation
	Errors     []ParseError
}

type ParseError struct {
	LineNumber int
	Message    string
}
