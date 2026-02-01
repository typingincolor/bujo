import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { format } from 'date-fns'
import { QuickStats } from './QuickStats'
import { DayEntries, Habit, Goal } from '@/types/bujo'

const currentMonth = format(new Date(), 'yyyy-MM')

const mockDays: DayEntries[] = [
  {
    date: '2026-01-25T00:00:00Z',
    entries: [
      { id: 1, content: 'Task 1', type: 'done', priority: 'none', parentId: null, loggedDate: '2026-01-25T10:00:00Z' },
      { id: 2, content: 'Task 2', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-25T11:00:00Z' },
    ],
  },
]

const mockHabits: Habit[] = [
  { id: 1, name: 'Exercise', streak: 5, completionRate: 0.8, dayHistory: [], todayLogged: true, todayCount: 1 },
  { id: 2, name: 'Reading', streak: 3, completionRate: 0.6, dayHistory: [], todayLogged: false, todayCount: 0 },
]

const mockGoals: Goal[] = [
  { id: 1, content: 'Learn TypeScript', month: currentMonth, status: 'done' },
  { id: 2, content: 'Build app', month: currentMonth, status: 'active' },
]

describe('QuickStats', () => {
  const defaultProps = {
    days: mockDays,
    habits: mockHabits,
    goals: mockGoals,
    overdueCount: 3,
  }

  it('renders compact card styling with reduced padding', () => {
    render(<QuickStats {...defaultProps} />)
    const cards = [
      screen.getByTestId('stat-card-tasks-completed'),
      screen.getByTestId('stat-card-pending-tasks'),
      screen.getByTestId('stat-card-habits-today'),
      screen.getByTestId('stat-card-monthly-goals'),
    ]
    cards.forEach(card => {
      expect(card).toHaveClass('p-3')
    })
  })

  it('uses smaller text sizes for metrics', () => {
    render(<QuickStats {...defaultProps} />)
    const counts = screen.getAllByTestId('stat-count')
    counts.forEach(count => {
      expect(count).toHaveClass('text-xl')
    })
  })

  it('uses smaller icon sizes', () => {
    render(<QuickStats {...defaultProps} />)
    const icons = screen.getAllByTestId('stat-icon')
    icons.forEach(icon => {
      expect(icon).toHaveClass('h-4', 'w-4')
    })
  })

  it('renders all four stat cards', () => {
    render(<QuickStats {...defaultProps} />)
    expect(screen.getByText('Tasks Completed')).toBeInTheDocument()
    expect(screen.getByText('Pending Tasks')).toBeInTheDocument()
    expect(screen.getByText('Habits Today')).toBeInTheDocument()
    expect(screen.getByText('Monthly Goals')).toBeInTheDocument()
  })

  it('displays correct values', () => {
    render(<QuickStats {...defaultProps} />)
    // 1 done task
    expect(screen.getByTestId('stat-card-tasks-completed')).toHaveTextContent('1')
    // 3 overdue
    expect(screen.getByTestId('stat-card-pending-tasks')).toHaveTextContent('3')
    // 1/2 habits
    expect(screen.getByTestId('stat-card-habits-today')).toHaveTextContent('1/2')
    // 1/2 goals
    expect(screen.getByTestId('stat-card-monthly-goals')).toHaveTextContent('1/2')
  })
})
