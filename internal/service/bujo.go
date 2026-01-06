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
	GetOverdue(ctx context.Context, date time.Time) ([]domain.Entry, error)
	Update(ctx context.Context, entry domain.Entry) error
}

type DayContextRepository interface {
	Upsert(ctx context.Context, dayCtx domain.DayContext) error
	GetByDate(ctx context.Context, date time.Time) (*domain.DayContext, error)
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

func (s *BujoService) SetLocation(ctx context.Context, date time.Time, location string) error {
	dayCtx := domain.DayContext{
		Date:     date,
		Location: &location,
	}
	return s.dayCtxRepo.Upsert(ctx, dayCtx)
}

func (s *BujoService) MarkDone(ctx context.Context, id int64) error {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("entry %d not found", id)
	}

	entry.Type = domain.EntryTypeDone
	return s.entryRepo.Update(ctx, *entry)
}
