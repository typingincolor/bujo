import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { EntryItem } from './EntryItem'
import { Entry } from '@/types/bujo'

describe('EntryItem context dot', () => {
  const baseEntry: Entry = {
    id: 1,
    content: 'Test entry',
    type: 'task',
    priority: 'none',
    parentId: null,
    loggedDate: '2026-01-25',
  }

  it('shows context dot when entry has parent', () => {
    const entryWithParent = { ...baseEntry, parentId: 99 }
    render(<EntryItem entry={entryWithParent} />)
    expect(screen.getByTestId('context-dot')).toBeInTheDocument()
  })

  it('does not show context dot when entry has no parent', () => {
    render(<EntryItem entry={baseEntry} />)
    expect(screen.queryByTestId('context-dot')).not.toBeInTheDocument()
  })

  it('context dot has muted styling', () => {
    const entryWithParent = { ...baseEntry, parentId: 99 }
    render(<EntryItem entry={entryWithParent} />)
    const dot = screen.getByTestId('context-dot')
    expect(dot).toHaveClass('bg-muted-foreground')
  })
})
