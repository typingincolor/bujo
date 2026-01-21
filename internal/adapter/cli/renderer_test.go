package cli

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func init() {
	// Enable color output for tests (normally disabled when not a TTY)
	color.NoColor = false
}

// stripANSI removes ANSI escape codes from a string for content verification
func stripANSI(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

func TestRenderMultiDayAgenda_OverdueTasksInDaySectionAreRed(t *testing.T) {
	today := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	parentID := int64(1)

	// Task from yesterday shown in yesterday's day section (within week view)
	// should be styled red because it's overdue relative to today
	agenda := &service.MultiDayAgenda{
		Days: []service.DayEntries{
			{
				Date: yesterday,
				Entries: []domain.Entry{
					{ID: 1, Type: domain.EntryTypeEvent, Content: "Meeting", ScheduledDate: &yesterday, Depth: 0},
					{ID: 2, Type: domain.EntryTypeTask, Content: "Overdue task", ParentID: &parentID, ScheduledDate: &yesterday, Depth: 1},
				},
			},
			{
				Date: today,
				Entries: []domain.Entry{
					{ID: 3, Type: domain.EntryTypeTask, Content: "Today task", ScheduledDate: &today, Depth: 0},
				},
			},
		},
	}

	result := RenderMultiDayAgenda(agenda, today)

	// The overdue task in yesterday's section should be red
	// We check that Red() was applied to "Overdue task" by looking for ANSI codes
	// Red color code is \x1b[31m
	assert.Contains(t, result, "Meeting")      // Event not red (events don't go overdue)
	assert.Contains(t, result, "Overdue task") // Task present
	assert.Contains(t, result, "Today task")   // Today's task present

	// Find the line containing "Overdue task" and check it has red ANSI code
	lines := strings.Split(result, "\n")
	var overdueTaskLine, todayTaskLine string
	for _, line := range lines {
		if strings.Contains(line, "Overdue task") {
			overdueTaskLine = line
		}
		if strings.Contains(line, "Today task") {
			todayTaskLine = line
		}
	}

	// \x1b[31m is the ANSI code for red
	assert.True(t, strings.Contains(overdueTaskLine, "\x1b[31m"),
		"Overdue task should be styled red, got: %q", overdueTaskLine)
	assert.False(t, strings.Contains(todayTaskLine, "\x1b[31m"),
		"Today's task should NOT be red, got: %q", todayTaskLine)
}

func TestRenderDailyAgenda_CancelledEntriesHaveStrikethrough(t *testing.T) {
	today := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)

	agenda := &service.DailyAgenda{
		Date: today,
		Today: []domain.Entry{
			{ID: 1, Type: domain.EntryTypeCancelled, Content: "Cancelled task", ScheduledDate: &today, Depth: 0},
			{ID: 2, Type: domain.EntryTypeTask, Content: "Normal task", ScheduledDate: &today, Depth: 0},
		},
	}

	result := RenderDailyAgenda(agenda)

	// Find lines
	lines := strings.Split(result, "\n")
	var cancelledLine, normalLine string
	for _, line := range lines {
		if strings.Contains(line, "Cancelled task") {
			cancelledLine = line
		}
		if strings.Contains(line, "Normal task") {
			normalLine = line
		}
	}

	// \x1b[9m is the ANSI code for strikethrough
	assert.True(t, strings.Contains(cancelledLine, "\x1b[9m"),
		"Cancelled task should have strikethrough, got: %q", cancelledLine)
	assert.False(t, strings.Contains(normalLine, "\x1b[9m"),
		"Normal task should NOT have strikethrough, got: %q", normalLine)
}

func TestRenderGoalsSection_ShowsGoalsWithProgress(t *testing.T) {
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	goals := []domain.Goal{
		{ID: 1, Content: "Learn Go", Status: domain.GoalStatusActive},
		{ID: 2, Content: "Read more books", Status: domain.GoalStatusDone},
		{ID: 3, Content: "Exercise daily", Status: domain.GoalStatusActive},
	}

	result := RenderGoalsSection(goals, month)
	stripped := stripANSI(result)

	assert.Contains(t, stripped, "January Goals")
	assert.Contains(t, stripped, "Learn Go")
	assert.Contains(t, stripped, "Read more books")
	assert.Contains(t, stripped, "Exercise daily")
	assert.Contains(t, stripped, "33%") // 1 out of 3 done
}

func TestRenderGoalsSection_EmptyWhenNoGoals(t *testing.T) {
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	result := RenderGoalsSection([]domain.Goal{}, month)

	assert.Empty(t, result)
}

func TestRenderGoalsSection_ShowsCheckmarkForDoneGoals(t *testing.T) {
	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	goals := []domain.Goal{
		{ID: 1, Content: "Completed goal", Status: domain.GoalStatusDone},
	}

	result := RenderGoalsSection(goals, month)
	stripped := stripANSI(result)

	assert.Contains(t, stripped, "✓")
	assert.Contains(t, stripped, "Completed goal")
	assert.Contains(t, stripped, "100%")
}

func TestRenderDailyAgenda_WithMoodAndWeather(t *testing.T) {
	today := time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC)
	location := "Home Office"
	mood := "Focused"
	weather := "Sunny"

	agenda := &service.DailyAgenda{
		Date:     today,
		Location: &location,
		Mood:     &mood,
		Weather:  &weather,
		Today:    []domain.Entry{},
	}

	result := RenderDailyAgenda(agenda)
	stripped := stripANSI(result)

	assert.Contains(t, stripped, "Home Office")
	assert.Contains(t, stripped, "Focused")
	assert.Contains(t, stripped, "Sunny")
}

func TestRenderMultiDayAgenda_WithMoodAndWeather(t *testing.T) {
	day1 := time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	today := day1

	location1 := "Home"
	mood1 := "Energetic"
	weather1 := "Cloudy"

	location2 := "Office"
	mood2 := "Calm"
	weather2 := "Rainy"

	agenda := &service.MultiDayAgenda{
		Days: []service.DayEntries{
			{
				Date:     day1,
				Location: &location1,
				Mood:     &mood1,
				Weather:  &weather1,
				Entries:  []domain.Entry{},
			},
			{
				Date:     day2,
				Location: &location2,
				Mood:     &mood2,
				Weather:  &weather2,
				Entries:  []domain.Entry{},
			},
		},
	}

	result := RenderMultiDayAgenda(agenda, today)
	stripped := stripANSI(result)

	assert.Contains(t, stripped, "Home")
	assert.Contains(t, stripped, "Energetic")
	assert.Contains(t, stripped, "Cloudy")

	assert.Contains(t, stripped, "Office")
	assert.Contains(t, stripped, "Calm")
	assert.Contains(t, stripped, "Rainy")
}
