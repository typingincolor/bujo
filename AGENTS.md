# AGENTS.md

## Purpose
This file is the working guide for coding agents in this repository. It reflects the current codebase and resolves documentation drift in `docs/ARCHITECTURE.md` and `docs/DATA.md`.

## Read First
1. `README.md` for product scope and interfaces.
2. `docs/ARCHITECTURE.md` for high-level layering (but see persistence correction below).
3. `docs/CLI.md`, `docs/TUI.md`, `docs/FRONTEND.md` for surface behavior.
4. `docs/WORKFLOWS.md` for expected user flows.
5. `docs/REMARKABLE.md` for reMarkable integration architecture, flow, and caveats.
6. `BUILD.md` and `Makefile` for build/test commands.

## Current Architecture (Code Truth)
- Language/runtime: Go 1.24 (`go.mod`), SQLite storage, Cobra CLI, Bubbletea TUI, Wails + React desktop UI.
- Entry point: `cmd/bujo/main.go`.
- CLI command handlers: `cmd/bujo/cmd/*.go`.
- Core domain: `internal/domain`.
- Application services: `internal/service`.
- SQLite repositories: `internal/repository/sqlite`.
- Wails adapter: `internal/adapter/wails`.
- HTTP adapter/integration endpoint: `internal/adapter/http`.
- Frontend UI: `frontend/src`.

## Critical Correction: Persistence Model
`docs/ARCHITECTURE.md` currently states that all repository mutations are append-only event sourcing. That is no longer true.

Use this model instead:

### 1) Mutable row model (in-place update/delete)
- `entries` is currently mutable in practice.
- `internal/repository/sqlite/entry_repository.go` uses direct `UPDATE entries ... WHERE id = ?` and `DELETE FROM entries ...`.
- Migration `internal/repository/sqlite/migrations/000029_purge_event_sourcing_history.up.sql` explicitly purged entry history rows.

Implication: do not design entry features assuming full temporal history/restore.

### 2) Versioned append model (event-sourced style)
These repositories still use `entity_id`, `version`, `valid_from`, `valid_to`, `op_type` with tombstones/current-row filters:
- `lists` (`list_repository.go`)
- `list_items` (`list_item_repository.go`)
- `habits` (`habit_repository.go`)
- `habit_logs` (`habit_log_repository.go`)
- `day_context` (`day_context_repository.go`)
- `goals` (`goal_repository.go`)

Implication: preserve close-old-row + insert-new-row behavior when changing these entities.

### 3) Scope of archive/history commands
- `ArchiveService` and `HistoryService` are currently list-item scoped.
- `internal/service/archive.go` and `internal/service/history.go` depend only on `ListItemRepository`.
- `bujo history ...` applies to list items, not generic entry history.

## Change Rules
- Prefer service-layer changes over calling repositories directly from adapters.
- Keep CLI/TUI/Wails behavior aligned by updating shared service logic first.
- If a change affects schema semantics, add a migration under `internal/repository/sqlite/migrations` and repository tests.
- When touching mutable `entries`, reason about parent-child integrity (`parent_id`) and date/sort behavior.
- When touching versioned entities, maintain version progression and `valid_to` closure semantics.
- Every code change must follow TDD: write or update a failing test first, implement the minimal fix, then refactor with tests passing.

## Testing Expectations
TDD is mandatory for all non-trivial code changes. Use a Red-Green-Refactor loop:
1. Add or adjust a test that fails for the desired behavior.
2. Implement the minimal code change to make it pass.
3. Refactor while keeping the full relevant suite green.

Run targeted tests for changed areas first, then broader suites.

Common commands:
- `go test ./...`
- `make test` (Go + frontend tests)
- `cd frontend && npm run test`
- `cd frontend && npm run lint`

Repository and service behavior is heavily test-covered in:
- `internal/repository/sqlite/*_test.go`
- `internal/service/*_test.go`
- `internal/tui/*_test.go`
- `frontend/src/**/*.test.tsx`

## Docs Hygiene
When behavior changes, update docs in `docs/` in the same change. Specifically avoid reintroducing the incorrect claim that all entities are fully event-sourced.

## Practical Navigation
- Entry operations: `internal/service/bujo.go` + `internal/repository/sqlite/entry_repository.go`
- List/list-item operations: `internal/service/list.go`
- Habit operations: `internal/service/habit.go`
- Goals: `internal/service/goal.go`
- Editable document flow: `internal/service/editable_view.go`
- Wails bindings: `internal/adapter/wails/app.go` and generated bindings in `frontend/src/wailsjs`

## If Unsure
Default to code truth over docs, especially for data mutation semantics. Confirm assumptions by checking repository implementations and migrations.
