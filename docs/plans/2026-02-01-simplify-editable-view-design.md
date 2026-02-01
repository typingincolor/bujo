# Simplify Editable Journal View

## Problem

The editable journal view grew to ~5,500 lines across 22 files. Complexity stems from maintaining entity identity through freeform text editing: entity IDs embedded in document text, a 5-way diff engine, clipboard handlers to strip/preserve IDs, a sort_order column, and mismatch detection infrastructure.

## Decision

Replace the parse-diff-apply pipeline with replace-all-on-save. On each save, delete all entries for the date and re-insert from parsed text in document order.

## Design

### Save flow

1. Parse document text into entries (symbol, content, priority, depth from indentation)
2. Validate: check syntax, hierarchy (no orphan children)
3. Delete all current entries for the date (event-sourced DELETE)
4. Re-insert each entry in document order, resolving parent-child via depth stack

Insertion order = display order. No sort_order column needed. `GetByDate` orders by `created_at`.

### Document format

Plain bullet journal syntax. No entity IDs in text.

```
. My task
  - A child note
o Meeting at 3pm
! . High priority task
```

### What we remove

- `document_diff.go` - diff engine with 5 operation types
- `migration_syntax.go` - migration parsing
- `entityIdHider.ts` - CodeMirror extension hiding UUIDs
- `editableParser.ts` + test - frontend-side parser
- `DeletionReviewDialog.tsx` - deletion review modal
- `sort_order` column + migration 000028
- `UpdateSortOrders` repository method
- Mismatch detection modal + clipboard debug UI
- Clipboard copy/cut/paste handlers in `bujoFolding.ts`
- Entity ID embedding in serialization
- DiffOperation/Changeset types from `editable_document.go`

### What we keep

- `editable_parser.go` - simplified, no entity IDs
- `editable_view.go` - rewritten, smaller
- `BujoEditor.tsx` - stripped to basic keybindings
- `bujoFolding.ts` - fold range calculation only
- `bujoTheme.ts`, `indentGuides.ts`, `errorMarkers.ts`, `priorityBadges.ts`
- `useEditableDocument.ts` - simplified (~100 lines)
- Draft recovery (localStorage)

### New code

- `DeleteByDate(ctx, date)` repository method - event-sourced batch delete
- Simplified `ParseDocument` domain function

### Trade-offs

- Every save replaces all entries. Last save wins. Acceptable for single-user.
- Entries lose entity_id continuity across saves. Old versions preserved as DELETE records.
- Migration syntax removed. Can add back later if needed.
