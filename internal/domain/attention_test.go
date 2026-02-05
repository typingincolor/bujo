package domain

import (
	"testing"
	"time"
)

func TestCalculateAttentionScore(t *testing.T) {
	now := time.Date(2026, 1, 27, 12, 0, 0, 0, time.UTC)

	newEntry := func(opts ...func(*Entry)) Entry {
		e := Entry{
			Type:      EntryTypeTask,
			Content:   "Test task",
			Priority:  PriorityNone,
			CreatedAt: now,
		}
		for _, opt := range opts {
			opt(&e)
		}
		return e
	}

	withCreatedAt := func(t time.Time) func(*Entry) {
		return func(e *Entry) { e.CreatedAt = t }
	}
	withScheduledDate := func(t time.Time) func(*Entry) {
		return func(e *Entry) { e.ScheduledDate = &t }
	}
	withPriority := func(p Priority) func(*Entry) {
		return func(e *Entry) { e.Priority = p }
	}
	withContent := func(c string) func(*Entry) {
		return func(e *Entry) { e.Content = c }
	}
	withType := func(et EntryType) func(*Entry) {
		return func(e *Entry) { e.Type = et }
	}
	withParent := func(id int64) func(*Entry) {
		return func(e *Entry) { e.ParentID = &id }
	}

	t.Run("returns zero score for plain task", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(), now, "")
		if result.Score != 0 {
			t.Errorf("expected score 0, got %d", result.Score)
		}
		if len(result.Indicators) != 0 {
			t.Errorf("expected no indicators, got %v", result.Indicators)
		}
	})

	t.Run("adds 50 points for past scheduled date", func(t *testing.T) {
		yesterday := now.AddDate(0, 0, -1)
		result := CalculateAttentionScore(newEntry(withScheduledDate(yesterday)), now, "")
		if result.Score < 50 {
			t.Errorf("expected score >= 50, got %d", result.Score)
		}
		assertContainsIndicator(t, result.Indicators, AttentionOverdue)
	})

	t.Run("no overdue points for future scheduled date", func(t *testing.T) {
		tomorrow := now.AddDate(0, 0, 1)
		result := CalculateAttentionScore(newEntry(withScheduledDate(tomorrow)), now, "")
		assertNotContainsIndicator(t, result.Indicators, AttentionOverdue)
	})

	t.Run("adds 30 points for low priority", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withPriority(PriorityLow)), now, "")
		if result.Score != 30 {
			t.Errorf("expected score 30, got %d", result.Score)
		}
		assertContainsIndicator(t, result.Indicators, AttentionPriority)
	})

	t.Run("adds 30 points for medium priority", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withPriority(PriorityMedium)), now, "")
		if result.Score != 30 {
			t.Errorf("expected score 30, got %d", result.Score)
		}
	})

	t.Run("adds 50 points for high priority", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withPriority(PriorityHigh)), now, "")
		if result.Score != 50 {
			t.Errorf("expected score 50 (30+20), got %d", result.Score)
		}
		assertContainsIndicator(t, result.Indicators, AttentionPriority)
	})

	t.Run("adds 15 points for items 3-7 days old", func(t *testing.T) {
		fourDaysAgo := now.AddDate(0, 0, -4)
		result := CalculateAttentionScore(newEntry(withCreatedAt(fourDaysAgo)), now, "")
		if result.Score != 15 {
			t.Errorf("expected score 15, got %d", result.Score)
		}
		assertContainsIndicator(t, result.Indicators, AttentionAging)
	})

	t.Run("adds 25 points for items older than 7 days", func(t *testing.T) {
		eightDaysAgo := now.AddDate(0, 0, -8)
		result := CalculateAttentionScore(newEntry(withCreatedAt(eightDaysAgo)), now, "")
		if result.Score != 25 {
			t.Errorf("expected score 25, got %d", result.Score)
		}
		assertContainsIndicator(t, result.Indicators, AttentionAging)
	})

	t.Run("no aging points for items 3 days old or less", func(t *testing.T) {
		twoDaysAgo := now.AddDate(0, 0, -2)
		result := CalculateAttentionScore(newEntry(withCreatedAt(twoDaysAgo)), now, "")
		if result.Score != 0 {
			t.Errorf("expected score 0, got %d", result.Score)
		}
	})

	t.Run("adds 20 points for urgent keyword", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withContent("This is urgent!")), now, "")
		if result.Score != 20 {
			t.Errorf("expected score 20, got %d", result.Score)
		}
	})

	t.Run("adds 20 points for asap keyword", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withContent("Need this ASAP")), now, "")
		if result.Score != 20 {
			t.Errorf("expected score 20, got %d", result.Score)
		}
	})

	t.Run("adds 20 points for blocker keyword", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withContent("This is a blocker")), now, "")
		if result.Score != 20 {
			t.Errorf("expected score 20, got %d", result.Score)
		}
	})

	t.Run("adds 20 points for waiting keyword", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withContent("waiting on response")), now, "")
		if result.Score != 20 {
			t.Errorf("expected score 20, got %d", result.Score)
		}
	})

	t.Run("adds 20 points for blocked keyword", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withContent("blocked by team")), now, "")
		if result.Score != 20 {
			t.Errorf("expected score 20, got %d", result.Score)
		}
	})

	t.Run("keyword matching is case insensitive", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withContent("URGENT meeting")), now, "")
		if result.Score != 20 {
			t.Errorf("expected score 20, got %d", result.Score)
		}
	})

	t.Run("adds 10 points for questions", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withType(EntryTypeQuestion)), now, "")
		if result.Score != 10 {
			t.Errorf("expected score 10, got %d", result.Score)
		}
	})

	t.Run("adds 5 points when parent is event", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withParent(1)), now, EntryTypeEvent)
		if result.Score != 5 {
			t.Errorf("expected score 5, got %d", result.Score)
		}
	})

	t.Run("no parent event points without parent", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(), now, EntryTypeEvent)
		if result.Score != 0 {
			t.Errorf("expected score 0, got %d", result.Score)
		}
	})

	t.Run("no parent event points when parent is not event", func(t *testing.T) {
		result := CalculateAttentionScore(newEntry(withParent(1)), now, EntryTypeTask)
		if result.Score != 0 {
			t.Errorf("expected score 0, got %d", result.Score)
		}
	})

	t.Run("combines multiple conditions", func(t *testing.T) {
		fourDaysAgo := now.AddDate(0, 0, -4)
		result := CalculateAttentionScore(
			newEntry(withPriority(PriorityHigh), withCreatedAt(fourDaysAgo)),
			now, "",
		)
		// 30 (priority) + 20 (high) + 15 (age) = 65
		if result.Score != 65 {
			t.Errorf("expected score 65, got %d", result.Score)
		}
	})

	t.Run("returns days old in result", func(t *testing.T) {
		fourDaysAgo := now.AddDate(0, 0, -4)
		result := CalculateAttentionScore(newEntry(withCreatedAt(fourDaysAgo)), now, "")
		if result.DaysOld != 4 {
			t.Errorf("expected daysOld 4, got %d", result.DaysOld)
		}
	})
}

func assertContainsIndicator(t *testing.T, indicators []AttentionIndicator, want AttentionIndicator) {
	t.Helper()
	for _, ind := range indicators {
		if ind == want {
			return
		}
	}
	t.Errorf("expected indicators to contain %s, got %v", want, indicators)
}

func assertNotContainsIndicator(t *testing.T, indicators []AttentionIndicator, unwanted AttentionIndicator) {
	t.Helper()
	for _, ind := range indicators {
		if ind == unwanted {
			t.Errorf("expected indicators not to contain %s, got %v", unwanted, indicators)
			return
		}
	}
}
