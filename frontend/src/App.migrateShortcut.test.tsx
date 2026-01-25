import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

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
  CreateHabit: vi.fn().mockResolvedValue(1),
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue(['Home', 'Office']),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
}))

import { GetAgenda } from './wailsjs/go/wails/App'

describe('App - Migrate Keyboard Shortcut', () => {
  const mockTaskEntry = createMockAgenda({
    Days: [createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Task to migrate', CreatedAt: '2026-01-17T10:00:00Z' }),
      ],
    })],
    Overdue: [],
  })

  const mockQuestionEntry = createMockAgenda({
    Days: [createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Question', Content: 'Question to migrate', CreatedAt: '2026-01-17T10:00:00Z' }),
      ],
    })],
    Overdue: [],
  })

  const mockNonMigratableEntries = createMockAgenda({
    Days: [createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Done', Content: 'Done task', CreatedAt: '2026-01-17T10:00:00Z' }),
        createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Note', Content: 'A note', CreatedAt: '2026-01-17T11:00:00Z' }),
        createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Event', Content: 'An event', CreatedAt: '2026-01-17T12:00:00Z' }),
      ],
    })],
    Overdue: [],
  })

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('pressing m opens migrate modal when a task entry is selected', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockTaskEntry)
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Task to migrate')).toBeInTheDocument()
    })

    await user.keyboard('m')

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
    })
  })

  it('pressing m opens migrate modal when a question entry is selected', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockQuestionEntry)
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Question to migrate')).toBeInTheDocument()
    })

    await user.keyboard('m')

    await waitFor(() => {
      expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
    })
  })

  it('pressing m does NOT open migrate modal for done entries', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockNonMigratableEntries)
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Done task')).toBeInTheDocument()
    })

    // First entry is Done type
    await user.keyboard('m')

    // Modal should NOT open
    await new Promise(resolve => setTimeout(resolve, 100))
    expect(screen.queryByText('Migrate Entry')).not.toBeInTheDocument()
  })

  it('pressing m does NOT open migrate modal for note entries', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockNonMigratableEntries)
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('A note')).toBeInTheDocument()
    })

    // Navigate to second entry (Note)
    await user.keyboard('j')
    await user.keyboard('m')

    // Modal should NOT open
    await new Promise(resolve => setTimeout(resolve, 100))
    expect(screen.queryByText('Migrate Entry')).not.toBeInTheDocument()
  })

  it('pressing m does NOT open migrate modal for event entries', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(mockNonMigratableEntries)
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('An event')).toBeInTheDocument()
    })

    // Navigate to third entry (Event)
    await user.keyboard('j')
    await user.keyboard('j')
    await user.keyboard('m')

    // Modal should NOT open
    await new Promise(resolve => setTimeout(resolve, 100))
    expect(screen.queryByText('Migrate Entry')).not.toBeInTheDocument()
  })

  it('pressing m does nothing when no entries exist', async () => {
    vi.mocked(GetAgenda).mockResolvedValue(createMockAgenda({
      Days: [createMockDayEntries({ Entries: [] })],
      Overdue: [],
    }))
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    await user.keyboard('m')

    // Modal should NOT open
    await new Promise(resolve => setTimeout(resolve, 100))
    expect(screen.queryByText('Migrate Entry')).not.toBeInTheDocument()
  })
})
