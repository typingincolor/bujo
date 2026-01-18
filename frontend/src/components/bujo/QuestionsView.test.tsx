import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QuestionsView } from './QuestionsView'
import { Entry } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  AnswerQuestion: vi.fn().mockResolvedValue(undefined),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  RetypeEntry: vi.fn().mockResolvedValue(undefined),
}))

import { AnswerQuestion, CancelEntry, CyclePriority, RetypeEntry } from '@/wailsjs/go/wails/App'

const createTestEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  content: 'Test question',
  type: 'question',
  priority: 'none',
  parentId: null,
  loggedDate: '2026-01-15',
  ...overrides,
})

describe('QuestionsView - Display', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the questions header with count', () => {
    render(<QuestionsView questions={[createTestEntry()]} />)
    expect(screen.getByText(/outstanding questions/i)).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('uses HelpCircle icon in header', () => {
    render(<QuestionsView questions={[createTestEntry()]} />)
    expect(screen.getByTestId('questions-icon')).toBeInTheDocument()
  })

  it('renders multiple question entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Question one' }),
      createTestEntry({ id: 2, content: 'Question two' }),
      createTestEntry({ id: 3, content: 'Question three' }),
    ]
    render(<QuestionsView questions={entries} />)

    expect(screen.getByText('Question one')).toBeInTheDocument()
    expect(screen.getByText('Question two')).toBeInTheDocument()
    expect(screen.getByText('Question three')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('shows empty state when no questions', () => {
    render(<QuestionsView questions={[]} />)
    expect(screen.getByText(/no outstanding questions/i)).toBeInTheDocument()
  })

  it('displays entry date', () => {
    render(<QuestionsView questions={[createTestEntry({ loggedDate: '2026-01-10' })]} />)
    expect(screen.getByText(/jan 10/i)).toBeInTheDocument()
  })

  it('shows question symbol', () => {
    render(<QuestionsView questions={[createTestEntry({ type: 'question' })]} />)
    expect(screen.getByTestId('entry-symbol')).toBeInTheDocument()
    expect(screen.getByTestId('entry-symbol')).toHaveTextContent('?')
  })

  it('shows priority indicator for high priority questions', () => {
    render(<QuestionsView questions={[createTestEntry({ priority: 'high' })]} />)
    expect(screen.getByText('!!!')).toBeInTheDocument()
  })

  it('filters to only show question entries', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'A question', type: 'question' }),
      createTestEntry({ id: 2, content: 'A task', type: 'task' }),
      createTestEntry({ id: 3, content: 'A note', type: 'note' }),
    ]
    render(<QuestionsView questions={entries} />)

    expect(screen.getByText('A question')).toBeInTheDocument()
    expect(screen.queryByText('A task')).not.toBeInTheDocument()
    expect(screen.queryByText('A note')).not.toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
  })
})

describe('QuestionsView - Answer Button', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows answer button for question entries', () => {
    render(<QuestionsView questions={[createTestEntry()]} />)
    expect(screen.getByTitle('Answer question')).toBeInTheDocument()
  })

  it('opens answer modal when clicking answer button', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ content: 'What is TDD?' })]} />)

    await user.click(screen.getByTitle('Answer question'))

    expect(screen.getByText('Answer Question')).toBeInTheDocument()
    // Text appears twice: once in list, once in modal
    expect(screen.getAllByText('What is TDD?').length).toBe(2)
  })

  it('calls AnswerQuestion when submitting answer', async () => {
    const user = userEvent.setup()
    const onEntryChanged = vi.fn()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} onEntryChanged={onEntryChanged} />)

    await user.click(screen.getByTitle('Answer question'))
    await user.type(screen.getByPlaceholderText(/enter your answer/i), 'Test Driven Development')
    await user.click(screen.getByRole('button', { name: /submit/i }))

    await waitFor(() => {
      expect(AnswerQuestion).toHaveBeenCalledWith(42, 'Test Driven Development')
    })
  })

  it('calls onEntryChanged after answering', async () => {
    const user = userEvent.setup()
    const onEntryChanged = vi.fn()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} onEntryChanged={onEntryChanged} />)

    await user.click(screen.getByTitle('Answer question'))
    await user.type(screen.getByPlaceholderText(/enter your answer/i), 'An answer')
    await user.click(screen.getByRole('button', { name: /submit/i }))

    await waitFor(() => {
      expect(onEntryChanged).toHaveBeenCalled()
    })
  })
})

describe('QuestionsView - Other Actions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows cancel button for questions', () => {
    render(<QuestionsView questions={[createTestEntry()]} />)
    expect(screen.getByTitle('Cancel entry')).toBeInTheDocument()
  })

  it('calls CancelEntry when cancel button is clicked', async () => {
    const user = userEvent.setup()
    const onEntryChanged = vi.fn()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} onEntryChanged={onEntryChanged} />)

    await user.click(screen.getByTitle('Cancel entry'))

    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('shows priority button', () => {
    render(<QuestionsView questions={[createTestEntry()]} />)
    expect(screen.getByTitle('Cycle priority')).toBeInTheDocument()
  })

  it('cycles priority when priority button is clicked', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} />)

    await user.click(screen.getByTitle('Cycle priority'))

    await waitFor(() => {
      expect(CyclePriority).toHaveBeenCalledWith(42)
    })
  })

  it('shows change type button', () => {
    render(<QuestionsView questions={[createTestEntry()]} />)
    expect(screen.getByTitle('Change type')).toBeInTheDocument()
  })

  it('cycles type when type button is clicked', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} />)

    await user.click(screen.getByTitle('Change type'))

    await waitFor(() => {
      expect(RetypeEntry).toHaveBeenCalledWith(42, expect.any(String))
    })
  })

  it('shows delete button', () => {
    render(<QuestionsView questions={[createTestEntry()]} />)
    expect(screen.getByTitle('Delete entry')).toBeInTheDocument()
  })
})

describe('QuestionsView - Collapse/Expand', () => {
  it('can collapse the questions section', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ content: 'Question here' })]} />)

    expect(screen.getByText('Question here')).toBeInTheDocument()

    await user.click(screen.getByTitle('Collapse'))

    expect(screen.queryByText('Question here')).not.toBeInTheDocument()
  })

  it('can expand collapsed section', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ content: 'Question here' })]} />)

    await user.click(screen.getByTitle('Collapse'))
    await user.click(screen.getByTitle('Expand'))

    expect(screen.getByText('Question here')).toBeInTheDocument()
  })
})

describe('QuestionsView - Context Display', () => {
  it('shows context when clicking on a question', async () => {
    const user = userEvent.setup()
    const entries = [
      createTestEntry({ id: 1, content: 'Parent note', type: 'note', parentId: null }),
      createTestEntry({ id: 2, content: 'Question with parent', type: 'question', parentId: 1 }),
    ]
    render(<QuestionsView questions={entries} />)

    // Only question should be visible initially
    expect(screen.getByText('Question with parent')).toBeInTheDocument()
    expect(screen.queryByText('Parent note')).not.toBeInTheDocument()

    // Click to expand
    await user.click(screen.getByText('Question with parent'))

    // Now parent should be visible in context
    await waitFor(() => {
      expect(screen.getByText('Parent note')).toBeInTheDocument()
    })
  })

  it('shows context indicator when question has parent', () => {
    const entries = [
      createTestEntry({ id: 1, content: 'Parent note', type: 'note', parentId: null }),
      createTestEntry({ id: 2, content: 'Question with parent', type: 'question', parentId: 1 }),
    ]
    render(<QuestionsView questions={entries} />)
    expect(screen.getByTitle('Has parent context')).toBeInTheDocument()
  })
})

describe('QuestionsView - Keyboard Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('navigates down with j key', async () => {
    const user = userEvent.setup()
    const entries = [
      createTestEntry({ id: 1, content: 'Question one' }),
      createTestEntry({ id: 2, content: 'Question two' }),
    ]
    render(<QuestionsView questions={entries} />)

    await user.keyboard('j')
    await user.keyboard('j')

    // Second question should be selected (ring-2 ring-primary)
    const secondQuestion = screen.getByText('Question two').closest('div[class*="ring-2"]')
    expect(secondQuestion).toBeInTheDocument()
  })

  it('navigates up with k key', async () => {
    const user = userEvent.setup()
    const entries = [
      createTestEntry({ id: 1, content: 'Question one' }),
      createTestEntry({ id: 2, content: 'Question two' }),
    ]
    render(<QuestionsView questions={entries} />)

    await user.keyboard('jj')
    await user.keyboard('k')

    // First question should be selected
    const firstQuestion = screen.getByText('Question one').closest('div[class*="ring-2"]')
    expect(firstQuestion).toBeInTheDocument()
  })

  it('cycles priority with p key', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} />)

    await user.keyboard('j')
    await user.keyboard('p')

    await waitFor(() => {
      expect(CyclePriority).toHaveBeenCalledWith(42)
    })
  })

  it('cycles type with t key', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} />)

    await user.keyboard('j')
    await user.keyboard('t')

    await waitFor(() => {
      expect(RetypeEntry).toHaveBeenCalledWith(42, expect.any(String))
    })
  })

  it('expands context with Enter key', async () => {
    const user = userEvent.setup()
    const entries = [
      createTestEntry({ id: 1, content: 'Parent note', type: 'note', parentId: null }),
      createTestEntry({ id: 2, content: 'Question with parent', type: 'question', parentId: 1 }),
    ]
    render(<QuestionsView questions={entries} />)

    await user.keyboard('j')
    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(screen.getByText('Parent note')).toBeInTheDocument()
    })
  })

  it('cancels entry with x key', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ id: 42 })]} />)

    await user.keyboard('j')
    await user.keyboard('x')

    await waitFor(() => {
      expect(CancelEntry).toHaveBeenCalledWith(42)
    })
  })

  it('opens answer modal with a key', async () => {
    const user = userEvent.setup()
    render(<QuestionsView questions={[createTestEntry({ content: 'What is TDD?' })]} />)

    await user.keyboard('j')
    await user.keyboard('a')

    expect(screen.getByText('Answer Question')).toBeInTheDocument()
  })
})
