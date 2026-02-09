package domain

import (
	"context"
	"time"
)

type EntryRepository interface {
	Insert(ctx context.Context, entry Entry) (int64, error)
	GetByID(ctx context.Context, id int64) (*Entry, error)
	GetByDate(ctx context.Context, date time.Time) ([]Entry, error)
	GetByDateRange(ctx context.Context, from, to time.Time) ([]Entry, error)
	GetAll(ctx context.Context) ([]Entry, error)
	GetOverdue(ctx context.Context) ([]Entry, error)
	GetWithChildren(ctx context.Context, id int64) ([]Entry, error)
	GetChildren(ctx context.Context, parentID int64) ([]Entry, error)
	Update(ctx context.Context, entry Entry) error
	Delete(ctx context.Context, id int64) error
	DeleteByDate(ctx context.Context, date time.Time) error
	DeleteAll(ctx context.Context) error
	DeleteWithChildren(ctx context.Context, id int64) error
	Search(ctx context.Context, opts SearchOptions) ([]Entry, error)
}

type HabitRepository interface {
	Insert(ctx context.Context, habit Habit) (int64, error)
	GetByID(ctx context.Context, id int64) (*Habit, error)
	GetByEntityID(ctx context.Context, entityID EntityID) (*Habit, error)
	GetByName(ctx context.Context, name string) (*Habit, error)
	GetOrCreate(ctx context.Context, name string, goalPerDay int) (*Habit, error)
	GetAll(ctx context.Context) ([]Habit, error)
	Update(ctx context.Context, habit Habit) error
	Delete(ctx context.Context, id int64) error
	DeleteAll(ctx context.Context) error
}

type HabitLogRepository interface {
	Insert(ctx context.Context, log HabitLog) (int64, error)
	GetByID(ctx context.Context, id int64) (*HabitLog, error)
	GetByHabitID(ctx context.Context, habitID int64) ([]HabitLog, error)
	GetRange(ctx context.Context, habitID int64, start, end time.Time) ([]HabitLog, error)
	GetRangeByEntityID(ctx context.Context, habitEntityID EntityID, start, end time.Time) ([]HabitLog, error)
	GetAllRange(ctx context.Context, start, end time.Time) ([]HabitLog, error)
	GetAll(ctx context.Context) ([]HabitLog, error)
	GetLastByHabitID(ctx context.Context, habitID int64) (*HabitLog, error)
	Delete(ctx context.Context, id int64) error
	DeleteAll(ctx context.Context) error
}

type DayContextRepository interface {
	Upsert(ctx context.Context, dayCtx DayContext) error
	GetByDate(ctx context.Context, date time.Time) (*DayContext, error)
	GetRange(ctx context.Context, start, end time.Time) ([]DayContext, error)
	GetAll(ctx context.Context) ([]DayContext, error)
	Delete(ctx context.Context, date time.Time) error
	DeleteAll(ctx context.Context) error
}

type ListRepository interface {
	Create(ctx context.Context, name string) (*List, error)
	InsertWithEntityID(ctx context.Context, list List) (int64, error)
	GetByID(ctx context.Context, id int64) (*List, error)
	GetByName(ctx context.Context, name string) (*List, error)
	GetByEntityID(ctx context.Context, entityID EntityID) (*List, error)
	GetAll(ctx context.Context) ([]List, error)
	Rename(ctx context.Context, id int64, newName string) error
	Delete(ctx context.Context, id int64) error
	DeleteAll(ctx context.Context) error
	GetItemCount(ctx context.Context, listID int64) (int, error)
	GetDoneCount(ctx context.Context, listID int64) (int, error)
}

type ListItemRepository interface {
	Insert(ctx context.Context, item ListItem) (int64, error)
	GetByID(ctx context.Context, id int64) (*ListItem, error)
	GetByEntityID(ctx context.Context, entityID EntityID) (*ListItem, error)
	GetByListEntityID(ctx context.Context, listEntityID EntityID) ([]ListItem, error)
	GetByListID(ctx context.Context, listID int64) ([]ListItem, error)
	GetAll(ctx context.Context) ([]ListItem, error)
	Update(ctx context.Context, item ListItem) error
	Delete(ctx context.Context, id int64) error
	DeleteAll(ctx context.Context) error
	GetHistory(ctx context.Context, entityID EntityID) ([]ListItem, error)
	GetAtVersion(ctx context.Context, entityID EntityID, version int) (*ListItem, error)
	CountArchivable(ctx context.Context, olderThan time.Time) (int, error)
	DeleteArchivable(ctx context.Context, olderThan time.Time) (int, error)
}

type TagRepository interface {
	InsertEntryTags(ctx context.Context, entryID int64, tags []string) error
	GetTagsForEntries(ctx context.Context, entryIDs []int64) (map[int64][]string, error)
	GetAllTags(ctx context.Context) ([]string, error)
	DeleteByEntryID(ctx context.Context, entryID int64) error
}

type MentionRepository interface {
	InsertEntryMentions(ctx context.Context, entryID int64, mentions []string) error
	GetMentionsForEntries(ctx context.Context, entryIDs []int64) (map[int64][]string, error)
	GetAllMentions(ctx context.Context) ([]string, error)
	DeleteByEntryID(ctx context.Context, entryID int64) error
}

type ChangeDetector interface {
	GetLastModified(ctx context.Context) (time.Time, error)
}

type EntryToListMover interface {
	MoveEntryToList(ctx context.Context, entry Entry, listEntityID EntityID) error
}
