import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
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

import { GetAgenda, MarkEntryDone, EditEntry } from './wailsjs/go/wails/App'


describe('App - Keyboard Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('pressing j moves selection down', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('j')

    const secondTask = screen.getByText('Second task').closest('[data-entry-id]')
    expect(secondTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing k moves selection up', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('jk')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing down arrow moves selection down', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}')

    const secondTask = screen.getByText('Second task').closest('[data-entry-id]')
    expect(secondTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing up arrow moves selection up', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}{ArrowUp}')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing Space toggles done on selected task', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('kkk')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('does not go below last entry when pressing j at bottom', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('jjjjj')

    const note = screen.getByText('A note').closest('[data-entry-id]')
    expect(note).toHaveAttribute('data-selected', 'true')
  })
})

describe('App - Click Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('clicking an entry updates the selection to that entry', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Initially first task is selected
    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')

    // Click on the note entry directly to select it
    const noteText = screen.getByText('A note')
    await user.click(noteText)

    // Now the note should be selected
    const noteEntry = screen.getByText('A note').closest('[data-entry-id]')
    await waitFor(() => {
      expect(noteEntry).toHaveAttribute('data-selected', 'true')
    })
    expect(firstTask).toHaveAttribute('data-selected', 'false')
  })

  it('clicking second task updates selection', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Click second task directly to select it
    const secondTaskText = screen.getByText('Second task')
    await user.click(secondTaskText)

    const secondTask = screen.getByText('Second task').closest('[data-entry-id]')
    await waitFor(() => {
      expect(secondTask).toHaveAttribute('data-selected', 'true')
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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByText('Tasks Completed')).toBeInTheDocument()
    expect(screen.getAllByText('Pending Tasks').length).toBeGreaterThan(0)
    expect(screen.getByText('Habits Today')).toBeInTheDocument()
    expect(screen.getAllByText(/monthly goals/i).length).toBeGreaterThan(0)
  })

  it('displays overdue count from agenda', async () => {
    const agendaWithOverdue = createMockAgenda({
      Days: [createMockDayEntries({
        Entries: [createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Today task' })],
      })],
      Overdue: [
        createMockEntry({ ID: 10, EntityID: 'e10', Type: 'Task', Content: 'Overdue task 1' }),
        createMockEntry({ ID: 11, EntityID: 'e11', Type: 'Task', Content: 'Overdue task 2' }),
        createMockEntry({ ID: 12, EntityID: 'e12', Type: 'Task', Content: 'Overdue task 3' }),
      ],
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaWithOverdue)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Today task')).toBeInTheDocument()
    })

    // Should show 3 overdue tasks in the QuickStats card
    const pendingTasksCard = screen.getByTestId('stat-card-pending-tasks')
    expect(pendingTasksCard).toHaveTextContent('3')
  })
})
