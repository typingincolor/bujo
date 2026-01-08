package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

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
				"o Meeting 1",
				"└── - note 1",
				"└── . Task 1",
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
				". Standalone task",
			},
			wantNotContain: []string{
				"└──",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderDailyAgenda(tt.agenda)

			for _, want := range tt.wantContains {
				assert.True(t, strings.Contains(result, want), "expected output to contain %q, got:\n%s", want, result)
			}

			for _, notWant := range tt.wantNotContain {
				assert.False(t, strings.Contains(result, notWant), "expected output NOT to contain %q, got:\n%s", notWant, result)
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

	result := RenderMultiDayAgenda(agenda)

	assert.Contains(t, result, "OVERDUE")
	assert.Contains(t, result, "o Meeting 1")
	assert.Contains(t, result, "└── - note 1")
}
