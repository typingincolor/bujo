# Context Panel UX Design

## Problem

The current `EntryContextPopover` adds friction to entry actions:
1. Click entry → opens popover → shows context + actions → click action (2 clicks minimum)
2. Context is important for decision-making but shouldn't block actions
3. SearchView's `ContextPill` takes up too much room
4. WeekSummary entries look like plain lists without type indicators

## Solution

### Core Interaction Model

**Remove popover as primary interaction:**
- Actions directly on entry row (hover reveals action bar, click symbol toggles done, right-click for menu)
- `EntryItem`'s native interactions re-enabled (currently disabled with `disableClick={true}`)

**Add toggleable context panel:**
- Press `c` to toggle context panel visibility globally
- Panel shows full hierarchy tree for currently selected entry
- Updates as selection changes (keyboard nav or click)
- Empty state message when selected entry has no ancestors

### Dot Indicator

**Appearance:**
- Small dot (4-6px) positioned to the left of entry symbol
- Muted color (`text-muted-foreground`)
- Only appears on entries with ancestors (`parentId !== null`)

**Behavior:**
- Pure visual indicator - not clickable or hoverable
- Replaces `ContextPill` component (much smaller footprint)

### Context Panel

**Trigger:**
- Keyboard shortcut `c` toggles visibility
- Optional: button in view header for discoverability

**Content:**
- Full hierarchy tree for currently selected entry
- Highlights selected entry within tree
- Empty state: "No context for this entry"

**Layout:**
- Side panel on right side of view
- Scrollable for large trees
- Persists while toggled on, updates with selection

**Implementation:**
- New `ContextPanel` component
- State lives in parent (App or view component)
- Reuses existing `EntryTree` component for rendering

### WeekSummary Entry Symbols

**Current:** Items show only content text with indicator badges

**Change:** Add `EntrySymbol` component before content:
- Meetings section: event symbol `○`
- Needs Attention section: task `.` or question `?`

**Example rendering:**
```
○ Weekly standup                    3 items
. Finish the quarterly report       ! aging
? Should we migrate to new API?     overdue
```

## Implementation Scope

### New Components
- `ContextPanel` - side panel showing hierarchy for selected entry

### Modified Components
- `EntryItem` - add dot indicator for entries with ancestors
- `DayView` - remove popover wrapper, add context panel, add `c` toggle
- `SearchView` - remove popover wrapper, remove ContextPill, remove inline expansion, add context panel, add `c` toggle
- `WeekSummary` - remove popover wrapper, add entry symbols, add context panel support

### Removed/Deprecated
- `ContextPill` - replaced by dot
- Inline ancestor expansion in SearchView (`expandedIds`, `toggleExpanded`, ancestor expansion state)
- `EntryContextPopover` as primary click wrapper (component may remain for other uses)

### Keyboard Shortcuts
- `c` - toggle context panel visibility (new)
- Existing shortcuts unchanged: `j/k` nav, `space` done, `x` cancel, `p` priority, `t` type cycle, `a` answer

## Files to Modify

1. `frontend/src/components/bujo/EntryItem.tsx` - add dot indicator
2. `frontend/src/components/bujo/DayView.tsx` - remove popover, add panel toggle
3. `frontend/src/components/bujo/SearchView.tsx` - remove popover, remove ContextPill, remove expansion, add panel toggle
4. `frontend/src/components/bujo/WeekSummary.tsx` - remove popover, add entry symbols
5. `frontend/src/components/bujo/ContextPanel.tsx` - new component
6. `frontend/src/components/bujo/ContextPill.tsx` - delete or deprecate
