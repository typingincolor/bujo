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

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn().mockReturnValue(() => {}),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}))

vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn().mockResolvedValue({
    Overdue: [],
    Days: [{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
  }),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  CreateHabit: vi.fn().mockResolvedValue(1),
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue(['Home', 'Office']),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
}))

import { GetAgenda, DeleteEntry, HasChildren, CancelEntry, UncancelEntry, CyclePriority, MigrateEntry } from './wailsjs/go/wails/App'


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
