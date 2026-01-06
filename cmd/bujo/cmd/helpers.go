package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tj/go-naturaldate"
)

func parseEntryID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid entry ID: %s", s)
	}
	return id, nil
}

func parseHabitNameOrID(s string) (name string, id int64, isID bool, err error) {
	if strings.HasPrefix(s, "#") {
		id, err = strconv.ParseInt(s[1:], 10, 64)
		if err != nil {
			return "", 0, false, fmt.Errorf("invalid habit ID: %s", s)
		}
		return "", id, true, nil
	}
	return s, 0, false, nil
}

func parsePastDate(s string) (time.Time, error) {
	now := time.Now()

	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	parsed, err := naturaldate.Parse(s, now, naturaldate.WithDirection(naturaldate.Past))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}

func parseFutureDate(s string) (time.Time, error) {
	now := time.Now()

	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return parsed, nil
	}

	parsed, err := naturaldate.Parse(s, now, naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", s)
	}

	return parsed, nil
}
