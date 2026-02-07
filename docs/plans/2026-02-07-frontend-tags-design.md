# Frontend Tag Display, Autocomplete & Management

**Issue:** #458
**Date:** 2026-02-07
**Depends on:** PR #459 (backend tag infrastructure - merged)

## Design Decisions

| Decision | Choice |
|----------|--------|
| Scope | All four capabilities in one PR |
| Tag display | Inline in content - style `#tag` text with subtle color, no separate chips |
| Tag filtering | Click tag navigates to SearchView with tag pre-filled |
| Autocomplete trigger | After `#` + 1 character |
| Tag management UI | Extend SearchView with tag browser panel |

## Architecture

Five implementation layers, built bottom-up:

1. **Data plumbing** - Wire tag data from backend to frontend types
2. **TagContent component** - Render inline styled tags in entry content
3. **SearchView tag integration** - Tag filtering and tag browser panel
4. **Editor autocomplete** - CodeMirror extension for `#tag` suggestions
5. **Tag management** - Rename/delete from SearchView tag panel

## Layer 1: Data Plumbing

### `frontend/src/wailsjs/go/models.ts`
Add `Tags: string[]` to `domain.Entry` class. Manual addition with comment noting next Wails regeneration will pick it up.

### `frontend/src/types/bujo.ts`
Add `tags?: string[]` to `Entry` interface. Optional because older entries may not have tags.

### `frontend/src/lib/transforms.ts`
Map `e.Tags` to `tags` in `transformEntry()`:
```typescript
tags: e.Tags || [],
```

### SearchView `SearchResult` type
Add `tags?: string[]` and map from backend results.

## Layer 2: TagContent Component

Pure rendering component: `frontend/src/components/bujo/TagContent.tsx`

Parses entry content string and renders `#tag` portions as styled clickable spans.

```
Input:  "Buy groceries #shopping #errands"
Output: ["Buy groceries ", <span class="tag">#shopping</span>, " ", <span class="tag">#errands</span>]
```

- Regex: `#([a-zA-Z][a-zA-Z0-9-]*)` (matches backend)
- Tag spans: subtle color differentiation, rounded background on hover, pointer cursor
- Respects entry type styling (done = done color, cancelled = line-through)
- `onTagClick(tagName)` callback prop for navigation
- Used in both `EntryItem.tsx` (line 204) and `SearchView.tsx` (line 338)

## Layer 3: SearchView Tag Extension

### Tag filtering mode
- New prop: `initialTagFilter?: string`
- When present, calls `SearchByTags([tag])` instead of `Search(query)`
- Shows "Filtered by #tag" indicator with clear button
- Search input still visible for further refinement

### Tag browser panel
- Collapsible section below search input
- Fetches all tags via `GetAllTags()`
- Horizontal flex-wrap list of small pill elements
- Each shows tag name + entry count
- Click applies as filter
- "..." menu on each pill for rename/delete

### Navigation flow
Click tag in EntryItem -> App navigates to SearchView with `initialTagFilter="tagname"` -> SearchView calls `SearchByTags(["tagname"])` -> results displayed

## Layer 4: Editor Autocomplete

Custom CodeMirror `CompletionSource` in `BujoEditor.tsx`:

1. Watches for `#` followed by 1+ word characters
2. Fetches suggestions from cached `GetAllTags()` call
3. Shows completion dropdown with matching tags
4. Replaces partial `#xxx` with `#selected-tag` on selection
5. Escape dismisses

## Layer 5: Tag Management (Deferred)

Backend `RenameTag` and `DeleteTag` endpoints do not exist yet. Tag management (rename/delete) is deferred to a follow-up issue. The tag browser panel will show all tags with counts but without rename/delete actions for now.

## New Files

- `frontend/src/components/bujo/TagContent.tsx` - Inline tag rendering
- `frontend/src/components/bujo/TagContent.test.tsx` - Tests
- `frontend/src/components/bujo/TagPanel.tsx` - Tag browser/management panel
- `frontend/src/components/bujo/TagPanel.test.tsx` - Tests
- `frontend/src/lib/codemirror/tagAutocomplete.ts` - CodeMirror extension
- `frontend/src/lib/codemirror/tagAutocomplete.test.ts` - Tests

## Modified Files

- `frontend/src/wailsjs/go/models.ts` - Add Tags field
- `frontend/src/types/bujo.ts` - Add tags to Entry
- `frontend/src/lib/transforms.ts` - Map tags in transform
- `frontend/src/lib/transforms.test.ts` - Test tag mapping
- `frontend/src/components/bujo/EntryItem.tsx` - Use TagContent
- `frontend/src/components/bujo/SearchView.tsx` - Tag filtering + TagPanel
- `frontend/src/components/bujo/SearchView.test.tsx` - Tests
- `frontend/src/lib/codemirror/BujoEditor.tsx` - Add autocomplete extension
- `frontend/src/App.tsx` - Pass initialTagFilter to SearchView
