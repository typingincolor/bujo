package domain

import (
	"errors"
	"time"
)

type GoalStatus string

const (
	GoalStatusActive GoalStatus = "active"
	GoalStatusDone   GoalStatus = "done"
)

type Goal struct {
	ID        int64
	EntityID  EntityID
	Content   string
	Month     time.Time
	Status    GoalStatus
	CreatedAt time.Time
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

func (g Goal) MonthKey() string {
	return g.Month.Format("2006-01")
}
