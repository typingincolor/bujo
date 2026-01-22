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
