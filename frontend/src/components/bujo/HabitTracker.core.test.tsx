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

import { CreateHabit, DeleteHabit, UndoHabitLogForDate } from '@/wailsjs/go/wails/App'

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
