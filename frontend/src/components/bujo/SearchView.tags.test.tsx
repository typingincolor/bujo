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
  GetAllMentions: vi.fn().mockResolvedValue([]),
  SearchByMentions: vi.fn().mockResolvedValue([]),
}))

import { Search, SearchByTags, GetAllTags } from '@/wailsjs/go/wails/App'

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

describe('SearchView - Tag Content Display', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders tags in search results as styled spans', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Buy groceries #shopping', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'groceries')

    await waitFor(() => {
      const tag = screen.getByText('#shopping')
      expect(tag).toHaveClass('tag')
    })
  })
})

describe('SearchView - Tag Filtering', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls SearchByTags when initialTagFilter is provided', async () => {
    vi.mocked(SearchByTags).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task #work', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    render(<SearchView initialTagFilter="work" />)

    await waitFor(() => {
      expect(SearchByTags).toHaveBeenCalledWith(['work'])
    })
  })

  it('shows filtered-by indicator when tag filter is active', async () => {
    vi.mocked(SearchByTags).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task #work', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    render(<SearchView initialTagFilter="work" />)

    await waitFor(() => {
      expect(screen.getByText('#work')).toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
  })

  it('clears tag filter when clear button is clicked', async () => {
    vi.mocked(SearchByTags).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task #work', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView initialTagFilter="work" />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /clear/i }))

    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /clear/i })).not.toBeInTheDocument()
    })
  })

  it('displays results from SearchByTags', async () => {
    vi.mocked(SearchByTags).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Buy milk #shopping', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Buy eggs #shopping', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    render(<SearchView initialTagFilter="shopping" />)

    await waitFor(() => {
      expect(screen.getByText('Buy milk')).toBeInTheDocument()
      expect(screen.getByText('Buy eggs')).toBeInTheDocument()
    })
  })

  it('does not call Search when initialTagFilter is provided', async () => {
    vi.mocked(SearchByTags).mockResolvedValue([])

    render(<SearchView initialTagFilter="work" />)

    await waitFor(() => {
      expect(SearchByTags).toHaveBeenCalled()
    })
    expect(Search).not.toHaveBeenCalled()
  })
})

describe('SearchView - Tag Panel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches and displays all tags', async () => {
    vi.mocked(GetAllTags).mockResolvedValue(['work', 'personal', 'urgent'] as never)

    render(<SearchView />)

    await waitFor(() => {
      expect(screen.getByText('#work')).toBeInTheDocument()
      expect(screen.getByText('#personal')).toBeInTheDocument()
      expect(screen.getByText('#urgent')).toBeInTheDocument()
    })
  })

  it('filters by tag when clicking a tag pill', async () => {
    vi.mocked(GetAllTags).mockResolvedValue(['work', 'personal'] as never)
    vi.mocked(SearchByTags).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Task #work', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    await waitFor(() => {
      expect(screen.getByText('#work')).toBeInTheDocument()
    })

    await user.click(screen.getByText('#work'))

    await waitFor(() => {
      expect(SearchByTags).toHaveBeenCalledWith(['work'])
    })
  })

  it('does not render tag panel when no tags exist', async () => {
    vi.mocked(GetAllTags).mockResolvedValue([] as never)

    render(<SearchView />)

    await waitFor(() => {
      expect(GetAllTags).toHaveBeenCalled()
    })
    expect(screen.queryByTestId('tag-panel')).not.toBeInTheDocument()
  })
})
