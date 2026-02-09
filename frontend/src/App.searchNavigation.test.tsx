import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'

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
  GetDayEntries: vi.fn().mockResolvedValue([{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }]),
  GetOverdue: vi.fn().mockResolvedValue([]),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  GetWeekSummary: vi.fn().mockResolvedValue({ Days: [] }),
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
  GetAllTags: vi.fn().mockResolvedValue([]),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
  GetEditableDocument: vi.fn().mockResolvedValue(''),
  ValidateEditableDocument: vi.fn().mockResolvedValue({ isValid: true, errors: [] }),
  ApplyEditableDocument: vi.fn().mockResolvedValue({ inserted: 0, deleted: 0 }),
  SearchEntries: vi.fn().mockResolvedValue([]),
  Search: vi.fn().mockResolvedValue([]),
  GetEntry: vi.fn().mockResolvedValue(null),
  RetypeEntry: vi.fn().mockResolvedValue(undefined),
  GetStats: vi.fn().mockResolvedValue({
    TotalEntries: 0,
    TasksCompleted: 0,
    ActiveHabits: 0,
    CurrentStreak: 0,
  }),
  GetVersion: vi.fn().mockResolvedValue('1.0.0'),
  GetAllMentions: vi.fn().mockResolvedValue([]),
  SearchByMentions: vi.fn().mockResolvedValue([]),
}))

import { Search, GetDayEntries } from './wailsjs/go/wails/App'

const createMockSearchResult = (overrides: Partial<{ ID: number; Content: string; Type: string; CreatedAt: string; ParentID: number | null; Priority: string }>) => ({
  ID: 1,
  EntityID: 'test-entity',
  Type: 'task',
  Content: 'Test content',
  Priority: 'none',
  ParentID: null,
  Depth: 0,
  CreatedAt: '2026-01-15T10:00:00Z',
  convertValues: vi.fn(),
  ...overrides,
})

describe('App - Search Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('navigates to journal view when double-clicking a search result', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockSearchResult({
        ID: 42,
        Content: 'Test entry from January 15th',
        Type: 'task',
        CreatedAt: '2026-01-15T10:00:00Z',
      }),
    ] as never)

    vi.mocked(GetDayEntries).mockResolvedValue([
      { Date: '2026-01-15T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' },
    ] as never)

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
    const searchButton = screen.getByRole('button', { name: /search/i })
    await user.click(searchButton)

    // Wait for search view to render
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
    })

    // Type a search query
    const searchInput = screen.getByPlaceholderText(/search entries/i)
    await user.type(searchInput, 'test')

    // Wait for search results
    await waitFor(() => {
      expect(screen.getByText('Test entry from January 15th')).toBeInTheDocument()
    })

    // Double-click the search result
    const result = screen.getByText('Test entry from January 15th').closest('[data-result-id]')
    expect(result).toBeInTheDocument()
    await user.dblClick(result!)

    // Should navigate to journal view (shows the CodeMirror editor)
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    // Header should show 'Journal' title indicating we're on the today view
    // Use getAllByText since 'Journal' appears in both sidebar and header
    const journalElements = screen.getAllByText('Journal')
    expect(journalElements.length).toBeGreaterThanOrEqual(2) // sidebar + header
  })

  it('shows back button on journal view after navigating from search', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockSearchResult({
        ID: 42,
        Content: 'Test entry',
        Type: 'task',
        CreatedAt: '2026-01-15T10:00:00Z',
      }),
    ] as never)

    vi.mocked(GetDayEntries).mockResolvedValue([
      { Date: '2026-01-15T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' },
    ] as never)

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
    await user.click(screen.getByRole('button', { name: /search/i }))

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
    })

    // Back button should NOT appear on search view
    expect(screen.queryByRole('button', { name: /go back/i })).not.toBeInTheDocument()

    // Search and double-click result
    await user.type(screen.getByPlaceholderText(/search entries/i), 'test')
    await waitFor(() => {
      expect(screen.getByText('Test entry')).toBeInTheDocument()
    })

    const result = screen.getByText('Test entry').closest('[data-result-id]')
    await user.dblClick(result!)

    // Now on journal view - back button SHOULD appear
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()
    })
  })

  it('back button returns to search view after double-click navigation', async () => {
    vi.mocked(Search).mockResolvedValue([
      createMockSearchResult({
        ID: 42,
        Content: 'Test entry',
        Type: 'task',
        CreatedAt: '2026-01-15T10:00:00Z',
      }),
    ] as never)

    vi.mocked(GetDayEntries).mockResolvedValue([
      { Date: '2026-01-15T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' },
    ] as never)

    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to search, double-click result
    await user.click(screen.getByRole('button', { name: /search/i }))
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
    })

    await user.type(screen.getByPlaceholderText(/search entries/i), 'test')
    await waitFor(() => {
      expect(screen.getByText('Test entry')).toBeInTheDocument()
    })

    await user.dblClick(screen.getByText('Test entry').closest('[data-result-id]')!)

    // Now on journal view
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    // Click back button
    await user.click(screen.getByRole('button', { name: /go back/i }))

    // Should be back on search view
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
    })
  })
})
