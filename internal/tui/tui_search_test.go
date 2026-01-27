package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func TestModel_DayView_CtrlS_EntersSearchMode(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
		{Entry: domain.Entry{ID: 2, Content: "Second task"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.searchMode.active {
		t.Error("searchMode should be active")
	}
	if !m.searchMode.forward {
		t.Error("searchMode should be forward")
	}
}

func TestModel_DayView_CtrlR_EntersReverseSearchMode(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
		{Entry: domain.Entry{ID: 2, Content: "Second task"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.searchMode.active {
		t.Error("searchMode should be active")
	}
	if m.searchMode.forward {
		t.Error("searchMode should be reverse (forward=false)")
	}
}

func TestModel_DayView_Search_TypingAddsToQuery(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.query != "test" {
		t.Errorf("expected query 'test', got '%s'", m.searchMode.query)
	}
}

func TestModel_DayView_Search_BackspaceRemovesFromQuery(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.query != "tes" {
		t.Errorf("expected query 'tes', got '%s'", m.searchMode.query)
	}
}

func TestModel_DayView_Search_SpaceAddsToQuery(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "my"}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.query != "my " {
		t.Errorf("expected query 'my ', got '%s'", m.searchMode.query)
	}
}

func TestModel_DayView_Search_EscCancelsSearch(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.active {
		t.Error("searchMode should not be active after Esc")
	}
	if m.searchMode.query != "" {
		t.Error("query should be cleared after Esc")
	}
}

func TestModel_DayView_Search_EnterExitsSearchMode(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First task"}},
		{Entry: domain.Entry{ID: 2, Content: "Second task"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "Second"}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchMode.active {
		t.Error("searchMode should not be active after Enter")
	}
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1, got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_IncrementalSearch_MovesToMatch(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	// Type "ban"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ban")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Banana), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_IncrementalSearch_StaysOnMatch(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry"}},
	}
	model.selectedIdx = 1 // Already on Banana
	model.searchMode = searchState{active: true, forward: true, query: "ban"}

	// Add more characters to refine search
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ana")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should stay on Banana since it still matches "banana"
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Banana), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_ForwardSearch_FindsNextMatch(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task one"}},
		{Entry: domain.Entry{ID: 2, Content: "Task two"}},
		{Entry: domain.Entry{ID: 3, Content: "Task three"}},
	}
	model.selectedIdx = 0 // On "Task one"
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// Press Ctrl+S to find next
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should move to "Task two"
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Task two), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_BackwardSearch_FindsPrevMatch(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task one"}},
		{Entry: domain.Entry{ID: 2, Content: "Task two"}},
		{Entry: domain.Entry{ID: 3, Content: "Task three"}},
	}
	model.selectedIdx = 2 // On "Task three"
	model.searchMode = searchState{active: true, forward: false, query: "Task"}

	// Press Ctrl+R to find previous
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should move to "Task two"
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Task two), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_DirectionSwitch_CtrlRThenCtrlS(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task one"}},
		{Entry: domain.Entry{ID: 2, Content: "Task two"}},
		{Entry: domain.Entry{ID: 3, Content: "Task three"}},
		{Entry: domain.Entry{ID: 4, Content: "Task four"}},
	}
	model.selectedIdx = 1 // On "Task two"
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// Press Ctrl+R to go backward
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("after Ctrl+R expected selectedIdx 0, got %d", m.selectedIdx)
	}
	if m.searchMode.forward {
		t.Error("forward should be false after Ctrl+R")
	}

	// Press Ctrl+S to go forward
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 1 {
		t.Errorf("after Ctrl+S expected selectedIdx 1, got %d", m.selectedIdx)
	}
	if !m.searchMode.forward {
		t.Error("forward should be true after Ctrl+S")
	}
}

func TestModel_DayView_Search_WrapsAround_Forward(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple here"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana there"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry time"}},
	}
	model.selectedIdx = 0 // On "Apple here"
	model.searchMode = searchState{active: true, forward: true, query: "Apple"}

	// Press Ctrl+S - should wrap around to beginning (only one match)
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should wrap to first item (same item - only match)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 after wrap, got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_WrapsAround_Backward(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Banana there"}},
		{Entry: domain.Entry{ID: 2, Content: "Cherry time"}},
		{Entry: domain.Entry{ID: 3, Content: "Apple here"}},
	}
	model.selectedIdx = 2 // On "Apple here"
	model.searchMode = searchState{active: true, forward: false, query: "Apple"}

	// Press Ctrl+R - should wrap around to end (only one match)
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should wrap to same item (only match)
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx 2 after wrap, got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_CaseInsensitive(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First item"}},
		{Entry: domain.Entry{ID: 2, Content: "SECOND ITEM"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("second")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (case insensitive match), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_ScrollFollowsSelection(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 10 // Small height to force scrolling

	// Create entries where match is beyond visible area
	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		content := fmt.Sprintf("Item %d", i)
		if i == 15 {
			content = "Special match"
		}
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}})
	}
	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Special")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 15 {
		t.Errorf("expected selectedIdx 15, got %d", m.selectedIdx)
	}
	// Scroll should have adjusted to show selected item
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed to show selected item")
	}
}

func TestModel_DayView_Search_NoMatch_StaysOnCurrent(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple"}},
		{Entry: domain.Entry{ID: 2, Content: "Banana"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("xyz")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should stay on current when no match
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 (no match), got %d", m.selectedIdx)
	}
}

func TestModel_DayView_Search_MultipleMatches_NextFindsDifferent(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task A"}},
		{Entry: domain.Entry{ID: 2, Content: "Note B"}},
		{Entry: domain.Entry{ID: 3, Content: "Task C"}},
		{Entry: domain.Entry{ID: 4, Content: "Note D"}},
		{Entry: domain.Entry{ID: 5, Content: "Task E"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// First Ctrl+S should go to Task C (index 2)
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 2 {
		t.Errorf("first Ctrl+S expected selectedIdx 2, got %d", m.selectedIdx)
	}

	// Second Ctrl+S should go to Task E (index 4)
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 4 {
		t.Errorf("second Ctrl+S expected selectedIdx 4, got %d", m.selectedIdx)
	}

	// Third Ctrl+S should wrap to Task A (index 0)
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("third Ctrl+S expected selectedIdx 0 (wrap), got %d", m.selectedIdx)
	}
}

// Capture Mode Search Tests - Direction Switching

// Week View Search Tests - Multiple days with headers

func TestModel_WeekView_Search_AcrossMultipleDays(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.currentView = ViewTypeReview
	model.viewMode = ViewModeWeek

	// Simulate entries from different days
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Monday task"}, DayHeader: "Monday, Jan 6"},
		{Entry: domain.Entry{ID: 2, Content: "Monday note"}},
		{Entry: domain.Entry{ID: 3, Content: "Tuesday task"}, DayHeader: "Tuesday, Jan 7"},
		{Entry: domain.Entry{ID: 4, Content: "Tuesday meeting"}},
		{Entry: domain.Entry{ID: 5, Content: "Wednesday task"}, DayHeader: "Wednesday, Jan 8"},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for "Tuesday"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Tuesday")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find "Tuesday task" at index 2
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx 2 (Tuesday task), got %d", m.selectedIdx)
	}

	// Press Ctrl+S to find next "Tuesday"
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Should find "Tuesday meeting" at index 3
	if m.selectedIdx != 3 {
		t.Errorf("expected selectedIdx 3 (Tuesday meeting), got %d", m.selectedIdx)
	}
}

func TestModel_WeekView_Search_WithDayHeaders_ScrollsCorrectly(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.currentView = ViewTypeReview
	model.viewMode = ViewModeWeek
	model.width = 80
	model.height = 12 // Small height to force scrolling

	// Create many entries across days
	entries := []EntryItem{}
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	id := int64(1)
	for _, day := range days {
		entries = append(entries, EntryItem{
			Entry:     domain.Entry{ID: id, Content: fmt.Sprintf("%s task 1", day)},
			DayHeader: fmt.Sprintf("%s, Jan", day),
		})
		id++
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: id, Content: fmt.Sprintf("%s task 2", day)}})
		id++
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: id, Content: fmt.Sprintf("%s task 3", day)}})
		id++
	}

	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for "Friday" which is far down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Friday")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find "Friday task 1" at index 12
	if m.selectedIdx != 12 {
		t.Errorf("expected selectedIdx 12 (Friday task 1), got %d", m.selectedIdx)
	}

	// Scroll should have adjusted
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed to show Friday entry")
	}
}

func TestModel_WeekView_Search_BackwardFromDifferentDay(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.currentView = ViewTypeReview
	model.viewMode = ViewModeWeek

	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Apple task"}, DayHeader: "Monday, Jan 6"},
		{Entry: domain.Entry{ID: 2, Content: "Banana task"}},
		{Entry: domain.Entry{ID: 3, Content: "Cherry task"}, DayHeader: "Tuesday, Jan 7"},
		{Entry: domain.Entry{ID: 4, Content: "Apple task"}}, // Another Apple on Tuesday
	}
	model.selectedIdx = 3 // On second "Apple task"
	model.searchMode = searchState{active: true, forward: false, query: "Apple"}

	// Press Ctrl+R to find previous Apple
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find first "Apple task" at index 0
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 (first Apple task), got %d", m.selectedIdx)
	}
}

func TestModel_WeekView_Search_NestedEntries(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.currentView = ViewTypeReview
	model.viewMode = ViewModeWeek

	// Entries with different indent levels (parent-child)
	parentID := int64(1)
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Project A"}, DayHeader: "Monday, Jan 6", Indent: 0},
		{Entry: domain.Entry{ID: 2, Content: "Subtask alpha", ParentID: &parentID}, Indent: 1},
		{Entry: domain.Entry{ID: 3, Content: "Subtask beta", ParentID: &parentID}, Indent: 1},
		{Entry: domain.Entry{ID: 4, Content: "Project B"}, DayHeader: "Tuesday, Jan 7", Indent: 0},
		{Entry: domain.Entry{ID: 5, Content: "Subtask alpha"}, Indent: 1}, // Another "alpha"
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for "alpha"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("alpha")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find first "Subtask alpha" at index 1
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (first Subtask alpha), got %d", m.selectedIdx)
	}

	// Press Ctrl+S to find next alpha
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Should find second "Subtask alpha" at index 4
	if m.selectedIdx != 4 {
		t.Errorf("expected selectedIdx 4 (second Subtask alpha), got %d", m.selectedIdx)
	}
}

func TestModel_WeekView_Search_WithOverdueEntries(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.currentView = ViewTypeReview
	model.viewMode = ViewModeWeek

	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Overdue task 1"}, DayHeader: "OVERDUE", IsOverdue: true},
		{Entry: domain.Entry{ID: 2, Content: "Overdue task 2"}, IsOverdue: true},
		{Entry: domain.Entry{ID: 3, Content: "Today task"}, DayHeader: "Monday, Jan 6"},
		{Entry: domain.Entry{ID: 4, Content: "Today task 2"}},
	}
	model.selectedIdx = 2 // On "Today task"
	model.searchMode = searchState{active: true, forward: false, query: "Overdue"}

	// Press Ctrl+R to find previous Overdue
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should find "Overdue task 2" at index 1
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 (Overdue task 2), got %d", m.selectedIdx)
	}
}

// Large data scrolling tests

func TestModel_Search_LargeData_ScrollsFromTopToBottom(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 20 // Can show ~14 entries

	// Create 100 entries
	entries := []EntryItem{}
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf("Item number %d", i)
		if i == 95 {
			content = "TARGET ITEM HERE"
		}
		entry := EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}}
		if i%10 == 0 {
			entry.DayHeader = fmt.Sprintf("Day %d", i/10)
		}
		entries = append(entries, entry)
	}

	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true}

	// Search for TARGET which is at index 95
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("TARGET")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 95 {
		t.Errorf("expected selectedIdx 95, got %d", m.selectedIdx)
	}

	// Verify scroll offset is valid (selectedIdx should be visible)
	if m.scrollOffset > 95 {
		t.Errorf("scrollOffset %d is too high, selectedIdx 95 won't be visible", m.scrollOffset)
	}
	if m.scrollOffset+14 < 95 {
		t.Errorf("scrollOffset %d is too low, selectedIdx 95 won't be visible", m.scrollOffset)
	}
}

func TestModel_Search_LargeData_ScrollsFromBottomToTop(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 20

	// Create 100 entries
	entries := []EntryItem{}
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf("Item number %d", i)
		if i == 5 {
			content = "TARGET ITEM HERE"
		}
		entry := EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}}
		if i%10 == 0 {
			entry.DayHeader = fmt.Sprintf("Day %d", i/10)
		}
		entries = append(entries, entry)
	}

	model.entries = entries
	model.selectedIdx = 95
	model.scrollOffset = 85 // Scrolled to bottom
	model.searchMode = searchState{active: true, forward: false, query: "TARGET"}

	// Search backward for TARGET which is at index 5
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 5 {
		t.Errorf("expected selectedIdx 5, got %d", m.selectedIdx)
	}

	// Verify scroll offset adjusted to show entry
	if m.scrollOffset > 5 {
		t.Errorf("scrollOffset %d is too high, selectedIdx 5 won't be visible", m.scrollOffset)
	}
}

func TestModel_Search_LargeData_MultipleSearchesTraverseAll(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 15

	// Create 50 entries with "task" in every 5th entry
	entries := []EntryItem{}
	expectedTaskIndices := []int{4, 9, 14, 19, 24, 29, 34, 39, 44, 49}
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf("Item %d", i)
		if i%5 == 4 {
			content = fmt.Sprintf("Task item %d", i)
		}
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}})
	}

	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0
	model.searchMode = searchState{active: true, forward: true, query: "Task"}

	// Press Ctrl+S repeatedly to find all tasks
	for i, expectedIdx := range expectedTaskIndices {
		msg := tea.KeyMsg{Type: tea.KeyCtrlS}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.selectedIdx != expectedIdx {
			t.Errorf("iteration %d: expected selectedIdx %d, got %d", i, expectedIdx, m.selectedIdx)
		}

		// Verify selectedIdx is visible (between scrollOffset and scrollOffset + visible area)
		if m.selectedIdx < m.scrollOffset {
			t.Errorf("iteration %d: selectedIdx %d is above scrollOffset %d", i, m.selectedIdx, m.scrollOffset)
		}

		model = m
	}

	// One more Ctrl+S should wrap to first task
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 4 {
		t.Errorf("after wrap: expected selectedIdx 4, got %d", m.selectedIdx)
	}
}

func TestModel_Search_LargeData_DirectionSwitchMidList(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 15

	// Create entries with matches at various positions
	entries := []EntryItem{}
	matchIndices := []int{10, 25, 40, 55, 70}
	for i := 0; i < 80; i++ {
		content := fmt.Sprintf("Item %d", i)
		for _, matchIdx := range matchIndices {
			if i == matchIdx {
				content = fmt.Sprintf("MATCH at %d", i)
			}
		}
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: content}})
	}

	model.entries = entries
	model.selectedIdx = 40 // Middle match
	model.scrollOffset = 35
	model.searchMode = searchState{active: true, forward: true, query: "MATCH"}

	// Go forward: should find 55
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 55 {
		t.Errorf("forward from 40: expected 55, got %d", m.selectedIdx)
	}

	// Switch direction (Ctrl+R): should find 40
	msg = tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 40 {
		t.Errorf("backward from 55: expected 40, got %d", m.selectedIdx)
	}

	// Continue backward: should find 25
	msg = tea.KeyMsg{Type: tea.KeyCtrlR}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 25 {
		t.Errorf("backward from 40: expected 25, got %d", m.selectedIdx)
	}

	// Switch to forward (Ctrl+S): should find 40
	msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.selectedIdx != 40 {
		t.Errorf("forward from 25: expected 40, got %d", m.selectedIdx)
	}

	// Verify scroll followed through all these jumps
	if m.scrollOffset > 40 || m.scrollOffset+14 < 40 {
		t.Errorf("final scrollOffset %d doesn't make selectedIdx 40 visible", m.scrollOffset)
	}
}

// Navigation Tests - Top and Bottom

func TestModel_Navigation_TopKey_GoesToFirstEntry(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First"}},
		{Entry: domain.Entry{ID: 2, Content: "Second"}},
		{Entry: domain.Entry{ID: 3, Content: "Third"}},
	}
	model.selectedIdx = 2
	model.scrollOffset = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0, got %d", m.selectedIdx)
	}
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0, got %d", m.scrollOffset)
	}
}

func TestModel_Navigation_BottomKey_GoesToLastEntry(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 20
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "First"}},
		{Entry: domain.Entry{ID: 2, Content: "Second"}},
		{Entry: domain.Entry{ID: 3, Content: "Third"}},
	}
	model.selectedIdx = 0
	model.scrollOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx 2, got %d", m.selectedIdx)
	}
}

func TestModel_Navigation_BottomKey_WithLargeList_ScrollsCorrectly(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 15 // Can show ~9 entries

	entries := []EntryItem{}
	for i := 0; i < 30; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.selectedIdx != 29 {
		t.Errorf("expected selectedIdx 29, got %d", m.selectedIdx)
	}
	// Scroll should have adjusted to show last entry
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed")
	}
}

// Done/Complete Tests

func TestModel_Done_MarksTaskComplete(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "Task to complete"}},
	}
	model.selectedIdx = 0

	// Space key marks as done
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should have triggered a command (the actual mark done happens via service)
	if cmd == nil {
		t.Error("expected a command to be returned")
	}
	_ = m // Model state is updated async via the command
}

func TestModel_Done_NoOpOnEmptyList(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("expected no command when list is empty")
	}
}

// Delete Tests

func TestModel_Delete_TriggersConfirmMode(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item to delete"}},
	}
	model.selectedIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should activate confirm mode synchronously
	if !m.confirmMode.active {
		t.Error("expected confirmMode to be active")
	}
	if m.confirmMode.entryID != 1 {
		t.Errorf("expected entryID 1, got %d", m.confirmMode.entryID)
	}
}

func TestModel_Delete_NoOpOnEmptyList(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("expected no command when list is empty")
	}
}

// Goto Date Tests

func TestModel_GotoMode_EntersOnSlash(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item"}},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.gotoMode.active {
		t.Error("gotoMode should be active")
	}
}

func TestModel_GotoMode_TypingAddsToInput(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.gotoMode = gotoState{active: true}
	model.gotoMode.input = createTextInput()
	model.gotoMode.input.Focus()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tomorrow")}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.gotoMode.input.Value() != "tomorrow" {
		t.Errorf("expected input 'tomorrow', got '%s'", m.gotoMode.input.Value())
	}
}

// Capture Mode Arrow Key Tests

// View Rendering Tests

func TestModel_View_SearchMode_ShowsSearchBar(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item one"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	view := model.View()

	if !strings.Contains(view, "Search") {
		t.Error("view should contain Search bar")
	}
	if !strings.Contains(view, "forward") {
		t.Error("view should show search direction")
	}
}

func TestModel_View_SearchMode_ShowsReverseDirection(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item one"}},
	}
	model.searchMode = searchState{active: true, forward: false, query: "test"}

	view := model.View()

	if !strings.Contains(view, "reverse") {
		t.Error("view should show reverse direction")
	}
}

func TestModel_View_SelectedEntry_IsHighlighted(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "First item"}},
		{Entry: domain.Entry{ID: 2, Type: domain.EntryTypeTask, Content: "Selected item"}},
	}
	model.selectedIdx = 1

	view := model.View()

	// The selected item should be present (exact styling depends on lipgloss)
	if !strings.Contains(view, "Selected item") {
		t.Error("view should contain the selected item text")
	}
}

func TestModel_View_DayHeader_IsRendered(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Task"}, DayHeader: "Monday, Jan 6"},
	}

	view := model.View()

	if !strings.Contains(view, "Monday") {
		t.Error("view should contain day header")
	}
}

func TestModel_View_OverdueHeader_IsRendered(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Overdue task"}, DayHeader: "OVERDUE", IsOverdue: true},
	}

	view := model.View()

	if !strings.Contains(view, "OVERDUE") {
		t.Error("view should contain OVERDUE header")
	}
}

func TestModel_View_EntrySymbols_AreRendered(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "Task"}},
		{Entry: domain.Entry{ID: 2, Type: domain.EntryTypeNote, Content: "Note"}},
		{Entry: domain.Entry{ID: 3, Type: domain.EntryTypeEvent, Content: "Event"}},
		{Entry: domain.Entry{ID: 4, Type: domain.EntryTypeDone, Content: "Done"}},
	}

	view := model.View()

	// Check Unicode symbols are present
	if !strings.Contains(view, "•") {
		t.Error("view should contain task symbol •")
	}
	if !strings.Contains(view, "–") {
		t.Error("view should contain note symbol –")
	}
	if !strings.Contains(view, "○") {
		t.Error("view should contain event symbol ○")
	}
	if !strings.Contains(view, "✓") {
		t.Error("view should contain done symbol ✓")
	}
}

func TestModel_View_ScrollIndicators_ShowMoreAbove(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 10 // Small height

	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 10
	model.scrollOffset = 5 // Scrolled down

	view := model.View()

	if !strings.Contains(view, "more above") {
		t.Error("view should show 'more above' indicator")
	}
}

func TestModel_View_ScrollIndicators_ShowMoreBelow(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 10 // Small height

	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 0
	model.scrollOffset = 0

	view := model.View()

	if !strings.Contains(view, "more below") {
		t.Error("view should show 'more below' indicator")
	}
}

func TestModel_View_Toolbar_ShowsJournalView(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal
	model.viewMode = ViewModeDay
	model.entries = []EntryItem{}

	view := model.View()

	if !strings.Contains(view, "Journal") {
		t.Error("view should show Journal in toolbar")
	}
}

func TestModel_View_Toolbar_ShowsReviewView(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeReview
	model.viewMode = ViewModeWeek
	model.entries = []EntryItem{}

	view := model.View()

	if !strings.Contains(view, "Review") {
		t.Error("view should show Review in toolbar")
	}
}

// Confirm Mode Tests

func TestModel_ConfirmMode_YConfirmsDelete(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item to delete"}},
	}
	model.confirmMode = confirmState{
		active:      true,
		entryID:     1,
		hasChildren: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := model.Update(msg)

	// Should return a delete command
	if cmd == nil {
		t.Error("expected a delete command")
	}
}

func TestModel_ConfirmMode_NCancels(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Content: "Item to delete"}},
	}
	model.confirmMode = confirmState{
		active:      true,
		entryID:     1,
		hasChildren: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.confirmMode.active {
		t.Error("confirmMode should not be active after pressing n")
	}
	if cmd != nil {
		t.Error("expected no command after cancel")
	}
}

// ensuredVisible Tests

func TestModel_EnsuredVisible_ScrollsUpWhenSelectedAbove(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 15

	entries := []EntryItem{}
	for i := 0; i < 30; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 2
	model.scrollOffset = 10 // Selected is above visible area

	m := model.ensuredVisible()

	if m.scrollOffset > 2 {
		t.Errorf("scrollOffset should be <= 2, got %d", m.scrollOffset)
	}
}

func TestModel_EnsuredVisible_ScrollsDownWhenSelectedBelow(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 15

	entries := []EntryItem{}
	for i := 0; i < 30; i++ {
		entries = append(entries, EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}})
	}
	model.entries = entries
	model.selectedIdx = 25
	model.scrollOffset = 0 // Selected is below visible area

	m := model.ensuredVisible()

	// Should have scrolled down to show item 25
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed")
	}
}

func TestModel_EnsuredVisible_WithDayHeaders_AccountsForExtraLines(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 12

	entries := []EntryItem{}
	for i := 0; i < 20; i++ {
		entry := EntryItem{Entry: domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Item %d", i)}}
		if i%5 == 0 {
			entry.DayHeader = fmt.Sprintf("Day %d", i/5)
		}
		entries = append(entries, entry)
	}
	model.entries = entries
	model.selectedIdx = 15
	model.scrollOffset = 0

	m := model.ensuredVisible()

	// Should have scrolled to show item 15
	// With headers, each header takes extra line, so scroll should adjust
	if m.scrollOffset == 0 {
		t.Error("scrollOffset should have changed to show item 15 with headers")
	}
}

// Paste Tests

// Unicode Symbol Conversion Tests

// Search Highlighting Tests

func TestModel_HighlightSearchTerm_HighlightsMatch(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	// Result should be different from input (contains ANSI codes)
	if result == line {
		t.Error("highlighted result should differ from original (should contain ANSI codes)")
	}
	// Original text should still be present
	if !strings.Contains(result, "this is a ") {
		t.Error("result should contain text before match")
	}
	if !strings.Contains(result, " line") {
		t.Error("result should contain text after match")
	}
}

func TestModel_HighlightSearchTerm_EmptyQuery_NoChange(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: ""}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	if result != line {
		t.Errorf("empty query should return unchanged line, got '%s'", result)
	}
}

func TestModel_HighlightSearchTerm_NoMatch_NoChange(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "xyz"}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	if result != line {
		t.Errorf("no match should return unchanged line, got '%s'", result)
	}
}

func TestModel_HighlightSearchTerm_CaseInsensitive(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "TEST"}

	line := "this is a test line"
	result := model.highlightSearchTerm(line)

	// Should highlight even though case differs
	if result == line {
		t.Error("case-insensitive match should be highlighted")
	}
}

func TestModel_HighlightSearchTerm_MultipleMatches(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	line := "test one and test two and test three"
	result := model.highlightSearchTerm(line)

	// Count how many times the original "test" appears vs the highlighted version
	// The original string has 3 "test" instances
	// After highlighting, each "test" should be wrapped in ANSI codes
	originalCount := strings.Count(line, "test")
	if originalCount != 3 {
		t.Errorf("test setup error: expected 3 matches in original, got %d", originalCount)
	}

	// The result should be longer due to ANSI codes
	if len(result) <= len(line) {
		t.Error("highlighted result should be longer than original due to ANSI codes")
	}
}

func TestModel_HighlightSearchTerm_PreservesNonMatchingCase(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "test"}

	line := "TEST and Test and test"
	result := model.highlightSearchTerm(line)

	// All three should be highlighted, but original case preserved in output
	// The ANSI codes will wrap each match
	if result == line {
		t.Error("all matches should be highlighted")
	}
	// Check that we can still find the original text patterns
	// (they'll be wrapped in ANSI codes but the text is there)
}

func TestModel_View_SearchHighlighting_AppliedToEntries(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "First item"}},
		{Entry: domain.Entry{ID: 2, Type: domain.EntryTypeTask, Content: "Second searchterm item"}},
		{Entry: domain.Entry{ID: 3, Type: domain.EntryTypeTask, Content: "Third item"}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "searchterm"}

	view := model.View()

	// The view should contain the search term (possibly with ANSI codes)
	if !strings.Contains(view, "searchterm") && !strings.Contains(view, "Second") {
		t.Error("view should contain the entry with search term")
	}

	// View without search mode
	model.searchMode = searchState{}
	viewNoSearch := model.View()

	// The non-search view should be shorter (no ANSI highlighting codes)
	// Actually both contain the text, but with search active, there are extra ANSI codes
	if !strings.Contains(viewNoSearch, "Second searchterm item") {
		t.Error("view without search should contain full entry text")
	}
}

func TestModel_View_SearchHighlighting_InSelectedEntry(t *testing.T) {
	model := New(nil)
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: 1, Type: domain.EntryTypeTask, Content: "Match here"}},
	}
	model.selectedIdx = 0
	model.searchMode = searchState{active: true, forward: true, query: "Match"}

	view := model.View()

	// The selected entry with a search match should render
	// Both selection styling and search highlighting should be applied
	if !strings.Contains(view, "Match") {
		t.Error("view should contain the matched text")
	}
}

func TestModel_HighlightSearchTerm_SpecialCharacters(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "•"}

	line := "• Task item"
	result := model.highlightSearchTerm(line)

	// Should highlight the bullet point
	if result == line {
		t.Error("special character should be highlighted")
	}
}

func TestModel_HighlightSearchTerm_PartialWord(t *testing.T) {
	model := New(nil)
	model.searchMode = searchState{active: true, forward: true, query: "ask"}

	line := "• Task item"
	result := model.highlightSearchTerm(line)

	// Should highlight "ask" within "Task"
	if result == line {
		t.Error("partial word match should be highlighted")
	}
}

func TestModel_SearchMode_ShowsAncestryForSelectedEntry(t *testing.T) {
	parent1ID := int64(1)
	parent2ID := int64(2)
	childID := int64(3)

	model := New(nil)
	model.width = 80
	model.height = 24
	model.entries = []EntryItem{
		{Entry: domain.Entry{ID: parent1ID, Content: "Project A", ParentID: nil}},
		{Entry: domain.Entry{ID: parent2ID, Content: "Phase 1", ParentID: &parent1ID}},
		{Entry: domain.Entry{ID: childID, Content: "Task detail", ParentID: &parent2ID}},
	}
	model.searchMode = searchState{active: true, forward: true, query: "detail"}
	model.selectedIdx = 2

	view := model.View()

	if !strings.Contains(view, "Project A") || !strings.Contains(view, "Phase 1") {
		t.Error("expected ancestry chain to be shown when in search mode")
	}
}

// ============================================================================
// Phase 4: Multi-View Architecture Tests
// ============================================================================
