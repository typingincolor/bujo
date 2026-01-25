# Context Panel UX Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace EntryContextPopover with direct entry actions and a toggleable side context panel.

**Architecture:** Remove popover wrappers from DayView/SearchView/WeekSummary, re-enable EntryItem's native actions, add ContextPanel component that shows hierarchy for selected entry, triggered by `c` key toggle.

**Tech Stack:** React, TypeScript, Radix UI, Tailwind CSS, Vitest/React Testing Library

---

## Task 1: Create ContextPanel Component

**Files:**
- Create: `frontend/src/components/bujo/ContextPanel.tsx`
- Create: `frontend/src/components/bujo/ContextPanel.test.tsx`

**Step 1: Write the failing test**

Create `frontend/src/components/bujo/ContextPanel.test.tsx`:

```tsx
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ContextPanel } from './ContextPanel'
import { Entry } from '@/types/bujo'

const mockEntries: Entry[] = [
  { id: 1, content: 'Root entry', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
  { id: 2, content: 'Child entry', type: 'note', priority: 'none', parentId: 1, loggedDate: '2026-01-25' },
  { id: 3, content: 'Grandchild entry', type: 'task', priority: 'high', parentId: 2, loggedDate: '2026-01-25' },
]

describe('ContextPanel', () => {
  it('renders empty state when selectedEntry has no ancestors', () => {
    const rootEntry = mockEntries[0]
    render(<ContextPanel selectedEntry={rootEntry} entries={mockEntries} />)
    expect(screen.getByText('No context for this entry')).toBeInTheDocument()
  })

  it('renders hierarchy tree when selectedEntry has ancestors', () => {
    const grandchild = mockEntries[2]
    render(<ContextPanel selectedEntry={grandchild} entries={mockEntries} />)
    expect(screen.getByText('Root entry')).toBeInTheDocument()
    expect(screen.getByText('Child entry')).toBeInTheDocument()
    expect(screen.getByText('Grandchild entry')).toBeInTheDocument()
  })

  it('highlights the selected entry in the tree', () => {
    const grandchild = mockEntries[2]
    render(<ContextPanel selectedEntry={grandchild} entries={mockEntries} />)
    const highlighted = screen.getByTestId(`context-panel-item-${grandchild.id}`)
    expect(highlighted).toHaveAttribute('data-highlighted', 'true')
  })

  it('renders nothing when selectedEntry is null', () => {
    const { container } = render(<ContextPanel selectedEntry={null} entries={mockEntries} />)
    expect(container.firstChild).toBeNull()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run ContextPanel.test.tsx`
Expected: FAIL with "Cannot find module './ContextPanel'"

**Step 3: Write minimal implementation**

Create `frontend/src/components/bujo/ContextPanel.tsx`:

```tsx
import { useMemo } from 'react'
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo'
import { cn } from '@/lib/utils'

interface ContextPanelProps {
  selectedEntry: Entry | null
  entries: Entry[]
}

export function ContextPanel({ selectedEntry, entries }: ContextPanelProps) {
  if (!selectedEntry) return null

  const entriesById = useMemo(() => new Map(entries.map(e => [e.id, e])), [entries])

  const ancestorPath = useMemo(() => {
    const path: Entry[] = []
    let current: Entry | undefined = selectedEntry
    while (current) {
      path.unshift(current)
      current = current.parentId ? entriesById.get(current.parentId) : undefined
    }
    return path
  }, [selectedEntry, entriesById])

  const hasAncestors = ancestorPath.length > 1

  if (!hasAncestors) {
    return (
      <div className="p-4 border-l border-border bg-card h-full">
        <p className="text-sm text-muted-foreground">No context for this entry</p>
      </div>
    )
  }

  return (
    <div className="p-4 border-l border-border bg-card h-full overflow-auto">
      <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
        Context
      </h3>
      <div className="space-y-1">
        {ancestorPath.map((entry, index) => {
          const isHighlighted = entry.id === selectedEntry.id
          const symbol = ENTRY_SYMBOLS[entry.type] || '-'

          return (
            <div
              key={entry.id}
              data-testid={`context-panel-item-${entry.id}`}
              data-highlighted={isHighlighted ? 'true' : undefined}
              style={{ paddingLeft: `${index * 16}px` }}
              className={cn(
                'py-1 px-2 rounded text-sm',
                isHighlighted ? 'bg-primary/10 font-medium' : 'text-muted-foreground'
              )}
            >
              <span className="font-mono mr-2">{symbol}</span>
              {entry.content}
            </div>
          )
        })}
      </div>
    </div>
  )
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run ContextPanel.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/ContextPanel.tsx frontend/src/components/bujo/ContextPanel.test.tsx
git commit -m "feat: add ContextPanel component for entry hierarchy display"
```

---

## Task 2: Add Context Dot Indicator to EntryItem

**Files:**
- Modify: `frontend/src/components/bujo/EntryItem.tsx`
- Create: `frontend/src/components/bujo/EntryItem.contextDot.test.tsx`

**Step 1: Write the failing test**

Create `frontend/src/components/bujo/EntryItem.contextDot.test.tsx`:

```tsx
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { EntryItem } from './EntryItem'
import { Entry } from '@/types/bujo'

describe('EntryItem context dot', () => {
  const baseEntry: Entry = {
    id: 1,
    content: 'Test entry',
    type: 'task',
    priority: 'none',
    parentId: null,
    loggedDate: '2026-01-25',
  }

  it('shows context dot when entry has parent', () => {
    const entryWithParent = { ...baseEntry, parentId: 99 }
    render(<EntryItem entry={entryWithParent} />)
    expect(screen.getByTestId('context-dot')).toBeInTheDocument()
  })

  it('does not show context dot when entry has no parent', () => {
    render(<EntryItem entry={baseEntry} />)
    expect(screen.queryByTestId('context-dot')).not.toBeInTheDocument()
  })

  it('context dot has muted styling', () => {
    const entryWithParent = { ...baseEntry, parentId: 99 }
    render(<EntryItem entry={entryWithParent} />)
    const dot = screen.getByTestId('context-dot')
    expect(dot).toHaveClass('bg-muted-foreground')
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run EntryItem.contextDot.test.tsx`
Expected: FAIL with "Unable to find an element by: [data-testid="context-dot"]"

**Step 3: Write minimal implementation**

Modify `frontend/src/components/bujo/EntryItem.tsx`. Add the context dot after the collapse indicator and before the symbol:

Find this section (around line 127-144):
```tsx
      {/* Collapse indicator */}
      {hasChildren ? (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onToggleCollapse?.();
          }}
          className="w-4 h-4 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
        >
          {isCollapsed ? (
            <ChevronRight className="w-3.5 h-3.5" />
          ) : (
            <ChevronDown className="w-3.5 h-3.5" />
          )}
        </button>
      ) : (
        <span className="w-4" />
      )}
```

Add the context dot right after the collapse indicator block:

```tsx
      {/* Context dot - indicates entry has ancestors */}
      {entry.parentId !== null ? (
        <span
          data-testid="context-dot"
          className="w-1.5 h-1.5 rounded-full bg-muted-foreground flex-shrink-0"
          title="Has parent context"
        />
      ) : (
        <span className="w-1.5" />
      )}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run EntryItem.contextDot.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/EntryItem.tsx frontend/src/components/bujo/EntryItem.contextDot.test.tsx
git commit -m "feat: add context dot indicator to EntryItem for entries with ancestors"
```

---

## Task 3: Remove EntryContextPopover from DayView

**Files:**
- Modify: `frontend/src/components/bujo/DayView.tsx`
- Modify: `frontend/src/components/bujo/DayView.test.tsx`

**Step 1: Write/update test for direct actions**

Add to `frontend/src/components/bujo/DayView.test.tsx` (or update existing):

```tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DayView } from './DayView'
import { DayEntries } from '@/types/bujo'

describe('DayView direct actions', () => {
  const mockDay: DayEntries = {
    date: '2026-01-25',
    entries: [
      { id: 1, content: 'Task entry', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
    ],
  }

  it('allows clicking entry symbol to toggle done without popover', async () => {
    const user = userEvent.setup()
    const onEntryChanged = vi.fn()
    render(<DayView day={mockDay} onEntryChanged={onEntryChanged} />)

    // Entry should be directly clickable, not wrapped in popover
    const entryItem = screen.getByTestId('entry-item')
    expect(entryItem).toBeInTheDocument()

    // No popover should exist
    expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
  })

  it('shows action bar on hover', async () => {
    const user = userEvent.setup()
    render(<DayView day={mockDay} />)

    const entryItem = screen.getByTestId('entry-item')
    await user.hover(entryItem)

    // Action bar should be visible on hover (not hidden behind popover)
    // The action bar contains buttons that are visible on group-hover
    expect(entryItem).toHaveClass('group')
  })
})
```

**Step 2: Run test to verify current state**

Run: `cd frontend && npm test -- --run DayView.test.tsx`
Note: Test may pass or fail depending on current implementation

**Step 3: Remove popover wrapper from DayView**

In `frontend/src/components/bujo/DayView.tsx`:

1. Remove the import:
```tsx
// Remove this line:
import { EntryContextPopover } from './EntryContextPopover';
```

2. In the `EntryTree` component, remove the conditional popover wrapping. Replace the entire conditional block (lines ~82-141) with just the `EntryItem` without wrapper:

```tsx
function EntryTree({ entries, allEntries, depth = 0, collapsedIds, selectedEntryId, onToggleCollapse, onToggleDone, onSelect, onEdit, onDelete, onCancel, onUncancel, onCyclePriority, onMigrate, onCycleType, onAddChild, onMoveToRoot, onMoveToList, onAnswer }: EntryTreeProps) {
  return (
    <>
      {entries.map((entry) => {
        const hasChildren = entry.children && entry.children.length > 0;
        const isCollapsed = collapsedIds.has(entry.id);

        return (
          <div key={entry.id}>
            <EntryItem
              entry={entry}
              depth={depth}
              isCollapsed={isCollapsed}
              hasChildren={hasChildren}
              hasParent={entry.parentId !== null}
              childCount={entry.children?.length || 0}
              isSelected={entry.id === selectedEntryId}
              onToggleCollapse={() => onToggleCollapse(entry.id)}
              onToggleDone={() => onToggleDone(entry.id)}
              onSelect={onSelect ? () => onSelect(entry.id) : undefined}
              onEdit={onEdit ? () => onEdit(entry) : undefined}
              onDelete={onDelete ? () => onDelete(entry) : undefined}
              onCancel={onCancel ? () => onCancel(entry) : undefined}
              onUncancel={onUncancel ? () => onUncancel(entry) : undefined}
              onCyclePriority={onCyclePriority ? () => onCyclePriority(entry) : undefined}
              onMigrate={onMigrate ? () => onMigrate(entry) : undefined}
              onCycleType={onCycleType ? () => onCycleType(entry) : undefined}
              onAddChild={onAddChild && entry.type !== 'question' ? () => onAddChild(entry) : undefined}
              onMoveToRoot={onMoveToRoot ? () => onMoveToRoot(entry) : undefined}
              onMoveToList={onMoveToList ? () => onMoveToList(entry) : undefined}
              onAnswer={onAnswer && entry.type === 'question' ? () => onAnswer(entry) : undefined}
            />
            {hasChildren && !isCollapsed && (
              <EntryTree
                entries={entry.children!}
                allEntries={allEntries}
                depth={depth + 1}
                collapsedIds={collapsedIds}
                selectedEntryId={selectedEntryId}
                onToggleCollapse={onToggleCollapse}
                onToggleDone={onToggleDone}
                onSelect={onSelect}
                onEdit={onEdit}
                onDelete={onDelete}
                onCancel={onCancel}
                onUncancel={onUncancel}
                onCyclePriority={onCyclePriority}
                onMigrate={onMigrate}
                onCycleType={onCycleType}
                onAddChild={onAddChild}
                onMoveToRoot={onMoveToRoot}
                onMoveToList={onMoveToList}
                onAnswer={onAnswer}
              />
            )}
          </div>
        );
      })}
    </>
  );
}
```

3. Remove unused props from `EntryTreeProps` interface:
   - Remove `onAction`
   - Remove `onNavigate`

4. Remove `handleAction` and `handleNavigate` functions from main `DayView` component

5. Remove `onAction` and `onNavigate` from the `EntryTree` call in the render

**Step 4: Run tests to verify they pass**

Run: `cd frontend && npm test -- --run DayView`
Expected: PASS (or update tests that relied on popover)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/DayView.tsx frontend/src/components/bujo/DayView.test.tsx
git commit -m "refactor: remove EntryContextPopover from DayView, enable direct actions"
```

---

## Task 4: Add Entry Symbols to WeekSummary

**Files:**
- Modify: `frontend/src/components/bujo/WeekSummary.tsx`
- Modify: `frontend/src/components/bujo/WeekSummary.test.tsx`

**Step 1: Write the failing test**

Add to `frontend/src/components/bujo/WeekSummary.test.tsx`:

```tsx
describe('WeekSummary entry symbols', () => {
  it('shows event symbol in meetings section', () => {
    const days: DayEntries[] = [{
      date: '2026-01-25',
      entries: [
        { id: 1, content: 'Team standup', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
        { id: 2, content: 'Action item', type: 'task', priority: 'none', parentId: 1, loggedDate: '2026-01-25' },
      ],
    }]
    render(<WeekSummary days={days} />)

    const meetingSection = screen.getByTestId('week-summary-meetings')
    expect(meetingSection).toHaveTextContent('○') // event symbol
  })

  it('shows task symbol in attention section for tasks', () => {
    const days: DayEntries[] = [{
      date: '2026-01-20', // 5 days ago to trigger attention
      entries: [
        { id: 1, content: 'Old task', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-20' },
      ],
    }]
    render(<WeekSummary days={days} />)

    const attentionSection = screen.getByTestId('week-summary-attention')
    expect(attentionSection).toHaveTextContent('•') // task symbol
  })

  it('shows question symbol in attention section for questions', () => {
    const days: DayEntries[] = [{
      date: '2026-01-20',
      entries: [
        { id: 1, content: 'Pending question', type: 'question', priority: 'none', parentId: null, loggedDate: '2026-01-20' },
      ],
    }]
    render(<WeekSummary days={days} />)

    const attentionSection = screen.getByTestId('week-summary-attention')
    expect(attentionSection).toHaveTextContent('?') // question symbol
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run WeekSummary.test.tsx`
Expected: FAIL (symbols not present in current implementation)

**Step 3: Add entry symbols to WeekSummary**

Modify `frontend/src/components/bujo/WeekSummary.tsx`:

1. Add import:
```tsx
import { EntrySymbol } from './EntrySymbol';
```

2. Update meetings section (around line 85-95) to include symbol:

```tsx
const meetingItem = (
  <button
    type="button"
    className="w-full flex items-center justify-between p-2 rounded-lg border border-border bg-card hover:bg-muted/50 cursor-pointer text-left"
  >
    <span className="flex items-center gap-2">
      <EntrySymbol type={event.type} />
      <span className="text-sm">{event.content}</span>
    </span>
    <span className="text-xs text-muted-foreground">
      {getChildCount(event.id)} items
    </span>
  </button>
);
```

3. Update attention section (around line 138-148) to include symbol:

```tsx
const attentionItem = (
  <button
    type="button"
    data-testid="attention-item"
    data-attention-item
    data-priority={isHighPriority ? 'high' : undefined}
    className="w-full flex items-center justify-between p-2 rounded-lg border border-border bg-card hover:bg-muted/50 cursor-pointer text-left"
  >
    <span className="flex items-center gap-2">
      <EntrySymbol type={entry.type} priority={entry.priority} />
      <span className="text-sm">{entry.content}</span>
    </span>
    {indicators.length > 0 && (
      <div className="flex gap-1" data-testid="attention-indicators">
        {indicators.map(indicator => (
          <span
            key={indicator}
            data-indicator={indicator}
            title={indicator}
            className={cn(
              'text-xs px-1.5 py-0.5 rounded',
              indicator === 'priority' && 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
              indicator === 'overdue' && 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
              indicator === 'aging' && 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
              indicator === 'migrated' && 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
            )}
          >
            {indicator === 'priority' ? '!' : indicator}
          </span>
        ))}
      </div>
    )}
  </button>
);
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run WeekSummary.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekSummary.tsx frontend/src/components/bujo/WeekSummary.test.tsx
git commit -m "feat: add entry type symbols to WeekSummary meetings and attention sections"
```

---

## Task 5: Remove EntryContextPopover from WeekSummary

**Files:**
- Modify: `frontend/src/components/bujo/WeekSummary.tsx`

**Step 1: Write test for popover removal**

Add to `frontend/src/components/bujo/WeekSummary.test.tsx`:

```tsx
describe('WeekSummary without popover', () => {
  it('does not render entry context popover wrapper', () => {
    const days: DayEntries[] = [{
      date: '2026-01-25',
      entries: [
        { id: 1, content: 'Event', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
        { id: 2, content: 'Child', type: 'task', priority: 'none', parentId: 1, loggedDate: '2026-01-25' },
      ],
    }]
    render(<WeekSummary days={days} />)
    expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
  })
})
```

**Step 2: Run test to see current state**

Run: `cd frontend && npm test -- --run WeekSummary.test.tsx`
Expected: May fail if popover is currently rendered

**Step 3: Remove popover from WeekSummary**

In `frontend/src/components/bujo/WeekSummary.tsx`:

1. Remove the import:
```tsx
// Remove this line:
import { EntryContextPopover } from './EntryContextPopover';
```

2. Remove the NOOP constants:
```tsx
// Remove these lines:
const NOOP_ACTION = () => {};
const NOOP_NAVIGATE = () => {};
```

3. In meetings section, replace the conditional popover wrapping with just the button (remove lines ~97-121 that wrap meetingItem):

```tsx
{eventsWithChildren.map(event => {
  const meetingItem = (
    <button
      key={event.id}
      type="button"
      className="w-full flex items-center justify-between p-2 rounded-lg border border-border bg-card hover:bg-muted/50 cursor-pointer text-left"
    >
      <span className="flex items-center gap-2">
        <EntrySymbol type={event.type} />
        <span className="text-sm">{event.content}</span>
      </span>
      <span className="text-xs text-muted-foreground">
        {getChildCount(event.id)} items
      </span>
    </button>
  );
  return meetingItem;
})}
```

4. Similarly for attention section, replace conditional popover wrapping with just the button (remove lines ~170-195).

5. Remove `onAction` and `onNavigate` from `WeekSummaryProps` interface if they become unused.

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run WeekSummary.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekSummary.tsx frontend/src/components/bujo/WeekSummary.test.tsx
git commit -m "refactor: remove EntryContextPopover from WeekSummary"
```

---

## Task 6: Remove EntryContextPopover and Inline Expansion from SearchView

**Files:**
- Modify: `frontend/src/components/bujo/SearchView.tsx`
- Modify: `frontend/src/components/bujo/SearchView.core.test.tsx`

**Step 1: Write test for simplified SearchView**

Add to or update `frontend/src/components/bujo/SearchView.core.test.tsx`:

```tsx
describe('SearchView simplified UI', () => {
  it('does not render context popover', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ID: 1, Content: 'Test', Type: 'task', Priority: 'none', CreatedAt: '2026-01-25', ParentID: 2 }
    ])

    render(<SearchView />)
    const input = screen.getByPlaceholderText('Search entries...')
    await userEvent.type(input, 'test')

    await waitFor(() => {
      expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
    })
  })

  it('shows context dot for entries with parents', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ID: 1, Content: 'Child entry', Type: 'task', Priority: 'none', CreatedAt: '2026-01-25', ParentID: 99 }
    ])

    render(<SearchView />)
    const input = screen.getByPlaceholderText('Search entries...')
    await userEvent.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTestId('context-dot')).toBeInTheDocument()
    })
  })

  it('does not show ContextPill', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ID: 1, Content: 'Test', Type: 'task', Priority: 'none', CreatedAt: '2026-01-25', ParentID: 2 }
    ])

    render(<SearchView />)
    const input = screen.getByPlaceholderText('Search entries...')
    await userEvent.type(input, 'test')

    await waitFor(() => {
      expect(screen.queryByTestId('context-pill')).not.toBeInTheDocument()
    })
  })
})
```

**Step 2: Run test to see current state**

Run: `cd frontend && npm test -- --run SearchView.core.test.tsx`
Expected: Tests may fail initially

**Step 3: Simplify SearchView**

In `frontend/src/components/bujo/SearchView.tsx`:

1. Remove imports:
```tsx
// Remove these:
import { ContextPill } from './ContextPill';
import { EntryContextPopover } from './EntryContextPopover';
```

2. Remove state related to inline expansion and popover:
```tsx
// Remove these useState hooks:
const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());
const [ancestorsMap, setAncestorsMap] = useState<Map<number, AncestorEntry[]>>(new Map());
const [popoverEntry, setPopoverEntry] = useState<SearchResult | null>(null);
const [allEntries, setAllEntries] = useState<Entry[]>([]);
const [openPopoverId, setOpenPopoverId] = useState<number | null>(null);
```

3. Remove functions:
   - `toggleExpanded`
   - `handleEntryClick`
   - `handlePopoverAction`
   - `handlePopoverNavigate`
   - `handlePopoverOpenChange`
   - `convertToEntry` (if only used for popover)

4. Simplify the result rendering to use a simple div with context dot instead of EntryContextPopover wrapper. Replace the entire result map (lines ~412-527) with:

```tsx
{results.map((result, index) => {
  const isSelected = index === selectedIndex;

  return (
    <div
      key={result.id}
      data-result-id={result.id}
      onDoubleClick={() => onNavigateToEntry?.(result)}
      className={cn(
        'p-3 rounded-lg border border-border cursor-pointer',
        'bg-card transition-colors group',
        !isSelected && 'hover:bg-secondary/30',
        isSelected && 'ring-2 ring-primary'
      )}
    >
      <div className="flex items-start gap-3">
        {/* Context dot */}
        {result.parentId !== null ? (
          <span
            data-testid="context-dot"
            className="w-1.5 h-1.5 rounded-full bg-muted-foreground flex-shrink-0 mt-2"
            title="Has parent context"
          />
        ) : (
          <span className="w-1.5" />
        )}

        <span className="inline-flex items-center gap-1 flex-shrink-0">
          {result.type === 'task' || result.type === 'done' ? (
            <button
              data-testid="entry-symbol"
              onClick={(e) => result.type === 'task' ? handleMarkDone(result.id, e) : handleMarkUndone(result.id, e)}
              title={result.type === 'task' ? 'Task' : 'Done'}
              className={cn(
                'text-lg font-mono w-5 text-center cursor-pointer hover:opacity-70 transition-opacity',
                result.type === 'done' && 'text-bujo-done',
                result.type === 'task' && 'text-bujo-task',
              )}
            >
              {getSymbol(result.type)}
            </button>
          ) : (
            <span
              data-testid="entry-symbol"
              className={cn(
                'text-lg font-mono w-5 text-center',
                result.type === 'note' && 'text-bujo-note',
                result.type === 'event' && 'text-bujo-event',
                result.type === 'cancelled' && 'text-bujo-cancelled',
              )}
            >
              {getSymbol(result.type)}
            </span>
          )}
          {result.priority !== 'none' && (
            <span className={cn(
              'text-xs font-bold',
              result.priority === 'low' && 'text-priority-low',
              result.priority === 'medium' && 'text-priority-medium',
              result.priority === 'high' && 'text-priority-high',
            )}>
              {PRIORITY_SYMBOLS[result.priority]}
            </span>
          )}
        </span>

        <div className="flex-1 min-w-0">
          <p className={cn(
            'text-sm',
            result.type === 'done' && 'text-bujo-done',
            result.type === 'cancelled' && 'line-through text-muted-foreground'
          )}>
            {result.content}
          </p>
          <p className="text-xs text-muted-foreground mt-1">
            {formatDate(result.date)}
          </p>
        </div>

        <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100">
          #{result.id}
        </span>
      </div>
    </div>
  );
})}
```

5. Keep `ancestorCounts` and `loadingCounts` state since they're used for... actually, remove these too since we're using a dot not a count badge. Simplify to just track which entries have parents (parentId !== null).

**Step 4: Run tests to verify they pass**

Run: `cd frontend && npm test -- --run SearchView`
Expected: PASS (update other SearchView tests as needed)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/SearchView.tsx frontend/src/components/bujo/SearchView.core.test.tsx
git commit -m "refactor: remove popover and inline expansion from SearchView, add context dot"
```

---

## Task 7: Add Context Panel Toggle to App

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/App.tsx` tests (create if needed)

**Step 1: Write failing test**

Create `frontend/src/App.contextPanel.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'

// Mock all the wails bindings
vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn().mockResolvedValue({ Days: [], Overdue: [] }),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
}))

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(() => () => {}),
}))

vi.mock('@/contexts/SettingsContext', () => ({
  useSettings: () => ({ defaultView: 'today', theme: 'system' }),
  SettingsProvider: ({ children }: { children: React.ReactNode }) => children,
}))

describe('App context panel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('toggles context panel with c key', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
    })

    // Initially panel is hidden
    expect(screen.queryByTestId('context-panel')).not.toBeInTheDocument()

    // Press c to show
    await user.keyboard('c')
    expect(screen.getByTestId('context-panel')).toBeInTheDocument()

    // Press c again to hide
    await user.keyboard('c')
    expect(screen.queryByTestId('context-panel')).not.toBeInTheDocument()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run App.contextPanel.test.tsx`
Expected: FAIL

**Step 3: Add context panel toggle to App**

In `frontend/src/App.tsx`:

1. Add import:
```tsx
import { ContextPanel } from '@/components/bujo/ContextPanel'
```

2. Add state for panel visibility and selected entry:
```tsx
const [showContextPanel, setShowContextPanel] = useState(false)
const [selectedEntry, setSelectedEntry] = useState<Entry | null>(null)
```

3. Add keyboard handler for 'c' key (in the existing keyboard handler or create new useEffect):
```tsx
useEffect(() => {
  const handleKeyDown = (e: KeyboardEvent) => {
    // Don't trigger if typing in input
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return

    if (e.key === 'c') {
      e.preventDefault()
      setShowContextPanel(prev => !prev)
    }
  }

  window.addEventListener('keydown', handleKeyDown)
  return () => window.removeEventListener('keydown', handleKeyDown)
}, [])
```

4. Add the ContextPanel to the layout. Modify the main content area to include the panel:

```tsx
<main className="flex-1 flex overflow-hidden">
  <div className={cn("flex-1 overflow-auto p-6", showContextPanel && "mr-64")}>
    {/* existing view content */}
  </div>

  {showContextPanel && (
    <aside className="w-64 flex-shrink-0" data-testid="context-panel">
      <ContextPanel
        selectedEntry={selectedEntry}
        entries={flatEntries}
      />
    </aside>
  )}
</main>
```

5. Wire up entry selection to update `selectedEntry` state when entries are clicked/selected.

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run App.contextPanel.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/App.tsx frontend/src/App.contextPanel.test.tsx
git commit -m "feat: add context panel toggle with c key in App"
```

---

## Task 8: Wire Up Entry Selection to Context Panel

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/components/bujo/DayView.tsx`

**Step 1: Write test for selection updating panel**

Add to `frontend/src/App.contextPanel.test.tsx`:

```tsx
it('updates context panel when entry is selected', async () => {
  const mockDays = [{
    date: '2026-01-25',
    entries: [
      { id: 1, content: 'Parent', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
      { id: 2, content: 'Child', type: 'note', priority: 'none', parentId: 1, loggedDate: '2026-01-25' },
    ],
  }]

  vi.mocked(GetAgenda).mockResolvedValue({ Days: mockDays, Overdue: [] })

  const user = userEvent.setup()
  render(<App />)

  await waitFor(() => {
    expect(screen.getByText('Parent')).toBeInTheDocument()
  })

  // Show context panel
  await user.keyboard('c')

  // Click on child entry
  await user.click(screen.getByText('Child'))

  // Panel should show the hierarchy
  await waitFor(() => {
    const panel = screen.getByTestId('context-panel')
    expect(panel).toHaveTextContent('Parent')
    expect(panel).toHaveTextContent('Child')
  })
})
```

**Step 2: Run test**

Run: `cd frontend && npm test -- --run App.contextPanel.test.tsx`

**Step 3: Wire up selection**

In `frontend/src/App.tsx`, update the `onSelectEntry` handler to also set the selected entry for the context panel:

```tsx
const handleSelectEntry = useCallback((entryId: number) => {
  // Find the entry in the flat list
  const entry = flatEntries.find(e => e.id === entryId)
  if (entry) {
    setSelectedEntry(entry)
  }
}, [flatEntries])
```

Pass this to DayView:
```tsx
<DayView
  day={days[selectedIndex]}
  selectedEntryId={selectedEntry?.id}
  onSelectEntry={handleSelectEntry}
  // ... other props
/>
```

**Step 4: Run tests**

Run: `cd frontend && npm test -- --run App.contextPanel`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/App.tsx frontend/src/App.contextPanel.test.tsx
git commit -m "feat: wire entry selection to context panel display"
```

---

## Task 9: Clean Up Unused Files and Run Full Test Suite

**Files:**
- Potentially delete: `frontend/src/components/bujo/ContextPill.tsx`
- Potentially delete: `frontend/src/components/bujo/ContextPill.test.tsx`
- Review: `frontend/src/components/bujo/EntryContextPopover.tsx`

**Step 1: Check for remaining usages**

Run: `cd frontend && grep -r "ContextPill" src/`
Run: `cd frontend && grep -r "EntryContextPopover" src/`

**Step 2: Delete unused files if no longer imported**

If ContextPill has no imports:
```bash
rm frontend/src/components/bujo/ContextPill.tsx
rm frontend/src/components/bujo/ContextPill.test.tsx
```

If EntryContextPopover has no imports (may want to keep for potential future use):
```bash
# Optional: delete or mark as deprecated
```

**Step 3: Run full test suite**

Run: `cd frontend && npm test`
Expected: All tests pass

**Step 4: Commit cleanup**

```bash
git add -A
git commit -m "chore: remove unused ContextPill component"
```

---

## Task 10: Final Integration Test and PR

**Step 1: Run full test suite**

```bash
cd frontend && npm test
```

**Step 2: Manual verification checklist**

- [ ] DayView: entries are directly actionable (no popover gate)
- [ ] DayView: context dot shows on entries with parents
- [ ] DayView: pressing 'c' toggles context panel
- [ ] DayView: selecting entry updates context panel
- [ ] SearchView: entries are directly actionable
- [ ] SearchView: context dot shows on entries with parents
- [ ] SearchView: no ContextPill badges
- [ ] WeekSummary: entry symbols show in Meetings section
- [ ] WeekSummary: entry symbols show in Needs Attention section
- [ ] Context panel: shows empty state for root entries
- [ ] Context panel: shows full hierarchy for nested entries
- [ ] Context panel: highlights selected entry

**Step 3: Create final commit if needed**

```bash
git status
# If any uncommitted changes:
git add -A
git commit -m "chore: final cleanup and test fixes"
```

**Step 4: Push and create PR**

```bash
git push -u origin feature/context-panel-ux
gh pr create --title "feat: Replace popover with context panel for entry hierarchy" --body "$(cat <<'EOF'
## Summary
- Remove EntryContextPopover as primary interaction in DayView, SearchView, WeekSummary
- Add toggleable ContextPanel sidebar (c key) showing entry hierarchy
- Add context dot indicator for entries with ancestors
- Add entry type symbols to WeekSummary meetings and attention sections
- Re-enable direct entry actions (hover bar, click symbol, context menu)

## Test plan
- [ ] All 883+ tests pass
- [ ] Manual verification of all views
- [ ] Keyboard shortcut 'c' toggles panel

Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
