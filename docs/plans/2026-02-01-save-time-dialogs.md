# Save-Time Dialogs: Migration & Move-to-List

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** When the user saves an editable document, detect entries marked with `>` (migrate) or `^` (move to list) and show the appropriate dialog before completing the save.

**Architecture:** The save flow in `useEditableDocument.ts` currently calls `ApplyEditableDocument` directly. We'll intercept the save to scan for `>` and `^` symbols, show dialogs, then call new backend methods that handle migration/move-to-list as part of the apply. The backend `ApplyChanges` will be extended to return entries needing special handling, and new Wails methods will process them.

**Tech Stack:** React, TypeScript, CodeMirror, Go, SQLite, Wails

---

## Design Decisions

- `>` symbol = migrate to future date. Single date picker for all `>` entries in the batch.
- `^` symbol = move to list. Single list picker for all `^` entries in the batch.
- After migration: entries stay in journal as `>` (existing behavior).
- After move-to-list: entries disappear from journal (existing behavior via `MoveEntryToList`).
- Both dialogs shown sequentially if both symbols are present (migrations first, then move-to-list).

## Current Save Flow

1. User presses Cmd+S
2. `useEditableDocument.save()` validates via `ValidateEditableDocument(doc)`
3. If valid, calls `ApplyEditableDocument(doc, date)` — deletes all entries for date, re-inserts from parsed doc
4. Reloads document

## New Save Flow

1. User presses Cmd+S
2. Frontend scans document text for lines starting with `>` or `^`
3. If `>` lines found: show MigrateBatchModal (date picker), user picks date or cancels
4. If `^` lines found: show ListPickerModal (list picker), user picks list or cancels
5. If user cancels either dialog: abort save entirely
6. Call new `ApplyEditableDocumentWithActions(doc, date, migrateDate?, listId?)` which:
   - Applies the document (delete+re-insert as before)
   - For entries marked `>`: calls MigrateEntry logic (create copy on target date)
   - For entries marked `^`: calls MoveEntryToList logic (create list item, delete entry)
7. Reload document

---

### Task 1: Add `^` symbol to Go domain parser

**Files:**
- Modify: `internal/domain/entry.go` (add `EntryTypeMovedToList`)
- Modify: `internal/domain/editable_parser.go` (add `^` to symbol maps)
- Test: `internal/domain/editable_parser_test.go`

**Step 1: Write failing test for `^` symbol parsing**

Add to `editable_parser_test.go`:

```go
func TestParseLineMovedToList(t *testing.T) {
	parser := NewEditableDocumentParser()
	result := parser.ParseLine("^ Buy groceries", 1)

	if !result.IsValid {
		t.Fatalf("expected valid, got error: %s", result.ErrorMessage)
	}
	if result.Symbol != EntryTypeMovedToList {
		t.Fatalf("expected movedToList, got %s", result.Symbol)
	}
	if result.Content != "Buy groceries" {
		t.Fatalf("expected 'Buy groceries', got %s", result.Content)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestParseLineMovedToList ./internal/domain/...`
Expected: FAIL — `EntryTypeMovedToList` undefined

**Step 3: Write minimal implementation**

In `internal/domain/entry.go`, add to EntryType constants:
```go
EntryTypeMovedToList EntryType = "movedToList"
```

Add to `validEntryTypes` map:
```go
EntryTypeMovedToList: "^",
```

In `internal/domain/editable_parser.go`, add to `editableSymbolToType`:
```go
'^': EntryTypeMovedToList,
```

Add to `typeToEditableSymbol`:
```go
EntryTypeMovedToList: '^',
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestParseLineMovedToList ./internal/domain/...`
Expected: PASS

**Step 5: Commit**

```
feat: Add movedToList entry type with ^ symbol
```

---

### Task 2: Add `^` symbol to frontend entry type system

**Files:**
- Modify: `frontend/src/types/bujo.ts` (add `movedToList` to EntryType, ENTRY_SYMBOLS)
- Modify: `frontend/src/lib/codemirror/entryTypeStyles.ts` (add `^` to symbolToType, add decoration)

**Step 1: Add `movedToList` to TypeScript types**

In `frontend/src/types/bujo.ts`:
- Add `'movedToList'` to `EntryType` union
- Add `movedToList: '^'` to `ENTRY_SYMBOLS`

**Step 2: Add `^` to CodeMirror entry styling**

In `frontend/src/lib/codemirror/entryTypeStyles.ts`:
- Add `'movedToList'` to `EntryStyleType`
- Add `'^': 'movedToList'` to `symbolToType`
- Add `movedToList: Decoration.line({ class: 'cm-entry-movedToList' })` to `lineDecorations`
- Update regex to include `^`: `^\s*([.\-ox>~?*^])\s`

**Step 3: Add CSS for moved-to-list styling**

Check existing entry type CSS file for patterns, add `cm-entry-movedToList` with a distinctive color (e.g., purple/indigo to distinguish from migration blue).

**Step 4: Commit**

```
feat: Add ^ symbol for move-to-list in frontend editor
```

---

### Task 3: Create `useSaveWithDialogs` hook — document scanning

**Files:**
- Create: `frontend/src/hooks/useSaveWithDialogs.ts`
- Test: `frontend/src/hooks/useSaveWithDialogs.test.ts`

This hook wraps the existing `save()` function and adds dialog detection logic.

**Step 1: Write failing test for document scanning**

```typescript
import { describe, it, expect } from 'vitest'
import { scanForSpecialEntries } from './useSaveWithDialogs'

describe('scanForSpecialEntries', () => {
  it('detects migrated entries', () => {
    const doc = ". Buy milk\n> Call dentist\n- A note"
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Call dentist'])
    expect(result.movedToListEntries).toEqual([])
  })

  it('detects moved-to-list entries', () => {
    const doc = ". Buy milk\n^ Fix bike\n- A note"
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual([])
    expect(result.movedToListEntries).toEqual(['Fix bike'])
  })

  it('detects both types', () => {
    const doc = "> Call dentist\n^ Fix bike"
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Call dentist'])
    expect(result.movedToListEntries).toEqual(['Fix bike'])
  })

  it('handles indented entries', () => {
    const doc = "  > Indented migrate\n  ^ Indented list"
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Indented migrate'])
    expect(result.movedToListEntries).toEqual(['Indented list'])
  })

  it('returns empty when no special entries', () => {
    const doc = ". Task\n- Note\no Event"
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual([])
    expect(result.movedToListEntries).toEqual([])
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/hooks/useSaveWithDialogs.test.ts`
Expected: FAIL — module not found

**Step 3: Write minimal implementation**

```typescript
export interface SpecialEntries {
  migratedEntries: string[]
  movedToListEntries: string[]
}

export function scanForSpecialEntries(doc: string): SpecialEntries {
  const lines = doc.split('\n')
  const migratedEntries: string[] = []
  const movedToListEntries: string[] = []

  for (const line of lines) {
    const match = line.match(/^\s*([>^])\s+(.+)/)
    if (match) {
      const [, symbol, content] = match
      if (symbol === '>') migratedEntries.push(content.trim())
      if (symbol === '^') movedToListEntries.push(content.trim())
    }
  }

  return { migratedEntries, movedToListEntries }
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/hooks/useSaveWithDialogs.test.ts`
Expected: PASS

**Step 5: Commit**

```
feat: Add scanForSpecialEntries for detecting migration and move-to-list
```

---

### Task 4: Create MigrateBatchModal component

**Files:**
- Create: `frontend/src/components/bujo/MigrateBatchModal.tsx`

Adapted from existing `MigrateModal.tsx` but shows a list of all entries being migrated and a single date picker.

**Step 1: Create the component**

```typescript
import { useState } from 'react'
import { cn } from '@/lib/utils'

interface MigrateBatchModalProps {
  isOpen: boolean
  entries: string[]
  onMigrate: (date: string) => void
  onCancel: () => void
}

export function MigrateBatchModal({ isOpen, entries, onMigrate, onCancel }: MigrateBatchModalProps) {
  const [selectedDate, setSelectedDate] = useState(() => {
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    return tomorrow.toISOString().split('T')[0]
  })

  if (!isOpen) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (selectedDate) {
      onMigrate(selectedDate)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background rounded-lg shadow-lg p-6 w-full max-w-md animate-fade-in">
        <h2 className="text-lg font-semibold mb-4">Migrate Entries</h2>
        <p className="text-sm text-muted-foreground mb-2">
          {entries.length === 1 ? 'Migrate this entry' : `Migrate ${entries.length} entries`} to a future date:
        </p>
        <ul className="mb-4 space-y-1">
          {entries.map((entry, i) => (
            <li key={i} className="text-sm text-foreground truncate pl-3 border-l-2 border-primary/50">
              {entry}
            </li>
          ))}
        </ul>
        <form onSubmit={handleSubmit}>
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
            min={new Date().toISOString().split('T')[0]}
            className={cn(
              'w-full px-3 py-2 rounded-lg border border-border bg-background',
              'focus:outline-none focus:ring-2 focus:ring-primary/50 mb-4'
            )}
            autoFocus
          />
          <div className="flex justify-end gap-2">
            <button type="button" onClick={onCancel}
              className="px-4 py-2 rounded-lg text-sm bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors">
              Cancel
            </button>
            <button type="submit"
              className="px-4 py-2 rounded-lg text-sm bg-primary text-primary-foreground hover:bg-primary/90 transition-colors">
              Migrate
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
```

**Step 2: Commit**

```
feat: Add MigrateBatchModal for batch migration on save
```

---

### Task 5: Create backend `ApplyEditableDocumentWithActions`

**Files:**
- Modify: `internal/service/editable_view.go` (add `ApplyChangesWithActions` method)
- Modify: `internal/adapter/wails/app.go` (add Wails binding)
- Test: `internal/service/editable_view_test.go`

The new method extends `ApplyChanges` to:
1. Apply the document normally (delete + re-insert)
2. For entries with type `migrated` (`>`): create copies on the migration target date (reusing `BujoService.MigrateEntry` logic)
3. For entries with type `movedToList` (`^`): create list items and delete the entries (reusing `MoveEntryToList` logic)

**Step 1: Write failing test for ApplyChangesWithActions with migration**

```go
func TestEditableViewService_ApplyChangesWithActions_Migration(t *testing.T) {
	// Setup: in-memory DB with entries repo, list repo
	// Doc: ". Keep this\n> Migrate this"
	// Call ApplyChangesWithActions(ctx, doc, today, &tomorrowDate, nil)
	// Assert: "Keep this" exists on today, "Migrate this" is migrated on today,
	//         copy of "Migrate this" exists on tomorrow as task
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestEditableViewService_ApplyChangesWithActions_Migration ./internal/service/...`
Expected: FAIL — method not defined

**Step 3: Write minimal implementation**

Add to `EditableViewService`:
```go
type ApplyActions struct {
	MigrateDate *time.Time
	ListID      *int64
}

func (s *EditableViewService) ApplyChangesWithActions(ctx context.Context, doc string, date time.Time, actions ApplyActions) (*ApplyChangesResult, error) {
	// 1. Call existing ApplyChanges
	result, err := s.ApplyChanges(ctx, doc, date)
	if err != nil {
		return nil, err
	}

	// 2. Get freshly inserted entries for this date
	entries, err := s.entryRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	// 3. For migrated entries: create copies on target date
	if actions.MigrateDate != nil {
		for _, entry := range entries {
			if entry.Type == domain.EntryTypeMigrated {
				newEntry := domain.Entry{
					Type:          domain.EntryTypeTask,
					Content:       entry.Content,
					Priority:      entry.Priority,
					ScheduledDate: actions.MigrateDate,
					CreatedAt:     time.Now(),
				}
				if _, err := s.entryRepo.Insert(ctx, newEntry); err != nil {
					return nil, err
				}
			}
		}
	}

	// 4. For movedToList entries: create list items and delete entries
	if actions.ListID != nil {
		for _, entry := range entries {
			if entry.Type == domain.EntryTypeMovedToList {
				if err := s.entryToListMover.MoveEntryToList(ctx, entry, /* listEntityID */); err != nil {
					return nil, err
				}
			}
		}
	}

	return result, nil
}
```

Note: The service will need access to the `entryToListMover` and `listRepo` — these dependencies need to be added to `EditableViewService`. The exact implementation will depend on how the list entity ID is resolved from the list row ID.

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestEditableViewService_ApplyChangesWithActions_Migration ./internal/service/...`
Expected: PASS

**Step 5: Write failing test for move-to-list action**

Similar pattern — verify entries with `^` get moved to the target list and removed from the journal.

**Step 6: Implement and verify**

**Step 7: Add Wails binding**

In `internal/adapter/wails/app.go`:
```go
func (a *App) ApplyEditableDocumentWithActions(doc string, date time.Time, migrateDate *time.Time, listID *int64) (*ApplyResult, error) {
	actions := service.ApplyActions{
		MigrateDate: migrateDate,
		ListID:      listID,
	}
	result, err := a.services.EditableView.ApplyChangesWithActions(a.ctx, doc, date, actions)
	if err != nil {
		return nil, err
	}
	return &ApplyResult{Inserted: result.Inserted, Deleted: result.Deleted}, nil
}
```

**Step 8: Regenerate Wails bindings**

Run: `wails generate`

**Step 9: Commit**

```
feat: Add ApplyEditableDocumentWithActions for migration and move-to-list on save
```

---

### Task 6: Wire up save flow with dialogs in EditableJournalView

**Files:**
- Modify: `frontend/src/components/bujo/EditableJournalView.tsx`
- Modify: `frontend/src/hooks/useEditableDocument.ts`

**Step 1: Add dialog state to EditableJournalView**

Add state for:
- `pendingMigrations: string[]` — entries to show in MigrateBatchModal
- `pendingMoveToList: string[]` — entries to show in ListPickerModal
- `pendingSaveDoc: string | null` — the document text waiting to be saved

**Step 2: Modify handleSave**

```typescript
const handleSave = useCallback(async () => {
  const { migratedEntries, movedToListEntries } = scanForSpecialEntries(document)

  if (migratedEntries.length > 0 || movedToListEntries.length > 0) {
    setPendingSaveDoc(document)
    if (migratedEntries.length > 0) {
      setPendingMigrations(migratedEntries)
      return // Show migration dialog first
    }
    if (movedToListEntries.length > 0) {
      setPendingMoveToList(movedToListEntries)
      return // Show list picker
    }
  }

  // No special entries — save directly
  const result = await save()
  if (!result.success && result.error) setSaveError(result.error)
}, [document, save])
```

**Step 3: Add dialog handlers**

```typescript
const handleMigrateBatch = useCallback(async (date: string) => {
  setMigrateDate(date)
  setPendingMigrations([])

  // Check if we also need list picker
  const { movedToListEntries } = scanForSpecialEntries(pendingSaveDoc!)
  if (movedToListEntries.length > 0) {
    setPendingMoveToList(movedToListEntries)
    return
  }

  // No more dialogs — do the save
  await completeSave(date, null)
}, [pendingSaveDoc])

const handleMoveToListSelect = useCallback(async (listId: number) => {
  setPendingMoveToList([])
  await completeSave(migrateDate, listId)
}, [migrateDate])

const completeSave = useCallback(async (migrateDate: string | null, listId: number | null) => {
  // Call ApplyEditableDocumentWithActions
  setPendingSaveDoc(null)
  setMigrateDate(null)
}, [])
```

**Step 4: Render dialogs**

Add `<MigrateBatchModal>` and `<ListPickerModal>` to JSX, controlled by `pendingMigrations` and `pendingMoveToList` state.

**Step 5: Commit**

```
feat: Wire save-time dialogs for migration and move-to-list
```

---

### Task 7: Integration testing and edge cases

**Files:**
- Test: `frontend/src/hooks/useSaveWithDialogs.test.ts` (add edge cases)
- Test: `internal/service/editable_view_test.go` (add edge cases)

**Edge cases to test:**
- User cancels migration dialog → save aborted, document unchanged
- User cancels list picker dialog → save aborted, document unchanged
- Document has both `>` and `^` entries → both dialogs shown sequentially
- Document has `>` with children → children are also migrated
- Document has `^` with children → error or children handled
- Empty document save (no special entries) → works as before
- Priority markers preserved through migration: `> !!! Important task`

**Step 1: Write and run tests for each edge case**

**Step 2: Commit**

```
test: Add edge case tests for save-time dialogs
```
