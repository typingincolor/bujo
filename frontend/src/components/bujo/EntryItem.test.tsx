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
})
