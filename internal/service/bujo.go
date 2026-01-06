package service

import (
	"context"
	"fmt"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type EntryRepository interface {
	Insert(ctx context.Context, entry domain.Entry) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Entry, error)
	GetByDate(ctx context.Context, date time.Time) ([]domain.Entry, error)
	GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error)
	GetOverdue(ctx context.Context, date time.Time) ([]domain.Entry, error)
	GetWithChildren(ctx context.Context, id int64) ([]domain.Entry, error)
	GetChildren(ctx context.Context, parentID int64) ([]domain.Entry, error)
	Update(ctx context.Context, entry domain.Entry) error
	Delete(ctx context.Context, id int64) error
	DeleteWithChildren(ctx context.Context, id int64) error
}

type DayContextRepository interface {
	Upsert(ctx context.Context, dayCtx domain.DayContext) error
	GetByDate(ctx context.Context, date time.Time) (*domain.DayContext, error)
	GetRange(ctx context.Context, start, end time.Time) ([]domain.DayContext, error)
	Delete(ctx context.Context, date time.Time) error
}

type BujoService struct {
	entryRepo  EntryRepository
	dayCtxRepo DayContextRepository
	parser     *domain.TreeParser
}

func NewBujoService(entryRepo EntryRepository, dayCtxRepo DayContextRepository, parser *domain.TreeParser) *BujoService {
	return &BujoService{
		entryRepo:  entryRepo,
		dayCtxRepo: dayCtxRepo,
		parser:     parser,
	}
}

type LogEntriesOptions struct {
	Date     time.Time
	Location *string
}

func (s *BujoService) LogEntries(ctx context.Context, input string, opts LogEntriesOptions) ([]int64, error) {
	entries, err := s.parser.Parse(input)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(entries))
	idMap := make(map[int]int64) // maps original index to database ID

	for i, entry := range entries {
		entry.ScheduledDate = &opts.Date
		entry.Location = opts.Location
		entry.CreatedAt = time.Now()

		// Update parent ID if this entry has a parent
		if entry.ParentID != nil {
			parentIdx := int(*entry.ParentID)
			if dbID, ok := idMap[parentIdx]; ok {
				entry.ParentID = &dbID
			}
		}

		id, err := s.entryRepo.Insert(ctx, entry)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
		idMap[i] = id
	}

	return ids, nil
}

type DailyAgenda struct {
	Date     time.Time
	Location *string
	Overdue  []domain.Entry
	Today    []domain.Entry
}

type DayEntries struct {
	Date     time.Time
	Location *string
	Entries  []domain.Entry
}

type MultiDayAgenda struct {
	Overdue []domain.Entry
	Days    []DayEntries
}

func (s *BujoService) GetDailyAgenda(ctx context.Context, date time.Time) (*DailyAgenda, error) {
	agenda := &DailyAgenda{
		Date: date,
	}

	// Get location for the day
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if dayCtx != nil {
		agenda.Location = dayCtx.Location
	}

	// Get overdue entries
	overdue, err := s.entryRepo.GetOverdue(ctx, date)
	if err != nil {
		return nil, err
	}
	agenda.Overdue = overdue

	// Get today's entries
	today, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	agenda.Today = today

	return agenda, nil
}

func (s *BujoService) GetMultiDayAgenda(ctx context.Context, from, to time.Time) (*MultiDayAgenda, error) {
	agenda := &MultiDayAgenda{}

	// Get overdue entries (before the from date)
	overdue, err := s.entryRepo.GetOverdue(ctx, from)
	if err != nil {
		return nil, err
	}
	agenda.Overdue = overdue

	// Get all entries in the range
	entries, err := s.entryRepo.GetByDateRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Get locations for the range
	locations, err := s.dayCtxRepo.GetRange(ctx, from, to)
	if err != nil {
		return nil, err
	}
	locationMap := make(map[string]*string)
	for _, loc := range locations {
		dateKey := loc.Date.Format("2006-01-02")
		locationMap[dateKey] = loc.Location
	}

	// Group entries by date
	entryMap := make(map[string][]domain.Entry)
	for _, entry := range entries {
		if entry.ScheduledDate != nil {
			dateKey := entry.ScheduledDate.Format("2006-01-02")
			entryMap[dateKey] = append(entryMap[dateKey], entry)
		}
	}

	// Build days array from from to to (inclusive)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		day := DayEntries{
			Date:     d,
			Location: locationMap[dateKey],
			Entries:  entryMap[dateKey],
		}
		agenda.Days = append(agenda.Days, day)
	}

	return agenda, nil
}

func (s *BujoService) getEntry(ctx context.Context, id int64) (*domain.Entry, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, fmt.Errorf("entry %d not found", id)
	}
	return entry, nil
}

func (s *BujoService) SetLocation(ctx context.Context, date time.Time, location string) error {
	dayCtx := domain.DayContext{
		Date:     date,
		Location: &location,
	}
	return s.dayCtxRepo.Upsert(ctx, dayCtx)
}

func (s *BujoService) GetLocationHistory(ctx context.Context, from, to time.Time) ([]domain.DayContext, error) {
	return s.dayCtxRepo.GetRange(ctx, from, to)
}

func (s *BujoService) GetLocation(ctx context.Context, date time.Time) (*string, error) {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if dayCtx == nil {
		return nil, nil
	}
	return dayCtx.Location, nil
}

func (s *BujoService) ClearLocation(ctx context.Context, date time.Time) error {
	return s.dayCtxRepo.Delete(ctx, date)
}

func (s *BujoService) MarkDone(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Type = domain.EntryTypeDone
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) Undo(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Type = domain.EntryTypeTask
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) EditEntry(ctx context.Context, id int64, newContent string) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Content = newContent
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) DeleteEntry(ctx context.Context, id int64) error {
	if _, err := s.getEntry(ctx, id); err != nil {
		return err
	}

	return s.entryRepo.DeleteWithChildren(ctx, id)
}

func (s *BujoService) DeleteEntryAndReparent(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	// Get children of this entry
	children, err := s.entryRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}

	// Reparent children to this entry's parent (may be nil for root)
	for _, child := range children {
		child.ParentID = entry.ParentID
		if entry.ParentID == nil {
			child.Depth = 0
		} else {
			child.Depth = entry.Depth
		}
		if err := s.entryRepo.Update(ctx, child); err != nil {
			return err
		}
	}

	// Now delete just the entry (not children)
	return s.entryRepo.Delete(ctx, id)
}

func (s *BujoService) HasChildren(ctx context.Context, id int64) (bool, error) {
	children, err := s.entryRepo.GetChildren(ctx, id)
	if err != nil {
		return false, err
	}
	return len(children) > 0, nil
}

func (s *BujoService) MigrateEntry(ctx context.Context, id int64, toDate time.Time) (int64, error) {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return 0, err
	}

	// Only tasks can be migrated
	if entry.Type != domain.EntryTypeTask {
		return 0, fmt.Errorf("only tasks can be migrated, this is a %s", entry.Type)
	}

	// Mark old entry as migrated
	entry.Type = domain.EntryTypeMigrated
	if err := s.entryRepo.Update(ctx, *entry); err != nil {
		return 0, err
	}

	// Create new task on target date
	newEntry := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       entry.Content,
		ScheduledDate: &toDate,
		CreatedAt:     time.Now(),
	}

	return s.entryRepo.Insert(ctx, newEntry)
}

type MoveOptions struct {
	NewParentID   *int64
	NewLoggedDate *time.Time
	MoveToRoot    *bool
}

func (s *BujoService) MoveEntry(ctx context.Context, id int64, opts MoveOptions) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	oldDepth := entry.Depth

	// Handle parent change
	if opts.MoveToRoot != nil && *opts.MoveToRoot {
		entry.ParentID = nil
		entry.Depth = 0
	} else if opts.NewParentID != nil {
		parent, err := s.entryRepo.GetByID(ctx, *opts.NewParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return fmt.Errorf("parent %d not found", *opts.NewParentID)
		}
		entry.ParentID = opts.NewParentID
		entry.Depth = parent.Depth + 1
	}

	// Handle logged date change
	if opts.NewLoggedDate != nil {
		entry.ScheduledDate = opts.NewLoggedDate
	}

	// Update the entry
	if err := s.entryRepo.Update(ctx, *entry); err != nil {
		return err
	}

	// Update children depths if parent changed
	depthDelta := entry.Depth - oldDepth
	if depthDelta != 0 {
		if err := s.updateChildrenDepths(ctx, id, depthDelta); err != nil {
			return err
		}
	}

	return nil
}

func (s *BujoService) updateChildrenDepths(ctx context.Context, parentID int64, depthDelta int) error {
	children, err := s.entryRepo.GetChildren(ctx, parentID)
	if err != nil {
		return err
	}

	for _, child := range children {
		child.Depth += depthDelta
		if err := s.entryRepo.Update(ctx, child); err != nil {
			return err
		}
		// Recursively update grandchildren
		if err := s.updateChildrenDepths(ctx, child.ID, depthDelta); err != nil {
			return err
		}
	}

	return nil
}

func (s *BujoService) GetEntryContext(ctx context.Context, id int64, ancestorLevels int) ([]domain.Entry, error) {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	// Walk up to find the root of the context we want to show
	rootID := id
	current := entry

	// Default behavior: go up to parent (if exists)
	if current.ParentID != nil {
		rootID = *current.ParentID
		parent, err := s.entryRepo.GetByID(ctx, rootID)
		if err != nil {
			return nil, err
		}
		current = parent
	}

	// Go up additional ancestor levels
	for i := 0; i < ancestorLevels && current.ParentID != nil; i++ {
		rootID = *current.ParentID
		parent, err := s.entryRepo.GetByID(ctx, rootID)
		if err != nil {
			return nil, err
		}
		current = parent
	}

	// Get the root and all its children
	return s.entryRepo.GetWithChildren(ctx, rootID)
}
