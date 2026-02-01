import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockDayEntries, createMockDays, createMockOverdue } from './test/mocks'

const mockDays = createMockDays([createMockDayEntries({})])
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
  GetEditableDocument: vi.fn().mockResolvedValue(''),
  ValidateEditableDocument: vi.fn().mockResolvedValue({ isValid: true, errors: [] }),
  ApplyEditableDocument: vi.fn().mockResolvedValue({ inserted: 0, deleted: 0 }),
  SearchEntries: vi.fn().mockResolvedValue([]),
  GetStats: vi.fn().mockResolvedValue({
    TotalEntries: 0,
    TasksCompleted: 0,
    ActiveHabits: 0,
    CurrentStreak: 0,
  }),
  GetVersion: vi.fn().mockResolvedValue('1.0.0'),
}))

import { GetDayEntries, GetOverdue } from './wailsjs/go/wails/App'

describe('App - Go to Today', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('shows Go to today button when viewing a different day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to previous day
    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    // Go to today button should be visible (not invisible class)
    await waitFor(() => {
      const jumpToTodayBtn = screen.getByTestId('jump-to-today')
      expect(jumpToTodayBtn).not.toHaveClass('invisible')
    })
  })

  it('hides Go to today button when viewing today', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Go to today button should have invisible class when viewing today
    const jumpToTodayBtn = screen.getByTestId('jump-to-today')
    expect(jumpToTodayBtn).toHaveClass('invisible')
  })

  it('clicking Go to today button navigates back to today', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to previous day
    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    await waitFor(() => {
      const jumpToTodayBtn = screen.getByTestId('jump-to-today')
      expect(jumpToTodayBtn).not.toHaveClass('invisible')
    })

    vi.mocked(GetDayEntries).mockClear()

    // Click Go to today button
    const todayButton = screen.getByTestId('jump-to-today')
    await user.click(todayButton)

    // Should trigger data refresh
    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })

    // Go to today button should become invisible after navigating back to today
    await waitFor(() => {
      const jumpToTodayBtn = screen.getByTestId('jump-to-today')
      expect(jumpToTodayBtn).toHaveClass('invisible')
    })
  })

  it('pressing T navigates to today when viewing a different day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to previous day
    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    await waitFor(() => {
      const jumpToTodayBtn = screen.getByTestId('jump-to-today')
      expect(jumpToTodayBtn).not.toHaveClass('invisible')
    })

    vi.mocked(GetDayEntries).mockClear()

    // Press T to go to today
    await user.keyboard('T')

    // Should trigger data refresh
    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })

    // Go to today button should become invisible
    await waitFor(() => {
      const jumpToTodayBtn = screen.getByTestId('jump-to-today')
      expect(jumpToTodayBtn).toHaveClass('invisible')
    })
  })
})

describe('App - Review View (formerly Past Week)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows "Weekly Review" label in sidebar navigation', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Sidebar should show "Weekly Review" for the week/review view
    expect(screen.getByRole('button', { name: /weekly review/i })).toBeInTheDocument()
  })

  it('shows "Weekly Review" as header title when review view is selected', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Click on Weekly Review in sidebar
    const reviewButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(reviewButton)

    // Header title should show "Weekly Review"
    await waitFor(() => {
      const headings = screen.getAllByRole('heading', { name: /weekly review/i })
      expect(headings.length).toBeGreaterThan(0)
    })
  })

  it('shows navigation controls in review view', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Click on Weekly Review in sidebar
    const reviewButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(reviewButton)

    // Should show prev/next navigation buttons
    await waitFor(() => {
      expect(screen.getByTitle('Previous week')).toBeInTheDocument()
      expect(screen.getByTitle('Next week')).toBeInTheDocument()
    })
  })

  it('disables next week button when viewing current week', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Click on Weekly Review in sidebar
    const reviewButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(reviewButton)

    // Next week button should be disabled when at current week
    await waitFor(() => {
      const nextButton = screen.getByTitle('Next week')
      expect(nextButton).toBeDisabled()
    })
  })
})
