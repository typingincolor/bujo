package domain

import (
	"context"
	"time"
)

type EntryRepository interface {
	Insert(ctx context.Context, entry Entry) (int64, error)
	GetByID(ctx context.Context, id int64) (*Entry, error)
	GetByEntityID(ctx context.Context, entityID EntityID) (*Entry, error)
	GetByDate(ctx context.Context, date time.Time) ([]Entry, error)
	GetByDateRange(ctx context.Context, from, to time.Time) ([]Entry, error)
	GetOverdue(ctx context.Context, date time.Time) ([]Entry, error)
	GetWithChildren(ctx context.Context, id int64) ([]Entry, error)
	GetChildren(ctx context.Context, parentID int64) ([]Entry, error)
	GetByListID(ctx context.Context, listID int64) ([]Entry, error)
	Update(ctx context.Context, entry Entry) error
	Delete(ctx context.Context, id int64) error
	DeleteWithChildren(ctx context.Context, id int64) error
	GetHistory(ctx context.Context, entityID EntityID) ([]Entry, error)
	GetAsOf(ctx context.Context, entityID EntityID, asOf time.Time) (*Entry, error)
	Search(ctx context.Context, opts SearchOptions) ([]Entry, error)
}

type HabitRepository interface {
	Insert(ctx context.Context, habit Habit) (int64, error)
	GetByID(ctx context.Context, id int64) (*Habit, error)
	GetByName(ctx context.Context, name string) (*Habit, error)
	GetOrCreate(ctx context.Context, name string, goalPerDay int) (*Habit, error)
	GetAll(ctx context.Context) ([]Habit, error)
	Update(ctx context.Context, habit Habit) error
	Delete(ctx context.Context, id int64) error
}

type HabitLogRepository interface {
	Insert(ctx context.Context, log HabitLog) (int64, error)
	GetByHabitID(ctx context.Context, habitID int64) ([]HabitLog, error)
	GetRange(ctx context.Context, habitID int64, start, end time.Time) ([]HabitLog, error)
	GetAllRange(ctx context.Context, start, end time.Time) ([]HabitLog, error)
	Delete(ctx context.Context, id int64) error
}

type DayContextRepository interface {
	Upsert(ctx context.Context, dayCtx DayContext) error
	GetByDate(ctx context.Context, date time.Time) (*DayContext, error)
	GetRange(ctx context.Context, start, end time.Time) ([]DayContext, error)
}

type SummaryRepository interface {
	Insert(ctx context.Context, summary Summary) (int64, error)
	Get(ctx context.Context, horizon SummaryHorizon, start, end time.Time) (*Summary, error)
	GetByHorizon(ctx context.Context, horizon SummaryHorizon) ([]Summary, error)
	Delete(ctx context.Context, id int64) error
}

type ListRepository interface {
	Create(ctx context.Context, name string) (*List, error)
	GetByID(ctx context.Context, id int64) (*List, error)
	GetByName(ctx context.Context, name string) (*List, error)
	GetByEntityID(ctx context.Context, entityID EntityID) (*List, error)
	GetAll(ctx context.Context) ([]List, error)
	Rename(ctx context.Context, id int64, newName string) error
	Delete(ctx context.Context, id int64) error
	GetItemCount(ctx context.Context, listID int64) (int, error)
	GetDoneCount(ctx context.Context, listID int64) (int, error)
}

type ListItemRepository interface {
	Insert(ctx context.Context, item ListItem) (int64, error)
	GetByID(ctx context.Context, id int64) (*ListItem, error)
	GetByEntityID(ctx context.Context, entityID EntityID) (*ListItem, error)
	GetByListEntityID(ctx context.Context, listEntityID EntityID) ([]ListItem, error)
	GetByListID(ctx context.Context, listID int64) ([]ListItem, error)
	Update(ctx context.Context, item ListItem) error
	Delete(ctx context.Context, id int64) error
	GetHistory(ctx context.Context, entityID EntityID) ([]ListItem, error)
	GetAtVersion(ctx context.Context, entityID EntityID, version int) (*ListItem, error)
	CountArchivable(ctx context.Context, olderThan time.Time) (int, error)
	DeleteArchivable(ctx context.Context, olderThan time.Time) (int, error)
}
