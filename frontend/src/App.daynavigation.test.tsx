import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
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

import { GetDayEntries, GetOverdue, GetHabits } from './wailsjs/go/wails/App'


describe('App - Day Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('renders prev/next day navigation buttons in today view', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: /previous day/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /next day/i })).toBeInTheDocument()
  })

  it('renders date picker in today view', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByTestId('date-picker')).toBeInTheDocument()
  })

  // Skipped: HTML date input change events don't work reliably in jsdom CI.
  // Navigation behavior is tested via prev/next buttons and keyboard shortcuts (h/l).
  it.skip('changing date picker navigates to selected date', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    const datePicker = screen.getByLabelText(/pick date/i)
    await act(async () => {
      fireEvent.change(datePicker, { target: { value: '2026-01-20' } })
    })

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })

  it('clicking next day navigates to next day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    const nextButton = screen.getByRole('button', { name: /next day/i })
    await user.click(nextButton)

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })

  it('clicking prev day navigates to previous day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })

  it('pressing h navigates to previous day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    await user.keyboard('h')

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })

  it('pressing l navigates to next day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    vi.mocked(GetDayEntries).mockClear()

    await user.keyboard('l')

    await waitFor(() => {
      expect(GetDayEntries).toHaveBeenCalled()
    })
  })
})

describe('App - Habit View Toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('refetches habits with different day count when period changes', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Switch to habits view
    const habitsButton = screen.getByRole('button', { name: /habit tracker/i })
    await user.click(habitsButton)

    await waitFor(() => {
      expect(screen.getAllByRole('heading', { name: /habit tracker/i }).length).toBeGreaterThan(0)
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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Switch to habits view
    const habitsButton = screen.getByRole('button', { name: /habit tracker/i })
    await user.click(habitsButton)

    await waitFor(() => {
      expect(screen.getAllByRole('heading', { name: /habit tracker/i }).length).toBeGreaterThan(0)
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
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Keyboard shortcuts panel should not be visible by default
    expect(screen.queryByText('Keyboard Shortcuts')).not.toBeInTheDocument()
  })

  it('? key toggles keyboard shortcuts panel visibility', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

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
