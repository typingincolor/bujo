package domain

import (
	"errors"
	"time"
)

type EntryType string

const (
	EntryTypeTask      EntryType = "task"
	EntryTypeNote      EntryType = "note"
	EntryTypeEvent     EntryType = "event"
	EntryTypeDone      EntryType = "done"
	EntryTypeMigrated  EntryType = "migrated"
	EntryTypeCancelled EntryType = "cancelled"
)

var validEntryTypes = map[EntryType]string{
	EntryTypeTask:      "•",
	EntryTypeNote:      "–",
	EntryTypeEvent:     "○",
	EntryTypeDone:      "✓",
	EntryTypeMigrated:  "→",
	EntryTypeCancelled: "✗",
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
	ParentID       *int64
	ParentEntityID *EntityID
	Depth          int
	Location       *string
	ScheduledDate  *time.Time
	CreatedAt      time.Time
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
	return e.Type == EntryTypeDone || e.Type == EntryTypeCancelled
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
	return e.ScheduledDate.Before(today)
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
