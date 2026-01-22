import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { HabitTracker } from './HabitTracker'
import { Habit } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  LogHabit: vi.fn().mockResolvedValue(undefined),
  CreateHabit: vi.fn().mockResolvedValue(1),
  DeleteHabit: vi.fn().mockResolvedValue(undefined),
  UndoHabitLog: vi.fn().mockResolvedValue(undefined),
  UndoHabitLogForDate: vi.fn().mockResolvedValue(undefined),
  SetHabitGoal: vi.fn().mockResolvedValue(undefined),
  LogHabitForDate: vi.fn().mockResolvedValue(undefined),
}))

import { LogHabitForDate } from '@/wailsjs/go/wails/App'

const createTestHabit = (overrides: Partial<Habit> = {}): Habit => ({
  id: 1,
  name: 'Test Habit',
  goal: 1,
  streak: 0,
  completionRate: 0,
  todayLogged: false,
  todayCount: 0,
  dayHistory: [
    { date: '2024-01-01', completed: false, count: 0 },
    { date: '2024-01-02', completed: false, count: 0 },
    { date: '2024-01-03', completed: false, count: 0 },
    { date: '2024-01-04', completed: false, count: 0 },
    { date: '2024-01-05', completed: false, count: 0 },
    { date: '2024-01-06', completed: false, count: 0 },
    { date: '2024-01-07', completed: false, count: 0 },
  ],
  ...overrides,
})

describe('HabitTracker - Click to Log', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-07T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('logs habit directly when clicking on a day circle', () => {
    const onHabitChanged = vi.fn()
    // Dec 29, 2024 is a Sunday, so week is Dec 29 - Jan 4
    const anchor = new Date('2024-01-01')
    const habit = createTestHabit({
      id: 42,
      dayHistory: [
        { date: '2023-12-29', completed: false, count: 0 },
        { date: '2023-12-30', completed: false, count: 0 },
        { date: '2023-12-31', completed: false, count: 0 },
        { date: '2024-01-01', completed: false, count: 0 },
        { date: '2024-01-02', completed: false, count: 0 },
        { date: '2024-01-03', completed: false, count: 0 },
        { date: '2024-01-04', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} onHabitChanged={onHabitChanged} anchorDate={anchor} />)

    // Click on a day circle
    const dayButton = screen.getByLabelText(/Log for 2024-01-01$/i)
    fireEvent.click(dayButton)

    // Should call LogHabitForDate immediately without confirmation
    expect(LogHabitForDate).toHaveBeenCalled()
    const call = vi.mocked(LogHabitForDate).mock.calls[0]
    expect(call[0]).toBe(42) // habit ID
    expect(call[1]).toBe(1) // count
  })

  it('displays count in day circle when habit is logged', () => {
    // Anchor to Jan 7 (system time) so Jan 1-7 are all visible
    const anchor = new Date('2024-01-07')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2024-01-07', completed: false, count: 0 },
        { date: '2024-01-06', completed: false, count: 0 },
        { date: '2024-01-05', completed: false, count: 0 },
        { date: '2024-01-04', completed: false, count: 0 },
        { date: '2024-01-03', completed: true, count: 1 },
        { date: '2024-01-02', completed: false, count: 0 },
        { date: '2024-01-01', completed: true, count: 3 },
      ]
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // Find cells by their dates and verify the count is displayed
    const loggedCell1 = screen.getByLabelText(/Logged for 2024-01-01/i)
    expect(loggedCell1).toHaveTextContent('3')
    const loggedCell2 = screen.getByLabelText(/Logged for 2024-01-03/i)
    expect(loggedCell2).toHaveTextContent('1')
  })
})

describe('HabitTracker - Week View Header', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-15T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows day labels (S M T W T F S) only once at the top in week view', () => {
    const anchor = new Date('2024-01-15')
    const habits = [
      createTestHabit({ id: 1, name: 'Habit 1' }),
      createTestHabit({ id: 2, name: 'Habit 2' }),
      createTestHabit({ id: 3, name: 'Habit 3' }),
    ]
    render(<HabitTracker habits={habits} anchorDate={anchor} />)

    // With 3 habits, there should be exactly 2 'S' letters (for Sunday and Saturday)
    // NOT 6 (2 per habit row) - header should only appear once at the top
    const sundayLabels = screen.getAllByText('S')
    expect(sundayLabels).toHaveLength(2) // Only Sunday and Saturday from single header

    // Verify other day labels appear only once each
    expect(screen.getAllByText('M')).toHaveLength(1)
    expect(screen.getAllByText('W')).toHaveLength(1)
    expect(screen.getAllByText('F')).toHaveLength(1)
    // T appears twice (Tuesday and Thursday) in the single header
    expect(screen.getAllByText('T')).toHaveLength(2)
  })

  it('shows Today button in week view', () => {
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} />)

    expect(screen.getByRole('button', { name: /today/i })).toBeInTheDocument()
  })

  it('calls onNavigate with current date when Today button is clicked', () => {
    const onNavigate = vi.fn()
    // Anchor to past week (Jan 1)
    const anchor = new Date('2024-01-01')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} onNavigate={onNavigate} />)

    fireEvent.click(screen.getByRole('button', { name: /today/i }))

    expect(onNavigate).toHaveBeenCalled()
    // The new anchor should be today (Jan 15, 2024 from fake timer)
    const newAnchor = onNavigate.mock.calls[0][0] as Date
    expect(newAnchor.getDate()).toBe(15)
    expect(newAnchor.getMonth()).toBe(0) // January
    expect(newAnchor.getFullYear()).toBe(2024)
  })
})

describe('HabitTracker - Week View Dynamic Day Labels', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    // Wednesday Jan 22, 2025
    vi.setSystemTime(new Date('2025-01-22T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows day labels matching actual days displayed (past 7 days ending with anchor)', () => {
    // Anchor is Wednesday Jan 22, 2025
    // Past 7 days: Thu 16, Fri 17, Sat 18, Sun 19, Mon 20, Tue 21, Wed 22
    // Labels should be: T, F, S, S, M, T, W (not S, M, T, W, T, F, S)
    const anchor = new Date('2025-01-22')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2025-01-16', completed: false, count: 0 },
        { date: '2025-01-17', completed: false, count: 0 },
        { date: '2025-01-18', completed: false, count: 0 },
        { date: '2025-01-19', completed: false, count: 0 },
        { date: '2025-01-20', completed: false, count: 0 },
        { date: '2025-01-21', completed: false, count: 0 },
        { date: '2025-01-22', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // The header should show labels matching the actual days:
    // Thu=T, Fri=F, Sat=S, Sun=S, Mon=M, Tue=T, Wed=W
    // So we expect: T, F, S, S, M, T, W
    const headerLabels = screen.getAllByText(/^[SMTWF]$/)

    // Should have exactly 7 labels in the header
    expect(headerLabels).toHaveLength(7)

    // Extract text content to verify order
    const labelTexts = headerLabels.map(el => el.textContent)
    expect(labelTexts).toEqual(['T', 'F', 'S', 'S', 'M', 'T', 'W'])
  })
})

describe('HabitTracker - No Re-render Animation', () => {
  it('habit rows do not have slide-in animation that would flicker on re-render', () => {
    const habit = createTestHabit({ name: 'Exercise' })
    const { container } = render(<HabitTracker habits={[habit]} />)

    // Find the habit row container
    const habitRow = container.querySelector('.group')
    expect(habitRow).toBeInTheDocument()

    // Should NOT have animate-slide-in class (causes flicker on re-render)
    expect(habitRow).not.toHaveClass('animate-slide-in')
  })
})

describe('HabitTracker - Calendar Grid View', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-15T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders CalendarNavigation with correct label for week view', () => {
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} />)

    // Should show navigation with week label (past 7 days ending with anchor)
    expect(screen.getByLabelText('Previous')).toBeInTheDocument()
    expect(screen.getByLabelText('Next')).toBeInTheDocument()
    expect(screen.getByText(/Jan 9.*Jan 15.*2024/)).toBeInTheDocument()
  })

  it('renders CalendarNavigation with correct label for month view', () => {
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} />)

    // Switch to month view using fireEvent (works with fake timers)
    fireEvent.click(screen.getByRole('button', { name: /week/i }))
    fireEvent.click(screen.getByRole('button', { name: /month/i }))

    expect(screen.getByText('January 2024')).toBeInTheDocument()
  })

  it('calls onNavigate when prev button clicked', () => {
    const onNavigate = vi.fn()
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} onNavigate={onNavigate} />)

    fireEvent.click(screen.getByLabelText('Previous'))

    expect(onNavigate).toHaveBeenCalled()
    // The new anchor should be 7 days before (week view default)
    const newAnchor = onNavigate.mock.calls[0][0] as Date
    expect(newAnchor.getDate()).toBe(8) // Jan 15 - 7 = Jan 8
  })

  it('calls onNavigate when next button clicked', () => {
    const onNavigate = vi.fn()
    // Use anchor before today (Jan 8) so next button is enabled
    const anchor = new Date('2024-01-08')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} onNavigate={onNavigate} />)

    fireEvent.click(screen.getByLabelText('Next'))

    expect(onNavigate).toHaveBeenCalled()
    // The new anchor should be 7 days after (week view default)
    const newAnchor = onNavigate.mock.calls[0][0] as Date
    expect(newAnchor.getDate()).toBe(15) // Jan 8 + 7 = Jan 15
  })

  it('renders week view as single row calendar grid', () => {
    const anchor = new Date('2024-01-15')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2024-01-14', completed: false, count: 0 },
        { date: '2024-01-15', completed: true, count: 1 },
        { date: '2024-01-16', completed: false, count: 0 },
        { date: '2024-01-17', completed: false, count: 0 },
        { date: '2024-01-18', completed: false, count: 0 },
        { date: '2024-01-19', completed: false, count: 0 },
        { date: '2024-01-20', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // Week view should show day-of-week header
    expect(screen.getAllByText('S').length).toBeGreaterThanOrEqual(2) // Sunday and Saturday
    expect(screen.getByText('M')).toBeInTheDocument()
    expect(screen.getByText('W')).toBeInTheDocument()
    expect(screen.getByText('F')).toBeInTheDocument()
  })

  it('renders month view as multi-row calendar grid', () => {
    const anchor = new Date('2024-01-15')
    const habit = createTestHabit({
      dayHistory: Array.from({ length: 31 }, (_, i) => ({
        date: `2024-01-${String(i + 1).padStart(2, '0')}`,
        completed: false,
        count: 0,
      }))
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // Switch to month view
    fireEvent.click(screen.getByRole('button', { name: /week/i }))
    fireEvent.click(screen.getByRole('button', { name: /month/i }))

    // Should show day numbers from the month - use aria-label to be specific
    expect(screen.getByLabelText(/Log for 2024-01-15/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Log for 2024-01-01/i)).toBeInTheDocument()
  })

  it('renders quarter view with three month calendars (past months)', () => {
    const anchor = new Date('2024-01-15')
    const habit = createTestHabit({
      dayHistory: Array.from({ length: 90 }, (_, i) => {
        // Generate dates from November 2023 through January 2024
        const date = new Date('2023-11-01')
        date.setDate(date.getDate() + i)
        return {
          date: date.toISOString().split('T')[0],
          completed: false,
          count: 0,
        }
      })
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // Switch to quarter view
    fireEvent.click(screen.getByRole('button', { name: /week/i }))
    fireEvent.click(screen.getByRole('button', { name: /quarter/i }))

    // Should show three month names (past: Nov, Dec, Jan)
    expect(screen.getByText('November')).toBeInTheDocument()
    expect(screen.getByText('December')).toBeInTheDocument()
    expect(screen.getByText('January')).toBeInTheDocument()
  })

  it('disables next button when anchor is at today (week view)', () => {
    // System time is Jan 15, 2024
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} />)

    // Next button should be disabled because navigating forward would show future dates
    expect(screen.getByLabelText('Next')).toBeDisabled()
  })

  it('enables next button when anchor is before today (week view)', () => {
    // System time is Jan 15, 2024, anchor is Jan 8
    const anchor = new Date('2024-01-08')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} />)

    // Next button should be enabled because there's room to navigate forward
    expect(screen.getByLabelText('Next')).toBeEnabled()
  })

  it('disables next button when anchor is at today (month view)', () => {
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} />)

    // Switch to month view
    fireEvent.click(screen.getByRole('button', { name: /week/i }))
    fireEvent.click(screen.getByRole('button', { name: /month/i }))

    // Next button should be disabled in current month
    expect(screen.getByLabelText('Next')).toBeDisabled()
  })

  it('disables next button when anchor is at today (quarter view)', () => {
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} />)

    // Switch to quarter view
    fireEvent.click(screen.getByRole('button', { name: /week/i }))
    fireEvent.click(screen.getByRole('button', { name: /quarter/i }))

    // Next button should be disabled when anchor is in the current quarter
    expect(screen.getByLabelText('Next')).toBeDisabled()
  })
})
