package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGoal_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		goal    Goal
		wantErr bool
	}{
		{
			name: "valid goal",
			goal: Goal{
				Content:   "Learn Go",
				Month:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "empty content is invalid",
			goal: Goal{
				Content:   "",
				Month:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "zero month is invalid",
			goal: Goal{
				Content:   "Learn Go",
				Month:     time.Time{},
				CreatedAt: now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.goal.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGoal_IsDone(t *testing.T) {
	goal := Goal{
		Content: "Learn Go",
		Status:  GoalStatusActive,
	}
	assert.False(t, goal.IsDone())

	goal.Status = GoalStatusDone
	assert.True(t, goal.IsDone())
}

func TestGoal_MarkDone(t *testing.T) {
	entityID := NewEntityID()
	original := Goal{
		ID:       1,
		EntityID: entityID,
		Content:  "Learn Go",
		Status:   GoalStatusActive,
		Month:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	updated := original.MarkDone()

	assert.Equal(t, GoalStatusDone, updated.Status)
	assert.Equal(t, GoalStatusActive, original.Status, "original should be unchanged")
	assert.Equal(t, original.ID, updated.ID, "other fields should be copied")
	assert.Equal(t, original.EntityID, updated.EntityID, "other fields should be copied")
	assert.Equal(t, original.Content, updated.Content, "other fields should be copied")
	assert.Equal(t, original.Month, updated.Month, "other fields should be copied")
}

func TestGoal_MarkActive(t *testing.T) {
	entityID := NewEntityID()
	original := Goal{
		ID:       1,
		EntityID: entityID,
		Content:  "Learn Go",
		Status:   GoalStatusDone,
		Month:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	updated := original.MarkActive()

	assert.Equal(t, GoalStatusActive, updated.Status)
	assert.Equal(t, GoalStatusDone, original.Status, "original should be unchanged")
	assert.Equal(t, original.ID, updated.ID, "other fields should be copied")
	assert.Equal(t, original.EntityID, updated.EntityID, "other fields should be copied")
	assert.Equal(t, original.Content, updated.Content, "other fields should be copied")
	assert.Equal(t, original.Month, updated.Month, "other fields should be copied")
}

func TestGoal_MonthKey(t *testing.T) {
	goal := Goal{
		Month: time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	key := goal.MonthKey()

	assert.Equal(t, "2026-01", key)
}

func TestGoal_IsMigrated(t *testing.T) {
	goal := Goal{
		Content: "Learn Go",
		Status:  GoalStatusActive,
	}
	assert.False(t, goal.IsMigrated())

	goal.Status = GoalStatusMigrated
	assert.True(t, goal.IsMigrated())
}

func TestGoal_MarkMigrated(t *testing.T) {
	entityID := NewEntityID()
	original := Goal{
		ID:       1,
		EntityID: entityID,
		Content:  "Learn Go",
		Status:   GoalStatusActive,
		Month:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	targetMonth := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	updated := original.MarkMigrated(targetMonth)

	assert.Equal(t, GoalStatusMigrated, updated.Status)
	assert.Equal(t, GoalStatusActive, original.Status, "original should be unchanged")
	assert.Equal(t, original.ID, updated.ID, "other fields should be copied")
	assert.Equal(t, original.EntityID, updated.EntityID, "other fields should be copied")
	assert.Equal(t, original.Content, updated.Content, "other fields should be copied")
	assert.Equal(t, original.Month, updated.Month, "original month should be preserved")
	assert.NotNil(t, updated.MigratedTo)
	assert.Equal(t, "2026-02", updated.MigratedTo.Format("2006-01"))
}

func TestGoal_UpdateContent(t *testing.T) {
	entityID := NewEntityID()
	original := Goal{
		ID:       1,
		EntityID: entityID,
		Content:  "Learn Go",
		Status:   GoalStatusActive,
		Month:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	updated := original.UpdateContent("Learn Go and Rust")

	assert.Equal(t, "Learn Go and Rust", updated.Content)
	assert.Equal(t, "Learn Go", original.Content, "original should be unchanged")
	assert.Equal(t, original.ID, updated.ID, "other fields should be copied")
	assert.Equal(t, original.EntityID, updated.EntityID, "other fields should be copied")
	assert.Equal(t, original.Status, updated.Status, "other fields should be copied")
	assert.Equal(t, original.Month, updated.Month, "other fields should be copied")
}

func TestGoal_IsCancelled(t *testing.T) {
	goal := Goal{
		Content: "Learn Go",
		Status:  GoalStatusActive,
	}
	assert.False(t, goal.IsCancelled())

	goal.Status = GoalStatusCancelled
	assert.True(t, goal.IsCancelled())
}

func TestGoal_MarkCancelled(t *testing.T) {
	entityID := NewEntityID()
	original := Goal{
		ID:       1,
		EntityID: entityID,
		Content:  "Learn Go",
		Status:   GoalStatusActive,
		Month:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	updated := original.MarkCancelled()

	assert.Equal(t, GoalStatusCancelled, updated.Status)
	assert.Equal(t, GoalStatusActive, original.Status, "original should be unchanged")
	assert.Equal(t, original.ID, updated.ID, "other fields should be copied")
	assert.Equal(t, original.EntityID, updated.EntityID, "other fields should be copied")
	assert.Equal(t, original.Content, updated.Content, "other fields should be copied")
	assert.Equal(t, original.Month, updated.Month, "other fields should be copied")
}
