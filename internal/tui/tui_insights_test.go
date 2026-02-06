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

	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.activeTab != InsightsTabInitiatives {
		t.Errorf("expected initiatives tab, got %d", m.insightsState.activeTab)
	}

	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.activeTab != InsightsTabTopics {
		t.Errorf("expected topics tab, got %d", m.insightsState.activeTab)
	}

	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.activeTab != InsightsTabDecisions {
		t.Errorf("expected decisions tab, got %d", m.insightsState.activeTab)
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

	if m.insightsState.activeTab != InsightsTabDecisions {
		t.Errorf("expected decisions tab (wrap), got %d", m.insightsState.activeTab)
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

func TestInsights_TabCycleIncludesInitiatives(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabActions

	msg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.activeTab != InsightsTabInitiatives {
		t.Errorf("expected initiatives tab after actions, got %d", m.insightsState.activeTab)
	}
}

func TestInsights_InitiativesLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.loading = true

	initiatives := []domain.InsightsInitiativePortfolio{
		{ID: 1, Name: "GenAI Integration", Status: "active", MentionCount: 2},
		{ID: 2, Name: "Tech Scorecard", Status: "active", MentionCount: 1},
	}

	msg := insightsInitiativesLoadedMsg{initiatives: initiatives}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to be false")
	}
	if len(m.insightsState.initiatives) != 2 {
		t.Errorf("expected 2 initiatives, got %d", len(m.insightsState.initiatives))
	}
}

func TestInsights_RenderInitiatives(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives
	model.insightsState.initiatives = []domain.InsightsInitiativePortfolio{
		{ID: 1, Name: "GenAI Integration", Status: "active", MentionCount: 2, LastMentionWeek: "2026-01-27"},
		{ID: 2, Name: "Tech Scorecard", Status: "active", MentionCount: 1, LastMentionWeek: "2026-01-20"},
		{ID: 3, Name: "Q1 OKRs", Status: "completed", MentionCount: 0},
	}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "Initiatives") {
		t.Error("expected output to contain Initiatives tab indicator")
	}
	if !strings.Contains(output, "GenAI Integration") {
		t.Error("expected output to contain initiative name")
	}
	if !strings.Contains(output, "Tech Scorecard") {
		t.Error("expected output to contain second initiative name")
	}
	if !strings.Contains(output, "active") {
		t.Error("expected output to contain status")
	}
}

func TestInsights_RenderInitiativesEmpty(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives
	model.insightsState.initiatives = []domain.InsightsInitiativePortfolio{}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "No initiatives") {
		t.Error("expected output to contain empty state message")
	}
}

func TestInsights_InitiativeDetailLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives
	model.insightsState.loading = true

	detail := &domain.InsightsInitiativeDetail{
		Initiative: domain.InsightsInitiative{
			ID:     1,
			Name:   "GenAI Integration",
			Status: "active",
		},
		Updates: []domain.InsightsInitiativeUpdate{
			{WeekStart: "2026-01-13", WeekEnd: "2026-01-19", UpdateText: "Started planning"},
			{WeekStart: "2026-01-27", WeekEnd: "2026-02-02", UpdateText: "Sprint completed"},
		},
		PendingActions: []domain.InsightsAction{
			{ID: 1, ActionText: "Review AI outputs", Priority: "high", Status: "pending"},
		},
		Decisions: []domain.InsightsDecision{
			{ID: 1, DecisionText: "Adopt Claude as primary AI provider"},
		},
	}

	msg := insightsInitiativeDetailLoadedMsg{detail: detail}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to be false")
	}
	if m.insightsState.initiativeDetail == nil {
		t.Error("expected initiative detail to be set")
	}
	if m.insightsState.initiativeDetail.Initiative.Name != "GenAI Integration" {
		t.Errorf("expected GenAI Integration, got %s", m.insightsState.initiativeDetail.Initiative.Name)
	}
}

func TestInsights_EnterOnInitiativesSetsSelectedID(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives
	model.insightsState.initiatives = []domain.InsightsInitiativePortfolio{
		{ID: 1, Name: "GenAI Integration", Status: "active"},
		{ID: 2, Name: "Tech Scorecard", Status: "active"},
	}
	model.insightsState.initiativeSelectedIdx = 1

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.selectedInitiativeID == nil {
		t.Fatal("expected selectedInitiativeID to be set")
	}
	if *m.insightsState.selectedInitiativeID != 2 {
		t.Errorf("expected initiative ID 2, got %d", *m.insightsState.selectedInitiativeID)
	}
}

func TestInsights_EscOnInitiativeDetailReturnsToList(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives
	id := int64(1)
	model.insightsState.selectedInitiativeID = &id
	model.insightsState.initiativeDetail = &domain.InsightsInitiativeDetail{
		Initiative: domain.InsightsInitiative{ID: 1, Name: "GenAI"},
	}

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.selectedInitiativeID != nil {
		t.Error("expected selectedInitiativeID to be nil after escape")
	}
	if m.insightsState.initiativeDetail != nil {
		t.Error("expected initiativeDetail to be nil after escape")
	}
}

func TestInsights_RenderInitiativeDetail(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives
	id := int64(1)
	model.insightsState.selectedInitiativeID = &id
	model.insightsState.initiativeDetail = &domain.InsightsInitiativeDetail{
		Initiative: domain.InsightsInitiative{
			ID:     1,
			Name:   "GenAI Integration",
			Status: "active",
		},
		Updates: []domain.InsightsInitiativeUpdate{
			{WeekStart: "2026-01-13", WeekEnd: "2026-01-19", UpdateText: "Started planning"},
		},
		PendingActions: []domain.InsightsAction{
			{ActionText: "Review AI outputs", Priority: "high", Status: "pending"},
		},
		Decisions: []domain.InsightsDecision{
			{DecisionText: "Adopt Claude as primary AI provider"},
		},
	}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "GenAI Integration") {
		t.Error("expected output to contain initiative name")
	}
	if !strings.Contains(output, "Started planning") {
		t.Error("expected output to contain update text")
	}
	if !strings.Contains(output, "Review AI outputs") {
		t.Error("expected output to contain action text")
	}
	if !strings.Contains(output, "Adopt Claude") {
		t.Error("expected output to contain decision text")
	}
}

func TestInsights_JKNavigatesInitiativesList(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives
	model.insightsState.initiatives = []domain.InsightsInitiativePortfolio{
		{ID: 1, Name: "GenAI Integration", Status: "active"},
		{ID: 2, Name: "Tech Scorecard", Status: "active"},
		{ID: 3, Name: "Q1 OKRs", Status: "completed"},
	}
	model.insightsState.initiativeSelectedIdx = 0

	// j moves down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.initiativeSelectedIdx != 1 {
		t.Errorf("expected idx 1 after j, got %d", m.insightsState.initiativeSelectedIdx)
	}

	// j again
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.initiativeSelectedIdx != 2 {
		t.Errorf("expected idx 2 after second j, got %d", m.insightsState.initiativeSelectedIdx)
	}

	// j at bottom stays
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.initiativeSelectedIdx != 2 {
		t.Errorf("expected idx 2 (clamped), got %d", m.insightsState.initiativeSelectedIdx)
	}

	// k moves up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.initiativeSelectedIdx != 1 {
		t.Errorf("expected idx 1 after k, got %d", m.insightsState.initiativeSelectedIdx)
	}
}

func TestInsights_TabCycleIncludesTopics(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabInitiatives

	msg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.activeTab != InsightsTabTopics {
		t.Errorf("expected topics tab after initiatives, got %d", m.insightsState.activeTab)
	}
}

func TestInsights_TopicsListLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.loading = true

	topics := []string{"GenAI", "Quarterly Planning", "Sprint Review"}

	msg := insightsTopicsListLoadedMsg{topics: topics}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to be false")
	}
	if len(m.insightsState.distinctTopics) != 3 {
		t.Errorf("expected 3 topics, got %d", len(m.insightsState.distinctTopics))
	}
}

func TestInsights_TopicTimelineLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.loading = true

	timeline := []domain.InsightsTopicTimeline{
		{Topic: "GenAI", Content: "AI features", Importance: "high", WeekStart: "2026-01-13", WeekEnd: "2026-01-19"},
		{Topic: "GenAI", Content: "AI sprint done", Importance: "high", WeekStart: "2026-01-27", WeekEnd: "2026-02-02"},
	}

	msg := insightsTopicTimelineLoadedMsg{timeline: timeline}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to be false")
	}
	if len(m.insightsState.topicTimeline) != 2 {
		t.Errorf("expected 2 timeline entries, got %d", len(m.insightsState.topicTimeline))
	}
}

func TestInsights_RenderTopicsList(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabTopics
	model.insightsState.distinctTopics = []string{"GenAI", "Quarterly Planning", "Sprint Review"}
	model.insightsState.topicSelectedIdx = 0

	output := model.renderInsightsContent()

	if !strings.Contains(output, "Topics") {
		t.Error("expected output to contain Topics tab indicator")
	}
	if !strings.Contains(output, "GenAI") {
		t.Error("expected output to contain topic name")
	}
	if !strings.Contains(output, "Quarterly Planning") {
		t.Error("expected output to contain second topic name")
	}
}

func TestInsights_RenderTopicsEmpty(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabTopics
	model.insightsState.distinctTopics = []string{}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "No topics") {
		t.Error("expected output to contain empty state message")
	}
}

func TestInsights_EnterOnTopicsLoadsTimeline(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabTopics
	model.insightsState.distinctTopics = []string{"GenAI", "Sprint Review"}
	model.insightsState.topicSelectedIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.selectedTopic != "GenAI" {
		t.Errorf("expected selectedTopic GenAI, got %s", m.insightsState.selectedTopic)
	}
}

func TestInsights_EscOnTopicTimelineReturnsToList(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabTopics
	model.insightsState.selectedTopic = "GenAI"
	model.insightsState.topicTimeline = []domain.InsightsTopicTimeline{
		{Topic: "GenAI", Content: "AI features", Importance: "high", WeekStart: "2026-01-13", WeekEnd: "2026-01-19"},
	}

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.selectedTopic != "" {
		t.Error("expected selectedTopic to be empty after escape")
	}
	if m.insightsState.topicTimeline != nil {
		t.Error("expected topicTimeline to be nil after escape")
	}
}

func TestInsights_RenderTopicTimeline(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabTopics
	model.insightsState.selectedTopic = "GenAI"
	model.insightsState.topicTimeline = []domain.InsightsTopicTimeline{
		{Topic: "GenAI", Content: "Integration planning for AI features", Importance: "high", WeekStart: "2026-01-13", WeekEnd: "2026-01-19"},
		{Topic: "GenAI", Content: "AI sprint completed", Importance: "medium", WeekStart: "2026-01-27", WeekEnd: "2026-02-02"},
	}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "GenAI") {
		t.Error("expected output to contain topic name")
	}
	if !strings.Contains(output, "Integration planning") {
		t.Error("expected output to contain timeline content")
	}
	if !strings.Contains(output, "2026-01-13") {
		t.Error("expected output to contain week start date")
	}
}

func TestInsights_JKNavigatesTopicsList(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabTopics
	model.insightsState.distinctTopics = []string{"GenAI", "Quarterly Planning", "Sprint Review"}
	model.insightsState.topicSelectedIdx = 0

	// j moves down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.topicSelectedIdx != 1 {
		t.Errorf("expected idx 1 after j, got %d", m.insightsState.topicSelectedIdx)
	}

	// j again
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.topicSelectedIdx != 2 {
		t.Errorf("expected idx 2 after second j, got %d", m.insightsState.topicSelectedIdx)
	}

	// j at bottom stays
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.topicSelectedIdx != 2 {
		t.Errorf("expected idx 2 (clamped), got %d", m.insightsState.topicSelectedIdx)
	}

	// k moves up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.topicSelectedIdx != 1 {
		t.Errorf("expected idx 1 after k, got %d", m.insightsState.topicSelectedIdx)
	}
}

func TestInsights_TabCycleIncludesDecisions(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabTopics

	msg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.activeTab != InsightsTabDecisions {
		t.Errorf("expected decisions tab after topics, got %d", m.insightsState.activeTab)
	}

	result, _ = m.Update(msg)
	m = result.(Model)

	if m.insightsState.activeTab != InsightsTabDashboard {
		t.Errorf("expected dashboard tab (wrap) after decisions, got %d", m.insightsState.activeTab)
	}
}

func TestInsights_DecisionsLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabDecisions
	model.insightsState.loading = true

	decisions := []domain.InsightsDecisionWithInitiatives{
		{
			InsightsDecision: domain.InsightsDecision{
				ID:           1,
				DecisionText: "Adopt new framework",
				Rationale:    "Better performance",
				Participants: "Team leads",
				DecisionDate: "2026-01-15",
			},
			Initiatives: "Platform Rewrite",
		},
	}

	result, _ := model.Update(insightsDecisionsLoadedMsg{decisions: decisions})
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to be false after decisions loaded")
	}
	if len(m.insightsState.decisions) != 1 {
		t.Errorf("expected 1 decision, got %d", len(m.insightsState.decisions))
	}
	if m.insightsState.decisions[0].DecisionText != "Adopt new framework" {
		t.Errorf("expected decision text, got %s", m.insightsState.decisions[0].DecisionText)
	}
}

func TestInsights_RenderDecisions(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabDecisions
	model.insightsState.decisions = []domain.InsightsDecisionWithInitiatives{
		{
			InsightsDecision: domain.InsightsDecision{
				ID:           1,
				DecisionText: "Adopt Claude",
				Rationale:    "Best AI",
				Participants: "Engineering",
				DecisionDate: "2026-01-15",
			},
			Initiatives: "AI Integration",
		},
	}

	view := model.View()

	if !strings.Contains(view, "Adopt Claude") {
		t.Error("expected view to contain decision text")
	}
	if !strings.Contains(view, "2026-01-15") {
		t.Error("expected view to contain decision date")
	}
	if !strings.Contains(view, "Engineering") {
		t.Error("expected view to contain participants")
	}
	if !strings.Contains(view, "AI Integration") {
		t.Error("expected view to contain linked initiatives")
	}
}

func TestInsights_RenderDecisionsEmpty(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabDecisions

	view := model.View()

	if !strings.Contains(view, "No decisions found") {
		t.Error("expected empty state message")
	}
}

func TestInsights_ViewTypeDisplayName(t *testing.T) {
	model := newInsightsModel()
	view := model.View()

	if !strings.Contains(view, "Insights") {
		t.Error("expected view to contain 'Insights' display name")
	}
}

func TestInsights_RKeyOnSummariesTriggersReportLoading(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries
	model.insightsState.weekSummary = &domain.InsightsSummary{
		WeekStart: "2026-01-27",
		WeekEnd:   "2026-02-02",
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	result, _ := model.Update(msg)
	m := result.(Model)

	if !m.insightsState.loading {
		t.Error("expected loading to be true after pressing r")
	}
}

func TestInsights_RKeyNoOpOnNonSummariesTab(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabDashboard
	initialWeek := model.insightsState.weekAnchor

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to remain false on non-summaries tab")
	}
	if !m.insightsState.weekAnchor.Equal(initialWeek) {
		t.Error("expected week anchor to remain unchanged")
	}
}

func TestInsights_WeeklyReportLoadedMsg(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries
	model.insightsState.loading = true

	report := &domain.InsightsWeeklyReport{
		Summary: &domain.InsightsSummary{
			ID:          1,
			WeekStart:   "2026-01-27",
			WeekEnd:     "2026-02-02",
			SummaryText: "Great week",
		},
		Topics: []domain.InsightsTopic{
			{ID: 1, Topic: "AI", Content: "Progress on AI", Importance: "high"},
		},
		InitiativeUpdates: []domain.InsightsInitiativeWeekUpdate{
			{InitiativeName: "GenAI Integration", UpdateText: "Sprint completed"},
		},
		Actions: []domain.InsightsAction{
			{ID: 1, ActionText: "Deploy model", Priority: "high", Status: "pending"},
		},
	}

	msg := insightsWeeklyReportLoadedMsg{report: report}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.loading {
		t.Error("expected loading to be false")
	}
	if m.insightsState.weeklyReport == nil {
		t.Error("expected weeklyReport to be set")
	}
	if m.insightsState.weeklyReport.Summary.SummaryText != "Great week" {
		t.Errorf("expected summary text, got %s", m.insightsState.weeklyReport.Summary.SummaryText)
	}
}

func TestInsights_EscOnWeeklyReportReturnsToSummaries(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries
	model.insightsState.weeklyReport = &domain.InsightsWeeklyReport{
		Summary: &domain.InsightsSummary{SummaryText: "Report"},
	}

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	result, _ := model.Update(msg)
	m := result.(Model)

	if m.insightsState.weeklyReport != nil {
		t.Error("expected weeklyReport to be nil after escape")
	}
}

func TestInsights_RenderWeeklyReport(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries
	model.insightsState.weekAnchor = time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC)
	model.insightsState.weeklyReport = &domain.InsightsWeeklyReport{
		Summary: &domain.InsightsSummary{
			WeekStart:   "2026-01-26",
			WeekEnd:     "2026-02-01",
			SummaryText: "Productive week with strong delivery",
		},
		Topics: []domain.InsightsTopic{
			{Topic: "AI Integration", Content: "Model deployed", Importance: "high"},
			{Topic: "Testing", Content: "Coverage improved", Importance: "medium"},
		},
		InitiativeUpdates: []domain.InsightsInitiativeWeekUpdate{
			{InitiativeName: "GenAI Integration", UpdateText: "Sprint 3 completed"},
		},
		Actions: []domain.InsightsAction{
			{ActionText: "Review deployment metrics", Priority: "high", Status: "pending"},
			{ActionText: "Update docs", Priority: "low", Status: "pending"},
		},
	}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "Weekly Report") {
		t.Error("expected output to contain Weekly Report header")
	}
	if !strings.Contains(output, "Productive week") {
		t.Error("expected output to contain summary text")
	}
	if !strings.Contains(output, "AI Integration") {
		t.Error("expected output to contain topic name")
	}
	if !strings.Contains(output, "Testing") {
		t.Error("expected output to contain second topic name")
	}
	if !strings.Contains(output, "GenAI Integration") {
		t.Error("expected output to contain initiative name")
	}
	if !strings.Contains(output, "Sprint 3 completed") {
		t.Error("expected output to contain initiative update text")
	}
	if !strings.Contains(output, "Review deployment metrics") {
		t.Error("expected output to contain action text")
	}
	if !strings.Contains(output, "Update docs") {
		t.Error("expected output to contain second action text")
	}
}

func TestInsights_RenderWeeklyReportEmpty(t *testing.T) {
	model := newInsightsModel()
	model.insightsState.activeTab = InsightsTabSummaries
	model.insightsState.weeklyReport = &domain.InsightsWeeklyReport{
		Summary: &domain.InsightsSummary{
			WeekStart:   "2026-01-26",
			WeekEnd:     "2026-02-01",
			SummaryText: "Quiet week",
		},
	}

	output := model.renderInsightsContent()

	if !strings.Contains(output, "Weekly Report") {
		t.Error("expected output to contain Weekly Report header")
	}
	if !strings.Contains(output, "Quiet week") {
		t.Error("expected output to contain summary text")
	}
}
