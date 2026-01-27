import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries } from './test/mocks'
import type { service, domain } from './wailsjs/go/models'
import { startOfDay, subDays, format } from 'date-fns'

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

import { GetDayEntries, GetOverdue, AddEntry, AddChildEntry } from './wailsjs/go/wails/App'

const mockStorage: Record<string, string> = {}
const mockLocalStorage = {
  getItem: vi.fn((key: string) => mockStorage[key] || null),
  setItem: vi.fn((key: string, value: string) => { mockStorage[key] = value }),
  removeItem: vi.fn((key: string) => { delete mockStorage[key] }),
  clear: vi.fn(() => { Object.keys(mockStorage).forEach(key => delete mockStorage[key]) }),
}

Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

describe('CaptureBar - Uses currentDate (not new Date())', () => {
  // Use real "today" and calculate yesterday from it
  const today = startOfDay(new Date())
  const yesterday = subDays(today, 1)
  const yesterdayStr = format(yesterday, 'yyyy-MM-dd')

  const mockTodayDays: service.DayEntries[] = [createMockDayEntries({
    Date: today.toISOString(),
    Entries: [
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Today task', CreatedAt: today.toISOString() }),
    ],
  })]

  const mockYesterdayDays: service.DayEntries[] = [createMockDayEntries({
    Date: yesterday.toISOString(),
    Entries: [
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Yesterday task', CreatedAt: yesterday.toISOString() }),
    ],
  })]

  const mockOverdue: domain.Entry[] = []

  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockTodayDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('AddEntry receives currentDate when submitting after navigating to past day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Today task')).toBeInTheDocument()
    })

    // Navigate to previous day (h key) - this updates currentDate to yesterday
    vi.mocked(GetDayEntries).mockResolvedValue(mockYesterdayDays)
    await user.keyboard('h')

    await waitFor(() => {
      expect(screen.getByText('Yesterday task')).toBeInTheDocument()
    })

    // Clear mock calls from navigation
    vi.mocked(AddEntry).mockClear()

    // Type and submit an entry
    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'New entry for yesterday{Enter}')

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalled()
    })

    // Verify AddEntry was called with yesterday's date (the currentDate after navigation)
    const [, dateArg] = vi.mocked(AddEntry).mock.calls[0]
    expect(dateArg).toContain(yesterdayStr)
  })

  it('AddChildEntry receives currentDate when submitting child after navigating to past day', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Today task')).toBeInTheDocument()
    })

    // Navigate to previous day (h key) - this updates currentDate to yesterday
    vi.mocked(GetDayEntries).mockResolvedValue(mockYesterdayDays)
    await user.keyboard('h')

    await waitFor(() => {
      expect(screen.getByText('Yesterday task')).toBeInTheDocument()
    })

    // Clear mock calls from navigation
    vi.mocked(AddChildEntry).mockClear()

    // Enter child mode for the entry (A key)
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    // Type and submit a child entry
    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Child entry for yesterday{Enter}')

    await waitFor(() => {
      expect(AddChildEntry).toHaveBeenCalled()
    })

    // Verify AddChildEntry was called with yesterday's date (the currentDate after navigation)
    const [, , dateArg] = vi.mocked(AddChildEntry).mock.calls[0]
    expect(dateArg).toContain(yesterdayStr)
  })
})
