import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchView } from './SearchView'

vi.mock('@/wailsjs/go/wails/App', () => ({
  Search: vi.fn().mockResolvedValue([]),
  SearchByTags: vi.fn().mockResolvedValue([]),
  SearchByMentions: vi.fn().mockResolvedValue([]),
  GetAllTags: vi.fn().mockResolvedValue([]),
  GetAllMentions: vi.fn().mockResolvedValue([]),
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

import { Search, SearchByMentions, GetAllMentions } from '@/wailsjs/go/wails/App'

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

describe('SearchView - Mention Content Display', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders @mentions in search results as styled spans', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Met with @john today', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'met')

    await waitFor(() => {
      const mention = screen.getByText('@john')
      expect(mention).toHaveClass('mention')
    })
  })
})

describe('SearchView - Mention Filtering', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls SearchByMentions when initialMentionFilter is provided', async () => {
    vi.mocked(SearchByMentions).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task with @john', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    render(<SearchView initialMentionFilter="john" />)

    await waitFor(() => {
      expect(SearchByMentions).toHaveBeenCalledWith(['john'])
    })
  })

  it('shows filtered-by indicator when mention filter is active', async () => {
    vi.mocked(SearchByMentions).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task with @john', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    render(<SearchView initialMentionFilter="john" />)

    await waitFor(() => {
      expect(screen.getByText(/@john/)).toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
  })

  it('clears mention filter when clear button is clicked', async () => {
    vi.mocked(SearchByMentions).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task with @john', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView initialMentionFilter="john" />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /clear/i }))

    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /clear/i })).not.toBeInTheDocument()
    })
  })

  it('displays results from SearchByMentions', async () => {
    vi.mocked(SearchByMentions).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Call @john about project', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Email @john re: meeting', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    render(<SearchView initialMentionFilter="john" />)

    await waitFor(() => {
      expect(screen.getByText('Call')).toBeInTheDocument()
      expect(screen.getByText('Email')).toBeInTheDocument()
    })
  })

  it('does not call Search when initialMentionFilter is provided', async () => {
    vi.mocked(SearchByMentions).mockResolvedValue([])

    render(<SearchView initialMentionFilter="john" />)

    await waitFor(() => {
      expect(SearchByMentions).toHaveBeenCalled()
    })
    expect(Search).not.toHaveBeenCalled()
  })
})

describe('SearchView - Mention Panel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches and displays all mentions', async () => {
    vi.mocked(GetAllMentions).mockResolvedValue(['john', 'alice', 'bob'] as never)

    render(<SearchView />)

    await waitFor(() => {
      expect(screen.getByText('@john')).toBeInTheDocument()
      expect(screen.getByText('@alice')).toBeInTheDocument()
      expect(screen.getByText('@bob')).toBeInTheDocument()
    })
  })

  it('filters by mention when clicking a mention pill', async () => {
    vi.mocked(GetAllMentions).mockResolvedValue(['john', 'alice'] as never)
    vi.mocked(SearchByMentions).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task with @john', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    await waitFor(() => {
      expect(screen.getByText('@john')).toBeInTheDocument()
    })

    await user.click(screen.getByText('@john'))

    await waitFor(() => {
      expect(SearchByMentions).toHaveBeenCalledWith(['john'])
    })
  })

  it('does not render mention panel when no mentions exist', async () => {
    vi.mocked(GetAllMentions).mockResolvedValue([] as never)

    render(<SearchView />)

    await waitFor(() => {
      expect(GetAllMentions).toHaveBeenCalled()
    })
    expect(screen.queryByTestId('mention-panel')).not.toBeInTheDocument()
  })
})
