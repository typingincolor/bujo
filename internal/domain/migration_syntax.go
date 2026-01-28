package domain

import (
	"errors"
	"strings"
	"time"
)

func ParseMigrationSyntax(line string, dateParser func(string) (time.Time, error)) (string, *time.Time, error) {
	if !strings.HasPrefix(line, ">[") {
		return line, nil, nil
	}

	closeBracket := strings.Index(line, "]")
	if closeBracket == -1 {
		return "", nil, errors.New("missing closing bracket in migration syntax")
	}

	dateStr := line[2:closeBracket]
	if dateStr == "" {
		return "", nil, errors.New("empty date in migration syntax")
	}

	if dateParser == nil {
		return "", nil, errors.New("date parser required for migration syntax")
	}

	parsedDate, err := dateParser(dateStr)
	if err != nil {
		return "", nil, errors.New("Cannot parse date")
	}

	content := strings.TrimSpace(line[closeBracket+1:])

	return content, &parsedDate, nil
}
