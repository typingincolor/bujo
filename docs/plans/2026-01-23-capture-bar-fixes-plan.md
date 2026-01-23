# CaptureBar Fixes and Context Popover Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix bugs identified in PR #414 and implement the context popover feature for entry interaction across views.

**Architecture:** Fix CaptureBar bugs (symbols, auto-grow, prefix detection), add WeekSummary click handling, implement EntryContextPopover with Radix UI, and add navigation history with back button.

**Tech Stack:** React 19, TypeScript, Radix UI Popover, Vitest, React Testing Library

---

## Task 1: Install Radix UI Popover

**Files:**
- Modify: `frontend/package.json`

**Step 1: Install dependency**

Run: `cd frontend && npm install @radix-ui/react-popover`

**Step 2: Verify installation**

Run: `cd frontend && npm ls @radix-ui/react-popover`
Expected: Shows installed version

**Step 3: Commit**

```bash
git add frontend/package.json frontend/package-lock.json
git commit -m "chore: add @radix-ui/react-popover dependency"
```

---

## Task 2: Fix CaptureBar Type Button Labels (Symbols Instead of Words)

**Files:**
- Modify: `frontend/src/components/bujo/CaptureBar.tsx:139-154`
- Modify: `frontend/src/components/bujo/__tests__/CaptureBar.test.tsx`

**Step 1: Write failing test**

Add test to `frontend/src/components/bujo/__tests__/CaptureBar.test.tsx`:

```typescript
describe('type button labels', () => {
  it('displays bullet journal symbols instead of word labels', () => {
    render(<CaptureBar onSubmit={vi.fn()} />)

    expect(screen.getByRole('button', { name: /task/i })).toHaveTextContent('.')
    expect(screen.getByRole('button', { name: /note/i })).toHaveTextContent('-')
    expect(screen.getByRole('button', { name: /event/i })).toHaveTextContent('o')
    expect(screen.getByRole('button', { name: /question/i })).toHaveTextContent('?')
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run CaptureBar.test.tsx`
Expected: FAIL - buttons show "task", "note", etc. instead of symbols

**Step 3: Write minimal implementation**

In `CaptureBar.tsx`, add constant and update button rendering:

```typescript
const TYPE_SYMBOLS: Record<CaptureType, string> = {
  task: '.',
  note: '-',
  event: 'o',
  question: '?',
}

// Update button in JSX (around line 151):
<button
  key={type}
  type="button"
  onClick={() => setSelectedType(type)}
  aria-pressed={selectedType === type}
  aria-label={type}
  className={cn(
    'w-8 h-8 text-sm rounded font-mono',
    selectedType === type
      ? 'bg-primary text-primary-foreground'
      : 'bg-muted text-muted-foreground hover:bg-muted/80'
  )}
>
  {TYPE_SYMBOLS[type]}
</button>
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run CaptureBar.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/CaptureBar.tsx frontend/src/components/bujo/__tests__/CaptureBar.test.tsx
git commit -m "fix: display bullet journal symbols on CaptureBar type buttons"
```

---

## Task 3: Fix CaptureBar Textarea Auto-Grow

**Files:**
- Modify: `frontend/src/components/bujo/CaptureBar.tsx`
- Modify: `frontend/src/components/bujo/__tests__/CaptureBar.test.tsx`

**Step 1: Write failing test**

```typescript
describe('textarea auto-grow', () => {
  it('expands textarea height for multiline content', async () => {
    render(<CaptureBar onSubmit={vi.fn()} />)
    const textarea = screen.getByTestId('capture-bar-input')

    // Initial height should be single line
    const initialHeight = textarea.scrollHeight

    // Type multiline content
    await userEvent.type(textarea, 'Line 1\nLine 2\nLine 3')

    // Height should have increased
    expect(textarea.style.height).not.toBe('')
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run CaptureBar.test.tsx`
Expected: FAIL - textarea has no dynamic height

**Step 3: Write minimal implementation**

Add auto-resize effect in `CaptureBar.tsx`:

```typescript
// Add effect after existing effects
useEffect(() => {
  const textarea = textareaRef.current
  if (textarea) {
    textarea.style.height = 'auto'
    textarea.style.height = `${textarea.scrollHeight}px`
  }
}, [content])
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run CaptureBar.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/CaptureBar.tsx frontend/src/components/bujo/__tests__/CaptureBar.test.tsx
git commit -m "fix: auto-grow CaptureBar textarea for multiline content"
```

---

## Task 4: Fix Prefix Detection Bug (First Character Cut Off)

**Files:**
- Modify: `frontend/src/components/bujo/CaptureBar.tsx:87-99`
- Modify: `frontend/src/components/bujo/__tests__/CaptureBar.test.tsx`

**Step 1: Write failing test**

```typescript
describe('prefix detection edge cases', () => {
  it('preserves content when typing dash followed by space and text', async () => {
    render(<CaptureBar onSubmit={vi.fn()} />)
    const textarea = screen.getByTestId('capture-bar-input')

    // Type "- hello" character by character
    await userEvent.type(textarea, '- hello')

    // Should switch to note type and preserve "hello"
    expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
    expect(textarea).toHaveValue('hello')
  })

  it('does not switch type when prefix appears mid-content', async () => {
    render(<CaptureBar onSubmit={vi.fn()} />)
    const textarea = screen.getByTestId('capture-bar-input')

    await userEvent.type(textarea, 'buy - groceries')

    // Should remain as task (default) and preserve full content
    expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
    expect(textarea).toHaveValue('buy - groceries')
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run CaptureBar.test.tsx`
Expected: FAIL - "hello" becomes "ello" or type doesn't switch correctly

**Step 3: Write minimal implementation**

Fix `handleChange` in `CaptureBar.tsx`:

```typescript
const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
  const newValue = e.target.value

  // Only check for prefix at the start of input (when content was empty)
  if (content === '') {
    for (const [prefix, type] of Object.entries(PREFIX_TO_TYPE)) {
      if (newValue === prefix) {
        setSelectedType(type)
        setContent('')
        return
      }
      // Check if user typed prefix followed by more content
      if (newValue.startsWith(prefix) && newValue.length > prefix.length) {
        setSelectedType(type)
        setContent(newValue.slice(prefix.length))
        return
      }
    }
  }

  setContent(newValue)
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run CaptureBar.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/CaptureBar.tsx frontend/src/components/bujo/__tests__/CaptureBar.test.tsx
git commit -m "fix: preserve content when detecting type prefix in CaptureBar"
```

---

## Task 5: Create useNavigationHistory Hook

**Files:**
- Create: `frontend/src/hooks/useNavigationHistory.ts`
- Create: `frontend/src/hooks/__tests__/useNavigationHistory.test.ts`

**Step 1: Write failing test**

Create `frontend/src/hooks/__tests__/useNavigationHistory.test.ts`:

```typescript
import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useNavigationHistory } from '../useNavigationHistory'

describe('useNavigationHistory', () => {
  it('starts with no history', () => {
    const { result } = renderHook(() => useNavigationHistory())

    expect(result.current.canGoBack).toBe(false)
    expect(result.current.history).toBeNull()
  })

  it('stores navigation state when pushing', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    expect(result.current.canGoBack).toBe(true)
    expect(result.current.history).toEqual({ view: 'week', scrollPosition: 100, entryId: 42 })
  })

  it('clears history on goBack', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    const returned = act(() => result.current.goBack())

    expect(result.current.canGoBack).toBe(false)
    expect(result.current.history).toBeNull()
  })

  it('returns history state on goBack', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    let returnedHistory: ReturnType<typeof result.current.goBack>
    act(() => {
      returnedHistory = result.current.goBack()
    })

    expect(returnedHistory!).toEqual({ view: 'week', scrollPosition: 100, entryId: 42 })
  })

  it('clearHistory removes history without returning it', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    act(() => {
      result.current.clearHistory()
    })

    expect(result.current.canGoBack).toBe(false)
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run useNavigationHistory.test.ts`
Expected: FAIL - module not found

**Step 3: Write minimal implementation**

Create `frontend/src/hooks/useNavigationHistory.ts`:

```typescript
import { useState, useCallback } from 'react'

export interface NavigationState {
  view: string
  scrollPosition: number
  entryId?: number
}

export function useNavigationHistory() {
  const [history, setHistory] = useState<NavigationState | null>(null)

  const pushHistory = useCallback((state: NavigationState) => {
    setHistory(state)
  }, [])

  const goBack = useCallback(() => {
    const current = history
    setHistory(null)
    return current
  }, [history])

  const clearHistory = useCallback(() => {
    setHistory(null)
  }, [])

  return {
    history,
    canGoBack: history !== null,
    pushHistory,
    goBack,
    clearHistory,
  }
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run useNavigationHistory.test.ts`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/hooks/useNavigationHistory.ts frontend/src/hooks/__tests__/useNavigationHistory.test.ts
git commit -m "feat: add useNavigationHistory hook for view navigation"
```

---

## Task 6: Create EntryTree Component

**Files:**
- Create: `frontend/src/components/bujo/EntryTree.tsx`
- Create: `frontend/src/components/bujo/__tests__/EntryTree.test.tsx`

**Step 1: Write failing test**

Create `frontend/src/components/bujo/__tests__/EntryTree.test.tsx`:

```typescript
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { EntryTree } from '../EntryTree'
import { Entry } from '@/types/bujo'

const mockEntries: Entry[] = [
  { id: 1, content: 'Root event', type: 'event', date: '2026-01-15', priority: 'none', parentId: null, children: [] },
  { id: 2, content: 'Child task', type: 'task', date: '2026-01-15', priority: 'none', parentId: 1, children: [] },
  { id: 3, content: 'Grandchild note', type: 'note', date: '2026-01-15', priority: 'none', parentId: 2, children: [] },
]

describe('EntryTree', () => {
  it('renders tree from root to highlighted entry', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    expect(screen.getByText('Root event')).toBeInTheDocument()
    expect(screen.getByText('Child task')).toBeInTheDocument()
    expect(screen.getByText('Grandchild note')).toBeInTheDocument()
  })

  it('shows bullet journal symbols for each entry type', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    const tree = screen.getByTestId('entry-tree')
    expect(tree.textContent).toContain('o') // event
    expect(tree.textContent).toContain('.') // task
    expect(tree.textContent).toContain('-') // note
  })

  it('highlights the target entry', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    const highlighted = screen.getByTestId('entry-tree-item-3')
    expect(highlighted).toHaveClass('bg-primary/10')
  })

  it('indents nested entries', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    const root = screen.getByTestId('entry-tree-item-1')
    const child = screen.getByTestId('entry-tree-item-2')
    const grandchild = screen.getByTestId('entry-tree-item-3')

    // Check padding-left increases with depth
    expect(root).toHaveStyle({ paddingLeft: '0px' })
    expect(child).toHaveStyle({ paddingLeft: '16px' })
    expect(grandchild).toHaveStyle({ paddingLeft: '32px' })
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run EntryTree.test.tsx`
Expected: FAIL - module not found

**Step 3: Write minimal implementation**

Create `frontend/src/components/bujo/EntryTree.tsx`:

```typescript
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo'
import { cn } from '@/lib/utils'

interface EntryTreeProps {
  entries: Entry[]
  highlightedEntryId: number
  rootEntryId: number
}

const INDENT_PX = 16

export function EntryTree({ entries, highlightedEntryId, rootEntryId }: EntryTreeProps) {
  const entriesById = new Map(entries.map(e => [e.id, e]))

  function buildPath(entryId: number): Entry[] {
    const path: Entry[] = []
    let current = entriesById.get(entryId)
    while (current) {
      path.unshift(current)
      current = current.parentId ? entriesById.get(current.parentId) : undefined
    }
    return path
  }

  const pathToHighlighted = buildPath(highlightedEntryId)
  const rootIndex = pathToHighlighted.findIndex(e => e.id === rootEntryId)
  const visiblePath = rootIndex >= 0 ? pathToHighlighted.slice(rootIndex) : pathToHighlighted

  return (
    <div data-testid="entry-tree" className="space-y-1">
      {visiblePath.map((entry, index) => {
        const isHighlighted = entry.id === highlightedEntryId
        const symbol = ENTRY_SYMBOLS[entry.type] || '-'

        return (
          <div
            key={entry.id}
            data-testid={`entry-tree-item-${entry.id}`}
            style={{ paddingLeft: `${index * INDENT_PX}px` }}
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
  )
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run EntryTree.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/EntryTree.tsx frontend/src/components/bujo/__tests__/EntryTree.test.tsx
git commit -m "feat: add EntryTree component for context visualization"
```

---

## Task 7: Create EntryContextPopover Component

**Files:**
- Create: `frontend/src/components/bujo/EntryContextPopover.tsx`
- Create: `frontend/src/components/bujo/__tests__/EntryContextPopover.test.tsx`

**Step 1: Write failing test**

Create `frontend/src/components/bujo/__tests__/EntryContextPopover.test.tsx`:

```typescript
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { EntryContextPopover } from '../EntryContextPopover'
import { Entry } from '@/types/bujo'

const mockEntries: Entry[] = [
  { id: 1, content: 'Root event', type: 'event', date: '2026-01-15', priority: 'none', parentId: null, children: [] },
  { id: 2, content: 'Child task', type: 'task', date: '2026-01-15', priority: 'none', parentId: 1, children: [] },
]

describe('EntryContextPopover', () => {
  it('renders trigger and opens popover on click', async () => {
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))

    expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
    expect(screen.getByText('Child task')).toBeInTheDocument()
  })

  it('shows quick action buttons based on entry type', async () => {
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))

    expect(screen.getByRole('button', { name: /done/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /priority/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /migrate/i })).toBeInTheDocument()
  })

  it('calls onAction when quick action clicked', async () => {
    const onAction = vi.fn()
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={onAction}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    await userEvent.click(screen.getByRole('button', { name: /done/i }))

    expect(onAction).toHaveBeenCalledWith(mockEntries[1], 'done')
  })

  it('calls onNavigate when "Go to entry" clicked', async () => {
    const onNavigate = vi.fn()
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={onNavigate}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    await userEvent.click(screen.getByText('Go to entry'))

    expect(onNavigate).toHaveBeenCalledWith(mockEntries[1])
  })

  it('closes on Escape key', async () => {
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()

    await userEvent.keyboard('{Escape}')

    expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
  })

  it('supports keyboard shortcuts for actions', async () => {
    const onAction = vi.fn()
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={onAction}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    await userEvent.keyboard(' ') // Space for done

    expect(onAction).toHaveBeenCalledWith(mockEntries[1], 'done')
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run EntryContextPopover.test.tsx`
Expected: FAIL - module not found

**Step 3: Write minimal implementation**

Create `frontend/src/components/bujo/EntryContextPopover.tsx`:

```typescript
import { ReactNode, useState, useCallback, useEffect } from 'react'
import * as Popover from '@radix-ui/react-popover'
import { Entry } from '@/types/bujo'
import { EntryTree } from './EntryTree'
import { cn } from '@/lib/utils'

type ActionType = 'done' | 'cancel' | 'priority' | 'migrate'

interface EntryContextPopoverProps {
  entry: Entry
  entries: Entry[]
  onAction: (entry: Entry, action: ActionType) => void
  onNavigate: (entry: Entry) => void
  children: ReactNode
}

function getAvailableActions(entry: Entry): ActionType[] {
  switch (entry.type) {
    case 'task':
      return ['done', 'priority', 'migrate']
    case 'question':
      return ['done', 'priority']
    case 'done':
      return ['cancel'] // undo
    case 'event':
    case 'note':
      return ['priority']
    default:
      return []
  }
}

function findRootId(entry: Entry, entries: Entry[]): number {
  const entriesById = new Map(entries.map(e => [e.id, e]))
  let current = entry
  while (current.parentId) {
    const parent = entriesById.get(current.parentId)
    if (!parent) break
    current = parent
  }
  return current.id
}

export function EntryContextPopover({
  entry,
  entries,
  onAction,
  onNavigate,
  children,
}: EntryContextPopoverProps) {
  const [open, setOpen] = useState(false)
  const availableActions = getAvailableActions(entry)
  const rootId = findRootId(entry, entries)

  const handleAction = useCallback((action: ActionType) => {
    onAction(entry, action)
    if (action === 'done' || action === 'cancel') {
      setOpen(false)
    }
  }, [entry, onAction])

  const handleNavigate = useCallback(() => {
    onNavigate(entry)
    setOpen(false)
  }, [entry, onNavigate])

  useEffect(() => {
    if (!open) return

    function handleKeyDown(e: KeyboardEvent) {
      switch (e.key) {
        case ' ':
          e.preventDefault()
          if (availableActions.includes('done')) handleAction('done')
          break
        case 'x':
          if (availableActions.includes('cancel')) handleAction('cancel')
          break
        case 'p':
          if (availableActions.includes('priority')) handleAction('priority')
          break
        case 'm':
          if (availableActions.includes('migrate')) handleAction('migrate')
          break
        case 'Enter':
          e.preventDefault()
          handleNavigate()
          break
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [open, availableActions, handleAction, handleNavigate])

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger asChild>
        {children}
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          data-testid="entry-context-popover"
          className="z-50 w-80 max-h-[400px] overflow-auto rounded-lg border border-border bg-card p-3 shadow-lg"
          sideOffset={4}
          collisionPadding={16}
        >
          <EntryTree
            entries={entries}
            highlightedEntryId={entry.id}
            rootEntryId={rootId}
          />

          <div className="mt-3 pt-3 border-t border-border flex items-center justify-between">
            <div className="flex gap-1">
              {availableActions.includes('done') && (
                <button
                  onClick={() => handleAction('done')}
                  aria-label="Mark done"
                  className="p-2 rounded hover:bg-muted"
                  title="Done (Space)"
                >
                  ✓
                </button>
              )}
              {availableActions.includes('cancel') && (
                <button
                  onClick={() => handleAction('cancel')}
                  aria-label="Cancel"
                  className="p-2 rounded hover:bg-muted"
                  title="Cancel (x)"
                >
                  ✕
                </button>
              )}
              {availableActions.includes('priority') && (
                <button
                  onClick={() => handleAction('priority')}
                  aria-label="Cycle priority"
                  className="p-2 rounded hover:bg-muted"
                  title="Priority (p)"
                >
                  !
                </button>
              )}
              {availableActions.includes('migrate') && (
                <button
                  onClick={() => handleAction('migrate')}
                  aria-label="Migrate"
                  className="p-2 rounded hover:bg-muted"
                  title="Migrate (m)"
                >
                  &gt;
                </button>
              )}
            </div>

            <button
              onClick={handleNavigate}
              className="text-sm text-primary hover:underline"
            >
              Go to entry →
            </button>
          </div>

          <Popover.Arrow className="fill-border" />
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  )
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run EntryContextPopover.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/EntryContextPopover.tsx frontend/src/components/bujo/__tests__/EntryContextPopover.test.tsx
git commit -m "feat: add EntryContextPopover component with quick actions"
```

---

## Task 8: Add WeekSummary Entry Click Handler and Popover Integration

**Files:**
- Modify: `frontend/src/components/bujo/WeekSummary.tsx`
- Modify: `frontend/src/components/bujo/__tests__/WeekSummary.test.tsx`

**Step 1: Write failing test**

Add to `frontend/src/components/bujo/__tests__/WeekSummary.test.tsx`:

```typescript
describe('entry interaction', () => {
  it('opens context popover when attention item clicked', async () => {
    const mockDays = [
      {
        date: '2026-01-15',
        entries: [
          { id: 1, content: 'Test task', type: 'task' as const, date: '2026-01-15', priority: 'high' as const, parentId: null, children: [] }
        ]
      }
    ]

    render(
      <WeekSummary
        days={mockDays}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
      />
    )

    await userEvent.click(screen.getByText('Test task'))

    expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
  })

  it('calls onAction when popover action triggered', async () => {
    const mockDays = [
      {
        date: '2026-01-15',
        entries: [
          { id: 1, content: 'Test task', type: 'task' as const, date: '2026-01-15', priority: 'none' as const, parentId: null, children: [] }
        ]
      }
    ]
    const onAction = vi.fn()

    render(
      <WeekSummary
        days={mockDays}
        onAction={onAction}
        onNavigate={vi.fn()}
      />
    )

    await userEvent.click(screen.getByText('Test task'))
    await userEvent.click(screen.getByRole('button', { name: /done/i }))

    expect(onAction).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }), 'done')
  })

  it('shows "Show all" button that opens all attention items', async () => {
    const mockDays = [
      {
        date: '2026-01-15',
        entries: Array.from({ length: 10 }, (_, i) => ({
          id: i + 1,
          content: `Task ${i + 1}`,
          type: 'task' as const,
          date: '2026-01-15',
          priority: 'none' as const,
          parentId: null,
          children: []
        }))
      }
    ]
    const onShowAll = vi.fn()

    render(
      <WeekSummary
        days={mockDays}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
        onShowAllAttention={onShowAll}
      />
    )

    await userEvent.click(screen.getByText('Show all'))

    expect(onShowAll).toHaveBeenCalled()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run WeekSummary.test.tsx`
Expected: FAIL - props don't exist, popover doesn't open

**Step 3: Write minimal implementation**

Update `WeekSummary.tsx`:

```typescript
import { DayEntries, Entry } from '@/types/bujo';
import { calculateAttentionScore, sortByAttentionScore } from '@/lib/attentionScore';
import { cn } from '@/lib/utils';
import { EntryContextPopover } from './EntryContextPopover';

type ActionType = 'done' | 'cancel' | 'priority' | 'migrate'

interface WeekSummaryProps {
  days: DayEntries[];
  onAction?: (entry: Entry, action: ActionType) => void;
  onNavigate?: (entry: Entry) => void;
  onShowAllAttention?: () => void;
}

// ... rest of implementation

// In the attention items section, wrap each item:
{topAttentionEntries.map(entry => {
  const { indicators } = calculateAttentionScore(entry, now);
  const isHighPriority = entry.priority === 'high';
  return (
    <EntryContextPopover
      key={entry.id}
      entry={entry}
      entries={allEntries}
      onAction={onAction || (() => {})}
      onNavigate={onNavigate || (() => {})}
    >
      <div
        data-testid="attention-item"
        data-attention-item
        data-priority={isHighPriority ? 'high' : undefined}
        className="flex items-center justify-between p-2 rounded-lg border border-border bg-card cursor-pointer hover:bg-muted/50"
      >
        {/* ... existing content ... */}
      </div>
    </EntryContextPopover>
  );
})}

// Update "Show all" button:
{hasMoreThanLimit && (
  <button
    onClick={onShowAllAttention}
    className="text-sm text-primary hover:underline"
  >
    Show all
  </button>
)}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run WeekSummary.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/WeekSummary.tsx frontend/src/components/bujo/__tests__/WeekSummary.test.tsx
git commit -m "feat: add entry click handling and popover to WeekSummary"
```

---

## Task 9: Add Back Button to Header

**Files:**
- Modify: `frontend/src/components/Header.tsx` (or equivalent)
- Create/Modify: `frontend/src/components/__tests__/Header.test.tsx`

**Step 1: Write failing test**

```typescript
describe('back button', () => {
  it('does not render back button when canGoBack is false', () => {
    render(<Header canGoBack={false} onBack={vi.fn()} />)

    expect(screen.queryByRole('button', { name: /back/i })).not.toBeInTheDocument()
  })

  it('renders back button when canGoBack is true', () => {
    render(<Header canGoBack={true} onBack={vi.fn()} />)

    expect(screen.getByRole('button', { name: /back/i })).toBeInTheDocument()
  })

  it('calls onBack when back button clicked', async () => {
    const onBack = vi.fn()
    render(<Header canGoBack={true} onBack={onBack} />)

    await userEvent.click(screen.getByRole('button', { name: /back/i }))

    expect(onBack).toHaveBeenCalled()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run Header.test.tsx`
Expected: FAIL - props don't exist

**Step 3: Write minimal implementation**

Add to Header component:

```typescript
interface HeaderProps {
  // ... existing props
  canGoBack?: boolean
  onBack?: () => void
}

// In JSX, add conditionally:
{canGoBack && onBack && (
  <button
    onClick={onBack}
    aria-label="Go back"
    className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
  >
    ← Back
  </button>
)}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run Header.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/components/Header.tsx frontend/src/components/__tests__/Header.test.tsx
git commit -m "feat: add back button to header for navigation history"
```

---

## Task 10: Integrate Navigation History in App

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/App.test.tsx` (if exists)

**Step 1: Write failing test**

```typescript
describe('navigation history', () => {
  it('shows back button after navigating from popover', async () => {
    render(<App />)

    // Navigate to week view, click an entry, click "Go to entry"
    // ... setup steps

    expect(screen.getByRole('button', { name: /back/i })).toBeInTheDocument()
  })

  it('returns to previous view when back clicked', async () => {
    render(<App />)

    // ... setup and navigate

    await userEvent.click(screen.getByRole('button', { name: /back/i }))

    // Should be back at week view
    expect(screen.getByTestId('week-summary')).toBeInTheDocument()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run App.test.tsx`
Expected: FAIL - navigation history not integrated

**Step 3: Write minimal implementation**

In `App.tsx`:

```typescript
import { useNavigationHistory } from '@/hooks/useNavigationHistory'

function App() {
  const { canGoBack, pushHistory, goBack, clearHistory } = useNavigationHistory()

  const handleNavigateToEntry = useCallback((entry: Entry) => {
    pushHistory({
      view: currentView,
      scrollPosition: window.scrollY,
      entryId: entry.id
    })
    // Navigate to journal view for entry's date
    setCurrentDate(new Date(entry.date))
    setCurrentView('journal')
    // Scroll to entry after render
    setTimeout(() => {
      document.getElementById(`entry-${entry.id}`)?.scrollIntoView({ behavior: 'smooth' })
    }, 100)
  }, [currentView, pushHistory])

  const handleBack = useCallback(() => {
    const history = goBack()
    if (history) {
      setCurrentView(history.view)
      setTimeout(() => window.scrollTo(0, history.scrollPosition), 100)
    }
  }, [goBack])

  // Clear history on manual navigation
  const handleViewChange = useCallback((view: string) => {
    clearHistory()
    setCurrentView(view)
  }, [clearHistory])

  return (
    <>
      <Header canGoBack={canGoBack} onBack={handleBack} />
      {/* Pass onNavigate to views */}
    </>
  )
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test -- --run App.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/App.tsx frontend/src/App.test.tsx
git commit -m "feat: integrate navigation history with back button support"
```

---

## Task 11: Add Global Migrate Keyboard Shortcut

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: Keyboard shortcut tests

**Step 1: Write failing test**

```typescript
describe('keyboard shortcuts', () => {
  it('triggers migrate on "m" key when entry selected', async () => {
    const onMigrate = vi.fn()
    render(<App onMigrate={onMigrate} />)

    // Select an entry
    await userEvent.keyboard('j') // move to first entry
    await userEvent.keyboard('m')

    expect(onMigrate).toHaveBeenCalled()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test`
Expected: FAIL - 'm' shortcut not handled

**Step 3: Write minimal implementation**

Add to App.tsx keyboard handler:

```typescript
case 'm':
  if (selectedEntry && selectedEntry.type === 'task') {
    handleMigrate(selectedEntry)
  }
  break
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && npm test`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat: add global 'm' keyboard shortcut for migrate"
```

---

## Task 12: Integration Testing

**Files:**
- Create: `frontend/src/components/bujo/__tests__/integration/contextPopover.test.tsx`

**Step 1: Write integration tests**

```typescript
describe('Context Popover Integration', () => {
  it('full flow: click entry -> see context -> take action -> entry updates', async () => {
    // Render full app with mock data
    // Click attention item in WeekSummary
    // Verify popover shows tree context
    // Click "Done" action
    // Verify entry marked as done
    // Verify popover closes
    // Verify attention list updates
  })

  it('full flow: click entry -> navigate -> back button returns', async () => {
    // Render full app
    // Click attention item
    // Click "Go to entry"
    // Verify navigated to journal view
    // Verify back button visible
    // Click back
    // Verify returned to week view
  })
})
```

**Step 2: Run integration tests**

Run: `cd frontend && npm test -- --run integration`
Expected: PASS

**Step 3: Commit**

```bash
git add frontend/src/components/bujo/__tests__/integration/
git commit -m "test: add integration tests for context popover flow"
```

---

## Task 13: Final Verification

**Step 1: Run all tests**

Run: `cd frontend && npm test`
Expected: All tests pass

**Step 2: Run linting**

Run: `cd frontend && npm run lint`
Expected: No errors

**Step 3: Run type checking**

Run: `cd frontend && npm run typecheck`
Expected: No errors

**Step 4: Manual smoke test**

1. Open app
2. Navigate to Week view
3. Click attention item - verify popover opens with tree
4. Press Space - verify entry marked done
5. Click another item, click "Go to entry"
6. Verify navigated to journal, back button visible
7. Click back - verify returned to week view
8. Test CaptureBar: type ". hello" - verify task created
9. Test CaptureBar: verify symbol buttons (. - o ?)
10. Test multiline in CaptureBar - verify auto-grow

**Step 5: Create final commit if any cleanup needed**

```bash
git add -A
git commit -m "chore: final cleanup and polish"
```

---

## Summary

This plan covers:

1. **Bug Fixes:**
   - CaptureBar type buttons show symbols (Task 2)
   - CaptureBar textarea auto-grow (Task 3)
   - Prefix detection bug fix (Task 4)

2. **New Features:**
   - useNavigationHistory hook (Task 5)
   - EntryTree component (Task 6)
   - EntryContextPopover component (Task 7)
   - WeekSummary popover integration (Task 8)
   - Back button in header (Task 9)
   - App navigation history integration (Task 10)
   - Global 'm' migrate shortcut (Task 11)

3. **Testing:**
   - Integration tests (Task 12)
   - Final verification (Task 13)

Each task follows TDD: write failing test, implement, verify pass, commit.
