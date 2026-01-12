package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/typingincolor/bujo/internal/domain"
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
	parts := strings.Split(result, " ")

	// Leftmost (position 0) = 6 days ago = completed + selected (styled with ANSI)
	if !strings.Contains(parts[0], "●") {
		t.Errorf("Leftmost (6 days ago, completed) should contain ●, got %s", parts[0])
	}

	// Rightmost (position 6) = today = completed (not selected) = ●
	if parts[6] != "●" {
		t.Errorf("Rightmost (today, completed) should be ●, got %s", parts[6])
	}

	// Second from left (position 1) = 5 days ago = not completed = ○
	if parts[1] != "○" {
		t.Errorf("Second from left (5 days ago, empty) should be ○, got %s", parts[1])
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

			// Selected day should contain ○ with ANSI styling (longer string)
			if !strings.Contains(parts[tt.wantPosition], "○") {
				t.Errorf("Expected ○ at position %d (%s), but got: %s",
					tt.wantPosition, tt.description, parts[tt.wantPosition])
			}
			// Selected position should be styled (contains ANSI escape codes, so longer)
			if len(parts[tt.wantPosition]) <= 3 {
				t.Errorf("Position %d should be styled (longer than plain char), got: %s",
					tt.wantPosition, parts[tt.wantPosition])
			}

			// Other positions should show plain ○ (empty, not selected)
			for i, part := range parts {
				if i != tt.wantPosition && part != "○" {
					t.Errorf("Position %d should be plain ○, got %s", i, part)
				}
			}
		})
	}
}

func TestFlattenEntries_AnsweredQuestionsAutoExpanded(t *testing.T) {
	today := time.Now()
	questionEntityID := uuid.New().String()
	answerEntityID := uuid.New().String()

	questionID := int64(1)
	question := domain.Entry{
		ID:       questionID,
		EntityID: domain.EntityID(questionEntityID),
		Type:     domain.EntryTypeAnswered,
		Content:  "What is the deadline?",
	}
	answer := domain.Entry{
		ID:       2,
		EntityID: domain.EntityID(answerEntityID),
		Type:     domain.EntryTypeNote,
		Content:  "The deadline is next Friday",
		ParentID: &questionID,
	}

	entries := []domain.Entry{question, answer}

	m := Model{
		collapsed: make(map[domain.EntityID]bool),
	}

	items := m.flattenEntries(entries, "Test Day", false, today)

	// Answered question should have HasChildren = true
	if !items[0].HasChildren {
		t.Errorf("Expected answered question to have children")
	}

	// Answered question should be expanded by default (HiddenChildCount = 0)
	if items[0].HiddenChildCount != 0 {
		t.Errorf("Expected answered question to be expanded (HiddenChildCount=0), got %d", items[0].HiddenChildCount)
	}

	// Answer should be included in the items
	if len(items) != 2 {
		t.Errorf("Expected 2 items (question + answer), got %d", len(items))
	}

	if len(items) >= 2 && items[1].Entry.Content != "The deadline is next Friday" {
		t.Errorf("Expected second item to be the answer, got %s", items[1].Entry.Content)
	}
}

func TestFlattenEntries_UnansweredQuestionsCollapsedByDefault(t *testing.T) {
	today := time.Now()
	questionEntityID := uuid.New().String()
	noteEntityID := uuid.New().String()

	questionID := int64(1)
	question := domain.Entry{
		ID:       questionID,
		EntityID: domain.EntityID(questionEntityID),
		Type:     domain.EntryTypeQuestion,
		Content:  "What is the deadline?",
	}
	note := domain.Entry{
		ID:       2,
		EntityID: domain.EntityID(noteEntityID),
		Type:     domain.EntryTypeNote,
		Content:  "Some context",
		ParentID: &questionID,
	}

	entries := []domain.Entry{question, note}

	m := Model{
		collapsed: make(map[domain.EntityID]bool),
	}

	items := m.flattenEntries(entries, "Test Day", false, today)

	// Unanswered question with children should be collapsed by default
	if !items[0].HasChildren {
		t.Errorf("Expected question to have children")
	}

	// Unanswered question should be collapsed (HiddenChildCount > 0)
	if items[0].HiddenChildCount == 0 {
		t.Errorf("Expected unanswered question to be collapsed by default")
	}

	// Only the question should be visible (child is hidden)
	if len(items) != 1 {
		t.Errorf("Expected 1 item (question only, child hidden), got %d", len(items))
	}
}
