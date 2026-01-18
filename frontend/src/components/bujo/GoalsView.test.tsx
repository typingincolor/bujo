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
  DeleteGoal: vi.fn().mockResolvedValue(undefined),
  MigrateGoal: vi.fn().mockResolvedValue(2),
  UpdateGoal: vi.fn().mockResolvedValue(undefined),
  CancelGoal: vi.fn().mockResolvedValue(undefined),
  UncancelGoal: vi.fn().mockResolvedValue(undefined),
}))

import { CreateGoal, DeleteGoal, MigrateGoal, UpdateGoal, CancelGoal, UncancelGoal, MarkGoalDone, MarkGoalActive } from '@/wailsjs/go/wails/App'

const currentMonth = format(new Date(), 'yyyy-MM')

const createTestGoal = (overrides: Partial<Goal> = {}): Goal => ({
  id: 1,
  content: 'Test goal',
  month: currentMonth,
  status: 'active',
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
      createTestGoal({ id: 1, status: 'done' }),
      createTestGoal({ id: 2, status: 'active' }),
    ]} />)

    expect(screen.getByText('1/2')).toBeInTheDocument()
  })
})

describe('GoalsView - Delete Goal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows delete button on goal item hover', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)
    expect(screen.getByTitle('Delete goal')).toBeInTheDocument()
  })

  it('shows confirmation dialog when delete button is clicked', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)

    await user.click(screen.getByTitle('Delete goal'))

    expect(screen.getByText('Delete Goal')).toBeInTheDocument()
    expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
  })

  it('calls DeleteGoal binding when confirming delete', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ id: 42, content: 'My Goal' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Delete goal'))

    const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
    const confirmButton = deleteButtons.find(btn => btn.textContent === 'Delete')
    expect(confirmButton).toBeDefined()
    await user.click(confirmButton!)

    await waitFor(() => {
      expect(DeleteGoal).toHaveBeenCalledWith(42)
    })
  })

  it('calls onGoalChanged after deleting goal', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Delete goal'))

    const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
    const confirmButton = deleteButtons.find(btn => btn.textContent === 'Delete')
    expect(confirmButton).toBeDefined()
    await user.click(confirmButton!)

    await waitFor(() => {
      expect(onGoalChanged).toHaveBeenCalled()
    })
  })

  it('closes dialog on cancel', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)

    await user.click(screen.getByTitle('Delete goal'))
    expect(screen.getByText('Delete Goal')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /^Cancel$/ }))

    expect(screen.queryByText('Delete Goal')).not.toBeInTheDocument()
  })

  it('does not call DeleteGoal when cancel is clicked', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)

    await user.click(screen.getByTitle('Delete goal'))
    await user.click(screen.getByRole('button', { name: /^Cancel$/ }))

    expect(DeleteGoal).not.toHaveBeenCalled()
  })
})

describe('GoalsView - Migrate Goal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows migrate button on goal item', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)
    expect(screen.getByTitle('Migrate goal')).toBeInTheDocument()
  })

  it('shows month picker dialog when migrate button is clicked', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)

    await user.click(screen.getByTitle('Migrate goal'))

    expect(screen.getByText(/migrate goal/i)).toBeInTheDocument()
  })

  it('calls MigrateGoal binding when confirming migration', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ id: 42, content: 'My Goal' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Migrate goal'))

    const confirmButton = screen.getByRole('button', { name: /^migrate$/i })
    await user.click(confirmButton)

    await waitFor(() => {
      expect(MigrateGoal).toHaveBeenCalledWith(42, expect.any(String))
    })
  })

  it('shows migrated indicator for migrated goals', () => {
    render(<GoalsView goals={[createTestGoal({
      content: 'Migrated Goal',
      status: 'migrated',
      migratedTo: '2026-02'
    })]} />)

    expect(screen.getByText(/migrated to/i)).toBeInTheDocument()
  })

  it('does not show migrate button for migrated goals', () => {
    render(<GoalsView goals={[createTestGoal({
      content: 'Migrated Goal',
      status: 'migrated',
      migratedTo: '2026-02'
    })]} />)

    expect(screen.queryByTitle('Migrate goal')).not.toBeInTheDocument()
  })
})

describe('GoalsView - Edit Goal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows edit button on goal item', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)
    expect(screen.getByTitle('Edit goal')).toBeInTheDocument()
  })

  it('shows edit input when edit button is clicked', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)

    await user.click(screen.getByTitle('Edit goal'))

    expect(screen.getByDisplayValue('My Goal')).toBeInTheDocument()
  })

  it('calls UpdateGoal binding when saving edit', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ id: 42, content: 'My Goal' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Edit goal'))
    const input = screen.getByDisplayValue('My Goal')
    await user.clear(input)
    await user.type(input, 'Updated Goal{Enter}')

    await waitFor(() => {
      expect(UpdateGoal).toHaveBeenCalledWith(42, 'Updated Goal')
    })
  })

  it('calls onGoalChanged after editing goal', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Edit goal'))
    const input = screen.getByDisplayValue('My Goal')
    await user.clear(input)
    await user.type(input, 'Updated Goal{Enter}')

    await waitFor(() => {
      expect(onGoalChanged).toHaveBeenCalled()
    })
  })

  it('cancels edit when Escape is pressed', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)

    await user.click(screen.getByTitle('Edit goal'))
    const input = screen.getByDisplayValue('My Goal')
    await user.type(input, '{Escape}')

    expect(screen.queryByDisplayValue('My Goal')).not.toBeInTheDocument()
    expect(screen.getByText('My Goal')).toBeInTheDocument()
  })

  it('does not show edit button for migrated goals', () => {
    render(<GoalsView goals={[createTestGoal({
      content: 'Migrated Goal',
      status: 'migrated',
      migratedTo: '2026-02'
    })]} />)

    expect(screen.queryByTitle('Edit goal')).not.toBeInTheDocument()
  })
})

describe('GoalsView - Click and Tick/Untick Behavior', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('clicking on a goal does not call toggle handler', async () => {
    const user = userEvent.setup()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal', status: 'active' })]} />)

    await user.click(screen.getByText('My Goal'))

    expect(MarkGoalDone).not.toHaveBeenCalled()
    expect(MarkGoalActive).not.toHaveBeenCalled()
  })

  it('shows tick button on active goals to mark as done', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'Active Goal', status: 'active' })]} />)

    expect(screen.getByTitle('Mark as done')).toBeInTheDocument()
  })

  it('shows untick button on done goals to mark as not done', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'Done Goal', status: 'done' })]} />)

    expect(screen.getByTitle('Mark as not done')).toBeInTheDocument()
  })

  it('calls MarkGoalDone when tick button is clicked', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ id: 42, content: 'Active Goal', status: 'active' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Mark as done'))

    await waitFor(() => {
      expect(MarkGoalDone).toHaveBeenCalledWith(42)
    })
  })

  it('calls MarkGoalActive when untick button is clicked', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ id: 42, content: 'Done Goal', status: 'done' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Mark as not done'))

    await waitFor(() => {
      expect(MarkGoalActive).toHaveBeenCalledWith(42)
    })
  })

  it('does not show tick/untick button on cancelled goals', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'Cancelled Goal', status: 'cancelled' })]} />)

    expect(screen.queryByTitle('Mark as done')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Mark as not done')).not.toBeInTheDocument()
  })

  it('does not show tick/untick button on migrated goals', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'Migrated Goal', status: 'migrated', migratedTo: '2026-02' })]} />)

    expect(screen.queryByTitle('Mark as done')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Mark as not done')).not.toBeInTheDocument()
  })

  it('shows task bullet symbol in mark as not done button', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'Done Goal', status: 'done' })]} />)
    const undoneButton = screen.getByTitle('Mark as not done')
    expect(undoneButton).toHaveTextContent('•')
  })
})

describe('GoalsView - Cancel Goal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows cancel button on active goal item', () => {
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} />)
    expect(screen.getByTitle('Cancel goal')).toBeInTheDocument()
  })

  it('calls CancelGoal binding when cancel button is clicked', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ id: 42, content: 'My Goal' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Cancel goal'))

    await waitFor(() => {
      expect(CancelGoal).toHaveBeenCalledWith(42)
    })
  })

  it('calls onGoalChanged after cancelling goal', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({ content: 'My Goal' })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Cancel goal'))

    await waitFor(() => {
      expect(onGoalChanged).toHaveBeenCalled()
    })
  })

  it('shows cancelled indicator for cancelled goals', () => {
    render(<GoalsView goals={[createTestGoal({
      content: 'Cancelled Goal',
      status: 'cancelled'
    })]} />)

    expect(screen.getByText('Cancelled Goal')).toHaveClass('line-through')
  })

  it('shows uncancel button for cancelled goals', () => {
    render(<GoalsView goals={[createTestGoal({
      content: 'Cancelled Goal',
      status: 'cancelled'
    })]} />)

    expect(screen.getByTitle('Restore goal')).toBeInTheDocument()
  })

  it('calls UncancelGoal binding when restore button is clicked', async () => {
    const user = userEvent.setup()
    const onGoalChanged = vi.fn()
    render(<GoalsView goals={[createTestGoal({
      id: 42,
      content: 'Cancelled Goal',
      status: 'cancelled'
    })]} onGoalChanged={onGoalChanged} />)

    await user.click(screen.getByTitle('Restore goal'))

    await waitFor(() => {
      expect(UncancelGoal).toHaveBeenCalledWith(42)
    })
  })

  it('does not show cancel button for migrated goals', () => {
    render(<GoalsView goals={[createTestGoal({
      content: 'Migrated Goal',
      status: 'migrated',
      migratedTo: '2026-02'
    })]} />)

    expect(screen.queryByTitle('Cancel goal')).not.toBeInTheDocument()
  })

  it('does not show cancel button for cancelled goals', () => {
    render(<GoalsView goals={[createTestGoal({
      content: 'Cancelled Goal',
      status: 'cancelled'
    })]} />)

    expect(screen.queryByTitle('Cancel goal')).not.toBeInTheDocument()
  })
})
