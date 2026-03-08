# Architecture

This document describes the architecture of bujo, a command-line Bullet Journal application.

## Overview

bujo follows **Hexagonal Architecture** (also known as Ports and Adapters), ensuring that business logic is isolated from external concerns like databases, CLI frameworks, and AI services.

```
┌─────────────────────────────────────────────────────────────┐
│                        Adapters                             │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐            │
│  │  CLI   │  │  TUI   │  │ Wails  │  │Insights│            │
│  │(Cobra) │  │(Bubble)│  │(Desktop│  │(SQLite)│            │
│  └───┬────┘  └───┬────┘  └───┬────┘  └───┬────┘            │
└──────┼───────────┼───────────┼───────────┼─────────────────┘
       │           │           │           │
       ▼           ▼           ▼           ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ BujoService  │  │ ListService  │  │ HabitService │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │BackupService │  │ArchiveService│  │HistoryService│       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │    Entry     │  │    List      │  │    Habit     │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  ListItem    │  │  DayContext  │  │    Goal      │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │  TreeParser  │  │   EntityID   │                         │
│  └──────────────┘  └──────────────┘                         │
└─────────────────────────────────────────────────────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │EntryRepository│ │ListRepository│  │HabitRepository│      │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │ListItemRepo  │  │DayContextRepo│                         │
│  └──────────────┘  └──────────────┘                         │
└─────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                       SQLite                                 │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
bujo/
├── cmd/bujo/                    # Application entry point
│   ├── main.go                  # Main function
│   └── cmd/                     # CLI commands (Cobra)
│       ├── root.go              # Root command setup
│       ├── add.go               # bujo add
│       ├── ls.go                # bujo ls
│       ├── tui.go               # bujo tui
│       ├── habit.go             # bujo habit
│       ├── list.go              # bujo list
│       ├── backup.go            # bujo backup
│       └── ...                  # Other commands
│
├── internal/
│   ├── domain/                  # Core business logic
│   │   ├── entry.go             # Entry types (task, note, event)
│   │   ├── habit.go             # Habit and HabitLog
│   │   ├── list.go              # List management
│   │   ├── list_item.go         # List items (separate from entries)
│   │   ├── context.go           # Day context (location, mood, weather)
│   │   ├── goal.go              # Monthly goals
│   │   ├── insights.go          # Insights read models
│   │   ├── parser.go            # TreeParser for hierarchical input
│   │   ├── entity_id.go         # UUID-based entity identification
│   │   ├── version.go           # Version metadata (for versioned entities)
│   │   └── repository.go        # Repository interfaces
│   │
│   ├── service/                 # Application services
│   │   ├── bujo.go              # BujoService (entries, agenda)
│   │   ├── list.go              # ListService
│   │   ├── habit.go             # HabitService
│   │   ├── goal.go              # GoalService
│   │   ├── stats.go             # StatsService
│   │   ├── backup.go            # BackupService
│   │   ├── archive.go           # ArchiveService (list item version cleanup)
│   │   └── history.go           # HistoryService (list item history)
│   │
│   ├── repository/sqlite/       # SQLite implementations
│   │   ├── entry_repository.go
│   │   ├── list_repository.go
│   │   ├── list_item_repository.go
│   │   ├── habit_repository.go
│   │   ├── day_context_repository.go
│   │   ├── goal_repository.go
│   │   ├── insights_repository.go
│   │   └── migrations/          # Database migrations
│   │
│   ├── adapter/
│   │   ├── cli/                 # CLI adapter helpers
│   │   ├── http/                # HTTP API + bookmarklet endpoint
│   │   ├── remarkable/          # reMarkable import and OCR pipeline
│   │   └── wails/               # Desktop app adapter
│   │       └── app.go           # Wails bindings to services
│   │
│   └── tui/                     # Terminal UI (Bubbletea)
│       ├── model.go             # TUI state model
│       ├── update.go            # Event handling
│       ├── view.go              # Rendering
│       ├── keymap.go            # Key bindings
│       ├── styles.go            # Lipgloss styles
│       └── draft.go             # Draft persistence
│
├── frontend/                    # Desktop app React frontend
│   ├── src/
│   │   ├── App.tsx              # Main application component
│   │   ├── components/bujo/     # UI components
│   │   ├── lib/                 # Utilities and transforms
│   │   └── wailsjs/             # Generated Wails bindings
│   └── package.json
│
└── docs/                        # Documentation
    ├── CLI.md                   # CLI command reference
    ├── TUI.md                   # TUI keyboard shortcuts
    ├── FRONTEND.md              # Desktop app guide
    └── ARCHITECTURE.md          # This file
```

## Layers

### Domain Layer (`internal/domain/`)

The domain layer contains pure business logic with no external dependencies. It defines:

- **Entity types**: `Entry`, `List`, `ListItem`, `Habit`, `HabitLog`, `DayContext`, `Goal`
- **Value objects**: `EntryType`, `EntityID`, `VersionInfo`, `OpType`
- **Business rules**: Validation, type conversions, date calculations
- **Repository interfaces**: Contracts for data access

**Key principle**: Domain code has zero imports from adapters or infrastructure.

### Service Layer (`internal/service/`)

Services orchestrate business operations by combining domain logic with repositories:

| Service | Responsibility |
|---------|---------------|
| `BujoService` | Entry CRUD, agenda queries, migration |
| `ListService` | List management, item operations |
| `HabitService` | Habit tracking, streaks, goals |
| `BackupService` | Database backup creation and verification |
| `ArchiveService` | List-item history cleanup |
| `HistoryService` | List-item version queries and restoration |

Services are stateless and depend only on repository interfaces.

### Repository Layer (`internal/repository/sqlite/`)

SQLite implementations of domain repository interfaces. Uses:

- **golang-migrate** for schema migrations
- A **hybrid persistence model** (mutable `entries`, versioned append for list/habit/context/goal entities)
- Version tombstones (`op_type='DELETE'`) and `valid_to` for versioned entities

### Adapter Layer

#### CLI Adapter (`cmd/bujo/cmd/`)

Cobra commands that parse user input and call services. Each command:
1. Parses flags and arguments
2. Calls appropriate service method
3. Formats output for the terminal

#### TUI Adapter (`internal/tui/`)

Bubbletea-based terminal UI with:
- **Model**: Application state (`Model` struct)
- **Update**: Event handling (keyboard, window resize)
- **View**: Rendering with Lipgloss styles

Features:
- Day/week view modes
- Inline editing and adding
- Capture mode for multi-entry input
- Incremental search (Ctrl+S/R)
- Draft persistence

#### Wails Adapter (`internal/adapter/wails/`)

Desktop application adapter using [Wails](https://wails.io/). Exposes service methods to the React frontend via Go bindings.

**Architecture:**
- `app.go`: Main App struct with bound methods
- Methods receive frontend calls and delegate to services
- Returns domain types (auto-serialized to JSON)

**Frontend** (`frontend/`):
- React with TypeScript
- Tailwind CSS for styling
- Generated bindings in `wailsjs/` directory
- Components mirror TUI functionality

#### HTTP/Insights/Import Adapters

- `internal/adapter/http/`: local HTTP server and integration endpoints (for example Gmail bookmarklet install/API)
- `internal/adapter/remarkable/`: reMarkable sync/import, rendering, OCR normalization
- Insights are stored/read through `internal/repository/sqlite/insights_repository.go` and surfaced in TUI/Wails

## Data Model

### Hybrid Persistence Model (Current)

bujo does **not** use one persistence strategy for every table.

#### Mutable table

- `entries` currently uses in-place updates/deletes in `entry_repository.go`.
- Entry history rows were purged by migration `000029_purge_event_sourcing_history.up.sql`.

#### Versioned append tables

These tables still use versioning columns (`entity_id`, `version`, `valid_from`, `valid_to`, `op_type`) with current-row filters:
- `lists`
- `list_items`
- `habits`
- `habit_logs`
- `day_context`
- `goals`

#### Archive/History Scope

- `ArchiveService` and `HistoryService` operate on **list items**.
- `bujo history ...` is list-item history, not global entry history.

### Tables

| Table | Purpose |
|-------|---------|
| `entries` | Journal entries (tasks, notes, events) |
| `lists` | Named lists for collections |
| `list_items` | Items within lists (separate from entries) |
| `habits` | Habit definitions |
| `habit_logs` | Habit completion logs |
| `day_context` | Daily location, mood, weather |
| `insights` (separate DB) | Read-only AI/analysis data loaded from `~/.bujo/claude-insights.db` when available |

### Entry Types

| Type | Symbol | Description |
|------|--------|-------------|
| Task | `.` / `•` | Todo item |
| Note | `-` / `–` | Information |
| Event | `o` / `○` | Scheduled occurrence |
| Done | `x` / `✓` | Completed task |
| Migrated | `>` / `→` | Moved to another day |

## Key Design Decisions

### Hexagonal Architecture

**Why**: Isolates business logic from frameworks, making it testable and adaptable.

**Trade-off**: More boilerplate (interfaces, dependency injection) but better long-term maintainability.

### Hybrid Persistence for Data

**Why**: Keeps high-churn journal entries simple and fast, while preserving version history where restore/audit is needed most (lists, habits, goals, day context).

**Trade-off**: Semantics differ by entity type, so docs/tests must be explicit about which operations are versioned.

### Separate List Items Table

**Why**: List items have different semantics than journal entries (no dates, no hierarchy). Separation prevents bugs like Issue #54 where operations could affect wrong entity types.

### TUI Draft Persistence

**Why**: Users shouldn't lose work if the app crashes during multi-entry capture.

**Implementation**: Auto-saves to `~/.bujo/capture_draft.txt` on each keystroke, prompts to restore on re-entry.

### TreeParser for Input

**Why**: Enables hierarchical entry creation from plain text using indentation.

**Example**:
```
. Parent task
  - Child note
  . Child task
```

## Testing Strategy

- **Domain layer**: 100% unit test coverage required
- **Service layer**: Integration tests with real SQLite (in-memory)
- **Repository layer**: Integration tests with migrations
- **TUI layer**: Unit tests for state transitions
- **CLI layer**: Integration tests for command behavior

All code follows TDD: tests written before implementation.

## Configuration

| Environment Variable | Purpose | Default |
|---------------------|---------|---------|
| `DB_PATH` | Database file location | `~/.bujo/bujo.db` |

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/wailsapp/wails/v2` | Desktop app framework |
| `github.com/mattn/go-sqlite3` | SQLite driver |
| `github.com/golang-migrate/migrate` | Schema migrations |
| `github.com/google/uuid` | Entity ID generation |
| `github.com/tj/go-naturaldate` | Natural language dates |
