package domain

import (
	"strings"
	"time"
)

type AttentionIndicator string

const (
	AttentionOverdue  AttentionIndicator = "overdue"
	AttentionPriority AttentionIndicator = "priority"
	AttentionAging    AttentionIndicator = "aging"
	AttentionMigrated AttentionIndicator = "migrated"
)

type AttentionResult struct {
	Score      int
	Indicators []AttentionIndicator
	DaysOld    int
}

const (
	attentionScoreOverdue       = 50
	attentionScorePriority      = 30
	attentionScoreHighPriority  = 20
	attentionScoreAgingOld      = 25
	attentionScoreAgingRecent   = 15
	attentionScoreUrgentKeyword = 20
	attentionScoreQuestion      = 10
	attentionScoreParentEvent   = 5
	attentionScoreMigration     = 15
	attentionAgingThresholdOld  = 7
	attentionAgingThresholdNew  = 3
)

var urgentKeywords = []string{"urgent", "asap", "blocker", "waiting", "blocked"}

func CalculateAttentionScore(entry Entry, now time.Time, parentType EntryType) AttentionResult {
	score := 0
	var indicators []AttentionIndicator

	if entry.ScheduledDate != nil && entry.ScheduledDate.Before(now) {
		score += attentionScoreOverdue
		indicators = append(indicators, AttentionOverdue)
	}

	if entry.Priority != PriorityNone && entry.Priority != "" {
		score += attentionScorePriority
		indicators = append(indicators, AttentionPriority)
		if entry.Priority == PriorityHigh {
			score += attentionScoreHighPriority
		}
	}

	daysOld := int(now.Sub(entry.CreatedAt).Hours() / 24)

	if daysOld > attentionAgingThresholdOld {
		score += attentionScoreAgingOld
		indicators = append(indicators, AttentionAging)
	} else if daysOld > attentionAgingThresholdNew {
		score += attentionScoreAgingRecent
		indicators = append(indicators, AttentionAging)
	}

	contentLower := strings.ToLower(entry.Content)
	for _, keyword := range urgentKeywords {
		if strings.Contains(contentLower, keyword) {
			score += attentionScoreUrgentKeyword
			break
		}
	}

	if entry.Type == EntryTypeQuestion {
		score += attentionScoreQuestion
	}

	if entry.MigrationCount > 0 {
		score += entry.MigrationCount * attentionScoreMigration
		indicators = append(indicators, AttentionMigrated)
	}

	if entry.ParentID != nil && parentType == EntryTypeEvent {
		score += attentionScoreParentEvent
	}

	return AttentionResult{
		Score:      score,
		Indicators: indicators,
		DaysOld:    daysOld,
	}
}
