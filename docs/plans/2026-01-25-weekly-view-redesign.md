# Weekly View Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a 2×3 calendar grid weekly view showing events and priority entries with context panel integration.

**Architecture:** New WeekView component with DayBox/WeekendBox children for calendar grid, reusing ContextTree and EntryActionBar from existing components. Filters to events + priority entries only.

**Tech Stack:** React 18, TypeScript, Tailwind CSS, Vitest, React Testing Library

---

## Task 1: Create WeekEntry Component (Presentational)

**Files:**
- Create: `frontend/src/components/bujo/WeekEntry.tsx`
- Create: `frontend/src/components/bujo/WeekEntry.test.tsx`

**Step 1: Write the failing test**

```typescript
// frontend/src/components/bujo/WeekEntry.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { WeekEntry } from './WeekEntry';
import { Entry } from '@/types/bujo';

describe('WeekEntry', () => {
  const mockEntry: Entry = {
    id: 1,
    content: 'Test entry',
    type: 'task',
    priority: 'high',
    parentId: null,
    createdAt: '2026-01-20T10:00:00Z',
    children: [],
  };

  it('renders entry with symbol and content', () => {
    render(<WeekEntry entry={mockEntry} />);
    expect(screen.getByText('•')).toBeInTheDocument();
    expect(screen.getByText('Test entry')).toBeInTheDocument();
  });

  it('shows priority indicator for high priority', () => {
    render(<WeekEntry entry={mockEntry} />);
    expect(screen.getByText('!!!')).toBeInTheDocument();
  });

  it('calls onSelect when clicked', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<WeekEntry entry={mockEntry} onSelect={onSelect} />);

    await user.click(screen.getByRole('button'));
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it('shows selected state', () => {
    const { container } = render(<WeekEntry entry={mockEntry} isSelected={true} />);
    expect(container.firstChild).toHaveClass('bg-primary/10');
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test WeekEntry.test.tsx`
Expected: FAIL with "Cannot find module './WeekEntry'"

**Step 3: Write minimal implementation**

```typescript
// frontend/src/components/bujo/WeekEntry.tsx
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';

interface WeekEntryProps {
  entry: Entry;
  isSelected?: boolean;
  onSelect?: () => void;
  datePrefix?: string;
}

export function WeekEntry({ entry, isSelected, onSelect, datePrefix }: WeekEntryProps) {
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];

  return (
    <div
      className={cn(
        'px-2 py-1.5 rounded-lg text-sm transition-colors',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
    >
      <button
        onClick={onSelect}
        className="flex items-center gap-2 text-left min-w-0 w-full"
      >
        {datePrefix && (
          <span className="text-muted-foreground text-xs flex-shrink-0">
            {datePrefix}
          </span>
        )}

        <span className="text-muted-foreground flex-shrink-0">
          {symbol}
        </span>

        {prioritySymbol && (
          <span className="text-orange-500 font-medium flex-shrink-0">
            {prioritySymbol}
          </span>
        )}

        <span className="flex-1 truncate">{entry.content}</span>
      </button>
    </div>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test WeekEntry.test.tsx`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekEntry.tsx frontend/src/components/bujo/WeekEntry.test.tsx
git commit -m "feat: add WeekEntry component for calendar grid items

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Add Entry Filtering Utility

**Files:**
- Create: `frontend/src/lib/weekViewFilters.ts`
- Create: `frontend/src/lib/weekViewFilters.test.ts`

**Step 1: Write the failing test**

```typescript
// frontend/src/lib/weekViewFilters.test.ts
import { describe, it, expect } from 'vitest';
import { filterWeekEntries } from './weekViewFilters';
import { Entry } from '@/types/bujo';

describe('filterWeekEntries', () => {
  it('includes all events regardless of priority', () => {
    const entries: Entry[] = [
      { id: 1, content: 'Meeting', type: 'event', priority: null, parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
      { id: 2, content: 'Another meeting', type: 'event', priority: 'high', parentId: null, createdAt: '2026-01-20T11:00:00Z', children: [] },
    ];

    const result = filterWeekEntries(entries);
    expect(result).toHaveLength(2);
  });

  it('includes entries with low priority', () => {
    const entries: Entry[] = [
      { id: 1, content: 'Task', type: 'task', priority: 'low', parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
    ];

    const result = filterWeekEntries(entries);
    expect(result).toHaveLength(1);
  });

  it('includes entries with medium priority', () => {
    const entries: Entry[] = [
      { id: 1, content: 'Task', type: 'task', priority: 'medium', parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
    ];

    const result = filterWeekEntries(entries);
    expect(result).toHaveLength(1);
  });

  it('includes entries with high priority', () => {
    const entries: Entry[] = [
      { id: 1, content: 'Task', type: 'task', priority: 'high', parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
    ];

    const result = filterWeekEntries(entries);
    expect(result).toHaveLength(1);
  });

  it('excludes entries without priority (unless events)', () => {
    const entries: Entry[] = [
      { id: 1, content: 'Task', type: 'task', priority: null, parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
      { id: 2, content: 'Note', type: 'note', priority: null, parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
    ];

    const result = filterWeekEntries(entries);
    expect(result).toHaveLength(0);
  });

  it('flattens hierarchical entries', () => {
    const entries: Entry[] = [
      {
        id: 1,
        content: 'Parent event',
        type: 'event',
        priority: null,
        parentId: null,
        createdAt: '2026-01-20T10:00:00Z',
        children: [
          { id: 2, content: 'Child task', type: 'task', priority: 'high', parentId: 1, createdAt: '2026-01-20T10:00:00Z', children: [] },
        ],
      },
    ];

    const result = filterWeekEntries(entries);
    expect(result).toHaveLength(2);
    expect(result.find(e => e.id === 1)).toBeDefined();
    expect(result.find(e => e.id === 2)).toBeDefined();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test weekViewFilters.test.ts`
Expected: FAIL with "Cannot find module './weekViewFilters'"

**Step 3: Write minimal implementation**

```typescript
// frontend/src/lib/weekViewFilters.ts
import { Entry } from '@/types/bujo';

function flattenEntries(entries: Entry[]): Entry[] {
  const result: Entry[] = [];
  function traverse(items: Entry[]) {
    for (const entry of items) {
      result.push(entry);
      if (entry.children && entry.children.length > 0) {
        traverse(entry.children);
      }
    }
  }
  traverse(entries);
  return result;
}

export function filterWeekEntries(entries: Entry[]): Entry[] {
  const flattened = flattenEntries(entries);

  return flattened.filter(entry =>
    entry.type === 'event' ||
    (entry.priority === 'low' || entry.priority === 'medium' || entry.priority === 'high')
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test weekViewFilters.test.ts`
Expected: PASS (6 tests)

**Step 5: Commit**

```bash
git add frontend/src/lib/weekViewFilters.ts frontend/src/lib/weekViewFilters.test.ts
git commit -m "feat: add week view entry filtering utility

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Create DayBox Component

**Files:**
- Create: `frontend/src/components/bujo/DayBox.tsx`
- Create: `frontend/src/components/bujo/DayBox.test.tsx`

**Step 1: Write the failing test**

```typescript
// frontend/src/components/bujo/DayBox.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { DayBox } from './DayBox';
import { Entry } from '@/types/bujo';

describe('DayBox', () => {
  const mockEntries: Entry[] = [
    { id: 1, content: 'Meeting', type: 'event', priority: null, parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
    { id: 2, content: 'Task', type: 'task', priority: 'high', parentId: null, createdAt: '2026-01-20T11:00:00Z', children: [] },
  ];

  it('renders day number and name', () => {
    render(<DayBox dayNumber={20} dayName="Mon" entries={[]} />);
    expect(screen.getByText('20')).toBeInTheDocument();
    expect(screen.getByText('Mon')).toBeInTheDocument();
  });

  it('shows "No events" when empty', () => {
    render(<DayBox dayNumber={20} dayName="Mon" entries={[]} />);
    expect(screen.getByText('No events')).toBeInTheDocument();
  });

  it('renders all entries', () => {
    render(<DayBox dayNumber={20} dayName="Mon" entries={mockEntries} />);
    expect(screen.getByText('Meeting')).toBeInTheDocument();
    expect(screen.getByText('Task')).toBeInTheDocument();
  });

  it('calls onSelectEntry when entry clicked', async () => {
    const user = userEvent.setup();
    const onSelectEntry = vi.fn();
    render(<DayBox dayNumber={20} dayName="Mon" entries={mockEntries} onSelectEntry={onSelectEntry} />);

    await user.click(screen.getByText('Meeting'));
    expect(onSelectEntry).toHaveBeenCalledWith(mockEntries[0]);
  });

  it('highlights selected entry', () => {
    const { container } = render(
      <DayBox dayNumber={20} dayName="Mon" entries={mockEntries} selectedEntry={mockEntries[0]} />
    );
    const selectedItem = container.querySelector('.bg-primary\\/10');
    expect(selectedItem).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test DayBox.test.tsx`
Expected: FAIL with "Cannot find module './DayBox'"

**Step 3: Write minimal implementation**

```typescript
// frontend/src/components/bujo/DayBox.tsx
import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';

interface DayBoxProps {
  dayNumber: number;
  dayName: string;
  entries: Entry[];
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
}

export function DayBox({ dayNumber, dayName, entries, selectedEntry, onSelectEntry }: DayBoxProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3 flex items-baseline gap-2">
        <span className="text-2xl font-semibold">{dayNumber}</span>
        <span className="text-sm text-muted-foreground">{dayName}</span>
      </div>

      <div className="space-y-1 max-h-64 overflow-y-auto">
        {entries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          entries.map(entry => (
            <WeekEntry
              key={entry.id}
              entry={entry}
              isSelected={selectedEntry?.id === entry.id}
              onSelect={() => onSelectEntry?.(entry)}
            />
          ))
        )}
      </div>
    </div>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test DayBox.test.tsx`
Expected: PASS (5 tests)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/DayBox.tsx frontend/src/components/bujo/DayBox.test.tsx
git commit -m "feat: add DayBox component for single day calendar view

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Create WeekendBox Component

**Files:**
- Create: `frontend/src/components/bujo/WeekendBox.tsx`
- Create: `frontend/src/components/bujo/WeekendBox.test.tsx`

**Step 1: Write the failing test**

```typescript
// frontend/src/components/bujo/WeekendBox.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { WeekendBox } from './WeekendBox';
import { Entry } from '@/types/bujo';

describe('WeekendBox', () => {
  const saturdayEntries: Entry[] = [
    { id: 1, content: 'Sat event', type: 'event', priority: null, parentId: null, createdAt: '2026-01-24T10:00:00Z', children: [] },
  ];

  const sundayEntries: Entry[] = [
    { id: 2, content: 'Sun task', type: 'task', priority: 'high', parentId: null, createdAt: '2026-01-25T10:00:00Z', children: [] },
  ];

  it('renders weekend header with date range', () => {
    render(<WeekendBox startDay={24} saturdayEntries={[]} sundayEntries={[]} />);
    expect(screen.getByText('24-25')).toBeInTheDocument();
    expect(screen.getByText('Weekend')).toBeInTheDocument();
  });

  it('shows "No events" when both days empty', () => {
    render(<WeekendBox startDay={24} saturdayEntries={[]} sundayEntries={[]} />);
    expect(screen.getByText('No events')).toBeInTheDocument();
  });

  it('prefixes Saturday entries with "Sat:"', () => {
    render(<WeekendBox startDay={24} saturdayEntries={saturdayEntries} sundayEntries={[]} />);
    expect(screen.getByText('Sat:')).toBeInTheDocument();
    expect(screen.getByText('Sat event')).toBeInTheDocument();
  });

  it('prefixes Sunday entries with "Sun:"', () => {
    render(<WeekendBox startDay={24} saturdayEntries={[]} sundayEntries={sundayEntries} />);
    expect(screen.getByText('Sun:')).toBeInTheDocument();
    expect(screen.getByText('Sun task')).toBeInTheDocument();
  });

  it('shows Saturday entries before Sunday entries', () => {
    const { container } = render(
      <WeekendBox startDay={24} saturdayEntries={saturdayEntries} sundayEntries={sundayEntries} />
    );
    const entries = container.querySelectorAll('button');
    expect(entries[0]).toHaveTextContent('Sat:');
    expect(entries[1]).toHaveTextContent('Sun:');
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test WeekendBox.test.tsx`
Expected: FAIL with "Cannot find module './WeekendBox'"

**Step 3: Write minimal implementation**

```typescript
// frontend/src/components/bujo/WeekendBox.tsx
import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';

interface WeekendBoxProps {
  startDay: number;
  saturdayEntries: Entry[];
  sundayEntries: Entry[];
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
}

export function WeekendBox({
  startDay,
  saturdayEntries,
  sundayEntries,
  selectedEntry,
  onSelectEntry,
}: WeekendBoxProps) {
  const allEntries = [
    ...saturdayEntries.map(e => ({ entry: e, prefix: 'Sat:' })),
    ...sundayEntries.map(e => ({ entry: e, prefix: 'Sun:' })),
  ];

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3 flex items-baseline gap-2">
        <span className="text-2xl font-semibold">{startDay}-{startDay + 1}</span>
        <span className="text-sm text-muted-foreground">Weekend</span>
      </div>

      <div className="space-y-1 max-h-64 overflow-y-auto">
        {allEntries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          allEntries.map(({ entry, prefix }) => (
            <WeekEntry
              key={entry.id}
              entry={entry}
              datePrefix={prefix}
              isSelected={selectedEntry?.id === entry.id}
              onSelect={() => onSelectEntry?.(entry)}
            />
          ))
        )}
      </div>
    </div>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test WeekendBox.test.tsx`
Expected: PASS (5 tests)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekendBox.tsx frontend/src/components/bujo/WeekendBox.test.tsx
git commit -m "feat: add WeekendBox component for combined Sat-Sun view

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Create WeekView Component (Orchestrator)

**Files:**
- Create: `frontend/src/components/bujo/WeekView.tsx`
- Create: `frontend/src/components/bujo/WeekView.test.tsx`

**Step 1: Write the failing test**

```typescript
// frontend/src/components/bujo/WeekView.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { WeekView } from './WeekView';
import { DayEntries } from '@/types/bujo';

describe('WeekView', () => {
  const mockWeekData: DayEntries[] = [
    {
      date: '2026-01-19',
      entries: [
        { id: 1, content: 'Mon meeting', type: 'event', priority: null, parentId: null, createdAt: '2026-01-19T10:00:00Z', children: [] },
      ],
    },
    {
      date: '2026-01-20',
      entries: [
        { id: 2, content: 'Tue task', type: 'task', priority: 'high', parentId: null, createdAt: '2026-01-20T10:00:00Z', children: [] },
      ],
    },
    {
      date: '2026-01-21',
      entries: [],
    },
    {
      date: '2026-01-22',
      entries: [],
    },
    {
      date: '2026-01-23',
      entries: [
        { id: 3, content: 'Fri event', type: 'event', priority: null, parentId: null, createdAt: '2026-01-23T10:00:00Z', children: [] },
      ],
    },
    {
      date: '2026-01-24',
      entries: [
        { id: 4, content: 'Sat lunch', type: 'event', priority: null, parentId: null, createdAt: '2026-01-24T12:00:00Z', children: [] },
      ],
    },
    {
      date: '2026-01-25',
      entries: [
        { id: 5, content: 'Sun task', type: 'task', priority: 'high', parentId: null, createdAt: '2026-01-25T10:00:00Z', children: [] },
      ],
    },
  ];

  it('renders 5 day boxes plus weekend box', () => {
    const { container } = render(<WeekView days={mockWeekData} />);
    const boxes = container.querySelectorAll('.rounded-lg.border');
    expect(boxes).toHaveLength(6);
  });

  it('renders week date range header', () => {
    render(<WeekView days={mockWeekData} />);
    expect(screen.getByText(/Jan 19.*Jan 25, 2026/)).toBeInTheDocument();
  });

  it('filters to events and priority entries only', () => {
    const withNonPriority: DayEntries[] = [
      {
        date: '2026-01-19',
        entries: [
          { id: 1, content: 'Meeting', type: 'event', priority: null, parentId: null, createdAt: '2026-01-19T10:00:00Z', children: [] },
          { id: 2, content: 'Task no priority', type: 'task', priority: null, parentId: null, createdAt: '2026-01-19T11:00:00Z', children: [] },
          { id: 3, content: 'Task with priority', type: 'task', priority: 'high', parentId: null, createdAt: '2026-01-19T12:00:00Z', children: [] },
        ],
      },
      ...mockWeekData.slice(1),
    ];

    render(<WeekView days={withNonPriority} />);
    expect(screen.getByText('Meeting')).toBeInTheDocument();
    expect(screen.getByText('Task with priority')).toBeInTheDocument();
    expect(screen.queryByText('Task no priority')).not.toBeInTheDocument();
  });

  it('shows context panel when entry selected', async () => {
    const user = userEvent.setup();
    render(<WeekView days={mockWeekData} />);

    await user.click(screen.getByText('Mon meeting'));
    expect(screen.getByText('Context')).toBeInTheDocument();
  });

  it('shows "No entry selected" initially', () => {
    render(<WeekView days={mockWeekData} />);
    expect(screen.getByText('No entry selected')).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test WeekView.test.tsx`
Expected: FAIL with "Cannot find module './WeekView'"

**Step 3: Write minimal implementation**

```typescript
// frontend/src/components/bujo/WeekView.tsx
import { useState, useMemo } from 'react';
import { DayEntries, Entry } from '@/types/bujo';
import { DayBox } from './DayBox';
import { WeekendBox } from './WeekendBox';
import { filterWeekEntries } from '@/lib/weekViewFilters';
import { format, parseISO } from 'date-fns';

interface WeekViewProps {
  days: DayEntries[];
}

interface TreeNode {
  entry: Entry;
  children: TreeNode[];
}

function buildTree(entries: Entry[]): TreeNode[] {
  if (entries.length === 0) return [];

  const entryMap = new Map<number, Entry>();
  const childrenMap = new Map<number | null, Entry[]>();

  for (const entry of entries) {
    entryMap.set(entry.id, entry);
    const parentId = entry.parentId;
    if (!childrenMap.has(parentId)) {
      childrenMap.set(parentId, []);
    }
    childrenMap.get(parentId)!.push(entry);
  }

  function buildNode(entry: Entry): TreeNode {
    const children = childrenMap.get(entry.id) || [];
    return {
      entry,
      children: children.map(buildNode),
    };
  }

  const roots = childrenMap.get(null) || [];
  return roots.map(buildNode);
}

function ContextTree({ nodes, selectedEntryId, depth = 0 }: { nodes: TreeNode[]; selectedEntryId?: number; depth?: number }) {
  return (
    <>
      {nodes.map((node) => (
        <div key={node.entry.id}>
          <div
            className="flex items-center gap-2 text-sm py-0.5 font-mono"
            style={{ paddingLeft: `${depth * 12}px` }}
          >
            <span className={node.entry.id === selectedEntryId ? 'text-foreground' : 'text-muted-foreground'}>
              {node.entry.content}
            </span>
          </div>
          {node.children.length > 0 && (
            <ContextTree nodes={node.children} selectedEntryId={selectedEntryId} depth={depth + 1} />
          )}
        </div>
      ))}
    </>
  );
}

export function WeekView({ days }: WeekViewProps) {
  const [selectedEntry, setSelectedEntry] = useState<Entry | undefined>();

  const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'];

  const weekDays = days.slice(0, 5).map((day, index) => ({
    ...day,
    dayName: dayNames[index],
    dayNumber: parseISO(day.date).getDate(),
  }));

  const saturday = days[5];
  const sunday = days[6];

  const filteredWeekDays = weekDays.map(day => ({
    ...day,
    entries: filterWeekEntries(day.entries),
  }));

  const filteredSaturday = saturday ? filterWeekEntries(saturday.entries) : [];
  const filteredSunday = sunday ? filterWeekEntries(sunday.entries) : [];

  const startDate = days[0] ? parseISO(days[0].date) : new Date();
  const endDate = days[6] ? parseISO(days[6].date) : new Date();
  const dateRange = `${format(startDate, 'MMM d')} – ${format(endDate, 'MMM d, yyyy')}`;

  const allEntries = days.flatMap(day => {
    const flatten = (entries: Entry[]): Entry[] => {
      const result: Entry[] = [];
      for (const entry of entries) {
        result.push(entry);
        if (entry.children && entry.children.length > 0) {
          result.push(...flatten(entry.children));
        }
      }
      return result;
    };
    return flatten(day.entries);
  });

  const contextTree = useMemo(() => buildTree(allEntries), [allEntries]);

  return (
    <div className="flex h-full gap-4">
      <div className="flex-1 overflow-y-auto">
        <div className="mb-4">
          <h2 className="text-lg font-semibold">Weekly Review</h2>
          <p className="text-sm text-muted-foreground">{dateRange}</p>
        </div>

        <div className="grid grid-cols-3 gap-4">
          {filteredWeekDays.map((day, index) => (
            <DayBox
              key={day.date}
              dayNumber={day.dayNumber}
              dayName={day.dayName}
              entries={day.entries}
              selectedEntry={selectedEntry}
              onSelectEntry={setSelectedEntry}
            />
          ))}

          {saturday && sunday && (
            <WeekendBox
              startDay={parseISO(saturday.date).getDate()}
              saturdayEntries={filteredSaturday}
              sundayEntries={filteredSunday}
              selectedEntry={selectedEntry}
              onSelectEntry={setSelectedEntry}
            />
          )}
        </div>
      </div>

      <div className="w-96 border-l border-border pl-4 overflow-y-auto">
        <div className="mb-3">
          <h3 className="text-sm font-medium">Context</h3>
        </div>

        {!selectedEntry ? (
          <p className="text-sm text-muted-foreground">No entry selected</p>
        ) : selectedEntry.parentId === null ? (
          <p className="text-sm text-muted-foreground">No context</p>
        ) : (
          <ContextTree nodes={contextTree} selectedEntryId={selectedEntry.id} />
        )}
      </div>
    </div>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test WeekView.test.tsx`
Expected: PASS (5 tests)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekView.tsx frontend/src/components/bujo/WeekView.test.tsx
git commit -m "feat: add WeekView orchestrator with calendar grid and context panel

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Add Entry Actions to WeekEntry

**Files:**
- Modify: `frontend/src/components/bujo/WeekEntry.tsx`
- Modify: `frontend/src/components/bujo/WeekEntry.test.tsx`

**Step 1: Write the failing test**

Add to existing `frontend/src/components/bujo/WeekEntry.test.tsx`:

```typescript
it('shows action bar on hover', async () => {
  const user = userEvent.setup();
  const callbacks = {
    onCancel: vi.fn(),
    onMigrate: vi.fn(),
  };

  const { container } = render(<WeekEntry entry={mockEntry} callbacks={callbacks} />);

  const entryDiv = container.firstChild as HTMLElement;
  await user.hover(entryDiv);

  expect(screen.getByLabelText(/mark done/i)).toBeInTheDocument();
});

it('calls callbacks when action buttons clicked', async () => {
  const user = userEvent.setup();
  const callbacks = {
    onCancel: vi.fn(),
  };

  const { container } = render(<WeekEntry entry={mockEntry} callbacks={callbacks} />);

  const entryDiv = container.firstChild as HTMLElement;
  await user.hover(entryDiv);
  await user.click(screen.getByLabelText(/mark done/i));

  expect(callbacks.onCancel).toHaveBeenCalledTimes(1);
});
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test WeekEntry.test.tsx`
Expected: FAIL with "callbacks is not a prop"

**Step 3: Write minimal implementation**

Modify `frontend/src/components/bujo/WeekEntry.tsx`:

```typescript
import { useState, useEffect } from 'react';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { EntryActionBar } from './EntryActions/EntryActionBar';
import { cn } from '@/lib/utils';

interface EntryCallbacks {
  onCancel?: () => void;
  onMigrate?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  onCyclePriority?: () => void;
  onMoveToList?: () => void;
}

interface WeekEntryProps {
  entry: Entry;
  isSelected?: boolean;
  onSelect?: () => void;
  datePrefix?: string;
  callbacks?: EntryCallbacks;
}

export function WeekEntry({ entry, isSelected, onSelect, datePrefix, callbacks }: WeekEntryProps) {
  const [isHovered, setIsHovered] = useState(false);
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown' || e.key === 'j' || e.key === 'k') {
        setIsHovered(false);
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <div
      className={cn(
        'group px-2 py-1.5 rounded-lg text-sm transition-colors',
        !isSelected && isHovered && 'bg-secondary/50',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <button
        onClick={onSelect}
        className="flex items-center gap-2 text-left min-w-0 w-full"
      >
        {datePrefix && (
          <span className="text-muted-foreground text-xs flex-shrink-0">
            {datePrefix}
          </span>
        )}

        <span className="text-muted-foreground flex-shrink-0">
          {symbol}
        </span>

        {prioritySymbol && (
          <span className="text-orange-500 font-medium flex-shrink-0">
            {prioritySymbol}
          </span>
        )}

        <span className="flex-1 truncate">{entry.content}</span>
      </button>

      {callbacks && (
        <div
          className={cn(
            'grid transition-all duration-150 ease-out grid-rows-[0fr]',
            isHovered && 'grid-rows-[1fr]'
          )}
        >
          <div className="overflow-hidden">
            <div
              className="pt-1"
              style={{ paddingLeft: 'calc(0.5rem + 0.5rem + 1ch)' }}
            >
              <EntryActionBar
                entry={entry}
                callbacks={callbacks}
                variant="always-visible"
                size="sm"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test WeekEntry.test.tsx`
Expected: PASS (all tests including new ones)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekEntry.tsx frontend/src/components/bujo/WeekEntry.test.tsx
git commit -m "feat: add entry actions to WeekEntry on hover

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Wire Up Entry Callbacks in WeekView

**Files:**
- Modify: `frontend/src/components/bujo/WeekView.tsx`
- Modify: `frontend/src/components/bujo/DayBox.tsx`
- Modify: `frontend/src/components/bujo/WeekendBox.tsx`

**Step 1: Write the failing test**

Add to `frontend/src/components/bujo/WeekView.test.tsx`:

```typescript
it('passes callbacks to entry components', () => {
  const callbacks = {
    onMarkDone: vi.fn(),
    onMigrate: vi.fn(),
  };

  render(<WeekView days={mockWeekData} callbacks={callbacks} />);
  // Test passes if no errors - implementation will thread callbacks through
});
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test WeekView.test.tsx`
Expected: FAIL with "callbacks is not a prop"

**Step 3: Write minimal implementation**

Update WeekView.tsx to accept and pass callbacks:

```typescript
export interface WeekViewCallbacks {
  onMarkDone?: (entry: Entry) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
  onCyclePriority?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
}

interface WeekViewProps {
  days: DayEntries[];
  callbacks?: WeekViewCallbacks;
}

export function WeekView({ days, callbacks = {} }: WeekViewProps) {
  // ... existing code ...

  const createEntryCallbacks = (entry: Entry) => ({
    onCancel: callbacks.onMarkDone ? () => callbacks.onMarkDone!(entry) : undefined,
    onMigrate: callbacks.onMigrate ? () => callbacks.onMigrate!(entry) : undefined,
    onEdit: callbacks.onEdit ? () => callbacks.onEdit!(entry) : undefined,
    onDelete: callbacks.onDelete ? () => callbacks.onDelete!(entry) : undefined,
    onCyclePriority: callbacks.onCyclePriority ? () => callbacks.onCyclePriority!(entry) : undefined,
    onMoveToList: callbacks.onMoveToList ? () => callbacks.onMoveToList!(entry) : undefined,
  });

  return (
    // ... existing JSX, but pass createEntryCallbacks to DayBox and WeekendBox
  );
}
```

Update DayBox.tsx to accept and pass callbacks:

```typescript
interface DayBoxProps {
  dayNumber: number;
  dayName: string;
  entries: Entry[];
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
  createEntryCallbacks?: (entry: Entry) => any;
}

export function DayBox({ dayNumber, dayName, entries, selectedEntry, onSelectEntry, createEntryCallbacks }: DayBoxProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      {/* ... header ... */}
      <div className="space-y-1 max-h-64 overflow-y-auto">
        {entries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          entries.map(entry => (
            <WeekEntry
              key={entry.id}
              entry={entry}
              isSelected={selectedEntry?.id === entry.id}
              onSelect={() => onSelectEntry?.(entry)}
              callbacks={createEntryCallbacks?.(entry)}
            />
          ))
        )}
      </div>
    </div>
  );
}
```

Update WeekendBox.tsx similarly.

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test WeekView.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekView.tsx frontend/src/components/bujo/DayBox.tsx frontend/src/components/bujo/WeekendBox.tsx
git commit -m "feat: wire up entry action callbacks through WeekView hierarchy

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Integrate WeekView into App.tsx

**Files:**
- Modify: `frontend/src/App.tsx`

**Step 1: Read current App.tsx structure**

Run: `cd frontend && head -100 src/App.tsx`

Understand how views are currently routed and how WeekSummary is integrated.

**Step 2: Write integration test**

Add to `frontend/src/App.test.tsx` (or create if doesn't exist):

```typescript
it('renders WeekView when view is "week"', () => {
  // Mock data setup
  render(<App />);
  // Trigger navigation to week view
  // Assert WeekView is rendered
});
```

**Step 3: Run test to verify it fails**

Run: `cd frontend && npm test App.test.tsx`

**Step 4: Add WeekView import and routing**

In App.tsx, import WeekView and add it to view routing logic similar to how other views are rendered.

**Step 5: Run test to verify it passes**

Run: `cd frontend && npm test App.test.tsx`

**Step 6: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat: integrate WeekView into App routing

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Update Navigation to Include WeekView

**Files:**
- Modify: `frontend/src/components/bujo/Sidebar.tsx` (or wherever navigation lives)

**Step 1: Find navigation component**

Run: `cd frontend && grep -r "Weekly" src/components/bujo/*.tsx`

**Step 2: Add WeekView to navigation**

Update navigation to include "Week" option that routes to new WeekView.

**Step 3: Test navigation**

Manual test: Click navigation → verify WeekView loads

**Step 4: Commit**

```bash
git add <navigation-file>
git commit -m "feat: add WeekView to navigation menu

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Execution Options

Plan complete and saved to `docs/plans/2026-01-25-weekly-view-redesign.md`.

Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
