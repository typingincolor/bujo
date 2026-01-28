# Editable Journal View Design

**Date:** 2026-01-28
**Status:** Ready for Implementation
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

**Priority:** Suffix with `!`, `!!`, or `!!!` after symbol:
```
. !!! Urgent task
. !! High priority
. ! Low priority
. Normal task
```

**Priority display (Frontend):** Priority markers render as styled numeric labels:
- `!!!` → label showing "1" (highest)
- `!!` → label showing "2"
- `!` → label showing "3" (lowest)

**Hierarchy:** 2-space indentation per level:
```
. Parent task
  . Child task
    . Grandchild
  - Sibling note
```

**Hierarchy controls (Frontend):**
- Tab/Shift+Tab to indent/outdent entries
- Visual indent guides (faint vertical lines) show hierarchy levels

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
. !! Urgent: fix prod bug
x Deployed hotfix
? How does the auth flow work
  - See docs/auth.md
>[tomorrow] . Schedule follow-up
```

### Sync Behavior

- **Draft mode:** All changes stay in local buffer until explicit save
- **Trigger:** Save on Ctrl+S (Frontend) / Ctrl+S (TUI)
- **Validation:** Full document parsed before sync; errors block save
- **Scope:** One day at a time (not multi-day editing)

### Crash Recovery (Frontend)

Auto-save draft to localStorage on every change:

1. On edit: debounce (500ms) save to `localStorage.bujo.draft.{date}`
2. Store: `{ document: string, deletedIds: string[], timestamp: number }`
3. On load: check for draft newer than last save timestamp
4. If draft exists: prompt "Restore unsaved changes from {time}?" with [Restore] [Discard]
5. On successful save: clear localStorage draft
6. Stale drafts (>7 days) cleaned up on app start

### ID Handling

- **Hidden:** Entry IDs not shown in editable view
- **Tracked:** Backend provides line → EntityID mapping on load
- **New entries:** Lines without mapping get new EntityID on save
- **Moves:** Line position changes tracked to update hierarchy

### Deletion Behavior (Frontend)

1. User deletes a line → line removed immediately (clean editing view)
2. Standard Ctrl+Z undo stack for immediate recovery
3. On Cmd+S: if deletions exist, show deletion review dialog
4. Dialog shows checkboxes for each deleted item; unchecking restores the entry
5. Restored entries return to their original position in the document
6. "Discard All Changes" button reverts entire document to last saved state
7. On save: confirmed deletions become soft deletes (event sourcing)
8. Post-save: can still `restore <id>` via CLI

**Deletion review dialog:**
```
┌─────────────────────────────────────────────────────┐
│  Save changes to Monday, Jan 27?                    │
│                                                     │
│  The following entries will be deleted:             │
│                                                     │
│  ☑️  Call dentist                                   │
│  ☑️  Old task                                       │
│                                                     │
│  Uncheck to restore before saving.                  │
│                                                     │
│  [Discard All Changes]          [Cancel]   [Save]   │
└─────────────────────────────────────────────────────┘
```

If no deletions exist, save proceeds without dialog.

### Error Handling

**Errors vs Warnings:**
- **Errors** (block save): Unknown symbol, orphan child, circular reference, missing content
- **Warnings** (allow save): Past migration date, same-day migration, odd indentation

**Display:**
- Invalid lines highlighted in red
- Error/warning message shown inline or in status bar
- Quick-fix suggestions where possible (e.g., `Unknown symbol "^" → [Delete line] [Change to "." (task)]`)

**Escape hatches:**
- Standard Ctrl+Z undo for immediate recovery
- "Discard All Changes" in save dialog reverts to last saved state

### Save Feedback (Frontend)

- **Unsaved indicator:** Dot in title/header when unsaved changes exist (e.g., `● Journal - Jan 27`)
- **Success confirmation:** Status bar at bottom shows "✓ Saved" with timestamp after successful save
- No popups or fading animations

### Discoverability

- Syntax reference (symbols, priority, migration) added to existing keyboard shortcut popup
- Users access via help icon or keyboard shortcut

### Parsing Architecture

Parsing happens in **both** frontend and backend with distinct roles:

**Frontend (TypeScript):** Lightweight parsing for real-time UX
- Syntax highlighting in editor
- Red highlighting on invalid lines as user types
- Quick-fix suggestion display
- Does not need to handle all edge cases

**Backend (Go):** Authoritative parsing on save
- Full validation with all edge cases
- Diff computation against existing entries
- Source of truth for what gets persisted
- Shared with TUI implementation

**Frontend parser scope:**

| Feature | Frontend handles | Backend handles |
|---------|-----------------|-----------------|
| Symbol recognition | ✓ | ✓ |
| Priority parsing | ✓ | ✓ |
| Indentation depth | ✓ | ✓ |
| Migration syntax | ✓ (pattern detection) | ✓ (date resolution) |
| Orphan detection | ✗ | ✓ |
| Circular references | ✗ | ✓ |
| Entity ID mapping | ✗ | ✓ |
| Diff computation | ✗ | ✓ |

**Migration date resolution:**

Frontend calls backend to resolve natural language dates, ensuring consistency:

1. User types `>[tomorrow]`
2. Frontend detects `>[...]` pattern, extracts date string
3. Frontend calls `ResolveDate("tomorrow")` API
4. Backend returns `{ iso: "2026-01-29", display: "Wed, Jan 29" }`
5. Frontend shows resolved date as inline hint
6. On save, backend uses same `dateutil.ParseFuture()` - guaranteed match

This avoids duplicating date parsing logic and ensures what user sees matches what gets saved.

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
3. First non-space char is symbol (`.`, `-`, `o`, `x`, `~`, `?`, `>`)
4. Symbol determines entry type
5. If symbol is `>`, parse `[date]` portion
6. After symbol, check for priority markers (`!`, `!!`, `!!!`)
7. Remainder is content

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

// ResolveDate parses a natural language date and returns ISO + display format
func (a *App) ResolveDate(input string) (*DateResolution, error) {
    resolved, err := dateutil.ParseFuture(input)
    if err != nil {
        return nil, err
    }
    return &DateResolution{
        ISO:     resolved.Format("2006-01-02"),
        Display: resolved.Format("Mon, Jan 2"),
    }, nil
}

type DateResolution struct {
    ISO     string `json:"iso"`     // e.g., "2026-01-29"
    Display string `json:"display"` // e.g., "Wed, Jan 29"
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
| 3.5 | `app.go` | Implement ResolveDate API for migration date preview |
| 3.6 | TypeScript | Generate TypeScript types |

### Phase 4: Frontend Implementation

| Task | File | Description |
|------|------|-------------|
| 4.1 | `editableParser.ts` | Lightweight parser for real-time validation and highlighting |
| 4.2 | `useEditableDocument.ts` | State management hook (tracks deletions, dirty state) |
| 4.3 | `EditableJournalView.tsx` | Main component with text editor |
| 4.4 | Editor integration | CodeMirror/Monaco with syntax highlighting |
| 4.5 | Visual features | Priority labels (1,2,3), indent guides, error highlighting |
| 4.6 | Migration date preview | Call ResolveDate API, show resolved date inline |
| 4.7 | `DeletionReviewDialog.tsx` | Save dialog with deletion checkboxes |
| 4.8 | Save feedback | Unsaved dot indicator, status bar with timestamp |
| 4.9 | Quick-fix suggestions | Inline error actions (delete line, change symbol) |
| 4.10 | Keyboard shortcuts | Cmd+S save, Tab/Shift+Tab indent, Escape exit |
| 4.11 | Help integration | Add syntax reference to keyboard shortcut popup |
| 4.12 | Crash recovery | localStorage auto-save + restore prompt on load |

### Phase 5: TUI Implementation

TUI follows Frontend patterns where possible. TUI-specific adaptations (e.g., deletion review without dialogs, date preview without inline hints) determined during implementation.

| Task | File | Description |
|------|------|-------------|
| 5.1 | `editable_mode.go` | Define editableViewState struct |
| 5.2 | `editable_mode.go` | Render editable text buffer |
| 5.3 | `editable_mode.go` | Text editing (insert, delete, cursor) |
| 5.4 | `editable_mode.go` | Deletion handling (TUI-specific flow TBD) |
| 5.5 | Key bindings | `e` toggle, Ctrl+S save, Escape exit |
| 5.6 | Validation display | Red/error highlighting for invalid lines |
| 5.7 | Integration | Hook into main TUI view switching |

### Phase 6: Cleanup (Dead Code Removal)

Remove obsolete UI code replaced by the editable view.

| Task | File | Description |
|------|------|-------------|
| 6.1 | Frontend | Remove action buttons (edit, delete, migrate) from journal view |
| 6.2 | Frontend | Remove entry edit modal/dialog components |
| 6.3 | Frontend | Remove inline edit handlers from JournalEntryItem |
| 6.4 | TUI | Remove separate edit mode commands |
| 6.5 | Wails | Remove deprecated API endpoints (if any) |
| 6.6 | Tests | Remove/update tests for removed components |
| 6.7 | Verify | Ensure no orphaned imports, unused types |

**Note:** Identify specific files during implementation. Some components may still be needed for other views (e.g., weekly review).

---

## Testing Strategy

### Domain Layer (100% coverage required)

**editable_parser_test.go:**
- Parse each symbol type (., -, o, x, ~, ?)
- Parse priority levels after symbol (`. !`, `. !!`, `. !!!`)
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
| Delete parent with children | Children also marked as deleted |
| Restore parent in dialog | Children remain deleted (user must uncheck each) |
| Restore entry | Returns to original line position |
| Delete all entries | Valid (empty day) |
| No deletions on save | Dialog skipped, save proceeds immediately |

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
3. User can change priority by adding/removing `!` markers (displayed as styled 1/2/3 labels)
4. User can create new entries by typing on new lines
5. User can delete entries (removed immediately, reviewed in save dialog)
6. User can migrate entries using `>[date]` syntax
7. User can restore deleted entries via checkbox in save dialog (returns to original position)
8. User can indent/outdent with Tab/Shift+Tab; visual indent guides show hierarchy
9. Invalid syntax highlighted in red with quick-fix suggestions
10. Errors block save; warnings allow save with notice
11. Unsaved changes indicated by dot in title; successful save shown in status bar
12. Changes persisted atomically on save
13. Event sourcing preserves full history
14. Syntax reference available in keyboard shortcut popup
15. Unsaved changes recovered after crash/reboot via localStorage
16. Obsolete UI components removed (action buttons, edit modals)
