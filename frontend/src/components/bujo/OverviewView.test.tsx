import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OverviewView } from './OverviewView'
import { Entry } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  RetypeEntry: vi.fn().mockResolvedValue(undefined),
}))

import { MarkEntryDone, MarkEntryUndone } from '@/wailsjs/go/wails/App'

const createTestEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  content: 'Test task',
  type: 'task',
  priority: 'none',
  parentId: null,
  loggedDate: '2026-01-15',
  ...overrides,
})

describe('OverviewView - Display', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the overview header with count', () => {
    render(<OverviewView overdueEntries={[createTestEntry()]} />)
    expect(screen.getByText(/pending tasks/i)).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('uses Clock icon in header (not AlertTriangle)', () => {
    render(<OverviewView overdueEntries={[createTestEntry()]} />)
    expect(screen.getByTestId('outstanding-icon')).toBeInTheDocument()
  })

  it('renders multiple overdue entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Task one' }),
      createTestEntry({ id: 2, content: 'Task two' }),
      createTestEntry({ id: 3, content: 'Task three' }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    expect(screen.getByText('Task one')).toBeInTheDocument()
    expect(screen.getByText('Task two')).toBeInTheDocument()
    expect(screen.getByText('Task three')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('shows empty state when no overdue entries', () => {
    render(<OverviewView overdueEntries={[]} />)
    expect(screen.getByText(/no pending tasks/i)).toBeInTheDocument()
  })

  it('displays entry date', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ loggedDate: '2026-01-10' })]} />)
    expect(screen.getByText(/jan 10/i)).toBeInTheDocument()
  })

  it('shows entry type symbol', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task' })]} />)
    // Task symbol should be visible (bullet point or similar)
    expect(screen.getByTestId('entry-symbol')).toBeInTheDocument()
  })

  it('shows priority indicator for high priority entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ priority: 'high' })]} />)
    expect(screen.getByText('!!!')).toBeInTheDocument()
  })
})

describe('OverviewView - Direct Click Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls onNavigateToEntry when clicking entry', async () => {
    const user = userEvent.setup()
    const onNavigateToEntry = vi.fn()
    const entry = createTestEntry({ id: 42, content: 'Test task' })
    render(<OverviewView overdueEntries={[entry]} onNavigateToEntry={onNavigateToEntry} />)

    await user.click(screen.getByText('Test task'))

    expect(onNavigateToEntry).toHaveBeenCalledWith(entry)
  })
})

describe('OverviewView - Grouping by Date', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('groups entries by date', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Buy groceries', loggedDate: '2026-01-10' }),
      createTestEntry({ id: 2, content: 'Call mom', loggedDate: '2026-01-10' }),
      createTestEntry({ id: 3, content: 'Finish report', loggedDate: '2026-01-11' }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Should have two date headers (h3 elements with date text)
    const headers = screen.getAllByRole('heading', { level: 3 })
    expect(headers).toHaveLength(2)
    expect(headers[0]).toHaveTextContent('Jan 10')
    expect(headers[1]).toHaveTextContent('Jan 11')
  })
})

describe('OverviewView - Collapsible', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('can collapse the overdue section', async () => {
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={[createTestEntry({ content: 'My task' })]} />)

    // Entry should be visible initially
    expect(screen.getByText('My task')).toBeInTheDocument()

    // Click collapse button
    await user.click(screen.getByTitle('Collapse'))

    // Entry should be removed from DOM when collapsed
    expect(screen.queryByText('My task')).not.toBeInTheDocument()
  })

  it('can expand collapsed section', async () => {
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={[createTestEntry({ content: 'My task' })]} />)

    // Collapse first
    await user.click(screen.getByTitle('Collapse'))
    expect(screen.queryByText('My task')).not.toBeInTheDocument()

    // Expand
    await user.click(screen.getByTitle('Expand'))
    expect(screen.getByText('My task')).toBeInTheDocument()
  })
})

describe('OverviewView - Visual Styling', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders outstanding task entries without red destructive color', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task', content: 'Outstanding task' })]} />)
    const content = screen.getByText('Outstanding task')
    expect(content).not.toHaveClass('text-destructive')
  })

  it('renders done entries with success color (not strikethrough)', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'done', content: 'Done task' })]} />)
    const content = screen.getByText('Done task')
    expect(content).toHaveClass('text-bujo-done')
    expect(content).not.toHaveClass('line-through')
  })

  it('renders cancelled entries with strikethrough style', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'cancelled', content: 'Cancelled task' })]} />)
    const content = screen.getByText('Cancelled task')
    expect(content).toHaveClass('line-through')
  })
})

describe('OverviewView - Entry Filtering', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows only task entries, hiding parent context entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Overdue task', type: 'task', parentId: 1 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Task should be visible
    expect(screen.getByText('Overdue task')).toBeInTheDocument()
    // Parent event should be hidden (not a task-type)
    expect(screen.queryByText('Parent event')).not.toBeInTheDocument()
  })

  it('shows context dot for task with parent', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Overdue task', type: 'task', parentId: 1 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Child task should show context dot
    expect(screen.getByTestId('context-dot')).toBeInTheDocument()
  })

  it('counts only task entries in the badge', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Task one', type: 'task', parentId: 1 }),
      createTestEntry({ id: 3, content: 'Task two', type: 'task', parentId: null }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Badge should show 2 (only tasks), not 3 (all entries)
    expect(screen.getByText('2')).toBeInTheDocument()
  })
})

describe('OverviewView - Keyboard Shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('selects first entry with j key', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'First task' }),
      createTestEntry({ id: 2, content: 'Second task' }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('j')

    await waitFor(() => {
      const firstTask = screen.getByText('First task').closest('.cursor-pointer')
      expect(firstTask).toHaveClass('ring-2')
    })
  })

  it('navigates down with j key', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'First task' }),
      createTestEntry({ id: 2, content: 'Second task' }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('jj') // Press j twice to select second

    await waitFor(() => {
      const secondTask = screen.getByText('Second task').closest('.cursor-pointer')
      expect(secondTask).toHaveClass('ring-2')
    })
  })

  it('navigates up with k key', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'First task' }),
      createTestEntry({ id: 2, content: 'Second task' }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('jjk') // Down twice, up once

    await waitFor(() => {
      const firstTask = screen.getByText('First task').closest('.cursor-pointer')
      expect(firstTask).toHaveClass('ring-2')
    })
  })

  it('navigates with arrow keys', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'First task' }),
      createTestEntry({ id: 2, content: 'Second task' }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('{ArrowDown}{ArrowDown}')

    await waitFor(() => {
      const secondTask = screen.getByText('Second task').closest('.cursor-pointer')
      expect(secondTask).toHaveClass('ring-2')
    })
  })

  it('toggles done with Space key for selected task', async () => {
    const entries = [createTestEntry({ id: 42, content: 'Test task', type: 'task' })]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('j ') // Select first, then Space

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(42)
    })
  })

  it('toggles undone with Space key for selected done entry', async () => {
    const entries = [createTestEntry({ id: 42, content: 'Done task', type: 'done' })]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('j ') // Select first, then Space

    await waitFor(() => {
      expect(MarkEntryUndone).toHaveBeenCalledWith(42)
    })
  })

  it('cancels entry with x key', async () => {
    const { CancelEntry } = await import('@/wailsjs/go/wails/App')
    const entries = [createTestEntry({ id: 42, content: 'Test task', type: 'task' })]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('jx') // Select first, then x

    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('uncancels entry with x key when cancelled', async () => {
    const { UncancelEntry } = await import('@/wailsjs/go/wails/App')
    const entries = [createTestEntry({ id: 42, content: 'Cancelled task', type: 'cancelled' })]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('jx') // Select first, then x

    await waitFor(() => {
      expect(UncancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('cycles priority with p key', async () => {
    const { CyclePriority } = await import('@/wailsjs/go/wails/App')
    const entries = [createTestEntry({ id: 42, content: 'Test task', type: 'task' })]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('jp') // Select first, then p

    await waitFor(() => {
      expect(CyclePriority).toHaveBeenCalledWith(42)
    })
  })

  it('cycles type with t key', async () => {
    const { RetypeEntry } = await import('@/wailsjs/go/wails/App')
    const entries = [createTestEntry({ id: 42, content: 'Test task', type: 'task' })]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('jt') // Select first, then t

    await waitFor(() => {
      expect(RetypeEntry).toHaveBeenCalledWith(42, 'note')
    })
  })

  it('navigates to entry with Enter key when onNavigateToEntry provided', async () => {
    const onNavigateToEntry = vi.fn()
    const entries = [createTestEntry({ id: 42, content: 'Test task', type: 'task' })]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} onNavigateToEntry={onNavigateToEntry} />)

    await user.keyboard('j{Enter}') // Select first, then Enter

    await waitFor(() => {
      expect(onNavigateToEntry).toHaveBeenCalledWith(entries[0])
    })
  })
})

describe('OverviewView - Context Dot (No Popover)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not render entry context popover', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Task', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-25' }),
    ]
    render(<OverviewView overdueEntries={entries} />)
    expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
  })

  it('shows context dot for entries with parent', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Child task', type: 'task', priority: 'none', parentId: 99, loggedDate: '2026-01-25' }),
    ]
    render(<OverviewView overdueEntries={entries} />)
    expect(screen.getByTestId('context-dot')).toBeInTheDocument()
  })

  it('does not show context dot for root entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Root task', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-25' }),
    ]
    render(<OverviewView overdueEntries={entries} />)
    expect(screen.queryByTestId('context-dot')).not.toBeInTheDocument()
  })
})

describe('OverviewView - Navigate to Entry (Direct Click)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls onNavigateToEntry when clicking entry directly', async () => {
    const user = userEvent.setup()
    const onNavigateToEntry = vi.fn()
    const entry = createTestEntry({ id: 42, content: 'Test task', type: 'task', loggedDate: '2026-01-15' })
    render(<OverviewView overdueEntries={[entry]} onNavigateToEntry={onNavigateToEntry} />)

    await user.click(screen.getByText('Test task'))

    expect(onNavigateToEntry).toHaveBeenCalledWith(entry)
  })

  it('navigates to any entry type when clicked', async () => {
    const user = userEvent.setup()
    const onNavigateToEntry = vi.fn()
    const entries = [
      createTestEntry({ id: 1, content: 'Task', type: 'task' }),
      createTestEntry({ id: 2, content: 'Done', type: 'done' }),
      createTestEntry({ id: 3, content: 'Cancelled', type: 'cancelled' }),
    ]
    render(<OverviewView overdueEntries={entries} onNavigateToEntry={onNavigateToEntry} />)

    // Click each entry
    await user.click(screen.getByText('Task'))
    expect(onNavigateToEntry).toHaveBeenCalledWith(entries[0])

    await user.click(screen.getByText('Done'))
    expect(onNavigateToEntry).toHaveBeenCalledWith(entries[1])

    await user.click(screen.getByText('Cancelled'))
    expect(onNavigateToEntry).toHaveBeenCalledWith(entries[2])
  })
})
