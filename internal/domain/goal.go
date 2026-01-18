package domain

import (
	"errors"
	"time"
)

type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusDone      GoalStatus = "done"
	GoalStatusMigrated  GoalStatus = "migrated"
	GoalStatusCancelled GoalStatus = "cancelled"
)

type Goal struct {
	ID         int64
	EntityID   EntityID
	Content    string
	Month      time.Time
	Status     GoalStatus
	MigratedTo *time.Time
	CreatedAt  time.Time
}

func (g Goal) Validate() error {
	if g.Content == "" {
		return errors.New("goal content cannot be empty")
	}
	if g.Month.IsZero() {
		return errors.New("goal month is required")
	}
	return nil
}

func (g Goal) IsDone() bool {
	return g.Status == GoalStatusDone
}

func (g Goal) MarkDone() Goal {
	g.Status = GoalStatusDone
	return g
}

func (g Goal) MarkActive() Goal {
	g.Status = GoalStatusActive
	return g
}

func (g Goal) IsMigrated() bool {
	return g.Status == GoalStatusMigrated
}

func (g Goal) MarkMigrated(toMonth time.Time) Goal {
	g.Status = GoalStatusMigrated
	g.MigratedTo = &toMonth
	return g
}

func (g Goal) MonthKey() string {
	return g.Month.Format("2006-01")
}

func (g Goal) UpdateContent(content string) Goal {
	g.Content = content
	return g
}

func (g Goal) IsCancelled() bool {
	return g.Status == GoalStatusCancelled
}

func (g Goal) MarkCancelled() Goal {
	g.Status = GoalStatusCancelled
	return g
}
