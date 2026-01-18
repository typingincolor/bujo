import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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

import { CreateHabit, DeleteHabit, UndoHabitLog, UndoHabitLogForDate, SetHabitGoal, LogHabitForDate } from '@/wailsjs/go/wails/App'

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

describe('HabitTracker - Create Habit', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows add habit button', () => {
    render(<HabitTracker habits={[]} />)
    expect(screen.getByRole('button', { name: /add habit/i })).toBeInTheDocument()
  })

  it('shows inline input when add habit button is clicked', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[]} />)

    await user.click(screen.getByRole('button', { name: /add habit/i }))

    expect(screen.getByPlaceholderText(/habit name/i)).toBeInTheDocument()
  })

  it('calls CreateHabit binding when submitting new habit', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByRole('button', { name: /add habit/i }))

    const input = screen.getByPlaceholderText(/habit name/i)
    await user.type(input, 'Morning Run{Enter}')

    await waitFor(() => {
      expect(CreateHabit).toHaveBeenCalledWith('Morning Run')
    })
  })

  it('calls onHabitChanged after creating habit', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByRole('button', { name: /add habit/i }))

    const input = screen.getByPlaceholderText(/habit name/i)
    await user.type(input, 'Morning Run{Enter}')

    await waitFor(() => {
      expect(onHabitChanged).toHaveBeenCalled()
    })
  })

  it('clears input after creating habit', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[]} />)

    await user.click(screen.getByRole('button', { name: /add habit/i }))

    const input = screen.getByPlaceholderText(/habit name/i) as HTMLInputElement
    await user.type(input, 'Morning Run{Enter}')

    await waitFor(() => {
      expect(input.value).toBe('')
    })
  })

  it('hides input when cancel button is clicked', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[]} />)

    await user.click(screen.getByRole('button', { name: /add habit/i }))
    expect(screen.getByPlaceholderText(/habit name/i)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /cancel/i }))
    expect(screen.queryByPlaceholderText(/habit name/i)).not.toBeInTheDocument()
  })

  it('hides input when Escape is pressed', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[]} />)

    await user.click(screen.getByRole('button', { name: /add habit/i }))
    const input = screen.getByPlaceholderText(/habit name/i)

    await user.type(input, '{Escape}')
    expect(screen.queryByPlaceholderText(/habit name/i)).not.toBeInTheDocument()
  })

  it('does not create habit with empty name', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[]} />)

    await user.click(screen.getByRole('button', { name: /add habit/i }))

    const input = screen.getByPlaceholderText(/habit name/i)
    await user.type(input, '{Enter}')

    expect(CreateHabit).not.toHaveBeenCalled()
  })
})

describe('HabitTracker - Display Habits', () => {
  it('renders habits', () => {
    render(<HabitTracker habits={[createTestHabit({ name: 'Exercise' })]} />)
    expect(screen.getByText('Exercise')).toBeInTheDocument()
  })

  it('shows empty state when no habits', () => {
    render(<HabitTracker habits={[]} />)
    expect(screen.getByText(/habit tracker/i)).toBeInTheDocument()
  })
})

describe('HabitTracker - Delete Habit', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows delete button on habit row', () => {
    render(<HabitTracker habits={[createTestHabit({ name: 'Exercise' })]} />)
    expect(screen.getByTitle('Delete habit')).toBeInTheDocument()
  })

  it('shows confirmation dialog when delete button is clicked', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit({ name: 'Exercise' })]} />)

    await user.click(screen.getByTitle('Delete habit'))

    expect(screen.getByText('Delete Habit')).toBeInTheDocument()
    expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
  })

  it('calls DeleteHabit binding when confirming delete', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ id: 42, name: 'Exercise' })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Delete habit'))

    const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
    const confirmButton = deleteButtons.find(btn => btn.textContent === 'Delete')
    expect(confirmButton).toBeDefined()
    await user.click(confirmButton!)

    await waitFor(() => {
      expect(DeleteHabit).toHaveBeenCalledWith(42)
    })
  })

  it('calls onHabitChanged after deleting habit', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ name: 'Exercise' })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Delete habit'))

    const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
    const confirmButton = deleteButtons.find(btn => btn.textContent === 'Delete')
    await user.click(confirmButton!)

    await waitFor(() => {
      expect(onHabitChanged).toHaveBeenCalled()
    })
  })

  it('closes dialog on cancel', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit({ name: 'Exercise' })]} />)

    await user.click(screen.getByTitle('Delete habit'))
    expect(screen.getByText('Delete Habit')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /cancel/i }))

    expect(screen.queryByText('Delete Habit')).not.toBeInTheDocument()
  })

  it('does not call DeleteHabit when cancel is clicked', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit({ name: 'Exercise' })]} />)

    await user.click(screen.getByTitle('Delete habit'))
    await user.click(screen.getByRole('button', { name: /cancel/i }))

    expect(DeleteHabit).not.toHaveBeenCalled()
  })
})

describe('HabitTracker - Period View Toggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows period selector with Week as default', () => {
    render(<HabitTracker habits={[createTestHabit()]} />)
    expect(screen.getByRole('button', { name: /week/i })).toBeInTheDocument()
  })

  it('shows Month option in period selector', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit()]} />)

    await user.click(screen.getByRole('button', { name: /week/i }))
    expect(screen.getByRole('button', { name: /month/i })).toBeInTheDocument()
  })

  it('shows Quarter option in period selector', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit()]} />)

    await user.click(screen.getByRole('button', { name: /week/i }))
    expect(screen.getByRole('button', { name: /quarter/i })).toBeInTheDocument()
  })

  it('calls onPeriodChange when selecting Month', async () => {
    const user = userEvent.setup()
    const onPeriodChange = vi.fn()
    render(<HabitTracker habits={[createTestHabit()]} onPeriodChange={onPeriodChange} />)

    await user.click(screen.getByRole('button', { name: /week/i }))
    await user.click(screen.getByRole('button', { name: /month/i }))

    expect(onPeriodChange).toHaveBeenCalledWith('month')
  })
})

describe('HabitTracker - Decrement via Cmd+Click', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls UndoHabitLogForDate when Cmd+clicking on a logged day circle', async () => {
    const onHabitChanged = vi.fn()
    const anchor = new Date('2024-01-01')
    const habit = createTestHabit({
      id: 42,
      dayHistory: [
        { date: '2024-01-01', completed: true, count: 2 },
        { date: '2024-01-02', completed: false, count: 0 },
        { date: '2024-01-03', completed: false, count: 0 },
        { date: '2024-01-04', completed: false, count: 0 },
        { date: '2024-01-05', completed: false, count: 0 },
        { date: '2024-01-06', completed: false, count: 0 },
        { date: '2024-01-07', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} onHabitChanged={onHabitChanged} anchorDate={anchor} />)

    // Find the button with the logged date
    const loggedButton = screen.getByLabelText(/Logged for 2024-01-01/i)
    fireEvent.click(loggedButton, { metaKey: true })

    await waitFor(() => {
      expect(UndoHabitLogForDate).toHaveBeenCalledWith(42, expect.any(String))
      // Verify the date string is for 2024-01-01
      const call = vi.mocked(UndoHabitLogForDate).mock.calls[0]
      expect(call[1]).toContain('2024-01-01')
    })
  })

  it('does not decrement when Cmd+clicking on a day with count 0', () => {
    const onHabitChanged = vi.fn()
    const anchor = new Date('2024-01-01')
    const habit = createTestHabit({
      id: 42,
      dayHistory: [
        { date: '2024-01-01', completed: false, count: 0 },
        { date: '2024-01-02', completed: false, count: 0 },
        { date: '2024-01-03', completed: false, count: 0 },
        { date: '2024-01-04', completed: false, count: 0 },
        { date: '2024-01-05', completed: false, count: 0 },
        { date: '2024-01-06', completed: false, count: 0 },
        { date: '2024-01-07', completed: false, count: 0 },
      ]
    })
    render(<HabitTracker habits={[habit]} onHabitChanged={onHabitChanged} anchorDate={anchor} />)

    // Find a button with no logs
    const dayButton = screen.getByLabelText(/Log for 2024-01-01$/i)
    fireEvent.click(dayButton, { metaKey: true })

    expect(UndoHabitLogForDate).not.toHaveBeenCalled()
  })
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

  it('shows goal input when goal button is clicked', async () => {
    const user = userEvent.setup()
    render(<HabitTracker habits={[createTestHabit()]} />)

    await user.click(screen.getByTitle('Set goal'))

    expect(screen.getByPlaceholderText(/daily goal/i)).toBeInTheDocument()
  })

  it('calls SetHabitGoal binding when submitting goal', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ id: 42 })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Set goal'))
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

    // CalendarNavigation shows the week range containing the anchor date
    // Jan 3, 2024 is Wednesday, so week is Dec 31 - Jan 6
    expect(screen.getByText(/Dec 31.*Jan 6.*2024/i)).toBeInTheDocument()
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

  it('displays days in calendar order (Sunday to Saturday)', () => {
    const anchor = new Date('2024-01-07')
    const habit = createTestHabit({
      dayHistory: [
        { date: '2024-01-07', completed: false, count: 0 }, // Sunday
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
    // Calendar grid orders days Sunday to Saturday
    // Jan 7 2024 is a Sunday, so week of Jan 7 is Jan 7-13
    // First circle should be Sunday (Jan 7)
    expect(dayCircles[0]).toHaveAttribute('aria-label', expect.stringContaining('2024-01-07'))
  })
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
    const anchor = new Date('2024-01-01')
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

    // Should show navigation with week label
    expect(screen.getByLabelText('Previous')).toBeInTheDocument()
    expect(screen.getByLabelText('Next')).toBeInTheDocument()
    expect(screen.getByText(/Jan 14.*Jan 20.*2024/)).toBeInTheDocument()
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
    const anchor = new Date('2024-01-15')
    render(<HabitTracker habits={[createTestHabit()]} anchorDate={anchor} onNavigate={onNavigate} />)

    fireEvent.click(screen.getByLabelText('Next'))

    expect(onNavigate).toHaveBeenCalled()
    // The new anchor should be 7 days after (week view default)
    const newAnchor = onNavigate.mock.calls[0][0] as Date
    expect(newAnchor.getDate()).toBe(22) // Jan 15 + 7 = Jan 22
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
})
