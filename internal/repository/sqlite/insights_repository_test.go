package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsightsRepository_IsAvailable(t *testing.T) {
	t.Run("available when db is set", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)
		assert.True(t, repo.IsAvailable())
	})

	t.Run("not available when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		assert.False(t, repo.IsAvailable())
	})
}

func TestInsightsRepository_GetLatestSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("returns most recent summary", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		summary, err := repo.GetLatestSummary(ctx)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Equal(t, "2026-01-27", summary.WeekStart)
		assert.Equal(t, "2026-02-02", summary.WeekEnd)
		assert.Contains(t, summary.SummaryText, "Jan 27")
	})

	t.Run("returns nil when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		summary, err := repo.GetLatestSummary(ctx)
		require.NoError(t, err)
		assert.Nil(t, summary)
	})
}

func TestInsightsRepository_GetSummaries(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns summaries ordered by week_start desc", func(t *testing.T) {
		summaries, err := repo.GetSummaries(ctx, 10)
		require.NoError(t, err)
		require.Len(t, summaries, 3)
		assert.Equal(t, "2026-01-27", summaries[0].WeekStart)
		assert.Equal(t, "2026-01-20", summaries[1].WeekStart)
		assert.Equal(t, "2026-01-13", summaries[2].WeekStart)
	})

	t.Run("respects limit", func(t *testing.T) {
		summaries, err := repo.GetSummaries(ctx, 2)
		require.NoError(t, err)
		require.Len(t, summaries, 2)
	})

	t.Run("returns empty when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		summaries, err := repo.GetSummaries(ctx, 10)
		require.NoError(t, err)
		assert.Empty(t, summaries)
	})
}

func TestInsightsRepository_GetTopicsForSummary(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns topics for summary", func(t *testing.T) {
		topics, err := repo.GetTopicsForSummary(ctx, 1)
		require.NoError(t, err)
		require.Len(t, topics, 2)
		assert.Equal(t, "GenAI", topics[0].Topic)
	})

	t.Run("returns empty for non-existent summary", func(t *testing.T) {
		topics, err := repo.GetTopicsForSummary(ctx, 999)
		require.NoError(t, err)
		assert.Empty(t, topics)
	})
}

func TestInsightsRepository_GetActiveInitiatives(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns only active initiatives", func(t *testing.T) {
		initiatives, err := repo.GetActiveInitiatives(ctx, 10)
		require.NoError(t, err)
		require.Len(t, initiatives, 2)
		for _, i := range initiatives {
			assert.Equal(t, "active", i.Status)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		initiatives, err := repo.GetActiveInitiatives(ctx, 1)
		require.NoError(t, err)
		require.Len(t, initiatives, 1)
	})
}

func TestInsightsRepository_GetPendingActions(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns pending actions sorted by priority then due_date", func(t *testing.T) {
		actions, err := repo.GetPendingActions(ctx)
		require.NoError(t, err)
		require.Len(t, actions, 3)
		assert.Equal(t, "high", actions[0].Priority)
		assert.Equal(t, "pending", actions[0].Status)
	})

	t.Run("includes week_start from joined summary", func(t *testing.T) {
		actions, err := repo.GetPendingActions(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, actions)
		assert.NotEmpty(t, actions[0].WeekStart)
	})
}

func TestInsightsRepository_GetRecentDecisions(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns decisions ordered by date desc", func(t *testing.T) {
		decisions, err := repo.GetRecentDecisions(ctx, 10)
		require.NoError(t, err)
		require.Len(t, decisions, 2)
		assert.Equal(t, "2026-01-28", decisions[0].DecisionDate)
	})

	t.Run("respects limit", func(t *testing.T) {
		decisions, err := repo.GetRecentDecisions(ctx, 1)
		require.NoError(t, err)
		require.Len(t, decisions, 1)
	})
}

func TestInsightsRepository_GetSummaryForWeek(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns summary matching week_start range", func(t *testing.T) {
		summary, err := repo.GetSummaryForWeek(ctx, "2026-01-20", "2026-01-27")
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Equal(t, "2026-01-20", summary.WeekStart)
		assert.Equal(t, "2026-01-26", summary.WeekEnd)
	})

	t.Run("finds summary when week_start falls within range", func(t *testing.T) {
		// Monday before the Tuesday week_start in test data
		summary, err := repo.GetSummaryForWeek(ctx, "2026-01-19", "2026-01-26")
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Equal(t, "2026-01-20", summary.WeekStart)
	})

	t.Run("returns nil for non-existent week", func(t *testing.T) {
		summary, err := repo.GetSummaryForWeek(ctx, "2025-12-01", "2025-12-08")
		require.NoError(t, err)
		assert.Nil(t, summary)
	})

	t.Run("returns nil when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		summary, err := repo.GetSummaryForWeek(ctx, "2026-01-20", "2026-01-27")
		require.NoError(t, err)
		assert.Nil(t, summary)
	})
}

func TestInsightsRepository_GetActionsForWeek(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns actions from the specified week range", func(t *testing.T) {
		actions, err := repo.GetActionsForWeek(ctx, "2026-01-27", "2026-02-03")
		require.NoError(t, err)
		require.Len(t, actions, 2)
		for _, a := range actions {
			assert.Equal(t, "2026-01-27", a.WeekStart)
		}
	})

	t.Run("finds actions when week_start falls within range", func(t *testing.T) {
		// Monday before the actual week_start in test data
		actions, err := repo.GetActionsForWeek(ctx, "2026-01-26", "2026-02-02")
		require.NoError(t, err)
		require.Len(t, actions, 2)
	})

	t.Run("returns empty for week with no actions", func(t *testing.T) {
		actions, err := repo.GetActionsForWeek(ctx, "2025-12-01", "2025-12-08")
		require.NoError(t, err)
		assert.Empty(t, actions)
	})

	t.Run("returns empty when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		actions, err := repo.GetActionsForWeek(ctx, "2026-01-27", "2026-02-03")
		require.NoError(t, err)
		assert.Empty(t, actions)
	})
}

func TestInsightsRepository_GetDaysSinceLastSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("returns positive number for seeded data", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)
		days, err := repo.GetDaysSinceLastSummary(ctx)
		require.NoError(t, err)
		assert.Greater(t, days, 0)
	})

	t.Run("returns zero when no summaries exist", func(t *testing.T) {
		db := setupEmptyInsightsTestDB(t)
		repo := NewInsightsRepository(db)
		days, err := repo.GetDaysSinceLastSummary(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, days)
	})

	t.Run("returns -1 when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		days, err := repo.GetDaysSinceLastSummary(ctx)
		require.NoError(t, err)
		assert.Equal(t, -1, days)
	})
}
