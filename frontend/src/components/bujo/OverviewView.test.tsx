import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OverviewView } from './OverviewView'
import { Entry } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
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
    expect(screen.getByText(/overdue tasks/i)).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
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
    expect(screen.getByText(/no overdue tasks/i)).toBeInTheDocument()
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
