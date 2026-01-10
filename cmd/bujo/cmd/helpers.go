package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
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

func isPureNumber(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func parseListNameOrID(s string) (name string, id int64, isID bool, err error) {
	if strings.HasPrefix(s, "#") {
		id, err = strconv.ParseInt(s[1:], 10, 64)
		if err != nil {
			return "", 0, false, fmt.Errorf("invalid list ID: %s", s)
		}
		return "", id, true, nil
	}
	return s, 0, false, nil
}

func resolveListID(ctx context.Context, s string) (int64, error) {
	name, id, isID, err := parseListNameOrID(s)
	if err != nil {
		return 0, err
	}
	if isID {
		return id, nil
	}
	list, err := listService.GetListByName(ctx, name)
	if err != nil {
		return 0, err
	}
	return list.ID, nil
}

func parsePastDate(s string) (time.Time, error) {
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

func parseDateOrToday(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	return parsePastDate(s)
}

func validateDateRange(from, to time.Time) error {
	if from.After(to) {
		return fmt.Errorf("--from date must be before --to date")
	}
	return nil
}

func parseAddArgs(args []string) (entries []string, location, date, file string, help, yes bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-a" || arg == "--at":
			if i+1 < len(args) {
				location = args[i+1]
				i++
			}
		case arg == "-d" || arg == "--date":
			if i+1 < len(args) {
				date = args[i+1]
				i++
			}
		case arg == "-f" || arg == "--file":
			if i+1 < len(args) {
				file = args[i+1]
				i++
			}
		case arg == "-y" || arg == "--yes":
			yes = true
		case strings.HasPrefix(arg, "-a="):
			location = arg[3:]
		case strings.HasPrefix(arg, "--at="):
			location = arg[5:]
		case strings.HasPrefix(arg, "-d="):
			date = arg[3:]
		case strings.HasPrefix(arg, "--date="):
			date = arg[7:]
		case strings.HasPrefix(arg, "-f="):
			file = arg[3:]
		case strings.HasPrefix(arg, "--file="):
			file = arg[7:]
		case arg == "-h" || arg == "--help":
			help = true
			return
		case arg == "--":
			entries = append(entries, args[i+1:]...)
			return
		case arg == "--db-path":
			if i+1 < len(args) {
				i++
			}
		case strings.HasPrefix(arg, "--db-path="):
			// skip
		case arg == "-v" || arg == "--verbose":
			// skip
		default:
			entries = append(entries, arg)
		}
	}
	return
}

func isNaturalLanguageDate(s string) bool {
	if s == "" {
		return false
	}
	// Check if it's ISO format (2006-01-02)
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return false
	}
	// Check if it's compact format (20060102)
	if _, err := time.Parse("20060102", s); err == nil {
		return false
	}
	// If not a recognized date format, it's natural language
	return true
}

func confirmDate(dateStr string, parsed time.Time, skipConfirm bool) (time.Time, error) {
	if !isNaturalLanguageDate(dateStr) || skipConfirm {
		return parsed, nil
	}

	formatted := parsed.Format("Monday, Jan 2, 2006")
	fmt.Fprintf(os.Stderr, "Using date: %s [Y/n]: ", formatted)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	response := strings.TrimSpace(strings.ToLower(input))

	if response == "" || response == "y" || response == "yes" {
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("cancelled")
}

func parseFutureDate(s string) (time.Time, error) {
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
