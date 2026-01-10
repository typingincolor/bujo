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
	GetDeleted(ctx context.Context) ([]domain.Entry, error)
	Restore(ctx context.Context, entityID domain.EntityID) (int64, error)
	Search(ctx context.Context, opts domain.SearchOptions) ([]domain.Entry, error)
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

func (s *BujoService) SetMood(ctx context.Context, date time.Time, mood string) error {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return err
	}
	if dayCtx == nil {
		dayCtx = &domain.DayContext{Date: date}
	}
	dayCtx.Mood = &mood
	return s.dayCtxRepo.Upsert(ctx, *dayCtx)
}

func (s *BujoService) GetMood(ctx context.Context, date time.Time) (*string, error) {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if dayCtx == nil {
		return nil, nil
	}
	return dayCtx.Mood, nil
}

func (s *BujoService) GetMoodHistory(ctx context.Context, from, to time.Time) ([]domain.DayContext, error) {
	return s.dayCtxRepo.GetRange(ctx, from, to)
}

func (s *BujoService) ClearMood(ctx context.Context, date time.Time) error {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return err
	}
	if dayCtx == nil {
		return nil
	}
	dayCtx.Mood = nil
	return s.dayCtxRepo.Upsert(ctx, *dayCtx)
}

func (s *BujoService) SetWeather(ctx context.Context, date time.Time, weather string) error {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return err
	}
	if dayCtx == nil {
		dayCtx = &domain.DayContext{Date: date}
	}
	dayCtx.Weather = &weather
	return s.dayCtxRepo.Upsert(ctx, *dayCtx)
}

func (s *BujoService) GetWeather(ctx context.Context, date time.Time) (*string, error) {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if dayCtx == nil {
		return nil, nil
	}
	return dayCtx.Weather, nil
}

func (s *BujoService) GetWeatherHistory(ctx context.Context, from, to time.Time) ([]domain.DayContext, error) {
	return s.dayCtxRepo.GetRange(ctx, from, to)
}

func (s *BujoService) ClearWeather(ctx context.Context, date time.Time) error {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return err
	}
	if dayCtx == nil {
		return nil
	}
	dayCtx.Weather = nil
	return s.dayCtxRepo.Upsert(ctx, *dayCtx)
}

func (s *BujoService) MarkDone(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	if entry.Type != domain.EntryTypeTask {
		return fmt.Errorf("only tasks can be marked done, this is a %s", entry.Type)
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

func (s *BujoService) CancelEntry(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Type = domain.EntryTypeCancelled
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) UncancelEntry(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Type = domain.EntryTypeTask
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) RetypeEntry(ctx context.Context, id int64, newType domain.EntryType) error {
	if !newType.IsValid() {
		return fmt.Errorf("invalid entry type: %s", newType)
	}

	if newType == domain.EntryTypeDone || newType == domain.EntryTypeMigrated || newType == domain.EntryTypeCancelled {
		return fmt.Errorf("cannot retype to %s, use the appropriate command instead", newType)
	}

	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Type = newType
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

func (s *BujoService) EditEntryPriority(ctx context.Context, id int64, priority domain.Priority) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Priority = priority
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

	// Get all children and save their original types before marking as migrated
	children, err := s.entryRepo.GetChildren(ctx, id)
	if err != nil {
		return 0, err
	}

	originalChildTypes := make([]domain.EntryType, len(children))
	for i, child := range children {
		originalChildTypes[i] = child.Type
	}

	// Mark old entry as migrated
	entry.Type = domain.EntryTypeMigrated
	if err := s.entryRepo.Update(ctx, *entry); err != nil {
		return 0, err
	}

	// Mark all children as migrated
	for i := range children {
		children[i].Type = domain.EntryTypeMigrated
		if err := s.entryRepo.Update(ctx, children[i]); err != nil {
			return 0, err
		}
	}

	// Create new task on target date
	newEntry := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       entry.Content,
		ScheduledDate: &toDate,
		CreatedAt:     time.Now(),
	}

	newParentID, err := s.entryRepo.Insert(ctx, newEntry)
	if err != nil {
		return 0, err
	}

	// Create new children linked to new parent with original types
	for i, child := range children {
		newChild := domain.Entry{
			Type:          originalChildTypes[i],
			Content:       child.Content,
			ParentID:      &newParentID,
			ScheduledDate: &toDate,
			CreatedAt:     time.Now(),
		}
		if _, err := s.entryRepo.Insert(ctx, newChild); err != nil {
			return 0, err
		}
	}

	return newParentID, nil
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

	// Update children dates if logged date changed
	if opts.NewLoggedDate != nil {
		if err := s.updateChildrenDates(ctx, id, *opts.NewLoggedDate); err != nil {
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

func (s *BujoService) updateChildrenDates(ctx context.Context, parentID int64, newDate time.Time) error {
	children, err := s.entryRepo.GetChildren(ctx, parentID)
	if err != nil {
		return err
	}

	for _, child := range children {
		child.ScheduledDate = &newDate
		if err := s.entryRepo.Update(ctx, child); err != nil {
			return err
		}
		if err := s.updateChildrenDates(ctx, child.ID, newDate); err != nil {
			return err
		}
	}

	return nil
}

func (s *BujoService) GetOutstandingTasks(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
	entries, err := s.entryRepo.GetByDateRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	var tasks []domain.Entry
	for _, entry := range entries {
		if entry.Type == domain.EntryTypeTask {
			tasks = append(tasks, entry)
		}
	}

	return tasks, nil
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

func (s *BujoService) GetDeletedEntries(ctx context.Context) ([]domain.Entry, error) {
	return s.entryRepo.GetDeleted(ctx)
}

func (s *BujoService) RestoreEntry(ctx context.Context, entityID domain.EntityID) (int64, error) {
	return s.entryRepo.Restore(ctx, entityID)
}

func (s *BujoService) ParseEntries(content string) ([]domain.Entry, error) {
	return s.parser.Parse(content)
}

func (s *BujoService) SearchEntries(ctx context.Context, opts domain.SearchOptions) ([]domain.Entry, error) {
	return s.entryRepo.Search(ctx, opts)
}
