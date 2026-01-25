import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
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

import { GetAgenda, GetHabits, SetMood, SetWeather, SetLocation } from './wailsjs/go/wails/App'


describe('App - No flicker on data refresh', () => {
  const originalError = console.error

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    // Suppress act() warnings in this test - the async flow is intentionally complex
    console.error = (...args: unknown[]) => {
      if (typeof args[0] === 'string' && args[0].includes('not wrapped in act')) return
      originalError(...args)
    }
  })

  afterEach(() => {
    console.error = originalError
  })

  it('does not show loading spinner when refreshing data after habit action', async () => {
    // Track loading spinner appearances
    let loadingSpinnerShownDuringRefresh = false
    let initialLoadComplete = false

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for initial load to complete
    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })
    initialLoadComplete = true

    // Navigate to habits view
    const habitsButton = screen.getByRole('button', { name: /habit tracker/i })
    fireEvent.click(habitsButton)

    // Verify we're in habits view
    await waitFor(() => {
      expect(screen.getAllByRole('heading', { name: /habit tracker/i }).length).toBeGreaterThan(0)
    })

    // Set up delayed mock to observe loading state during refresh
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let resolveGetHabits: (value: any) => void
    vi.mocked(GetHabits).mockImplementation(() => new Promise(resolve => {
      resolveGetHabits = resolve
      // Check loading state while API is in flight
      setTimeout(() => {
        if (initialLoadComplete && screen.queryByText('Loading your journal...')) {
          loadingSpinnerShownDuringRefresh = true
        }
      }, 10)
    }))

    // Trigger loadData by simulating habit creation (which calls onHabitChanged -> loadData)
    const addButton = screen.getByRole('button', { name: /add habit/i })
    fireEvent.click(addButton)

    // Find the input and submit a new habit
    const habitInput = screen.getByPlaceholderText('Habit name')
    fireEvent.change(habitInput, { target: { value: 'Test habit' } })
    fireEvent.keyDown(habitInput, { key: 'Enter' })

    // Small delay to let the API call start
    await new Promise(resolve => setTimeout(resolve, 50))

    // Resolve the pending API call wrapped in act() to avoid warning
    await act(async () => {
      resolveGetHabits!({ Habits: [] })
    })

    // Wait for everything to settle
    await waitFor(() => {
      expect(screen.getAllByRole('heading', { name: /habit tracker/i }).length).toBeGreaterThan(0)
    })

    // CRITICAL: Loading spinner should NOT have appeared during refresh
    expect(loadingSpinnerShownDuringRefresh).toBe(false)
  })
})

describe('App - Day Context (Mood/Weather/Location)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset GetHabits mock that may have been changed by previous tests
    vi.mocked(GetHabits).mockResolvedValue({ Habits: [] } as unknown as Awaited<ReturnType<typeof GetHabits>>)
  })

  it('displays current mood emoji in header when mood is set', async () => {
    const mockWithMood = createMockAgenda({
      Days: [createMockDayEntries({
        Mood: 'happy',
        Entries: [],
      })],
      Overdue: [],
    })
    vi.mocked(GetAgenda).mockResolvedValue(mockWithMood)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Happy mood emoji should be displayed
    expect(screen.getByText('😊')).toBeInTheDocument()
  })

  it('displays current weather emoji in header when weather is set', async () => {
    const mockWithWeather = createMockAgenda({
      Days: [createMockDayEntries({
        Weather: 'sunny',
        Entries: [],
      })],
      Overdue: [],
    })
    vi.mocked(GetAgenda).mockResolvedValue(mockWithWeather)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Sunny weather emoji should be displayed
    expect(screen.getByText('☀️')).toBeInTheDocument()
  })

  it('displays current location in header when location is set', async () => {
    const mockWithLocation = createMockAgenda({
      Days: [createMockDayEntries({
        Location: 'Home Office',
        Entries: [],
      })],
      Overdue: [],
    })
    vi.mocked(GetAgenda).mockResolvedValue(mockWithLocation)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Location should be displayed (appears in both Header and DayView)
    const locationElements = screen.getAllByText('Home Office')
    expect(locationElements.length).toBeGreaterThanOrEqual(1)
  })

  it('calls SetMood and refreshes data when selecting mood', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Click mood button to open picker
    await user.click(screen.getByTitle('Set mood'))

    // Select happy mood
    await user.click(screen.getByText('😊'))

    await waitFor(() => {
      expect(SetMood).toHaveBeenCalled()
    })

    // Data should be refreshed (GetAgenda is called twice per loadData - once for today, once for review)
    expect(GetAgenda).toHaveBeenCalledTimes(4)
  })

  it('calls SetWeather and refreshes data when selecting weather', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Click weather button to open picker
    await user.click(screen.getByTitle('Set weather'))

    // Select sunny weather
    await user.click(screen.getByText('☀️'))

    await waitFor(() => {
      expect(SetWeather).toHaveBeenCalled()
    })

    // Data should be refreshed (GetAgenda is called twice per loadData - once for today, once for review)
    expect(GetAgenda).toHaveBeenCalledTimes(4)
  })

  it('calls SetLocation and refreshes data when setting location', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Click location button to open picker
    await user.click(screen.getByTitle('Set location'))

    // Type a location and press Enter
    const input = screen.getByPlaceholderText('Enter location...')
    await user.type(input, 'Coffee Shop{Enter}')

    await waitFor(() => {
      expect(SetLocation).toHaveBeenCalled()
    })

    // Data should be refreshed (GetAgenda is called twice per loadData - once for today, once for review)
    expect(GetAgenda).toHaveBeenCalledTimes(4)
  })
})

describe('App - CaptureModal integration', () => {
  const localStorageMock = {
    getItem: vi.fn().mockReturnValue(null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
    Object.defineProperty(window, 'localStorage', {
      value: localStorageMock,
      writable: true,
    })
  })

  it('pressing c opens CaptureModal', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('c')

    await waitFor(() => {
      expect(screen.getByText('Capture Entries')).toBeInTheDocument()
    })
  })

  it('clicking capture button opens CaptureModal', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureButton = screen.getByTitle('Open capture modal')
    await user.click(captureButton)

    await waitFor(() => {
      expect(screen.getByText('Capture Entries')).toBeInTheDocument()
    })
  })

  it('closing CaptureModal returns focus to main view', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('c')

    await waitFor(() => {
      expect(screen.getByText('Capture Entries')).toBeInTheDocument()
    })

    // Press Escape or click Cancel to close
    await user.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByText('Capture Entries')).not.toBeInTheDocument()
    })
  })

  it('CaptureModal refreshes data after creating entries', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('c')

    await waitFor(() => {
      expect(screen.getByText('Capture Entries')).toBeInTheDocument()
    })

    const textarea = screen.getByPlaceholderText(/enter entries/i)
    await user.type(textarea, '. New task')

    vi.mocked(GetAgenda).mockClear()

    // Click Save Entries button
    const saveButton = screen.getByRole('button', { name: /save entries/i })
    await user.click(saveButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })
})
