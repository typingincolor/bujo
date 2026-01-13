package dateutil

import (
	"fmt"
	"time"

	"github.com/tj/go-naturaldate"
)

func ParsePast(s string) (time.Time, error) {
	now := time.Now()

	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	if parsed, err := time.Parse("20060102", s); err == nil {
		return parsed, nil
	}

	parsed, err := naturaldate.Parse(s, now, naturaldate.WithDirection(naturaldate.Past))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}

func ParseFuture(s string) (time.Time, error) {
	now := time.Now()

	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	if parsed, err := time.Parse("20060102", s); err == nil {
		return parsed, nil
	}

	parsed, err := naturaldate.Parse(s, now, naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}
