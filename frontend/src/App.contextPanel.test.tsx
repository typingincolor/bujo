import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

const mockEntriesAgenda = createMockAgenda({
  Days: [createMockDayEntries({
    Entries: [
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Parent task', CreatedAt: '2026-01-17T10:00:00Z' }),
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Child task', ParentID: 1, Depth: 1, CreatedAt: '2026-01-17T11:00:00Z' }),
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
  AddChildEntry: vi.fn().mockResolvedValue([1]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  MoveEntryToList: vi.fn().mockResolvedValue(undefined),
  MoveEntryToRoot: vi.fn().mockResolvedValue(undefined),
  CreateHabit: vi.fn().mockResolvedValue(1),
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue(['Home', 'Office']),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
  Search: vi.fn().mockResolvedValue([]),
}))

import { GetAgenda } from './wailsjs/go/wails/App'

describe('App context panel toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('toggles context panel with Shift+C key', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Parent task')).toBeInTheDocument()
    })

    // Initially panel is hidden
    expect(screen.queryByTestId('context-panel')).not.toBeInTheDocument()

    // Press Shift+C to show (uppercase C simulates Shift+c)
    await user.keyboard('C')
    expect(screen.getByTestId('context-panel')).toBeInTheDocument()

    // Press Shift+C again to hide
    await user.keyboard('C')
    expect(screen.queryByTestId('context-panel')).not.toBeInTheDocument()
  })

  it('does not toggle context panel with lowercase c (reserved for capture modal)', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Parent task')).toBeInTheDocument()
    })

    // Initially panel is hidden
    expect(screen.queryByTestId('context-panel')).not.toBeInTheDocument()

    // Press lowercase c - should NOT show context panel (it opens CaptureModal instead)
    await user.keyboard('c')
    expect(screen.queryByTestId('context-panel')).not.toBeInTheDocument()
  })

  it('shows selected entry context in panel when in today view', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Parent task')).toBeInTheDocument()
    })

    // Navigate to second entry (child task) using j
    await user.keyboard('j')

    // Open context panel (uppercase C simulates Shift+c)
    await user.keyboard('C')

    // Panel should show the child task context
    await waitFor(() => {
      const panel = screen.getByTestId('context-panel')
      expect(panel).toBeInTheDocument()
      // Should show child task in context
      expect(panel).toHaveTextContent('Child task')
    })
  })

  it('context panel is available in habits view', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to habits view (sidebar button is named "Habit Tracker")
    const habitsButton = screen.getByRole('button', { name: /habit tracker/i })
    await user.click(habitsButton)

    // Just wait for the click to process and view to change
    await waitFor(() => {
      const headings = screen.getAllByRole('heading', { name: /habit tracker/i })
      expect(headings.length).toBeGreaterThan(0)
    })

    // Open context panel with Shift+C (uppercase C)
    await user.keyboard('C')
    expect(screen.getByTestId('context-panel')).toBeInTheDocument()
  })

  it('context panel is available in search view', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to search view
    const searchButton = screen.getByRole('button', { name: /^search$/i })
    await user.click(searchButton)

    // Wait for view to change - search view should show a search input
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
    })

    // Open context panel with Shift+C (uppercase C)
    await user.keyboard('C')
    expect(screen.getByTestId('context-panel')).toBeInTheDocument()
  })

  it('updates context panel when entry is clicked', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Parent task')).toBeInTheDocument()
    })

    // Show context panel first with Shift+C
    await user.keyboard('C')
    expect(screen.getByTestId('context-panel')).toBeInTheDocument()

    // Click on child entry
    await user.click(screen.getByText('Child task'))

    // Panel should show the child task
    await waitFor(() => {
      const panel = screen.getByTestId('context-panel')
      expect(panel).toHaveTextContent('Child task')
    })
  })

  it('updates context panel when search result is clicked', async () => {
    // Mock search results
    const { Search } = await import('./wailsjs/go/wails/App')
    vi.mocked(Search).mockResolvedValue([
      createMockEntry({ ID: 10, EntityID: 'e10', Type: 'Task', Content: 'Search result task', ParentID: null, CreatedAt: '2026-01-15T10:00:00Z' }),
    ])

    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to search view
    const searchButton = screen.getByRole('button', { name: /^search$/i })
    await user.click(searchButton)

    // Wait for search view
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
    })

    // Open context panel with Shift+C
    await user.keyboard('C')
    expect(screen.getByTestId('context-panel')).toBeInTheDocument()

    // Type in search box and search
    const searchInput = screen.getByPlaceholderText(/search entries/i)
    await user.type(searchInput, 'Search result')
    await user.keyboard('{Enter}')

    // Wait for search results
    await waitFor(() => {
      expect(screen.getByText('Search result task')).toBeInTheDocument()
    })

    // Click on search result
    await user.click(screen.getByText('Search result task'))

    // Panel should show the search result entry
    await waitFor(() => {
      const panel = screen.getByTestId('context-panel')
      expect(panel).toHaveTextContent('Search result task')
    })
  })
})
