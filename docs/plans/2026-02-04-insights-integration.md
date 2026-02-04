# Insights Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate read-only `claude-insights.db` into bujo desktop app with Dashboard, Summaries, and Actions views, while deprecating the existing broken Gemini AI summary system.

**Architecture:** Two-layer (domain types + repository). No service layer since all data is read-only. Repository returns empty results when insights DB is missing. Frontend uses three-tab InsightsView accessed from sidebar.

**Tech Stack:** Go 1.23, SQLite (read-only), Wails v2, React/TypeScript, Tailwind CSS

---

## Phase 1: Deprecation — Remove Existing AI/Gemini Summary System

### Task 1: Remove AI Adapter Directory

**Files:**
- Delete: `internal/adapter/ai/` (entire directory — 15 files)

**Step 1: Delete the AI adapter directory**

```bash
rm -rf internal/adapter/ai/
```

**Step 2: Verify no other code imports it**

Run: `grep -r '"github.com/typingincolor/bujo/internal/adapter/ai"' --include="*.go" .`
Expected: Only `cmd/bujo/cmd/root.go` (handled in Task 4)

**Step 3: Commit**

```bash
git add -A
git commit -m "chore: remove AI adapter directory (Gemini/Ollama integration)"
```

---

### Task 2: Remove Domain Summary Types

**Files:**
- Delete: `internal/domain/summary.go`
- Delete: `internal/domain/summary_test.go`
- Delete: `internal/domain/prompt.go`
- Delete: `internal/domain/prompt_test.go`
- Modify: `internal/domain/repository.go:59-66` — remove SummaryRepository interface
- Modify: `internal/domain/export.go:14` — remove Summaries field from ExportData

**Step 1: Delete summary and prompt source + test files**

```bash
rm internal/domain/summary.go internal/domain/summary_test.go
rm internal/domain/prompt.go internal/domain/prompt_test.go
```

**Step 2: Remove SummaryRepository interface from repository.go**

In `internal/domain/repository.go`, delete lines 59-66:

```go
// DELETE THIS BLOCK:
type SummaryRepository interface {
	Insert(ctx context.Context, summary Summary) (int64, error)
	Get(ctx context.Context, horizon SummaryHorizon, start, end time.Time) (*Summary, error)
	GetByHorizon(ctx context.Context, horizon SummaryHorizon) ([]Summary, error)
	GetAll(ctx context.Context) ([]Summary, error)
	Delete(ctx context.Context, id int64) error
	DeleteAll(ctx context.Context) error
}
```

**Step 3: Remove Summaries field from ExportData**

In `internal/domain/export.go`, remove the `Summaries []Summary` line from the ExportData struct.

Before:
```go
type ExportData struct {
	Version     string       `json:"version"`
	ExportedAt  time.Time    `json:"exported_at"`
	Entries     []Entry      `json:"entries"`
	Habits      []Habit      `json:"habits"`
	HabitLogs   []HabitLog   `json:"habit_logs"`
	DayContexts []DayContext `json:"day_contexts"`
	Summaries   []Summary    `json:"summaries"`
	Lists       []List       `json:"lists"`
	ListItems   []ListItem   `json:"list_items"`
	Goals       []Goal       `json:"goals"`
}
```

After:
```go
type ExportData struct {
	Version     string       `json:"version"`
	ExportedAt  time.Time    `json:"exported_at"`
	Entries     []Entry      `json:"entries"`
	Habits      []Habit      `json:"habits"`
	HabitLogs   []HabitLog   `json:"habit_logs"`
	DayContexts []DayContext `json:"day_contexts"`
	Lists       []List       `json:"lists"`
	ListItems   []ListItem   `json:"list_items"`
	Goals       []Goal       `json:"goals"`
}
```

**Step 4: Verify domain tests pass**

Run: `go test ./internal/domain/... -v`
Expected: PASS (no summary-related tests remain)

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: remove Summary and Prompt domain types"
```

---

### Task 3: Remove Summary Repository Implementation

**Files:**
- Delete: `internal/repository/sqlite/summary_repository.go`
- Delete: `internal/repository/sqlite/summary_repository_test.go`

**Step 1: Delete repository files**

```bash
rm internal/repository/sqlite/summary_repository.go
rm internal/repository/sqlite/summary_repository_test.go
```

**Step 2: Verify repository tests pass**

Run: `go test ./internal/repository/sqlite/... -v`
Expected: PASS

**Step 3: Commit**

```bash
git add -A
git commit -m "chore: remove SQLite summary repository"
```

---

### Task 4: Remove Summary Service

**Files:**
- Delete: `internal/service/summary.go`
- Delete: `internal/service/summary_test.go`
- Modify: `internal/app/factory.go:20` — remove Summary field from Services struct

**Step 1: Delete service files**

```bash
rm internal/service/summary.go internal/service/summary_test.go
```

**Step 2: Remove Summary field from Services struct**

In `internal/app/factory.go`, remove line 20 (`Summary *service.SummaryService`):

Before:
```go
type Services struct {
	DB              *sql.DB
	Bujo            *service.BujoService
	Habit           *service.HabitService
	List            *service.ListService
	Goal            *service.GoalService
	Stats           *service.StatsService
	Summary         *service.SummaryService
	ChangeDetection *service.ChangeDetectionService
	EditableView    *service.EditableViewService
}
```

After:
```go
type Services struct {
	DB              *sql.DB
	Bujo            *service.BujoService
	Habit           *service.HabitService
	List            *service.ListService
	Goal            *service.GoalService
	Stats           *service.StatsService
	ChangeDetection *service.ChangeDetectionService
	EditableView    *service.EditableViewService
}
```

**Step 3: Verify service tests pass**

Run: `go test ./internal/service/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove SummaryService and Services.Summary field"
```

---

### Task 5: Remove Summary from Export/Import Services

**Files:**
- Modify: `internal/service/export.go` — remove all summary repository references
- Modify: `internal/service/export_test.go` — remove mock summary repositories

**Step 1: Remove summary from ExportService**

In `internal/service/export.go`:

1. Delete `ExportSummaryRepository` interface (lines 27-29)
2. Delete `summaryRepo` field from `ExportService` struct (line 48)
3. Remove `summaryRepo ExportSummaryRepository` parameter from `NewExportService` (line 59) and its assignment in the return struct (line 69)
4. Delete the `data.Summaries` block in `Export` method (lines 314-320)
5. Delete `ImportSummaryRepository` interface (lines 97-100)
6. Delete `summaryRepo` field from `ImportService` struct (line 124)
7. Remove `summaryRepo ImportSummaryRepository` parameter from `NewImportService` (line 135) and its assignment (line 145)
8. Delete the `data.Summaries` loop in `Import` method (lines 194-198)
9. Delete `s.summaryRepo.DeleteAll` call in `clearAllData` (line 253-255)

**Step 2: Update export_test.go**

Remove `mockSummaryRepoForExport` and `mockSummaryRepoForImport` mock types. Remove `summaryRepo` parameter from all `NewExportService` and `NewImportService` calls in tests.

**Step 3: Run export tests**

Run: `go test ./internal/service/... -v -run TestExport`
Expected: PASS

Run: `go test ./internal/service/... -v -run TestImport`
Expected: PASS

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove summary from export/import services"
```

---

### Task 6: Remove Summary from CLI and Wails Adapter

**Files:**
- Delete: `cmd/bujo/cmd/summary.go` — CLI summary command
- Modify: `cmd/bujo/cmd/root.go` — remove summaryService, summaryRepo, AI client init
- Modify: `internal/adapter/wails/app.go:361-373` — remove GetSummary method

**Step 1: Delete CLI summary command**

```bash
rm cmd/bujo/cmd/summary.go
```

**Step 2: Update root.go**

In `cmd/bujo/cmd/root.go`:

1. Remove `"github.com/typingincolor/bujo/internal/adapter/ai"` import (line 13)
2. Remove `summaryService *service.SummaryService` variable (line 32)
3. Remove `summaryRepo` creation and usage (lines 97-118):
   - Delete `summaryRepo := sqlite.NewSummaryRepository(db)` (line 97)
   - Remove `summaryRepo` from `NewExportService` call (line 100) — becomes:
     ```go
     exportService = service.NewExportService(
         entryRepo, habitRepo, habitLogRepo, dayCtxRepo,
         listRepo, listItemRepo, goalRepo,
     )
     ```
   - Remove `summaryRepo` from `NewImportService` call (line 104) — becomes:
     ```go
     importService = service.NewImportService(
         entryRepo, habitRepo, habitLogRepo, dayCtxRepo,
         listRepo, listItemRepo, goalRepo,
     )
     ```
   - Delete AI client init block (lines 106-118)
4. Delete `getDefaultPromptsDir` function (lines 159-166)

**Step 3: Remove GetSummary from Wails adapter**

In `internal/adapter/wails/app.go`, delete the GetSummary method (lines 361-373):

```go
// DELETE THIS ENTIRE METHOD:
func (a *App) GetSummary(date time.Time) (string, error) {
	if a.services.Summary == nil {
		return "", nil
	}
	summary, err := a.services.Summary.GetSummary(a.ctx, domain.SummaryHorizonDaily, date)
	if err != nil {
		return "", err
	}
	if summary == nil {
		return "", nil
	}
	return summary.Content, nil
}
```

**Step 4: Verify full build**

Run: `go build ./...`
Expected: BUILD SUCCESS

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: remove summary CLI command, AI init from root, GetSummary from Wails"
```

---

### Task 7: Remove Summary from TUI

**Files:**
- Modify: `internal/tui/model.go` — remove summary-related fields, config, methods
- Modify: `internal/tui/messages.go` — remove summary message types
- Modify: `internal/tui/update.go` — remove summary update handlers
- Modify: `internal/tui/view.go` — remove summary rendering

**Step 1: Clean summary references from TUI**

In `internal/tui/model.go`:
- Remove `SummaryService *service.SummaryService` from Config struct (line 38)
- Remove `summaryService *service.SummaryService` from Model struct (line 49)
- Remove `summaryState summaryState` and `summaryCollapsed bool` fields (lines 86-87)
- Remove `summaryState` type definition (around line 257)
- Remove `summaryService: cfg.SummaryService` from New function (around line 398)
- Remove `summaryState` and `summaryCollapsed` initialization (around lines 416-417)
- Remove `loadSummaryCmd` method and related streaming types (around lines 970-1038)

In `internal/tui/messages.go`:
- Remove `summaryLoadedMsg`, `summaryTokenMsg`, `summaryErrorMsg` message types

In `internal/tui/update.go`:
- Remove case handlers for summary messages

In `internal/tui/view.go`:
- Remove summary rendering section

**Step 2: Run TUI tests**

Run: `go test ./internal/tui/... -v`
Expected: PASS

**Step 3: Run full test suite**

Run: `go test ./... -count=1`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove summary state, messages, and rendering from TUI"
```

---

### Task 8: Add Migration to Drop Summaries Table

**Files:**
- Create: `internal/repository/sqlite/migrations/000031_drop_summaries.up.sql`
- Create: `internal/repository/sqlite/migrations/000031_drop_summaries.down.sql`

**Step 1: Create up migration**

```sql
-- 000031_drop_summaries.up.sql
DROP TABLE IF EXISTS summaries;
```

**Step 2: Create down migration**

```sql
-- 000031_drop_summaries.down.sql
CREATE TABLE IF NOT EXISTS summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    horizon TEXT NOT NULL CHECK(horizon IN ('daily', 'weekly')),
    content TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_to DATETIME,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK(op_type IN ('INSERT', 'UPDATE', 'DELETE'))
);
```

**Step 3: Verify migration runs**

Run: `go test ./internal/repository/sqlite/... -v -run TestMigration`
Expected: PASS (or run the app with a test DB to verify migration applies)

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: add migration 000031 to drop summaries table"
```

---

## Phase 2: Domain Types — Insights Data Model

### Task 9: Write Failing Tests for Insights Domain Types

**Files:**
- Create: `internal/domain/insights_test.go`

**Step 1: Write the failing test**

RED: Writing failing test for InsightsDashboard.Status() method.

```go
package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsightsDashboard_StatusReady(t *testing.T) {
	d := InsightsDashboard{
		LatestSummary: &InsightsSummary{ID: 1, WeekStart: "2026-01-27"},
	}
	assert.Equal(t, "ready", d.Status)
}

func TestInsightsDashboard_StatusEmpty(t *testing.T) {
	d := InsightsDashboard{}
	assert.Equal(t, "empty", d.Status)
}

func TestInsightsAction_IsOverdue(t *testing.T) {
	tests := []struct {
		name    string
		action  InsightsAction
		today   string
		want    bool
	}{
		{
			name:   "overdue action",
			action: InsightsAction{DueDate: "2026-01-15", Status: "pending"},
			today:  "2026-02-04",
			want:   true,
		},
		{
			name:   "future action",
			action: InsightsAction{DueDate: "2026-03-01", Status: "pending"},
			today:  "2026-02-04",
			want:   false,
		},
		{
			name:   "no due date",
			action: InsightsAction{DueDate: "", Status: "pending"},
			today:  "2026-02-04",
			want:   false,
		},
		{
			name:   "completed action not overdue",
			action: InsightsAction{DueDate: "2026-01-15", Status: "completed"},
			today:  "2026-02-04",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.action.IsOverdue(tt.today))
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/... -v -run TestInsights`
Expected: FAIL — `InsightsDashboard`, `InsightsSummary`, `InsightsAction` undefined

**Step 3: Commit**

```bash
git add -A
git commit -m "test: add failing tests for insights domain types"
```

---

### Task 10: Implement Insights Domain Types

**Files:**
- Create: `internal/domain/insights.go`

**Step 1: Implement the types**

GREEN: Implementing insights domain types.

```go
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

type InsightsDashboard struct {
	LatestSummary        *InsightsSummary
	ActiveInitiatives    []InsightsInitiative
	HighPriorityActions  []InsightsAction
	RecentDecisions      []InsightsDecision
	DaysSinceLastSummary int
	Status               string
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/domain/... -v -run TestInsights`
Expected: PASS

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add insights domain types"
```

---

## Phase 3: Repository — Insights Data Access

### Task 11: Create Test Helper for Insights DB

**Files:**
- Create: `internal/repository/sqlite/insights_test_helper_test.go`

**Step 1: Write the test helper**

This helper creates an in-memory SQLite database with the full insights schema and seeds it with realistic test data.

```go
package sqlite

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func setupInsightsTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	schema := `
		CREATE TABLE summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			week_start TEXT NOT NULL,
			week_end TEXT NOT NULL,
			summary_text TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE topics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			summary_id INTEGER NOT NULL,
			topic TEXT NOT NULL,
			content TEXT,
			importance TEXT CHECK(importance IN ('high', 'medium', 'low')),
			FOREIGN KEY (summary_id) REFERENCES summaries(id)
		);

		CREATE TABLE initiatives (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			status TEXT CHECK(status IN ('active', 'planning', 'blocked', 'completed', 'on-hold')),
			description TEXT,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE initiative_mentions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			summary_id INTEGER NOT NULL,
			initiative_id INTEGER NOT NULL,
			update_text TEXT,
			FOREIGN KEY (summary_id) REFERENCES summaries(id),
			FOREIGN KEY (initiative_id) REFERENCES initiatives(id)
		);

		CREATE TABLE actions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			summary_id INTEGER NOT NULL,
			action_text TEXT NOT NULL,
			priority TEXT CHECK(priority IN ('high', 'medium', 'low')),
			status TEXT CHECK(status IN ('pending', 'completed', 'blocked', 'cancelled')),
			due_date TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (summary_id) REFERENCES summaries(id)
		);

		CREATE TABLE decisions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			decision_text TEXT NOT NULL,
			rationale TEXT,
			participants TEXT,
			expected_outcomes TEXT,
			decision_date TEXT NOT NULL,
			summary_id INTEGER,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (summary_id) REFERENCES summaries(id)
		);

		CREATE TABLE decision_initiatives (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			decision_id INTEGER NOT NULL,
			initiative_id INTEGER NOT NULL,
			FOREIGN KEY (decision_id) REFERENCES decisions(id),
			FOREIGN KEY (initiative_id) REFERENCES initiatives(id)
		);

		CREATE TABLE metadata (
			key TEXT PRIMARY KEY,
			value TEXT
		);

		INSERT INTO metadata (key, value) VALUES ('version', '1.1');
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	seedInsightsData(t, db)
	return db
}

func seedInsightsData(t *testing.T, db *sql.DB) {
	t.Helper()

	// 3 weeks of summaries
	_, err := db.Exec(`
		INSERT INTO summaries (id, week_start, week_end, summary_text, created_at) VALUES
		(1, '2026-01-13', '2026-01-19', 'Week of Jan 13: Focused on GenAI integration and team planning.', '2026-01-20 09:00:00'),
		(2, '2026-01-20', '2026-01-26', 'Week of Jan 20: Major progress on tech scorecard. Team retrospective.', '2026-01-27 09:00:00'),
		(3, '2026-01-27', '2026-02-02', 'Week of Jan 27: Sprint completion and quarterly planning kickoff.', '2026-02-03 09:00:00')
	`)
	require.NoError(t, err)

	// Topics
	_, err = db.Exec(`
		INSERT INTO topics (summary_id, topic, content, importance) VALUES
		(1, 'GenAI', 'Integration planning for AI features', 'high'),
		(1, 'Team Planning', 'Q1 roadmap discussion', 'medium'),
		(2, 'Tech Scorecard', 'Completed initial assessment', 'high'),
		(3, 'Quarterly Planning', 'Q2 goals defined', 'high'),
		(3, 'Sprint Review', 'All stories completed', 'medium')
	`)
	require.NoError(t, err)

	// Initiatives
	_, err = db.Exec(`
		INSERT INTO initiatives (id, name, status, description, last_updated) VALUES
		(1, 'GenAI Integration', 'active', 'Integrate AI capabilities into core platform', '2026-01-27'),
		(2, 'Tech Scorecard', 'active', 'Technology assessment framework', '2026-01-26'),
		(3, 'Q1 OKRs', 'completed', 'First quarter objectives', '2026-01-20')
	`)
	require.NoError(t, err)

	// Initiative mentions
	_, err = db.Exec(`
		INSERT INTO initiative_mentions (summary_id, initiative_id, update_text) VALUES
		(1, 1, 'Started planning AI integration approach'),
		(2, 2, 'Completed tech scorecard assessment'),
		(3, 1, 'AI integration sprint completed')
	`)
	require.NoError(t, err)

	// Actions
	_, err = db.Exec(`
		INSERT INTO actions (summary_id, action_text, priority, status, due_date, created_at) VALUES
		(1, 'Review AI vendor proposals', 'high', 'completed', '2026-01-20', '2026-01-20 09:00:00'),
		(2, 'Schedule tech scorecard review', 'medium', 'pending', '2026-02-10', '2026-01-27 09:00:00'),
		(3, 'Prepare Q2 planning materials', 'high', 'pending', '2026-02-05', '2026-02-03 09:00:00'),
		(3, 'Update team onboarding docs', 'low', 'pending', NULL, '2026-02-03 09:00:00')
	`)
	require.NoError(t, err)

	// Decisions
	_, err = db.Exec(`
		INSERT INTO decisions (decision_text, rationale, participants, expected_outcomes, decision_date, summary_id, created_at) VALUES
		('Adopt Claude as primary AI provider', 'Best performance on code tasks', 'Engineering team', 'Improved developer productivity', '2026-01-15', 1, '2026-01-20 09:00:00'),
		('Move to biweekly sprints', 'Better planning cadence', 'Team leads', 'More predictable delivery', '2026-01-28', 3, '2026-02-03 09:00:00')
	`)
	require.NoError(t, err)
}
```

**Step 2: Verify the helper compiles**

Run: `go test ./internal/repository/sqlite/... -v -run TestNothing -count=1`
Expected: PASS (no tests match, but compilation succeeds)

**Step 3: Commit**

```bash
git add -A
git commit -m "test: add insights test helper with in-memory DB and seed data"
```

---

### Task 12: Write Failing Tests for InsightsRepository

**Files:**
- Create: `internal/repository/sqlite/insights_repository_test.go`

**Step 1: Write failing tests**

RED: Writing failing tests for all InsightsRepository methods.

```go
package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsightsRepository_IsAvailable(t *testing.T) {
	t.Run("available when db is set", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)
		assert.True(t, repo.IsAvailable())
	})

	t.Run("not available when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		assert.False(t, repo.IsAvailable())
	})
}

func TestInsightsRepository_GetLatestSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("returns most recent summary", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)

		summary, err := repo.GetLatestSummary(ctx)
		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Equal(t, "2026-01-27", summary.WeekStart)
		assert.Equal(t, "2026-02-02", summary.WeekEnd)
		assert.Contains(t, summary.SummaryText, "Jan 27")
	})

	t.Run("returns nil when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		summary, err := repo.GetLatestSummary(ctx)
		require.NoError(t, err)
		assert.Nil(t, summary)
	})
}

func TestInsightsRepository_GetSummaries(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns summaries ordered by week_start desc", func(t *testing.T) {
		summaries, err := repo.GetSummaries(ctx, 10)
		require.NoError(t, err)
		require.Len(t, summaries, 3)
		assert.Equal(t, "2026-01-27", summaries[0].WeekStart)
		assert.Equal(t, "2026-01-20", summaries[1].WeekStart)
		assert.Equal(t, "2026-01-13", summaries[2].WeekStart)
	})

	t.Run("respects limit", func(t *testing.T) {
		summaries, err := repo.GetSummaries(ctx, 2)
		require.NoError(t, err)
		require.Len(t, summaries, 2)
	})

	t.Run("returns empty when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		summaries, err := repo.GetSummaries(ctx, 10)
		require.NoError(t, err)
		assert.Empty(t, summaries)
	})
}

func TestInsightsRepository_GetTopicsForSummary(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns topics for summary", func(t *testing.T) {
		topics, err := repo.GetTopicsForSummary(ctx, 1)
		require.NoError(t, err)
		require.Len(t, topics, 2)
		assert.Equal(t, "GenAI", topics[0].Topic)
	})

	t.Run("returns empty for non-existent summary", func(t *testing.T) {
		topics, err := repo.GetTopicsForSummary(ctx, 999)
		require.NoError(t, err)
		assert.Empty(t, topics)
	})
}

func TestInsightsRepository_GetActiveInitiatives(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns only active initiatives", func(t *testing.T) {
		initiatives, err := repo.GetActiveInitiatives(ctx, 10)
		require.NoError(t, err)
		require.Len(t, initiatives, 2)
		for _, i := range initiatives {
			assert.Equal(t, "active", i.Status)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		initiatives, err := repo.GetActiveInitiatives(ctx, 1)
		require.NoError(t, err)
		require.Len(t, initiatives, 1)
	})
}

func TestInsightsRepository_GetPendingActions(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns pending actions sorted by priority then due_date", func(t *testing.T) {
		actions, err := repo.GetPendingActions(ctx)
		require.NoError(t, err)
		require.Len(t, actions, 3) // 3 pending actions in seed data
		assert.Equal(t, "high", actions[0].Priority)
		assert.Equal(t, "pending", actions[0].Status)
	})

	t.Run("includes week_start from joined summary", func(t *testing.T) {
		actions, err := repo.GetPendingActions(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, actions)
		assert.NotEmpty(t, actions[0].WeekStart)
	})
}

func TestInsightsRepository_GetRecentDecisions(t *testing.T) {
	ctx := context.Background()
	db := setupInsightsTestDB(t)
	repo := NewInsightsRepository(db)

	t.Run("returns decisions ordered by date desc", func(t *testing.T) {
		decisions, err := repo.GetRecentDecisions(ctx, 10)
		require.NoError(t, err)
		require.Len(t, decisions, 2)
		assert.Equal(t, "2026-01-28", decisions[0].DecisionDate)
	})

	t.Run("respects limit", func(t *testing.T) {
		decisions, err := repo.GetRecentDecisions(ctx, 1)
		require.NoError(t, err)
		require.Len(t, decisions, 1)
	})
}

func TestInsightsRepository_GetDaysSinceLastSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("returns positive number for seeded data", func(t *testing.T) {
		db := setupInsightsTestDB(t)
		repo := NewInsightsRepository(db)
		days, err := repo.GetDaysSinceLastSummary(ctx)
		require.NoError(t, err)
		assert.Greater(t, days, 0)
	})

	t.Run("returns -1 when db is nil", func(t *testing.T) {
		repo := NewInsightsRepository(nil)
		days, err := repo.GetDaysSinceLastSummary(ctx)
		require.NoError(t, err)
		assert.Equal(t, -1, days)
	})
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/repository/sqlite/... -v -run TestInsightsRepository`
Expected: FAIL — `NewInsightsRepository` undefined

**Step 3: Commit**

```bash
git add -A
git commit -m "test: add failing tests for InsightsRepository"
```

---

### Task 13: Implement InsightsRepository

**Files:**
- Create: `internal/repository/sqlite/insights_repository.go`

**Step 1: Implement the repository**

GREEN: Implementing InsightsRepository with all queried methods.

```go
package sqlite

import (
	"context"
	"database/sql"

	"github.com/typingincolor/bujo/internal/domain"
)

type InsightsRepository struct {
	db *sql.DB
}

func NewInsightsRepository(db *sql.DB) *InsightsRepository {
	return &InsightsRepository{db: db}
}

func (r *InsightsRepository) IsAvailable() bool {
	return r.db != nil
}

func (r *InsightsRepository) GetLatestSummary(ctx context.Context) (*domain.InsightsSummary, error) {
	if r.db == nil {
		return nil, nil
	}

	row := r.db.QueryRowContext(ctx,
		`SELECT id, week_start, week_end, summary_text, created_at
		 FROM summaries ORDER BY week_start DESC LIMIT 1`)

	var s domain.InsightsSummary
	err := row.Scan(&s.ID, &s.WeekStart, &s.WeekEnd, &s.SummaryText, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *InsightsRepository) GetSummaries(ctx context.Context, limit int) ([]domain.InsightsSummary, error) {
	if r.db == nil {
		return []domain.InsightsSummary{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, week_start, week_end, summary_text, created_at
		 FROM summaries ORDER BY week_start DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.InsightsSummary
	for rows.Next() {
		var s domain.InsightsSummary
		if err := rows.Scan(&s.ID, &s.WeekStart, &s.WeekEnd, &s.SummaryText, &s.CreatedAt); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	if summaries == nil {
		summaries = []domain.InsightsSummary{}
	}
	return summaries, rows.Err()
}

func (r *InsightsRepository) GetTopicsForSummary(ctx context.Context, summaryID int64) ([]domain.InsightsTopic, error) {
	if r.db == nil {
		return []domain.InsightsTopic{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, summary_id, topic, content, importance
		 FROM topics WHERE summary_id = ?
		 ORDER BY CASE importance
			WHEN 'high' THEN 1
			WHEN 'medium' THEN 2
			WHEN 'low' THEN 3
		 END`, summaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []domain.InsightsTopic
	for rows.Next() {
		var t domain.InsightsTopic
		var content sql.NullString
		if err := rows.Scan(&t.ID, &t.SummaryID, &t.Topic, &content, &t.Importance); err != nil {
			return nil, err
		}
		t.Content = content.String
		topics = append(topics, t)
	}
	if topics == nil {
		topics = []domain.InsightsTopic{}
	}
	return topics, rows.Err()
}

func (r *InsightsRepository) GetActiveInitiatives(ctx context.Context, limit int) ([]domain.InsightsInitiative, error) {
	if r.db == nil {
		return []domain.InsightsInitiative{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, status, description, last_updated
		 FROM initiatives WHERE status = 'active'
		 ORDER BY last_updated DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var initiatives []domain.InsightsInitiative
	for rows.Next() {
		var i domain.InsightsInitiative
		var desc sql.NullString
		if err := rows.Scan(&i.ID, &i.Name, &i.Status, &desc, &i.LastUpdated); err != nil {
			return nil, err
		}
		i.Description = desc.String
		initiatives = append(initiatives, i)
	}
	if initiatives == nil {
		initiatives = []domain.InsightsInitiative{}
	}
	return initiatives, rows.Err()
}

func (r *InsightsRepository) GetPendingActions(ctx context.Context) ([]domain.InsightsAction, error) {
	if r.db == nil {
		return []domain.InsightsAction{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.summary_id, a.action_text, a.priority, a.status,
		        COALESCE(a.due_date, ''), a.created_at, s.week_start
		 FROM actions a
		 JOIN summaries s ON a.summary_id = s.id
		 WHERE a.status = 'pending'
		 ORDER BY
			CASE a.priority
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			a.due_date ASC NULLS LAST,
			s.week_start DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []domain.InsightsAction
	for rows.Next() {
		var a domain.InsightsAction
		if err := rows.Scan(&a.ID, &a.SummaryID, &a.ActionText, &a.Priority,
			&a.Status, &a.DueDate, &a.CreatedAt, &a.WeekStart); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	if actions == nil {
		actions = []domain.InsightsAction{}
	}
	return actions, rows.Err()
}

func (r *InsightsRepository) GetRecentDecisions(ctx context.Context, limit int) ([]domain.InsightsDecision, error) {
	if r.db == nil {
		return []domain.InsightsDecision{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, decision_text, rationale, participants,
		        expected_outcomes, decision_date, summary_id, created_at
		 FROM decisions
		 ORDER BY decision_date DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []domain.InsightsDecision
	for rows.Next() {
		var d domain.InsightsDecision
		var rationale, participants, outcomes sql.NullString
		if err := rows.Scan(&d.ID, &d.DecisionText, &rationale, &participants,
			&outcomes, &d.DecisionDate, &d.SummaryID, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.Rationale = rationale.String
		d.Participants = participants.String
		d.ExpectedOutcomes = outcomes.String
		decisions = append(decisions, d)
	}
	if decisions == nil {
		decisions = []domain.InsightsDecision{}
	}
	return decisions, rows.Err()
}

func (r *InsightsRepository) GetDaysSinceLastSummary(ctx context.Context) (int, error) {
	if r.db == nil {
		return -1, nil
	}

	row := r.db.QueryRowContext(ctx,
		`SELECT CAST(julianday('now') - julianday(MAX(week_start)) AS INTEGER)
		 FROM summaries`)

	var days sql.NullInt64
	if err := row.Scan(&days); err != nil {
		return -1, nil
	}
	if !days.Valid {
		return -1, nil
	}
	return int(days.Int64), nil
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/repository/sqlite/... -v -run TestInsightsRepository`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: implement InsightsRepository with read-only queries"
```

---

### Task 14: Add OpenInsightsDB Helper to Factory

**Files:**
- Modify: `internal/app/factory.go` — add InsightsRepo field, OpenInsightsDB function, wire in Create

**Step 1: Write failing test for OpenInsightsDB**

Create `internal/app/insights_db_test.go`:

```go
package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenInsightsDB_MissingFile(t *testing.T) {
	db, err := OpenInsightsDB("/nonexistent/path/insights.db")
	assert.NoError(t, err)
	assert.Nil(t, db)
}
```

Run: `go test ./internal/app/... -v -run TestOpenInsightsDB`
Expected: FAIL — `OpenInsightsDB` undefined

**Step 2: Implement OpenInsightsDB and wire into factory**

Create `internal/app/insights_db.go`:

```go
package app

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func OpenInsightsDB(dbPath string) (*sql.DB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, nil
	}

	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func DefaultInsightsDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/bujo-companion/claude-insights.db"
}
```

**Step 3: Add InsightsRepo to Services struct and wire in factory**

In `internal/app/factory.go`:

Add import:
```go
"github.com/typingincolor/bujo/internal/repository/sqlite"
```

Note: `sqlite` is already imported. Add InsightsRepo to Services:

```go
type Services struct {
	DB              *sql.DB
	Bujo            *service.BujoService
	Habit           *service.HabitService
	List            *service.ListService
	Goal            *service.GoalService
	Stats           *service.StatsService
	ChangeDetection *service.ChangeDetectionService
	EditableView    *service.EditableViewService
	InsightsRepo    *sqlite.InsightsRepository
}
```

In `Create` method, add after `services := f.createServices(db)`:

```go
insightsDB, err := OpenInsightsDB(DefaultInsightsDBPath())
if err != nil {
	// Non-fatal: insights are optional
	insightsDB = nil
}
services.InsightsRepo = sqlite.NewInsightsRepository(insightsDB)
if insightsDB != nil {
	origCleanup := cleanup
	cleanup = func() {
		_ = insightsDB.Close()
		origCleanup()
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/app/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add OpenInsightsDB helper and wire InsightsRepo into factory"
```

---

## Phase 4: Wails Adapter — Backend API Methods

### Task 15: Add Insights Methods to Wails App

**Files:**
- Modify: `internal/adapter/wails/app.go` — add insights methods

**Step 1: Add insights methods**

Add to `internal/adapter/wails/app.go`:

```go
func (a *App) IsInsightsAvailable() bool {
	if a.services.InsightsRepo == nil {
		return false
	}
	return a.services.InsightsRepo.IsAvailable()
}

func (a *App) GetInsightsDashboard() (*domain.InsightsDashboard, error) {
	repo := a.services.InsightsRepo
	if repo == nil || !repo.IsAvailable() {
		return &domain.InsightsDashboard{Status: "not_initialized"}, nil
	}

	latest, err := repo.GetLatestSummary(a.ctx)
	if err != nil {
		return nil, err
	}

	initiatives, err := repo.GetActiveInitiatives(a.ctx, 5)
	if err != nil {
		return nil, err
	}

	actions, err := repo.GetPendingActions(a.ctx)
	if err != nil {
		return nil, err
	}

	// Filter to high priority for dashboard
	var highPriority []domain.InsightsAction
	for _, action := range actions {
		if action.Priority == "high" {
			highPriority = append(highPriority, action)
		}
	}

	decisions, err := repo.GetRecentDecisions(a.ctx, 3)
	if err != nil {
		return nil, err
	}

	days, err := repo.GetDaysSinceLastSummary(a.ctx)
	if err != nil {
		return nil, err
	}

	status := "ready"
	if latest == nil {
		status = "empty"
	}

	return &domain.InsightsDashboard{
		LatestSummary:        latest,
		ActiveInitiatives:    initiatives,
		HighPriorityActions:  highPriority,
		RecentDecisions:      decisions,
		DaysSinceLastSummary: days,
		Status:               status,
	}, nil
}

func (a *App) GetInsightsSummaries(limit int) ([]domain.InsightsSummary, error) {
	repo := a.services.InsightsRepo
	if repo == nil || !repo.IsAvailable() {
		return []domain.InsightsSummary{}, nil
	}
	return repo.GetSummaries(a.ctx, limit)
}

func (a *App) GetInsightsSummaryDetail(summaryID int64) ([]domain.InsightsTopic, error) {
	repo := a.services.InsightsRepo
	if repo == nil || !repo.IsAvailable() {
		return []domain.InsightsTopic{}, nil
	}
	return repo.GetTopicsForSummary(a.ctx, summaryID)
}

func (a *App) GetInsightsActions() ([]domain.InsightsAction, error) {
	repo := a.services.InsightsRepo
	if repo == nil || !repo.IsAvailable() {
		return []domain.InsightsAction{}, nil
	}
	return repo.GetPendingActions(a.ctx)
}
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add insights Wails adapter methods"
```

---

## Phase 5: Frontend — React Components

### Task 16: Update Sidebar with Insights Navigation

**Files:**
- Modify: `frontend/src/components/bujo/Sidebar.tsx` — rename stats label, add insights view type

**Step 1: Update ViewType and navItems**

In `frontend/src/components/bujo/Sidebar.tsx`:

1. Add `'insights'` to ViewType union:
```typescript
export type ViewType = 'today' | 'pending' | 'week' | 'questions' | 'habits' | 'lists' | 'goals' | 'search' | 'stats' | 'insights' | 'settings' | 'editable';
```

2. Rename 'stats' label from 'Insights' to 'Statistics' and add new insights item. Add `Lightbulb` to imports from lucide-react:
```typescript
import { ..., Lightbulb } from 'lucide-react';
```

3. Update navItems:
```typescript
const navItems: { view: ViewType; icon: React.ElementType; label: string }[] = [
  { view: 'today', icon: FileEdit, label: 'Journal' },
  { view: 'pending', icon: Clock, label: 'Pending Tasks' },
  { view: 'week', icon: CalendarDays, label: 'Weekly Review' },
  { view: 'questions', icon: HelpCircle, label: 'Open Questions' },
  { view: 'habits', icon: Flame, label: 'Habit Tracker' },
  { view: 'lists', icon: List, label: 'Lists' },
  { view: 'goals', icon: Target, label: 'Monthly Goals' },
  { view: 'search', icon: Search, label: 'Search' },
  { view: 'stats', icon: BarChart3, label: 'Statistics' },
  { view: 'insights', icon: Lightbulb, label: 'Insights' },
];
```

**Note:** In a future iteration, `IsInsightsAvailable()` can be used to conditionally show/hide the insights nav item.

**Step 2: Commit**

```bash
git add -A
git commit -m "feat: add insights to sidebar navigation, rename stats to Statistics"
```

---

### Task 17: Create InsightsView Container with Tabs

**Files:**
- Create: `frontend/src/components/bujo/InsightsView.tsx`

**Step 1: Create the tab container**

```tsx
import { useState } from 'react';
import { cn } from '@/lib/utils';
import { InsightsDashboard } from './InsightsDashboard';
import { InsightsSummaries } from './InsightsSummaries';
import { InsightsActions } from './InsightsActions';

type InsightsTab = 'dashboard' | 'summaries' | 'actions';

const tabs: { id: InsightsTab; label: string }[] = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'summaries', label: 'Summaries' },
  { id: 'actions', label: 'Actions' },
];

export function InsightsView() {
  const [activeTab, setActiveTab] = useState<InsightsTab>('dashboard');

  return (
    <div className="flex flex-col h-full">
      <div className="border-b border-border px-4">
        <div className="flex gap-4">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'py-2 px-1 text-sm border-b-2 transition-colors',
                activeTab === tab.id
                  ? 'border-primary text-primary font-medium'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              )}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 overflow-auto p-4">
        {activeTab === 'dashboard' && <InsightsDashboard />}
        {activeTab === 'summaries' && <InsightsSummaries />}
        {activeTab === 'actions' && <InsightsActions />}
      </div>
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add -A
git commit -m "feat: create InsightsView container with tab navigation"
```

---

### Task 18: Create InsightsDashboard Component

**Files:**
- Create: `frontend/src/components/bujo/InsightsDashboard.tsx`

**Step 1: Create the dashboard component**

```tsx
import { useEffect, useState } from 'react';
import { GetInsightsDashboard } from '../../../wailsjs/go/wails/App';

interface Dashboard {
  LatestSummary: {
    ID: number;
    WeekStart: string;
    WeekEnd: string;
    SummaryText: string;
    CreatedAt: string;
  } | null;
  ActiveInitiatives: {
    ID: number;
    Name: string;
    Status: string;
    Description: string;
    LastUpdated: string;
  }[];
  HighPriorityActions: {
    ID: number;
    ActionText: string;
    Priority: string;
    DueDate: string;
    WeekStart: string;
  }[];
  RecentDecisions: {
    ID: number;
    DecisionText: string;
    Rationale: string;
    DecisionDate: string;
  }[];
  DaysSinceLastSummary: number;
  Status: string;
}

export function InsightsDashboard() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    GetInsightsDashboard()
      .then((data: Dashboard) => setDashboard(data))
      .catch((err: Error) => setError(err.message));
  }, []);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load insights: {error}</div>;
  }

  if (!dashboard) {
    return <div className="text-muted-foreground text-sm">Loading...</div>;
  }

  if (dashboard.Status === 'not_initialized') {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Insights not available.</p>
        <p className="text-sm text-muted-foreground mt-2">
          Generate weekly summaries with Claude to see insights here.
        </p>
      </div>
    );
  }

  if (dashboard.Status === 'empty') {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No insights yet.</p>
        <p className="text-sm text-muted-foreground mt-2">
          Generate your first weekly summary to get started.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Staleness indicator */}
      {dashboard.DaysSinceLastSummary > 0 && (
        <div className="text-xs text-muted-foreground">
          Last summary: {dashboard.LatestSummary?.WeekStart} — {dashboard.LatestSummary?.WeekEnd}
          {' '}({dashboard.DaysSinceLastSummary} days ago)
        </div>
      )}

      {/* Latest Summary */}
      {dashboard.LatestSummary && (
        <div className="border border-border rounded-lg p-4">
          <h3 className="text-sm font-medium mb-2">
            Latest Summary ({dashboard.LatestSummary.WeekStart} — {dashboard.LatestSummary.WeekEnd})
          </h3>
          <p className="text-sm text-muted-foreground whitespace-pre-wrap">
            {dashboard.LatestSummary.SummaryText.length > 500
              ? dashboard.LatestSummary.SummaryText.substring(0, 500) + '...'
              : dashboard.LatestSummary.SummaryText}
          </p>
        </div>
      )}

      {/* Active Initiatives */}
      {dashboard.ActiveInitiatives?.length > 0 && (
        <div className="border border-border rounded-lg p-4">
          <h3 className="text-sm font-medium mb-2">Active Initiatives</h3>
          <ul className="space-y-1">
            {dashboard.ActiveInitiatives.map((i) => (
              <li key={i.ID} className="text-sm flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-green-500" />
                {i.Name}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* High Priority Actions */}
      {dashboard.HighPriorityActions?.length > 0 && (
        <div className="border border-border rounded-lg p-4">
          <h3 className="text-sm font-medium mb-2">High Priority Actions</h3>
          <ul className="space-y-1">
            {dashboard.HighPriorityActions.map((a) => (
              <li key={a.ID} className="text-sm flex justify-between">
                <span>{a.ActionText}</span>
                {a.DueDate && (
                  <span className="text-xs text-muted-foreground">{a.DueDate}</span>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Recent Decisions */}
      {dashboard.RecentDecisions?.length > 0 && (
        <div className="border border-border rounded-lg p-4">
          <h3 className="text-sm font-medium mb-2">Recent Decisions</h3>
          <ul className="space-y-2">
            {dashboard.RecentDecisions.map((d) => (
              <li key={d.ID} className="text-sm">
                <div className="font-medium">{d.DecisionText}</div>
                <div className="text-xs text-muted-foreground">{d.DecisionDate}</div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add -A
git commit -m "feat: create InsightsDashboard component"
```

---

### Task 19: Create InsightsSummaries Component

**Files:**
- Create: `frontend/src/components/bujo/InsightsSummaries.tsx`

**Step 1: Create the summaries list component**

```tsx
import { useEffect, useState } from 'react';
import { GetInsightsSummaries, GetInsightsSummaryDetail } from '../../../wailsjs/go/wails/App';
import { cn } from '@/lib/utils';

interface Summary {
  ID: number;
  WeekStart: string;
  WeekEnd: string;
  SummaryText: string;
  CreatedAt: string;
}

interface Topic {
  ID: number;
  SummaryID: number;
  Topic: string;
  Content: string;
  Importance: string;
}

export function InsightsSummaries() {
  const [summaries, setSummaries] = useState<Summary[]>([]);
  const [expandedID, setExpandedID] = useState<number | null>(null);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    GetInsightsSummaries(10)
      .then((data: Summary[]) => setSummaries(data))
      .catch((err: Error) => setError(err.message));
  }, []);

  const toggleSummary = async (id: number) => {
    if (expandedID === id) {
      setExpandedID(null);
      setTopics([]);
      return;
    }
    setExpandedID(id);
    try {
      const detail = await GetInsightsSummaryDetail(id);
      setTopics(detail);
    } catch (err) {
      setTopics([]);
    }
  };

  if (error) {
    return <div className="text-destructive text-sm">Failed to load summaries: {error}</div>;
  }

  if (summaries.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No weekly summaries yet.</p>
      </div>
    );
  }

  const importanceBadge = (importance: string) => {
    const colors: Record<string, string> = {
      high: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
      medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      low: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
    };
    return (
      <span className={cn('px-1.5 py-0.5 rounded text-xs', colors[importance] || colors.low)}>
        {importance}
      </span>
    );
  };

  return (
    <div className="space-y-3">
      {summaries.map((s) => (
        <div key={s.ID} className="border border-border rounded-lg">
          <button
            onClick={() => toggleSummary(s.ID)}
            className="w-full text-left p-4 hover:bg-muted/50 transition-colors"
          >
            <div className="flex justify-between items-center">
              <h3 className="text-sm font-medium">
                {s.WeekStart} — {s.WeekEnd}
              </h3>
              <span className="text-xs text-muted-foreground">{s.CreatedAt.split(' ')[0]}</span>
            </div>
            {expandedID !== s.ID && (
              <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                {s.SummaryText}
              </p>
            )}
          </button>
          {expandedID === s.ID && (
            <div className="px-4 pb-4 space-y-3">
              <p className="text-sm whitespace-pre-wrap">{s.SummaryText}</p>
              {topics.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-muted-foreground mb-1">Topics</h4>
                  <div className="flex flex-wrap gap-1">
                    {topics.map((t) => (
                      <span key={t.ID} className="inline-flex items-center gap-1">
                        {importanceBadge(t.Importance)}
                        <span className="text-xs">{t.Topic}</span>
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add -A
git commit -m "feat: create InsightsSummaries component with expand/collapse"
```

---

### Task 20: Create InsightsActions Component

**Files:**
- Create: `frontend/src/components/bujo/InsightsActions.tsx`

**Step 1: Create the actions list component**

```tsx
import { useEffect, useState } from 'react';
import { GetInsightsActions } from '../../../wailsjs/go/wails/App';
import { cn } from '@/lib/utils';

interface Action {
  ID: number;
  SummaryID: number;
  ActionText: string;
  Priority: string;
  Status: string;
  DueDate: string;
  CreatedAt: string;
  WeekStart: string;
}

export function InsightsActions() {
  const [actions, setActions] = useState<Action[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    GetInsightsActions()
      .then((data: Action[]) => setActions(data))
      .catch((err: Error) => setError(err.message));
  }, []);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load actions: {error}</div>;
  }

  if (actions.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No pending actions.</p>
      </div>
    );
  }

  const today = new Date().toISOString().split('T')[0];

  const priorityBadge = (priority: string) => {
    const colors: Record<string, string> = {
      high: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
      medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      low: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
    };
    return (
      <span className={cn('px-1.5 py-0.5 rounded text-xs', colors[priority] || colors.low)}>
        {priority}
      </span>
    );
  };

  return (
    <div className="space-y-2">
      {actions.map((a) => {
        const isOverdue = a.DueDate && a.DueDate < today;
        return (
          <div
            key={a.ID}
            className={cn(
              'border border-border rounded-lg p-3 flex items-start justify-between gap-3',
              isOverdue && 'border-red-300 dark:border-red-800'
            )}
          >
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                {priorityBadge(a.Priority)}
                <span className="text-xs text-muted-foreground">
                  from {a.WeekStart}
                </span>
              </div>
              <p className="text-sm">{a.ActionText}</p>
            </div>
            {a.DueDate && (
              <div className={cn(
                'text-xs whitespace-nowrap',
                isOverdue ? 'text-red-600 dark:text-red-400 font-medium' : 'text-muted-foreground'
              )}>
                {isOverdue ? 'Overdue: ' : 'Due: '}{a.DueDate}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add -A
git commit -m "feat: create InsightsActions component with priority badges"
```

---

### Task 21: Integrate InsightsView into App.tsx

**Files:**
- Modify: `frontend/src/App.tsx` — add insights view to routing

**Step 1: Add InsightsView import and rendering**

In `frontend/src/App.tsx`:

1. Add import at top:
```typescript
import { InsightsView } from '@/components/bujo/InsightsView';
```

2. Add `'insights'` to `validViews` array (around line 50)

3. Add to `viewTitles` record (around line 613):
```typescript
insights: 'Insights',
```

4. Also rename the stats entry if still labeled 'Insights':
```typescript
stats: 'Statistics',
```

5. Add rendering block alongside other view conditionals (around line 796):
```tsx
{view === 'insights' && (<InsightsView />)}
```

**Step 2: Generate Wails bindings**

Run: `cd frontend && npm run generate` (or however Wails bindings are regenerated)

If the bindings generator isn't set up, the Wails build will auto-generate them. Verify the app builds:

Run: `wails build` (or `wails dev`)

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: integrate InsightsView into App.tsx"
```

---

## Phase 6: Verification

### Task 22: Run Full Test Suite

**Step 1: Run all Go tests**

Run: `go test ./... -count=1`
Expected: ALL PASS

**Step 2: Run go vet**

Run: `go vet ./...`
Expected: No issues

**Step 3: Build the binary**

Run: `go build ./...`
Expected: BUILD SUCCESS

**Step 4: Update INTEGRATION-SPEC.md status table**

In `docs/INTEGRATION-SPEC.md`, update the Implementation Status table:

| Use Case | Status | Notes |
|----------|--------|-------|
| 1. Weekly Summary Dashboard | Implemented | |
| 5. Action Items Widget | Implemented | |
| 7. Insights Dashboard | Implemented | |
| Database Access Layer | Implemented | InsightsRepository with optional read-only DB |
| Domain Types | Implemented | Summary, Topic, Initiative, Action, Decision structs |
| Deprecate Gemini AI | Implemented | Removed 22+ files, dropped summaries table |

**Step 5: Commit**

```bash
git add -A
git commit -m "docs: update INTEGRATION-SPEC.md implementation status"
```

---

## Summary

**Total tasks:** 22
**Phase 1 (Deprecation):** Tasks 1-8 — Remove 22+ files, clean up references, add drop migration
**Phase 2 (Domain):** Tasks 9-10 — TDD domain types
**Phase 3 (Repository):** Tasks 11-14 — TDD repository with test helper, factory wiring
**Phase 4 (Wails):** Task 15 — Backend API methods
**Phase 5 (Frontend):** Tasks 16-21 — Sidebar, InsightsView, Dashboard, Summaries, Actions, App integration
**Phase 6 (Verification):** Task 22 — Full test suite, docs update
