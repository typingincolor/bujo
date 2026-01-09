# Architecture

This document describes the architecture of bujo, a command-line Bullet Journal application.

## Overview

bujo follows **Hexagonal Architecture** (also known as Ports and Adapters), ensuring that business logic is isolated from external concerns like databases, CLI frameworks, and AI services.

```
┌─────────────────────────────────────────────────────────────┐
│                        Adapters                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   CLI        │  │    TUI       │  │     AI       │       │
│  │  (Cobra)     │  │ (Bubbletea)  │  │  (Gemini)    │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
└─────────┼─────────────────┼─────────────────┼───────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
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
│  │  ListItem    │  │  DayContext  │  │   Summary    │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │  TreeParser  │  │ VersionInfo  │                         │
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
│  │ListItemRepo  │  │ContextRepo   │                         │
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
│   │   ├── summary.go           # AI summary types
│   │   ├── context.go           # Day context (location, mood, weather)
│   │   ├── parser.go            # TreeParser for hierarchical input
│   │   ├── entity_id.go         # UUID-based entity identification
│   │   ├── version.go           # Event sourcing version info
│   │   └── repository.go        # Repository interfaces
│   │
│   ├── service/                 # Application services
│   │   ├── bujo.go              # BujoService (entries, agenda)
│   │   ├── list.go              # ListService
│   │   ├── habit.go             # HabitService
│   │   ├── backup.go            # BackupService
│   │   ├── archive.go           # ArchiveService (version cleanup)
│   │   └── history.go           # HistoryService (version queries)
│   │
│   ├── repository/sqlite/       # SQLite implementations
│   │   ├── entry_repository.go
│   │   ├── list_repository.go
│   │   ├── list_item_repository.go
│   │   ├── habit_repository.go
│   │   ├── context_repository.go
│   │   └── migrations/          # Database migrations
│   │
│   ├── adapter/
│   │   ├── cli/                 # CLI adapter helpers
│   │   └── ai/                  # Gemini AI integration
│   │
│   └── tui/                     # Terminal UI (Bubbletea)
│       ├── model.go             # TUI state model
│       ├── update.go            # Event handling
│       ├── view.go              # Rendering
│       ├── keys.go              # Key bindings
│       ├── styles.go            # Lipgloss styles
│       └── draft.go             # Draft persistence
│
└── docs/                        # Documentation
    ├── ARCHITECTURE.md          # This file
    └── issue-event-sourcing-refactor.md
```

## Layers

### Domain Layer (`internal/domain/`)

The domain layer contains pure business logic with no external dependencies. It defines:

- **Entity types**: `Entry`, `List`, `ListItem`, `Habit`, `HabitLog`, `DayContext`, `Summary`
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
| `ArchiveService` | Old version cleanup |
| `HistoryService` | Version history queries and restoration |

Services are stateless and depend only on repository interfaces.

### Repository Layer (`internal/repository/sqlite/`)

SQLite implementations of domain repository interfaces. Uses:

- **golang-migrate** for schema migrations
- **Event sourcing** pattern for audit trails
- **Soft deletes** via `valid_to` timestamps

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

#### AI Adapter (`internal/adapter/ai/`)

Gemini API integration for AI-powered summaries and reflections.

## Data Model

### Event Sourcing (MANDATORY)

**All repository mutations MUST create new versioned rows. No in-place updates.**

All tables use event sourcing with these columns:

```sql
row_id INTEGER PRIMARY KEY,     -- Unique version identifier
entity_id TEXT NOT NULL,        -- UUID, stable across versions
version INTEGER NOT NULL,       -- Incremental counter
valid_from TEXT NOT NULL,       -- When this version became active
valid_to TEXT,                  -- NULL = current, set when superseded
op_type TEXT NOT NULL           -- INSERT, UPDATE, or DELETE
```

#### Repository Mutation Pattern

Every `Update` method must:
1. Get current version by ID
2. Begin transaction
3. Close current version (`SET valid_to = now`)
4. Get next version number (`MAX(version) + 1`)
5. Insert new row with `op_type = 'UPDATE'`
6. Commit transaction

```go
// CORRECT: Event sourcing pattern
func (r *Repo) Update(ctx, entity) error {
    current := r.GetByID(ctx, entity.ID)
    tx.Exec("UPDATE table SET valid_to = ? WHERE entity_id = ?", now, current.EntityID)
    tx.Exec("INSERT INTO table (..., op_type) VALUES (..., 'UPDATE')")
}

// WRONG: In-place update destroys history
func (r *Repo) Update(ctx, entity) error {
    r.db.Exec("UPDATE table SET col = ? WHERE id = ?", val, id)  // NEVER DO THIS
}
```

#### GetByID Semantics

`GetByID(id)` follows the entity through versions:
1. Look up `entity_id` for the given row `id`
2. Return current version (`valid_to IS NULL AND op_type != 'DELETE'`)

This ensures IDs remain stable references even after updates.

#### Benefits

- Complete audit trails
- Point-in-time queries (`GetAsOf`)
- Safe undo/restore of deleted items
- No data loss from updates

### Tables

| Table | Purpose |
|-------|---------|
| `entries` | Journal entries (tasks, notes, events) |
| `lists` | Named lists for collections |
| `list_items` | Items within lists (separate from entries) |
| `habits` | Habit definitions |
| `habit_logs` | Habit completion logs |
| `day_context` | Daily location, mood, weather |
| `summaries` | Cached AI summaries |

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

### Event Sourcing for Data

**Why**: Provides audit trails, enables undo, and supports temporal queries.

**Trade-off**: Database grows over time (mitigated by archive command).

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
| `GEMINI_API_KEY` | AI features API key | (none) |

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/mattn/go-sqlite3` | SQLite driver |
| `github.com/golang-migrate/migrate` | Schema migrations |
| `github.com/google/uuid` | Entity ID generation |
| `github.com/araddon/dateparse` | Natural language dates |
