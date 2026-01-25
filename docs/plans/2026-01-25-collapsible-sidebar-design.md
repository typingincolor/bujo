# Collapsible Journal Sidebar Design

**Date:** 2026-01-25
**Status:** Approved
**Feature:** Make the pending tasks panel on the journal page collapsible with keyboard shortcut and mouse button

## Overview

Add collapse functionality to the journal view sidebar (pending tasks + context sections) with both keyboard (`[` key) and mouse (chevron button) controls.

## Design Decisions

### Collapse Behavior
- **Scope:** Collapse affects the entire sidebar (both pending tasks AND context sections)
- **Rationale:** These sections only make sense together - context shows hierarchy of selected pending task

### Keyboard Shortcut
- **Key:** `[` (left bracket)
- **Rationale:**
  - Single key, no modifier needed
  - Easy to reach from home row
  - Consistent with developer tools (VS Code sidebar toggle)
  - Visual mnemonic (bracket shape suggests sidebar)

### Button Design
- **Position:** Small chevron button in top-right of sidebar header (next to "Pending Tasks (X)")
- **Visibility:** Always visible
- **Icon:** ChevronLeft when expanded, ChevronRight when collapsed
- **Rationale:** Discoverability - users should immediately see collapse affordance

## State Management

**Location:** `App.tsx`
```tsx
const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
```

**Why App.tsx:**
- Keyboard shortcut handler lives there
- Sidebar conditionally rendered based on `view === 'today'`
- Simple boolean toggle, no complex logic needed

**Persistence:**
- Does NOT persist across sessions initially
- State resets to `false` (expanded) on app launch
- Can add to settings context later if desired

## Visual Design

### Expanded State
- Width: Existing `JOURNAL_SIDEBAR_WIDTH_CLASS` constant
- Content: Full pending tasks + context sections visible
- Toggle button: ChevronLeft icon in header top-right

### Collapsed State
- Width: `w-10` (40px)
- Content: All content hidden
- Toggle button: ChevronRight icon, vertically centered
- Main content: Does NOT resize (prevents layout shift)

### Transitions
- Width: `transition-all duration-300 ease-in-out`
- Content: Opacity fade
- Icon: Instant swap (no rotation animation)

### Button Styling
- Size: ~20x20px clickable area with padding
- Colors: Muted, slightly darker on hover
- Position when collapsed: Centered in narrow strip

## Component Structure

### JournalSidebar Props
```tsx
interface JournalSidebarProps {
  // ... existing props
  isCollapsed?: boolean;
  onToggleCollapse?: () => void;
}
```

### App.tsx Integration
```tsx
{view === 'today' && (
  <aside
    className={cn(
      'h-screen border-l border-border bg-background overflow-hidden transition-all duration-300 ease-in-out',
      isSidebarCollapsed ? 'w-10' : JOURNAL_SIDEBAR_WIDTH_CLASS
    )}
  >
    <JournalSidebar
      isCollapsed={isSidebarCollapsed}
      onToggleCollapse={() => setIsSidebarCollapsed(prev => !prev)}
      // ... other existing props
    />
  </aside>
)}
```

## Keyboard Integration

### Shortcut Handler
Add to existing keyboard event handler in `App.tsx`:
```tsx
if (e.key === '[') {
  e.preventDefault()
  setIsSidebarCollapsed(prev => !prev)
  return
}
```

### Scope
- Active only when `view === 'today'`
- Works regardless of panel focus (main or sidebar)
- Does NOT interfere with text input

### Help Documentation
Update `KeyboardShortcuts.tsx` in "Navigation" section:
```
[ - Toggle sidebar
```

## Edge Cases

### 1. Selection State Preservation
- Maintain `sidebarSelectedEntry` and `sidebarSelectedIndex` when collapsed
- If user expands, previously selected entry remains highlighted

### 2. Keyboard Navigation While Collapsed
- Arrow keys still work to change selection even when content hidden
- Allows keyboard-first users to navigate then expand to see selection

### 3. Tab Key Behavior
- Tab still switches focus between main/sidebar panels
- When collapsed sidebar gains focus, show subtle glow/border on collapsed strip

### 4. View Switching
- When switching away from 'today' view, reset `isSidebarCollapsed` to `false`
- Ensures sidebar is expanded when returning to journal view

## Testing Strategy

**Unit Tests:**
- JournalSidebar renders toggle button
- Button click calls onToggleCollapse callback
- Collapsed prop hides content correctly

**Integration Tests:**
- `[` keyboard shortcut toggles state
- CSS classes update on state change
- Selection state preserved across collapse/expand

**Visual Regression:**
- Collapsed state renders at 40px width
- Transition is smooth
- Button positioned correctly in both states

**Accessibility:**
- Button has proper aria-label
- Keyboard shortcut documented in help
- Focus indicator visible when collapsed sidebar has focus

## Implementation Order

1. Add state to App.tsx
2. Add keyboard shortcut handler
3. Update JournalSidebar props and UI
4. Update KeyboardShortcuts help modal
5. Write tests
6. Manual testing of edge cases
