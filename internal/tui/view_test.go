package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/typingincolor/bujo/internal/service"
)

func TestRenderSparkline_OrderMatchesLabels(t *testing.T) {
	// Setup: Create a model with habit history
	// History is ordered [0]=today, [1]=yesterday, etc.
	// Visual should be: oldest (left) -> today (right)
	history := []service.DayStatus{
		{Date: time.Now(), Completed: true, Count: 1},                        // today - should be rightmost
		{Date: time.Now().AddDate(0, 0, -1), Completed: false, Count: 0},     // yesterday
		{Date: time.Now().AddDate(0, 0, -2), Completed: true, Count: 1},      // 2 days ago
		{Date: time.Now().AddDate(0, 0, -3), Completed: false, Count: 0},     // 3 days ago
		{Date: time.Now().AddDate(0, 0, -4), Completed: true, Count: 1},      // 4 days ago
		{Date: time.Now().AddDate(0, 0, -5), Completed: false, Count: 0},     // 5 days ago
		{Date: time.Now().AddDate(0, 0, -6), Completed: true, Count: 1},      // 6 days ago - should be leftmost
	}

	m := Model{
		habitState: habitState{
			selectedIdx:    0,
			selectedDayIdx: 0, // leftmost (oldest = 6 days ago)
		},
	}

	result := m.renderSparkline(history, true)

	// Parse the result - split by space, ignoring the selection brackets
	parts := strings.Split(result, " ")

	// First element (leftmost) should be 6 days ago (history[6]) = completed (●)
	// Last element (rightmost) should be today (history[0]) = completed (●)
	// The pattern should be: ● ○ ● ○ ● ○ ● (6 days ago to today)
	// With history[6]=completed, history[5]=not, history[4]=completed, etc.

	// Leftmost should show history[6] = completed
	if !strings.Contains(parts[0], "●") {
		t.Errorf("Leftmost (6 days ago) should be completed (●), got %s", parts[0])
	}

	// Rightmost should show history[0] = completed
	if !strings.Contains(parts[6], "●") {
		t.Errorf("Rightmost (today) should be completed (●), got %s", parts[6])
	}

	// Second from left should show history[5] = not completed
	if !strings.Contains(parts[1], "○") {
		t.Errorf("Second from left (5 days ago) should be not completed (○), got %s", parts[1])
	}
}

func TestRenderSparkline_SelectionHighlightsCorrectDay(t *testing.T) {
	history := make([]service.DayStatus, 7)
	for i := range history {
		history[i] = service.DayStatus{
			Date:      time.Now().AddDate(0, 0, -i),
			Completed: false,
			Count:     0,
		}
	}

	tests := []struct {
		name           string
		selectedDayIdx int
		wantPosition   int // 0=leftmost, 6=rightmost
		description    string
	}{
		{
			name:           "selectedDayIdx=0 highlights leftmost (oldest)",
			selectedDayIdx: 0,
			wantPosition:   0,
			description:    "oldest day (6 days ago)",
		},
		{
			name:           "selectedDayIdx=6 highlights rightmost (today)",
			selectedDayIdx: 6,
			wantPosition:   6,
			description:    "today",
		},
		{
			name:           "selectedDayIdx=3 highlights middle",
			selectedDayIdx: 3,
			wantPosition:   3,
			description:    "3 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				habitState: habitState{
					selectedIdx:    0,
					selectedDayIdx: tt.selectedDayIdx,
				},
			}

			result := m.renderSparkline(history, true)
			parts := strings.Split(result, " ")

			// Check that the selection bracket is at the expected position
			if !strings.Contains(parts[tt.wantPosition], "[") {
				t.Errorf("Expected selection at position %d (%s), but got: %v",
					tt.wantPosition, tt.description, parts)
			}

			// Check that other positions don't have brackets
			for i, part := range parts {
				if i != tt.wantPosition && strings.Contains(part, "[") {
					t.Errorf("Unexpected selection at position %d, expected only at %d",
						i, tt.wantPosition)
				}
			}
		})
	}
}
