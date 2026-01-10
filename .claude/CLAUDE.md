# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **About this file (v4.0.0):** Project-specific guidelines for bujo, a Go-based CLI Bullet Journal.
>
> **Architecture:**
> - **CLAUDE.md** (this file): Project context, commands, and architecture
> - **Skills**: Detailed patterns loaded on-demand (tdd, testing, functional, refactoring, planning)
> - **Agents**: Specialized subprocesses for verification and analysis
> - **spec.md**: Complete technical specification

## Project Overview

bujo is a high-performance, Go-based command-line Bullet Journal for macOS. It captures tasks, notes, events, habits, and locations with AI-powered reflections using Google's Gemini API.

**Tech Stack:** Go 1.23, SQLite, Cobra CLI, Gemini API

## Build Commands

```bash
go build -o bujo ./cmd/bujo     # Build binary
go test ./...                    # Run all tests
go test ./internal/domain/...   # Run domain tests only
go test -v -run TestName ./...  # Run specific test
go test -cover ./...            # Run with coverage
go vet ./...                    # Static analysis
```

## Architecture

Hexagonal Architecture with clear separation:

```
cmd/bujo/           CLI entry point (Cobra adapter)
internal/
  domain/           Core business logic (100% TDD coverage required)
    entry.go        Entry types: Task (.), Note (-), Event (o), Done (x), Migrated (>)
    habit.go        Habit tracking with multi-log support
    summary.go      AI summary types
    parser.go       TreeParser for hierarchical input
  service/          Stateless services (BujoService, HabitService)
  repository/       SQLite repository implementations
  adapter/
    cli/            Cobra command handlers
    ai/             Gemini integration
```

**Key Principle:** Business logic isolated in `internal/domain`. CLI and future web server are adapters to shared logic.

## Data Model

SQLite with these tables:
- `entries`: Hierarchical items with `parent_id` for tree structure
- `habits`: Recurring habits with `goal_per_day`
- `habit_logs`: Multiple logs per habit per day
- `day_context`: Daily location, mood, weather
- `summaries`: Cached AI summaries (daily/weekly/quarterly/annual)

## Event Sourcing (MANDATORY)

**ALL entities use event sourcing.** Every mutation creates a new versioned row. No exceptions.

### Event Sourcing Columns (on all entity tables)

| Column | Purpose |
|--------|---------|
| `entity_id` | Stable UUID identifying the logical entity across versions |
| `version` | Incrementing version number (1, 2, 3...) |
| `valid_from` | Timestamp when this version became active |
| `valid_to` | Timestamp when this version was superseded (NULL = current) |
| `op_type` | Operation type: `INSERT`, `UPDATE`, or `DELETE` |

### Repository Pattern

Every repository mutation MUST follow this pattern:

```go
func (r *Repo) Update(ctx context.Context, entity Entity) error {
    // 1. Get current version
    current, err := r.GetByID(ctx, entity.ID)

    // 2. Close current version
    tx.Exec("UPDATE table SET valid_to = ? WHERE entity_id = ? AND valid_to IS NULL", now, current.EntityID)

    // 3. Get next version number
    tx.QueryRow("SELECT MAX(version) FROM table WHERE entity_id = ?", current.EntityID)

    // 4. Insert new version
    tx.Exec("INSERT INTO table (..., entity_id, version, valid_from, op_type) VALUES (..., ?, ?, ?, 'UPDATE')",
        current.EntityID, maxVersion+1, now)
}
```

### GetByID Semantics

`GetByID(id)` returns the **current version** of the entity, even if `id` refers to an old row:
1. Look up `entity_id` for the given `id`
2. Return current version (`valid_to IS NULL AND op_type != 'DELETE'`) for that `entity_id`

### Why This Matters

- Full audit trail of all changes
- Point-in-time queries (`GetAsOf`)
- Undo/restore deleted items
- No data loss from updates

**If you write a repository method that does `UPDATE ... SET` without creating a new version, STOP and rewrite it.**

## Core Philosophy

**TEST-DRIVEN DEVELOPMENT IS NON-NEGOTIABLE.** Every line of production code must be written in response to a failing test. No exceptions.

## TDD Checkpoints (MANDATORY)

Before writing ANY production code, you MUST:

1. **Say "RED:" explicitly** - Write "RED: Writing failing test for [specific behavior]"
2. **Write the test first** - Create the test file/function
3. **Run and show failure** - Execute `go test` and display the failure
4. **Only then say "GREEN:"** - Write "GREEN: Implementing [specific behavior]"
5. **Write minimum code** - Just enough to pass the test
6. **Run and show success** - Execute `go test` and display passing
7. **Say "REFACTOR:"** - Assess if refactoring adds value, do it if so

**If you catch yourself writing production code without a failing test, STOP IMMEDIATELY:**
- Delete the production code you wrote
- Go back to step 1
- This applies even for "simple" fixes

## "Hard to Test" is Not an Excuse

If something feels hard to test (CLI commands, integration points, etc.):

1. **Extract the logic** into a pure, testable function
2. **Write the test FIRST** for that extracted function
3. **Then integrate** the tested function into the harder-to-test code

Never skip tests because:
- "It's just CLI glue code"
- "It's an integration point"
- "It's a simple one-liner"
- "I'll add tests later"

## Development Workflow

**RED-GREEN-REFACTOR:**
1. RED: Write failing test first (NO production code without failing test)
2. GREEN: Write MINIMUM code to pass test
3. REFACTOR: Assess improvement opportunities (only if adds value)
4. **Wait for commit approval** before every commit

## Code Style

**Functional patterns adapted for Go:**
- Prefer immutability where practical
- Pure functions wherever possible
- No nested if/else - use early returns
- No comments - code should be self-documenting
- Table-driven tests for comprehensive coverage

**Testing:**
- Test behavior, not implementation
- Use test helpers/factories for test data
- Tests document expected business behavior
- 100% coverage in `internal/domain` before other layers

## 12-Factor Patterns

- **Config:** Environment variables (`GEMINI_API_KEY`, `DB_PATH`)
- **Logs:** Diagnostic messages to stderr, data output to stdout
- **Dependencies:** Strict via go.mod

## Implementation Order

1. Domain Layer first: Types + TreeParser with 100% TDD
2. Service Layer: BujoService, HabitService (UI-agnostic)
3. Infrastructure: SQLite repositories with golang-migrate
4. Adapter (CLI): Cobra commands
5. Adapter (AI): Gemini with rolling summary logic

For detailed specifications, see `spec.md`.
