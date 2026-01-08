package domain

import (
	"errors"
	"time"
)

type SummaryHorizon string

const (
	SummaryHorizonDaily     SummaryHorizon = "daily"
	SummaryHorizonWeekly    SummaryHorizon = "weekly"
	SummaryHorizonQuarterly SummaryHorizon = "quarterly"
	SummaryHorizonAnnual    SummaryHorizon = "annual"
)

var validHorizons = map[SummaryHorizon]bool{
	SummaryHorizonDaily:     true,
	SummaryHorizonWeekly:    true,
	SummaryHorizonQuarterly: true,
	SummaryHorizonAnnual:    true,
}

func (h SummaryHorizon) IsValid() bool {
	return validHorizons[h]
}

type Summary struct {
	ID        int64
	EntityID  EntityID
	Horizon   SummaryHorizon
	Content   string
	StartDate time.Time
	EndDate   time.Time
	CreatedAt time.Time
}

func (s Summary) PeriodLength() int {
	days := s.EndDate.Sub(s.StartDate).Hours() / 24
	return int(days) + 1
}

func (s Summary) IsRecent(now time.Time) bool {
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	createdDate := time.Date(s.CreatedAt.Year(), s.CreatedAt.Month(), s.CreatedAt.Day(), 0, 0, 0, 0, s.CreatedAt.Location())
	return nowDate.Equal(createdDate)
}

func (s Summary) Validate() error {
	if !s.Horizon.IsValid() {
		return errors.New("invalid summary horizon")
	}
	if s.Content == "" {
		return errors.New("summary content cannot be empty")
	}
	if s.StartDate.IsZero() {
		return errors.New("start date is required")
	}
	if s.EndDate.IsZero() {
		return errors.New("end date is required")
	}
	if s.EndDate.Before(s.StartDate) {
		return errors.New("end date cannot be before start date")
	}
	return nil
}
