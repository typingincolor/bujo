# Editable Journal View Design

**Date:** 2026-01-28
**Status:** Draft
**Feature:** Transform journal view into an editable document where entries can be modified inline

## Overview

Replace the current read-only journal display + action buttons with a live editable document. Users can modify entry type, priority, content, hierarchy, and migration status by directly editing text. Changes are validated and synced on explicit save.

## Design Decisions

### Document Format

Entries rendered as ASCII text with symbols:

| Symbol | Entry Type | Example |
|--------|------------|---------|
| `.` | Task | `. Buy groceries` |
| `-` | Note | `- Meeting went well` |
| `o` | Event | `o Team standup at 10am` |
| `x` | Done | `x Finished report` |
| `~` | Cancelled | `~ No longer needed` |
| `?` | Question | `? How does auth work` |
| `>` | Migrated | (display only, see migration syntax) |

**Priority:** Prefix with `!`, `!!`, or `!!!` before symbol:
```
!!! . Urgent task
!! . High priority
! . Low priority
. Normal task
```

**Hierarchy:** 2-space indentation per level:
```
. Parent task
  . Child task
    . Grandchild
  - Sibling note
```

**Migration syntax:** Change symbol to `>` with target date:
```
>[2026-01-29] Call dentist
>[tomorrow] Review PR
>[next monday] Submit report
```

### Example Document

```
── Monday, Jan 27 ──────────────────
. Buy groceries
  . Milk
  . Eggs
- Meeting went well
!! . Urgent: fix prod bug
x Deployed hotfix
? How does the auth flow work
  - See docs/auth.md
>[tomorrow] . Schedule follow-up

── Pending Deletion ────────────────
~ Old task I removed
```

### Sync Behavior

- **Draft mode:** All changes stay in local buffer until explicit save
- **Trigger:** Save on Ctrl+S (Frontend) / Ctrl+S (TUI)
- **Validation:** Full document parsed before sync; errors block save
- **Scope:** One day at a time (not multi-day editing)

### ID Handling

- **Hidden:** Entry IDs not shown in editable view
- **Tracked:** Backend provides line → EntityID mapping on load
- **New entries:** Lines without mapping get new EntityID on save
- **Moves:** Line position changes tracked to update hierarchy

### Deletion Behavior

1. User deletes a line
2. Line moves to "Pending Deletion" section at bottom
3. User can restore by moving line back into day section
4. On save: pending deletions become soft deletes (event sourcing)
5. Post-save: can still `restore <id>` via CLI

### Error Handling

- Invalid lines highlighted in red
- Error message shown inline or in status bar
- Save blocked until all errors resolved
- Undo: Editor-level undo (standard text editing)

### Concurrent Edits

- Unlikely scenario (Frontend + TUI simultaneously)
- No special handling initially
- Event sourcing provides conflict resolution via versioning

---

## Architecture

### Layer Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    UI Layer                                 │
│  ┌─────────────────────────┐  ┌─────────────────────────┐   │
│  │   Frontend (React)      │  │   TUI (Bubble Tea)      │   │
│  │   EditableJournalView   │  │   editableViewMode      │   │
│  └────────────┬────────────┘  └────────────┬────────────┘   │
│               │                            │                │
│               └──────────┬─────────────────┘                │
└──────────────────────────┼──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   Adapter Layer                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │   Wails App                                         │    │
│  │   - GetEditableDocument(date) → string + mapping    │    │
│  │   - ValidateDocument(doc) → errors[]                │    │
│  │   - ApplyDocument(doc, date, deletes) → error       │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   Service Layer                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │   EditableViewService (new)                         │    │
│  │   - GetEditableDocument()                           │    │
│  │   - ValidateDocument()                              │    │
│  │   - ApplyChanges()                                  │    │
│  └─────────────────────────┬───────────────────────────┘    │
│                            │                                │
│  ┌─────────────────────────▼───────────────────────────┐    │
│  │   BujoService (existing - unchanged)                │    │
│  │   - LogEntries(), EditEntry(), DeleteEntry()        │    │
│  │   - MigrateEntry(), RestoreEntry()                  │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   Domain Layer                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │   EditableDocumentParser (new)                      │    │
│  │   - Parse(text, existing) → ParsedDocument          │    │
│  │   - Serialize(entries) → string                     │    │
│  └─────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────┐    │
│  │   DocumentDiffer (new)                              │    │
│  │   - Diff(original, parsed) → Changeset              │    │
│  └─────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────┐    │
│  │   MigrationSyntaxParser (new)                       │    │
│  │   - Parse(">[date] content") → (content, date, err) │    │
│  │   - Uses existing dateutil.ParseFuture()            │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

---

## Domain Layer Detail

### File: `internal/domain/editable_document.go`

```go
// ParsedLine represents a single line from the editable document
type ParsedLine struct {
    LineNumber    int
    Raw           string
    Depth         int            // indentation level (0, 1, 2...)
    Symbol        rune           // '.', '-', 'o', 'x', '~', '?'
    Priority      Priority
    Content       string
    EntityID      *EntityID      // nil for new entries
    MigrateTarget *time.Time     // parsed from >[date] syntax
    IsValid       bool
    ErrorMessage  string
}

// EditableDocument represents the parsed state of an editable view
type EditableDocument struct {
    Date            time.Time
    Lines           []ParsedLine
    PendingDeletes  []EntityID
    OriginalMapping map[EntityID]int  // entity → original line number
}

// DiffOperation represents a single change to apply
type DiffOperation struct {
    Type        DiffOpType
    EntityID    *EntityID      // for existing entries
    Entry       Entry          // new/updated entry data
    MigrateDate *time.Time     // for migrate operations
    NewParentID *EntityID      // for reparent operations
    LineNumber  int            // for error reporting
}

type DiffOpType int

const (
    DiffOpInsert DiffOpType = iota
    DiffOpUpdate
    DiffOpDelete
    DiffOpMigrate
    DiffOpReparent
)

// Changeset is the result of diffing original vs edited document
type Changeset struct {
    Operations []DiffOperation
    Errors     []ParseError
}

type ParseError struct {
    LineNumber int
    Message    string
}
```

### File: `internal/domain/editable_parser.go`

```go
type EditableDocumentParser struct {
    dateParser func(string) (time.Time, error)
}

func NewEditableDocumentParser(dateParser func(string) (time.Time, error)) *EditableDocumentParser

// Parse converts text document + existing entries into ParsedDocument
func (p *EditableDocumentParser) Parse(input string, existing []Entry) (*EditableDocument, error)

// Serialize converts entries into editable text format
func (p *EditableDocumentParser) Serialize(entries []Entry) string

// ParseLine parses a single line
func (p *EditableDocumentParser) ParseLine(line string, lineNum int) ParsedLine
```

**Parsing rules:**
1. Lines starting with `──` are headers (skip)
2. Count leading spaces / 2 = depth
3. First non-space char is priority (`!`) or symbol
4. Symbol determines entry type
5. If symbol is `>`, parse `[date]` portion
6. Remainder is content

### File: `internal/domain/document_diff.go`

```go
// ComputeDiff compares original entries with parsed document
func ComputeDiff(original []Entry, parsed *EditableDocument) *Changeset
```

**Diff algorithm:**
1. Build map of EntityID → original entry
2. For each parsed line:
   - If EntityID exists: compare for changes (type, priority, content, parent)
   - If EntityID nil: mark as insert
   - If migration syntax: mark as migrate
3. For EntityIDs not in parsed: mark as delete (unless in PendingDeletes)
4. Validate hierarchy (no orphan children)

### File: `internal/domain/migration_syntax.go`

```go
// ParseMigrationSyntax extracts date from >[date] prefix
// Returns: content without prefix, target date, error
func ParseMigrationSyntax(line string, dateParser func(string) (time.Time, error)) (string, *time.Time, error)
```

**Supported formats:**
- `>[2026-01-29]` - ISO date
- `>[tomorrow]` - natural language (via dateutil.ParseFuture)
- `>[next monday]` - natural language

---

## Service Layer Detail

### File: `internal/service/editable_view.go`

```go
type EditableViewService struct {
    bujoService *BujoService
    parser      *domain.EditableDocumentParser
}

func NewEditableViewService(bujoService *BujoService) *EditableViewService

// GetEditableDocument returns serialized entries for a date
func (s *EditableViewService) GetEditableDocument(ctx context.Context, date time.Time) (*EditableDocumentResponse, error)

// ValidateDocument parses and validates without applying
func (s *EditableViewService) ValidateDocument(ctx context.Context, document string, date time.Time) (*ValidationResult, error)

// ApplyChanges computes diff and applies atomically
func (s *EditableViewService) ApplyChanges(ctx context.Context, document string, date time.Time, pendingDeletes []string) error

type EditableDocumentResponse struct {
    Document    string
    LineMapping []LineMapping
}

type LineMapping struct {
    LineNumber int
    EntityID   string
}

type ValidationResult struct {
    IsValid  bool
    Errors   []LineError
    Warnings []LineWarning
}

type LineError struct {
    LineNumber int
    Message    string
}

type LineWarning struct {
    LineNumber int
    Message    string
}
```

**ApplyChanges flow:**
1. Get existing entries for date
2. Parse document with parser
3. Compute diff
4. If diff has errors, return validation error
5. Begin transaction
6. Apply each operation via BujoService methods
7. Commit transaction

---

## Wails Adapter Detail

### File: `internal/adapter/wails/app.go` (additions)

```go
// GetEditableDocument returns the editable view for a date
func (a *App) GetEditableDocument(dateStr string) (*EditableDocumentResponse, error) {
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return nil, err
    }
    return a.editableViewService.GetEditableDocument(a.ctx, date)
}

// ValidateEditableDocument checks document for errors
func (a *App) ValidateEditableDocument(document string, dateStr string) (*ValidationResult, error) {
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return nil, err
    }
    return a.editableViewService.ValidateDocument(a.ctx, document, date)
}

// ApplyEditableDocument saves changes from editable view
func (a *App) ApplyEditableDocument(document string, dateStr string, pendingDeletes []string) error {
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return err
    }
    return a.editableViewService.ApplyChanges(a.ctx, document, date, pendingDeletes)
}
```

---

## Implementation Phases

### Phase 1: Domain Layer (TDD)

| Task | File | Description |
|------|------|-------------|
| 1.1 | `editable_document.go` | Define types: ParsedLine, EditableDocument, DiffOperation, Changeset |
| 1.2 | `editable_parser_test.go` | Tests for parsing single lines (all symbol types, priorities) |
| 1.3 | `editable_parser.go` | Implement ParseLine() |
| 1.4 | `editable_parser_test.go` | Tests for full document parsing with hierarchy |
| 1.5 | `editable_parser.go` | Implement Parse() |
| 1.6 | `editable_parser_test.go` | Tests for serialization |
| 1.7 | `editable_parser.go` | Implement Serialize() |
| 1.8 | `migration_syntax_test.go` | Tests for migration syntax parsing |
| 1.9 | `migration_syntax.go` | Implement ParseMigrationSyntax() |
| 1.10 | `document_diff_test.go` | Tests for diff computation |
| 1.11 | `document_diff.go` | Implement ComputeDiff() |

### Phase 2: Service Layer

| Task | File | Description |
|------|------|-------------|
| 2.1 | `editable_view_test.go` | Tests for GetEditableDocument |
| 2.2 | `editable_view.go` | Implement GetEditableDocument |
| 2.3 | `editable_view_test.go` | Tests for ValidateDocument |
| 2.4 | `editable_view.go` | Implement ValidateDocument |
| 2.5 | `editable_view_test.go` | Tests for ApplyChanges |
| 2.6 | `editable_view.go` | Implement ApplyChanges |

### Phase 3: Wails Adapter

| Task | File | Description |
|------|------|-------------|
| 3.1 | `app.go` | Add EditableViewService dependency |
| 3.2 | `app.go` | Implement GetEditableDocument API |
| 3.3 | `app.go` | Implement ValidateEditableDocument API |
| 3.4 | `app.go` | Implement ApplyEditableDocument API |
| 3.5 | TypeScript | Generate TypeScript types |

### Phase 4: Frontend Implementation

| Task | File | Description |
|------|------|-------------|
| 4.1 | `useEditableDocument.ts` | State management hook |
| 4.2 | `EditableJournalView.tsx` | Main component with text editor |
| 4.3 | Editor integration | CodeMirror/Monaco with syntax highlighting |
| 4.4 | Validation display | Red highlighting for error lines |
| 4.5 | Pending deletion UI | Section at bottom with restore buttons |
| 4.6 | Keyboard shortcuts | Cmd+S save, Escape exit |
| 4.7 | Edit mode toggle | Button to enter/exit edit mode |

### Phase 5: TUI Implementation

| Task | File | Description |
|------|------|-------------|
| 5.1 | `editable_mode.go` | Define editableViewState struct |
| 5.2 | `editable_mode.go` | Render editable text buffer |
| 5.3 | `editable_mode.go` | Text editing (insert, delete, cursor) |
| 5.4 | `editable_mode.go` | Pending deletion section |
| 5.5 | Key bindings | `e` toggle, Ctrl+S save, Escape exit |
| 5.6 | Validation display | Red highlighting for errors |
| 5.7 | Integration | Hook into main TUI view switching |

---

## Testing Strategy

### Domain Layer (100% coverage required)

**editable_parser_test.go:**
- Parse each symbol type (., -, o, x, ~, ?)
- Parse priority levels (!, !!, !!!)
- Parse indentation depths (0, 1, 2, 3+)
- Parse migration syntax variations
- Handle malformed lines gracefully
- Serialize round-trip: parse(serialize(entries)) == entries

**document_diff_test.go:**
- Detect insert (new line without EntityID)
- Detect update (content/type/priority changed)
- Detect delete (EntityID missing from document)
- Detect migrate (> with date)
- Detect reparent (indentation changed)
- Handle empty document
- Handle all-deleted scenario

### Service Layer

**editable_view_test.go:**
- GetEditableDocument returns correct format
- ValidateDocument catches errors
- ApplyChanges persists correctly
- Transaction rollback on error

### Integration Tests

- Full round-trip: load → edit → save → reload
- Migration creates entry on target date
- Deletion uses event sourcing (can restore)

---

## Edge Cases

### Parsing

| Case | Behavior |
|------|----------|
| Empty line | Skip (not an entry) |
| Only whitespace | Skip |
| Unknown symbol | Error: "Unknown entry type" |
| Tab indentation | Normalize to spaces |
| Odd indentation (3 spaces) | Round to nearest level |
| Missing content | Error: "Entry content required" |
| Multiple `>` symbols | First `>` is symbol, rest is content |

### Hierarchy

| Case | Behavior |
|------|----------|
| Child without parent | Error: "Orphan entry at line X" |
| Deep nesting (10+ levels) | Allow (no limit) |
| Reparent to self | Error: "Cannot parent to self" |
| Circular reference | Error: "Circular reference detected" |

### Migration

| Case | Behavior |
|------|----------|
| Past date | Warning: "Migrating to past date" |
| Invalid date | Error: "Cannot parse date" |
| Same date | Warning: "Migrating to same day" |
| Migrate cancelled entry | Error: "Cannot migrate cancelled entry" |

### Deletion

| Case | Behavior |
|------|----------|
| Delete parent with children | Children move to pending deletion too |
| Restore parent only | Children remain in pending |
| Delete all entries | Valid (empty day) |

---

## Open Items

1. **Answered questions:** Currently `★` symbol. In editable view, use `?` with indented `-` for answer. Need to handle display of existing answered questions.

2. **Multi-line content:** Current spec assumes single-line entries. May need continuation syntax for long notes.

3. **Editor choice (Frontend):** CodeMirror vs Monaco vs textarea. CodeMirror lighter weight, Monaco more features.

4. **Syntax highlighting:** Custom grammar for entry format. May be overkill for v1.

---

## Success Criteria

1. User can edit entry content by directly modifying text
2. User can change entry type by changing prefix symbol
3. User can change priority by adding/removing `!` markers
4. User can create new entries by typing on new lines
5. User can delete entries (move to pending section)
6. User can migrate entries using `>[date]` syntax
7. User can restore deleted entries before save
8. Invalid syntax highlighted in red with error message
9. Save blocked until all errors resolved
10. Changes persisted atomically on save
11. Event sourcing preserves full history
