import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
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

import { GetAgenda, AddEntry } from './wailsjs/go/wails/App'


describe('App - Inline Entry Creation (a/A/r shortcuts)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('pressing r shows inline input for root entry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('r')

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/type entry/i)).toBeInTheDocument()
    })
  })

  it('pressing a shows inline input for sibling entry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('a')

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/add entry/i)).toBeInTheDocument()
    })
  })

  it('pressing A shows inline input for child entry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/add child entry/i)).toBeInTheDocument()
    })
  })

  it('submitting inline input calls AddEntry and refreshes data', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('r')

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/type entry/i)).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText(/type entry/i)
    await user.type(input, '. New root task{Enter}')

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalledWith('. New root task', expect.any(String))
    })
  })

  it('pressing Escape cancels inline input', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('r')

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/type entry/i)).toBeInTheDocument()
    })

    await user.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByPlaceholderText(/type entry/i)).not.toBeInTheDocument()
    })
  })

  it('clicking + button shows inline input for root entry', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const addButton = screen.getByTitle('Add new entry')
    await user.click(addButton)

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/type entry/i)).toBeInTheDocument()
    })
  })
})

describe('App - Go to Today', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('shows Go to today button when viewing a different day', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Navigate to previous day
    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    // Go to today button should appear (distinct from sidebar Today nav button)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /go to today/i })).toBeInTheDocument()
    })
  })

  it('hides Go to today button when viewing today', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Go to today button should NOT be visible when viewing today
    expect(screen.queryByRole('button', { name: /go to today/i })).not.toBeInTheDocument()
  })

  it('clicking Go to today button navigates back to today', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Navigate to previous day
    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /go to today/i })).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    // Click Go to today button
    const todayButton = screen.getByRole('button', { name: /go to today/i })
    await user.click(todayButton)

    // Should trigger data refresh
    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })

    // Go to today button should disappear after navigating back to today
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /go to today/i })).not.toBeInTheDocument()
    })
  })

  it('pressing T navigates to today when viewing a different day', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Navigate to previous day
    const prevButton = screen.getByRole('button', { name: /previous day/i })
    await user.click(prevButton)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /go to today/i })).toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    // Press T to go to today
    await user.keyboard('T')

    // Should trigger data refresh
    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })

    // Go to today button should disappear
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /go to today/i })).not.toBeInTheDocument()
    })
  })
})

describe('App - Review View (formerly Past Week)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows "Weekly Review" label in sidebar navigation', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Sidebar should show "Weekly Review" for the week/review view
    expect(screen.getByRole('button', { name: /weekly review/i })).toBeInTheDocument()
  })

  it('shows "Weekly Review" as header title when review view is selected', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Click on Weekly Review in sidebar
    const reviewButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(reviewButton)

    // Header title should show "Weekly Review"
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /weekly review/i })).toBeInTheDocument()
    })
  })

  it('shows navigation controls in review view', async () => {
    const user = userEvent.setup()
    render(<App />)

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
    render(<App />)

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
