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

import { Search, GetEntryAncestors } from '@/wailsjs/go/wails/App'

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

describe('SearchView - Context Display', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows context when clicking on a search result with parent', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 2, Content: 'Child task', Type: 'task', ParentID: 1, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Parent event', Type: 'event', ParentID: null, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'child')

    await waitFor(() => {
      expect(screen.getByText('Child task')).toBeInTheDocument()
    })

    // Click on the result to show context
    await user.click(screen.getByText('Child task'))

    await waitFor(() => {
      expect(screen.getByText('Parent event')).toBeInTheDocument()
    })
  })

  it('hides context when clicking on expanded result again', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 2, Content: 'Child task', Type: 'task', ParentID: 1, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Parent event', Type: 'event', ParentID: null, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'child')

    await waitFor(() => {
      expect(screen.getByText('Child task')).toBeInTheDocument()
    })

    // Click to expand
    await user.click(screen.getByText('Child task'))
    await waitFor(() => {
      expect(screen.getByText('Parent event')).toBeInTheDocument()
    })

    // Click again to collapse
    await user.click(screen.getByText('Child task'))
    await waitFor(() => {
      expect(screen.queryByText('Parent event')).not.toBeInTheDocument()
    })
  })

  it('shows multi-level context for deeply nested entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 3, Content: 'Grandchild task', Type: 'task', ParentID: 2, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Grandparent event', Type: 'event', ParentID: null, CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Parent note', Type: 'note', ParentID: 1, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'grandchild')

    await waitFor(() => {
      expect(screen.getByText('Grandchild task')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Grandchild task'))

    await waitFor(() => {
      expect(screen.getByText('Grandparent event')).toBeInTheDocument()
      expect(screen.getByText('Parent note')).toBeInTheDocument()
    })
  })

  it('indents ancestors to show hierarchy', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 3, Content: 'Grandchild task', Type: 'task', ParentID: 2, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Grandparent event', Type: 'event', ParentID: null, CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Parent note', Type: 'note', ParentID: 1, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'grandchild')

    await waitFor(() => {
      expect(screen.getByText('Grandchild task')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Grandchild task'))

    await waitFor(() => {
      // Root ancestor (grandparent) should have no indentation
      const grandparentRow = screen.getByText('Grandparent event').closest('div')
      expect(grandparentRow).toHaveStyle({ paddingLeft: '0px' })

      // Second level (parent) should be indented
      const parentRow = screen.getByText('Parent note').closest('div')
      expect(parentRow).toHaveStyle({ paddingLeft: '20px' })
    })
  })

  it('indents the main result to continue hierarchy', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 3, Content: 'Grandchild task', Type: 'task', ParentID: 2, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Grandparent event', Type: 'event', ParentID: null, CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Parent note', Type: 'note', ParentID: 1, CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'grandchild')

    await waitFor(() => {
      expect(screen.getByText('Grandchild task')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Grandchild task'))

    await waitFor(() => {
      // Main result should be indented one level deeper than parent (2 ancestors = 40px)
      const mainResultRow = screen.getByText('Grandchild task').closest('[data-result-id]')
      expect(mainResultRow).toHaveStyle({ paddingLeft: '40px' })
    })
  })
})
