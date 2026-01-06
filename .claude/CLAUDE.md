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

**Tech Stack:** Go 1.25.5, SQLite, Cobra CLI, Gemini API

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

## Core Philosophy

**TEST-DRIVEN DEVELOPMENT IS NON-NEGOTIABLE.** Every line of production code must be written in response to a failing test. No exceptions.

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
