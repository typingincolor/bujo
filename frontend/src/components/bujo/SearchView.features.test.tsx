import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchView } from './SearchView'

vi.mock('@/wailsjs/go/wails/App', () => ({
  Search: vi.fn().mockResolvedValue([]),
  GetEntry: vi.fn().mockResolvedValue(null),
  GetEntryAncestors: vi.fn().mockResolvedValue([]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(1),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  RetypeEntry: vi.fn().mockResolvedValue(undefined),
}))

import { Search, MarkEntryDone, MarkEntryUndone } from '@/wailsjs/go/wails/App'

const createMockEntry = (overrides: Partial<{ ID: number; Content: string; Type: string; CreatedAt: string; ParentID: number | null }>) => ({
  ID: 1,
  EntityID: 'test-entity',
  Type: 'task',
  Content: 'Test content',
  Priority: 'none',
  ParentID: null,
  Depth: 0,
  CreatedAt: '2024-01-15T10:00:00Z',
  convertValues: vi.fn(),
  ...overrides,
})

describe('SearchView - Double Click Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls onNavigateToEntry when double-clicking a search result', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const onNavigateToEntry = vi.fn()
    const user = userEvent.setup()
    render(<SearchView onNavigateToEntry={onNavigateToEntry} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test entry')).toBeInTheDocument()
    })

    const result = screen.getByText('Test entry').closest('[data-result-id]')
    expect(result).toBeInTheDocument()

    await user.dblClick(result!)

    expect(onNavigateToEntry).toHaveBeenCalledWith({
      id: 42,
      content: 'Test entry',
      type: 'task',
      priority: 'none',
      date: '2024-01-15T10:00:00Z',
      parentId: null,
    })
  })

  it('does not call onNavigateToEntry on single click', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const onNavigateToEntry = vi.fn()
    const user = userEvent.setup()
    render(<SearchView onNavigateToEntry={onNavigateToEntry} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test entry')).toBeInTheDocument()
    })

    const result = screen.getByText('Test entry').closest('[data-result-id]')
    await user.click(result!)

    expect(onNavigateToEntry).not.toHaveBeenCalled()
  })
})

describe('SearchView - Symbol Click Toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls MarkEntryDone when clicking symbol for task entry', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Task to complete', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'task')

    await waitFor(() => {
      expect(screen.getByTitle('Task')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Task'))

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(42)
    })
  })

  it('calls MarkEntryUndone when clicking symbol for done entry', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Completed task', Type: 'done', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'done')

    await waitFor(() => {
      expect(screen.getByTitle('Done')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Done'))

    await waitFor(() => {
      expect(MarkEntryUndone).toHaveBeenCalledWith(42)
    })
  })

  it('symbol shows task bullet for task entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'task')

    await waitFor(() => {
      const symbolButton = screen.getByTitle('Task')
      expect(symbolButton).toHaveTextContent('•')
    })
  })

  it('symbol shows checkmark for done entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Done entry', Type: 'done', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'done')

    await waitFor(() => {
      const symbolButton = screen.getByTitle('Done')
      expect(symbolButton).toHaveTextContent('✓')
    })
  })

  it('symbol is not clickable for cancelled entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Cancelled entry', Type: 'cancelled', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'cancelled')

    await waitFor(() => {
      expect(screen.getByText('Cancelled entry')).toBeInTheDocument()
    })

    // Symbol for cancelled entries should not be a button with task/done title
    expect(screen.queryByTitle('Task')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Done')).not.toBeInTheDocument()
  })
})

describe('SearchView - Entry Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls onSelectEntry when clicking a search result', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const onSelectEntry = vi.fn()
    const user = userEvent.setup()
    render(<SearchView onSelectEntry={onSelectEntry} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test entry')).toBeInTheDocument()
    })

    const result = screen.getByText('Test entry').closest('[data-result-id]')
    await user.click(result!)

    expect(onSelectEntry).toHaveBeenCalledWith({
      id: 42,
      content: 'Test entry',
      type: 'task',
      priority: 'none',
      date: '2024-01-15T10:00:00Z',
      parentId: null,
    })
  })
})
