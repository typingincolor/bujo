# Insights Integration Design (MVP)

**Date:** 2026-02-04
**Spec:** docs/INTEGRATION-SPEC.md
**Branch:** feature/insights-integration

## Scope

MVP integration of `claude-insights.db` into the bujo desktop app. Three views behind a single "Insights" sidebar item with internal tabs:

1. **Dashboard** - Executive summary with latest summary, active initiatives, high-priority actions, recent decisions, staleness indicator
2. **Weekly Summaries** - Chronological list with expand/collapse and topic badges
3. **Action Items** - Pending actions sorted by priority/due date (read-only)

## Architecture

Two layers only: domain types + repository. No service layer (read-only data, no business logic).

### New Files

```
internal/
  domain/
    insights.go                         # Structs
  repository/sqlite/
    insights_repository.go              # Read-only queries
    insights_repository_test.go         # In-memory seeded DB tests

internal/adapter/wails/
    app.go                              # New methods added

frontend/src/components/bujo/
    InsightsView.tsx                    # Container with tabs
    InsightsDashboard.tsx              # Dashboard cards
    InsightsSummaries.tsx              # Weekly summary list
    InsightsActions.tsx                # Action items list
```

### Domain Types

```go
// internal/domain/insights.go

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
    Importance string // high, medium, low
}

type InsightsInitiative struct {
    ID          int64
    Name        string
    Status      string // active, planning, blocked, completed, on-hold
    Description string
    LastUpdated string
}

type InsightsAction struct {
    ID         int64
    SummaryID  int64
    ActionText string
    Priority   string // high, medium, low
    Status     string // pending, completed, blocked, cancelled
    DueDate    string
    CreatedAt  string
    WeekStart  string // denormalized from summaries join
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

type InsightsDashboard struct {
    LatestSummary        *InsightsSummary
    ActiveInitiatives    []InsightsInitiative
    HighPriorityActions  []InsightsAction
    RecentDecisions      []InsightsDecision
    DaysSinceLastSummary int
    Status               string // ready, not_initialized, empty
}
```

### Repository

```go
type InsightsRepository struct {
    db *sql.DB // nil when insights DB doesn't exist
}
```

**Factory helper:**
- `OpenInsightsDB()` returns `(*sql.DB, error)` — returns `nil, nil` if file doesn't exist
- Opens `~/bujo-companion/claude-insights.db` with `?mode=ro`
- Called from `ServiceFactory.Create()`

**Methods:**

| Method | Returns | Used By |
|--------|---------|---------|
| `GetLatestSummary(ctx)` | `*InsightsSummary` | Dashboard |
| `GetSummaries(ctx, limit)` | `[]InsightsSummary` | Summaries tab |
| `GetTopicsForSummary(ctx, summaryID)` | `[]InsightsTopic` | Summary detail |
| `GetActiveInitiatives(ctx, limit)` | `[]InsightsInitiative` | Dashboard |
| `GetPendingActions(ctx)` | `[]InsightsAction` | Dashboard + Actions tab |
| `GetRecentDecisions(ctx, limit)` | `[]InsightsDecision` | Dashboard |
| `GetDaysSinceLastSummary(ctx)` | `int` | Dashboard staleness |
| `IsAvailable()` | `bool` | Sidebar visibility |

Every method: `if r.db == nil { return empty }` as first line.

### Wails Adapter

New methods on `App` struct:

```go
func (a *App) GetInsightsDashboard() (*domain.InsightsDashboard, error)
func (a *App) GetInsightsSummaries(limit int) ([]domain.InsightsSummary, error)
func (a *App) GetInsightsSummaryDetail(summaryID int64) (*domain.InsightsSummary, []domain.InsightsTopic, error)
func (a *App) GetInsightsActions() ([]domain.InsightsAction, error)
func (a *App) IsInsightsAvailable() bool
```

`GetInsightsDashboard` composes multiple repository calls. Others are thin pass-throughs.

### Frontend

**InsightsView.tsx** - Container with three tabs (Dashboard, Summaries, Actions). Follows existing component patterns.

**InsightsDashboard.tsx** - Four cards:
- Latest weekly summary (truncated, expandable)
- Top active initiatives (name + status badge)
- High priority actions (text + due date)
- Recent decisions (text + date)
- Staleness indicator ("Last summary: Jan 27 - Feb 2, 3 days ago")

**InsightsSummaries.tsx** - Scrollable list, week range titles, expand to full markdown, topic badges on each summary.

**InsightsActions.tsx** - Pending actions by priority then due date. Priority badges. Due date highlighting for overdue. Read-only in v1.

**Sidebar:** `IsInsightsAvailable()` controls visibility. Hidden when DB doesn't exist (no onboarding UI in v1).

## Testing

- In-memory SQLite DB with full insights schema from spec appendix
- Test helper seeds 2-3 weeks of realistic data
- Each test gets fresh seeded DB
- Test nil DB path for graceful degradation
- TDD: all repository methods test-driven

## Deprecation: Existing AI/Gemini Summary Functionality

This integration replaces the existing Gemini-based AI summaries which never worked properly. Files to remove:

**Go source (15 files):**
- `internal/adapter/ai/` — entire directory (client.go, gemini.go, generator.go, config.go, prompt_loader.go, local/client.go, prompts/*.txt)
- `internal/domain/summary.go` — Summary, SummaryHorizon types
- `internal/domain/prompt.go` — PromptType, PromptTemplate types
- `internal/repository/sqlite/summary_repository.go`
- `internal/service/summary.go` — SummaryService
- `cmd/bujo/cmd/summary.go` — CLI summary command

**Test files (7):**
- `internal/adapter/ai/gemini_test.go`, `config_test.go`, `prompt_loader_test.go`
- `internal/domain/summary_test.go`, `prompt_test.go`
- `internal/repository/sqlite/summary_repository_test.go`
- `internal/service/summary_test.go`

**References to update:**
- `internal/app/factory.go` — remove Summary field from Services
- `internal/adapter/wails/app.go` — remove GetSummary method
- `internal/domain/repository.go` — remove SummaryRepository interface
- `cmd/bujo/cmd/root.go` — remove summaryService, summaryRepo, AI client init
- `internal/tui/` — remove summary state, messages, rendering (model.go, messages.go, update.go, view.go)
- `internal/service/export.go` — remove ExportSummaryRepository usage

**Database:** New migration to drop `summaries` table (migration 000005)

**Docs to update:** AI_SETUP.md, ARCHITECTURE.md, CLI.md, WORKFLOWS.md

**Environment variables removed:** BUJO_AI_ENABLED, BUJO_AI_PROVIDER, GEMINI_API_KEY, BUJO_MODEL

## Not In Scope (Future)

- Topic timeline view
- Initiative portfolio/detail views
- Decision log
- Weekly report generator
- Onboarding UI when DB missing
- Write-back to insights DB (marking actions done)
- Schema version migration handling
