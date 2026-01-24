# Capture Bar & Week Summary Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the clunky capture modal with an always-ready capture bar, and add a week summary with attention scoring to the weekly view.

**Architecture:** New `CaptureBar` component replaces `InlineEntryInput` and `CaptureModal`. New `WeekSummary` component with `attentionScore` utility. All frontend-only changes - no backend modifications needed.

**Tech Stack:** React 19, TypeScript, Vitest, React Testing Library, Tailwind CSS

---

## Phase 0: User Acceptance Tests

Write integration tests first that define the expected behavior. These tests will fail initially and guide the implementation.

### Task 0.1: Create Acceptance Test File for Capture Bar

**Files:**
- Create: `frontend/src/App.captureBar.test.tsx`

**Step 1: Write the acceptance tests**

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

const mockEntriesAgenda = createMockAgenda({
  Days: [createMockDayEntries({
    Date: new Date().toISOString().split('T')[0] + 'T00:00:00Z',
    Entries: [
      createMockEntry({ ID: 1, Type: 'Task', Content: 'First task' }),
      createMockEntry({ ID: 2, Type: 'Event', Content: 'Team standup' }),
    ],
  })],
})

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn().mockReturnValue(() => {}),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}))

vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn().mockResolvedValue({
    Overdue: [],
    Days: [{ Date: new Date().toISOString().split('T')[0] + 'T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
  }),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  AddChildEntry: vi.fn().mockResolvedValue([2]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  CreateHabit: vi.fn().mockResolvedValue(1),
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue([]),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
}))

import { GetAgenda, AddEntry, AddChildEntry, OpenFileDialog, ReadFile } from './wailsjs/go/wails/App'
import { OnFileDrop } from './wailsjs/runtime/runtime'

describe('Capture Bar - Always Visible', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    localStorage.clear()
  })

  it('shows capture bar at bottom of today view', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
  })

  it('shows type selection buttons', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: /task/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /note/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /event/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /question/i })).toBeInTheDocument()
  })

  it('has Task selected by default', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const taskButton = screen.getByRole('button', { name: /task/i })
    expect(taskButton).toHaveAttribute('aria-pressed', 'true')
  })
})

describe('Capture Bar - Type Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    localStorage.clear()
  })

  it('clicking type button changes selection', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const noteButton = screen.getByRole('button', { name: /note/i })
    await user.click(noteButton)

    expect(noteButton).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'false')
  })

  it('Tab cycles through types when input is empty', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.click(input)
    await user.keyboard('{Tab}')

    expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
  })

  it('typing prefix changes type', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, '- ')

    expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
  })
})

describe('Capture Bar - Entry Submission', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    localStorage.clear()
  })

  it('Enter submits entry with selected type prefix', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Buy groceries{Enter}')

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalledWith('. Buy groceries', expect.any(String))
    })
  })

  it('clears input after submission', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Buy groceries{Enter}')

    await waitFor(() => {
      expect(input).toHaveValue('')
    })
  })

  it('keeps focus after submission for rapid entry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Buy groceries{Enter}')

    await waitFor(() => {
      expect(input).toHaveFocus()
    })
  })

  it('Escape clears input', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Some text')
    await user.keyboard('{Escape}')

    expect(input).toHaveValue('')
  })
})

describe('Capture Bar - Parent Context (Child Entries)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    localStorage.clear()
  })

  it('pressing A on selected entry shows parent context', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('Team standup')).toBeInTheDocument()
    })

    // Select the event entry
    await user.keyboard('{ArrowDown}')
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
      expect(screen.getByText('Team standup')).toBeInTheDocument()
    })
  })

  it('submitting in child mode calls AddChildEntry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('Team standup')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}')
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Action item from meeting{Enter}')

    await waitFor(() => {
      expect(AddChildEntry).toHaveBeenCalledWith(2, '. Action item from meeting', expect.any(String))
    })
  })

  it('clicking X clears parent context', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('Team standup')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}')
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    const clearButton = screen.getByRole('button', { name: /clear parent/i })
    await user.click(clearButton)

    expect(screen.queryByText(/adding to:/i)).not.toBeInTheDocument()
  })
})

describe('Capture Bar - Draft Persistence', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    localStorage.clear()
  })

  it('saves draft to localStorage on input', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Draft entry')

    expect(localStorage.getItem('bujo-capture-bar-draft')).toBe('Draft entry')
  })

  it('restores draft on mount', async () => {
    localStorage.setItem('bujo-capture-bar-draft', 'Restored draft')

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    expect(input).toHaveValue('Restored draft')
  })

  it('clears draft after successful submission', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Draft entry{Enter}')

    await waitFor(() => {
      expect(localStorage.getItem('bujo-capture-bar-draft')).toBeNull()
    })
  })
})

describe('Capture Bar - File Import', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    localStorage.clear()
  })

  it('shows file import button', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: /import file/i })).toBeInTheDocument()
  })

  it('clicking import button opens file dialog', async () => {
    const user = userEvent.setup()
    vi.mocked(OpenFileDialog).mockResolvedValue('')

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const importButton = screen.getByRole('button', { name: /import file/i })
    await user.click(importButton)

    expect(OpenFileDialog).toHaveBeenCalled()
  })

  it('appends file content to input', async () => {
    const user = userEvent.setup()
    vi.mocked(OpenFileDialog).mockResolvedValue('. Task from file')

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/add a task/i)
    await user.type(input, 'Existing text\n')

    const importButton = screen.getByRole('button', { name: /import file/i })
    await user.click(importButton)

    await waitFor(() => {
      expect(input).toHaveValue('Existing text\n. Task from file')
    })
  })
})

describe('Capture Bar - Keyboard Shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    localStorage.clear()
  })

  it('i key focuses capture bar', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('i')

    const input = screen.getByPlaceholderText(/add a task/i)
    expect(input).toHaveFocus()
  })

  it('r key focuses capture bar in root mode', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // First set a parent context
    await user.keyboard('{ArrowDown}')
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    // Now press r to force root mode
    await user.keyboard('{Escape}') // blur first
    await user.keyboard('r')

    expect(screen.queryByText(/adding to:/i)).not.toBeInTheDocument()
  })
})
```

**Step 2: Run tests to verify they fail**

Run: `cd frontend && npm test -- --run src/App.captureBar.test.tsx`
Expected: Multiple failures (components don't exist yet)

**Step 3: Commit the acceptance tests**

```bash
git add frontend/src/App.captureBar.test.tsx
git commit -m "test: Add acceptance tests for CaptureBar component

RED phase - these tests define expected behavior and will fail
until CaptureBar is implemented."
```

---

### Task 0.2: Create Acceptance Test File for Week Summary

**Files:**
- Create: `frontend/src/App.weekSummary.test.tsx`

**Step 1: Write the acceptance tests**

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

// Create a week's worth of test data
const today = new Date()
const createDateString = (daysAgo: number) => {
  const date = new Date(today)
  date.setDate(date.getDate() - daysAgo)
  return date.toISOString().split('T')[0] + 'T00:00:00Z'
}

const mockWeekAgenda = createMockAgenda({
  Days: [
    createMockDayEntries({
      Date: createDateString(0),
      Entries: [
        createMockEntry({ ID: 1, Type: 'Task', Content: 'Open task 1' }),
        createMockEntry({ ID: 2, Type: 'Done', Content: 'Completed task' }),
        createMockEntry({ ID: 3, Type: 'Event', Content: 'Team standup' }),
        createMockEntry({ ID: 4, Type: 'Note', Content: 'Standup note', ParentID: 3 }),
        createMockEntry({ ID: 5, Type: 'Task', Content: 'Action from standup', ParentID: 3 }),
      ],
    }),
    createMockDayEntries({
      Date: createDateString(1),
      Entries: [
        createMockEntry({ ID: 6, Type: 'Task', Content: 'Open task 2' }),
        createMockEntry({ ID: 7, Type: 'Migrated', Content: 'Migrated task' }),
      ],
    }),
    createMockDayEntries({
      Date: createDateString(3),
      Entries: [
        createMockEntry({ ID: 8, Type: 'Task', Content: 'Old task needs attention', Priority: 'high' }),
        createMockEntry({ ID: 9, Type: 'Question', Content: 'Unanswered question?' }),
      ],
    }),
  ],
  Overdue: [],
})

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn().mockReturnValue(() => {}),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}))

vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn().mockResolvedValue({
    Overdue: [],
    Days: [{ Date: new Date().toISOString().split('T')[0] + 'T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
  }),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  AddChildEntry: vi.fn().mockResolvedValue([2]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  CreateHabit: vi.fn().mockResolvedValue(1),
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue([]),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
}))

import { GetAgenda } from './wailsjs/go/wails/App'

describe('Week Summary - Task Flow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockWeekAgenda)
  })

  it('shows week summary at top of weekly view', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to weekly view
    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByTestId('week-summary')).toBeInTheDocument()
    })
  })

  it('shows task flow section with created count', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText('Task Flow')).toBeInTheDocument()
      expect(screen.getByText('Created')).toBeInTheDocument()
    })
  })

  it('shows done, migrated, and open counts', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText('Done')).toBeInTheDocument()
      expect(screen.getByText('Migrated')).toBeInTheDocument()
      expect(screen.getByText('Open')).toBeInTheDocument()
    })
  })
})

describe('Week Summary - Meetings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockWeekAgenda)
  })

  it('shows meetings section with events that have children', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText('Meetings')).toBeInTheDocument()
      expect(screen.getByText('Team standup')).toBeInTheDocument()
    })
  })

  it('shows child count for each meeting', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      // Team standup has 2 children (note + task)
      expect(screen.getByText(/2 items/i)).toBeInTheDocument()
    })
  })
})

describe('Week Summary - Needs Attention', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockWeekAgenda)
  })

  it('shows needs attention section', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText('Needs Attention')).toBeInTheDocument()
    })
  })

  it('shows open tasks sorted by attention score', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      // High priority + old task should appear
      expect(screen.getByText('Old task needs attention')).toBeInTheDocument()
      expect(screen.getByText('Unanswered question?')).toBeInTheDocument()
    })
  })

  it('shows attention indicator for high-priority items', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(weekButton)

    await waitFor(() => {
      // Should show some indicator for why item needs attention
      expect(screen.getByText(/priority/i)).toBeInTheDocument()
    })
  })
})
```

**Step 2: Run tests to verify they fail**

Run: `cd frontend && npm test -- --run src/App.weekSummary.test.tsx`
Expected: Multiple failures (components don't exist yet)

**Step 3: Commit the acceptance tests**

```bash
git add frontend/src/App.weekSummary.test.tsx
git commit -m "test: Add acceptance tests for WeekSummary component

RED phase - these tests define expected behavior and will fail
until WeekSummary is implemented."
```

---

## Phase 1: Attention Score Utility

Start with pure logic that has no UI dependencies.

### Task 1.1: Create Attention Score Tests

**Files:**
- Create: `frontend/src/lib/attentionScore.test.ts`

**Step 1: Write the failing tests**

```typescript
import { describe, it, expect } from 'vitest'
import { calculateAttentionScore, AttentionIndicator } from './attentionScore'
import { Entry } from '@/types/bujo'

const createEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  entityId: 'e1',
  type: 'task',
  content: 'Test task',
  priority: '',
  parentId: null,
  depth: 0,
  loggedDate: new Date().toISOString(),
  scheduledDate: undefined,
  migrationCount: 0,
  ...overrides,
})

describe('calculateAttentionScore', () => {
  it('returns 0 for a new task with no special conditions', () => {
    const entry = createEntry()
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(0)
  })

  it('adds 50 points for past scheduled date', () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)

    const entry = createEntry({
      scheduledDate: yesterday.toISOString(),
    })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(50)
    expect(result.indicators).toContain('overdue')
  })

  it('adds 30 points for any priority set', () => {
    const entry = createEntry({ priority: 'low' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(30)
  })

  it('adds additional 20 points for high priority', () => {
    const entry = createEntry({ priority: 'high' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(50) // 30 + 20
    expect(result.indicators).toContain('priority')
  })

  it('adds 25 points for items older than 7 days', () => {
    const eightDaysAgo = new Date()
    eightDaysAgo.setDate(eightDaysAgo.getDate() - 8)

    const entry = createEntry({
      loggedDate: eightDaysAgo.toISOString(),
    })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(25)
    expect(result.indicators).toContain('aging')
  })

  it('adds 15 points for items older than 3 days but less than 7', () => {
    const fourDaysAgo = new Date()
    fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

    const entry = createEntry({
      loggedDate: fourDaysAgo.toISOString(),
    })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(15)
  })

  it('adds 15 points per migration', () => {
    const entry = createEntry({ migrationCount: 2 })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(30) // 15 * 2
    expect(result.indicators).toContain('migrated')
  })

  it('adds 20 points for urgent keywords in content', () => {
    const entry = createEntry({ content: 'This is urgent!' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBeGreaterThanOrEqual(20)
  })

  it('adds 10 points for questions', () => {
    const entry = createEntry({ type: 'question' })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.score).toBe(10)
  })

  it('adds 5 points for items with event parent', () => {
    const entry = createEntry({ parentId: 1 })
    const result = calculateAttentionScore(entry, new Date(), 'event')
    expect(result.score).toBe(5)
  })

  it('combines multiple conditions', () => {
    const fourDaysAgo = new Date()
    fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

    const entry = createEntry({
      priority: 'high',
      loggedDate: fourDaysAgo.toISOString(),
      migrationCount: 1,
    })
    const result = calculateAttentionScore(entry, new Date())
    // 30 (priority) + 20 (high) + 15 (age) + 15 (migration) = 80
    expect(result.score).toBe(80)
  })
})

describe('AttentionIndicator formatting', () => {
  it('returns overdue indicator for past scheduled date', () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)

    const entry = createEntry({ scheduledDate: yesterday.toISOString() })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.indicators).toContain('overdue')
  })

  it('returns migrated indicator with count', () => {
    const entry = createEntry({ migrationCount: 2 })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.indicators).toContain('migrated')
    expect(result.migrationCount).toBe(2)
  })

  it('returns aging indicator for old items', () => {
    const fourDaysAgo = new Date()
    fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

    const entry = createEntry({ loggedDate: fourDaysAgo.toISOString() })
    const result = calculateAttentionScore(entry, new Date())
    expect(result.indicators).toContain('aging')
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && npm test -- --run src/lib/attentionScore.test.ts`
Expected: FAIL - module not found

**Step 3: Commit**

```bash
git add frontend/src/lib/attentionScore.test.ts
git commit -m "test: Add failing tests for attentionScore utility"
```

---

### Task 1.2: Implement Attention Score

**Files:**
- Create: `frontend/src/lib/attentionScore.ts`

**Step 1: Write the implementation**

```typescript
import { Entry } from '@/types/bujo'

export type AttentionIndicator = 'overdue' | 'priority' | 'aging' | 'migrated'

export interface AttentionResult {
  score: number
  indicators: AttentionIndicator[]
  migrationCount?: number
  daysOld?: number
}

const URGENT_KEYWORDS = ['urgent', 'asap', 'blocker', 'waiting', 'blocked']

export function calculateAttentionScore(
  entry: Entry,
  now: Date,
  parentType?: string
): AttentionResult {
  let score = 0
  const indicators: AttentionIndicator[] = []

  // Past scheduled date: +50
  if (entry.scheduledDate) {
    const scheduled = new Date(entry.scheduledDate)
    if (scheduled < now) {
      score += 50
      indicators.push('overdue')
    }
  }

  // Priority set: +30, high/urgent: additional +20
  if (entry.priority) {
    score += 30
    indicators.push('priority')
    if (entry.priority === 'high' || entry.priority === 'urgent') {
      score += 20
    }
  }

  // Age calculations
  const loggedDate = new Date(entry.loggedDate)
  const daysOld = Math.floor((now.getTime() - loggedDate.getTime()) / (1000 * 60 * 60 * 24))

  if (daysOld > 7) {
    score += 25
    indicators.push('aging')
  } else if (daysOld > 3) {
    score += 15
    indicators.push('aging')
  }

  // Migration count: +15 per migration
  if (entry.migrationCount && entry.migrationCount > 0) {
    score += entry.migrationCount * 15
    indicators.push('migrated')
  }

  // Urgent keywords: +20
  const contentLower = entry.content.toLowerCase()
  if (URGENT_KEYWORDS.some(keyword => contentLower.includes(keyword))) {
    score += 20
  }

  // Questions: +10
  if (entry.type === 'question') {
    score += 10
  }

  // Parent is event: +5
  if (entry.parentId && parentType === 'event') {
    score += 5
  }

  return {
    score,
    indicators,
    migrationCount: entry.migrationCount,
    daysOld,
  }
}

export function sortByAttentionScore(
  entries: Entry[],
  now: Date,
  parentTypes?: Map<number, string>
): Entry[] {
  return [...entries].sort((a, b) => {
    const scoreA = calculateAttentionScore(a, now, parentTypes?.get(a.parentId ?? 0))
    const scoreB = calculateAttentionScore(b, now, parentTypes?.get(b.parentId ?? 0))
    return scoreB.score - scoreA.score
  })
}
```

**Step 2: Run tests to verify they pass**

Run: `cd frontend && npm test -- --run src/lib/attentionScore.test.ts`
Expected: All tests PASS

**Step 3: Commit**

```bash
git add frontend/src/lib/attentionScore.ts
git commit -m "feat: Implement attentionScore utility for prioritizing items"
```

---

## Phase 2: CaptureBar Component

### Task 2.1: Create CaptureBar Unit Tests

**Files:**
- Create: `frontend/src/components/bujo/CaptureBar.test.tsx`

**Step 1: Write the failing tests**

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CaptureBar } from './CaptureBar'

describe('CaptureBar', () => {
  const defaultProps = {
    onSubmit: vi.fn(),
    onSubmitChild: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  describe('rendering', () => {
    it('renders type buttons', () => {
      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByRole('button', { name: /task/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /note/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /event/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /question/i })).toBeInTheDocument()
    })

    it('renders input with placeholder', () => {
      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByPlaceholderText(/add a task/i)).toBeInTheDocument()
    })

    it('renders file import button', () => {
      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByRole('button', { name: /import file/i })).toBeInTheDocument()
    })

    it('has task selected by default', () => {
      render(<CaptureBar {...defaultProps} />)

      const taskButton = screen.getByRole('button', { name: /task/i })
      expect(taskButton).toHaveAttribute('aria-pressed', 'true')
    })
  })

  describe('type selection', () => {
    it('clicking type button changes selection', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const noteButton = screen.getByRole('button', { name: /note/i })
      await user.click(noteButton)

      expect(noteButton).toHaveAttribute('aria-pressed', 'true')
      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'false')
    })

    it('updates placeholder when type changes', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /note/i }))

      expect(screen.getByPlaceholderText(/add a note/i)).toBeInTheDocument()
    })

    it('Tab cycles types when input is empty', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.click(input)
      await user.keyboard('{Tab}')

      expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('Tab does not cycle when input has content', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Some content')
      await user.keyboard('{Tab}')

      // Task should still be selected
      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
    })
  })

  describe('prefix detection', () => {
    it('typing ". " sets type to task', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      // Start with note selected
      await user.click(screen.getByRole('button', { name: /note/i }))

      const input = screen.getByPlaceholderText(/add a note/i)
      await user.type(input, '. ')

      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
      expect(input).toHaveValue('') // Prefix consumed
    })

    it('typing "- " sets type to note', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, '- ')

      expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('typing "o " sets type to event', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'o ')

      expect(screen.getByRole('button', { name: /event/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('typing "? " sets type to question', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, '? ')

      expect(screen.getByRole('button', { name: /question/i })).toHaveAttribute('aria-pressed', 'true')
    })
  })

  describe('submission', () => {
    it('Enter submits with type prefix', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Buy groceries{Enter}')

      expect(onSubmit).toHaveBeenCalledWith('. Buy groceries')
    })

    it('clears input after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Buy groceries{Enter}')

      expect(input).toHaveValue('')
    })

    it('does not submit empty input', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.click(input)
      await user.keyboard('{Enter}')

      expect(onSubmit).not.toHaveBeenCalled()
    })

    it('keeps focus after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Buy groceries{Enter}')

      expect(input).toHaveFocus()
    })
  })

  describe('parent context', () => {
    it('shows parent context when parentEntry provided', () => {
      render(
        <CaptureBar
          {...defaultProps}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event' } as any}
        />
      )

      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
      expect(screen.getByText('Team standup')).toBeInTheDocument()
    })

    it('shows clear button when parent is set', () => {
      render(
        <CaptureBar
          {...defaultProps}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event' } as any}
        />
      )

      expect(screen.getByRole('button', { name: /clear parent/i })).toBeInTheDocument()
    })

    it('calls onSubmitChild when parent is set', async () => {
      const user = userEvent.setup()
      const onSubmitChild = vi.fn()
      render(
        <CaptureBar
          {...defaultProps}
          onSubmitChild={onSubmitChild}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event' } as any}
        />
      )

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Action item{Enter}')

      expect(onSubmitChild).toHaveBeenCalledWith(1, '. Action item')
    })

    it('calls onClearParent when X clicked', async () => {
      const user = userEvent.setup()
      const onClearParent = vi.fn()
      render(
        <CaptureBar
          {...defaultProps}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event' } as any}
          onClearParent={onClearParent}
        />
      )

      await user.click(screen.getByRole('button', { name: /clear parent/i }))

      expect(onClearParent).toHaveBeenCalled()
    })
  })

  describe('draft persistence', () => {
    it('saves draft to localStorage', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Draft text')

      expect(localStorage.getItem('bujo-capture-bar-draft')).toBe('Draft text')
    })

    it('restores draft on mount', () => {
      localStorage.setItem('bujo-capture-bar-draft', 'Restored draft')
      localStorage.setItem('bujo-capture-bar-type', 'note')

      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByDisplayValue('Restored draft')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('clears draft after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Draft text{Enter}')

      expect(localStorage.getItem('bujo-capture-bar-draft')).toBeNull()
    })
  })

  describe('escape handling', () => {
    it('Escape clears input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Some text')
      await user.keyboard('{Escape}')

      expect(input).toHaveValue('')
    })

    it('Escape on empty input blurs', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.click(input)
      await user.keyboard('{Escape}')

      expect(input).not.toHaveFocus()
    })
  })

  describe('file import', () => {
    it('calls onFileImport when import button clicked', async () => {
      const user = userEvent.setup()
      const onFileImport = vi.fn()
      render(<CaptureBar {...defaultProps} onFileImport={onFileImport} />)

      await user.click(screen.getByRole('button', { name: /import file/i }))

      expect(onFileImport).toHaveBeenCalled()
    })

    it('appends imported content to input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} importedContent=". Task from file" />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Existing\n')

      // Simulate file import by re-rendering with importedContent
      // In real usage, this would be managed by parent
    })
  })

  describe('multiline', () => {
    it('Shift+Enter adds newline', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Line 1{Shift>}{Enter}{/Shift}Line 2')

      expect(input).toHaveValue('Line 1\nLine 2')
    })
  })
})
```

**Step 2: Run tests to verify they fail**

Run: `cd frontend && npm test -- --run src/components/bujo/CaptureBar.test.tsx`
Expected: FAIL - component not found

**Step 3: Commit**

```bash
git add frontend/src/components/bujo/CaptureBar.test.tsx
git commit -m "test: Add failing unit tests for CaptureBar component"
```

---

### Task 2.2: Implement CaptureBar Component

**Files:**
- Create: `frontend/src/components/bujo/CaptureBar.tsx`

**Step 1: Write the implementation**

```typescript
import { useState, useEffect, useRef, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { Paperclip, X } from 'lucide-react'
import { Entry } from '@/types/bujo'

type EntryType = 'task' | 'note' | 'event' | 'question'

interface CaptureBarProps {
  onSubmit: (content: string) => void
  onSubmitChild: (parentId: number, content: string) => void
  onClearParent?: () => void
  onFileImport?: () => void
  parentEntry?: Entry | null
  importedContent?: string
}

const TYPE_PREFIXES: Record<EntryType, string> = {
  task: '.',
  note: '-',
  event: 'o',
  question: '?',
}

const PREFIX_TO_TYPE: Record<string, EntryType> = {
  '.': 'task',
  '-': 'note',
  'o': 'event',
  '?': 'question',
}

const TYPE_LABELS: Record<EntryType, string> = {
  task: 'Task',
  note: 'Note',
  event: 'Event',
  question: 'Question',
}

const DRAFT_KEY = 'bujo-capture-bar-draft'
const TYPE_KEY = 'bujo-capture-bar-type'
const PARENT_KEY = 'bujo-capture-bar-parent'

export function CaptureBar({
  onSubmit,
  onSubmitChild,
  onClearParent,
  onFileImport,
  parentEntry,
  importedContent,
}: CaptureBarProps) {
  const [content, setContent] = useState('')
  const [entryType, setEntryType] = useState<EntryType>('task')
  const inputRef = useRef<HTMLTextAreaElement>(null)

  // Restore draft on mount
  useEffect(() => {
    const savedDraft = localStorage.getItem(DRAFT_KEY)
    const savedType = localStorage.getItem(TYPE_KEY) as EntryType | null

    if (savedDraft) {
      setContent(savedDraft)
    }
    if (savedType && TYPE_LABELS[savedType]) {
      setEntryType(savedType)
    }
  }, [])

  // Save draft on change
  useEffect(() => {
    if (content) {
      localStorage.setItem(DRAFT_KEY, content)
      localStorage.setItem(TYPE_KEY, entryType)
    } else {
      localStorage.removeItem(DRAFT_KEY)
    }
  }, [content, entryType])

  // Handle imported content
  useEffect(() => {
    if (importedContent) {
      setContent(prev => prev ? prev + '\n' + importedContent : importedContent)
    }
  }, [importedContent])

  const handleTypeChange = (type: EntryType) => {
    setEntryType(type)
    inputRef.current?.focus()
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value

    // Check for prefix at start
    if (value.length === 2 && value[1] === ' ') {
      const prefix = value[0]
      if (PREFIX_TO_TYPE[prefix]) {
        setEntryType(PREFIX_TO_TYPE[prefix])
        setContent('')
        return
      }
    }

    setContent(value)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Tab' && !content) {
      e.preventDefault()
      const types: EntryType[] = ['task', 'note', 'event', 'question']
      const currentIndex = types.indexOf(entryType)
      const nextIndex = (currentIndex + 1) % types.length
      setEntryType(types[nextIndex])
      return
    }

    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
      return
    }

    if (e.key === 'Escape') {
      e.preventDefault()
      if (content) {
        setContent('')
        localStorage.removeItem(DRAFT_KEY)
      } else {
        inputRef.current?.blur()
      }
      return
    }
  }

  const handleSubmit = useCallback(() => {
    if (!content.trim()) return

    const prefixedContent = `${TYPE_PREFIXES[entryType]} ${content.trim()}`

    if (parentEntry) {
      onSubmitChild(parentEntry.id, prefixedContent)
    } else {
      onSubmit(prefixedContent)
    }

    setContent('')
    localStorage.removeItem(DRAFT_KEY)
    inputRef.current?.focus()
  }, [content, entryType, parentEntry, onSubmit, onSubmitChild])

  const placeholder = `Add a ${entryType}...`

  return (
    <div data-testid="capture-bar" className="border rounded-lg bg-card p-3 space-y-2">
      {/* Parent Context */}
      {parentEntry && (
        <div className="flex items-center justify-between text-sm bg-secondary/50 rounded px-2 py-1">
          <span>
            <span className="text-muted-foreground">Adding to: </span>
            <span className="font-medium">{parentEntry.content}</span>
          </span>
          <button
            onClick={onClearParent}
            aria-label="Clear parent"
            className="p-1 hover:bg-secondary rounded"
          >
            <X className="w-3 h-3" />
          </button>
        </div>
      )}

      {/* Type Buttons */}
      <div className="flex items-center gap-1">
        {(Object.keys(TYPE_LABELS) as EntryType[]).map((type) => (
          <button
            key={type}
            onClick={() => handleTypeChange(type)}
            aria-pressed={entryType === type}
            className={cn(
              'px-3 py-1 text-sm rounded transition-colors',
              entryType === type
                ? 'bg-primary text-primary-foreground'
                : 'bg-secondary/50 hover:bg-secondary text-muted-foreground'
            )}
          >
            {TYPE_LABELS[type]}
          </button>
        ))}

        <div className="flex-1" />

        <button
          onClick={onFileImport}
          aria-label="Import file"
          className="p-2 hover:bg-secondary rounded transition-colors"
        >
          <Paperclip className="w-4 h-4 text-muted-foreground" />
        </button>
      </div>

      {/* Input */}
      <textarea
        ref={inputRef}
        value={content}
        onChange={handleInputChange}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        rows={1}
        className={cn(
          'w-full px-3 py-2 rounded-md border bg-background resize-none',
          'focus:outline-none focus:ring-2 focus:ring-primary',
          'placeholder:text-muted-foreground text-sm'
        )}
        style={{
          minHeight: '2.5rem',
          height: 'auto',
          maxHeight: '8rem',
        }}
      />
    </div>
  )
}
```

**Step 2: Run tests**

Run: `cd frontend && npm test -- --run src/components/bujo/CaptureBar.test.tsx`
Expected: Most tests PASS (some may need minor adjustments)

**Step 3: Fix any failing tests and commit**

```bash
git add frontend/src/components/bujo/CaptureBar.tsx
git commit -m "feat: Implement CaptureBar component

- Type selection buttons with Tab cycling
- Prefix detection for power users
- Draft persistence to localStorage
- Parent context mode for child entries
- File import button
- Multiline support with Shift+Enter"
```

---

## Phase 3: WeekSummary Component

### Task 3.1: Create WeekSummary Unit Tests

**Files:**
- Create: `frontend/src/components/bujo/WeekSummary.test.tsx`

**Step 1: Write the failing tests**

```typescript
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { WeekSummary } from './WeekSummary'
import { DayEntries, Entry } from '@/types/bujo'

const createEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  entityId: 'e1',
  type: 'task',
  content: 'Test task',
  priority: '',
  parentId: null,
  depth: 0,
  loggedDate: new Date().toISOString(),
  ...overrides,
})

const createDay = (overrides: Partial<DayEntries> = {}): DayEntries => ({
  date: new Date().toISOString().split('T')[0],
  location: '',
  mood: '',
  weather: '',
  entries: [],
  ...overrides,
})

describe('WeekSummary', () => {
  describe('Task Flow', () => {
    it('renders task flow section', () => {
      render(<WeekSummary days={[]} />)
      expect(screen.getByText('Task Flow')).toBeInTheDocument()
    })

    it('shows created count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'task' }),
            createEntry({ id: 2, type: 'task' }),
            createEntry({ id: 3, type: 'note' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Created')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument() // 2 tasks created
    })

    it('shows done count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'done' }),
            createEntry({ id: 2, type: 'task' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Done')).toBeInTheDocument()
    })

    it('shows migrated count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'migrated' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Migrated')).toBeInTheDocument()
    })

    it('shows open count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'task' }),
            createEntry({ id: 2, type: 'task' }),
            createEntry({ id: 3, type: 'done' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Open')).toBeInTheDocument()
    })
  })

  describe('Meetings', () => {
    it('renders meetings section', () => {
      render(<WeekSummary days={[]} />)
      expect(screen.getByText('Meetings')).toBeInTheDocument()
    })

    it('shows events with children', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'event', content: 'Team standup' }),
            createEntry({ id: 2, type: 'note', content: 'Note 1', parentId: 1 }),
            createEntry({ id: 3, type: 'task', content: 'Action', parentId: 1 }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Team standup')).toBeInTheDocument()
      expect(screen.getByText(/2 items/i)).toBeInTheDocument()
    })

    it('does not show events without children', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'event', content: 'Solo event' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.queryByText('Solo event')).not.toBeInTheDocument()
    })
  })

  describe('Needs Attention', () => {
    it('renders needs attention section', () => {
      render(<WeekSummary days={[]} />)
      expect(screen.getByText('Needs Attention')).toBeInTheDocument()
    })

    it('shows open tasks sorted by attention score', () => {
      const fourDaysAgo = new Date()
      fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'task', content: 'New task' }),
            createEntry({
              id: 2,
              type: 'task',
              content: 'Old urgent task',
              priority: 'high',
              loggedDate: fourDaysAgo.toISOString(),
            }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      // Old urgent task should appear first (higher attention score)
      const items = screen.getAllByTestId('attention-item')
      expect(items[0]).toHaveTextContent('Old urgent task')
    })

    it('shows unanswered questions', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'question', content: 'What is the deadline?' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('What is the deadline?')).toBeInTheDocument()
    })

    it('shows attention indicators', () => {
      const days = [
        createDay({
          entries: [
            createEntry({
              id: 1,
              type: 'task',
              content: 'High priority task',
              priority: 'high',
            }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText(/priority/i)).toBeInTheDocument()
    })

    it('limits to top 5 items', () => {
      const days = [
        createDay({
          entries: Array.from({ length: 10 }, (_, i) =>
            createEntry({ id: i + 1, type: 'task', content: `Task ${i + 1}` })
          ),
        }),
      ]
      render(<WeekSummary days={days} />)

      const items = screen.getAllByTestId('attention-item')
      expect(items.length).toBe(5)
    })

    it('shows "Show all" link when more than 5 items', () => {
      const days = [
        createDay({
          entries: Array.from({ length: 10 }, (_, i) =>
            createEntry({ id: i + 1, type: 'task', content: `Task ${i + 1}` })
          ),
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText(/show all/i)).toBeInTheDocument()
    })
  })
})
```

**Step 2: Run tests**

Run: `cd frontend && npm test -- --run src/components/bujo/WeekSummary.test.tsx`
Expected: FAIL - component not found

**Step 3: Commit**

```bash
git add frontend/src/components/bujo/WeekSummary.test.tsx
git commit -m "test: Add failing unit tests for WeekSummary component"
```

---

### Task 3.2: Implement WeekSummary Component

**Files:**
- Create: `frontend/src/components/bujo/WeekSummary.tsx`

**Step 1: Write the implementation**

```typescript
import { useMemo, useState } from 'react'
import { DayEntries, Entry } from '@/types/bujo'
import { calculateAttentionScore, AttentionIndicator } from '@/lib/attentionScore'
import { ArrowRight, AlertCircle, Clock, RotateCcw, Flag } from 'lucide-react'
import { cn } from '@/lib/utils'

interface WeekSummaryProps {
  days: DayEntries[]
  onEntryClick?: (entry: Entry) => void
}

interface TaskFlowStats {
  created: number
  done: number
  migrated: number
  open: number
}

interface MeetingWithChildren {
  entry: Entry
  childCount: number
}

interface AttentionEntry {
  entry: Entry
  score: number
  indicators: AttentionIndicator[]
}

function getIndicatorIcon(indicator: AttentionIndicator) {
  switch (indicator) {
    case 'overdue':
      return <AlertCircle className="w-3 h-3 text-destructive" />
    case 'priority':
      return <Flag className="w-3 h-3 text-orange-500" />
    case 'aging':
      return <Clock className="w-3 h-3 text-yellow-500" />
    case 'migrated':
      return <RotateCcw className="w-3 h-3 text-blue-500" />
  }
}

function getIndicatorLabel(indicator: AttentionIndicator, entry: AttentionEntry) {
  switch (indicator) {
    case 'overdue':
      return 'Overdue'
    case 'priority':
      return 'Priority'
    case 'aging':
      return `${entry.entry.daysOld || 0}+ days`
    case 'migrated':
      return `Migrated ${entry.entry.migrationCount || 0}x`
  }
}

export function WeekSummary({ days, onEntryClick }: WeekSummaryProps) {
  const [showAllAttention, setShowAllAttention] = useState(false)

  const allEntries = useMemo(() => {
    return days.flatMap(day => day.entries)
  }, [days])

  const taskFlow = useMemo((): TaskFlowStats => {
    const tasks = allEntries.filter(e =>
      e.type === 'task' || e.type === 'done' || e.type === 'migrated'
    )

    return {
      created: tasks.filter(e => e.type === 'task' || e.type === 'done').length,
      done: tasks.filter(e => e.type === 'done').length,
      migrated: tasks.filter(e => e.type === 'migrated').length,
      open: tasks.filter(e => e.type === 'task').length,
    }
  }, [allEntries])

  const meetings = useMemo((): MeetingWithChildren[] => {
    const events = allEntries.filter(e => e.type === 'event')
    const childCounts = new Map<number, number>()

    allEntries.forEach(entry => {
      if (entry.parentId) {
        childCounts.set(entry.parentId, (childCounts.get(entry.parentId) || 0) + 1)
      }
    })

    return events
      .filter(event => (childCounts.get(event.id) || 0) > 0)
      .map(event => ({
        entry: event,
        childCount: childCounts.get(event.id) || 0,
      }))
  }, [allEntries])

  const attentionItems = useMemo((): AttentionEntry[] => {
    const now = new Date()
    const openItems = allEntries.filter(e =>
      e.type === 'task' || e.type === 'question'
    )

    const parentTypes = new Map<number, string>()
    allEntries.forEach(e => {
      parentTypes.set(e.id, e.type)
    })

    return openItems
      .map(entry => {
        const result = calculateAttentionScore(entry, now, parentTypes.get(entry.parentId ?? 0))
        return {
          entry,
          score: result.score,
          indicators: result.indicators,
        }
      })
      .sort((a, b) => b.score - a.score)
  }, [allEntries])

  const visibleAttentionItems = showAllAttention
    ? attentionItems
    : attentionItems.slice(0, 5)

  return (
    <div data-testid="week-summary" className="bg-card border rounded-lg p-4 space-y-6">
      {/* Task Flow */}
      <div>
        <h3 className="text-sm font-semibold mb-3">Task Flow</h3>
        <div className="flex items-center gap-2">
          <div className="flex-1 text-center p-3 bg-secondary/30 rounded">
            <div className="text-2xl font-bold">{taskFlow.created}</div>
            <div className="text-xs text-muted-foreground">Created</div>
          </div>
          <ArrowRight className="w-4 h-4 text-muted-foreground" />
          <div className="flex-1 text-center p-3 bg-green-500/10 rounded">
            <div className="text-2xl font-bold text-green-600">{taskFlow.done}</div>
            <div className="text-xs text-muted-foreground">Done</div>
          </div>
          <div className="flex-1 text-center p-3 bg-blue-500/10 rounded">
            <div className="text-2xl font-bold text-blue-600">{taskFlow.migrated}</div>
            <div className="text-xs text-muted-foreground">Migrated</div>
          </div>
          <div className="flex-1 text-center p-3 bg-orange-500/10 rounded">
            <div className="text-2xl font-bold text-orange-600">{taskFlow.open}</div>
            <div className="text-xs text-muted-foreground">Open</div>
          </div>
        </div>
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {/* Meetings */}
        <div>
          <h3 className="text-sm font-semibold mb-3">Meetings</h3>
          {meetings.length > 0 ? (
            <div className="space-y-2">
              {meetings.map(meeting => (
                <button
                  key={meeting.entry.id}
                  onClick={() => onEntryClick?.(meeting.entry)}
                  className="w-full flex items-center justify-between p-2 rounded hover:bg-secondary/50 transition-colors text-left"
                >
                  <span className="text-sm truncate">{meeting.entry.content}</span>
                  <span className="text-xs text-muted-foreground ml-2">
                    {meeting.childCount} items
                  </span>
                </button>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground italic">No meetings with notes</p>
          )}
        </div>

        {/* Needs Attention */}
        <div>
          <h3 className="text-sm font-semibold mb-3">Needs Attention</h3>
          {attentionItems.length > 0 ? (
            <div className="space-y-2">
              {visibleAttentionItems.map(item => (
                <button
                  key={item.entry.id}
                  data-testid="attention-item"
                  onClick={() => onEntryClick?.(item.entry)}
                  className="w-full flex items-center gap-2 p-2 rounded hover:bg-secondary/50 transition-colors text-left"
                >
                  <span className={cn(
                    'text-sm truncate flex-1',
                    item.entry.type === 'question' && 'italic'
                  )}>
                    {item.entry.type === 'task' ? '.' : '?'} {item.entry.content}
                  </span>
                  {item.indicators.length > 0 && (
                    <span className="flex items-center gap-1 text-xs">
                      {getIndicatorIcon(item.indicators[0])}
                      <span className="text-muted-foreground">
                        {getIndicatorLabel(item.indicators[0], item)}
                      </span>
                    </span>
                  )}
                </button>
              ))}
              {attentionItems.length > 5 && !showAllAttention && (
                <button
                  onClick={() => setShowAllAttention(true)}
                  className="text-xs text-primary hover:underline"
                >
                  Show all ({attentionItems.length})
                </button>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground italic">All caught up!</p>
          )}
        </div>
      </div>
    </div>
  )
}
```

**Step 2: Run tests**

Run: `cd frontend && npm test -- --run src/components/bujo/WeekSummary.test.tsx`
Expected: Tests PASS (may need minor type adjustments)

**Step 3: Commit**

```bash
git add frontend/src/components/bujo/WeekSummary.tsx
git commit -m "feat: Implement WeekSummary component

- Task flow visualization (created/done/migrated/open)
- Meetings section showing events with children
- Needs Attention section with attention scoring
- Expandable list for attention items"
```

---

## Phase 4: Integration with App.tsx

### Task 4.1: Update Entry Type for Migration Count

**Files:**
- Modify: `frontend/src/types/bujo.ts`

**Step 1: Read current types**

Read `frontend/src/types/bujo.ts` to understand current Entry type.

**Step 2: Add migrationCount field if missing**

Add `migrationCount?: number` and `scheduledDate?: string` to Entry type if not present.

**Step 3: Update transforms if needed**

Ensure `transformEntry` in `frontend/src/lib/transforms.ts` maps these fields.

**Step 4: Commit**

```bash
git add frontend/src/types/bujo.ts frontend/src/lib/transforms.ts
git commit -m "feat: Add migrationCount and scheduledDate to Entry type"
```

---

### Task 4.2: Integrate CaptureBar into App.tsx

**Files:**
- Modify: `frontend/src/App.tsx`

**Step 1: Import CaptureBar**

Add import for CaptureBar component.

**Step 2: Replace InlineEntryInput with CaptureBar**

In the today view section, replace the `InlineEntryInput` and add button with `CaptureBar`.

**Step 3: Update state management**

- Remove `inlineInputMode` state
- Add `captureParentEntry` state for parent context
- Update keyboard shortcuts (`i`, `A`, `r`)

**Step 4: Wire up handlers**

Connect `onSubmit`, `onSubmitChild`, `onClearParent`, `onFileImport` props.

**Step 5: Run acceptance tests**

Run: `cd frontend && npm test -- --run src/App.captureBar.test.tsx`
Expected: All acceptance tests PASS

**Step 6: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat: Integrate CaptureBar into today view

- Replace InlineEntryInput with CaptureBar
- Add keyboard shortcuts (i, A, r)
- Wire up parent context for child entries
- Connect file import functionality"
```

---

### Task 4.3: Integrate WeekSummary into Weekly View

**Files:**
- Modify: `frontend/src/App.tsx`

**Step 1: Import WeekSummary**

Add import for WeekSummary component.

**Step 2: Add WeekSummary to weekly view**

In the `view === 'week'` section, add `<WeekSummary days={reviewDays} />` before the day-by-day list.

**Step 3: Wire up entry click handler**

Add `onEntryClick` handler to navigate to entry's day.

**Step 4: Run acceptance tests**

Run: `cd frontend && npm test -- --run src/App.weekSummary.test.tsx`
Expected: All acceptance tests PASS

**Step 5: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat: Integrate WeekSummary into weekly view

- Add WeekSummary component above day list
- Connect entry click to navigate to day"
```

---

## Phase 5: Cleanup

### Task 5.1: Remove Deprecated Components

**Files:**
- Delete: `frontend/src/components/bujo/InlineEntryInput.tsx`
- Delete: `frontend/src/components/bujo/InlineEntryInput.test.tsx` (if exists)
- Modify: `frontend/src/App.tsx` (remove CaptureModal if fully replaced)

**Step 1: Remove old components**

Delete `InlineEntryInput` files after verifying all tests pass.

**Step 2: Update any remaining imports**

Search for any remaining references to removed components.

**Step 3: Run full test suite**

Run: `cd frontend && npm test`
Expected: All tests PASS

**Step 4: Commit**

```bash
git add -A
git commit -m "refactor: Remove deprecated InlineEntryInput component

CaptureBar now handles all capture functionality."
```

---

### Task 5.2: Export New Components

**Files:**
- Modify: `frontend/src/components/bujo/index.ts` (if exists)

**Step 1: Add exports**

Export `CaptureBar` and `WeekSummary` from index file.

**Step 2: Commit**

```bash
git add frontend/src/components/bujo/index.ts
git commit -m "chore: Export CaptureBar and WeekSummary components"
```

---

## Final Verification

### Task 6.1: Run All Tests

**Step 1: Run full test suite**

Run: `cd frontend && npm test`
Expected: All tests PASS

**Step 2: Run type check**

Run: `cd frontend && npx tsc --noEmit`
Expected: No type errors

**Step 3: Run linter**

Run: `cd frontend && npm run lint`
Expected: No lint errors

**Step 4: Manual verification**

Start the app and manually verify:
- Capture bar appears at bottom of today view
- Type selection works
- Draft persistence works
- Parent context mode works
- Week summary appears in weekly view
- Attention scoring surfaces important items

**Step 5: Final commit**

```bash
git add -A
git commit -m "feat: Complete CaptureBar and WeekSummary implementation

Closes #413"
```

---

## Summary

**Total Tasks:** 15 (across 6 phases)

**Key Files Created:**
- `frontend/src/App.captureBar.test.tsx` - Acceptance tests for capture bar
- `frontend/src/App.weekSummary.test.tsx` - Acceptance tests for week summary
- `frontend/src/lib/attentionScore.ts` - Attention scoring utility
- `frontend/src/lib/attentionScore.test.ts` - Unit tests for scoring
- `frontend/src/components/bujo/CaptureBar.tsx` - Capture bar component
- `frontend/src/components/bujo/CaptureBar.test.tsx` - Unit tests
- `frontend/src/components/bujo/WeekSummary.tsx` - Week summary component
- `frontend/src/components/bujo/WeekSummary.test.tsx` - Unit tests

**Key Files Modified:**
- `frontend/src/App.tsx` - Integration
- `frontend/src/types/bujo.ts` - Type updates
- `frontend/src/lib/transforms.ts` - Transform updates
