import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

const mockAgendaWithOverdue = createMockAgenda({
  Days: [createMockDayEntries({
    Entries: [
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Main panel task', CreatedAt: '2026-01-17T10:00:00Z' }),
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Note', Content: 'Main panel note', CreatedAt: '2026-01-17T11:00:00Z' }),
    ],
  })],
  Overdue: [
    createMockEntry({ ID: 10, EntityID: 'e10', Type: 'Task', Content: 'Overdue task 1' }),
    createMockEntry({ ID: 11, EntityID: 'e11', Type: 'Task', Content: 'Overdue task 2' }),
  ],
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
  GetEntryContext: vi.fn().mockResolvedValue([]),
}))

import { GetAgenda } from './wailsjs/go/wails/App'

describe('App - Panel Navigation with Tab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockAgendaWithOverdue)
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
        expect(sidebarEntry).toHaveClass('bg-accent')
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
        expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      })

      // Tab to sidebar
      await user.keyboard('{Tab}')

      await waitFor(() => {
        const firstEntry = screen.getByText('Overdue task 1').closest('.group')
        expect(firstEntry).toHaveClass('bg-accent')
      })

      // Press j to move to next overdue task
      await user.keyboard('j')

      await waitFor(() => {
        const secondEntry = screen.getByText('Overdue task 2').closest('.group')
        expect(secondEntry).toHaveClass('bg-accent')
      })
    })
  })
})

describe('App - Mutual Exclusion of Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockAgendaWithOverdue)
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

    // Tab to sidebar to select something there
    await user.keyboard('{Tab}')

    await waitFor(() => {
      const sidebarEntry = screen.getByText('Overdue task 1').closest('.group')
      expect(sidebarEntry).toHaveClass('bg-accent')
    })

    // Click on main panel entry
    await user.click(screen.getByText('Main panel note'))

    await waitFor(() => {
      // Main panel entry should be selected
      const mainPanelEntry = screen.getByText('Main panel note').closest('[data-entry-id]')
      expect(mainPanelEntry).toHaveAttribute('data-selected', 'true')
      // Sidebar entry should NOT be selected anymore
      const sidebarEntry = screen.getByText('Overdue task 1').closest('.group')
      expect(sidebarEntry).not.toHaveClass('bg-accent')
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
      expect(sidebarEntry).toHaveClass('bg-accent')
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
      (el) => el?.classList.contains('bg-accent')
    )

    // Total selected should be exactly 1
    const totalSelected = selectedMainEntries.length + selectedSidebarEntries.length
    expect(totalSelected).toBe(1)
  })
})
