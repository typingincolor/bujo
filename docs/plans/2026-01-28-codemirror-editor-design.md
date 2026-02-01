# CodeMirror Editor Integration Design

**Date:** 2026-01-28
**Status:** Ready for Implementation
**Feature:** Replace textarea with CodeMirror 6 for rich editing experience

## Overview

Integrate CodeMirror 6 into the Editable Journal View to enable visual features not possible with a plain textarea: priority badges, indent guides, error highlighting, and code folding.

## Design Decisions

### Package Choice

**CodeMirror 6** via `@uiw/react-codemirror`

- Lightweight (~130KB gzipped) vs Monaco (~2MB)
- Excellent extension system for custom syntaxes
- Active development, good React integration

### Visual Features

| Feature | Implementation |
|---------|----------------|
| Priority badges | `Decoration.replace()` swaps `!`, `!!`, `!!!` with styled 1/2/3 badges |
| Indent guides | `Decoration.line()` adds CSS classes, `::before` pseudo-elements draw vertical lines |
| Error highlighting | CodeMirror linting API provides red underlines + gutter icons |
| Code folding | Custom fold service based on indentation depth |

### Priority Badge Display

User types exclamation marks, editor displays colored badges:

```
. !!! Urgent task  →  . [1] Urgent task
. !! High priority →  . [2] High priority
. ! Low priority   →  . [3] Low priority
```

Badges use existing CSS variables: `--priority-high`, `--priority-medium`, `--priority-low`

### Indent Guides

Faint vertical lines at each indentation level (2 spaces = 1 level):

```
. Parent task
│ . Child task
│ │ . Grandchild
│ - Sibling note
```

### Error Display

- Red wavy underline under invalid content
- Red dot icon in left gutter
- Integrates with existing `editableParser` validation

### Code Folding

Entries with children can be collapsed:

```
. Parent task          ▼  (expanded)
  . Child 1
  . Child 2

. Parent task          ▶  (collapsed)
```

Fold markers appear in gutter for entries with nested children.

## Architecture

### File Structure

```
src/
  components/bujo/
    EditableJournalView.tsx      # Uses BujoEditor instead of textarea
    BujoEditor.tsx               # CodeMirror wrapper component
  lib/
    codemirror/
      bujoLanguage.ts            # Custom language definition
      bujoTheme.ts               # Theme using app CSS variables
      priorityBadges.ts          # Decoration plugin for badges
      indentGuides.ts            # Decoration plugin for vertical lines
      bujoFolding.ts             # Custom fold service
      errorMarkers.ts            # Linting integration
      index.ts                   # Combined extension export
```

### BujoEditor Component

```typescript
interface BujoEditorProps {
  value: string
  onChange: (value: string) => void
  onSave?: () => void
  onImport?: () => void
  validationErrors?: ValidationError[]
  disabled?: boolean
}
```

Key configuration:
- No line numbers (journal feel, not code)
- Fold gutter enabled for collapsing entries
- Tab/Shift+Tab for indent/outdent (2 spaces)
- Cmd+S triggers save, Cmd+I triggers import

### Theme Integration

CodeMirror theme references app CSS variables for automatic light/dark mode support:

```typescript
const bujoTheme = EditorView.theme({
  '&': {
    backgroundColor: 'hsl(var(--background))',
    color: 'hsl(var(--foreground))',
  },
  '.bujo-task': { color: 'hsl(var(--bujo-task))' },
  '.bujo-done': { color: 'hsl(var(--bujo-done))' },
  '.bujo-event': { color: 'hsl(var(--bujo-event))' },
  // ... etc
})
```

## Implementation Tasks

| Task | File | Description |
|------|------|-------------|
| 4.4.1 | `package.json` | Install `@uiw/react-codemirror` and CodeMirror dependencies |
| 4.4.2 | `bujoTheme.ts` | Create theme using CSS variables |
| 4.4.3 | `priorityBadges.ts` | Decoration plugin for 1/2/3 badges |
| 4.4.4 | `indentGuides.ts` | Decoration plugin for vertical lines |
| 4.4.5 | `bujoFolding.ts` | Custom fold service for entry hierarchy |
| 4.4.6 | `errorMarkers.ts` | Linting integration with editableParser |
| 4.4.7 | `BujoEditor.tsx` | Wrapper component with all extensions |
| 4.4.8 | `EditableJournalView.tsx` | Replace textarea with BujoEditor |

## Testing Strategy

### Unit Tests (Pure Functions)

```typescript
// priorityBadges.test.ts
- findPriorityMarkers: finds markers after symbols
- findPriorityMarkers: ignores ! in content

// indentGuides.test.ts
- getIndentDepth: calculates depth from leading spaces

// bujoFolding.test.ts
- getFoldRange: returns null for childless entries
- getFoldRange: returns range covering nested children
```

### Component Tests

```typescript
// BujoEditor.test.tsx
- renders document content
- calls onChange when text edited
- calls onSave on Cmd+S
- displays priority badges
- shows error underline for invalid lines
```

### What We Don't Test

CodeMirror internals - trust the library. Test our extension logic and integration points.

## Success Criteria

1. Priority markers display as colored 1/2/3 badges
2. Indent guides show hierarchy visually
3. Invalid lines have red underline + gutter icon
4. Parent entries can be folded/unfolded
5. Theme matches app light/dark mode
6. All existing keyboard shortcuts work (Cmd+S, Cmd+I, Tab, Shift+Tab)
7. Existing tests continue to pass
