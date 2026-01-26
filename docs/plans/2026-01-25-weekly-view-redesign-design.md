# Weekly View Redesign

## Goal

Replace the current WeekSummary component with a calendar-grid weekly view that shows events and priority entries in a 2×3 day-box layout with context panel integration.

## Architecture

### Component Hierarchy

```
WeekView (orchestrator)
├── WeekCalendarGrid (2×3 grid layout)
│   ├── DayBox × 5 (Mon-Fri individual boxes)
│   └── WeekendBox (Sat-Sun combined)
│       └── WeekEntry (clickable entry items)
└── Context Panel (reused from JournalView pattern)
```

### Key Differences from Current WeekSummary

- **Calendar grid layout** instead of attention-score summary lists
- **Filtered entries** showing events + priority items only (not all entries)
- **Weekend consolidation** combining Sat-Sun into single box with inline date labels
- **Context panel integration** showing ancestry tree for selected entry
- **Entry actions** on hover/selection (same pattern as JournalView)

### Component Reuse

- `ContextTree` from JournalSidebar (without pending tasks section)
- `EntryActionBar` for entry actions
- `EntrySymbol` for type/priority indicators
- Same callback pattern as JournalSidebar: `onMarkDone`, `onMigrate`, `onEdit`, etc.

## Data Flow

### Input Data

The WeekView receives `DayEntries[]` for 7 days in the current week from parent component (App.tsx).

Parent is responsible for:
- Week date range calculation
- Loading entries for the week
- Handling prev/next week navigation

### Entry Filtering Logic

For each day, flatten hierarchical entries and filter to show:

```typescript
const visibleEntries = allEntries.filter(entry =>
  entry.type === 'event' ||
  (entry.priority === 'low' || entry.priority === 'medium' || entry.priority === 'high')
);
```

**Show:**
- All events (○) regardless of priority
- Any entry type with priority set (!, !!, !!!)
  - Tasks (•)
  - Notes (-)
  - Questions (?)
  - Done (x)
  - Migrated (>)

**Hide:**
- Entries without priority (unless they're events)

### Grouping

- **Mon-Fri:** Group entries by single date
- **Weekend box:** Combine Sat+Sun entries, prefix each with "Sat:" or "Sun:" inline label

### Selection State

```
User clicks entry in calendar
  ↓
Set selectedEntry state
  ↓
Context panel updates with ancestry tree
  ↓
Hover shows action bar below entry
```

## UI Layout

### Calendar Grid Structure

```
┌─────────────┬─────────────┬─────────────┐
│ 19 Mon      │ 20 Tue      │ 21 Wed      │
│             │             │             │
│ [entries]   │ [entries]   │ [entries]   │
│             │             │             │
├─────────────┼─────────────┼─────────────┤
│ 22 Thu      │ 23 Fri      │ 24-25       │
│             │             │ Weekend     │
│ [entries]   │ [entries]   │ [entries]   │
│             │             │             │
└─────────────┴─────────────┴─────────────┘
```

Grid uses CSS Grid: `grid-cols-3` with equal column widths.

### Day Box Components

**Header:**
- Date number (larger) + day name (smaller)
- Example: "19" (text-2xl) "Mon" (text-sm text-muted-foreground)

**Empty state:**
- "No events" in muted text when no filtered entries

**Entry list:**
- Scrollable area with `max-h-64` or similar
- Vertical spacing between entries
- Each entry is clickable button

**Entry display:**
- Symbol (○, •, ?, etc.)
- Priority indicator if present (!, !!, !!!)
- Content (truncated with ellipsis if too long)
- Example: `• !!! Draft project timeline - URGENT`

### Weekend Box Specifics

**Header:**
- "24-25 Weekend" format

**Entry prefix:**
- Inline date label before symbol
- Example: `Sat: ○ Lunch with Sarah`
- Example: `Sun: • !!! Fix login bug - blocker`

**Mixing:**
- Entries from both days in single list
- Sorted by date (Sat first, then Sun)

### Visual Styling

**Box styles:**
- Rounded border: `rounded-lg`
- Border: `border border-border`
- Background: `bg-card`
- Padding: `p-4`

**Entry selection:**
- Selected: `bg-primary/10 ring-1 ring-primary/30`
- Hover: `bg-secondary/50`
- Transition: `transition-colors`

**Priority indicators:**
- !!! (high): Red text `text-red-500`
- !! (medium): Orange text `text-orange-500`
- ! (low): Yellow text `text-yellow-500`

**Responsive:**
- Grid stays 2×3 on all screen sizes
- Individual boxes scroll internally if needed

## Context Panel Integration

### Panel Structure

Right sidebar shows context tree without pending tasks section:

```
┌─────────────────┐
│ Context         │
│─────────────────│
│                 │
│ [Context Tree]  │
│                 │
│ (scrollable)    │
│                 │
└─────────────────┘
```

### Display States

1. **No selection:** "No entry selected" placeholder
2. **Root entry selected:** "No context" placeholder
3. **Child entry selected:** Full ancestry tree with selected entry highlighted

### Context Tree Display

Reuse `ContextTree` component from JournalSidebar:

- Indented hierarchy showing ancestors
- Entry symbols inline
- Selected entry in normal weight/color
- Ancestor entries in muted color
- Truncation for long content

### Entry Actions

**Trigger:**
When entry is clicked in calendar grid → becomes selected

**Action bar display:**
- Hover over entry in calendar → action bar slides down below entry
- Same slide-down animation as JournalSidebar: `grid-rows-[0fr]` → `grid-rows-[1fr]`
- Action bar shows below entry content within day box

**Actions available:**
- Mark Done (cancel icon for tasks)
- Migrate (forward arrow)
- Edit (pencil)
- Delete (trash)
- Cycle Priority (! → !! → !!! → no priority)
- Move to List (checklist icon)

**Implementation:**
- Reuse `EntryActionBar` component
- Props: `variant="always-visible"` and `size="sm"`
- Callbacks: `onCancel`, `onMigrate`, `onEdit`, `onDelete`, `onCyclePriority`, `onMoveToList`

### Callback Pattern

```typescript
interface WeekViewCallbacks {
  onMarkDone?: (entry: Entry) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
  onCyclePriority?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
}
```

Parent component provides callbacks, WeekView passes to EntryActionBar.

## Testing Strategy

### Component Tests

1. **WeekView orchestration:**
   - Renders with empty week
   - Renders with entries
   - Filters to events + priority only
   - Groups entries by day correctly
   - Combines weekend entries

2. **DayBox component:**
   - Shows header with date/day
   - Shows "No events" when empty
   - Lists filtered entries
   - Handles entry selection
   - Shows action bar on hover

3. **WeekendBox component:**
   - Shows combined header
   - Prefixes entries with "Sat:" or "Sun:"
   - Sorts entries by date

4. **Entry interactions:**
   - Click selects entry
   - Hover shows actions
   - Action callbacks fire correctly

5. **Context panel:**
   - Shows "No entry selected" initially
   - Shows "No context" for root entries
   - Shows ancestry tree for child entries
   - Highlights selected entry

### Integration Tests

1. Full week navigation flow
2. Entry selection → context display
3. Entry actions → data updates
4. Week change → data refresh

## Migration from WeekSummary

### Breaking Changes

WeekSummary is replaced entirely - different purpose and interface.

### Transition Plan

1. Create new WeekView component alongside WeekSummary
2. Add new route/view option in App.tsx
3. Test thoroughly
4. Replace WeekSummary in navigation
5. Remove old WeekSummary component

### Route/Navigation

Update App.tsx view routing:
- Current: "week-summary" route shows WeekSummary
- New: "week" route shows WeekView
- Navigation sidebar: Change "Weekly Review" to point to new WeekView

## Tech Stack

- React 18 with TypeScript
- Tailwind CSS for styling
- Existing BuJo types and utilities
- Reused components: EntryActionBar, EntrySymbol, ContextTree

## Open Questions

None - design validated with user.
