import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockDays, createMockOverdue } from './test/mocks'

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
  GetAllTags: vi.fn().mockResolvedValue([]),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
  GetEditableDocument: vi.fn().mockResolvedValue(''),
  ValidateEditableDocument: vi.fn().mockResolvedValue({ isValid: true, errors: [] }),
  ApplyEditableDocument: vi.fn().mockResolvedValue({ inserted: 0, deleted: 0 }),
}))

import { GetDayEntries, GetOverdue } from './wailsjs/go/wails/App'

const weekDays = createMockDays([
  createMockDayEntries({
    Date: '2026-01-19T00:00:00Z',
    Entries: [
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Monday task', CreatedAt: '2026-01-19T10:00:00Z' }),
    ],
  }),
  createMockDayEntries({
    Date: '2026-01-20T00:00:00Z',
    Entries: [
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Tuesday task', CreatedAt: '2026-01-20T10:00:00Z' }),
    ],
  }),
])
const weekOverdue = createMockOverdue([])

describe('WeekView Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('pressing [ toggles context panel collapsed state in week view', async () => {
    const user = userEvent.setup()
    vi.mocked(GetDayEntries).mockResolvedValue(weekDays)
    vi.mocked(GetOverdue).mockResolvedValue(weekOverdue)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Switch to week view
    const reviewButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(reviewButton)

    await waitFor(() => {
      const main = screen.getByRole('main')
      expect(within(main).getByLabelText('Toggle context panel')).toBeInTheDocument()
    })

    // Context panel should be collapsed initially (no Context heading visible)
    const main = screen.getByRole('main')
    expect(within(main).queryByText('Context')).not.toBeInTheDocument()

    // Press [ to expand
    await user.keyboard('{[}')

    // Context panel should be visible (Context heading appears)
    await waitFor(() => {
      expect(within(main).getByText('Context')).toBeInTheDocument()
    })

    // Press [ again to collapse
    await user.keyboard('{[}')

    // Context panel should be hidden again
    await waitFor(() => {
      expect(within(main).queryByText('Context')).not.toBeInTheDocument()
    })
  })

  it('renders WeekView when view is "week"', async () => {
    const user = userEvent.setup()
    vi.mocked(GetDayEntries).mockResolvedValue(weekDays)
    vi.mocked(GetOverdue).mockResolvedValue(weekOverdue)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const reviewButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(reviewButton)

    await waitFor(() => {
      const main = screen.getByRole('main')
      // Context panel is collapsed by default, so Context heading should NOT be visible
      expect(within(main).queryByText('Context')).not.toBeInTheDocument()
      // Toggle button should be visible
      expect(within(main).getByLabelText('Toggle context panel')).toBeInTheDocument()
    })
  })
})
