package domain

import (
	"errors"
	"time"
)

type EntryType string

const (
	EntryTypeTask        EntryType = "task"
	EntryTypeNote        EntryType = "note"
	EntryTypeEvent       EntryType = "event"
	EntryTypeDone        EntryType = "done"
	EntryTypeMigrated    EntryType = "migrated"
	EntryTypeCancelled   EntryType = "cancelled"
	EntryTypeQuestion    EntryType = "question"
	EntryTypeAnswered    EntryType = "answered"
	EntryTypeAnswer      EntryType = "answer"
	EntryTypeMovedToList EntryType = "movedToList"
)

type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

var validPriorities = map[Priority]string{
	PriorityNone:   "",
	PriorityLow:    "!",
	PriorityMedium: "!!",
	PriorityHigh:   "!!!",
}

func (p Priority) IsValid() bool {
	_, ok := validPriorities[p]
	return ok
}

func (p Priority) Symbol() string {
	return validPriorities[p]
}

func (p Priority) Cycle() Priority {
	switch p {
	case PriorityNone:
		return PriorityLow
	case PriorityLow:
		return PriorityMedium
	case PriorityMedium:
		return PriorityHigh
	default:
		return PriorityNone
	}
}

func ParsePriority(s string) (Priority, error) {
	if s == "" {
		return PriorityNone, nil
	}
	p := Priority(s)
	if !p.IsValid() {
		return PriorityNone, errors.New("invalid priority: must be none, low, medium, or high")
	}
	return p, nil
}

var validEntryTypes = map[EntryType]string{
	EntryTypeTask:        "•",
	EntryTypeNote:        "–",
	EntryTypeEvent:       "○",
	EntryTypeDone:        "✓",
	EntryTypeMigrated:    "→",
	EntryTypeCancelled:   "✗",
	EntryTypeQuestion:    "?",
	EntryTypeAnswered:    "★",
	EntryTypeAnswer:      "↳",
	EntryTypeMovedToList: "^",
}

func (et EntryType) IsValid() bool {
	_, ok := validEntryTypes[et]
	return ok
}

func (et EntryType) Symbol() string {
	return validEntryTypes[et]
}

type Entry struct {
	ID             int64
	EntityID       EntityID
	Type           EntryType
	Content        string
	Priority       Priority
	ParentID       *int64
	ParentEntityID *EntityID
	Depth          int
	Location       *string
	ScheduledDate  *time.Time
	CreatedAt      time.Time
	SortOrder      int
	MigrationCount int
}

func NewEntry(entryType EntryType, content string, scheduledDate *time.Time) Entry {
	return Entry{
		EntityID:      NewEntityID(),
		Type:          entryType,
		Content:       content,
		ScheduledDate: scheduledDate,
		CreatedAt:     time.Now(),
	}
}

func (e Entry) IsComplete() bool {
	return e.Type == EntryTypeDone || e.Type == EntryTypeCancelled || e.Type == EntryTypeAnswered
}

func (e Entry) HasParent() bool {
	return e.ParentEntityID != nil && !e.ParentEntityID.IsEmpty()
}

func (e Entry) IsOverdue(today time.Time) bool {
	if e.IsComplete() {
		return false
	}
	if e.Type == EntryTypeNote || e.Type == EntryTypeEvent {
		return false
	}
	if e.ScheduledDate == nil {
		return false
	}
	scheduledDate := e.ScheduledDate.Year()*10000 + int(e.ScheduledDate.Month())*100 + e.ScheduledDate.Day()
	todayDate := today.Year()*10000 + int(today.Month())*100 + today.Day()
	return scheduledDate < todayDate
}

func (e Entry) Validate() error {
	if !e.Type.IsValid() {
		return errors.New("invalid entry type")
	}
	if e.Content == "" {
		return errors.New("content cannot be empty")
	}
	if e.Depth < 0 {
		return errors.New("depth cannot be negative")
	}
	return nil
}

// cycleableTypes defines which entry types can have their type changed.
// SYNC: This logic is duplicated in frontend/src/components/bujo/EntryActions/types.ts
// for instant UI validation. When modifying, update both files.
var cycleableTypes = map[EntryType]bool{
	EntryTypeTask:     true,
	EntryTypeNote:     true,
	EntryTypeEvent:    true,
	EntryTypeQuestion: true,
}

// Entry action validation methods.
// SYNC: These rules are duplicated in frontend/src/components/bujo/EntryActions/types.ts
// (ACTION_REGISTRY appliesTo functions) for instant UI validation.
// When modifying validation rules, update both files.

func (e Entry) CanCancel() bool {
	return e.Type != EntryTypeCancelled
}

func (e Entry) CanUncancel() bool {
	return e.Type == EntryTypeCancelled
}

func (e Entry) CanCycleType() bool {
	return cycleableTypes[e.Type]
}

func (e Entry) CanEdit() bool {
	return e.Type != EntryTypeCancelled
}

func (e Entry) CanMigrate() bool {
	return e.Type == EntryTypeTask
}

func (e Entry) CanAnswer() bool {
	return e.Type == EntryTypeQuestion
}

func (e Entry) CanAddChild() bool {
	return e.Type != EntryTypeQuestion
}

func (e Entry) CanMoveToList() bool {
	return e.Type == EntryTypeTask
}

func (e Entry) CanMoveToRoot() bool {
	return e.ParentID != nil
}

func (e Entry) CanCyclePriority() bool {
	return true
}

func (e Entry) CanDelete() bool {
	return true
}
