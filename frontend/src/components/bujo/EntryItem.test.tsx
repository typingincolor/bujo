import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { EntryItem } from './EntryItem'
import { Entry } from '@/types/bujo'

const createTestEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  type: 'task',
  content: 'Test entry',
  priority: 'none',
  parentId: null,
  loggedDate: '2026-01-17',
  children: [],
  ...overrides,
})

describe('EntryItem', () => {
  it('renders entry content', () => {
    render(<EntryItem entry={createTestEntry({ content: 'My task' })} />)
    expect(screen.getByText('My task')).toBeInTheDocument()
  })

  it('shows edit button on hover', () => {
    render(<EntryItem entry={createTestEntry()} onEdit={() => {}} />)
    const button = screen.getByTitle('Edit entry')
    expect(button).toBeInTheDocument()
  })

  it('shows delete button on hover', () => {
    render(<EntryItem entry={createTestEntry()} onDelete={() => {}} />)
    const button = screen.getByTitle('Delete entry')
    expect(button).toBeInTheDocument()
  })

  it('calls onEdit when edit button is clicked', () => {
    const onEdit = vi.fn()
    render(<EntryItem entry={createTestEntry()} onEdit={onEdit} />)

    fireEvent.click(screen.getByTitle('Edit entry'))
    expect(onEdit).toHaveBeenCalledTimes(1)
  })

  it('calls onDelete when delete button is clicked', () => {
    const onDelete = vi.fn()
    render(<EntryItem entry={createTestEntry()} onDelete={onDelete} />)

    fireEvent.click(screen.getByTitle('Delete entry'))
    expect(onDelete).toHaveBeenCalledTimes(1)
  })

  it('does not render edit button when onEdit is not provided', () => {
    render(<EntryItem entry={createTestEntry()} />)
    expect(screen.queryByTitle('Edit entry')).not.toBeInTheDocument()
  })

  it('does not render delete button when onDelete is not provided', () => {
    render(<EntryItem entry={createTestEntry()} />)
    expect(screen.queryByTitle('Delete entry')).not.toBeInTheDocument()
  })

  it('stops propagation when clicking edit button', () => {
    const onEdit = vi.fn()
    const onToggleDone = vi.fn()
    render(
      <EntryItem
        entry={createTestEntry()}
        onEdit={onEdit}
        onToggleDone={onToggleDone}
      />
    )

    fireEvent.click(screen.getByTitle('Edit entry'))
    expect(onEdit).toHaveBeenCalledTimes(1)
    expect(onToggleDone).not.toHaveBeenCalled()
  })

  it('stops propagation when clicking delete button', () => {
    const onDelete = vi.fn()
    const onToggleDone = vi.fn()
    render(
      <EntryItem
        entry={createTestEntry()}
        onDelete={onDelete}
        onToggleDone={onToggleDone}
      />
    )

    fireEvent.click(screen.getByTitle('Delete entry'))
    expect(onDelete).toHaveBeenCalledTimes(1)
    expect(onToggleDone).not.toHaveBeenCalled()
  })

  describe('question/answer entry types', () => {
    it('renders question entry with ? symbol', () => {
      render(<EntryItem entry={createTestEntry({ type: 'question', content: 'What is TDD?' })} />)
      expect(screen.getByText('?')).toBeInTheDocument()
      expect(screen.getByText('What is TDD?')).toBeInTheDocument()
    })

    it('renders answered entry with ★ symbol', () => {
      render(<EntryItem entry={createTestEntry({ type: 'answered', content: 'What is TDD?' })} />)
      expect(screen.getByText('★')).toBeInTheDocument()
      expect(screen.getByText('What is TDD?')).toBeInTheDocument()
    })

    it('renders answer entry with ↳ symbol', () => {
      render(<EntryItem entry={createTestEntry({ type: 'answer', content: 'Test-Driven Development' })} />)
      expect(screen.getByText('↳')).toBeInTheDocument()
      expect(screen.getByText('Test-Driven Development')).toBeInTheDocument()
    })

    it('question entries are not toggleable', () => {
      const onToggleDone = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'question' })} onToggleDone={onToggleDone} />)

      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onToggleDone).not.toHaveBeenCalled()
    })

    it('answered entries are not toggleable', () => {
      const onToggleDone = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'answered' })} onToggleDone={onToggleDone} />)

      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onToggleDone).not.toHaveBeenCalled()
    })

    it('answer entries are not toggleable', () => {
      const onToggleDone = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'answer' })} onToggleDone={onToggleDone} />)

      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onToggleDone).not.toHaveBeenCalled()
    })

    it('shows answer button for question entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'question', content: 'What is TDD?' })}
          onAnswer={() => {}}
        />
      )
      expect(screen.getByTitle('Answer question')).toBeInTheDocument()
    })

    it('calls onAnswer when answer button is clicked', () => {
      const onAnswer = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'question', content: 'What is TDD?' })}
          onAnswer={onAnswer}
        />
      )

      fireEvent.click(screen.getByTitle('Answer question'))
      expect(onAnswer).toHaveBeenCalledTimes(1)
    })

    it('does not show answer button for non-question entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onAnswer={() => {}}
        />
      )
      expect(screen.queryByTitle('Answer question')).not.toBeInTheDocument()
    })
  })

  describe('cancel/uncancel functionality', () => {
    it('shows cancel button for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCancel={() => {}}
        />
      )
      expect(screen.getByTitle('Cancel entry')).toBeInTheDocument()
    })

    it('shows uncancel button for cancelled entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'cancelled' })}
          onUncancel={() => {}}
        />
      )
      expect(screen.getByTitle('Uncancel entry')).toBeInTheDocument()
    })

    it('calls onCancel when cancel button is clicked', () => {
      const onCancel = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCancel={onCancel}
        />
      )

      fireEvent.click(screen.getByTitle('Cancel entry'))
      expect(onCancel).toHaveBeenCalledTimes(1)
    })

    it('calls onUncancel when uncancel button is clicked', () => {
      const onUncancel = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'cancelled' })}
          onUncancel={onUncancel}
        />
      )

      fireEvent.click(screen.getByTitle('Uncancel entry'))
      expect(onUncancel).toHaveBeenCalledTimes(1)
    })

    it('renders cancelled entry with strikethrough style', () => {
      render(<EntryItem entry={createTestEntry({ type: 'cancelled', content: 'Cancelled task' })} />)
      const content = screen.getByText('Cancelled task')
      expect(content).toHaveClass('line-through')
    })

    it('does not show cancel button for cancelled entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'cancelled' })}
          onCancel={() => {}}
        />
      )
      expect(screen.queryByTitle('Cancel entry')).not.toBeInTheDocument()
    })

    it('does not show uncancel button for non-cancelled entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onUncancel={() => {}}
        />
      )
      expect(screen.queryByTitle('Uncancel entry')).not.toBeInTheDocument()
    })
  })

  describe('priority functionality', () => {
    it('shows priority button for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCyclePriority={() => {}}
        />
      )
      expect(screen.getByTitle('Cycle priority')).toBeInTheDocument()
    })

    it('calls onCyclePriority when priority button is clicked', () => {
      const onCyclePriority = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCyclePriority={onCyclePriority}
        />
      )

      fireEvent.click(screen.getByTitle('Cycle priority'))
      expect(onCyclePriority).toHaveBeenCalledTimes(1)
    })

    it('displays priority indicator for high priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'high' })} />)
      expect(screen.getByText('!!!')).toBeInTheDocument()
    })

    it('displays priority indicator for medium priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'medium' })} />)
      expect(screen.getByText('!!')).toBeInTheDocument()
    })

    it('displays priority indicator for low priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'low' })} />)
      expect(screen.getByText('!')).toBeInTheDocument()
    })

    it('does not display priority indicator for none priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'none' })} />)
      // None priority shouldn't show any indicator
      expect(screen.queryByText('!')).not.toBeInTheDocument()
      expect(screen.queryByText('!!')).not.toBeInTheDocument()
      expect(screen.queryByText('!!!')).not.toBeInTheDocument()
    })
  })
})
