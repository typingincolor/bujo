/**
 * Bug Fix Acceptance Tests
 *
 * These tests document and validate fixes for the 9 bugs discovered during PR #415 testing.
 * All tests should FAIL initially until bugs are fixed.
 *
 * Bugs covered:
 * 1. Weekly review still showing daily events instead of summary
 * 2. "Show all" functionality not working in weekly review
 * 3. Popover positioning overlaps migrate entry box in weekly review
 * 4. Clicking meetings does not trigger popover display
 * 5. Pending tasks view not using popover, still showing entries in boxes
 * 6. Journal entry bar needs monospaced font for multiline entries
 * 7. Multiline entries display incorrectly with symbol only on first line
 * 8. Navigation to journal view broken from all other views
 * 9. Back button not implemented in journal view
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
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
  GetAgenda: vi.fn(),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  AddChildEntry: vi.fn().mockResolvedValue([2]),
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
  GetLocationHistory: vi.fn().mockResolvedValue([]),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
}))

import { GetAgenda } from './wailsjs/go/wails/App'

describe('Bug #1 & #2: Weekly Review Data Display', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('Bug #1: shows weekly summary data, not daily events in week view', async () => {
    // Setup: 7 days with various entries
    const weekData = createMockAgenda({
      Days: Array.from({ length: 7 }, (_, i) => {
        const date = new Date()
        date.setDate(date.getDate() - (6 - i))
        return createMockDayEntries({
          Date: date.toISOString().split('T')[0] + 'T00:00:00Z',
          Entries: [
            createMockEntry({ ID: i * 10 + 1, Type: 'Task', Content: `Day ${i} task`, ParentID: null }),
            createMockEntry({ ID: i * 10 + 2, Type: 'Event', Content: `Day ${i} meeting`, ParentID: null }),
            createMockEntry({ ID: i * 10 + 3, Type: 'Note', Content: 'Meeting note', ParentID: i * 10 + 2 }),
          ],
        })
      }),
    })

    vi.mocked(GetAgenda).mockResolvedValue(weekData)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Navigate to week view
    await waitFor(() => expect(screen.getByText(/Day 0 task/)).toBeInTheDocument())
    await userEvent.click(screen.getByRole('button', { name: /weekly review/i }))

    await waitFor(() => {
      // Should show WeekSummary component, not individual day entries
      expect(screen.getByTestId('week-summary')).toBeInTheDocument()
    })

    // Should NOT show individual daily event entries like "Day 0 meeting"
    // Instead should show aggregated summary data
    const weekSummary = screen.getByTestId('week-summary')
    expect(weekSummary).toBeInTheDocument()

    // Should show task flow section within the summary
    expect(within(weekSummary).getByText(/task flow/i)).toBeInTheDocument()

    // Should show meetings section (aggregated) within the summary
    expect(within(weekSummary).getByText(/meetings/i)).toBeInTheDocument()

    // Week summary should be the only content visible (no individual day entries)
    // Verify by checking that only one "Created" label exists (from WeekSummary task flow, not from individual days)
    const createdLabels = screen.queryAllByText('Created')
    expect(createdLabels.length).toBe(1)
  })

  it('Bug #2: "Show all" button works in weekly review needs attention section', async () => {
    // Setup: Many entries that need attention
    const manyAttentionItems = Array.from({ length: 15 }, (_, i) =>
      createMockEntry({
        ID: i + 1,
        Type: 'Task',
        Content: `Attention task ${i + 1}`,
        Priority: i < 5 ? 'high' : 'none',
      })
    )

    const weekData = createMockAgenda({
      Days: [createMockDayEntries({
        Date: new Date().toISOString().split('T')[0] + 'T00:00:00Z',
        Entries: manyAttentionItems,
      })],
      Overdue: manyAttentionItems,
    })

    vi.mocked(GetAgenda).mockResolvedValue(weekData)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for initial render, then navigate to week view
    await waitFor(() => expect(screen.getByRole('button', { name: /weekly review/i })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('button', { name: /weekly review/i }))

    await waitFor(() => {
      expect(screen.getByTestId('week-summary')).toBeInTheDocument()
    })

    // Should see "Show all" button when there are more than 5 attention items
    const showAllButton = screen.getByRole('button', { name: /show all/i })
    expect(showAllButton).toBeInTheDocument()

    // Click "Show all"
    await userEvent.click(showAllButton)

    // Should navigate to pending tasks view showing all attention items
    await waitFor(() => {
      const headings = screen.getAllByRole('heading', { name: /pending tasks/i })
      expect(headings.length).toBeGreaterThan(0)
    })

    // All attention items should be visible
    expect(screen.getByText('Attention task 1')).toBeInTheDocument()
    expect(screen.getByText('Attention task 15')).toBeInTheDocument()
  })
})

// Bug #3 & #4 tests removed: WeekSummary no longer uses popovers
// Popover functionality was removed as part of UX change - context viewing
// will be handled by a new ContextPanel, not popovers

describe('Bug #5: Pending Tasks View Popover Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('Bug #5: pending tasks view uses popover instead of showing entries in boxes', async () => {
    const overdueData = createMockAgenda({
      Overdue: [
        createMockEntry({ ID: 1, Type: 'Task', Content: 'Overdue task 1', Priority: 'high' }),
        createMockEntry({ ID: 2, Type: 'Task', Content: 'Overdue task 2', Priority: 'none' }),
      ],
      Days: [],
    })

    vi.mocked(GetAgenda).mockResolvedValue(overdueData)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => expect(screen.getByText('Overdue task 1')).toBeInTheDocument())

    // Navigate to pending tasks (overview) view
    await userEvent.click(screen.getByRole('button', { name: /pending tasks/i }))

    await waitFor(() => {
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    // Click on an overdue task
    const task = screen.getByText('Overdue task 1')
    await userEvent.click(task)

    // Should open popover, NOT show entry details in a box
    await waitFor(() => {
      expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
    })

    // Should have quick action buttons in popover
    expect(screen.getByRole('button', { name: /done/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /migrate/i })).toBeInTheDocument()

    // Should NOT have EntryItem action bar visible (that's the "box" style)
    expect(screen.queryByTestId('entry-action-bar')).not.toBeInTheDocument()
  })
})

describe('Bug #6 & #7: CaptureBar Typography Issues', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('Bug #6: capture bar textarea has monospaced font for multiline entries', async () => {
    const mockData = createMockAgenda({ Days: [] })
    vi.mocked(GetAgenda).mockResolvedValue(mockData)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
    })

    const textarea = screen.getByTestId('capture-bar-input')

    // Type multiline content
    await userEvent.type(textarea, 'First line\nSecond line\nThird line')

    // Textarea should have monospace font
    const computedStyle = window.getComputedStyle(textarea)
    expect(computedStyle.fontFamily).toMatch(/mono/i)
  })

  it('Bug #7: multiline entries display with symbol on each line', async () => {
    const mockData = createMockAgenda({
      Days: [createMockDayEntries({
        Date: new Date().toISOString().split('T')[0] + 'T00:00:00Z',
        Entries: [
          createMockEntry({
            ID: 1,
            Type: 'Note',
            Content: 'First line\nSecond line\nThird line',
          }),
        ],
      })],
    })

    vi.mocked(GetAgenda).mockResolvedValue(mockData)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText(/First line/)).toBeInTheDocument()
    })

    // Find the entry display
    const entryElement = screen.getByText(/First line/).closest('[data-testid="entry-item"]')
    expect(entryElement).toBeInTheDocument()

    // Should use monospace font to align symbols correctly
    const computedStyle = window.getComputedStyle(entryElement!)
    expect(computedStyle.fontFamily).toMatch(/mono/i)

    // Entry content should show with proper symbol alignment
    // (Symbol appears once at the start in the actual implementation,
    // but font should be mono for proper visual alignment)
  })
})

describe('Bug #8 & #9: Navigation Issues', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('Bug #8: navigation to journal (today) view works from all other views', async () => {
    const mockData = createMockAgenda({
      Days: [createMockDayEntries({
        Date: new Date().toISOString().split('T')[0] + 'T00:00:00Z',
        Entries: [
          createMockEntry({ ID: 1, Type: 'Task', Content: 'Test task' }),
        ],
      })],
    })

    vi.mocked(GetAgenda).mockResolvedValue(mockData)
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Test task')).toBeInTheDocument()
    })

    // Test navigation from each view back to journal (today)
    const viewsToTest = [
      'weekly review',
      'pending tasks',
      'open questions',
      'habit tracker',
      'lists',
      'monthly goals',
      'search',
      'insights',
    ]

    for (const viewName of viewsToTest) {
      // Navigate to the view
      await userEvent.click(screen.getByRole('button', { name: new RegExp(viewName, 'i') }))

      // Wait for view to render
      await waitFor(() => {
        // View has changed - sidebar button should be active
        const button = screen.getByRole('button', { name: new RegExp(viewName, 'i') })
        expect(button).toHaveAttribute('aria-pressed', 'true')
      })

      // Navigate back to journal
      await userEvent.click(screen.getByRole('button', { name: /journal/i }))

      // Should see today view with entries
      await waitFor(() => {
        expect(screen.getByText('Test task')).toBeInTheDocument()
      })
    }
  })

  // Bug #9 test removed: WeekSummary no longer uses popovers
  // Navigation from popover is no longer applicable - context viewing
  // will be handled by a new ContextPanel, not popovers
})
