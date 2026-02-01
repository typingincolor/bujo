package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type BujoService struct {
	entryRepo        domain.EntryRepository
	dayCtxRepo       domain.DayContextRepository
	parser           *domain.TreeParser
	listRepo         domain.ListRepository
	listItemRepo     domain.ListItemRepository
	entryToListMover domain.EntryToListMover
}

func NewBujoService(entryRepo domain.EntryRepository, dayCtxRepo domain.DayContextRepository, parser *domain.TreeParser) *BujoService {
	return &BujoService{
		entryRepo:  entryRepo,
		dayCtxRepo: dayCtxRepo,
		parser:     parser,
	}
}

func NewBujoServiceWithLists(entryRepo domain.EntryRepository, dayCtxRepo domain.DayContextRepository, parser *domain.TreeParser, listRepo domain.ListRepository, listItemRepo domain.ListItemRepository, entryToListMover domain.EntryToListMover) *BujoService {
	return &BujoService{
		entryRepo:        entryRepo,
		dayCtxRepo:       dayCtxRepo,
		parser:           parser,
		listRepo:         listRepo,
		listItemRepo:     listItemRepo,
		entryToListMover: entryToListMover,
	}
}

type LogEntriesOptions struct {
	Date     time.Time
	Location *string
	ParentID *int64
}

func (s *BujoService) LogEntries(ctx context.Context, input string, opts LogEntriesOptions) ([]int64, error) {
	entries, err := s.parser.Parse(input)
	if err != nil {
		return nil, err
	}

	var parentDepth int
	if opts.ParentID != nil {
		parent, err := s.entryRepo.GetByID(ctx, *opts.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, fmt.Errorf("parent entry %d not found", *opts.ParentID)
		}
		if parent.Type == domain.EntryTypeQuestion {
			return nil, fmt.Errorf("cannot add children to questions, use answer instead")
		}
		parentDepth = parent.Depth + 1
	}

	ids := make([]int64, 0, len(entries))
	idMap := make(map[int]int64) // maps original index to database ID

	for i, entry := range entries {
		entry.ScheduledDate = &opts.Date
		entry.Location = opts.Location
		entry.CreatedAt = time.Now()

		if entry.ParentID != nil {
			parentIdx := int(*entry.ParentID)
			if dbID, ok := idMap[parentIdx]; ok {
				entry.ParentID = &dbID
			}
		}

		if entry.ParentID == nil && opts.ParentID != nil && entry.Depth == 0 {
			entry.ParentID = opts.ParentID
		}

		if opts.ParentID != nil {
			entry.Depth += parentDepth
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
	Mood     *string
	Weather  *string
	Today    []domain.Entry
}

type DayEntries struct {
	Date     time.Time
	Location *string
	Mood     *string
	Weather  *string
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

	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if dayCtx != nil {
		agenda.Location = dayCtx.Location
		agenda.Mood = dayCtx.Mood
		agenda.Weather = dayCtx.Weather
	}

	today, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	agenda.Today = today

	return agenda, nil
}

func (s *BujoService) GetMultiDayAgenda(ctx context.Context, from, to time.Time) (*MultiDayAgenda, error) {
	agenda := &MultiDayAgenda{}

	overdue, err := s.entryRepo.GetOverdue(ctx)
	if err != nil {
		return nil, err
	}
	agenda.Overdue = overdue

	days, err := s.GetDayEntries(ctx, from, to)
	if err != nil {
		return nil, err
	}
	agenda.Days = days

	return agenda, nil
}

func (s *BujoService) GetDayEntries(ctx context.Context, from, to time.Time) ([]DayEntries, error) {
	entries, err := s.entryRepo.GetByDateRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	dayContexts, err := s.dayCtxRepo.GetRange(ctx, from, to)
	if err != nil {
		return nil, err
	}
	contextMap := make(map[string]*domain.DayContext)
	for i := range dayContexts {
		dateKey := dayContexts[i].Date.Format("2006-01-02")
		contextMap[dateKey] = &dayContexts[i]
	}

	entryMap := make(map[string][]domain.Entry)
	for _, entry := range entries {
		if entry.ScheduledDate != nil {
			dateKey := entry.ScheduledDate.Format("2006-01-02")
			entryMap[dateKey] = append(entryMap[dateKey], entry)
		}
	}

	var days []DayEntries
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		day := DayEntries{
			Date:    d,
			Entries: entryMap[dateKey],
		}
		if dayCtx := contextMap[dateKey]; dayCtx != nil {
			day.Location = dayCtx.Location
			day.Mood = dayCtx.Mood
			day.Weather = dayCtx.Weather
		}
		days = append(days, day)
	}

	return days, nil
}

func (s *BujoService) GetOverdue(ctx context.Context) ([]domain.Entry, error) {
	return s.entryRepo.GetOverdue(ctx)
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

func (s *BujoService) GetEntry(ctx context.Context, id int64) (*domain.Entry, error) {
	return s.getEntry(ctx, id)
}

func (s *BujoService) SetLocation(ctx context.Context, date time.Time, location string) error {
	dayCtx, err := s.dayCtxRepo.GetByDate(ctx, date)
	if err != nil {
		return err
	}
	if dayCtx == nil {
		dayCtx = &domain.DayContext{Date: date}
	}
	dayCtx.Location = &location
	return s.dayCtxRepo.Upsert(ctx, *dayCtx)
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

	wasAnswer := entry.Type == domain.EntryTypeAnswer
	parentID := entry.ParentID

	entry.Type = domain.EntryTypeCancelled
	if err := s.entryRepo.Update(ctx, *entry); err != nil {
		return err
	}

	if wasAnswer && parentID != nil {
		if err := s.reopenParentQuestionIfNeeded(ctx, *parentID); err != nil {
			return err
		}
	}

	return nil
}

func (s *BujoService) UncancelEntry(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	if entry.Type != domain.EntryTypeCancelled {
		return nil
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

	if !entry.CanCycleType() {
		return s.retypeErrorMessage(entry.Type)
	}

	entry.Type = newType
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) retypeErrorMessage(entryType domain.EntryType) error {
	switch entryType {
	case domain.EntryTypeCancelled:
		return fmt.Errorf("cannot change type of cancelled entry: uncancel it first")
	case domain.EntryTypeDone:
		return fmt.Errorf("cannot change type of completed entry: use undo to reopen it first")
	case domain.EntryTypeMigrated:
		return fmt.Errorf("cannot change type of migrated entry: the entry has moved to a new date")
	case domain.EntryTypeAnswered:
		return fmt.Errorf("cannot change type of answered entry: reopen the question first")
	case domain.EntryTypeAnswer:
		return fmt.Errorf("cannot change type of answer entry: answers are tied to their parent question")
	default:
		return fmt.Errorf("cannot change type of %s entry", entryType)
	}
}

func (s *BujoService) EditEntry(ctx context.Context, id int64, newContent string) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	if !entry.CanEdit() {
		return fmt.Errorf("cannot edit cancelled entry: uncancel it first to make changes")
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

func (s *BujoService) CyclePriority(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Priority = entry.Priority.Cycle()
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) DeleteEntry(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	wasAnswer := entry.Type == domain.EntryTypeAnswer
	parentID := entry.ParentID

	if err := s.entryRepo.DeleteWithChildren(ctx, id); err != nil {
		return err
	}

	if wasAnswer && parentID != nil {
		if err := s.reopenParentQuestionIfNeeded(ctx, *parentID); err != nil {
			return err
		}
	}

	return nil
}

func (s *BujoService) DeleteEntryAndReparent(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	children, err := s.entryRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}

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

	if entry.Type != domain.EntryTypeTask {
		return 0, fmt.Errorf("only tasks can be migrated, this is a %s", entry.Type)
	}

	children, err := s.entryRepo.GetChildren(ctx, id)
	if err != nil {
		return 0, err
	}

	originalChildTypes := make([]domain.EntryType, len(children))
	for i, child := range children {
		originalChildTypes[i] = child.Type
	}

	entry.Type = domain.EntryTypeMigrated
	if err := s.entryRepo.Update(ctx, *entry); err != nil {
		return 0, err
	}

	for i := range children {
		children[i].Type = domain.EntryTypeMigrated
		if err := s.entryRepo.Update(ctx, children[i]); err != nil {
			return 0, err
		}
	}

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

	for i, child := range children {
		newChild := domain.Entry{
			Type:          originalChildTypes[i],
			Content:       child.Content,
			ParentID:      &newParentID,
			Depth:         child.Depth,
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

	moveToRoot := opts.MoveToRoot != nil && *opts.MoveToRoot
	if moveToRoot {
		entry.ParentID = nil
		entry.Depth = 0
	}

	if !moveToRoot && opts.NewParentID != nil {
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

	if opts.NewLoggedDate != nil {
		entry.ScheduledDate = opts.NewLoggedDate
	}

	if err := s.entryRepo.Update(ctx, *entry); err != nil {
		return err
	}

	depthDelta := entry.Depth - oldDepth
	if depthDelta != 0 {
		if err := s.updateChildrenDepths(ctx, entry.ID, depthDelta); err != nil {
			return err
		}
	}

	if opts.NewLoggedDate != nil {
		if err := s.updateChildrenDates(ctx, entry.ID, *opts.NewLoggedDate); err != nil {
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

	rootID := id
	current := entry

	if current.ParentID != nil {
		rootID = *current.ParentID
		parent, err := s.entryRepo.GetByID(ctx, rootID)
		if err != nil {
			return nil, err
		}
		current = parent
	}

	for i := 0; i < ancestorLevels && current.ParentID != nil; i++ {
		rootID = *current.ParentID
		parent, err := s.entryRepo.GetByID(ctx, rootID)
		if err != nil {
			return nil, err
		}
		current = parent
	}

	return s.entryRepo.GetWithChildren(ctx, rootID)
}

func (s *BujoService) ParseEntries(content string) ([]domain.Entry, error) {
	return s.parser.Parse(content)
}

func (s *BujoService) SearchEntries(ctx context.Context, opts domain.SearchOptions) ([]domain.Entry, error) {
	return s.entryRepo.Search(ctx, opts)
}

func (s *BujoService) GetEntryAncestors(ctx context.Context, id int64) ([]domain.Entry, error) {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	var ancestors []domain.Entry
	current := entry

	for current.ParentID != nil {
		parent, err := s.entryRepo.GetByID(ctx, *current.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			break
		}
		ancestors = append([]domain.Entry{*parent}, ancestors...)
		current = parent
	}

	return ancestors, nil
}

func (s *BujoService) GetEntriesAncestorsMap(ctx context.Context, ids []int64) (map[int64][]domain.Entry, error) {
	if len(ids) == 0 {
		return make(map[int64][]domain.Entry), nil
	}

	result := make(map[int64][]domain.Entry)

	entriesMap := make(map[int64]*domain.Entry)
	for _, id := range ids {
		entry, err := s.entryRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			entriesMap[id] = entry
		}
	}

	parentIDsSet := make(map[int64]bool)
	for _, entry := range entriesMap {
		current := entry
		for current.ParentID != nil {
			parentIDsSet[*current.ParentID] = true
			current = entriesMap[*current.ParentID]
			if current == nil {
				break
			}
		}
	}

	for parentID := range parentIDsSet {
		if _, exists := entriesMap[parentID]; !exists {
			parent, err := s.entryRepo.GetByID(ctx, parentID)
			if err != nil {
				continue
			}
			if parent != nil {
				entriesMap[parentID] = parent
			}
		}
	}

	for _, id := range ids {
		entry := entriesMap[id]
		if entry == nil {
			continue
		}

		var ancestors []domain.Entry
		current := entry

		for current.ParentID != nil {
			parent := entriesMap[*current.ParentID]
			if parent == nil {
				break
			}
			ancestors = append([]domain.Entry{*parent}, ancestors...)
			current = parent
		}

		result[id] = ancestors
	}

	return result, nil
}

func (s *BujoService) MarkAnswered(ctx context.Context, id int64, answerText string) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	if entry.Type != domain.EntryTypeQuestion {
		return fmt.Errorf("only questions can be marked answered, this is a %s", entry.Type)
	}

	if answerText == "" {
		return fmt.Errorf("answer text is required")
	}

	if len(answerText) > 512 {
		return fmt.Errorf("answer text too long (max 512 characters, got %d)", len(answerText))
	}

	entry.Type = domain.EntryTypeAnswered
	if err := s.entryRepo.Update(ctx, *entry); err != nil {
		return err
	}

	updated, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if updated == nil {
		return fmt.Errorf("question entry not found after update")
	}

	answerEntry := domain.Entry{
		Type:          domain.EntryTypeAnswer,
		Content:       answerText,
		ParentID:      &updated.ID,
		Depth:         updated.Depth + 1,
		ScheduledDate: updated.ScheduledDate,
		CreatedAt:     time.Now(),
	}

	if _, err := s.entryRepo.Insert(ctx, answerEntry); err != nil {
		return fmt.Errorf("failed to add answer: %w", err)
	}

	return nil
}

func (s *BujoService) ReopenQuestion(ctx context.Context, id int64) error {
	entry, err := s.getEntry(ctx, id)
	if err != nil {
		return err
	}

	entry.Type = domain.EntryTypeQuestion
	return s.entryRepo.Update(ctx, *entry)
}

func (s *BujoService) reopenParentQuestionIfNeeded(ctx context.Context, parentID int64) error {
	parent, err := s.entryRepo.GetByID(ctx, parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return nil
	}

	if parent.Type == domain.EntryTypeAnswered {
		parent.Type = domain.EntryTypeQuestion
		return s.entryRepo.Update(ctx, *parent)
	}

	return nil
}

func (s *BujoService) ExportEntryMarkdown(ctx context.Context, id int64) (string, error) {
	entries, err := s.entryRepo.GetWithChildren(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get entry: %w", err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("entry %d not found", id)
	}

	return formatEntriesAsMarkdown(entries), nil
}

func formatEntriesAsMarkdown(entries []domain.Entry) string {
	childrenMap := make(map[int64][]domain.Entry)
	var root domain.Entry

	for _, entry := range entries {
		if entry.ParentID == nil {
			root = entry
		} else {
			childrenMap[*entry.ParentID] = append(childrenMap[*entry.ParentID], entry)
		}
	}

	return formatEntryMarkdown(root, childrenMap, 0)
}

func formatEntryMarkdown(entry domain.Entry, children map[int64][]domain.Entry, depth int) string {
	indent := strings.Repeat("  ", depth)
	symbol := entry.Type.Symbol()

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s%s %s\n", indent, symbol, entry.Content)

	for _, child := range children[entry.ID] {
		sb.WriteString(formatEntryMarkdown(child, children, depth+1))
	}

	return sb.String()
}

func (s *BujoService) MoveEntryToList(ctx context.Context, entryID int64, listID int64) error {
	entry, err := s.getEntry(ctx, entryID)
	if err != nil {
		return err
	}

	if !entry.CanMoveToList() {
		return fmt.Errorf("only tasks can be moved to lists, this is a %s", entry.Type)
	}

	children, err := s.entryRepo.GetChildren(ctx, entry.ID)
	if err != nil {
		return fmt.Errorf("failed to check for children: %w", err)
	}
	if len(children) > 0 {
		return fmt.Errorf("cannot move entry with children to a list")
	}

	list, err := s.listRepo.GetByID(ctx, listID)
	if err != nil {
		return err
	}
	if list == nil {
		return fmt.Errorf("list not found: %d", listID)
	}

	if err := s.entryToListMover.MoveEntryToList(ctx, *entry, list.EntityID); err != nil {
		return fmt.Errorf("failed to move entry to list: %w", err)
	}

	return nil
}
