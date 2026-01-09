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

func TestRenderDailyAgenda_OverdueHierarchy(t *testing.T) {
	today := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	parentID := int64(1)

	tests := []struct {
		name           string
		agenda         *service.DailyAgenda
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "overdue entries preserve parent-child hierarchy",
			agenda: &service.DailyAgenda{
				Date: today,
				Overdue: []domain.Entry{
					{ID: 1, Type: domain.EntryTypeEvent, Content: "Meeting 1", ScheduledDate: &yesterday, Depth: 0},
					{ID: 2, Type: domain.EntryTypeNote, Content: "note 1", ParentID: &parentID, ScheduledDate: &yesterday, Depth: 1},
					{ID: 3, Type: domain.EntryTypeTask, Content: "Task 1", ParentID: func() *int64 { id := int64(2); return &id }(), ScheduledDate: &yesterday, Depth: 2},
				},
				Today: []domain.Entry{},
			},
			wantContains: []string{
				"OVERDUE",
				"○ Meeting 1",
				"  – note 1",
				"    • Task 1",
			},
			wantNotContain: []string{},
		},
		{
			name: "overdue root entries without children render without tree prefix",
			agenda: &service.DailyAgenda{
				Date: today,
				Overdue: []domain.Entry{
					{ID: 1, Type: domain.EntryTypeTask, Content: "Standalone task", ScheduledDate: &yesterday, Depth: 0},
				},
				Today: []domain.Entry{},
			},
			wantContains: []string{
				"OVERDUE",
				"• Standalone task",
			},
			wantNotContain: []string{
				"└──",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderDailyAgenda(tt.agenda)
			stripped := stripANSI(result)

			for _, want := range tt.wantContains {
				assert.True(t, strings.Contains(stripped, want), "expected output to contain %q, got:\n%s", want, stripped)
			}

			for _, notWant := range tt.wantNotContain {
				assert.False(t, strings.Contains(stripped, notWant), "expected output NOT to contain %q, got:\n%s", notWant, stripped)
			}
		})
	}
}

func TestRenderMultiDayAgenda_OverdueHierarchy(t *testing.T) {
	today := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	parentID := int64(1)

	agenda := &service.MultiDayAgenda{
		Overdue: []domain.Entry{
			{ID: 1, Type: domain.EntryTypeEvent, Content: "Meeting 1", ScheduledDate: &yesterday, Depth: 0},
			{ID: 2, Type: domain.EntryTypeNote, Content: "note 1", ParentID: &parentID, ScheduledDate: &yesterday, Depth: 1},
		},
		Days: []service.DayEntries{
			{Date: today, Entries: []domain.Entry{}},
		},
	}

	result := RenderMultiDayAgenda(agenda, today)
	stripped := stripANSI(result)

	assert.Contains(t, stripped, "OVERDUE")
	assert.Contains(t, stripped, "○ Meeting 1")
	assert.Contains(t, stripped, "  – note 1")
}

func TestRenderMultiDayAgenda_OverdueTasksInDaySectionAreRed(t *testing.T) {
	today := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	parentID := int64(1)

	// Task from yesterday shown in yesterday's day section (within week view)
	// should be styled red because it's overdue relative to today
	agenda := &service.MultiDayAgenda{
		Overdue: []domain.Entry{}, // nothing in overdue section
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
