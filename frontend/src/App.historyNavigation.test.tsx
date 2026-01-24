import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
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
      render(<App />)

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
      })

      // Back button should not be present initially
      expect(screen.queryByRole('button', { name: /go back/i })).not.toBeInTheDocument()
    })
  })

  describe('navigation from WeekSummary', () => {
    it('shows back button after navigating from attention item popover', async () => {
      const user = userEvent.setup()
      render(<App />)

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
      })

      // Navigate to weekly review
      const reviewButton = screen.getByRole('button', { name: /weekly review/i })
      await user.click(reviewButton)

      await waitFor(() => {
        expect(screen.getByTestId('week-summary')).toBeInTheDocument()
      })

      // Click on an attention item to open popover
      const weekSummary = screen.getByTestId('week-summary')
      const attentionItem = within(weekSummary).getByText('Task needing attention')
      await user.click(attentionItem)

      // Wait for popover to appear
      await waitFor(() => {
        expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
      })

      // Click "Go to" button to navigate
      const goToButton = screen.getByRole('button', { name: /go to/i })
      await user.click(goToButton)

      // Should now be in today/journal view and back button should be visible
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()
      })
    })

    it('clicking back button returns to previous view', async () => {
      const user = userEvent.setup()
      render(<App />)

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
      })

      // Navigate to weekly review
      const reviewButton = screen.getByRole('button', { name: /weekly review/i })
      await user.click(reviewButton)

      await waitFor(() => {
        expect(screen.getByTestId('week-summary')).toBeInTheDocument()
      })

      // Click on an attention item to open popover
      const weekSummary = screen.getByTestId('week-summary')
      const attentionItem = within(weekSummary).getByText('Task needing attention')
      await user.click(attentionItem)

      await waitFor(() => {
        expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
      })

      // Click "Go to" button to navigate
      const goToButton = screen.getByRole('button', { name: /go to/i })
      await user.click(goToButton)

      // Wait for navigation and back button
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()
      })

      // Click back button
      const backButton = screen.getByRole('button', { name: /go back/i })
      await user.click(backButton)

      // Should be back in weekly review
      await waitFor(() => {
        expect(screen.getByTestId('week-summary')).toBeInTheDocument()
      })
    })

    it('supports multi-level navigation history', async () => {
      const user = userEvent.setup()
      render(<App />)

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
      })

      // Navigate to weekly review (pushes today to history)
      const reviewButton = screen.getByRole('button', { name: /weekly review/i })
      await user.click(reviewButton)

      await waitFor(() => {
        expect(screen.getByTestId('week-summary')).toBeInTheDocument()
      })

      // Click on an attention item to open popover
      const weekSummary = screen.getByTestId('week-summary')
      const attentionItem = within(weekSummary).getByText('Task needing attention')
      await user.click(attentionItem)

      await waitFor(() => {
        expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
      })

      // Click "Go to" button to navigate (pushes week to history)
      const goToButton = screen.getByRole('button', { name: /go to/i })
      await user.click(goToButton)

      // Wait for back button (history: [today, week])
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()
      })

      // Click back button - should return to week view
      const backButton = screen.getByRole('button', { name: /go back/i })
      await user.click(backButton)

      await waitFor(() => {
        expect(screen.getByTestId('week-summary')).toBeInTheDocument()
      })

      // Back button should still be visible (history: [today])
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()
      })

      // Click back again - should return to today view
      const backButtonAgain = screen.getByRole('button', { name: /go back/i })
      await user.click(backButtonAgain)

      await waitFor(() => {
        expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
      })

      // Back button should now disappear (history: [])
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /go back/i })).not.toBeInTheDocument()
      })
    })
  })

  describe('manual navigation to today clears history', () => {
    it('clears history when manually navigating to today view via sidebar', async () => {
      const user = userEvent.setup()
      render(<App />)

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
