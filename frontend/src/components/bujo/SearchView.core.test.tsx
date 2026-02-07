import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchView } from './SearchView'

vi.mock('@/wailsjs/go/wails/App', () => ({
  Search: vi.fn().mockResolvedValue([]),
  SearchByTags: vi.fn().mockResolvedValue([]),
  GetAllTags: vi.fn().mockResolvedValue([]),
  GetEntry: vi.fn().mockResolvedValue(null),
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

import { Search } from '@/wailsjs/go/wails/App'

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


describe('SearchView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders search input', () => {
    render(<SearchView />)
    expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
  })

  it('renders search icon', () => {
    render(<SearchView />)
    expect(screen.getByTestId('search-icon')).toBeInTheDocument()
  })

  it('calls Search binding when typing', async () => {
    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test query')

    await waitFor(() => {
      expect(Search).toHaveBeenCalled()
    })
  })

  it('displays search results', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Another entry', Type: 'note', CreatedAt: '2024-01-14T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test entry')).toBeInTheDocument()
      expect(screen.getByText('Another entry')).toBeInTheDocument()
    })
  })

  it('shows entry type symbols in results', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'task')

    await waitFor(() => {
      expect(screen.getByText('•')).toBeInTheDocument()
    })
  })

  it('shows date in results', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText(/jan 15/i)).toBeInTheDocument()
    })
  })

  it('shows no results message when search returns empty', async () => {
    vi.mocked(Search).mockResolvedValue([])

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'nonexistent')

    await waitFor(() => {
      expect(screen.getByText(/no results found/i)).toBeInTheDocument()
    })
  })

  it('shows initial state message when no search performed', () => {
    render(<SearchView />)
    expect(screen.getByText(/enter a search term/i)).toBeInTheDocument()
  })

  it('clears results when search input is cleared', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test entry')).toBeInTheDocument()
    })

    await user.clear(input)

    await waitFor(() => {
      expect(screen.queryByText('Test entry')).not.toBeInTheDocument()
      expect(screen.getByText(/enter a search term/i)).toBeInTheDocument()
    })
  })

  it('shows entry ID on hover', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('#42')).toBeInTheDocument()
    })
  })
})

describe('SearchView simplified UI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not render context popover', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test', Type: 'task', ParentID: 2, CreatedAt: '2026-01-25T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)
    const input = screen.getByPlaceholderText('Search entries...')
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
    })
  })

  it('shows context dot for entries with parents', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Child entry', Type: 'task', ParentID: 99, CreatedAt: '2026-01-25T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)
    const input = screen.getByPlaceholderText('Search entries...')
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTestId('context-dot')).toBeInTheDocument()
    })
  })

  it('does not show context dot for root entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Root entry', Type: 'task', ParentID: null, CreatedAt: '2026-01-25T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)
    const input = screen.getByPlaceholderText('Search entries...')
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Root entry')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('context-dot')).not.toBeInTheDocument()
  })

  it('does not show ContextPill', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test', Type: 'task', ParentID: 2, CreatedAt: '2026-01-25T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)
    const input = screen.getByPlaceholderText('Search entries...')
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('context-pill')).not.toBeInTheDocument()
  })
})
