import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockAgenda } from './test/mocks'

const mockAgendaWithOverdue = createMockAgenda({
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
    Days: [{ Date: '2026-01-25T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
  }),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MoveEntryToList: vi.fn().mockResolvedValue(undefined),
  MoveEntryToRoot: vi.fn().mockResolvedValue(undefined),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  GetEntryContext: vi.fn().mockResolvedValue([]),
  RetypeEntry: vi.fn().mockResolvedValue(undefined),
}))

import { GetAgenda } from './wailsjs/go/wails/App'

describe('App - Sidebar Collapse', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockAgendaWithOverdue)
  })

  describe('Sidebar collapse with [ keyboard shortcut', () => {
    it('pressing [ toggles sidebar collapsed state', async () => {
      const user = userEvent.setup()
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      // Wait for app to load
      await waitFor(() => {
        expect(screen.getByTestId('overdue-sidebar')).toBeInTheDocument()
      })

      // Sidebar should be collapsed initially (starts collapsed by default)
      expect(screen.queryByText('Overdue task 1')).not.toBeInTheDocument()

      // Press [ to expand
      await user.keyboard('{[}')

      // Sidebar content should be visible
      await waitFor(() => {
        expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      })

      // Press [ again to collapse
      await user.keyboard('{[}')

      // Sidebar content should be hidden again
      await waitFor(() => {
        expect(screen.queryByText('Overdue task 1')).not.toBeInTheDocument()
      })
    })

    it('[ shortcut only works on journal view', async () => {
      const user = userEvent.setup()
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      // Wait for app to load
      await waitFor(() => {
        expect(screen.getByTestId('overdue-sidebar')).toBeInTheDocument()
      })

      // Switch to a different view (habits)
      await user.click(screen.getByRole('button', { pressed: false, name: /habit tracker/i }))

      await waitFor(() => {
        expect(screen.queryByTestId('overdue-sidebar')).not.toBeInTheDocument()
      })

      // Press [ should have no effect (sidebar not visible anyway)
      await user.keyboard('{[}')

      // Verify we're still on habits view
      expect(screen.queryByTestId('overdue-sidebar')).not.toBeInTheDocument()
    })
  })

  describe('Sidebar collapse with mouse button', () => {
    it('clicking toggle button collapses sidebar', async () => {
      const user = userEvent.setup()
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      // Wait for app to load
      await waitFor(() => {
        expect(screen.getByTestId('overdue-sidebar')).toBeInTheDocument()
      })

      // Sidebar should be collapsed initially (starts collapsed by default)
      expect(screen.queryByText('Overdue task 1')).not.toBeInTheDocument()

      // Click the toggle button to expand
      await user.click(screen.getByRole('button', { name: /toggle sidebar/i }))

      // Sidebar content should be visible
      await waitFor(() => {
        expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      })

      // Click again to collapse
      await user.click(screen.getByRole('button', { name: /toggle sidebar/i }))

      // Sidebar content should be hidden again
      await waitFor(() => {
        expect(screen.queryByText('Overdue task 1')).not.toBeInTheDocument()
      })
    })
  })

  describe('Sidebar width styling', () => {
    it('does not apply static width class when sidebar is expanded', async () => {
      const user = userEvent.setup()
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      // Wait for app to load
      await waitFor(() => {
        expect(screen.getByTestId('overdue-sidebar')).toBeInTheDocument()
      })

      // Expand the sidebar (starts collapsed by default)
      await user.keyboard('{[}')

      await waitFor(() => {
        expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      })

      // Find the sidebar's parent aside element
      const sidebar = screen.getByTestId('overdue-sidebar')
      const asideElement = sidebar.closest('aside')

      expect(asideElement).toBeInTheDocument()

      // Should NOT have the static width class w-[32rem] when expanded
      expect(asideElement?.className).not.toContain('w-[32rem]')
    })

    it('applies w-10 class when sidebar is collapsed', async () => {
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      // Wait for app to load
      await waitFor(() => {
        expect(screen.getByTestId('overdue-sidebar')).toBeInTheDocument()
      })

      // Sidebar starts collapsed by default
      expect(screen.queryByText('Overdue task 1')).not.toBeInTheDocument()

      // Find the sidebar's parent aside element
      const sidebar = screen.getByTestId('overdue-sidebar')
      const asideElement = sidebar.closest('aside')

      // Should have width 2.5rem (w-10 equivalent) when collapsed
      expect(asideElement).toHaveStyle({ width: '2.5rem' })
    })
  })
})
