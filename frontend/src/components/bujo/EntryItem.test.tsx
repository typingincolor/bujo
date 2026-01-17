import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { EntryItem } from './EntryItem'
import { Entry } from '@/types/bujo'

const createTestEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  entityId: 'e1',
  type: 'task',
  content: 'Test entry',
  priority: null,
  parentId: null,
  depth: 0,
  createdAt: '2026-01-17',
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
})
