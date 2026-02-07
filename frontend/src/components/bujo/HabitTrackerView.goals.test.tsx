import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HabitTracker } from './HabitTrackerView'
import { Habit } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  LogHabit: vi.fn().mockResolvedValue(undefined),
  CreateHabit: vi.fn().mockResolvedValue(1),
  DeleteHabit: vi.fn().mockResolvedValue(undefined),
  UndoHabitLog: vi.fn().mockResolvedValue(undefined),
  UndoHabitLogForDate: vi.fn().mockResolvedValue(undefined),
  SetHabitGoal: vi.fn().mockResolvedValue(undefined),
  SetHabitWeeklyGoal: vi.fn().mockResolvedValue(undefined),
  SetHabitMonthlyGoal: vi.fn().mockResolvedValue(undefined),
  LogHabitForDate: vi.fn().mockResolvedValue(undefined),
}))

import { SetHabitGoal, SetHabitWeeklyGoal, SetHabitMonthlyGoal } from '@/wailsjs/go/wails/App'

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

describe('HabitTracker - Goal Display', () => {
  it('shows Target icon with goal number instead of text', () => {
    const habit = createTestHabit({ goal: 3 })
    render(<HabitTracker habits={[habit]} />)

    // Should NOT show "Goal: 3/day" text
    expect(screen.queryByText(/Goal:.*3.*day/i)).not.toBeInTheDocument()

    // Should show Target icon with goal number
    const goalIndicator = screen.getByLabelText(/daily goal.*3/i)
    expect(goalIndicator).toBeInTheDocument()
  })

  it('does not show goal indicator when goal is not set', () => {
    const habit = createTestHabit({ goal: undefined })
    render(<HabitTracker habits={[habit]} />)

    expect(screen.queryByLabelText(/daily goal/i)).not.toBeInTheDocument()
  })
})

describe('HabitTracker - Set Habit Goal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows goal button on habit row', () => {
    render(<HabitTracker habits={[createTestHabit()]} />)
    expect(screen.getByTitle('Set goal')).toBeInTheDocument()
  })

  it('shows goal input when goal button is clicked and type selected', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit()]} />)

    await user.click(screen.getByTitle('Set goal'))
    await user.click(screen.getByText(/daily/i))

    expect(screen.getByPlaceholderText(/daily goal/i)).toBeInTheDocument()
  })

  it('calls SetHabitGoal binding when submitting daily goal', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ id: 42 })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Set goal'))
    await user.click(screen.getByText(/daily/i))
    const input = screen.getByPlaceholderText(/daily goal/i)
    await user.type(input, '3{Enter}')

    await waitFor(() => {
      expect(SetHabitGoal).toHaveBeenCalledWith(42, 3)
    })
  })

  it('calls onHabitChanged after setting goal', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit()]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Set goal'))
    await user.click(screen.getByText(/daily/i))
    const input = screen.getByPlaceholderText(/daily goal/i)
    await user.type(input, '3{Enter}')

    await waitFor(() => {
      expect(onHabitChanged).toHaveBeenCalled()
    })
  })
})

describe('HabitTracker - Horizontal Scroll Layout', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-07T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders day circles in scrollable container', () => {
    // Jan 7, 2024 is a Sunday, so week is Jan 7-13
    const anchor = new Date('2024-01-07')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2024-01-07', completed: false, count: 0 },
        { date: '2024-01-08', completed: false, count: 0 },
        { date: '2024-01-09', completed: false, count: 0 },
        { date: '2024-01-10', completed: false, count: 0 },
        { date: '2024-01-11', completed: false, count: 0 },
        { date: '2024-01-12', completed: false, count: 0 },
        { date: '2024-01-13', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // Calendar grid is rendered inside an overflow-x-auto container
    const dayCircles = screen.getAllByRole('button', { name: /Log for 2024-01/i })
    expect(dayCircles.length).toBe(7)
    // The parent container should have overflow-x-auto for scrolling
    const scrollContainer = dayCircles[0].closest('.overflow-x-auto')
    expect(scrollContainer).toBeInTheDocument()
  })
})

describe('HabitTracker - Today Indicator', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-07T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows visible ring around today circle', () => {
    const anchor = new Date('2024-01-07')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2024-01-07', completed: false, count: 0 },
        { date: '2024-01-06', completed: false, count: 0 },
        { date: '2024-01-05', completed: false, count: 0 },
        { date: '2024-01-04', completed: false, count: 0 },
        { date: '2024-01-03', completed: false, count: 0 },
        { date: '2024-01-02', completed: false, count: 0 },
        { date: '2024-01-01', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // Find today's cell (Jan 7, 2024) - it should have ring-bujo-today class
    const todayButton = screen.getByLabelText(/Log for 2024-01-07/i)
    expect(todayButton).toHaveClass('ring-bujo-today')
  })
})

describe('HabitTracker - Date Range Indicator', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-07T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('displays date range between period selector and add button', () => {
    // Anchor to Jan 3, 2024 (Wednesday) - week will be Dec 31, 2023 - Jan 6, 2024
    const anchor = new Date('2024-01-03')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2024-01-07', completed: false, count: 0 },
        { date: '2024-01-06', completed: false, count: 0 },
        { date: '2024-01-05', completed: false, count: 0 },
        { date: '2024-01-04', completed: false, count: 0 },
        { date: '2024-01-03', completed: false, count: 0 },
        { date: '2024-01-02', completed: false, count: 0 },
        { date: '2024-01-01', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    // CalendarNavigation shows past 7 days ending with anchor date
    // Jan 3, 2024 anchor means Dec 28 - Jan 3
    expect(screen.getByText(/Dec 28.*Jan 3.*2024/i)).toBeInTheDocument()
  })
})

describe('HabitTracker - Day Order Display', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-07T12:00:00'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('displays days in past-to-present order (oldest on left, anchor on right)', () => {
    const anchor = new Date('2024-01-07')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2024-01-07', completed: false, count: 0 }, // Sunday (anchor)
        { date: '2024-01-06', completed: false, count: 0 }, // Saturday
        { date: '2024-01-05', completed: false, count: 0 }, // Friday
        { date: '2024-01-04', completed: false, count: 0 }, // Thursday
        { date: '2024-01-03', completed: false, count: 0 }, // Wednesday
        { date: '2024-01-02', completed: false, count: 0 }, // Tuesday
        { date: '2024-01-01', completed: false, count: 0 }, // Monday
      ]
    })
    render(<HabitTracker habits={[habit]} anchorDate={anchor} />)

    const dayCircles = screen.getAllByRole('button', { name: /Log for 2024-01/i })
    // Past 7 days ending with anchor: Jan 1-7
    // First circle should be oldest (Jan 1), last should be anchor (Jan 7)
    expect(dayCircles[0]).toHaveAttribute('aria-label', expect.stringContaining('2024-01-01'))
    expect(dayCircles[6]).toHaveAttribute('aria-label', expect.stringContaining('2024-01-07'))
  })
})

describe('HabitTracker - Weekly Goal Display', () => {
  it('shows weekly goal indicator when goalPerWeek is set', () => {
    const habit = createTestHabit({ goalPerWeek: 5 })
    render(<HabitTracker habits={[habit]} />)

    const goalIndicator = screen.getByLabelText(/weekly goal.*5/i)
    expect(goalIndicator).toBeInTheDocument()
  })

  it('shows weekly progress when available', () => {
    const habit = createTestHabit({ goalPerWeek: 5, weeklyProgress: 60 })
    render(<HabitTracker habits={[habit]} />)

    expect(screen.getByText(/60%/)).toBeInTheDocument()
  })

  it('does not show weekly goal indicator when not set', () => {
    const habit = createTestHabit({ goalPerWeek: undefined })
    render(<HabitTracker habits={[habit]} />)

    expect(screen.queryByLabelText(/weekly goal/i)).not.toBeInTheDocument()
  })
})

describe('HabitTracker - Monthly Goal Display', () => {
  it('shows monthly goal indicator when goalPerMonth is set', () => {
    const habit = createTestHabit({ goalPerMonth: 20 })
    render(<HabitTracker habits={[habit]} />)

    const goalIndicator = screen.getByLabelText(/monthly goal.*20/i)
    expect(goalIndicator).toBeInTheDocument()
  })

  it('does not show monthly goal indicator when not set', () => {
    const habit = createTestHabit({ goalPerMonth: undefined })
    render(<HabitTracker habits={[habit]} />)

    expect(screen.queryByLabelText(/monthly goal/i)).not.toBeInTheDocument()
  })
})

describe('HabitTracker - Goal Type Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows goal type dropdown when goal button is clicked', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit()]} />)

    await user.click(screen.getByTitle('Set goal'))

    expect(screen.getByText(/daily/i)).toBeInTheDocument()
    expect(screen.getByText(/weekly/i)).toBeInTheDocument()
    expect(screen.getByText(/monthly/i)).toBeInTheDocument()
  })

  it('calls SetHabitWeeklyGoal when submitting weekly goal', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ id: 42 })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Set goal'))
    await user.click(screen.getByText(/weekly/i))
    const input = screen.getByPlaceholderText(/weekly goal/i)
    await user.type(input, '5{Enter}')

    await waitFor(() => {
      expect(SetHabitWeeklyGoal).toHaveBeenCalledWith(42, 5)
    })
  })

  it('calls SetHabitMonthlyGoal when submitting monthly goal', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ id: 42 })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Set goal'))
    await user.click(screen.getByText(/monthly/i))
    const input = screen.getByPlaceholderText(/monthly goal/i)
    await user.type(input, '20{Enter}')

    await waitFor(() => {
      expect(SetHabitMonthlyGoal).toHaveBeenCalledWith(42, 20)
    })
  })

  it('defaults to daily goal type', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit()]} />)

    await user.click(screen.getByTitle('Set goal'))
    await user.click(screen.getByText(/daily/i))

    expect(screen.getByPlaceholderText(/daily goal/i)).toBeInTheDocument()
  })
})
