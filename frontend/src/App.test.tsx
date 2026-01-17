import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

const mockEntriesAgenda = createMockAgenda({
  Days: [createMockDayEntries({
    Entries: [
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'First task', CreatedAt: '2026-01-17T10:00:00Z' }),
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Second task', CreatedAt: '2026-01-17T11:00:00Z' }),
      createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Note', Content: 'A note', CreatedAt: '2026-01-17T12:00:00Z' }),
    ],
  })],
  Overdue: [],
})

const mockSearchResults = [
  createMockEntry({ ID: 10, EntityID: 'e10', Type: 'Task', Content: 'Buy groceries', CreatedAt: '2026-01-15T10:00:00Z' }),
  createMockEntry({ ID: 11, EntityID: 'e11', Type: 'Note', Content: 'Grocery list ideas', CreatedAt: '2026-01-14T10:00:00Z' }),
]

vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn().mockResolvedValue({
    Overdue: [],
    Days: [{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
  }),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  Search: vi.fn().mockResolvedValue([]),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
}))

import { GetAgenda, AddEntry, MarkEntryDone, Search, EditEntry, DeleteEntry, HasChildren } from './wailsjs/go/wails/App'

describe('App - AddEntryBar integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls AddEntry binding when adding entry via AddEntryBar', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText("What's on your mind?")
    fireEvent.change(input, { target: { value: 'Test task' } })

    const submitButton = screen.getByRole('button', { name: '' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalledWith('. Test task', expect.any(String))
    })
  })

  it('refreshes data after adding entry', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const input = screen.getByPlaceholderText("What's on your mind?")
    fireEvent.change(input, { target: { value: 'Test task' } })

    const submitButton = screen.getByRole('button', { name: '' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })

  it('clears input after successful add', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText("What's on your mind?") as HTMLInputElement
    fireEvent.change(input, { target: { value: 'Test task' } })

    const submitButton = screen.getByRole('button', { name: '' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(input.value).toBe('')
    })
  })
})

describe('App - Keyboard Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('pressing j moves selection down', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('j')

    const secondTask = screen.getByText('Second task').closest('[data-entry-id]')
    expect(secondTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing k moves selection up', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('jk')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing down arrow moves selection down', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}')

    const secondTask = screen.getByText('Second task').closest('[data-entry-id]')
    expect(secondTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing up arrow moves selection up', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}{ArrowUp}')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing Space toggles done on selected task', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard(' ')

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(1)
    })
  })

  it('does not go above first entry when pressing k at top', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('kkk')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('does not go below last entry when pressing j at bottom', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('jjjjj')

    const note = screen.getByText('A note').closest('[data-entry-id]')
    expect(note).toHaveAttribute('data-selected', 'true')
  })
})

describe('App - Search functionality', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    vi.mocked(Search).mockResolvedValue(mockSearchResults)
  })

  it('calls Search binding when typing in search input', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search entries...')
    await user.type(searchInput, 'groceries')

    await waitFor(() => {
      expect(Search).toHaveBeenCalledWith('groceries')
    }, { timeout: 1000 })
  })

  it('displays search results in dropdown', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search entries...')
    await user.type(searchInput, 'groceries')

    await waitFor(() => {
      expect(screen.getByText('Buy groceries')).toBeInTheDocument()
    })
  })

  it('clears search results when input is cleared', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search entries...')
    await user.type(searchInput, 'groceries')

    await waitFor(() => {
      expect(screen.getByText('Buy groceries')).toBeInTheDocument()
    })

    await user.clear(searchInput)

    await waitFor(() => {
      expect(screen.queryByText('Buy groceries')).not.toBeInTheDocument()
    })
  })
})

describe('App - Edit Entry', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('pressing e opens edit modal for selected entry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('e')

    await waitFor(() => {
      expect(screen.getByText('Edit Entry')).toBeInTheDocument()
      expect(screen.getByDisplayValue('First task')).toBeInTheDocument()
    })
  })

  it('calls EditEntry binding when saving edit', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('e')

    await waitFor(() => {
      expect(screen.getByDisplayValue('First task')).toBeInTheDocument()
    })

    const input = screen.getByDisplayValue('First task')
    await user.clear(input)
    await user.type(input, 'Updated task')

    fireEvent.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => {
      expect(EditEntry).toHaveBeenCalledWith(1, 'Updated task')
    })
  })

  it('closes edit modal on cancel', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('e')

    await waitFor(() => {
      expect(screen.getByText('Edit Entry')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))

    await waitFor(() => {
      expect(screen.queryByText('Edit Entry')).not.toBeInTheDocument()
    })
  })
})

describe('App - Delete Entry', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    vi.mocked(HasChildren).mockResolvedValue(false)
  })

  it('pressing d opens delete confirmation for selected entry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('d')

    await waitFor(() => {
      expect(screen.getByText('Delete Entry')).toBeInTheDocument()
    })
  })

  it('calls DeleteEntry binding when confirming delete', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('d')

    await waitFor(() => {
      expect(screen.getByText('Delete Entry')).toBeInTheDocument()
    })

    const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
    const dialogDeleteButton = deleteButtons.find(btn => btn.textContent === 'Delete')
    expect(dialogDeleteButton).toBeDefined()
    fireEvent.click(dialogDeleteButton!)

    await waitFor(() => {
      expect(DeleteEntry).toHaveBeenCalledWith(1)
    })
  })

  it('shows warning when entry has children', async () => {
    vi.mocked(HasChildren).mockResolvedValue(true)
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('d')

    await waitFor(() => {
      expect(screen.getByText(/will also delete/i)).toBeInTheDocument()
    })
  })

  it('closes delete dialog on cancel', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('d')

    await waitFor(() => {
      expect(screen.getByText('Delete Entry')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))

    await waitFor(() => {
      expect(screen.queryByText('Delete Entry')).not.toBeInTheDocument()
    })
  })
})
