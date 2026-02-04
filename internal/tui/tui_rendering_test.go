package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
	"github.com/typingincolor/bujo/internal/testutil"
)

func TestTheme_Default_HasAllColors(t *testing.T) {
	theme := DefaultTheme
	if !theme.HasAllColors() {
		t.Error("default theme should have all colors defined")
	}
}

func TestTheme_Dark_HasAllColors(t *testing.T) {
	theme := DarkTheme
	if !theme.HasAllColors() {
		t.Error("dark theme should have all colors defined")
	}
}

func TestTheme_Light_HasAllColors(t *testing.T) {
	theme := LightTheme
	if !theme.HasAllColors() {
		t.Error("light theme should have all colors defined")
	}
}

func TestTheme_Solarized_HasAllColors(t *testing.T) {
	theme := SolarizedTheme
	if !theme.HasAllColors() {
		t.Error("solarized theme should have all colors defined")
	}
}

func TestTheme_GetTheme_ReturnsCorrectTheme(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"default", "default"},
		{"dark", "dark"},
		{"light", "light"},
		{"solarized", "solarized"},
	}

	for _, tt := range tests {
		theme := GetTheme(tt.name)
		if theme.Name != tt.expected {
			t.Errorf("GetTheme(%s) returned theme with name %s, expected %s", tt.name, theme.Name, tt.expected)
		}
	}
}

func TestTheme_GetTheme_InvalidTheme_ReturnsDefault(t *testing.T) {
	theme := GetTheme("nonexistent")
	if theme.Name != "default" {
		t.Errorf("GetTheme with invalid name should return default theme, got %s", theme.Name)
	}
}

func TestTheme_AvailableThemes_ReturnsList(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) != 4 {
		t.Errorf("expected 4 themes, got %d", len(themes))
	}

	expected := map[string]bool{"default": true, "dark": true, "light": true, "solarized": true}
	for _, theme := range themes {
		if !expected[theme] {
			t.Errorf("unexpected theme %s", theme)
		}
	}
}

func TestTheme_NewThemeStyles_CreatesStyles(t *testing.T) {
	theme := DefaultTheme
	styles := NewThemeStyles(theme)

	// Check that styles are not empty (have some rendering capability)
	// We can verify by checking that Render produces non-empty output
	testStr := "test"
	if styles.Toolbar.Render(testStr) == "" {
		t.Error("Toolbar style should render text")
	}
	if styles.Header.Render(testStr) == "" {
		t.Error("Header style should render text")
	}
	if styles.Done.Render(testStr) == "" {
		t.Error("Done style should render text")
	}
	if styles.Selected.Render(testStr) == "" {
		t.Error("Selected style should render text")
	}
}

// Config Tests

func TestConfig_DefaultTUIConfig_ReturnsDefaults(t *testing.T) {
	config := DefaultTUIConfig()

	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
	if config.Theme != "default" {
		t.Errorf("expected default theme 'default', got '%s'", config.Theme)
	}
	if config.DateFormat != "Mon, Jan 2 2006" {
		t.Errorf("expected default date format, got '%s'", config.DateFormat)
	}
	if !config.ShowHelp {
		t.Error("expected ShowHelp to be true by default")
	}
}

func TestConfig_LoadTUIConfigFromPath_LoadsValidYAML(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	content := []byte("default_view: habits\ntheme: dark\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	config := LoadTUIConfigFromPath(configPath)

	if config.DefaultView != "habits" {
		t.Errorf("expected 'habits', got '%s'", config.DefaultView)
	}
	if config.Theme != "dark" {
		t.Errorf("expected 'dark', got '%s'", config.Theme)
	}
}

func TestConfig_LoadTUIConfigFromPath_PartialFile_UsesDefaults(t *testing.T) {
	// Create a temporary config file with only theme
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	content := []byte("theme: solarized\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	config := LoadTUIConfigFromPath(configPath)

	if config.Theme != "solarized" {
		t.Errorf("expected 'solarized', got '%s'", config.Theme)
	}
	// Default view should use default
	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
}

func TestConfig_LoadTUIConfigFromPath_NoFile_UsesDefaults(t *testing.T) {
	config := LoadTUIConfigFromPath("/nonexistent/path/config.yaml")

	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
	if config.Theme != "default" {
		t.Errorf("expected default theme 'default', got '%s'", config.Theme)
	}
}

func TestConfig_LoadTUIConfigFromPath_InvalidYAML_UsesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	content := []byte("invalid: yaml: content: [")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	config := LoadTUIConfigFromPath(configPath)

	if config.DefaultView != "journal" {
		t.Errorf("expected default view 'journal', got '%s'", config.DefaultView)
	}
}

func TestConfig_GetViewType_ReturnsCorrectType(t *testing.T) {
	tests := []struct {
		defaultView string
		expected    ViewType
	}{
		{"journal", ViewTypeJournal},
		{"habits", ViewTypeHabits},
		{"lists", ViewTypeLists},
		{"unknown", ViewTypeJournal},
	}

	for _, tt := range tests {
		config := TUIConfig{DefaultView: tt.defaultView}
		if config.GetViewType() != tt.expected {
			t.Errorf("GetViewType() for '%s' expected %v, got %v", tt.defaultView, tt.expected, config.GetViewType())
		}
	}
}

func TestConfig_ConfigPaths_ReturnsMultiplePaths(t *testing.T) {
	paths := ConfigPaths()
	if len(paths) < 1 {
		t.Error("ConfigPaths should return at least one path")
	}

	// Check that paths end with expected suffixes
	foundConfigDir := false
	foundBujoDir := false
	for _, p := range paths {
		if strings.Contains(p, "bujo/config.yaml") || strings.Contains(p, "bujo\\config.yaml") {
			foundConfigDir = true
		}
		if strings.Contains(p, ".bujo/config.yaml") || strings.Contains(p, ".bujo\\config.yaml") {
			foundBujoDir = true
		}
	}

	if !foundConfigDir && !foundBujoDir {
		t.Errorf("ConfigPaths should include standard config paths, got: %v", paths)
	}
}

func TestRenderEntry_SelectedMigratedEntry_HasReadableForeground(t *testing.T) {
	model := New(nil)
	item := EntryItem{
		Entry: domain.Entry{
			Type:    domain.EntryTypeMigrated,
			Content: "Migrated task",
		},
	}

	rendered := model.renderEntry(item, true)

	migratedFgCode := "\x1b[38;5;8m"

	if strings.Contains(rendered, migratedFgCode) {
		t.Error("selected migrated entry should NOT have gray foreground (color 8) - would be unreadable against gray background")
	}
	if !strings.Contains(rendered, "Migrated task") {
		t.Error("rendered output should contain the task content")
	}
}

func TestRenderEntry_UnselectedMigratedEntry_HasDimStyle(t *testing.T) {
	model := New(nil)
	item := EntryItem{
		Entry: domain.Entry{
			Type:    domain.EntryTypeMigrated,
			Content: "Migrated task",
		},
	}

	rendered := model.renderEntry(item, false)

	if !strings.Contains(rendered, "Migrated task") {
		t.Error("rendered output should contain the task content")
	}
	if !strings.Contains(rendered, "→") {
		t.Error("migrated entry should show the migrated symbol")
	}
}

func TestRemoveHabitLogForDateCmd_NoLogsToRemove_ShouldNotReturnError(t *testing.T) {
	bujoSvc, habitSvc, listSvc, goalSvc := setupTestServices(t)
	ctx := context.Background()

	today := time.Now()
	err := habitSvc.LogHabitForDate(ctx, "Meditation", 1, today)
	if err != nil {
		t.Fatalf("failed to create habit: %v", err)
	}

	status, err := habitSvc.GetTrackerStatus(ctx, today, 7)
	if err != nil {
		t.Fatalf("failed to get tracker status: %v", err)
	}
	habitID := status.Habits[0].ID

	model := NewWithConfig(Config{
		BujoService:  bujoSvc,
		HabitService: habitSvc,
		ListService:  listSvc,
		GoalService:  goalSvc,
	})

	twoDaysAgo := today.AddDate(0, 0, -2)
	cmd := model.removeHabitLogForDateCmd(habitID, twoDaysAgo)
	msg := cmd()

	if _, isError := msg.(errMsg); isError {
		t.Error("removeHabitLogForDateCmd should not return errMsg when no logs exist for the date")
	}
}

func TestHabitView_WeekOffset_DefaultsToZero(t *testing.T) {
	model := New(nil)
	if model.habitState.weekOffset != 0 {
		t.Errorf("weekOffset should default to 0, got %d", model.habitState.weekOffset)
	}
}

func TestHabitView_PrevPeriod_IncrementsWeekOffset(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.weekOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.habitState.weekOffset != 1 {
		t.Errorf("pressing '[' should increment weekOffset to 1, got %d", m.habitState.weekOffset)
	}
}

func TestHabitView_NextPeriod_DecrementsWeekOffset(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.weekOffset = 2

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.habitState.weekOffset != 1 {
		t.Errorf("pressing ']' should decrement weekOffset to 1, got %d", m.habitState.weekOffset)
	}
}

func TestHabitView_NextPeriod_CannotGoToFuture(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.weekOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.habitState.weekOffset != 0 {
		t.Errorf("pressing ']' at weekOffset=0 should not go negative, got %d", m.habitState.weekOffset)
	}
}

func TestHabitView_ToggleView_ResetsWeekOffsetToAvoidFuture(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.habitState.viewMode = HabitViewModeWeek
	model.habitState.weekOffset = 0

	// Navigate back 2 weeks in week mode
	model.habitState.weekOffset = 2

	// Now press 'w' to toggle to month mode
	// This should reset weekOffset to 0
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// After toggle, weekOffset should be 0
	if m.habitState.weekOffset != 0 {
		t.Errorf("after toggling view mode, weekOffset should be reset to 0, got %d", m.habitState.weekOffset)
	}

	// Reference date should be current day or recent
	refDateMonth := m.getHabitReferenceDate()
	todayNormalized := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	refDateNormalized := time.Date(refDateMonth.Year(), refDateMonth.Month(), refDateMonth.Day(), 0, 0, 0, 0, refDateMonth.Location())
	daysOffset := todayNormalized.Sub(refDateNormalized).Hours() / 24

	// Allow ~35 day margin: since month-mode offset of 1 = ~30 days, this ensures
	// we're within approximately one month of the current date after the toggle
	const maxDaysMargin = 35
	if daysOffset > maxDaysMargin {
		t.Errorf("after toggling to month mode, reference date should be recent (within ~30 days), but is %d days in past", int(daysOffset))
	}
}

func TestHabitView_DayLabels_WeekView_ShowsDayNames(t *testing.T) {
	model := New(nil)
	model.habitState.viewMode = HabitViewModeWeek

	labels := model.renderDayLabels(HabitDaysWeek)

	// Week view should show day-of-week letters
	for _, dayLetter := range []string{"S", "M", "T", "W", "F"} {
		if !strings.Contains(labels, dayLetter) {
			t.Errorf("week view labels should contain '%s', got: %s", dayLetter, labels)
		}
	}
}

func TestHabitView_DayLabels_MonthView_ShowsDateMarkers(t *testing.T) {
	model := New(nil)
	model.habitState.viewMode = HabitViewModeMonth

	labels := model.renderDayLabels(HabitDaysMonth)

	// Month view should NOT show day-of-week letters
	if strings.Count(labels, "S") > 2 || strings.Count(labels, "M") > 2 {
		t.Errorf("month view labels should show dates not day letters, got: %s", labels)
	}

	// Should contain numeric date markers
	hasNumber := false
	for _, r := range labels {
		if r >= '0' && r <= '9' {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		t.Errorf("month view labels should contain date numbers, got: %s", labels)
	}
}

func TestHabitView_DayLabels_QuarterView_ShowsMonthMarkers(t *testing.T) {
	model := New(nil)
	model.habitState.viewMode = HabitViewModeQuarter

	labels := model.renderDayLabels(HabitDaysQuarter)

	// Quarter view should show month abbreviations
	monthAbbrevs := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	hasMonth := false
	for _, abbrev := range monthAbbrevs {
		if strings.Contains(labels, abbrev) {
			hasMonth = true
			break
		}
	}
	if !hasMonth {
		t.Errorf("quarter view labels should contain month abbreviations, got: %s", labels)
	}
}

func TestNavigationStack_InitiallyEmpty(t *testing.T) {
	model := New(nil)

	if len(model.viewStack) != 0 {
		t.Errorf("viewStack should be empty initially, got %d items", len(model.viewStack))
	}
}

func TestNavigationStack_PushWhenSwitchingViews(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeHabits {
		t.Errorf("expected ViewTypeHabits, got %v", m.currentView)
	}
	if len(m.viewStack) != 1 {
		t.Errorf("viewStack should have 1 item after switching views, got %d", len(m.viewStack))
	}
	if len(m.viewStack) > 0 && m.viewStack[0] != ViewTypeJournal {
		t.Errorf("viewStack[0] should be ViewTypeJournal, got %v", m.viewStack[0])
	}
}

func TestNavigationStack_PopWhenPressingEsc(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.viewStack = []ViewType{ViewTypeJournal}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.currentView != ViewTypeJournal {
		t.Errorf("should navigate back to ViewTypeJournal, got %v", m.currentView)
	}
	if len(m.viewStack) != 0 {
		t.Errorf("viewStack should be empty after pop, got %d items", len(m.viewStack))
	}
}

func TestNavigationStack_QShowsConfirmEvenWithStack(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeHabits
	model.viewStack = []ViewType{ViewTypeJournal}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.quitConfirmMode.active {
		t.Error("pressing q should show quit confirmation, even with views in stack")
	}
	if m.currentView != ViewTypeHabits {
		t.Errorf("should still be in ViewTypeHabits, got %v", m.currentView)
	}
}

func TestNavigationStack_ShowConfirmWhenEmptyStack(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.viewStack = []ViewType{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.quitConfirmMode.active {
		t.Error("should show quit confirmation when at root view")
	}
	if m.currentView != ViewTypeJournal {
		t.Error("should stay on journal view while confirming")
	}
}

func TestNavigationStack_ConfirmYes_Quits(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.quitConfirmMode.active = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("confirming quit with 'y' should return a command")
	}
}

func TestNavigationStack_ConfirmNo_CancelsQuit(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal
	model.quitConfirmMode.active = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.quitConfirmMode.active {
		t.Error("quit confirm mode should be deactivated after pressing 'n'")
	}
}

func TestQuitConfirmView_ShowsWhenActive(t *testing.T) {
	model := New(nil)
	model.quitConfirmMode.active = true

	view := model.View()

	if !strings.Contains(view, "Quit") {
		t.Error("quit confirm view should contain 'Quit'")
	}
	if !strings.Contains(view, "Are you sure") {
		t.Error("quit confirm view should ask 'Are you sure'")
	}
}

// ============================================================================
// Undo Functionality Tests
// ============================================================================

func TestModel_UndoMarkDone(t *testing.T) {
	model := New(nil)
	taskEntry := domain.Entry{
		ID:      1,
		Type:    domain.EntryTypeTask,
		Content: "Test task",
	}
	model.entries = []EntryItem{
		{Entry: taskEntry},
	}
	model.selectedIdx = 0

	msgSpace := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := model.Update(msgSpace)
	m := newModel.(Model)

	if m.undoState.operation == UndoOpNone {
		t.Error("expected undo state to be set after marking done")
	}

	msgUndo := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	newModel, _ = m.Update(msgUndo)
	m = newModel.(Model)

	if m.undoState.operation != UndoOpNone {
		t.Error("expected undo state to be cleared after undo")
	}
}

func TestModel_WeeklyView_MarkDone_PreservesFocus(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypeJournal

	// Create multiple entries in weekly view context
	entries := []EntryItem{
		{Entry: domain.Entry{ID: 1, EntityID: "entity-1", Type: domain.EntryTypeTask, Content: "Task 1"}},
		{Entry: domain.Entry{ID: 2, EntityID: "entity-2", Type: domain.EntryTypeTask, Content: "Task 2"}},
		{Entry: domain.Entry{ID: 3, EntityID: "entity-3", Type: domain.EntryTypeTask, Content: "Task 3"}},
		{Entry: domain.Entry{ID: 4, EntityID: "entity-4", Type: domain.EntryTypeTask, Content: "Task 4"}},
	}
	model.entries = entries

	// Select the 3rd entry (index 2)
	model.selectedIdx = 2
	selectedEntityID := model.entries[2].Entry.EntityID

	// Create an agenda with the same entries
	// This simulates what happens when the agenda is reloaded
	today := time.Now()
	dayEntries := service.DayEntries{
		Date:    today,
		Entries: []domain.Entry{entries[0].Entry, entries[1].Entry, entries[2].Entry, entries[3].Entry},
	}

	daysMsg := daysLoadedMsg{
		days: []service.DayEntries{dayEntries},
	}

	newModel, _ := model.Update(daysMsg)
	m := newModel.(Model)

	// Focus should be preserved on the same entry EntityID
	// Find the new index of the selected entry
	newIdx := -1
	for idx, item := range m.entries {
		if item.Entry.EntityID == selectedEntityID {
			newIdx = idx
			break
		}
	}

	if newIdx == -1 {
		t.Error("selected entry not found after days reload")
	}

	if m.selectedIdx != newIdx {
		t.Errorf("focus was not preserved, expected index %d, got %d", newIdx, m.selectedIdx)
	}

	// The selectedIdx should still be within valid bounds
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.entries) {
		t.Errorf("selectedIdx is out of bounds after daysLoadedMsg, got %d, len=%d", m.selectedIdx, len(m.entries))
	}
}

func TestModel_WeeklyView_PastCompletedTasks_DisplayedInRed(t *testing.T) {
	model := New(nil)

	// Create entries: a recently completed task and an older completed task
	today := time.Now()
	oldDate := today.AddDate(0, 0, -7)

	model.entries = []EntryItem{
		{
			Entry: domain.Entry{
				ID:      1,
				Type:    domain.EntryTypeDone,
				Content: "Completed today",
				// This entry was just completed
			},
		},
		{
			Entry: domain.Entry{
				ID:            2,
				Type:          domain.EntryTypeDone,
				Content:       "Completed in past",
				ScheduledDate: &oldDate,
				// This entry was completed in the past
			},
			IsOverdue: true, // Mark as a past entry
		},
	}

	// Render the past completed entry (it's overdue)
	rendered := model.renderEntry(model.entries[1], false)

	// The entry should be styled with ANSI escape codes
	// Check for presence of ANSI escape sequence (\x1b[) which indicates styling was applied
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("expected past completed task to be styled with ANSI codes, but no styling found")
	}
}

func TestRenderJournalContent_EmptyMessage_DayView(t *testing.T) {
	model := New(nil)
	model.viewMode = ViewModeDay
	model.entries = []EntryItem{}
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24

	view := model.renderJournalContent()

	if strings.Contains(view, "7 days") {
		t.Error("day view should not mention '7 days' when there are no entries")
	}
	if !strings.Contains(view, "today") {
		t.Error("day view should say 'No entries for today' when there are no entries")
	}
}

func TestRenderJournalContent_EmptyMessage_WeekView(t *testing.T) {
	model := New(nil)
	model.viewMode = ViewModeWeek
	model.entries = []EntryItem{}
	model.days = []service.DayEntries{}
	model.width = 80
	model.height = 24

	view := model.renderJournalContent()

	if !strings.Contains(view, "7 days") {
		t.Error("week view should say 'No entries for the last 7 days' when there are no entries")
	}
}

func TestPendingTasks_NavigateDown(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	today := time.Now()
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "Task 1", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 2, Content: "Task 2", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 3, Content: "Task 3", Type: domain.EntryTypeTask, ScheduledDate: &today},
	}
	model.pendingTasksState.selectedIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.pendingTasksState.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1, got %d", m.pendingTasksState.selectedIdx)
	}
}

func TestPendingTasks_NavigateUp(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	today := time.Now()
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "Task 1", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 2, Content: "Task 2", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 3, Content: "Task 3", Type: domain.EntryTypeTask, ScheduledDate: &today},
	}
	model.pendingTasksState.selectedIdx = 2

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.pendingTasksState.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1, got %d", m.pendingTasksState.selectedIdx)
	}
}

func TestPendingTasks_NavigateDownAtBounds(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	today := time.Now()
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "Task 1", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 2, Content: "Task 2", Type: domain.EntryTypeTask, ScheduledDate: &today},
	}
	model.pendingTasksState.selectedIdx = 1

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.pendingTasksState.selectedIdx != 1 {
		t.Errorf("expected selectedIdx to stay at 1 (at bounds), got %d", m.pendingTasksState.selectedIdx)
	}
}

func TestPendingTasks_NavigateUpAtBounds(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	today := time.Now()
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "Task 1", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 2, Content: "Task 2", Type: domain.EntryTypeTask, ScheduledDate: &today},
	}
	model.pendingTasksState.selectedIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.pendingTasksState.selectedIdx != 0 {
		t.Errorf("expected selectedIdx to stay at 0 (at bounds), got %d", m.pendingTasksState.selectedIdx)
	}
}

func TestPendingTasks_ScrollOffsetAdjustsWhenNavigatingBeyondViewport(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 10

	today := time.Now()
	entries := make([]domain.Entry, 20)
	for i := range entries {
		entries[i] = domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Task %d", i+1), Type: domain.EntryTypeTask, ScheduledDate: &today}
	}
	model.pendingTasksState.entries = entries
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.scrollOffset = 0

	for i := 0; i < 8; i++ {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := model.Update(msg)
		model = newModel.(Model)
	}

	if model.pendingTasksState.scrollOffset <= 0 {
		t.Errorf("expected scrollOffset to be > 0 after navigating beyond viewport, got %d", model.pendingTasksState.scrollOffset)
	}
}

func TestPendingTasks_RenderShowsScrollIndicators(t *testing.T) {
	today := time.Now()
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 15

	entries := make([]domain.Entry, 20)
	for i := range entries {
		entries[i] = domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Task %d", i+1), Type: domain.EntryTypeTask, ScheduledDate: &today}
	}
	model.pendingTasksState.entries = entries
	model.pendingTasksState.selectedIdx = 10
	model.pendingTasksState.scrollOffset = 5

	view := model.View()

	if !strings.Contains(view, "↑") {
		t.Error("expected scroll up indicator when scrollOffset > 0")
	}
	if !strings.Contains(view, "↓") {
		t.Error("expected scroll down indicator when more items below viewport")
	}
}

func TestPendingTasks_RenderOnlyShowsVisibleEntries(t *testing.T) {
	today := time.Now()
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 15

	entries := make([]domain.Entry, 20)
	for i := range entries {
		entries[i] = domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Task_Item_%02d", i+1), Type: domain.EntryTypeTask, ScheduledDate: &today}
	}
	model.pendingTasksState.entries = entries
	model.pendingTasksState.selectedIdx = 10
	model.pendingTasksState.scrollOffset = 8

	view := model.View()

	if strings.Contains(view, "Task_Item_01") {
		t.Error("Task 1 should not be visible when scrollOffset is 8")
	}
	if !strings.Contains(view, "Task_Item_09") {
		t.Error("Task 9 should be visible (index 8, at scrollOffset)")
	}
}

func TestPendingTasks_ShowsContextIndicatorForEntriesWithParents(t *testing.T) {
	today := time.Now()
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	parentID := int64(100)
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "Standalone task", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 2, Content: "Nested task", Type: domain.EntryTypeTask, ScheduledDate: &today, ParentID: &parentID},
	}
	model.pendingTasksState.selectedIdx = 0

	grandparentID := int64(99)
	model.pendingTasksState.parentChains = map[int64][]domain.Entry{
		2: {
			{ID: 100, Content: "Parent event", Type: domain.EntryTypeEvent},
			{ID: 99, Content: "Grandparent", Type: domain.EntryTypeEvent, ParentID: &grandparentID},
		},
	}

	view := model.View()

	if !strings.Contains(view, "↳") {
		t.Error("expected context indicator ↳ for nested task with parents")
	}
}

func TestPendingTasks_ShowsParentChainForSelectedEntry(t *testing.T) {
	today := time.Now()
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	parentID := int64(100)
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 2, Content: "Nested task", Type: domain.EntryTypeTask, ScheduledDate: &today, ParentID: &parentID},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.expandedID = 2

	model.pendingTasksState.parentChains = map[int64][]domain.Entry{
		2: {
			{ID: 100, Content: "Parent event", Type: domain.EntryTypeEvent},
		},
	}

	view := model.View()

	if !strings.Contains(view, "Parent event") {
		t.Error("expected parent context to be shown for selected entry")
	}
	if !strings.Contains(view, ">") {
		t.Error("expected breadcrumb indicator in parent chain display")
	}
}

func TestPendingTasks_EnterTogglesContextExpansion(t *testing.T) {
	today := time.Now()
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	parentID := int64(100)
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 2, Content: "Nested task", Type: domain.EntryTypeTask, ScheduledDate: &today, ParentID: &parentID},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.expandedID = 0
	model.pendingTasksState.parentChains = map[int64][]domain.Entry{
		2: {{ID: 100, Content: "Parent event", Type: domain.EntryTypeEvent}},
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.pendingTasksState.expandedID != 2 {
		t.Errorf("expected expandedID to be 2 after Enter, got %d", m.pendingTasksState.expandedID)
	}

	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.pendingTasksState.expandedID != 0 {
		t.Errorf("expected expandedID to be 0 after second Enter (toggle off), got %d", m.pendingTasksState.expandedID)
	}
}

func TestPendingTasks_EnterTriggersParentChainLoadingWhenNotCached(t *testing.T) {
	today := time.Now()
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 24

	parentID := int64(100)
	model.pendingTasksState.entries = []domain.Entry{
		{ID: 2, Content: "Nested task", Type: domain.EntryTypeTask, ScheduledDate: &today, ParentID: &parentID},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.expandedID = 0
	model.pendingTasksState.parentChains = make(map[int64][]domain.Entry)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("expected a command to be returned to load parent chain, got nil")
	}
}

func TestPendingTasks_ParentChainLoadedMsgStoresChainAndExpandsEntry(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.pendingTasksState.parentChains = make(map[int64][]domain.Entry)

	parentChain := []domain.Entry{
		{ID: 100, Content: "Parent event", Type: domain.EntryTypeEvent},
	}
	msg := parentChainLoadedMsg{
		entryID: 2,
		chain:   parentChain,
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.pendingTasksState.expandedID != 2 {
		t.Errorf("expected expandedID to be 2 after parentChainLoadedMsg, got %d", m.pendingTasksState.expandedID)
	}

	chain, ok := m.pendingTasksState.parentChains[2]
	if !ok {
		t.Error("expected parent chain to be stored in parentChains map")
	}
	if len(chain) != 1 {
		t.Errorf("expected parent chain length to be 1, got %d", len(chain))
	}
}

func TestPendingTasks_GroupsTasksByDateWithHeaders(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 30

	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "Task from two days ago", Type: domain.EntryTypeTask, ScheduledDate: &twoDaysAgo},
		{ID: 2, Content: "Yesterday task 1", Type: domain.EntryTypeTask, ScheduledDate: &yesterday},
		{ID: 3, Content: "Yesterday task 2", Type: domain.EntryTypeTask, ScheduledDate: &yesterday},
		{ID: 4, Content: "Today task", Type: domain.EntryTypeTask, ScheduledDate: &today},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.parentChains = make(map[int64][]domain.Entry)

	view := model.View()

	if !strings.Contains(view, twoDaysAgo.Format("Mon, Jan 2")) {
		t.Errorf("expected date header for two days ago (%s) in view", twoDaysAgo.Format("Mon, Jan 2"))
	}
	if !strings.Contains(view, yesterday.Format("Mon, Jan 2")) {
		t.Errorf("expected date header for yesterday (%s) in view", yesterday.Format("Mon, Jan 2"))
	}
	if !strings.Contains(view, today.Format("Mon, Jan 2")) {
		t.Errorf("expected date header for today (%s) in view", today.Format("Mon, Jan 2"))
	}
}

func TestPendingTasks_DoesNotShowDateOnEachTaskLine(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 30

	today := time.Now().Truncate(24 * time.Hour)

	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "First task", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 2, Content: "Second task", Type: domain.EntryTypeTask, ScheduledDate: &today},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.parentChains = make(map[int64][]domain.Entry)

	view := model.View()

	dateOnLine := today.Format("2006-01-02")
	occurrences := strings.Count(view, dateOnLine)
	if occurrences > 1 {
		t.Errorf("expected date %s to NOT appear on each task line, found %d occurrences", dateOnLine, occurrences)
	}
}

func TestPendingTasks_TasksWithinGroupShowWithoutDate(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 30

	today := time.Now().Truncate(24 * time.Hour)

	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "My task content", Type: domain.EntryTypeTask, ScheduledDate: &today},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.parentChains = make(map[int64][]domain.Entry)

	view := model.View()

	if !strings.Contains(view, "• My task content") {
		t.Error("expected task line to show just the symbol and content without date bracket prefix")
	}
	if strings.Contains(view, "["+today.Format("2006-01-02")+"]") {
		t.Error("expected task line to NOT have date in brackets")
	}
}

func TestPendingTasks_ShowsDateHeaderWhenScrolled(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 15

	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	entries := make([]domain.Entry, 20)
	for i := 0; i < 10; i++ {
		entries[i] = domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Yesterday task %d", i+1), Type: domain.EntryTypeTask, ScheduledDate: &yesterday}
	}
	for i := 10; i < 20; i++ {
		entries[i] = domain.Entry{ID: int64(i + 1), Content: fmt.Sprintf("Today task %d", i-9), Type: domain.EntryTypeTask, ScheduledDate: &today}
	}
	model.pendingTasksState.entries = entries
	model.pendingTasksState.selectedIdx = 5
	model.pendingTasksState.scrollOffset = 5
	model.pendingTasksState.parentChains = make(map[int64][]domain.Entry)

	view := model.View()

	if !strings.Contains(view, yesterday.Format("Mon, Jan 2")) {
		t.Errorf("expected date header for yesterday (%s) when scrolled into that date group", yesterday.Format("Mon, Jan 2"))
	}
}

func TestPendingTasks_ParentChainIndentationIsCorrect(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 30

	today := time.Now().Truncate(24 * time.Hour)
	parentID := int64(100)

	model.pendingTasksState.entries = []domain.Entry{
		{ID: 2, Content: "Child task", Type: domain.EntryTypeTask, ScheduledDate: &today, ParentID: &parentID},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.expandedID = 2
	model.pendingTasksState.parentChains = map[int64][]domain.Entry{
		2: {
			{ID: 100, Content: "Parent note", Type: domain.EntryTypeNote},
		},
	}

	view := testutil.StripAnsi(model.View())

	if !strings.Contains(view, "  > – Parent note") {
		t.Errorf("expected parent to be indented with '  > – Parent note', got:\n%s", view)
	}
	if !strings.Contains(view, "    • Child task") {
		t.Errorf("expected child task to be more indented than parent with '    • Child task', got:\n%s", view)
	}
}

func TestPendingTasks_ContextIndicatorIsVisible(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 30

	today := time.Now().Truncate(24 * time.Hour)
	parentID := int64(100)

	model.pendingTasksState.entries = []domain.Entry{
		{ID: 1, Content: "Task without context", Type: domain.EntryTypeTask, ScheduledDate: &today},
		{ID: 2, Content: "Task with context", Type: domain.EntryTypeTask, ScheduledDate: &today, ParentID: &parentID},
	}
	model.pendingTasksState.selectedIdx = 0
	model.pendingTasksState.parentChains = map[int64][]domain.Entry{
		2: {
			{ID: 100, Content: "Parent note", Type: domain.EntryTypeNote},
		},
	}

	view := testutil.StripAnsi(model.View())

	// Task with context should have a clear indicator (not just small [1])
	if !strings.Contains(view, "↳") {
		t.Errorf("expected context indicator with ↳ symbol for task with parents, got:\n%s", view)
	}

	// Task without context should NOT have the indicator
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Task without context") && strings.Contains(line, "↳") {
			t.Errorf("task without context should not have ↳ indicator")
		}
	}
}

func TestPendingTasks_ScrollingAccountsForDateHeaders(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 20 // Small height to force scrolling issues

	// Create entries across 5 different dates (each date header takes a line)
	entries := make([]domain.Entry, 0, 15)
	for i := 0; i < 15; i++ {
		date := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		entries = append(entries, domain.Entry{
			ID:            int64(i + 1),
			Content:       fmt.Sprintf("Task %d", i+1),
			Type:          domain.EntryTypeTask,
			ScheduledDate: &date,
		})
	}
	model.pendingTasksState.entries = entries
	model.pendingTasksState.selectedIdx = 0

	view := testutil.StripAnsi(model.View())
	lines := strings.Split(view, "\n")

	// Count actual content lines (excluding empty lines at end)
	contentLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			contentLines++
		}
	}

	// Should not exceed height (with some margin for header/footer)
	maxExpectedLines := model.height + 2 // small buffer for final newlines
	if contentLines > maxExpectedLines {
		t.Errorf("view has %d content lines but height is %d, content:\n%s", contentLines, model.height, view)
	}
}

func TestPendingTasks_ScrollingToBottomShowsLastEntry(t *testing.T) {
	model := New(nil)
	model.currentView = ViewTypePendingTasks
	model.width = 80
	model.height = 20

	// Create entries across multiple dates
	entries := make([]domain.Entry, 0, 10)
	for i := 0; i < 10; i++ {
		date := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		entries = append(entries, domain.Entry{
			ID:            int64(i + 1),
			Content:       fmt.Sprintf("Task %d", i+1),
			Type:          domain.EntryTypeTask,
			ScheduledDate: &date,
		})
	}
	model.pendingTasksState.entries = entries

	// Navigate to bottom
	model.pendingTasksState.selectedIdx = len(entries) - 1
	model = model.ensurePendingTaskVisible()

	view := testutil.StripAnsi(model.View())

	// Last entry (Task 10) should be visible
	if !strings.Contains(view, "Task 10") {
		t.Errorf("expected last entry 'Task 10' to be visible after scrolling to bottom, got:\n%s", view)
	}

	// Selection indicator should be on Task 10
	lines := strings.Split(view, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Task 10") && strings.Contains(line, ">") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Task 10 to be selected (have > indicator), got:\n%s", view)
	}
}
