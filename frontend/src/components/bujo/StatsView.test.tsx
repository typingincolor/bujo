import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { StatsView } from './StatsView'
import { Habit, Goal, DayEntries } from '@/types/bujo'
import { format } from 'date-fns'
import { GetDayEntries } from '@/wailsjs/go/wails/App'
import { service } from '@/wailsjs/go/models'

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetDayEntries: vi.fn().mockResolvedValue([]),
}))

vi.mock('./ActivityHeatmap', () => ({
  ActivityHeatmap: () => <div data-testid="activity-heatmap" />,
}))

vi.mock('./TrendsChart', () => ({
  TrendsChart: () => <div data-testid="trends-chart" />,
}))

vi.mock('./TaskDurationChart', () => ({
  TaskDurationChart: () => <div data-testid="task-duration-chart" />,
}))

const mockedGetDayEntries = vi.mocked(GetDayEntries)

const currentMonth = format(new Date(), 'yyyy-MM')

const createTestDay = (entries: DayEntries['entries'] = []): DayEntries => ({
  date: new Date().toISOString(),
  entries,
})

function toApiEntries(entries: DayEntries['entries']) {
  return entries.map(e => ({
    ID: e.id,
    Content: e.content,
    Type: e.type.charAt(0).toUpperCase() + e.type.slice(1),
    Priority: e.priority ? e.priority.charAt(0).toUpperCase() + e.priority.slice(1) : 'None',
    ParentID: e.parentId ?? 0,
    CreatedAt: e.loggedDate || new Date().toISOString(),
    Children: (e.children || []).map((c: DayEntries['entries'][0]) => ({
      ID: c.id,
      Content: c.content,
      Type: c.type.charAt(0).toUpperCase() + c.type.slice(1),
      Priority: 'None',
      ParentID: c.parentId ?? 0,
      CreatedAt: c.loggedDate || new Date().toISOString(),
      Children: [],
    })),
  }))
}

function toApiDays(days: DayEntries[]) {
  return days.map(d => ({
    Date: d.date,
    Entries: toApiEntries(d.entries),
    Location: '',
    Mood: '',
    Weather: '',
  }))
}

const createTestHabit = (overrides: Partial<Habit> = {}): Habit => ({
  id: 1,
  name: 'Test habit',
  todayLogged: false,
  todayCount: 0,
  streak: 0,
  completionRate: 0,
  dayHistory: [],
  ...overrides,
})

const createTestGoal = (overrides: Partial<Goal> = {}): Goal => ({
  id: 1,
  content: 'Test goal',
  month: currentMonth,
  status: 'active',
  ...overrides,
})

describe('StatsView', () => {
  beforeEach(() => {
    mockedGetDayEntries.mockReset()
    mockedGetDayEntries.mockResolvedValue([])
  })

  it('renders stats title', async () => {
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/insights/i)).toBeInTheDocument()
    })
  })

  it('displays total entry count', async () => {
    const days = [createTestDay([
      { id: 1, content: 'Task 1', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Note 1', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    mockedGetDayEntries.mockResolvedValue(toApiDays(days) as unknown as service.DayEntries[])
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText('2')).toBeInTheDocument()
    })
    expect(screen.getByText(/total entries/i)).toBeInTheDocument()
  })

  it('displays task count and percentage', async () => {
    const days = [createTestDay([
      { id: 1, content: 'Task 1', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Task 2', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 3, content: 'Note 1', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    mockedGetDayEntries.mockResolvedValue(toApiDays(days) as unknown as service.DayEntries[])
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/67%/)).toBeInTheDocument()
    })
    expect(screen.getByText(/tasks/i)).toBeInTheDocument()
  })

  it('displays completion rate', async () => {
    const days = [createTestDay([
      { id: 1, content: 'Done task', type: 'done', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Pending task', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    mockedGetDayEntries.mockResolvedValue(toApiDays(days) as unknown as service.DayEntries[])
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/50%/)).toBeInTheDocument()
    })
    expect(screen.getByText(/completion rate/i)).toBeInTheDocument()
  })

  it('displays active habits count', async () => {
    const habits = [
      createTestHabit({ id: 1, name: 'Habit 1' }),
      createTestHabit({ id: 2, name: 'Habit 2' }),
    ]
    render(<StatsView habits={habits} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/active habits/i)).toBeInTheDocument()
    })
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('displays best streak', async () => {
    const habits = [
      createTestHabit({ id: 1, streak: 5 }),
      createTestHabit({ id: 2, streak: 12 }),
      createTestHabit({ id: 3, streak: 3 }),
    ]
    render(<StatsView habits={habits} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/best streak/i)).toBeInTheDocument()
    })
    expect(screen.getByText(/12 days/i)).toBeInTheDocument()
  })

  it('displays monthly goals progress', async () => {
    const goals = [
      createTestGoal({ id: 1, status: 'done' }),
      createTestGoal({ id: 2, status: 'done' }),
      createTestGoal({ id: 3, status: 'active' }),
    ]
    render(<StatsView habits={[]} goals={goals} />)
    await waitFor(() => {
      expect(screen.getByText(/monthly goals/i)).toBeInTheDocument()
    })
    expect(screen.getByText('2/3')).toBeInTheDocument()
  })

  it('shows zero stats when no data', async () => {
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/total entries/i)).toBeInTheDocument()
    })
    expect(screen.getAllByText('0').length).toBeGreaterThan(0)
  })

  it('displays note count', async () => {
    const days = [createTestDay([
      { id: 1, content: 'Note 1', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Note 2', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    mockedGetDayEntries.mockResolvedValue(toApiDays(days) as unknown as service.DayEntries[])
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/notes/i)).toBeInTheDocument()
    })
  })

  it('displays event count', async () => {
    const days = [createTestDay([
      { id: 1, content: 'Event 1', type: 'event', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    mockedGetDayEntries.mockResolvedValue(toApiDays(days) as unknown as service.DayEntries[])
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByText(/events/i)).toBeInTheDocument()
    })
  })

  it('renders ActivityHeatmap', async () => {
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByTestId('activity-heatmap')).toBeInTheDocument()
    })
  })

  it('renders TrendsChart', async () => {
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByTestId('trends-chart')).toBeInTheDocument()
    })
  })

  it('renders TaskDurationChart', async () => {
    render(<StatsView habits={[]} goals={[]} />)
    await waitFor(() => {
      expect(screen.getByTestId('task-duration-chart')).toBeInTheDocument()
    })
  })
})
