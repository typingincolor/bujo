package domain

type InsightsSummary struct {
	ID          int64
	WeekStart   string
	WeekEnd     string
	SummaryText string
	CreatedAt   string
}

type InsightsTopic struct {
	ID         int64
	SummaryID  int64
	Topic      string
	Content    string
	Importance string
}

type InsightsInitiative struct {
	ID          int64
	Name        string
	Status      string
	Description string
	LastUpdated string
}

type InsightsAction struct {
	ID         int64
	SummaryID  int64
	ActionText string
	Priority   string
	Status     string
	DueDate    string
	CreatedAt  string
	WeekStart  string
}

func (a InsightsAction) IsOverdue(today string) bool {
	if a.DueDate == "" || a.Status != "pending" {
		return false
	}
	return a.DueDate < today
}

type InsightsDecision struct {
	ID               int64
	DecisionText     string
	Rationale        string
	Participants     string
	ExpectedOutcomes string
	DecisionDate     string
	SummaryID        *int64
	CreatedAt        string
}

type InsightsDecisionWithInitiatives struct {
	InsightsDecision
	Initiatives string
}

type InsightsInitiativePortfolio struct {
	ID              int64
	Name            string
	Status          string
	Description     string
	LastUpdated     string
	MentionCount    int
	LastMentionWeek string
	ActivityWeeks   string
}

type InsightsInitiativeUpdate struct {
	WeekStart  string
	WeekEnd    string
	UpdateText string
}

type InsightsInitiativeDetail struct {
	Initiative     InsightsInitiative
	Updates        []InsightsInitiativeUpdate
	PendingActions []InsightsAction
	Decisions      []InsightsDecision
}

type InsightsTopicTimeline struct {
	Topic      string
	Content    string
	Importance string
	WeekStart  string
	WeekEnd    string
}

type InsightsInitiativeWeekUpdate struct {
	InitiativeName string
	UpdateText     string
}

type InsightsWeeklyReport struct {
	Summary           *InsightsSummary
	Topics            []InsightsTopic
	InitiativeUpdates []InsightsInitiativeWeekUpdate
	Actions           []InsightsAction
}

type InsightsDashboard struct {
	LatestSummary        *InsightsSummary
	ActiveInitiatives    []InsightsInitiative
	HighPriorityActions  []InsightsAction
	RecentDecisions      []InsightsDecision
	DaysSinceLastSummary int
	Status               string
}
