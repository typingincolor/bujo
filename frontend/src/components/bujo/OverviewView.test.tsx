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

describe('OverviewView - Interactions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls MarkEntryDone when clicking complete button', async () => {
    const user = userEvent.setup()
    const onEntryChanged = vi.fn()
    render(<OverviewView overdueEntries={[createTestEntry({ id: 42 })]} onEntryChanged={onEntryChanged} />)

    await user.click(screen.getByTitle('Mark done'))

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(42)
    })
  })

  it('calls onEntryChanged after marking entry done', async () => {
    const user = userEvent.setup()
    const onEntryChanged = vi.fn()
    render(<OverviewView overdueEntries={[createTestEntry()]} onEntryChanged={onEntryChanged} />)

    await user.click(screen.getByTitle('Mark done'))

    await waitFor(() => {
      expect(onEntryChanged).toHaveBeenCalled()
    })
  })

  it('calls MarkEntryUndone when clicking undo on done entry', async () => {
    const user = userEvent.setup()
    const onEntryChanged = vi.fn()
    render(<OverviewView overdueEntries={[createTestEntry({ id: 42, type: 'done' })]} onEntryChanged={onEntryChanged} />)

    await user.click(screen.getByTitle('Mark undone'))

    await waitFor(() => {
      expect(MarkEntryUndone).toHaveBeenCalledWith(42)
    })
  })

  it('shows checkmark symbol for done entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'done' })]} />)
    const undoneButton = screen.getByTitle('Mark undone')
    expect(undoneButton).toHaveTextContent('✓')
  })

  it('shows cancel button for non-cancelled entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task' })]} />)
    expect(screen.getByTitle('Cancel entry')).toBeInTheDocument()
  })

  it('shows uncancel button for cancelled entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'cancelled' })]} />)
    expect(screen.getByTitle('Uncancel entry')).toBeInTheDocument()
  })

  it('calls CancelEntry when cancel button is clicked', async () => {
    const { CancelEntry } = await import('@/wailsjs/go/wails/App')
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={[createTestEntry({ id: 42, type: 'task' })]} />)

    await user.click(screen.getByTitle('Cancel entry'))

    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('calls UncancelEntry when uncancel button is clicked', async () => {
    const { UncancelEntry } = await import('@/wailsjs/go/wails/App')
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={[createTestEntry({ id: 42, type: 'cancelled' })]} />)

    await user.click(screen.getByTitle('Uncancel entry'))

    await waitFor(() => {
      expect(UncancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('shows delete button for all entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task' })]} />)
    expect(screen.getByTitle('Delete entry')).toBeInTheDocument()
  })

  it('shows edit button for non-cancelled entries when onEdit provided', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task' })]} onEdit={vi.fn()} />)
    expect(screen.getByTitle('Edit entry')).toBeInTheDocument()
  })

  it('does not show edit button for cancelled entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'cancelled' })]} onEdit={vi.fn()} />)
    expect(screen.queryByTitle('Edit entry')).not.toBeInTheDocument()
  })

  it('shows migrate button for task entries when onMigrate provided', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task' })]} onMigrate={vi.fn()} />)
    expect(screen.getByTitle('Migrate entry')).toBeInTheDocument()
  })

  it('calls onMigrate when migrate button is clicked', async () => {
    const user = userEvent.setup()
    const onMigrate = vi.fn()
    const entry = createTestEntry({ id: 42, type: 'task' })
    render(<OverviewView overdueEntries={[entry]} onMigrate={onMigrate} />)

    await user.click(screen.getByTitle('Migrate entry'))

    expect(onMigrate).toHaveBeenCalledWith(entry)
  })

  it('shows priority button for all entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task' })]} />)
    expect(screen.getByTitle('Cycle priority')).toBeInTheDocument()
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

describe('OverviewView - Context Display', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows only task entries by default, hiding parent context entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Overdue task', type: 'task', parentId: 1 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Task should be visible
    expect(screen.getByText('Overdue task')).toBeInTheDocument()
    // Parent event should be hidden by default
    expect(screen.queryByText('Parent event')).not.toBeInTheDocument()
  })

  it('shows context when clicking on a task', async () => {
    const user = userEvent.setup()
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Overdue task', type: 'task', parentId: 1 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Click on the task to expand context
    await user.click(screen.getByText('Overdue task'))

    // Parent event should now be visible
    expect(screen.getByText('Parent event')).toBeInTheDocument()
  })

  it('hides context when clicking on an expanded task again', async () => {
    const user = userEvent.setup()
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Overdue task', type: 'task', parentId: 1 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Click to expand
    await user.click(screen.getByText('Overdue task'))
    expect(screen.getByText('Parent event')).toBeInTheDocument()

    // Click again to collapse
    await user.click(screen.getByText('Overdue task'))
    expect(screen.queryByText('Parent event')).not.toBeInTheDocument()
  })

  it('shows multi-level context when clicking on a deeply nested task', async () => {
    const user = userEvent.setup()
    const entries = [
      createTestEntry({ id: 1, content: 'Grandparent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Parent note', type: 'note', parentId: 1 }),
      createTestEntry({ id: 3, content: 'Overdue task', type: 'task', parentId: 2 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Click on the task to expand context
    await user.click(screen.getByText('Overdue task'))

    // Both parent note and grandparent event should be visible
    expect(screen.getByText('Parent note')).toBeInTheDocument()
    expect(screen.getByText('Grandparent event')).toBeInTheDocument()
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

  it('expands context with Enter key', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Task with parent', type: 'task', parentId: 1 }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    await user.keyboard('j{Enter}') // Select first visible task, then Enter

    await waitFor(() => {
      expect(screen.getByText('Parent event')).toBeInTheDocument()
    })
  })
})

describe('OverviewView - Context Pill', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows context pill with ancestor count when entry has parent and is not expanded', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Task with parent', type: 'task', parentId: 1 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toBeInTheDocument()
    expect(pill).toHaveTextContent('1')
  })

  it('shows correct ancestor count for deeply nested entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Grandparent', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Parent', type: 'note', parentId: 1 }),
      createTestEntry({ id: 3, content: 'Child task', type: 'task', parentId: 2 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toHaveTextContent('2')
  })

  it('does not show context pill when entry has no parent', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Root task', type: 'task', parentId: null }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    expect(screen.getByText('Root task')).toBeInTheDocument()
    expect(screen.queryByTestId('context-pill')).not.toBeInTheDocument()
  })

  it('hides context pill when entry is expanded', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Task with parent', type: 'task', parentId: 1 }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    expect(screen.getByTestId('context-pill')).toBeInTheDocument()

    // Click pill to expand
    await user.click(screen.getByTestId('context-pill'))

    await waitFor(() => {
      expect(screen.getByText('Parent event')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('context-pill')).not.toBeInTheDocument()
  })

  it('clicking context pill toggles expand/collapse', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Task with parent', type: 'task', parentId: 1 }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    // Click pill to expand
    await user.click(screen.getByTestId('context-pill'))
    await waitFor(() => {
      expect(screen.getByText('Parent event')).toBeInTheDocument()
    })

    // Click entry row to collapse (pill is hidden when expanded)
    await user.click(screen.getByText('Task with parent'))
    await waitFor(() => {
      expect(screen.queryByText('Parent event')).not.toBeInTheDocument()
    })

    // Pill should reappear
    expect(screen.getByTestId('context-pill')).toBeInTheDocument()
  })

  it('context pill click does not trigger other entry actions', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Task with parent', type: 'task', parentId: 1 }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    // Click pill - should only expand, not mark done
    await user.click(screen.getByTestId('context-pill'))

    await waitFor(() => {
      expect(screen.getByText('Parent event')).toBeInTheDocument()
    })
    expect(MarkEntryDone).not.toHaveBeenCalled()
  })

  it('context starts collapsed by default', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Task with parent', type: 'task', parentId: 1 }),
    ]
    render(<OverviewView overdueEntries={entries} />)

    // Parent should not be visible initially
    expect(screen.queryByText('Parent event')).not.toBeInTheDocument()
    // Pill should be visible
    expect(screen.getByTestId('context-pill')).toBeInTheDocument()
  })
})

describe('OverviewView - Symbol Click Toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls MarkEntryDone when clicking symbol for task entry', async () => {
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={[createTestEntry({ id: 42, type: 'task' })]} />)

    const symbolButton = screen.getByTitle('Mark done')
    await user.click(symbolButton)

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(42)
    })
  })

  it('calls MarkEntryUndone when clicking symbol for done entry', async () => {
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={[createTestEntry({ id: 42, type: 'done' })]} />)

    const symbolButton = screen.getByTitle('Mark undone')
    await user.click(symbolButton)

    await waitFor(() => {
      expect(MarkEntryUndone).toHaveBeenCalledWith(42)
    })
  })

  it('symbol shows task bullet for task entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'task' })]} />)
    const symbolButton = screen.getByTitle('Mark done')
    expect(symbolButton).toHaveTextContent('•')
  })

  it('symbol shows checkmark for done entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'done' })]} />)
    const symbolButton = screen.getByTitle('Mark undone')
    expect(symbolButton).toHaveTextContent('✓')
  })

  it('symbol is not clickable for cancelled entries', () => {
    render(<OverviewView overdueEntries={[createTestEntry({ type: 'cancelled' })]} />)
    expect(screen.queryByTitle('Mark done')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Mark undone')).not.toBeInTheDocument()
  })

  it('symbol click does not toggle row expand', async () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent event', type: 'event', parentId: null }),
      createTestEntry({ id: 2, content: 'Child task', type: 'task', parentId: 1 }),
    ]
    const user = userEvent.setup()
    render(<OverviewView overdueEntries={entries} />)

    const symbolButton = screen.getByTitle('Mark done')
    await user.click(symbolButton)

    // Parent should NOT be visible (expand was not triggered)
    expect(screen.queryByText('Parent event')).not.toBeInTheDocument()
  })
})

describe('OverviewView - Action Icon Placeholders', () => {
  it('renders same number of action slots for task and cancelled entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Task entry', type: 'task' }),
      createTestEntry({ id: 2, content: 'Cancelled entry', type: 'cancelled' }),
    ]

    render(<OverviewView overdueEntries={entries} />)

    const taskRow = screen.getByText('Task entry').closest('[data-entry-id]')
    const cancelledRow = screen.getByText('Cancelled entry').closest('[data-entry-id]')

    const taskActionSlots = taskRow?.querySelectorAll('[data-action-slot]')
    const cancelledActionSlots = cancelledRow?.querySelectorAll('[data-action-slot]')

    expect(taskActionSlots?.length).toBeGreaterThan(0)
    expect(taskActionSlots?.length).toBe(cancelledActionSlots?.length)
  })

  it('renders same number of action slots for task and done entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Task entry', type: 'task' }),
      createTestEntry({ id: 2, content: 'Done entry', type: 'done' }),
    ]

    render(<OverviewView overdueEntries={entries} />)

    const taskRow = screen.getByText('Task entry').closest('[data-entry-id]')
    const doneRow = screen.getByText('Done entry').closest('[data-entry-id]')

    const taskActionSlots = taskRow?.querySelectorAll('[data-action-slot]')
    const doneActionSlots = doneRow?.querySelectorAll('[data-action-slot]')

    expect(taskActionSlots?.length).toBeGreaterThan(0)
    expect(taskActionSlots?.length).toBe(doneActionSlots?.length)
  })
})
