import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatsView } from './StatsView'
import { DayEntries, Habit, Goal } from '@/types/bujo'
import { format } from 'date-fns'

const currentMonth = format(new Date(), 'yyyy-MM')

const createTestDay = (entries: DayEntries['entries'] = []): DayEntries => ({
  date: new Date().toISOString(),
  entries,
})

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
  completed: false,
  ...overrides,
})

describe('StatsView', () => {
  it('renders stats title', () => {
    render(<StatsView days={[]} habits={[]} goals={[]} />)
    expect(screen.getByText(/statistics/i)).toBeInTheDocument()
  })

  it('displays total entry count', () => {
    const days = [createTestDay([
      { id: 1, content: 'Task 1', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Note 1', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    render(<StatsView days={days} habits={[]} goals={[]} />)
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText(/total entries/i)).toBeInTheDocument()
  })

  it('displays task count and percentage', () => {
    const days = [createTestDay([
      { id: 1, content: 'Task 1', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Task 2', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 3, content: 'Note 1', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    render(<StatsView days={days} habits={[]} goals={[]} />)
    expect(screen.getByText(/tasks/i)).toBeInTheDocument()
    expect(screen.getByText(/67%/)).toBeInTheDocument()
  })

  it('displays completion rate', () => {
    const days = [createTestDay([
      { id: 1, content: 'Done task', type: 'done', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Pending task', type: 'task', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    render(<StatsView days={days} habits={[]} goals={[]} />)
    expect(screen.getByText(/completion rate/i)).toBeInTheDocument()
    expect(screen.getByText(/50%/)).toBeInTheDocument()
  })

  it('displays active habits count', () => {
    const habits = [
      createTestHabit({ id: 1, name: 'Habit 1' }),
      createTestHabit({ id: 2, name: 'Habit 2' }),
    ]
    render(<StatsView days={[]} habits={habits} goals={[]} />)
    expect(screen.getByText(/active habits/i)).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('displays best streak', () => {
    const habits = [
      createTestHabit({ id: 1, streak: 5 }),
      createTestHabit({ id: 2, streak: 12 }),
      createTestHabit({ id: 3, streak: 3 }),
    ]
    render(<StatsView days={[]} habits={habits} goals={[]} />)
    expect(screen.getByText(/best streak/i)).toBeInTheDocument()
    expect(screen.getByText(/12 days/i)).toBeInTheDocument()
  })

  it('displays monthly goals progress', () => {
    const goals = [
      createTestGoal({ id: 1, completed: true }),
      createTestGoal({ id: 2, completed: true }),
      createTestGoal({ id: 3, completed: false }),
    ]
    render(<StatsView days={[]} habits={[]} goals={goals} />)
    expect(screen.getByText(/monthly goals/i)).toBeInTheDocument()
    expect(screen.getByText('2/3')).toBeInTheDocument()
  })

  it('shows zero stats when no data', () => {
    render(<StatsView days={[]} habits={[]} goals={[]} />)
    expect(screen.getByText(/total entries/i)).toBeInTheDocument()
    expect(screen.getAllByText('0').length).toBeGreaterThan(0)
  })

  it('displays note count', () => {
    const days = [createTestDay([
      { id: 1, content: 'Note 1', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
      { id: 2, content: 'Note 2', type: 'note', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    render(<StatsView days={days} habits={[]} goals={[]} />)
    expect(screen.getByText(/notes/i)).toBeInTheDocument()
  })

  it('displays event count', () => {
    const days = [createTestDay([
      { id: 1, content: 'Event 1', type: 'event', priority: 'none', parentId: null, loggedDate: '', children: [] },
    ])]
    render(<StatsView days={days} habits={[]} goals={[]} />)
    expect(screen.getByText(/events/i)).toBeInTheDocument()
  })
})
