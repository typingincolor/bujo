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

import { GetAgenda } from './wailsjs/go/wails/App'

// Week data with multiple days containing various entry types
const weekAgendaWithTaskFlow = createMockAgenda({
  Days: [
    // Day 1: Monday - 2 tasks created, 1 done
    createMockDayEntries({
      Date: '2026-01-19T00:00:00Z',
      Entries: [
        createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Task created Monday', CreatedAt: '2026-01-19T10:00:00Z' }),
        createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Done', Content: 'Task completed Monday', CreatedAt: '2026-01-19T11:00:00Z' }),
      ],
    }),
    // Day 2: Tuesday - 1 task migrated, 1 new task
    createMockDayEntries({
      Date: '2026-01-20T00:00:00Z',
      Entries: [
        createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Migrated', Content: 'Task migrated Tuesday', CreatedAt: '2026-01-20T10:00:00Z' }),
        createMockEntry({ ID: 4, EntityID: 'e4', Type: 'Task', Content: 'Task created Tuesday', CreatedAt: '2026-01-20T11:00:00Z' }),
      ],
    }),
    // Day 3: Wednesday - 2 tasks done
    createMockDayEntries({
      Date: '2026-01-21T00:00:00Z',
      Entries: [
        createMockEntry({ ID: 5, EntityID: 'e5', Type: 'Done', Content: 'Task completed Wednesday 1', CreatedAt: '2026-01-21T10:00:00Z' }),
        createMockEntry({ ID: 6, EntityID: 'e6', Type: 'Done', Content: 'Task completed Wednesday 2', CreatedAt: '2026-01-21T11:00:00Z' }),
      ],
    }),
  ],
  Overdue: [],
})

// Week data with meetings (events with children)
const weekAgendaWithMeetings = createMockAgenda({
  Days: [
    createMockDayEntries({
      Date: '2026-01-19T00:00:00Z',
      Entries: [
        // Meeting 1: Event with 2 children
        createMockEntry({ ID: 10, EntityID: 'e10', Type: 'Event', Content: 'Team standup', CreatedAt: '2026-01-19T09:00:00Z' }),
        createMockEntry({ ID: 11, EntityID: 'e11', Type: 'Note', Content: 'Discussed sprint goals', ParentID: 10, Depth: 1, CreatedAt: '2026-01-19T09:15:00Z' }),
        createMockEntry({ ID: 12, EntityID: 'e12', Type: 'Task', Content: 'Follow up with team', ParentID: 10, Depth: 1, CreatedAt: '2026-01-19T09:20:00Z' }),
        // Meeting 2: Event with 3 children
        createMockEntry({ ID: 13, EntityID: 'e13', Type: 'Event', Content: 'Project review', CreatedAt: '2026-01-19T14:00:00Z' }),
        createMockEntry({ ID: 14, EntityID: 'e14', Type: 'Note', Content: 'Review notes', ParentID: 13, Depth: 1, CreatedAt: '2026-01-19T14:15:00Z' }),
        createMockEntry({ ID: 15, EntityID: 'e15', Type: 'Task', Content: 'Update documentation', ParentID: 13, Depth: 1, CreatedAt: '2026-01-19T14:20:00Z' }),
        createMockEntry({ ID: 16, EntityID: 'e16', Type: 'Task', Content: 'Schedule follow-up', ParentID: 13, Depth: 1, CreatedAt: '2026-01-19T14:25:00Z' }),
        // Event without children (should NOT appear in meetings list)
        createMockEntry({ ID: 17, EntityID: 'e17', Type: 'Event', Content: 'Lunch break', CreatedAt: '2026-01-19T12:00:00Z' }),
      ],
    }),
  ],
  Overdue: [],
})

// Week data with items needing attention
const weekAgendaWithNeedsAttention = createMockAgenda({
  Days: [
    createMockDayEntries({
      Date: '2026-01-19T00:00:00Z',
      Entries: [
        // High priority task (should appear first in needs attention)
        createMockEntry({ ID: 20, EntityID: 'e20', Type: 'Task', Content: 'Urgent deadline task', Priority: 'High', CreatedAt: '2026-01-19T10:00:00Z' }),
        // Old open task (high attention score due to age)
        createMockEntry({ ID: 21, EntityID: 'e21', Type: 'Task', Content: 'Old unfinished task', CreatedAt: '2026-01-12T10:00:00Z' }),
        // Unanswered question
        createMockEntry({ ID: 22, EntityID: 'e22', Type: 'Question', Content: 'Should we refactor this?', CreatedAt: '2026-01-19T11:00:00Z' }),
        // Normal task (lower attention score)
        createMockEntry({ ID: 23, EntityID: 'e23', Type: 'Task', Content: 'Regular task', CreatedAt: '2026-01-19T12:00:00Z' }),
        // Done task (should NOT appear in needs attention)
        createMockEntry({ ID: 24, EntityID: 'e24', Type: 'Done', Content: 'Completed task', CreatedAt: '2026-01-19T13:00:00Z' }),
      ],
    }),
  ],
  Overdue: [],
})

describe('WeekSummary - Task Flow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows week summary section at top of weekly view', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithTaskFlow)

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    // Navigate to weekly review
    const reviewButton = screen.getByRole('button', { name: /weekly review/i })
    await user.click(reviewButton)

    await waitFor(() => {
      expect(screen.getByTestId('week-summary')).toBeInTheDocument()
    })
  })

  it('shows task flow section with Created count', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithTaskFlow)

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
      // Should show Task Flow section
      expect(screen.getByText(/task flow/i)).toBeInTheDocument()
      // Should show Created count (2 tasks created: Monday and Tuesday)
      expect(screen.getByTestId('task-flow-created')).toHaveTextContent('2')
    })
  })

  it('shows task flow section with Done count', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithTaskFlow)

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
      // Should show Done count (3 tasks done: 1 Monday, 2 Wednesday)
      expect(screen.getByTestId('task-flow-done')).toHaveTextContent('3')
    })
  })

  it('shows task flow section with Migrated count', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithTaskFlow)

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
      // Should show Migrated count (1 task migrated on Tuesday)
      expect(screen.getByTestId('task-flow-migrated')).toHaveTextContent('1')
    })
  })

  it('shows task flow section with Open count', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithTaskFlow)

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
      // Should show Open count (2 tasks still open: Task created Monday, Task created Tuesday)
      expect(screen.getByTestId('task-flow-open')).toHaveTextContent('2')
    })
  })
})

describe('WeekSummary - Meetings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows meetings section with events that have children', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithMeetings)

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
      // Should show Meetings section
      expect(screen.getByText(/meetings/i)).toBeInTheDocument()
      // Should show meetings with children (scoped to week-summary to avoid duplicates in daily entries)
      const weekSummary = screen.getByTestId('week-summary')
      expect(within(weekSummary).getByText('Team standup')).toBeInTheDocument()
      expect(within(weekSummary).getByText('Project review')).toBeInTheDocument()
    })
  })

  it('shows child count for each meeting', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithMeetings)

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
      // Team standup has 2 children
      expect(screen.getByText(/2 items/i)).toBeInTheDocument()
      // Project review has 3 children
      expect(screen.getByText(/3 items/i)).toBeInTheDocument()
    })
  })

  it('does NOT show events without children in meetings list', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithMeetings)

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
      const weekSummary = screen.getByTestId('week-summary')
      expect(within(weekSummary).getByText('Team standup')).toBeInTheDocument()
    })

    // "Lunch break" is an event without children - should NOT be in meetings section
    const meetingsSection = screen.getByTestId('week-summary-meetings')
    expect(meetingsSection).not.toHaveTextContent('Lunch break')
  })
})

describe('WeekSummary - Needs Attention', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows needs attention section', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithNeedsAttention)

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
      expect(screen.getByText(/needs attention/i)).toBeInTheDocument()
    })
  })

  it('shows open tasks sorted by attention score (high priority first)', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithNeedsAttention)

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
      const attentionSection = screen.getByTestId('week-summary-attention')
      // High priority task should appear (first due to priority)
      expect(attentionSection).toHaveTextContent('Urgent deadline task')
      // Old task should appear (high score due to age)
      expect(attentionSection).toHaveTextContent('Old unfinished task')
    })
  })

  it('shows unanswered questions in needs attention', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithNeedsAttention)

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
      const attentionSection = screen.getByTestId('week-summary-attention')
      expect(attentionSection).toHaveTextContent('Should we refactor this?')
    })
  })

  it('shows attention indicator for high-priority items', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithNeedsAttention)

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
      // High priority item should have a priority indicator (scoped to week-summary)
      const weekSummary = screen.getByTestId('week-summary')
      const highPriorityItem = within(weekSummary).getByText('Urgent deadline task').closest('[data-attention-item]')
      expect(highPriorityItem).toHaveAttribute('data-priority', 'high')
    })
  })

  it('does NOT show completed tasks in needs attention', async () => {
    const user = userEvent.setup()
    vi.mocked(GetAgenda).mockResolvedValue(weekAgendaWithNeedsAttention)

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
      expect(screen.getByTestId('week-summary-attention')).toBeInTheDocument()
    })

    // Completed task should NOT appear in needs attention
    const attentionSection = screen.getByTestId('week-summary-attention')
    expect(attentionSection).not.toHaveTextContent('Completed task')
  })
})
