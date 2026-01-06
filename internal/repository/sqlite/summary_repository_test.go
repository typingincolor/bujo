package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestSummaryRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSummaryRepository(db)
	ctx := context.Background()

	summary := domain.Summary{
		Horizon:   domain.SummaryHorizonWeekly,
		Content:   "Weekly reflection content",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
		CreatedAt: time.Now(),
	}

	id, err := repo.Insert(ctx, summary)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestSummaryRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSummaryRepository(db)
	ctx := context.Background()

	summary := domain.Summary{
		Horizon:   domain.SummaryHorizonWeekly,
		Content:   "Weekly reflection",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
		CreatedAt: time.Now(),
	}
	_, err := repo.Insert(ctx, summary)
	require.NoError(t, err)

	result, err := repo.Get(ctx, domain.SummaryHorizonWeekly, summary.StartDate, summary.EndDate)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domain.SummaryHorizonWeekly, result.Horizon)
	assert.Equal(t, "Weekly reflection", result.Content)
}

func TestSummaryRepository_Get_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSummaryRepository(db)
	ctx := context.Background()

	result, err := repo.Get(ctx, domain.SummaryHorizonDaily,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestSummaryRepository_GetByHorizon(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSummaryRepository(db)
	ctx := context.Background()

	summaries := []domain.Summary{
		{
			Horizon:   domain.SummaryHorizonWeekly,
			Content:   "Week 1",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
			CreatedAt: time.Now(),
		},
		{
			Horizon:   domain.SummaryHorizonWeekly,
			Content:   "Week 2",
			StartDate: time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
			CreatedAt: time.Now(),
		},
		{
			Horizon:   domain.SummaryHorizonDaily,
			Content:   "Day 1",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt: time.Now(),
		},
	}
	for _, s := range summaries {
		_, err := repo.Insert(ctx, s)
		require.NoError(t, err)
	}

	results, err := repo.GetByHorizon(ctx, domain.SummaryHorizonWeekly)

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestSummaryRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSummaryRepository(db)
	ctx := context.Background()

	summary := domain.Summary{
		Horizon:   domain.SummaryHorizonDaily,
		Content:   "To be deleted",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt: time.Now(),
	}
	id, err := repo.Insert(ctx, summary)
	require.NoError(t, err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err)

	result, err := repo.Get(ctx, domain.SummaryHorizonDaily, summary.StartDate, summary.EndDate)
	require.NoError(t, err)
	assert.Nil(t, result)
}
