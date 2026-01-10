package service

import (
	"context"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type SummaryEntryRepository interface {
	GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error)
}

type SummaryRepository interface {
	Get(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error)
	Insert(ctx context.Context, summary domain.Summary) (int64, error)
}

type SummaryGenerator interface {
	GenerateSummary(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon) (string, error)
}

type SummaryService struct {
	entryRepo   SummaryEntryRepository
	summaryRepo SummaryRepository
	generator   SummaryGenerator
}

func NewSummaryService(entryRepo SummaryEntryRepository, summaryRepo SummaryRepository, generator SummaryGenerator) *SummaryService {
	return &SummaryService{
		entryRepo:   entryRepo,
		summaryRepo: summaryRepo,
		generator:   generator,
	}
}

func (s *SummaryService) GetSummary(ctx context.Context, horizon domain.SummaryHorizon, refDate time.Time) (*domain.Summary, error) {
	return s.GetSummaryWithRefresh(ctx, horizon, refDate, false)
}

func (s *SummaryService) GetSummaryWithRefresh(ctx context.Context, horizon domain.SummaryHorizon, refDate time.Time, forceRefresh bool) (*domain.Summary, error) {
	startDate, endDate := s.calculateDateRange(horizon, refDate)

	if !forceRefresh {
		cached, err := s.summaryRepo.Get(ctx, horizon, startDate, endDate)
		if err != nil {
			return nil, err
		}

		if cached != nil && cached.IsRecent(refDate) {
			return cached, nil
		}
	}

	entries, err := s.entryRepo.GetByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	content, err := s.generator.GenerateSummary(ctx, entries, horizon)
	if err != nil {
		return nil, err
	}

	summary := domain.Summary{
		EntityID:  domain.NewEntityID(),
		Horizon:   horizon,
		Content:   content,
		StartDate: startDate,
		EndDate:   endDate,
		CreatedAt: time.Now(),
	}

	id, err := s.summaryRepo.Insert(ctx, summary)
	if err != nil {
		return nil, err
	}

	summary.ID = id
	return &summary, nil
}

func (s *SummaryService) calculateDateRange(horizon domain.SummaryHorizon, refDate time.Time) (time.Time, time.Time) {
	refDate = time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, refDate.Location())

	switch horizon {
	case domain.SummaryHorizonDaily:
		return refDate, refDate

	case domain.SummaryHorizonWeekly:
		weekday := int(refDate.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday is 7
		}
		monday := refDate.AddDate(0, 0, -(weekday - 1))
		sunday := monday.AddDate(0, 0, 6)
		return monday, sunday

	case domain.SummaryHorizonQuarterly:
		quarter := (refDate.Month()-1)/3 + 1
		startMonth := time.Month((quarter-1)*3 + 1)
		start := time.Date(refDate.Year(), startMonth, 1, 0, 0, 0, 0, refDate.Location())
		end := start.AddDate(0, 3, -1)
		return start, end

	case domain.SummaryHorizonAnnual:
		start := time.Date(refDate.Year(), 1, 1, 0, 0, 0, 0, refDate.Location())
		end := time.Date(refDate.Year(), 12, 31, 0, 0, 0, 0, refDate.Location())
		return start, end

	default:
		return refDate, refDate
	}
}
