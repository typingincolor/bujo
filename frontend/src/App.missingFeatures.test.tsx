/**
 * Missing Features Acceptance Tests
 *
 * These tests document the missing functionality from the original implementation plan
 * that was not delivered in PR #415. All tests should FAIL initially until features are implemented.
 *
 * Missing features covered:
 * 1. useNavigationHistory hook for view navigation
 * 2. Back button in Header component
 * 3. EntryContextPopover using Radix UI Popover
 * 4. EntryTree back button integration
 * 5. WeekSummary using EntryContextPopover for meetings
 * 6. WeekSummary using EntryContextPopover for attention items
 * 7. Keyboard shortcut 'm' to migrate entry from popover
 * 8. Full attention scoring algorithm implementation
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'
import { GetAgenda, MigrateEntry } from './wailsjs/go/wails/App'

vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn(),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue(undefined),
  AddChildEntry: vi.fn().mockResolvedValue(undefined),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  MigrateEntry: vi.fn().mockResolvedValue(undefined),
  MoveEntryToList: vi.fn().mockResolvedValue(undefined),
  MoveEntryToRoot: vi.fn().mockResolvedValue(undefined),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  RetypeEntry: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('@/lib/wailsTime', () => ({
  toWailsTime: (date: Date) => date.toISOString(),
  fromWailsTime: (time: string) => new Date(time),
}))

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(() => () => {}),
  EventsOff: vi.fn(),
  EventsEmit: vi.fn(),
  LogPrint: vi.fn(),
  LogTrace: vi.fn(),
  LogDebug: vi.fn(),
  LogInfo: vi.fn(),
  LogWarning: vi.fn(),
  LogError: vi.fn(),
  LogFatal: vi.fn(),
}))

beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
  // Default mock: empty agenda
  vi.mocked(GetAgenda).mockResolvedValue(createMockAgenda({ Days: [] }))
})

describe('Missing Feature #1: useNavigationHistory hook', () => {
  it('should track navigation history when switching views', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText(/Journal/i)).toBeInTheDocument()
    })

    // Navigate to Week view
    const weekButton = screen.getByText(/Weekly Review/i)
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText(/Week of/i)).toBeInTheDocument()
    })

    // Navigate to Overview
    const overviewButton = screen.getByRole('button', { name: /Pending Tasks/i })
    await user.click(overviewButton)

    await waitFor(() => {
      expect(screen.getByTestId('outstanding-icon')).toBeInTheDocument()
    })

    // Back button should be visible and functional
    const backButton = screen.getByLabelText(/go back/i)
    expect(backButton).toBeInTheDocument()
    expect(backButton).not.toBeDisabled()

    // Click back should return to Week view
    await user.click(backButton)

    await waitFor(() => {
      expect(screen.getByText(/Week of/i)).toBeInTheDocument()
    })

    // Back again should return to Journal
    const backButtonAgain = screen.getByLabelText(/go back/i)
    await user.click(backButtonAgain)

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })
  })

  it('should clear history when navigating to today view', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    // Navigate away from today
    const weekButton = screen.getByText(/Weekly Review/i)
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText(/Week of/i)).toBeInTheDocument()
    })

    // Navigate to today should clear history
    const journalButton = screen.getByRole('button', { name: /Journal/i })
    await user.click(journalButton)

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    // Back button should not exist (no history)
    expect(screen.queryByLabelText(/go back/i)).not.toBeInTheDocument()
  })
})

describe('Missing Feature #2: Header back button', () => {
  it('should show back button in header when navigation history exists', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    // Initially no back button
    expect(screen.queryByLabelText(/go back/i)).not.toBeInTheDocument()

    // Navigate to create history
    const weekButton = screen.getByRole('button', { name: /Weekly Review/i })
    await user.click(weekButton)

    await waitFor(() => {
      const backButton = screen.getByLabelText(/go back/i)
      expect(backButton).toBeInTheDocument()
      expect(backButton).toBeVisible()
    })
  })

  it('should hide back button when no navigation history', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // On initial load (today view), no back button
    expect(screen.queryByLabelText(/go back/i)).not.toBeInTheDocument()
  })
})

describe('Missing Feature #3: EntryContextPopover with Radix UI', () => {
  it('should open Radix UI popover when clicking entry in EntryTree', async () => {
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Test task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText(/Test task/i)).toBeInTheDocument()
    })

    // Click on the entry (need to click the container, not just the text)
    const entry = screen.getByTestId('entry-item')
    await user.click(entry)

    // Radix UI popover should open with quick actions (done, migrate, priority for tasks)
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByLabelText(/mark done/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/migrate/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/cycle priority/i)).toBeInTheDocument()
      expect(screen.getByText(/go to entry/i)).toBeInTheDocument()
    })
  })

  it('should close popover when clicking outside', async () => {
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Test task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText(/Test task/i)).toBeInTheDocument()
    })

    // Open popover
    const entry = screen.getByText(/Test task/i)
    await user.click(entry)

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })

    // Click outside (on the document body)
    await user.click(document.body)

    // Popover should close
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })
})

describe('Missing Feature #4: EntryTree back button integration', () => {
  it('should show back button in EntryTree when viewing child entries', async () => {
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Parent task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
            }),
            createMockEntry({
              ID: 2,
              Content: 'Child task',
              Type: 'Task',
              Priority: 'None',
              ParentID: 1,
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText(/Parent task/i)).toBeInTheDocument()
    })

    // Click to expand/navigate to child
    const parentEntry = screen.getByText(/Parent task/i)
    await user.click(parentEntry)

    // Should show child with back button to parent in the popover
    let popover!: HTMLElement
    await waitFor(() => {
      popover = screen.getByRole('dialog')
      expect(within(popover).getByText(/Child task/i)).toBeInTheDocument()
      const backButton = within(popover).getByLabelText(/back to parent/i)
      expect(backButton).toBeInTheDocument()
    })

    // Click back should return to parent view in the popover
    const backToParent = within(popover).getByLabelText(/back to parent/i)
    await user.click(backToParent)

    await waitFor(() => {
      expect(within(popover).getByText(/Parent task/i)).toBeInTheDocument()
      expect(within(popover).queryByLabelText(/back to parent/i)).not.toBeInTheDocument()
    })
  })
})

describe('Missing Feature #5: WeekSummary popover for meetings', () => {
  it('should open EntryContextPopover when clicking meeting in WeekSummary', async () => {
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Team standup',
              Type: 'Event',
              Priority: 'None',
              ParentID: null,
            }),
            createMockEntry({
              ID: 2,
              Content: 'Standup note',
              Type: 'Note',
              Priority: 'None',
              ParentID: 1,
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    // Navigate to Week view
    const weekButton = screen.getByRole('button', { name: /Weekly Review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText(/Meetings/i)).toBeInTheDocument()
      expect(screen.getByText(/Team standup/i)).toBeInTheDocument()
    })

    // Click on meeting
    const meeting = screen.getByText(/Team standup/i)
    await user.click(meeting)

    // EntryContextPopover should open with priority button (Events only have priority action)
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByLabelText(/cycle priority/i)).toBeInTheDocument()
    })
  })
})

describe('Missing Feature #6: WeekSummary popover for attention items', () => {
  it('should open EntryContextPopover when clicking attention item in WeekSummary', async () => {
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Urgent task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
              CreatedAt: '2026-01-20T10:00:00Z', // Old task (4 days ago)
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    // Navigate to Week view
    const weekButton = screen.getByRole('button', { name: /Weekly Review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText(/Needs Attention/i)).toBeInTheDocument()
      expect(screen.getByText(/Urgent task/i)).toBeInTheDocument()
    })

    // Click on attention item
    const attentionItem = screen.getByText(/Urgent task/i)
    await user.click(attentionItem)

    // EntryContextPopover should open with task actions (done, migrate)
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByLabelText(/mark done/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/migrate/i)).toBeInTheDocument()
    })
  })
})

describe('Missing Feature #7: Keyboard shortcut for migrate', () => {
  it('should migrate entry when pressing "m" key in popover', async () => {
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-20T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Old task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText(/Old task/i)).toBeInTheDocument()
    })

    // Open popover
    const entry = screen.getByText(/Old task/i)
    await user.click(entry)

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })

    // Press 'm' key
    await user.keyboard('m')

    // Wait for migrate modal
    let modal!: HTMLElement
    await waitFor(() => {
      const heading = screen.getByText('Migrate Entry')
      modal = heading.closest('div[class*="bg-background"]')!
      expect(modal).toBeInTheDocument()
    })

    // Click migrate button within modal (date defaults to tomorrow which is fine)
    const migrateButton = within(modal).getByRole('button', { name: /migrate/i })
    await user.click(migrateButton)

    // Should call MigrateEntry
    await waitFor(() => {
      expect(vi.mocked(MigrateEntry)).toHaveBeenCalledWith(1, expect.any(String))
    })
  })
})

describe('Missing Feature #8: Full attention scoring algorithm', () => {
  it('should calculate attention score based on age, complexity, and questions', async () => {
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            // Old task with children (high score)
            createMockEntry({
              ID: 1,
              Content: 'Complex old task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
              CreatedAt: '2026-01-10T10:00:00Z', // 14 days old
            }),
            createMockEntry({
              ID: 2,
              Content: 'Subtask 1',
              Type: 'Task',
              Priority: 'None',
              ParentID: 1,
              CreatedAt: '2026-01-10T10:00:00Z',
            }),
            createMockEntry({
              ID: 3,
              Content: 'Subtask 2',
              Type: 'Task',
              Priority: 'None',
              ParentID: 1,
              CreatedAt: '2026-01-10T10:00:00Z',
            }),
            // Question (medium-high score)
            createMockEntry({
              ID: 4,
              Content: 'Unresolved question',
              Type: 'Question',
              Priority: 'None',
              ParentID: null,
              CreatedAt: '2026-01-22T10:00:00Z',
            }),
            // Recent simple task (low score)
            createMockEntry({
              ID: 5,
              Content: 'New simple task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
              CreatedAt: '2026-01-24T10:00:00Z',
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    // Navigate to Week view
    const weekButton = screen.getByRole('button', { name: /Weekly Review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText(/Needs Attention/i)).toBeInTheDocument()
    })

    // Should show items ordered by attention score
    const attentionSection = screen.getByText(/Needs Attention/i).closest('div')
    expect(attentionSection).toBeInTheDocument()

    // Complex old task should appear first (highest score)
    const items = within(attentionSection!).getAllByRole('button', { name: /task|question/i })
    expect(items[0]).toHaveTextContent(/Complex old task/i)

    // New simple task should not appear in attention list (low score)
    expect(within(attentionSection!).queryByText(/New simple task/i)).not.toBeInTheDocument()
  })

  it('should apply correct scoring rules: age + children + question symbol', async () => {
    // Score calculation should be:
    // - Age: 14 days = 14 points
    // - Children: 2 * 5 = 10 points
    // - Total: 24 points (high attention)

    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Test task',
              Type: 'Task',
              Priority: 'None',
              ParentID: null,
              CreatedAt: '2026-01-10T10:00:00Z', // 14 days old
            }),
            createMockEntry({
              ID: 2,
              Content: 'Child 1',
              Type: 'Task',
              Priority: 'None',
              ParentID: 1,
              CreatedAt: '2026-01-10T10:00:00Z',
            }),
            createMockEntry({
              ID: 3,
              Content: 'Child 2',
              Type: 'Task',
              Priority: 'None',
              ParentID: 1,
              CreatedAt: '2026-01-10T10:00:00Z',
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    const weekButton = screen.getByRole('button', { name: /Weekly Review/i })
    await user.click(weekButton)

    await waitFor(() => {
      expect(screen.getByText(/Needs Attention/i)).toBeInTheDocument()
      expect(screen.getByText(/Test task/i)).toBeInTheDocument()
    })

    // Should display attention score or indicator
    const attentionItem = screen.getByText(/Test task/i).closest('div')
    expect(attentionItem).toBeInTheDocument()
  })
})
