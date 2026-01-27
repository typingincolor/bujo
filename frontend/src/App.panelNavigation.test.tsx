import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockDays, createMockOverdue } from './test/mocks'

const mockDays = createMockDays([createMockDayEntries({
  Entries: [
    createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
    createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Note', Content: 'Main panel note', CreatedAt: '2026-01-17T11:00:00Z' }),
  ],
})])
const mockOverdue = createMockOverdue([
  createMockEntry({ ID: 10, EntityID: 'e10', Type: 'Task', Content: 'Overdue task 1' }),
  createMockEntry({ ID: 11, EntityID: 'e11', Type: 'Task', Content: 'Overdue task 2' }),
])

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
  GetEntryContext: vi.fn().mockResolvedValue([]),
  RetypeEntry: vi.fn().mockResolvedValue(undefined),
}))

import { GetDayEntries, GetOverdue } from './wailsjs/go/wails/App'

describe('App - Panel Navigation with Tab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  describe('Tab key switches focus between panels', () => {
    it('pressing Tab moves focus from main panel to sidebar', async () => {
      const user = userEvent.setup()
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      await waitFor(() => {
        expect(screen.getByText('Main panel task')).toBeInTheDocument()
      })

      // Expand sidebar (starts collapsed by default)
      const toggleButton = screen.getByLabelText('Toggle sidebar')
      await user.click(toggleButton)

      await waitFor(() => {
        expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      })

      // Initially main panel has selection
      const mainPanelEntry = screen.getByText('Main panel task').closest('[data-entry-id]')
      expect(mainPanelEntry).toHaveAttribute('data-selected', 'true')

      // Press Tab to move to sidebar
      await user.keyboard('{Tab}')

      // Now sidebar should have selection, main panel should not
      await waitFor(() => {
        // Main panel entry should no longer be selected
        expect(mainPanelEntry).toHaveAttribute('data-selected', 'false')
        // First overdue entry in sidebar should be highlighted
        const sidebarEntry = screen.getByText('Overdue task 1').closest('.group')
        expect(sidebarEntry).toHaveClass('ring-primary/30')
      })
    })

    it('pressing Tab again moves focus back to main panel', async () => {
      const user = userEvent.setup()
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      await waitFor(() => {
        expect(screen.getByText('Main panel task')).toBeInTheDocument()
      })

      // Expand sidebar (starts collapsed by default)
      const toggleButton = screen.getByLabelText('Toggle sidebar')
      await user.click(toggleButton)

      await waitFor(() => {
        expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      })

      // Tab to sidebar, then Tab again back to main
      await user.keyboard('{Tab}{Tab}')

      await waitFor(() => {
        const mainPanelEntry = screen.getByText('Main panel task').closest('[data-entry-id]')
        expect(mainPanelEntry).toHaveAttribute('data-selected', 'true')
      })
    })
  })

  describe('Arrow keys work in sidebar when focused', () => {
    it('j/ArrowDown moves selection down in sidebar', async () => {
      const user = userEvent.setup()
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      await waitFor(() => {
        expect(screen.getByText('Main panel task')).toBeInTheDocument()
      })

      // Expand sidebar (starts collapsed by default)
      const toggleButton = screen.getByLabelText('Toggle sidebar')
      await user.click(toggleButton)

      await waitFor(() => {
        expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      })

      // Tab to sidebar
      await user.keyboard('{Tab}')

      await waitFor(() => {
        const firstEntry = screen.getByText('Overdue task 1').closest('.group')
        expect(firstEntry).toHaveClass('ring-primary/30')
      })

      // Press j to move to next overdue task
      await user.keyboard('j')

      await waitFor(() => {
        const secondEntry = screen.getByText('Overdue task 2').closest('.group')
        expect(secondEntry).toHaveClass('ring-primary/30')
      })
    })
  })
})

describe('App - Mutual Exclusion of Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('selecting in main panel clears sidebar selection', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar (starts collapsed by default)
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Tab to sidebar to select something there
    await user.keyboard('{Tab}')

    await waitFor(() => {
      const sidebarEntry = screen.getByText('Overdue task 1').closest('.group')
      expect(sidebarEntry).toHaveClass('ring-primary/30')
    })

    // Click on main panel entry
    await user.click(screen.getByText('Main panel note'))

    await waitFor(() => {
      // Main panel entry should be selected
      const mainPanelEntry = screen.getByText('Main panel note').closest('[data-entry-id]')
      expect(mainPanelEntry).toHaveAttribute('data-selected', 'true')
      // Sidebar entry should NOT be selected anymore
      const sidebarEntry = screen.getByText('Overdue task 1').closest('.group')
      expect(sidebarEntry).not.toHaveClass('ring-primary/30')
    })
  })

  it('clicking in sidebar clears main panel selection', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar (starts collapsed by default)
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Initially main panel task is selected
    const mainPanelEntry = screen.getByText('Main panel task').closest('[data-entry-id]')
    expect(mainPanelEntry).toHaveAttribute('data-selected', 'true')

    // Click on sidebar entry
    await user.click(screen.getByText('Overdue task 1'))

    await waitFor(() => {
      // Main panel entry should NOT be selected anymore
      expect(mainPanelEntry).toHaveAttribute('data-selected', 'false')
      // Sidebar entry should be selected
      const sidebarEntry = screen.getByText('Overdue task 1').closest('.group')
      expect(sidebarEntry).toHaveClass('ring-primary/30')
    })
  })

  it('clicking row then pressing arrow key results in only one highlight', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Initially first entry is selected
    const firstEntry = screen.getByText('Main panel task').closest('[data-entry-id]')
    expect(firstEntry).toHaveAttribute('data-selected', 'true')

    // Click on second entry (note)
    await user.click(screen.getByText('Main panel note'))

    await waitFor(() => {
      const noteEntry = screen.getByText('Main panel note').closest('[data-entry-id]')
      expect(noteEntry).toHaveAttribute('data-selected', 'true')
      // First entry should NOT be selected anymore
      expect(firstEntry).toHaveAttribute('data-selected', 'false')
    })

    // Press down arrow
    await user.keyboard('{ArrowDown}')

    // After pressing arrow, only one entry should be selected
    await waitFor(() => {
      const allEntries = screen.getAllByTestId('entry-item')
      const selectedEntries = allEntries.filter(
        (el) => el.getAttribute('data-selected') === 'true'
      )
      expect(selectedEntries.length).toBe(1)
    })
  })

  it('only one entry is highlighted across entire screen', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar (starts collapsed by default)
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Find all elements with selection indicators
    const mainPanelEntries = screen.getAllByTestId('entry-item')
    const selectedMainEntries = mainPanelEntries.filter(
      (el) => el.getAttribute('data-selected') === 'true'
    )

    // Get sidebar entries with bg-accent class
    const sidebarEntries = screen.getAllByText(/Overdue task/).map(
      (el) => el.closest('.group')
    ).filter(Boolean)
    const selectedSidebarEntries = sidebarEntries.filter(
      (el) => el?.classList.contains('ring-primary/30')
    )

    // Total selected should be exactly 1
    const totalSelected = selectedMainEntries.length + selectedSidebarEntries.length
    expect(totalSelected).toBe(1)
  })

  it('data refresh after sidebar selection does not cause dual highlight', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Expand sidebar (starts collapsed by default)
    const toggleButton = screen.getByLabelText('Toggle sidebar')
    await user.click(toggleButton)

    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Click on sidebar entry
    await user.click(screen.getByText('Overdue task 1'))

    await waitFor(() => {
      const sidebarEntry = screen.getByText('Overdue task 1').closest('.group')
      expect(sidebarEntry).toHaveClass('ring-primary/30')
    })

    // Simulate data refresh by updating the mock and triggering re-render
    vi.mocked(GetDayEntries).mockResolvedValue(createMockDays([createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
        createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Note', Content: 'Main panel note', CreatedAt: '2026-01-17T11:00:00Z' }),
        createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Task', Content: 'New task added', CreatedAt: '2026-01-17T12:00:00Z' }),
      ],
    })]))

    // Trigger a data refresh (e.g., by simulating an action that triggers loadData)
    // For now, we'll simulate the 'data:changed' event
    const { EventsOn } = await import('./wailsjs/runtime/runtime')
    const eventCallback = vi.mocked(EventsOn).mock.calls[0]?.[1]
    if (eventCallback) {
      eventCallback()
    }

    await waitFor(() => {
      expect(screen.getByText('New task added')).toBeInTheDocument()
    })

    // After data refresh, should still only have one highlighted entry
    // The sidebar selection should be preserved, main panel should NOT be highlighted
    const mainPanelEntries = screen.getAllByTestId('entry-item')
    const selectedMainEntries = mainPanelEntries.filter(
      (el) => el.getAttribute('data-selected') === 'true'
    )

    const sidebarEntries = screen.getAllByText(/Overdue task/).map(
      (el) => el.closest('.group')
    ).filter(Boolean)
    const selectedSidebarEntries = sidebarEntries.filter(
      (el) => el?.classList.contains('ring-primary/30')
    )

    // Total selected should be exactly 1 (only sidebar)
    const totalSelected = selectedMainEntries.length + selectedSidebarEntries.length
    expect(totalSelected).toBe(1)
    expect(selectedSidebarEntries.length).toBe(1) // Sidebar selection preserved
    expect(selectedMainEntries.length).toBe(0) // No main panel selection
  })
})

describe('App - Keyboard Action Shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('p key cycles priority on selected entry', async () => {
    const user = userEvent.setup()
    const { CyclePriority } = await import('./wailsjs/go/wails/App')

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Press p to cycle priority
    await user.keyboard('p')

    await waitFor(() => {
      expect(CyclePriority).toHaveBeenCalledWith(1)
    })
  })

  it('t key cycles type on selected entry', async () => {
    const user = userEvent.setup()
    const { RetypeEntry } = await import('./wailsjs/go/wails/App')

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Main panel task')).toBeInTheDocument()
    })

    // Press t to cycle type
    await user.keyboard('t')

    // Should call RetypeEntry to change type (task -> note)
    await waitFor(() => {
      expect(RetypeEntry).toHaveBeenCalledWith(1, 'note')
    })
  })
})
