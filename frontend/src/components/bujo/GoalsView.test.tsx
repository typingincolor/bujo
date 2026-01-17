import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { GoalsView } from './GoalsView'
import { Goal } from '@/types/bujo'
import { format } from 'date-fns'

vi.mock('@/wailsjs/go/wails/App', () => ({
  MarkGoalDone: vi.fn().mockResolvedValue(undefined),
  MarkGoalActive: vi.fn().mockResolvedValue(undefined),
  CreateGoal: vi.fn().mockResolvedValue(1),
}))

import { CreateGoal } from '@/wailsjs/go/wails/App'

const currentMonth = format(new Date(), 'yyyy-MM')

const createTestGoal = (overrides: Partial<Goal> = {}): Goal => ({
  id: 1,
  content: 'Test goal',
  month: currentMonth,
  completed: false,
  ...overrides,
})

describe('GoalsView - Create Goal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows add goal button', () => {
    render(<GoalsView goals={[]} />)
    expect(screen.getByRole('button', { name: /add goal/i })).toBeInTheDocument()
  })

  it('shows inline input when add goal button is clicked', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[]} />)

    await user.click(screen.getByRole('button', { name: /add goal/i }))

    expect(screen.getByPlaceholderText(/new goal/i)).toBeInTheDocument()
  })

  it('calls CreateGoal binding when submitting new goal', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByRole('button', { name: /add goal/i }))

    const input = screen.getByPlaceholderText(/new goal/i)
    await user.type(input, 'Learn TypeScript{Enter}')

    await waitFor(() => {
      expect(CreateGoal).toHaveBeenCalledWith('Learn TypeScript', expect.any(String))
    })
  })

  it('calls onGoalChanged after creating goal', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByRole('button', { name: /add goal/i }))

    const input = screen.getByPlaceholderText(/new goal/i)
    await user.type(input, 'Learn TypeScript{Enter}')

    await waitFor(() => {
      expect(onGoalChanged).toHaveBeenCalled()
    })
  })

  it('clears input after creating goal', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[]} />)

    await user.click(screen.getByRole('button', { name: /add goal/i }))

    const input = screen.getByPlaceholderText(/new goal/i) as HTMLInputElement
    await user.type(input, 'Learn TypeScript{Enter}')

    await waitFor(() => {
      expect(input.value).toBe('')
    })
  })

  it('hides input when cancel button is clicked', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[]} />)

    await user.click(screen.getByRole('button', { name: /add goal/i }))
    expect(screen.getByPlaceholderText(/new goal/i)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /cancel/i }))
    expect(screen.queryByPlaceholderText(/new goal/i)).not.toBeInTheDocument()
  })

  it('hides input when Escape is pressed', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[]} />)

    await user.click(screen.getByRole('button', { name: /add goal/i }))
    const input = screen.getByPlaceholderText(/new goal/i)

    await user.type(input, '{Escape}')
    expect(screen.queryByPlaceholderText(/new goal/i)).not.toBeInTheDocument()
  })

  it('does not create goal with empty content', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[]} />)

    await user.click(screen.getByRole('button', { name: /add goal/i }))

    const input = screen.getByPlaceholderText(/new goal/i)
    await user.type(input, '{Enter}')

    expect(CreateGoal).not.toHaveBeenCalled()
  })
})

describe('GoalsView - Toggle Goals', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders goals for current month', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)
    expect(screen.getByText('My Goal')).toBeInTheDocument()
  })

  it('shows progress bar with correct progress', () => {
    render(<GoalsView goals={[
      createTestGoal({ id: 1, completed: true }),
      createTestGoal({ id: 2, completed: false }),
    ]} />)

    expect(screen.getByText('1/2')).toBeInTheDocument()
  })
})
