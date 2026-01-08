package domain

import (
	"errors"
	"time"
)

type DayContext struct {
	EntityID EntityID
	Date     time.Time
	Location *string
	Mood     *string
	Weather  *string
}

func (c DayContext) Validate() error {
	if c.Date.IsZero() {
		return errors.New("date is required")
	}
	if c.Location != nil && *c.Location == "" {
		return errors.New("location cannot be empty string")
	}
	return nil
}

func (c DayContext) HasLocation() bool {
	return c.Location != nil
}

func (c DayContext) GetLocation() string {
	if c.Location == nil {
		return ""
	}
	return *c.Location
}
