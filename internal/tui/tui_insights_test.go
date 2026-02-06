package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/typingincolor/bujo/internal/domain"
)

func newInsightsModel() Model {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeInsights
	return model
}

func TestInsights_IKeySwitchesToInsightsView(t *testing.T) {
	model := New(nil)
	model.width = 80
	model.height = 24
	model.currentView = ViewTypeJournal

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.currentView != ViewTypeInsights {
		t.Errorf("expected ViewTypeInsights, got %d", m.currentView)
	}
}

func TestInsights_DefaultTabIsDashboard(t *testing.T) {
	model := newInsightsModel()

	if model.insightsState.activeTab != InsightsTabDashboard {
		t.Errorf("expected dashboard tab, got %d", model.insightsState.activeTab)
	}
}

func TestInsights_TabKeySwitchesTab(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabDashboard

	msg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.activeTab != InsightsTabSummaries {
		t.Errorf("expected summaries tab, got %d", m.insightsState.activeTab)
	}

	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.activeTab != InsightsTabActions {
		t.Errorf("expected actions tab, got %d", m.insightsState.activeTab)
	}

	// wraps around
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.activeTab != InsightsTabDashboard {
		t.Errorf("expected dashboard tab (wrap), got %d", m.insightsState.activeTab)
	}
}

func TestInsights_ShiftTabGoesBack(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.activeTab != InsightsTabDashboard {
		t.Errorf("expected dashboard tab, got %d", m.insightsState.activeTab)
	}

	// wraps around backwards
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.activeTab != InsightsTabActions {
		t.Errorf("expected actions tab (wrap), got %d", m.insightsState.activeTab)
	}
}

func TestInsights_HLNavigatesWeeks(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries
	initialWeek := model.insightsState.weekAnchor

	// h = previous week
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	result, _ := model.Update(msg)
	m := result.(Model)

	expectedPrev := initialWeek.AddDate(0, 0, -7)
	if !m.insightsState.weekAnchor.Equal(expectedPrev) {
		t.Errorf("expected week %v, got %v", expectedPrev, m.insightsState.weekAnchor)
	}

	// l = next week
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	result, _ = m.Update(msg)
	m = result.(Model)

	if !m.insightsState.weekAnchor.Equal(initialWeek) {
		t.Errorf("expected week %v, got %v", initialWeek, m.insightsState.weekAnchor)
	}
}

func TestInsights_HLNoOpOnDashboard(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabDashboard
	initialWeek := model.insightsState.weekAnchor

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	result, _ := model.Update(msg)
	m := result.(Model)

	if !m.insightsState.weekAnchor.Equal(initialWeek) {
		t.Errorf("week should not change on dashboard, got %v", m.insightsState.weekAnchor)
	}
}

func TestInsights_DashboardLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.loading = true

	dashboard := domain.InsightsDashboard{
		LatestSummary: &domain.InsightsSummary{
			ID:          1,
			WeekStart:   "2026-01-26",
			WeekEnd:     "2026-02-01",
			SummaryText: "Test summary",
		},
		ActiveInitiatives:    []domain.InsightsInitiative{},
		HighPriorityActions:  []domain.InsightsAction{},
		RecentDecisions:      []domain.InsightsDecision{},
		DaysSinceLastSummary: 3,
		Status:               "ok",
	}

	msg := insightsDashboardLoadedMsg{dashboard: dashboard}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to be false")
	}
	if m.insightsState.dashboard.LatestSummary == nil {
		t.Error("expected dashboard to have latest summary")
	}
	if m.insightsState.dashboard.DaysSinceLastSummary != 3 {
		t.Errorf("expected 3 days since last summary, got %d", m.insightsState.dashboard.DaysSinceLastSummary)
	}
}

func TestInsights_SummaryLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries

	summary := &domain.InsightsSummary{
		ID:          1,
		WeekStart:   "2026-01-26",
		WeekEnd:     "2026-02-01",
		SummaryText: "Week summary text",
	}
	topics := []domain.InsightsTopic{
		{ID: 1, SummaryID: 1, Topic: "Testing", Content: "Good progress", Importance: "high"},
	}

	msg := insightsSummaryLoadedMsg{summary: summary, topics: topics}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.weekSummary == nil {
		t.Error("expected week summary to be set")
	}
	if len(m.insightsState.weekTopics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(m.insightsState.weekTopics))
	}
}

func TestInsights_ActionsLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabActions

	actions := []domain.InsightsAction{
		{ID: 1, ActionText: "Do thing", Priority: "high", Status: "pending"},
		{ID: 2, ActionText: "Do other", Priority: "low", Status: "pending"},
	}

	msg := insightsActionsLoadedMsg{actions: actions}
	result, _ := model.Update(msg)
	m := result.(Model)

	if len(m.insightsState.weekActions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(m.insightsState.weekActions))
	}
}

func TestInsights_EscGoesBack(t *testing.T) {
	model := newInsightsModel()
	model.viewStack = []ViewType{ViewTypeJournal}

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.currentView != ViewTypeJournal {
		t.Errorf("expected journal view after esc, got %d", m.currentView)
	}
}

func TestInsights_RenderDashboard(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.dashboard = domain.InsightsDashboard{
		LatestSummary: &domain.InsightsSummary{
			WeekStart:   "2026-01-26",
			WeekEnd:     "2026-02-01",
			SummaryText: "Great week of progress",
		},
		ActiveInitiatives:    []domain.InsightsInitiative{{Name: "Project Alpha", Status: "active"}},
		HighPriorityActions:  []domain.InsightsAction{{ActionText: "Fix bug", Priority: "high"}},
		RecentDecisions:      []domain.InsightsDecision{{DecisionText: "Use Go"}},
		DaysSinceLastSummary: 2,
		Status:               "ok",
	}

	output := model.renderInsightsContent()

	if output == "" {
		t.Error("expected non-empty render output")
	}
	if !strings.Contains(output, "Dashboard") {
		t.Error("expected output to contain Dashboard tab indicator")
	}
	if !strings.Contains(output, "Great week") {
		t.Error("expected output to contain summary text")
	}
	if !strings.Contains(output, "Project Alpha") {
		t.Error("expected output to contain initiative name")
	}
}

func TestInsights_RenderSummaries(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries
	model.insightsState.weekSummary = &domain.InsightsSummary{
		WeekStart:   "2026-01-26",
		WeekEnd:     "2026-02-01",
		SummaryText: "Summary for week",
	}
	model.insightsState.weekTopics = []domain.InsightsTopic{
		{Topic: "TDD", Content: "Applied consistently", Importance: "high"},
	}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "Summaries") {
		t.Error("expected output to contain Summaries tab indicator")
	}
	if !strings.Contains(output, "Summary for week") {
		t.Error("expected output to contain summary text")
	}
	if !strings.Contains(output, "TDD") {
		t.Error("expected output to contain topic name")
	}
}

func TestInsights_RenderActions(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabActions
	model.insightsState.weekActions = []domain.InsightsAction{
		{ActionText: "Write tests", Priority: "high", Status: "pending"},
		{ActionText: "Deploy app", Priority: "low", Status: "pending"},
	}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "Actions") {
		t.Error("expected output to contain Actions tab indicator")
	}
	if !strings.Contains(output, "Write tests") {
		t.Error("expected output to contain action text")
	}
	if !strings.Contains(output, "Deploy app") {
		t.Error("expected output to contain second action text")
	}
}

func TestInsights_RenderNoData(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.dashboard = domain.InsightsDashboard{}

	output := model.renderInsightsContent()

	if output == "" {
		t.Error("expected non-empty render even with no data")
	}
}

func TestInsights_RenderLoading(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.loading = true

	output := model.renderInsightsContent()

	if !strings.Contains(output, "Loading") {
		t.Error("expected loading indicator")
	}
}

func TestInsights_WeekAnchorInitializesToCurrentWeekMonday(t *testing.T) {
	model := newInsightsModel()

	anchor := model.insightsState.weekAnchor
	if anchor.Weekday() != time.Monday {
		t.Errorf("expected week anchor to be Monday, got %v", anchor.Weekday())
	}
}

func TestInsights_HelpText(t *testing.T) {
	model := newInsightsModel()
	help := model.renderContextHelp()

	if !strings.Contains(help, "tab") {
		t.Error("expected help text to mention tab key")
	}
	if !strings.Contains(help, "h/l") {
		t.Error("expected help text to mention h/l for week nav")
	}
}

func TestInsights_ViewTypeDisplayName(t *testing.T) {
	model := newInsightsModel()
	view := model.View()

	if !strings.Contains(view, "Insights") {
		t.Error("expected view to contain 'Insights' display name")
	}
}
