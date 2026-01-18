import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
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
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  CreateHabit: vi.fn().mockResolvedValue(1),
}))

import { GetAgenda, GetHabits, AddEntry, MarkEntryDone, Search, EditEntry, DeleteEntry, HasChildren, CancelEntry, UncancelEntry, CyclePriority, MigrateEntry } from './wailsjs/go/wails/App'

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

    // Find the Cancel button in the modal (not the Cancel entry buttons)
    const cancelButtons = screen.getAllByRole('button', { name: /cancel/i })
    const modalCancelButton = cancelButtons.find(btn => btn.textContent === 'Cancel')
    expect(modalCancelButton).toBeDefined()
    fireEvent.click(modalCancelButton!)

    await waitFor(() => {
      expect(screen.queryByText('Edit Entry')).not.toBeInTheDocument()
    })
  })
})

describe('App - QuickStats', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('renders QuickStats component in today view', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByText('Tasks Completed')).toBeInTheDocument()
    expect(screen.getByText('Pending Tasks')).toBeInTheDocument()
    expect(screen.getByText('Habits Today')).toBeInTheDocument()
    expect(screen.getByText('Monthly Goals')).toBeInTheDocument()
  })
})

describe('App - Day Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('renders prev/next day navigation buttons in today view', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: /previous day/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /next day/i })).toBeInTheDocument()
  })

  it('renders date picker in today view', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByLabelText(/pick date/i)).toBeInTheDocument()
  })

  it('changing date picker navigates to selected date', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const datePicker = screen.getByLabelText(/pick date/i)
    fireEvent.change(datePicker, { target: { value: '2026-01-20' } })

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })

  it('clicking next day navigates to next day', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const nextButton = screen.getByRole('button', { name: /next day/i })
    await user.click(nextButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })

  it('clicking prev day navigates to previous day', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })

  it('pressing h navigates to previous day', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    await user.keyboard('h')

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })

  it('pressing l navigates to next day', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    await user.keyboard('l')

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })
})

describe('App - Habit View Toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('refetches habits with different day count when period changes', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Switch to habits view
    const habitsButton = screen.getByRole('button', { name: /habits/i })
    await user.click(habitsButton)

    await waitFor(() => {
      expect(screen.getByText('Habit Tracker')).toBeInTheDocument()
    })

    vi.mocked(GetHabits).mockClear()

    // Click on period selector (shows "week" by default) - it's inside the HabitTracker component
    const periodButton = screen.getByRole('button', { name: /^week$/i })
    await user.click(periodButton)

    // Click on Month option
    const monthButton = screen.getByRole('button', { name: /^month$/i })
    await user.click(monthButton)

    await waitFor(() => {
      expect(GetHabits).toHaveBeenCalledWith(45)
    })
  })

  it('pressing w key cycles habit period from week to month', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Switch to habits view
    const habitsButton = screen.getByRole('button', { name: /habits/i })
    await user.click(habitsButton)

    await waitFor(() => {
      expect(screen.getByText('Habit Tracker')).toBeInTheDocument()
    })

    // Verify we're in week view
    expect(screen.getByRole('button', { name: /^week$/i })).toBeInTheDocument()

    vi.mocked(GetHabits).mockClear()

    // Press 'w' key to cycle to month view - dispatch proper KeyboardEvent
    const event = new KeyboardEvent('keydown', {
      key: 'w',
      bubbles: true,
      cancelable: true,
    })
    await act(async () => {
      window.dispatchEvent(event)
    })

    await waitFor(() => {
      expect(GetHabits).toHaveBeenCalledWith(45)
    }, { timeout: 2000 })

    // Period button should now show 'month'
    expect(screen.getByRole('button', { name: /^month$/i })).toBeInTheDocument()
  })
})

describe('App - Keyboard Shortcuts Panel Toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keyboard shortcuts panel is hidden by default', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Keyboard shortcuts panel should not be visible by default
    expect(screen.queryByText('Keyboard Shortcuts')).not.toBeInTheDocument()
  })

  it('? key toggles keyboard shortcuts panel visibility', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Initially hidden
    expect(screen.queryByText('Keyboard Shortcuts')).not.toBeInTheDocument()

    // Press ? to show (no modifier key needed)
    const showEvent = new KeyboardEvent('keydown', {
      key: '?',
      bubbles: true,
      cancelable: true,
    })
    await act(async () => {
      window.dispatchEvent(showEvent)
    })

    // Should now be visible
    await waitFor(() => {
      expect(screen.getByText('Keyboard Shortcuts')).toBeInTheDocument()
    })

    // Press ? again to hide
    const hideEvent = new KeyboardEvent('keydown', {
      key: '?',
      bubbles: true,
      cancelable: true,
    })
    await act(async () => {
      window.dispatchEvent(hideEvent)
    })

    // Should be hidden again
    await waitFor(() => {
      expect(screen.queryByText('Keyboard Shortcuts')).not.toBeInTheDocument()
    })
  })
})

describe('App - Cancel/Uncancel Entry', () => {
  const mockWithCancelledEntry = createMockAgenda({
    Days: [createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Active task', CreatedAt: '2026-01-17T10:00:00Z' }),
        createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Cancelled', Content: 'Cancelled task', CreatedAt: '2026-01-17T11:00:00Z' }),
      ],
    })],
    Overdue: [],
  })

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('clicking cancel button calls CancelEntry binding', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const cancelButton = screen.getAllByTitle('Cancel entry')[0]
    fireEvent.click(cancelButton)

    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(1)
    })
  })

  it('clicking uncancel button calls UncancelEntry binding', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockWithCancelledEntry)
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('Cancelled task')).toBeInTheDocument()
    })

    const uncancelButton = screen.getByTitle('Uncancel entry')
    fireEvent.click(uncancelButton)

    await waitFor(() => {
      expect(UncancelEntry).toHaveBeenCalledWith(2)
    })
  })

  it('refreshes data after cancelling entry', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const cancelButton = screen.getAllByTitle('Cancel entry')[0]
    fireEvent.click(cancelButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })

  it('refreshes data after uncancelling entry', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockWithCancelledEntry)
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('Cancelled task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const uncancelButton = screen.getByTitle('Uncancel entry')
    fireEvent.click(uncancelButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })
})

describe('App - Priority Cycling', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('clicking priority button calls CyclePriority binding', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const priorityButton = screen.getAllByTitle('Cycle priority')[0]
    fireEvent.click(priorityButton)

    await waitFor(() => {
      expect(CyclePriority).toHaveBeenCalledWith(1)
    })
  })

  it('refreshes data after cycling priority', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const priorityButton = screen.getAllByTitle('Cycle priority')[0]
    fireEvent.click(priorityButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
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

    // Find the Cancel button in the dialog (not the Cancel entry buttons)
    const cancelButtons = screen.getAllByRole('button', { name: /cancel/i })
    const dialogCancelButton = cancelButtons.find(btn => btn.textContent === 'Cancel')
    expect(dialogCancelButton).toBeDefined()
    fireEvent.click(dialogCancelButton!)

    await waitFor(() => {
      expect(screen.queryByText('Delete Entry')).not.toBeInTheDocument()
    })
  })
})

describe('App - Task Migration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('clicking migrate button opens migrate modal', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const migrateButton = screen.getAllByTitle('Migrate entry')[0]
    fireEvent.click(migrateButton)

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
    })
  })

  it('shows entry content in migrate modal', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const migrateButton = screen.getAllByTitle('Migrate entry')[0]
    fireEvent.click(migrateButton)

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
      // The modal shows the entry content in its message (uses smart quotes)
      expect(screen.getByText(/Migrate.*First task.*to a future date/)).toBeInTheDocument()
    })
  })

  it('calls MigrateEntry binding when confirming migration', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const migrateButton = screen.getAllByTitle('Migrate entry')[0]
    fireEvent.click(migrateButton)

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
    })

    // Find the date input and modal buttons within the modal
    const modal = document.querySelector('.fixed.inset-0')
    expect(modal).toBeTruthy()

    const dateInput = modal!.querySelector('input[type="date"]') as HTMLInputElement
    fireEvent.change(dateInput, { target: { value: '2026-01-25' } })

    // Click the Migrate submit button in the modal
    const migrateSubmitButton = modal!.querySelector('button[type="submit"]') as HTMLButtonElement
    fireEvent.click(migrateSubmitButton)

    await waitFor(() => {
      expect(MigrateEntry).toHaveBeenCalledWith(1, expect.any(String))
    })
  })

  it('closes migrate modal on cancel', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const migrateButton = screen.getAllByTitle('Migrate entry')[0]
    fireEvent.click(migrateButton)

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
    })

    // Find the Cancel button in the modal
    const cancelButtons = screen.getAllByRole('button', { name: /cancel/i })
    const modalCancelButton = cancelButtons.find(btn => btn.textContent === 'Cancel')
    expect(modalCancelButton).toBeDefined()
    fireEvent.click(modalCancelButton!)

    await waitFor(() => {
      expect(screen.queryByText('Migrate Entry')).not.toBeInTheDocument()
    })
  })

  it('refreshes data after migrating entry', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const migrateButton = screen.getAllByTitle('Migrate entry')[0]
    fireEvent.click(migrateButton)

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    // Find and click the submit button in the modal
    const modal = document.querySelector('.fixed.inset-0')
    expect(modal).toBeTruthy()
    const migrateSubmitButton = modal!.querySelector('button[type="submit"]') as HTMLButtonElement
    fireEvent.click(migrateSubmitButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })
})

describe('App - No flicker on data refresh', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('does not show loading spinner when refreshing data after habit action', async () => {
    // Track loading spinner appearances
    let loadingSpinnerShownDuringRefresh = false
    let initialLoadComplete = false

    render(<App />)

    // Wait for initial load to complete
    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })
    initialLoadComplete = true

    // Navigate to habits view
    const habitsButton = screen.getByRole('button', { name: /habits/i })
    fireEvent.click(habitsButton)

    // Verify we're in habits view
    await waitFor(() => {
      expect(screen.getByText('Habit Tracker')).toBeInTheDocument()
    })

    // Set up delayed mock to observe loading state during refresh
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let resolveGetHabits: (value: any) => void
    vi.mocked(GetHabits).mockImplementation(() => new Promise(resolve => {
      resolveGetHabits = resolve
      // Check loading state while API is in flight
      setTimeout(() => {
        if (initialLoadComplete && screen.queryByText('Loading your journal...')) {
          loadingSpinnerShownDuringRefresh = true
        }
      }, 10)
    }))

    // Trigger loadData by simulating habit creation (which calls onHabitChanged -> loadData)
    const addButton = screen.getByRole('button', { name: /add habit/i })
    fireEvent.click(addButton)

    // Find the input and submit a new habit
    const habitInput = screen.getByPlaceholderText('Habit name')
    fireEvent.change(habitInput, { target: { value: 'Test habit' } })
    fireEvent.keyDown(habitInput, { key: 'Enter' })

    // Small delay to let the API call start
    await new Promise(resolve => setTimeout(resolve, 50))

    // Resolve the pending API call
    resolveGetHabits!({ Habits: [] })

    // Wait for everything to settle
    await waitFor(() => {
      expect(screen.getByText('Habit Tracker')).toBeInTheDocument()
    })

    // CRITICAL: Loading spinner should NOT have appeared during refresh
    expect(loadingSpinnerShownDuringRefresh).toBe(false)
  })
})
