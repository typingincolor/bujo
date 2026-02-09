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

func TestInsightsRepository_GetInitiativePortfolio(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all initiatives sorted by status priority then mention count", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		portfolio, err := repo.GetInitiativePortfolio(ctx)
		require.NoError(t, err)
		require.Len(t, portfolio, 3)

		// Active initiatives first (GenAI has 2 mentions, Tech Scorecard has 1)
		assert.Equal(t, "GenAI Integration", portfolio[0].Name)
		assert.Equal(t, "active", portfolio[0].Status)
		assert.Equal(t, 2, portfolio[0].MentionCount)

		assert.Equal(t, "Tech Scorecard", portfolio[1].Name)
		assert.Equal(t, "active", portfolio[1].Status)
		assert.Equal(t, 1, portfolio[1].MentionCount)

		// Completed last
		assert.Equal(t, "Q1 OKRs", portfolio[2].Name)
		assert.Equal(t, "completed", portfolio[2].Status)
		assert.Equal(t, 0, portfolio[2].MentionCount)
	})

	t.Run("includes last mention week and activity weeks", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		portfolio, err := repo.GetInitiativePortfolio(ctx)
		require.NoError(t, err)

		// GenAI Integration: mentioned in weeks Jan 13 and Jan 27
		assert.Equal(t, "2026-01-27", portfolio[0].LastMentionWeek)
		assert.Contains(t, portfolio[0].ActivityWeeks, "2026-01-13")
		assert.Contains(t, portfolio[0].ActivityWeeks, "2026-01-27")
	})

	t.Run("returns empty when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		portfolio, err := repo.GetInitiativePortfolio(ctx)
		require.NoError(t, err)
		assert.Empty(t, portfolio)
	})

	t.Run("returns empty for empty database", func(t *testing.T) {
		db := setupEmptyInsightsTestDB(t)
		repo := NewInsightsRepository(db)
		portfolio, err := repo.GetInitiativePortfolio(ctx)
		require.NoError(t, err)
		assert.Empty(t, portfolio)
	})
}

func TestInsightsRepository_GetInitiativeDetail(t *testing.T) {
	ctx := context.Background()

	t.Run("returns full detail for initiative with mentions, actions, and decisions", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		detail, err := repo.GetInitiativeDetail(ctx, 1)
		require.NoError(t, err)
		require.NotNil(t, detail)

		assert.Equal(t, "GenAI Integration", detail.Initiative.Name)
		assert.Equal(t, "active", detail.Initiative.Status)

		require.Len(t, detail.Updates, 2)
		assert.Equal(t, "2026-01-13", detail.Updates[0].WeekStart)
		assert.Equal(t, "Started planning AI integration approach", detail.Updates[0].UpdateText)
		assert.Equal(t, "2026-01-27", detail.Updates[1].WeekStart)
		assert.Equal(t, "AI integration sprint completed", detail.Updates[1].UpdateText)

		require.Len(t, detail.PendingActions, 2)
		for _, a := range detail.PendingActions {
			assert.Equal(t, "pending", a.Status)
		}

		require.Len(t, detail.Decisions, 1)
		assert.Equal(t, "Adopt Claude as primary AI provider", detail.Decisions[0].DecisionText)
	})

	t.Run("returns detail for initiative with no mentions", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		detail, err := repo.GetInitiativeDetail(ctx, 3)
		require.NoError(t, err)
		require.NotNil(t, detail)

		assert.Equal(t, "Q1 OKRs", detail.Initiative.Name)
		assert.Empty(t, detail.Updates)
		assert.Empty(t, detail.PendingActions)
		require.Len(t, detail.Decisions, 1)
		assert.Equal(t, "Move to biweekly sprints", detail.Decisions[0].DecisionText)
	})

	t.Run("returns nil for non-existent initiative", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		detail, err := repo.GetInitiativeDetail(ctx, 999)
		require.NoError(t, err)
		assert.Nil(t, detail)
	})

	t.Run("returns nil when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		detail, err := repo.GetInitiativeDetail(ctx, 1)
		require.NoError(t, err)
		assert.Nil(t, detail)
	})
}

func TestInsightsRepository_GetDistinctTopics(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all unique topic names sorted alphabetically", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		topics, err := repo.GetDistinctTopics(ctx)
		require.NoError(t, err)
		require.Len(t, topics, 5)
		assert.Equal(t, "GenAI", topics[0])
		assert.Equal(t, "Quarterly Planning", topics[1])
		assert.Equal(t, "Sprint Review", topics[2])
		assert.Equal(t, "Team Planning", topics[3])
		assert.Equal(t, "Tech Scorecard", topics[4])
	})

	t.Run("returns empty when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		topics, err := repo.GetDistinctTopics(ctx)
		require.NoError(t, err)
		assert.Empty(t, topics)
	})

	t.Run("returns empty for empty database", func(t *testing.T) {
		db := setupEmptyInsightsTestDB(t)
		repo := NewInsightsRepository(db)
		topics, err := repo.GetDistinctTopics(ctx)
		require.NoError(t, err)
		assert.Empty(t, topics)
	})
}

func TestInsightsRepository_GetTopicTimeline(t *testing.T) {
	ctx := context.Background()

	t.Run("returns topic mentions across weeks for given topic", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		timeline, err := repo.GetTopicTimeline(ctx, "GenAI")
		require.NoError(t, err)
		require.Len(t, timeline, 1)
		assert.Equal(t, "GenAI", timeline[0].Topic)
		assert.Equal(t, "Integration planning for AI features", timeline[0].Content)
		assert.Equal(t, "high", timeline[0].Importance)
		assert.Equal(t, "2026-01-13", timeline[0].WeekStart)
		assert.Equal(t, "2026-01-19", timeline[0].WeekEnd)
	})

	t.Run("returns empty for non-existent topic", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		timeline, err := repo.GetTopicTimeline(ctx, "NonExistent")
		require.NoError(t, err)
		assert.Empty(t, timeline)
	})

	t.Run("returns empty when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		timeline, err := repo.GetTopicTimeline(ctx, "GenAI")
		require.NoError(t, err)
		assert.Empty(t, timeline)
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

func TestInsightsRepository_GetWeeklyReport(t *testing.T) {
	ctx := context.Background()

	t.Run("returns full report for week with data", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		report, err := repo.GetWeeklyReport(ctx, "2026-01-27", "2026-02-03")
		require.NoError(t, err)
		require.NotNil(t, report)

		require.NotNil(t, report.Summary)
		assert.Equal(t, "2026-01-27", report.Summary.WeekStart)

		require.Len(t, report.Topics, 2)
		assert.Equal(t, "Quarterly Planning", report.Topics[0].Topic)
		assert.Equal(t, "Sprint Review", report.Topics[1].Topic)

		require.Len(t, report.InitiativeUpdates, 1)
		assert.Equal(t, "GenAI Integration", report.InitiativeUpdates[0].InitiativeName)
		assert.Equal(t, "AI integration sprint completed", report.InitiativeUpdates[0].UpdateText)

		require.Len(t, report.Actions, 2)
		assert.Equal(t, "high", report.Actions[0].Priority)
		assert.Equal(t, "low", report.Actions[1].Priority)
	})

	t.Run("returns nil report for week with no summary", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		report, err := repo.GetWeeklyReport(ctx, "2025-12-01", "2025-12-08")
		require.NoError(t, err)
		assert.Nil(t, report)
	})

	t.Run("returns nil when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)

		report, err := repo.GetWeeklyReport(ctx, "2026-01-27", "2026-02-03")
		require.NoError(t, err)
		assert.Nil(t, report)
	})

	t.Run("returns report with single topic and initiative update", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		report, err := repo.GetWeeklyReport(ctx, "2026-01-20", "2026-01-27")
		require.NoError(t, err)
		require.NotNil(t, report)

		assert.Equal(t, "2026-01-20", report.Summary.WeekStart)
		require.Len(t, report.Topics, 1)
		assert.Equal(t, "Tech Scorecard", report.Topics[0].Topic)

		require.Len(t, report.InitiativeUpdates, 1)
		assert.Equal(t, "Tech Scorecard", report.InitiativeUpdates[0].InitiativeName)
		assert.Equal(t, "Completed tech scorecard assessment", report.InitiativeUpdates[0].UpdateText)
	})
}

func TestInsightsRepository_GetDecisionsWithInitiatives(t *testing.T) {
	ctx := context.Background()

	t.Run("returns decisions with linked initiative names", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		decisions, err := repo.GetDecisionsWithInitiatives(ctx)
		require.NoError(t, err)
		require.Len(t, decisions, 2)

		assert.Equal(t, "Move to biweekly sprints", decisions[0].DecisionText)
		assert.Equal(t, "Q1 OKRs", decisions[0].Initiatives)

		assert.Equal(t, "Adopt Claude as primary AI provider", decisions[1].DecisionText)
		assert.Equal(t, "GenAI Integration", decisions[1].Initiatives)
	})

	t.Run("handles decisions with NULL nullable fields", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		_, err := db.Exec(`
			INSERT INTO decisions (decision_text, rationale, participants, expected_outcomes, decision_date, summary_id, created_at)
			VALUES ('Minimal decision', NULL, NULL, NULL, '2026-02-01', NULL, '2026-02-01 09:00:00')
		`)
		require.NoError(t, err)

		decisions, err := repo.GetDecisionsWithInitiatives(ctx)
		require.NoError(t, err)
		require.Len(t, decisions, 3)

		assert.Equal(t, "Minimal decision", decisions[0].DecisionText)
		assert.Empty(t, decisions[0].Rationale)
		assert.Empty(t, decisions[0].Participants)
		assert.Empty(t, decisions[0].ExpectedOutcomes)
		assert.Empty(t, decisions[0].Initiatives)
	})

	t.Run("returns empty slice when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)

		decisions, err := repo.GetDecisionsWithInitiatives(ctx)
		require.NoError(t, err)
		assert.Empty(t, decisions)
	})
}
