import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HabitTracker } from './HabitTracker'
import { Habit } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  LogHabit: vi.fn().mockResolvedValue(undefined),
  CreateHabit: vi.fn().mockResolvedValue(1),
}))

import { CreateHabit } from '@/wailsjs/go/wails/App'

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
