import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HabitTracker } from './HabitTracker'
import { Habit } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  LogHabit: vi.fn().mockResolvedValue(undefined),
  CreateHabit: vi.fn().mockResolvedValue(1),
  DeleteHabit: vi.fn().mockResolvedValue(undefined),
  UndoHabitLog: vi.fn().mockResolvedValue(undefined),
  SetHabitGoal: vi.fn().mockResolvedValue(undefined),
  LogHabitForDate: vi.fn().mockResolvedValue(undefined),
}))

import { CreateHabit, DeleteHabit, UndoHabitLog, SetHabitGoal, LogHabitForDate } from '@/wailsjs/go/wails/App'

const createTestHabit = (overrides: Partial<Habit> = {}): Habit => ({
  id: 1,
  name: 'Test Habit',
  goal: 1,
  streak: 0,
  completionRate: 0,
  todayLogged: false,
  todayCount: 0,
  history: [false, false, false, false, false, false, false],
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

describe('HabitTracker - Undo Habit Log', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows undo button on habit row when habit is logged today', () => {
    render(<HabitTracker habits={[createTestHabit({ todayLogged: true, todayCount: 1 })]} />)
    expect(screen.getByTitle('Undo last log')).toBeInTheDocument()
  })

  it('does not show undo button when habit is not logged today', () => {
    render(<HabitTracker habits={[createTestHabit({ todayLogged: false })]} />)
    expect(screen.queryByTitle('Undo last log')).not.toBeInTheDocument()
  })

  it('calls UndoHabitLog binding when undo button is clicked', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ id: 42, todayLogged: true, todayCount: 1 })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Undo last log'))

    await waitFor(() => {
      expect(UndoHabitLog).toHaveBeenCalledWith(42)
    })
  })

  it('calls onHabitChanged after undo', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    render(<HabitTracker habits={[createTestHabit({ todayLogged: true, todayCount: 1 })]} onHabitChanged={onHabitChanged} />)

    await user.click(screen.getByTitle('Undo last log'))

    await waitFor(() => {
      expect(onHabitChanged).toHaveBeenCalled()
    })
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

describe('HabitTracker - Log Habit For Specific Date', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows date picker when clicking on a past day in sparkline', async () => {
    const user = userEvent.setup()
    // Create habit with history array of 7 days
    const habit = createTestHabit({
      history: [false, false, false, false, false, false, false]
    })
    render(<HabitTracker habits={[habit]} />)

    // Click on a past day (first day in sparkline which is 6 days ago)
    const sparklineDots = screen.getAllByRole('button', { name: /log for/i })
    expect(sparklineDots.length).toBeGreaterThan(0)
    await user.click(sparklineDots[0])

    // Check for the dialog title
    expect(screen.getByText('Log Habit for Date')).toBeInTheDocument()
  })

  it('calls LogHabitForDate binding when logging for specific date', async () => {
    const user = userEvent.setup()
    const onHabitChanged = vi.fn()
    const habit = createTestHabit({
      id: 42,
      history: [false, false, false, false, false, false, false]
    })
    render(<HabitTracker habits={[habit]} onHabitChanged={onHabitChanged} />)

    // Click on a past day sparkline dot
    const sparklineDots = screen.getAllByRole('button', { name: /log for/i })
    await user.click(sparklineDots[0])

    // Confirm logging
    await user.click(screen.getByRole('button', { name: /confirm/i }))

    await waitFor(() => {
      expect(LogHabitForDate).toHaveBeenCalled()
      const call = vi.mocked(LogHabitForDate).mock.calls[0]
      expect(call[0]).toBe(42) // habit ID
      expect(call[1]).toBe(1) // count
    })
  })
})
