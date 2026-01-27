import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockDays, createMockOverdue } from './test/mocks'

const mockDays = createMockDays([createMockDayEntries({
  Entries: [
    createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'First task', CreatedAt: '2026-01-17T10:00:00Z' }),
    createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Second task', CreatedAt: '2026-01-17T11:00:00Z' }),
    createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Note', Content: 'A note', CreatedAt: '2026-01-17T12:00:00Z' }),
  ],
})])
const mockOverdue = createMockOverdue([])

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn().mockReturnValue(() => {}),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}))

vi.mock('./wailsjs/go/wails/App', () => ({
  GetDayEntries: vi.fn().mockResolvedValue([{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }]),
  GetOverdue: vi.fn().mockResolvedValue([]),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  GetEntryContext: vi.fn().mockResolvedValue([]),
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

import { GetDayEntries, GetOverdue, DeleteEntry, HasChildren, CancelEntry, UncancelEntry, CyclePriority, MigrateEntry, MarkEntryDone } from './wailsjs/go/wails/App'


describe('App - Cancel/Uncancel Entry', () => {
  const mockDaysWithCancelledEntry = createMockDays([createMockDayEntries({
    Entries: [
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Active task', CreatedAt: '2026-01-17T10:00:00Z' }),
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Cancelled', Content: 'Cancelled task', CreatedAt: '2026-01-17T11:00:00Z' }),
    ],
  })])

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('clicking cancel button calls CancelEntry binding', async () => {
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysWithCancelledEntry)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    const cancelButton = screen.getAllByTitle('Cancel entry')[0]
    fireEvent.click(cancelButton)

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })

  it('refreshes data after uncancelling entry', async () => {
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysWithCancelledEntry)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Cancelled task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    const uncancelButton = screen.getByTitle('Uncancel entry')
    fireEvent.click(uncancelButton)

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })
})

describe('App - Priority Cycling', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('clicking priority button calls CyclePriority binding', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    const priorityButton = screen.getAllByTitle('Cycle priority')[0]
    fireEvent.click(priorityButton)

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })
})

describe('App - Delete Entry', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
    vi.mocked(HasChildren).mockResolvedValue(false)
  })

  it('pressing d opens delete confirmation for selected entry', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('clicking migrate button opens migrate modal', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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

    // Submit the form
    const form = dateInput.closest('form') as HTMLFormElement
    fireEvent.submit(form)

    await waitFor(() => {
      expect(MigrateEntry).toHaveBeenCalledWith(1, expect.any(String))
    })
  })

  it('closes migrate modal on cancel', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const migrateButton = screen.getAllByTitle('Migrate entry')[0]
    fireEvent.click(migrateButton)

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    // Find and click the submit button in the modal
    const modal = document.querySelector('.fixed.inset-0')
    expect(modal).toBeTruthy()
    const migrateSubmitButton = modal!.querySelector('button[type="submit"]') as HTMLButtonElement
    fireEvent.click(migrateSubmitButton)

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })
})

describe('App - Sidebar Entry Visibility After Actions', () => {
  const mockDaysWithOverdue = createMockDays([createMockDayEntries({
    Entries: [
      createMockEntry({ ID: 100, EntityID: 'e100', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
    ],
  })])
  const mockOverdueEntries = createMockOverdue([
    createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Overdue task 1', CreatedAt: '2026-01-15T10:00:00Z' }),
    createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Overdue task 2', CreatedAt: '2026-01-15T11:00:00Z' }),
  ])

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysWithOverdue)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdueEntries)
  })

  it('keeps entry visible in sidebar after marking done', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for main panel to render first
    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar (starts collapsed by default)
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    // Wait for sidebar to render with overdue entries
    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Verify both tasks are visible in sidebar
    expect(screen.getByText('Overdue task 2')).toBeInTheDocument()

    // Mock the reload to return the entry as done (would normally disappear)
    const mockDaysAfterMarkDone = createMockDays([createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 100, EntityID: 'e100', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Done', Content: 'Overdue task 1', CreatedAt: '2026-01-15T10:00:00Z' }),
      ],
    })])
    const mockOverdueAfterMarkDone = createMockOverdue([
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Overdue task 2', CreatedAt: '2026-01-15T11:00:00Z' }),
    ])
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysAfterMarkDone)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdueAfterMarkDone)

    // Clear mocks to track reload
    vi.mocked(GetDayEntries).mockClear()

    // Click on first task in sidebar to select it (makes action bar visible)
    const sidebar = screen.getByTestId('overdue-sidebar')
    await user.click(within(sidebar).getByText('Overdue task 1'))

    // Now click mark done button (action bar is visible because entry is selected)
    const doneButton = within(sidebar).getByTitle('Mark as done')
    await user.click(doneButton)

    // Wait for the mark done API call AND data reload
    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(1)
    })
    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })

    // Give React time to process the state update from the reload
    await waitFor(() => {
      // Entry should STILL be visible in sidebar (even though it's now 'done')
      // This is the key assertion - currently fails because activelyCyclingEntry not set
      // Scope to sidebar to ensure it's in the Pending Tasks section, not main panel
      const sidebarAfter = screen.getByTestId('overdue-sidebar')
      expect(within(sidebarAfter).getByText('Overdue task 1')).toBeInTheDocument()
    })
  })

  it('keeps entry visible in sidebar after cancelling', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for main panel to render first
    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar (starts collapsed by default)
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    // Wait for sidebar to render with overdue entries
    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Verify both tasks are visible in sidebar
    expect(screen.getByText('Overdue task 2')).toBeInTheDocument()

    // Mock the reload to return the entry as cancelled (would normally disappear)
    const mockDaysAfterCancel = createMockDays([createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 100, EntityID: 'e100', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Cancelled', Content: 'Overdue task 1', CreatedAt: '2026-01-15T10:00:00Z' }),
      ],
    })])
    const mockOverdueAfterCancel = createMockOverdue([
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Overdue task 2', CreatedAt: '2026-01-15T11:00:00Z' }),
    ])
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysAfterCancel)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdueAfterCancel)

    // Clear mocks to track reload
    vi.mocked(GetDayEntries).mockClear()

    // Click on first task in sidebar to select it (makes action bar visible)
    const sidebar = screen.getByTestId('overdue-sidebar')
    await user.click(within(sidebar).getByText('Overdue task 1'))

    // Now click cancel button (action bar is visible because entry is selected)
    const cancelButton = within(sidebar).getByTitle('Cancel entry')
    await user.click(cancelButton)

    // Wait for the cancel API call AND data reload
    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(1)
    })
    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })

    // Give React time to process the state update from the reload
    await waitFor(() => {
      // Entry should STILL be visible in sidebar (even though it's now 'cancelled')
      // This is the key assertion - currently fails because activelyCyclingEntry not set
      // Scope to sidebar to ensure it's in the Pending Tasks section, not main panel
      const sidebarAfter = screen.getByTestId('overdue-sidebar')
      expect(within(sidebarAfter).getByText('Overdue task 1')).toBeInTheDocument()
    })
  })

  it('shows unmark done button after marking entry done', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for main panel to render first
    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar (starts collapsed by default)
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    // Wait for sidebar to render with overdue entries
    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Mock the reload to return the entry as done
    const mockDaysAfterDone = createMockDays([createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 100, EntityID: 'e100', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Done', Content: 'Overdue task 1', CreatedAt: '2026-01-15T10:00:00Z' }),
      ],
    })])
    const mockOverdueAfterDone = createMockOverdue([
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Overdue task 2', CreatedAt: '2026-01-15T11:00:00Z' }),
    ])
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysAfterDone)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdueAfterDone)

    // Click on first task in sidebar to select it (makes action bar visible)
    const sidebar = screen.getByTestId('overdue-sidebar')
    await user.click(within(sidebar).getByText('Overdue task 1'))

    // Now click mark done button (action bar is visible because entry is selected)
    const doneButton = within(sidebar).getByTitle('Mark as done')
    await user.click(doneButton)

    // Wait for the mark done API call AND data reload
    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(1)
    })
    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })

    // After marking done, the entry should show an "Unmark done" button
    // Note: Entry remains selected, so action bar is still visible
    await waitFor(() => {
      const sidebarAfter = screen.getByTestId('overdue-sidebar')
      // The entry is still visible
      expect(within(sidebarAfter).getByText('Overdue task 1')).toBeInTheDocument()
      // And should now have an "Unmark done" button instead of "Mark as done"
      expect(within(sidebarAfter).getByTitle('Unmark done')).toBeInTheDocument()
    })
  })

  it('snapshots pending tasks when sidebar expands and keeps them stable until re-expanded', async () => {
    // Setup: three tasks in sidebar
    const mockInitialDays = createMockDays([createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 100, EntityID: 'e100', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
      ],
    })])
    const mockInitialOverdue = createMockOverdue([
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Overdue task 1', CreatedAt: '2026-01-15T10:00:00Z' }),
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Overdue task 2', CreatedAt: '2026-01-15T11:00:00Z' }),
      createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Task', Content: 'Overdue task 3', CreatedAt: '2026-01-15T12:00:00Z' }),
    ])

    vi.mocked(GetDayEntries).mockResolvedValue(mockInitialDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockInitialOverdue)
    vi.mocked(MarkEntryDone).mockResolvedValue(undefined)

    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for main panel to render first
    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar - this snapshots the pending tasks
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    // Wait for sidebar to render with overdue entries
    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // After marking task 1 done, server returns it as Done type (removed from Overdue)
    const mockDaysAfterFirst = createMockDays([createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 100, EntityID: 'e100', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Done', Content: 'Overdue task 1', CreatedAt: '2026-01-15T10:00:00Z' }),
      ],
    })])
    const mockOverdueAfterFirst = createMockOverdue([
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Overdue task 2', CreatedAt: '2026-01-15T11:00:00Z' }),
      createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Task', Content: 'Overdue task 3', CreatedAt: '2026-01-15T12:00:00Z' }),
    ])

    // Select first task to show action bar
    const sidebar = screen.getByTestId('overdue-sidebar')
    await user.click(within(sidebar).getByText('Overdue task 1'))

    // Click mark done button
    const doneButton = within(sidebar).getByTitle('Mark as done')
    await user.click(doneButton)

    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysAfterFirst)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdueAfterFirst)

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(1)
    })

    // Task 1 should STILL be visible (snapshot doesn't change)
    // All three tasks remain visible with original state
    await waitFor(() => {
      const sidebarAfter = screen.getByTestId('overdue-sidebar')
      expect(within(sidebarAfter).getByText('Overdue task 1')).toBeInTheDocument()
      expect(within(sidebarAfter).getByText('Overdue task 2')).toBeInTheDocument()
      expect(within(sidebarAfter).getByText('Overdue task 3')).toBeInTheDocument()
    })

    // Select second task to show action bar
    // Note: Task 1 now shows "Unmark done", so "Mark as done" buttons are only for tasks 2 and 3
    await user.click(within(screen.getByTestId('overdue-sidebar')).getByText('Overdue task 2'))

    const doneButtonSecond = within(screen.getByTestId('overdue-sidebar')).getByTitle('Mark as done')
    await user.click(doneButtonSecond)

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(2)
    })

    // All tasks STILL visible - snapshot is stable
    const sidebarStable = screen.getByTestId('overdue-sidebar')
    expect(within(sidebarStable).getByText('Overdue task 1')).toBeInTheDocument()
    expect(within(sidebarStable).getByText('Overdue task 2')).toBeInTheDocument()
    expect(within(sidebarStable).getByText('Overdue task 3')).toBeInTheDocument()

    // Now collapse and re-expand the sidebar - this should refresh the snapshot
    const mockDaysAfterBoth = createMockDays([createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 100, EntityID: 'e100', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
      ],
    })])
    const mockOverdueAfterBoth = createMockOverdue([
      createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Task', Content: 'Overdue task 3', CreatedAt: '2026-01-15T12:00:00Z' }),
    ])
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysAfterBoth)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdueAfterBoth)

    // Collapse sidebar
    await user.click(toggleButton)

    // Re-expand sidebar - should show fresh data (only task 3 remains)
    await user.click(toggleButton)

    await waitFor(() => {
      const sidebarRefreshed = screen.getByTestId('overdue-sidebar')
      // Only task 3 should be visible now
      expect(within(sidebarRefreshed).getByText('Overdue task 3')).toBeInTheDocument()
      expect(within(sidebarRefreshed).queryByText('Overdue task 1')).not.toBeInTheDocument()
      expect(within(sidebarRefreshed).queryByText('Overdue task 2')).not.toBeInTheDocument()
    })
  })
})

describe('App - Migrate Entry with Children', () => {
  const mockDaysAfterMigration = createMockDays([
    createMockDayEntries({
      Date: '2026-01-17T00:00:00Z',
      Entries: [
        createMockEntry({
          ID: 1,
          EntityID: 'e1',
          Type: 'Migrated',
          Content: 'Parent task',
          ParentID: null,
          Depth: 0,
          CreatedAt: '2026-01-17T10:00:00Z'
        }),
        createMockEntry({
          ID: 2,
          EntityID: 'e2',
          Type: 'Migrated',
          Content: 'Child note',
          ParentID: 1,
          Depth: 1,
          CreatedAt: '2026-01-17T10:01:00Z'
        }),
      ],
    }),
  ])

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('maintains correct indentation for migrated child notes', async () => {
    // Show the old location after migration
    vi.mocked(GetDayEntries).mockResolvedValue(mockDaysAfterMigration)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for render
    await waitFor(() => {
      expect(screen.getByText('Parent task')).toBeInTheDocument()
    })

    // Expand parent first (parents default to collapsed)
    const expandButton = screen.getByRole('button', { name: '' })
    fireEvent.click(expandButton)

    // Now child should be visible
    expect(screen.getByText('Child note')).toBeInTheDocument()

    // Check that both parent and child are rendered as migrated
    const allEntries = screen.getAllByTestId('entry-item')

    // Find entries by checking data-entry-id attribute
    const parent = allEntries.find(el => el.getAttribute('data-entry-id') === '1')
    const child = allEntries.find(el => el.getAttribute('data-entry-id') === '2')

    expect(parent).toBeDefined()
    expect(child).toBeDefined()

    // Check indentation: parent should have depth 0 (8px), child should have depth 1 (28px)
    const parentPadding = window.getComputedStyle(parent!).paddingLeft
    const childPadding = window.getComputedStyle(child!).paddingLeft

    // This test will FAIL if the bug exists - migrated children lose indentation
    expect(parentPadding).toBe('8px')
    expect(childPadding).toBe('28px')  // depth 1 = 1 * 20 + 8 = 28px
  })
})
