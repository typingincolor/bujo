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

import { Search, GetEntryAncestors, MarkEntryDone, MarkEntryUndone } from '@/wailsjs/go/wails/App'

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

describe('SearchView - Actions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows tick button for task entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Mark done')).toBeInTheDocument()
    })
  })

  it('shows untick button for done entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Done task', Type: 'done', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'done')

    await waitFor(() => {
      expect(screen.getByTitle('Mark undone')).toBeInTheDocument()
    })
  })

  it('calls MarkEntryDone when tick button is clicked', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Mark done')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Mark done'))

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(42)
    })
  })

  it('calls MarkEntryUndone when untick button is clicked', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Done task', Type: 'done', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'done')

    await waitFor(() => {
      expect(screen.getByTitle('Mark undone')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Mark undone'))

    await waitFor(() => {
      expect(MarkEntryUndone).toHaveBeenCalledWith(42)
    })
  })

  it('shows checkmark symbol for done entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Done task', Type: 'done', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'done')

    await waitFor(() => {
      const undoneButton = screen.getByTitle('Mark undone')
      expect(undoneButton).toHaveTextContent('✓')
    })
  })

  it('shows cancel button for non-cancelled entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Cancel entry')).toBeInTheDocument()
    })
  })

  it('shows uncancel button for cancelled entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Cancelled task', Type: 'cancelled', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'cancelled')

    await waitFor(() => {
      expect(screen.getByTitle('Uncancel entry')).toBeInTheDocument()
    })
  })

  it('renders cancelled entries with strikethrough style', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Cancelled task', Type: 'cancelled', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'cancelled')

    await waitFor(() => {
      const content = screen.getByText('Cancelled task')
      expect(content).toHaveClass('line-through')
    })
  })

  it('renders done entries with success color (not strikethrough)', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Done task', Type: 'done', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'done')

    await waitFor(() => {
      const content = screen.getByText('Done task')
      expect(content).not.toHaveClass('line-through')
      expect(content).toHaveClass('text-bujo-done')
    })
  })

  it('calls CancelEntry when cancel button is clicked', async () => {
    const { CancelEntry } = await import('@/wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Cancel entry')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Cancel entry'))

    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('calls UncancelEntry when uncancel button is clicked', async () => {
    const { UncancelEntry } = await import('@/wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Cancelled task', Type: 'cancelled', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'cancelled')

    await waitFor(() => {
      expect(screen.getByTitle('Uncancel entry')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Uncancel entry'))

    await waitFor(() => {
      expect(UncancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('shows delete button for all entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Delete entry')).toBeInTheDocument()
    })
  })

  it('shows edit button for non-cancelled entries when onEdit provided', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView onEdit={vi.fn()} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Edit entry')).toBeInTheDocument()
    })
  })

  it('does not show edit button for cancelled entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test cancelled', Type: 'cancelled', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView onEdit={vi.fn()} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test cancelled')).toBeInTheDocument()
    })

    expect(screen.queryByTitle('Edit entry')).not.toBeInTheDocument()
  })

  it('shows migrate button for task entries when onMigrate provided', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView onMigrate={vi.fn()} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Migrate entry')).toBeInTheDocument()
    })
  })

  it('calls onMigrate when migrate button is clicked', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    const onMigrate = vi.fn()
    render(<SearchView onMigrate={onMigrate} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Migrate entry')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Migrate entry'))

    expect(onMigrate).toHaveBeenCalledWith(expect.objectContaining({ id: 42, type: 'task', content: 'Test task' }))
  })

  it('shows priority button for all entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Cycle priority')).toBeInTheDocument()
    })
  })

  it('displays priority indicator for entries with priority', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 1, Content: 'High priority task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), Priority: 'high' },
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'priority')

    await waitFor(() => {
      expect(screen.getByText('!!!')).toBeInTheDocument()
    })
  })

  it('updates priority indicator after cycling priority', async () => {
    const { CyclePriority, GetEntry } = await import('@/wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), Priority: 'none' },
    ] as never)
    vi.mocked(GetEntry).mockResolvedValue({
      ...createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      Priority: 'low',
    } as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Cycle priority')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Cycle priority'))

    await waitFor(() => {
      expect(CyclePriority).toHaveBeenCalledWith(42)
      expect(screen.getByText('!')).toBeInTheDocument()
    })
  })
})

describe('SearchView - Keyboard Shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('selects first result with j key when search results exist', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'First result', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Second result', Type: 'task', CreatedAt: '2024-01-14T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('First result')).toBeInTheDocument()
    })

    // Blur the input to allow keyboard navigation
    await user.tab()
    await user.keyboard('j')

    await waitFor(() => {
      const firstResult = screen.getByText('First result').closest('[data-result-id]')
      expect(firstResult?.parentElement).toHaveClass('ring-2')
    })
  })

  it('navigates down with j key', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'First result', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Second result', Type: 'task', CreatedAt: '2024-01-14T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('First result')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('jj') // Press j twice to select second result

    await waitFor(() => {
      const secondResult = screen.getByText('Second result').closest('[data-result-id]')
      expect(secondResult?.parentElement).toHaveClass('ring-2')
    })
  })

  it('navigates up with k key', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'First result', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Second result', Type: 'task', CreatedAt: '2024-01-14T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('First result')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('jjk') // Go down twice, then up once

    await waitFor(() => {
      const firstResult = screen.getByText('First result').closest('[data-result-id]')
      expect(firstResult?.parentElement).toHaveClass('ring-2')
    })
  })

  it('navigates with arrow keys', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'First result', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, Content: 'Second result', Type: 'task', CreatedAt: '2024-01-14T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('First result')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('{ArrowDown}{ArrowDown}')

    await waitFor(() => {
      const secondResult = screen.getByText('Second result').closest('[data-result-id]')
      expect(secondResult?.parentElement).toHaveClass('ring-2')
    })
  })

  it('toggles done with Space key for selected task', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test task')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('j ') // Select first, then Space

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(42)
    })
  })

  it('toggles undone with Space key for selected done entry', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Done task', Type: 'done', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'done')

    await waitFor(() => {
      expect(screen.getByText('Done task')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('j ') // Select first, then Space

    await waitFor(() => {
      expect(MarkEntryUndone).toHaveBeenCalledWith(42)
    })
  })

  it('cancels entry with x key', async () => {
    const { CancelEntry } = await import('@/wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test task')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('jx') // Select first, then x

    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('uncancels entry with x key when cancelled', async () => {
    const { UncancelEntry } = await import('@/wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Cancelled task', Type: 'cancelled', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'cancelled')

    await waitFor(() => {
      expect(screen.getByText('Cancelled task')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('jx') // Select first, then x

    await waitFor(() => {
      expect(UncancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('cycles priority with p key', async () => {
    const { CyclePriority } = await import('@/wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test task')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('jp') // Select first, then p

    await waitFor(() => {
      expect(CyclePriority).toHaveBeenCalledWith(42)
    })
  })

  it('cycles type with t key', async () => {
    const { RetypeEntry } = await import('@/wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test task')).toBeInTheDocument()
    })

    await user.tab()
    await user.keyboard('jt') // Select first, then t

    await waitFor(() => {
      expect(RetypeEntry).toHaveBeenCalledWith(42, 'note')
    })
  })

  it('expands context with Enter key', async () => {
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

    await user.tab()
    await user.keyboard('j{Enter}') // Select first, then Enter

    await waitFor(() => {
      expect(screen.getByText('Parent event')).toBeInTheDocument()
    })
  })

  it('does not trigger shortcuts when typing in search input', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'jjjpptx') // These should be part of the search query, not shortcuts

    await waitFor(() => {
      expect(Search).toHaveBeenLastCalledWith('jjjpptx')
    })

    // No actions should have been triggered
    expect(MarkEntryDone).not.toHaveBeenCalled()
  })
})

describe('SearchView - Context Pill', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows context pill when entry has parent and is not expanded', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 1, Content: 'Child entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), ParentID: 5 },
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'child')

    await waitFor(() => {
      const pill = screen.getByTestId('context-pill')
      expect(pill).toBeInTheDocument()
    })
  })

  it('does not show context pill when entry has no parent', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Root entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z', ParentID: null }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'root')

    await waitFor(() => {
      expect(screen.getByText('Root entry')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('context-pill')).not.toBeInTheDocument()
  })

  it('hides context pill when entry is expanded', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 1, Content: 'Child entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), ParentID: 5 },
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 5, Content: 'Parent entry', Type: 'note', CreatedAt: '2024-01-14T10:00:00Z', ParentID: null }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'child')

    await waitFor(() => {
      expect(screen.getByTestId('context-pill')).toBeInTheDocument()
    })

    // Click to expand
    await user.click(screen.getByText('Child entry'))

    await waitFor(() => {
      expect(screen.getByText('Parent entry')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('context-pill')).not.toBeInTheDocument()
  })

  it('clicking context pill toggles expand/collapse', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 1, Content: 'Child entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), ParentID: 5 },
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 5, Content: 'Parent entry', Type: 'note', CreatedAt: '2024-01-14T10:00:00Z', ParentID: null }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'child')

    await waitFor(() => {
      expect(screen.getByTestId('context-pill')).toBeInTheDocument()
    })

    // Click pill to expand
    await user.click(screen.getByTestId('context-pill'))

    await waitFor(() => {
      expect(screen.getByText('Parent entry')).toBeInTheDocument()
    })
  })

  it('context pill click does not trigger other entry actions', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 1, Content: 'Child task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), ParentID: 5 },
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 5, Content: 'Parent entry', Type: 'note', CreatedAt: '2024-01-14T10:00:00Z', ParentID: null }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'child')

    await waitFor(() => {
      expect(screen.getByTestId('context-pill')).toBeInTheDocument()
    })

    // Click pill - should only expand, not mark done
    await user.click(screen.getByTestId('context-pill'))

    await waitFor(() => {
      expect(screen.getByText('Parent entry')).toBeInTheDocument()
    })
    expect(MarkEntryDone).not.toHaveBeenCalled()
  })

  it('shows ancestor count in pill after loading', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 1, Content: 'Child entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), ParentID: 5 },
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 5, Content: 'Parent', Type: 'note', CreatedAt: '2024-01-14T10:00:00Z', ParentID: null }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'child')

    await waitFor(() => {
      const pill = screen.getByTestId('context-pill')
      expect(pill).toHaveTextContent('1')
    })
  })

  it('shows correct count for deeply nested entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      { ...createMockEntry({ ID: 1, Content: 'Nested child', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }), ParentID: 3 },
    ] as never)
    vi.mocked(GetEntryAncestors).mockResolvedValue([
      createMockEntry({ ID: 2, Content: 'Root', Type: 'event', CreatedAt: '2024-01-13T10:00:00Z', ParentID: null }),
      createMockEntry({ ID: 3, Content: 'Parent', Type: 'note', CreatedAt: '2024-01-14T10:00:00Z', ParentID: 2 }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'nested')

    await waitFor(() => {
      const pill = screen.getByTestId('context-pill')
      expect(pill).toHaveTextContent('2')
    })
  })
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

describe('SearchView - Move to List', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows move to list button for task entries when onMoveToList provided', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView onMoveToList={vi.fn()} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Move to list')).toBeInTheDocument()
    })
  })

  it('does not show move to list button for non-task entries', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test note', Type: 'note', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView onMoveToList={vi.fn()} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test note')).toBeInTheDocument()
    })

    expect(screen.queryByTitle('Move to list')).not.toBeInTheDocument()
  })

  it('does not show move to list button when onMoveToList not provided', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test task')).toBeInTheDocument()
    })

    expect(screen.queryByTitle('Move to list')).not.toBeInTheDocument()
  })

  it('calls onMoveToList when move to list button is clicked', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const onMoveToList = vi.fn()
    const user = userEvent.setup()
    render(<SearchView onMoveToList={onMoveToList} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Move to list')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Move to list'))

    expect(onMoveToList).toHaveBeenCalledWith(expect.objectContaining({ id: 42, type: 'task', content: 'Test task' }))
  })
})

describe('SearchView - Navigate to Entry Button', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows go to date button when onNavigateToEntry provided', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 1, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const user = userEvent.setup()
    render(<SearchView onNavigateToEntry={vi.fn()} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Go to date')).toBeInTheDocument()
    })
  })

  it('does not show go to date button when onNavigateToEntry not provided', async () => {
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

    expect(screen.queryByTitle('Go to date')).not.toBeInTheDocument()
  })

  it('calls onNavigateToEntry when go to date button is clicked', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const onNavigateToEntry = vi.fn()
    const user = userEvent.setup()
    render(<SearchView onNavigateToEntry={onNavigateToEntry} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Go to date')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Go to date'))

    expect(onNavigateToEntry).toHaveBeenCalledWith(expect.objectContaining({ id: 42, type: 'task', content: 'Test entry' }))
  })

  it('go to date button works independently of double-click navigation', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test entry', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const onNavigateToEntry = vi.fn()
    const user = userEvent.setup()
    render(<SearchView onNavigateToEntry={onNavigateToEntry} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByTitle('Go to date')).toBeInTheDocument()
    })

    // First use the button
    await user.click(screen.getByTitle('Go to date'))
    expect(onNavigateToEntry).toHaveBeenCalledTimes(1)

    // Then also verify double-click still works
    const result = screen.getByText('Test entry').closest('[data-result-id]')
    await user.dblClick(result!)
    expect(onNavigateToEntry).toHaveBeenCalledTimes(2)
  })
})

describe('SearchView - Add Child Callback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('accepts onAddChild prop without error', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 42, Content: 'Test task', Type: 'task', CreatedAt: '2024-01-15T10:00:00Z' }),
    ] as never)

    const onAddChild = vi.fn()
    const user = userEvent.setup()
    // Should render without error when onAddChild is provided
    render(<SearchView onAddChild={onAddChild} />)

    const input = screen.getByPlaceholderText(/search entries/i)
    await user.type(input, 'test')

    await waitFor(() => {
      expect(screen.getByText('Test task')).toBeInTheDocument()
    })

    // Verify EntryActionBar is rendered (callbacks are wired internally)
    expect(screen.getByTestId('entry-action-bar')).toBeInTheDocument()
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
      expect(screen.getByTitle('Mark done')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Mark done'))

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
      expect(screen.getByTitle('Mark undone')).toBeInTheDocument()
    })

    await user.click(screen.getByTitle('Mark undone'))

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
      const symbolButton = screen.getByTitle('Mark done')
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
      const symbolButton = screen.getByTitle('Mark undone')
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

    // Symbol for cancelled entries should not be a button with mark done/undone title
    expect(screen.queryByTitle('Mark done')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Mark undone')).not.toBeInTheDocument()
  })
})
