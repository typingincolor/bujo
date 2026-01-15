package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestSummaryService_GetSummary(t *testing.T) {
	t.Run("generates and caches daily summary", func(t *testing.T) {
		today := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
		entries := []domain.Entry{
			{ID: 1, Type: domain.EntryTypeTask, Content: "Task 1"},
			{ID: 2, Type: domain.EntryTypeDone, Content: "Task 2"},
		}

		entryRepo := &mockEntryRepoForSummary{
			getByDateRangeFunc: func(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
				assert.Equal(t, today, from)
				assert.Equal(t, today, to)
				return entries, nil
			},
		}

		summaryRepo := &mockSummaryRepo{
			getFunc: func(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error) {
				return nil, nil // no cached summary
			},
			insertFunc: func(ctx context.Context, summary domain.Summary) (int64, error) {
				assert.Equal(t, domain.SummaryHorizonDaily, summary.Horizon)
				assert.Equal(t, "AI generated summary", summary.Content)
				return 1, nil
			},
		}

		generator := &mockSummaryGenerator{
			generateFunc: func(ctx context.Context, e []domain.Entry, h domain.SummaryHorizon) (string, error) {
				assert.Equal(t, entries, e)
				assert.Equal(t, domain.SummaryHorizonDaily, h)
				return "AI generated summary", nil
			},
		}

		svc := NewSummaryService(entryRepo, summaryRepo, generator)

		result, err := svc.GetSummary(context.Background(), domain.SummaryHorizonDaily, today)

		require.NoError(t, err)
		assert.Equal(t, "AI generated summary", result.Content)
	})

	// Smart caching tests:
	// - Completed periods (end date < today) always use cache
	// - Ongoing periods (end date >= today) always regenerate

	t.Run("uses cached summary for completed period", func(t *testing.T) {
		// Yesterday is a completed daily period - should always use cache
		yesterday := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)

		cachedSummary := &domain.Summary{
			ID:        1,
			Horizon:   domain.SummaryHorizonDaily,
			Content:   "Old cached summary",
			StartDate: yesterday,
			EndDate:   yesterday,
			CreatedAt: time.Date(2026, 1, 9, 10, 0, 0, 0, time.UTC),
		}

		summaryRepo := &mockSummaryRepo{
			getFunc: func(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error) {
				return cachedSummary, nil
			},
		}

		// Should NOT call entry repo or generator since period is complete
		entryRepo := &mockEntryRepoForSummary{
			getByDateRangeFunc: func(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
				t.Fatal("should not fetch entries for completed period")
				return nil, nil
			},
		}

		generator := &mockSummaryGenerator{
			generateFunc: func(ctx context.Context, e []domain.Entry, h domain.SummaryHorizon) (string, error) {
				t.Fatal("should not generate for completed period")
				return "", nil
			},
		}

		svc := NewSummaryService(entryRepo, summaryRepo, generator)

		result, err := svc.GetSummary(context.Background(), domain.SummaryHorizonDaily, yesterday)

		require.NoError(t, err)
		assert.Equal(t, "Old cached summary", result.Content)
	})

	t.Run("regenerates for ongoing period even with cache", func(t *testing.T) {
		// Today is an ongoing period - should always regenerate
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
		entries := []domain.Entry{{ID: 1, Type: domain.EntryTypeTask, Content: "Task"}}

		cachedSummary := &domain.Summary{
			ID:        1,
			Horizon:   domain.SummaryHorizonDaily,
			Content:   "Morning summary",
			StartDate: today,
			EndDate:   today,
			CreatedAt: time.Date(2026, 1, 10, 8, 0, 0, 0, time.UTC),
		}

		entryRepo := &mockEntryRepoForSummary{
			getByDateRangeFunc: func(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
				return entries, nil
			},
		}

		summaryRepo := &mockSummaryRepo{
			getFunc: func(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error) {
				return cachedSummary, nil
			},
			insertFunc: func(ctx context.Context, summary domain.Summary) (int64, error) {
				return 2, nil
			},
		}

		generator := &mockSummaryGenerator{
			generateFunc: func(ctx context.Context, e []domain.Entry, h domain.SummaryHorizon) (string, error) {
				return "Afternoon summary", nil
			},
		}

		svc := NewSummaryService(entryRepo, summaryRepo, generator)

		result, err := svc.GetSummary(context.Background(), domain.SummaryHorizonDaily, today)

		require.NoError(t, err)
		assert.Equal(t, "Afternoon summary", result.Content)
	})

	t.Run("forceRefresh bypasses cache", func(t *testing.T) {
		today := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
		entries := []domain.Entry{{ID: 1, Type: domain.EntryTypeTask, Content: "Task"}}

		cachedSummary := &domain.Summary{
			ID:        1,
			Horizon:   domain.SummaryHorizonDaily,
			Content:   "Cached summary",
			StartDate: today,
			EndDate:   today,
			CreatedAt: time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC), // recent cache
		}

		entryRepo := &mockEntryRepoForSummary{
			getByDateRangeFunc: func(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
				return entries, nil
			},
		}

		summaryRepo := &mockSummaryRepo{
			getFunc: func(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error) {
				return cachedSummary, nil
			},
			insertFunc: func(ctx context.Context, summary domain.Summary) (int64, error) {
				return 2, nil
			},
		}

		generator := &mockSummaryGenerator{
			generateFunc: func(ctx context.Context, e []domain.Entry, h domain.SummaryHorizon) (string, error) {
				return "Fresh summary", nil
			},
		}

		svc := NewSummaryService(entryRepo, summaryRepo, generator)

		result, err := svc.GetSummaryWithRefresh(context.Background(), domain.SummaryHorizonDaily, today, true)

		require.NoError(t, err)
		assert.Equal(t, "Fresh summary", result.Content)
	})

	t.Run("calculates weekly date range correctly", func(t *testing.T) {
		refDate := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC) // Friday

		entryRepo := &mockEntryRepoForSummary{
			getByDateRangeFunc: func(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
				// Week should be Mon Jan 5 to Sun Jan 11
				assert.Equal(t, time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), from)
				assert.Equal(t, time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC), to)
				return []domain.Entry{}, nil
			},
		}

		summaryRepo := &mockSummaryRepo{
			getFunc: func(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error) {
				return nil, nil
			},
			insertFunc: func(ctx context.Context, summary domain.Summary) (int64, error) {
				return 1, nil
			},
		}

		generator := &mockSummaryGenerator{
			generateFunc: func(ctx context.Context, e []domain.Entry, h domain.SummaryHorizon) (string, error) {
				return "Weekly summary", nil
			},
		}

		svc := NewSummaryService(entryRepo, summaryRepo, generator)

		_, err := svc.GetSummary(context.Background(), domain.SummaryHorizonWeekly, refDate)

		require.NoError(t, err)
	})
}

type mockEntryRepoForSummary struct {
	getByDateRangeFunc func(ctx context.Context, from, to time.Time) ([]domain.Entry, error)
}

func (m *mockEntryRepoForSummary) GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
	return m.getByDateRangeFunc(ctx, from, to)
}

type mockSummaryRepo struct {
	getFunc    func(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error)
	insertFunc func(ctx context.Context, summary domain.Summary) (int64, error)
}

func (m *mockSummaryRepo) Get(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error) {
	return m.getFunc(ctx, horizon, start, end)
}

func (m *mockSummaryRepo) Insert(ctx context.Context, summary domain.Summary) (int64, error) {
	return m.insertFunc(ctx, summary)
}

type mockSummaryGenerator struct {
	generateFunc func(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon) (string, error)
}

func (m *mockSummaryGenerator) GenerateSummary(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon) (string, error) {
	return m.generateFunc(ctx, entries, horizon)
}

func (m *mockSummaryGenerator) GenerateSummaryStream(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon, callback func(token string)) error {
	response, err := m.generateFunc(ctx, entries, horizon)
	if err != nil {
		return err
	}
	callback(response)
	return nil
}
