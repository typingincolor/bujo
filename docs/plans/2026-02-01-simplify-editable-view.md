# Simplify Editable Journal View — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the parse-diff-apply pipeline with replace-all-on-save, reducing ~5,500 lines / 22 files to ~2,000 lines / 12 files.

**Architecture:** On save, hard-delete all current entries for the date and re-insert from parsed text in document order. Insertion order = display order. No entity IDs in document text. No diff engine. No event-sourced deletes for this operation — replace-all-on-save makes per-entry audit trails meaningless.

**Tech Stack:** Go 1.23, SQLite, Wails, React, CodeMirror 6, TypeScript

---

### Task 1: Delete backend files that are no longer needed

Delete the diff engine, migration syntax, and their tests.

**Files:**
- Delete: `internal/domain/document_diff.go`
- Delete: `internal/domain/document_diff_test.go`
- Delete: `internal/domain/migration_syntax.go`
- Delete: `internal/domain/migration_syntax_test.go`

**Step 1: Run existing tests to confirm green baseline**

Run: `go test ./internal/domain/... -count=1`
Expected: All tests pass (some may reference deleted code)

**Step 2: Delete the files**

```bash
rm internal/domain/document_diff.go internal/domain/document_diff_test.go
rm internal/domain/migration_syntax.go internal/domain/migration_syntax_test.go
```

**Step 3: Fix compilation errors**

After deletion, `editable_parser.go` calls `ParseMigrationSyntax` — remove that call. In `editable_view.go`, `ComputeDiff` is called — this will be addressed in Task 4 when we rewrite the service. For now, comment out or stub the call so domain tests pass.

Remove from `editable_parser.go` (lines 60-67):

```go
// DELETE this block:
content, migrateDate, err := ParseMigrationSyntax(line, p.dateParser)
if err == nil && migrateDate != nil {
    parsedLine.MigrateTarget = migrateDate
    innerLine := p.ParseLine(content, lineNum)
    parsedLine.Symbol = innerLine.Symbol
    parsedLine.Content = innerLine.Content
    parsedLine.Priority = innerLine.Priority
}
```

**Step 4: Run domain tests**

Run: `go test ./internal/domain/... -count=1`
Expected: PASS (service tests may fail — that's expected, we fix them in Task 4)

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: remove diff engine and migration syntax files"
```

---

### Task 2: Simplify domain types in `editable_document.go`

Remove entity ID, migration, diff, and changeset types from the domain.

**Files:**
- Modify: `internal/domain/editable_document.go`

**Step 1: Rewrite `editable_document.go`**

Replace the entire file with:

```go
package domain

import "time"

type ParsedLine struct {
	LineNumber   int
	Raw          string
	Depth        int
	Symbol       EntryType
	Priority     Priority
	Content      string
	IsValid      bool
	IsHeader     bool
	ErrorMessage string
}

type EditableDocument struct {
	Date  time.Time
	Lines []ParsedLine
}

type ParseError struct {
	LineNumber int
	Message    string
}
```

**Step 2: Fix compilation — update `editable_parser.go`**

The `Parse` function references `OriginalMapping`, `EntityID`, `MigrateTarget`, and content-matching logic. Simplify it:

```go
func (p *EditableDocumentParser) Parse(input string, existing []Entry) (*EditableDocument, error) {
	lines := strings.Split(input, "\n")
	doc := &EditableDocument{
		Lines: make([]ParsedLine, 0, len(lines)),
	}

	for i, line := range lines {
		lineNum := i + 1

		if strings.TrimSpace(line) == "" {
			doc.Lines = append(doc.Lines, ParsedLine{
				LineNumber: lineNum,
				Raw:        line,
				IsValid:    false,
			})
			continue
		}

		parsedLine := p.ParseLine(line, lineNum)
		doc.Lines = append(doc.Lines, parsedLine)
	}

	return doc, nil
}
```

Also remove the `dateParser` field from `EditableDocumentParser` since migration syntax is gone:

```go
type EditableDocumentParser struct{}

func NewEditableDocumentParser() *EditableDocumentParser {
	return &EditableDocumentParser{}
}
```

Update `ParseLine` to remove entity ID bracket parsing (lines 202-209). Remove the `if strings.HasPrefix(rest, "[")` block entirely.

Remove the `MigrateTarget` field usage — already handled by removing the migration syntax call.

**Step 3: Fix `Serialize` — remove entity ID output**

In `serializeEntryLine`, remove lines 166-169:

```go
// DELETE this block:
if !entry.EntityID.IsEmpty() {
    result.WriteString("[")
    result.WriteString(entry.EntityID.String())
    result.WriteString("] ")
}
```

**Step 4: Run domain tests**

Run: `go test ./internal/domain/... -count=1`
Expected: Parser tests pass. Some tests referencing entity IDs or migration syntax in editable_parser_test.go may need updating.

**Step 5: Fix any failing parser tests**

Update tests that expected entity IDs in parsed output or serialized strings.

**Step 6: Commit**

```bash
git add -A
git commit -m "refactor: simplify domain types, remove entity IDs from document model"
```

---

### Task 3: Add `DeleteByDate` to repository

**Files:**
- Modify: `internal/domain/repository.go`
- Modify: `internal/repository/sqlite/entry_repository.go`
- Test: `internal/repository/sqlite/entry_repository_test.go` (or a new test if needed)

**Step 1: Write failing test for `DeleteByDate`**

RED: Writing failing test for DeleteByDate.

```go
func TestEntryRepository_DeleteByDate(t *testing.T) {
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	defer db.Close()

	repo := NewEntryRepository(db)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	entry1 := domain.Entry{
		Type:          domain.EntryTypeTask,
		Content:       "Task one",
		ScheduledDate: &date,
		CreatedAt:     time.Now(),
	}
	entry2 := domain.Entry{
		Type:          domain.EntryTypeNote,
		Content:       "Note two",
		ScheduledDate: &date,
		CreatedAt:     time.Now(),
	}

	_, err = repo.Insert(ctx, entry1)
	require.NoError(t, err)
	_, err = repo.Insert(ctx, entry2)
	require.NoError(t, err)

	entries, err := repo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	count, err := repo.DeleteByDate(ctx, date)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	entries, err = repo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Empty(t, entries)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/repository/sqlite/... -run TestEntryRepository_DeleteByDate -v`
Expected: FAIL — `DeleteByDate` method doesn't exist

**Step 3: Add interface method**

In `internal/domain/repository.go`, add to `EntryRepository` interface:

```go
DeleteByDate(ctx context.Context, date time.Time) (int, error)
```

Remove `UpdateSortOrders` from the interface at the same time:

```go
// DELETE this line:
UpdateSortOrders(ctx context.Context, orders map[EntityID]int) error
```

**Step 4: Implement `DeleteByDate`**

Hard-delete all current entries for the date. No event-sourced DELETE rows — with replace-all-on-save, per-entry audit trails are noise.

In `internal/repository/sqlite/entry_repository.go`:

```go
func (r *EntryRepository) DeleteByDate(ctx context.Context, date time.Time) (int, error) {
	dateStr := date.Format("2006-01-02")

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM entries WHERE scheduled_date = ? AND (valid_to IS NULL OR valid_to = '')
	`, dateStr)
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}
```

**Step 5: Remove `UpdateSortOrders` implementation**

Delete the `UpdateSortOrders` method from `entry_repository.go` (lines 428-445).

**Step 6: Run test to verify it passes**

Run: `go test ./internal/repository/sqlite/... -run TestEntryRepository_DeleteByDate -v`
Expected: PASS

**Step 7: Write test for DeleteByDate preserving other dates**

```go
func TestEntryRepository_DeleteByDate_PreservesOtherDates(t *testing.T) {
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	defer db.Close()

	repo := NewEntryRepository(db)
	ctx := context.Background()
	date1 := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	_, err = repo.Insert(ctx, domain.Entry{
		Type: domain.EntryTypeTask, Content: "Day 1 task",
		ScheduledDate: &date1, CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	_, err = repo.Insert(ctx, domain.Entry{
		Type: domain.EntryTypeTask, Content: "Day 2 task",
		ScheduledDate: &date2, CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	count, err := repo.DeleteByDate(ctx, date1)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	entries, err := repo.GetByDate(ctx, date2)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "Day 2 task", entries[0].Content)
}
```

Run: `go test ./internal/repository/sqlite/... -run TestEntryRepository_DeleteByDate -v`
Expected: PASS

**Step 8: Commit**

```bash
git add -A
git commit -m "feat: add DeleteByDate repository method with hard deletes"
```

---

### Task 3b: Drop event sourcing from entry repository

Remove event sourcing from entries entirely. Simplify Insert/Update/Delete to direct SQL operations. Remove ES-specific methods from interface and implementation. Update all callers.

**Scope:** Only entries. Other repositories (habits, lists, day_context, list_items, goals) keep event sourcing.

**Files:**
- Modify: `internal/domain/repository.go` — remove `GetByEntityID`, `GetHistory`, `GetAsOf`, `GetDeleted`, `Restore`, `UpdateSortOrders` from `EntryRepository` interface; add `DeleteByDate`
- Modify: `internal/repository/sqlite/entry_repository.go` — simplify Insert (plain INSERT), Update (direct UPDATE), Delete (direct DELETE), remove `valid_to`/`op_type` filters from queries, remove ES-specific methods
- Modify: `internal/service/bujo.go` — remove `GetDeletedEntries`/`RestoreEntry` methods; fix `UpdateEntry` to not use `GetHistory`; fix `updateChildrenDepths`/`updateChildrenDates` to not use `GetByEntityID` (with simplified Update, row ID doesn't change)
- Modify: `internal/service/export.go` — remove `GetByEntityID` from `ImportEntryRepository` interface; change merge logic to use `GetByID` or just insert
- Modify: `internal/adapter/wails/app.go` — remove Wails bindings for `GetDeletedEntries`/`RestoreEntry` if they exist
- Delete tests: remove/rewrite ES-specific tests in `entry_repository_ops_test.go` and `entry_repository_crud_test.go`
- Update: `internal/service/bujo_context_test.go` — remove `TestBujoService_GetDeletedEntries` test
- Update: `internal/service/export_test.go` — remove `GetByEntityID` from mock entry repo
- Update: `internal/tui/acceptance_journal_test.go` — remove/update deleted entries UAT

**Step 1: Simplify `EntryRepository` interface in `repository.go`**

Remove these methods from the `EntryRepository` interface:
```go
// DELETE these lines:
GetByEntityID(ctx context.Context, entityID EntityID) (*Entry, error)
GetHistory(ctx context.Context, entityID EntityID) ([]Entry, error)
GetAsOf(ctx context.Context, entityID EntityID, asOf time.Time) (*Entry, error)
GetDeleted(ctx context.Context) ([]Entry, error)
Restore(ctx context.Context, entityID EntityID) (int64, error)
UpdateSortOrders(ctx context.Context, orders map[EntityID]int) error
```

Add:
```go
DeleteByDate(ctx context.Context, date time.Time) (int, error)
```

**Step 2: Simplify `entry_repository.go`**

- `Insert`: Remove entity_id UUID generation, version, valid_from, op_type. Plain INSERT with the provided fields. Still populate entity_id/version/valid_from/op_type columns with defaults for schema compatibility (empty entity_id, version=1, valid_from=now, op_type='INSERT').
- `GetByID`: Simplify to single query `SELECT ... WHERE id = ?` (no entity_id lookup, no valid_to/op_type filter).
- `GetByDate`: Remove `valid_to IS NULL AND op_type != 'DELETE'` filter. Keep `ORDER BY created_at, id`.
- `GetByDateRange`, `GetAll`, `GetOverdue`, `GetWithChildren`, `GetChildren`, `Search`: Remove valid_to/op_type filters.
- `Update`: Direct `UPDATE entries SET ... WHERE id = ?` instead of close-current-version + insert-new-version. The row ID stays the same.
- `Delete`: Direct `DELETE FROM entries WHERE id = ?` instead of creating a DELETE version row.
- `DeleteAll`: Direct `DELETE FROM entries` (no filter needed).
- `DeleteWithChildren`: Direct DELETE instead of event-sourced delete.
- Remove methods: `GetByEntityID`, `GetHistory`, `GetAsOf`, `GetDeleted`, `Restore`, `UpdateSortOrders`.
- Add: `DeleteByDate` (already defined in Task 3).

**Step 3: Update `bujo.go` callers**

1. `UpdateEntry` (line ~392): Remove `GetHistory` call. When uncancelling, just default to `EntryTypeTask` instead of looking at history:
```go
// REPLACE the GetHistory block with:
previousType := domain.EntryTypeTask
```

2. `EditEntry` (line ~640): Remove `GetByEntityID` call. With direct UPDATE, the row ID doesn't change, so use the original entry's ID:
```go
// REPLACE:
//   updatedEntry, err := s.entryRepo.GetByEntityID(ctx, entityID)
//   newID := updatedEntry.ID
// WITH:
newID := entry.ID
```

3. `updateChildrenDepths` (line ~674): Same pattern — after Update, use original child.ID:
```go
// REPLACE:
//   updatedChild, err := s.entryRepo.GetByEntityID(ctx, entityID)
//   s.updateChildrenDepths(ctx, updatedChild.ID, depthDelta)
// WITH:
s.updateChildrenDepths(ctx, child.ID, depthDelta)
```

4. `updateChildrenDates` (line ~698): Same pattern:
```go
// REPLACE:
//   updatedChild, err := s.entryRepo.GetByEntityID(ctx, entityID)
//   s.updateChildrenDates(ctx, updatedChild.ID, newDate)
// WITH:
s.updateChildrenDates(ctx, child.ID, newDate)
```

5. Remove `GetDeletedEntries` and `RestoreEntry` methods (lines 756-762).

**Step 4: Update `export.go`**

Remove `GetByEntityID` from `ImportEntryRepository` interface. For merge imports, entries without entity IDs can't be deduplicated — just always insert:

```go
type ImportEntryRepository interface {
	Insert(ctx context.Context, entry domain.Entry) (int64, error)
	DeleteAll(ctx context.Context) error
}
```

Update import loop to remove the merge check for entries:
```go
for _, entry := range data.Entries {
	if _, err := s.entryRepo.Insert(ctx, entry); err != nil {
		return err
	}
}
```

**Step 5: Update `export_test.go` mock**

Remove `GetByEntityID` from `mockImportEntryRepo`.

**Step 6: Update Wails adapter**

Remove `GetDeletedEntries` and `RestoreEntry` bindings from `app.go` if they exist.

**Step 7: Update entry repository tests**

- `entry_repository_ops_test.go`: Delete tests for `GetByEntityID`, `GetHistory`, `GetAsOf`, `GetDeleted`, `Restore`. Update the Update test to verify in-place update (same ID returned). Update the Delete test to verify row is actually removed.
- `entry_repository_crud_test.go`: Delete `TestEntryRepository_UpdateSortOrders` test. Update the `GetByDate` ordering test.
- `bujo_context_test.go`: Delete `TestBujoService_GetDeletedEntries` test and the restore test.
- `acceptance_journal_test.go`: Remove or skip the "View and Restore Deleted Entries" UAT.

**Step 8: Run all Go tests**

Run: `go test ./... -count=1`
Expected: PASS

**Step 9: Commit**

```bash
git add -A
git commit -m "refactor: drop event sourcing from entries, simplify to direct CRUD operations"
```

---

### Task 4: Rewrite `EditableViewService`

Rewrite ApplyChanges to use replace-all-on-save. Remove `bujoService` dependency.

**Files:**
- Modify: `internal/service/editable_view.go`
- Rewrite: `internal/service/editable_view_test.go`
- Modify: `internal/app/factory.go:76`

**Step 1: Write failing tests for new ApplyChanges**

RED: Rewrite tests for the simplified service.

Replace `internal/service/editable_view_test.go` with:

```go
package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupEditableViewService(t *testing.T) (*EditableViewService, *BujoService, *sqlite.EntryRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	parser := domain.NewTreeParser()

	bujoService := NewBujoService(entryRepo, dayCtxRepo, parser)
	editableViewService := NewEditableViewService(entryRepo)
	return editableViewService, bujoService, entryRepo
}

func TestGetEditableDocument_EmptyDay(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Empty(t, doc)
}

func TestGetEditableDocument_SingleEntry(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(context.Background(), ". Buy groceries", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". Buy groceries")
}

func TestGetEditableDocument_MultipleEntries(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(context.Background(), ". Task one\n- Note two\no Event three", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". Task one")
	require.Contains(t, doc, "- Note two")
	require.Contains(t, doc, "o Event three")
}

func TestGetEditableDocument_WithHierarchy(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(context.Background(), ". Parent task\n  - Child note", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". Parent task")
	require.Contains(t, doc, "- Child note")
}

func TestGetEditableDocument_WithPriority(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	ids, err := bujoSvc.LogEntries(context.Background(), ". Important task", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(context.Background(), ids[0])
	require.NoError(t, err)
	entry.Priority = domain.PriorityHigh
	err = entryRepo.Update(context.Background(), *entry)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". !!! Important task")
}

func TestValidateDocument_ValidDocument(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument(". Task one\n- Note two")

	require.True(t, result.IsValid)
	require.Empty(t, result.Errors)
	require.Len(t, result.ParsedLines, 2)
}

func TestValidateDocument_InvalidLine(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument("invalid line without symbol")

	require.False(t, result.IsValid)
	require.Len(t, result.Errors, 1)
	require.Equal(t, 1, result.Errors[0].LineNumber)
}

func TestValidateDocument_OrphanChild(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument("  . Orphan child at depth 1")

	require.False(t, result.IsValid)
	require.Len(t, result.Errors, 1)
}

func TestValidateDocument_MixedValidInvalid(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument(". Valid task\ninvalid line\n- Valid note")

	require.False(t, result.IsValid)
	require.Len(t, result.Errors, 1)
	require.Equal(t, 2, result.Errors[0].LineNumber)
	require.Len(t, result.ParsedLines, 3)
}

func TestValidateDocument_EmptyDocument(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument("")

	require.True(t, result.IsValid)
	require.Empty(t, result.Errors)
	require.Empty(t, result.ParsedLines)
}

func TestApplyChanges_InsertEntries(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	result, err := svc.ApplyChanges(ctx, ". Task one\n- Note two", date)

	require.NoError(t, err)
	require.Equal(t, 2, result.Inserted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "Task one", entries[0].Content)
	require.Equal(t, "Note two", entries[1].Content)
}

func TestApplyChanges_ReplacesExistingEntries(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Old task\n- Old note", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	result, err := svc.ApplyChanges(ctx, ". New task\no New event", date)

	require.NoError(t, err)
	require.Equal(t, 2, result.Inserted)
	require.Equal(t, 2, result.Deleted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "New task", entries[0].Content)
	require.Equal(t, "New event", entries[1].Content)
}

func TestApplyChanges_ChildEntriesGetParentID(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	result, err := svc.ApplyChanges(ctx, ". Parent task\n  - Child note\n    - Grandchild note", date)

	require.NoError(t, err)
	require.Equal(t, 3, result.Inserted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	parent := entries[0]
	child := entries[1]
	grandchild := entries[2]

	require.Equal(t, "Parent task", parent.Content)
	require.Equal(t, "Child note", child.Content)
	require.Equal(t, "Grandchild note", grandchild.Content)

	require.Nil(t, parent.ParentID)
	require.NotNil(t, child.ParentID)
	require.Equal(t, parent.ID, *child.ParentID)
	require.NotNil(t, grandchild.ParentID)
	require.Equal(t, child.ID, *grandchild.ParentID)
}

func TestApplyChanges_SiblingsShareParent(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". Parent\n  - Sibling one\n  - Sibling two", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	require.NotNil(t, entries[1].ParentID)
	require.NotNil(t, entries[2].ParentID)
	require.Equal(t, entries[0].ID, *entries[1].ParentID)
	require.Equal(t, entries[0].ID, *entries[2].ParentID)
}

func TestApplyChanges_PreservesDocumentOrder(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". First\n- Second\no Third", date)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	lines := splitNonEmpty(doc)
	require.Len(t, lines, 3)
	require.Contains(t, lines[0], "First")
	require.Contains(t, lines[1], "Second")
	require.Contains(t, lines[2], "Third")
}

func TestApplyChanges_ReorderPreserved(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". First\n- Second\no Third", date)
	require.NoError(t, err)

	_, err = svc.ApplyChanges(ctx, "o Third\n. First\n- Second", date)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	lines := splitNonEmpty(doc)
	require.Len(t, lines, 3)
	require.Contains(t, lines[0], "Third")
	require.Contains(t, lines[1], "First")
	require.Contains(t, lines[2], "Second")
}

func TestApplyChanges_ValidationErrors(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, "invalid line", date)
	require.Error(t, err)
}

func TestApplyChanges_EmptyDocument(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Task to remove", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	result, err := svc.ApplyChanges(ctx, "", date)
	require.NoError(t, err)
	require.Equal(t, 0, result.Inserted)
	require.Equal(t, 1, result.Deleted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestApplyChanges_WithPriority(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". !!! High priority task", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, domain.PriorityHigh, entries[0].Priority)
	require.Equal(t, "High priority task", entries[0].Content)
}

func TestApplyChanges_RoundTrip(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	original := ". Parent task\n  - Child one\n  - Child two\no Standalone event"
	_, err := svc.ApplyChanges(ctx, original, date)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)

	_, err = svc.ApplyChanges(ctx, doc, date)
	require.NoError(t, err)

	doc2, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Equal(t, doc, doc2)
}

func splitNonEmpty(s string) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/service/... -count=1 -v`
Expected: FAIL — `NewEditableViewService` signature changed, `ApplyChanges` signature changed

**Step 3: Rewrite `editable_view.go`**

```go
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type EditableViewService struct {
	entryRepo domain.EntryRepository
}

func NewEditableViewService(entryRepo domain.EntryRepository) *EditableViewService {
	return &EditableViewService{
		entryRepo: entryRepo,
	}
}

func (s *EditableViewService) GetEditableDocument(ctx context.Context, date time.Time) (string, error) {
	entries, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return "", err
	}

	return domain.Serialize(entries), nil
}

type ValidationResult struct {
	IsValid     bool
	Errors      []domain.ParseError
	ParsedLines []domain.ParsedLine
}

func (s *EditableViewService) ValidateDocument(doc string) ValidationResult {
	parser := domain.NewEditableDocumentParser()
	parsed, err := parser.Parse(doc, nil)
	if err != nil {
		return ValidationResult{
			IsValid: false,
			Errors:  []domain.ParseError{{LineNumber: 0, Message: err.Error()}},
		}
	}

	errors := make([]domain.ParseError, 0)

	var parentStack []int
	for _, line := range parsed.Lines {
		if strings.TrimSpace(line.Raw) == "" {
			continue
		}

		if !line.IsValid && !line.IsHeader {
			errors = append(errors, domain.ParseError{
				LineNumber: line.LineNumber,
				Message:    line.ErrorMessage,
			})
		}

		if line.IsValid && !line.IsHeader {
			if line.Depth > 0 && len(parentStack) == 0 {
				errors = append(errors, domain.ParseError{
					LineNumber: line.LineNumber,
					Message:    "Orphan child: no parent at depth 0",
				})
			}

			for len(parentStack) > line.Depth {
				parentStack = parentStack[:len(parentStack)-1]
			}

			if line.Depth >= len(parentStack) {
				parentStack = append(parentStack, line.LineNumber)
			} else {
				parentStack[line.Depth] = line.LineNumber
				parentStack = parentStack[:line.Depth+1]
			}
		}
	}

	validLines := make([]domain.ParsedLine, 0)
	for _, line := range parsed.Lines {
		if strings.TrimSpace(line.Raw) != "" {
			validLines = append(validLines, line)
		}
	}

	return ValidationResult{
		IsValid:     len(errors) == 0,
		Errors:      errors,
		ParsedLines: validLines,
	}
}

type ApplyChangesResult struct {
	Inserted int
	Deleted  int
}

func (s *EditableViewService) ApplyChanges(ctx context.Context, doc string, date time.Time) (*ApplyChangesResult, error) {
	validation := s.ValidateDocument(doc)
	if !validation.IsValid {
		return nil, fmt.Errorf("validation failed: %s", validation.Errors[0].Message)
	}

	deleted, err := s.entryRepo.DeleteByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	result := &ApplyChangesResult{Deleted: deleted}

	// depthStack[d] = row ID of the most recent entry at depth d
	var depthStack []int64

	for _, line := range validation.ParsedLines {
		if !line.IsValid || line.IsHeader {
			continue
		}

		entry := domain.Entry{
			Type:          line.Symbol,
			Content:       line.Content,
			Priority:      line.Priority,
			Depth:         line.Depth,
			ScheduledDate: &date,
			CreatedAt:     time.Now(),
		}

		if line.Depth > 0 && len(depthStack) > line.Depth-1 {
			parentID := depthStack[line.Depth-1]
			entry.ParentID = &parentID
		}

		rowID, err := s.entryRepo.Insert(ctx, entry)
		if err != nil {
			return nil, err
		}

		if line.Depth >= len(depthStack) {
			depthStack = append(depthStack, rowID)
		} else {
			depthStack[line.Depth] = rowID
			depthStack = depthStack[:line.Depth+1]
		}

		result.Inserted++
	}

	return result, nil
}
```

**Step 4: Update factory**

In `internal/app/factory.go:76`, change:

```go
// FROM:
EditableView: service.NewEditableViewService(entryRepo, bujoService),
// TO:
EditableView: service.NewEditableViewService(entryRepo),
```

**Step 5: Update `GetByDate` ordering**

In `internal/repository/sqlite/entry_repository.go`, change the `GetByDate` ORDER BY:

```go
// FROM:
ORDER BY sort_order, created_at, entity_id
// TO:
ORDER BY created_at, id
```

**Step 6: Run tests**

Run: `go test ./internal/service/... -count=1 -v`
Expected: PASS

**Step 7: Commit**

```bash
git add -A
git commit -m "feat: rewrite EditableViewService with replace-all-on-save"
```

---

### Task 5: Update Wails adapter

Simplify the Wails bindings — remove `pendingDeletes` param, remove `GetEditableDocumentWithEntries`, simplify `ApplyResult`.

**Files:**
- Modify: `internal/adapter/wails/app.go`

**Step 1: Update Wails adapter**

Remove `EditableEntryInfo` struct (line 441-444), `EditableDocumentWithEntries` struct (line 446-449), `GetEditableDocumentWithEntries` method (line 455-480).

Update `ApplyResult` to remove `Updated` and `Migrated` fields:

```go
type ApplyResult struct {
	Inserted int `json:"inserted"`
	Deleted  int `json:"deleted"`
}
```

Simplify `ApplyEditableDocument`:

```go
func (a *App) ApplyEditableDocument(doc string, date time.Time) (*ApplyResult, error) {
	result, err := a.services.EditableView.ApplyChanges(a.ctx, doc, date)
	if err != nil {
		return nil, err
	}

	return &ApplyResult{
		Inserted: result.Inserted,
		Deleted:  result.Deleted,
	}, nil
}
```

**Step 2: Run backend tests**

Run: `go test ./... -count=1`
Expected: PASS

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: simplify Wails editable document bindings"
```

---

### Task 6: Delete frontend files no longer needed

**Files:**
- Delete: `frontend/src/lib/editableParser.ts`
- Delete: `frontend/src/lib/__tests__/editableParser.test.ts` (if exists)
- Delete: `frontend/src/lib/codemirror/entityIdHider.ts`
- Delete: `frontend/src/lib/codemirror/__tests__/entityIdHider.test.ts` (if exists)
- Delete: `frontend/src/lib/codemirror/migrationDatePreview.ts`
- Delete: `frontend/src/lib/codemirror/__tests__/migrationDatePreview.test.ts` (if exists)
- Delete: `frontend/src/components/bujo/DeletionReviewDialog.tsx`
- Delete: `frontend/src/components/bujo/__tests__/DeletionReviewDialog.test.tsx` (if exists)

**Step 1: Delete the files**

```bash
rm -f frontend/src/lib/editableParser.ts
rm -f frontend/src/lib/__tests__/editableParser.test.ts
rm -f frontend/src/lib/codemirror/entityIdHider.ts
rm -f frontend/src/lib/codemirror/__tests__/entityIdHider.test.ts
rm -f frontend/src/lib/codemirror/migrationDatePreview.ts
rm -f frontend/src/lib/codemirror/__tests__/migrationDatePreview.test.ts
rm -f frontend/src/components/bujo/DeletionReviewDialog.tsx
rm -f frontend/src/components/bujo/__tests__/DeletionReviewDialog.test.tsx
```

**Step 2: Commit**

```bash
git add -A
git commit -m "chore: delete unused frontend files (entityIdHider, editableParser, DeletionReviewDialog, migrationDatePreview)"
```

---

### Task 7: Simplify `bujoFolding.ts`

Remove clipboard handlers, entity ID functions. Keep fold logic only.

**Files:**
- Modify: `frontend/src/lib/codemirror/bujoFolding.ts`

**Step 1: Rewrite `bujoFolding.ts`**

```typescript
import { EditorState, Extension } from '@codemirror/state'
import { foldService, foldedRanges } from '@codemirror/language'
import { codeFolding, foldGutter, foldKeymap } from '@codemirror/language'
import { keymap } from '@codemirror/view'

export interface FoldRange {
  from: number
  to: number
}

const INDENT_SIZE = 2

function getIndentDepth(line: string): number {
  const leadingSpaces = line.match(/^(\s*)/)?.[1].length ?? 0
  return Math.floor(leadingSpaces / INDENT_SIZE)
}

export function getFoldRange(lines: string[], lineIndex: number): FoldRange | null {
  const currentLine = lines[lineIndex]
  if (!currentLine || currentLine.trim() === '') return null
  if (lineIndex >= lines.length - 1) return null

  const currentDepth = getIndentDepth(currentLine)
  let lastChildIndex = lineIndex

  for (let i = lineIndex + 1; i < lines.length; i++) {
    const line = lines[i]
    if (line.trim() === '') continue

    const depth = getIndentDepth(line)
    if (depth <= currentDepth) break

    lastChildIndex = i
  }

  if (lastChildIndex === lineIndex) return null

  return { from: lineIndex, to: lastChildIndex }
}

function bujoFoldServiceFn(state: EditorState, lineStart: number, _lineEnd: number): { from: number; to: number } | null {
  const doc = state.doc
  const lines: string[] = []
  for (let i = 1; i <= doc.lines; i++) {
    lines.push(doc.line(i).text)
  }

  const lineObj = doc.lineAt(lineStart)
  const lineIndex = lineObj.number - 1

  const range = getFoldRange(lines, lineIndex)
  if (!range) return null

  return {
    from: lineObj.to,
    to: doc.line(range.to + 1).to,
  }
}

export function expandRangeForFolds(state: EditorState, from: number, to: number): { from: number; to: number } {
  const folded = foldedRanges(state)
  if (folded.size === 0) return { from, to }

  let expandedFrom = from
  let expandedTo = to

  const cursor = folded.iter()
  while (cursor.value) {
    if (cursor.from >= expandedFrom && cursor.from <= expandedTo) {
      expandedTo = Math.max(expandedTo, cursor.to)
    }
    if (cursor.to >= expandedFrom && cursor.to <= expandedTo) {
      expandedFrom = Math.min(expandedFrom, cursor.from)
    }
    cursor.next()
  }

  return { from: expandedFrom, to: expandedTo }
}

export function bujoFoldExtension(): Extension {
  return [
    foldService.of(bujoFoldServiceFn),
    codeFolding(),
    foldGutter(),
    keymap.of(foldKeymap),
  ]
}
```

**Step 2: Run frontend tests if available**

Run: `cd frontend && npx vitest run --reporter=verbose 2>&1 | head -50`
Expected: Tests related to folding still pass; deleted file tests fail (expected since we deleted them)

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: simplify bujoFolding to fold logic only, remove clipboard handlers"
```

---

### Task 8: Simplify `BujoEditor.tsx`

Remove entity ID hider, migration date preview, and related props.

**Files:**
- Modify: `frontend/src/lib/codemirror/BujoEditor.tsx`

**Step 1: Update imports and remove deleted extensions**

Remove these imports:
```typescript
// DELETE:
import { entityIdHiderExtension, entityIdAtomicRanges } from './entityIdHider'
import {
  migrationDatePreviewExtension,
  setResolvedDates,
  findMigrationDates,
  ResolvedDateInfo,
} from './migrationDatePreview'
import type { DocumentError } from '../editableParser'
```

Remove `resolveDate` prop and all migration date resolution logic (the `resolveMigrationDates` callback, `useEffect` for it, `lastResolvedValueRef`, `resolvedCacheRef`).

Remove `entityIdHiderExtension()`, `entityIdAtomicRanges()`, `migrationDatePreviewExtension()` from the extensions array.

**Step 2: Run frontend build check**

Run: `cd frontend && npx tsc --noEmit 2>&1 | head -30`
Expected: No errors from BujoEditor.tsx

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: simplify BujoEditor, remove entity ID hider and migration date preview"
```

---

### Task 9: Simplify `useEditableDocument.ts`

Rewrite to ~100 lines. Remove deletion tracking, entry mappings, debug infrastructure, mismatch detection.

**Files:**
- Modify: `frontend/src/hooks/useEditableDocument.ts`

**Step 1: Rewrite the hook**

```typescript
import { useState, useEffect, useCallback, useRef } from 'react'
import {
  GetEditableDocument,
  ValidateEditableDocument,
  ApplyEditableDocument,
} from '../wailsjs/go/wails/App'
import { toWailsTime } from '@/lib/wailsTime'

export interface ValidationError {
  lineNumber: number
  message: string
}

export interface ApplyResult {
  inserted: number
  deleted: number
}

export interface SaveResult {
  success: boolean
  error?: string
  result?: ApplyResult
}

export interface EditableDocumentState {
  document: string
  setDocument: (doc: string) => void
  isLoading: boolean
  error: string | null
  isDirty: boolean
  validationErrors: ValidationError[]
  save: () => Promise<SaveResult>
  discardChanges: () => void
  lastSaved: Date | null
  hasDraft: boolean
  restoreDraft: () => void
  discardDraft: () => void
}

const DEBOUNCE_MS = 500

function formatDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function getDraftKey(date: Date): string {
  return `bujo.draft.${formatDateKey(date)}`
}

export function useEditableDocument(date: Date): EditableDocumentState {
  const [document, setDocumentState] = useState('')
  const [originalDocument, setOriginalDocument] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([])
  const [lastSaved, setLastSaved] = useState<Date | null>(null)
  const [hasDraft, setHasDraft] = useState(false)

  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const originalDocumentRef = useRef(originalDocument)
  const draftKey = getDraftKey(date)

  useEffect(() => {
    originalDocumentRef.current = originalDocument
  }, [originalDocument])

  const checkForDraft = useCallback((loadedDocument: string) => {
    const stored = localStorage.getItem(draftKey)
    if (stored) {
      try {
        const draft = JSON.parse(stored)
        if (draft.document !== loadedDocument) {
          setHasDraft(true)
        } else {
          localStorage.removeItem(draftKey)
          setHasDraft(false)
        }
      } catch {
        localStorage.removeItem(draftKey)
        setHasDraft(false)
      }
    } else {
      setHasDraft(false)
    }
  }, [draftKey])

  useEffect(() => {
    let cancelled = false

    async function loadDocument() {
      setIsLoading(true)
      setError(null)
      setValidationErrors([])

      try {
        const result = await GetEditableDocument(toWailsTime(date))
        if (!cancelled) {
          setDocumentState(result)
          setOriginalDocument(result)
          setIsLoading(false)
          checkForDraft(result)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : String(err))
          setIsLoading(false)
        }
      }
    }

    loadDocument()

    return () => {
      cancelled = true
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
        debounceTimerRef.current = null
      }
    }
  }, [date, checkForDraft])

  const saveDraft = useCallback(
    (doc: string) => {
      localStorage.setItem(draftKey, JSON.stringify({ document: doc, timestamp: Date.now() }))
    },
    [draftKey]
  )

  const clearDraft = useCallback(() => {
    localStorage.removeItem(draftKey)
    setHasDraft(false)
  }, [draftKey])

  const validateDocument = useCallback(async (doc: string) => {
    try {
      const result = await ValidateEditableDocument(doc)
      setValidationErrors(result.errors || [])
    } catch {
      // Validation failures are not critical errors
    }
  }, [])

  const setDocument = useCallback(
    (newDoc: string) => {
      setDocumentState(newDoc)

      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      debounceTimerRef.current = setTimeout(() => {
        validateDocument(newDoc)
        if (newDoc !== originalDocumentRef.current) {
          saveDraft(newDoc)
        }
      }, DEBOUNCE_MS)
    },
    [validateDocument, saveDraft]
  )

  const isDirty = document !== originalDocument

  const save = useCallback(async (): Promise<SaveResult> => {
    try {
      const validation = await ValidateEditableDocument(document)
      if (!validation.isValid) {
        setValidationErrors(validation.errors || [])
        return { success: false, error: 'Validation failed' }
      }

      const result = await ApplyEditableDocument(document, toWailsTime(date))

      setOriginalDocument(document)
      setLastSaved(new Date())
      clearDraft()

      return { success: true, result }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      return { success: false, error: errorMessage }
    }
  }, [document, date, clearDraft])

  const discardChanges = useCallback(() => {
    setDocumentState(originalDocument)
    setValidationErrors([])
  }, [originalDocument])

  const restoreDraft = useCallback(() => {
    const stored = localStorage.getItem(draftKey)
    if (stored) {
      const draft = JSON.parse(stored)
      setDocumentState(draft.document)
      setHasDraft(false)
    }
  }, [draftKey])

  const discardDraft = useCallback(() => {
    clearDraft()
  }, [clearDraft])

  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
    }
  }, [])

  return {
    document,
    setDocument,
    isLoading,
    error,
    isDirty,
    validationErrors,
    save,
    discardChanges,
    lastSaved,
    hasDraft,
    restoreDraft,
    discardDraft,
  }
}
```

**Step 2: Commit**

```bash
git add -A
git commit -m "refactor: simplify useEditableDocument hook, remove deletion tracking and debug infrastructure"
```

---

### Task 10: Simplify `EditableJournalView.tsx`

Remove deletion dialog, debug UI, mismatch display. Update to match simplified hook interface.

**Files:**
- Modify: `frontend/src/components/bujo/EditableJournalView.tsx`

**Step 1: Update the component**

Remove the `DeletionReviewDialog` import. Update the destructured hook values to match the new interface:

```typescript
const {
  document,
  setDocument,
  isLoading,
  error,
  isDirty,
  validationErrors,
  save,
  discardChanges,
  lastSaved,
  hasDraft,
  restoreDraft,
  discardDraft,
} = useEditableDocument(date)
```

Remove all references to:
- `deletedEntries`, `restoreDeletion`, `debugLog`, `debugMismatch`, `clearDebugMismatch`
- `showDeletionDialog`, `showDebug`, `clipboardLog` state
- The clipboard debug event listener (`bujo-clipboard-debug`)
- The deletion dialog rendering
- The debug panel rendering
- The `MismatchDiff` component

**Step 2: Run frontend type check**

Run: `cd frontend && npx tsc --noEmit 2>&1 | head -30`
Expected: No errors

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: simplify EditableJournalView, remove deletion dialog and debug UI"
```

---

### Task 11: Delete sort_order migration files

**Files:**
- Delete: `internal/repository/sqlite/migrations/000028_add_sort_order.up.sql`
- Delete: `internal/repository/sqlite/migrations/000028_add_sort_order.down.sql`

**Step 1: Check if sort_order column is referenced elsewhere**

Before deleting, verify no other queries depend on sort_order beyond what we've already changed. The `scanEntry` method in `entry_repository.go` scans it — update that to stop scanning it, or leave it for backward compatibility (the column can remain in the DB; we just don't use it for ordering).

**Decision:** Leave the migration files. The column exists in production databases. Removing the migration would break `OpenAndMigrate` for existing databases. Instead, we just stop using `sort_order` for ordering (already done in Task 4 by changing `ORDER BY`). Remove `UpdateSortOrders` method only (already done in Task 3).

**Step 2: Remove `SortOrder` from Entry struct usage in new code**

The `SortOrder` field on `Entry` (line 104 of entry.go) can stay — it's populated by `scanEntry` but we just don't rely on it. No changes needed.

**Step 3: No commit needed — skip this task**

---

### Task 12: Run full test suite and verify

**Step 1: Run all Go tests**

Run: `go test ./... -count=1`
Expected: PASS

**Step 2: Run frontend type check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No errors

**Step 3: Run frontend tests**

Run: `cd frontend && npx vitest run`
Expected: All remaining tests pass

**Step 4: Verify Wails bindings regeneration**

Run: `cd frontend && ls src/wailsjs/go/wails/`
Note: Wails TypeScript bindings may need regeneration. If `ApplyEditableDocument` still expects 3 args in the generated bindings, run `wails generate module` or equivalent.

**Step 5: Final commit if any fixes needed**

```bash
git add -A
git commit -m "chore: fix any remaining compilation or test issues after simplification"
```

---

## Summary

| Task | Description | Files Changed |
|------|-------------|---------------|
| 1 | Delete diff engine and migration syntax | -4 files |
| 2 | Simplify domain types | editable_document.go, editable_parser.go |
| 3 | Add DeleteByDate repository method | repository.go, entry_repository.go |
| 3b | Drop event sourcing from entries | entry_repository.go, bujo.go, export.go, app.go, tests |
| 4 | Rewrite EditableViewService | editable_view.go, editable_view_test.go, factory.go |
| 5 | Update Wails adapter | app.go |
| 6 | Delete unused frontend files | -8 files |
| 7 | Simplify bujoFolding.ts | bujoFolding.ts |
| 8 | Simplify BujoEditor.tsx | BujoEditor.tsx |
| 9 | Simplify useEditableDocument.ts | useEditableDocument.ts |
| 10 | Simplify EditableJournalView.tsx | EditableJournalView.tsx |
| 11 | Sort order cleanup (skip — leave migration) | — |
| 12 | Full verification | — |
