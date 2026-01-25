import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn()
  window.scrollTo = vi.fn() as unknown as typeof window.scrollTo
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

import { GetAgenda } from './wailsjs/go/wails/App'

// Week data with items needing attention for testing navigation
const weekAgendaWithAttentionItems = createMockAgenda({
  Days: [
    createMockDayEntries({
      Date: '2026-01-19T00:00:00Z',
      Entries: [
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Task needing attention', Priority: 'High', CreatedAt: '2026-01-19T10:00:00Z' }),
        createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Note', Content: 'A note entry', CreatedAt: '2026-01-19T11:00:00Z' }),
      ],
    }),
    createMockDayEntries({
      Date: '2026-01-20T00:00:00Z',
      Entries: [
        createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Task', Content: 'Another task', CreatedAt: '2026-01-20T10:00:00Z' }),
      ],
    }),
  ],
  Overdue: [],
})

describe('App - Navigation History', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithAttentionItems)
  })

  describe('initial state', () => {
    it('Header does not show back button initially (canGoBack is false)', async () => {
      render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
      })

      // Back button should not be present initially
      expect(screen.queryByRole('button', { name: /go back/i })).not.toBeInTheDocument()
    })
  })

  // Tests for navigation from WeekSummary popover removed:
  // WeekSummary no longer uses popovers - the UX has changed.
  // Context viewing will be handled by a new ContextPanel, not popovers.

  describe('manual navigation to today clears history', () => {
    it('clears history when manually navigating to today view via sidebar', async () => {
      const user = userEvent.setup()
      render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
      })

      // Navigate to weekly review (pushes today to history)
      const reviewButton = screen.getByRole('button', { name: /weekly review/i })
      await user.click(reviewButton)

      await waitFor(() => {
        expect(screen.getByTestId('week-summary')).toBeInTheDocument()
      })

      // Back button should be visible (can go back to today)
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()
      })

      // Manually navigate to today via sidebar
      const todayButton = screen.getByRole('button', { name: /journal/i })
      await user.click(todayButton)

      await waitFor(() => {
        expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
      })

      // Back button should be gone (history cleared by navigating to today)
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /go back/i })).not.toBeInTheDocument()
      })
    })
  })
})
