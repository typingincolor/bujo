import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { EntryTree } from '../EntryTree'
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo'

const mockEntries: Entry[] = [
  { id: 1, content: 'Root event', type: 'event', loggedDate: '2026-01-15', priority: 'none', parentId: null, children: [] },
  { id: 2, content: 'Child task', type: 'task', loggedDate: '2026-01-15', priority: 'none', parentId: 1, children: [] },
  { id: 3, content: 'Grandchild note', type: 'note', loggedDate: '2026-01-15', priority: 'none', parentId: 2, children: [] },
]

describe('EntryTree', () => {
  it('renders tree from root to highlighted entry', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    expect(screen.getByText('Root event')).toBeInTheDocument()
    expect(screen.getByText('Child task')).toBeInTheDocument()
    expect(screen.getByText('Grandchild note')).toBeInTheDocument()
  })

  it('shows bullet journal symbols for each entry type', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    const tree = screen.getByTestId('entry-tree')
    expect(tree.textContent).toContain(ENTRY_SYMBOLS.event) // event: '○'
    expect(tree.textContent).toContain(ENTRY_SYMBOLS.task)  // task: '•'
    expect(tree.textContent).toContain(ENTRY_SYMBOLS.note)  // note: '–'
  })

  it('highlights the target entry', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    const highlighted = screen.getByTestId('entry-tree-item-3')
    expect(highlighted).toHaveClass('bg-primary/10')
  })

  it('indents nested entries', () => {
    render(
      <EntryTree
        entries={mockEntries}
        highlightedEntryId={3}
        rootEntryId={1}
      />
    )

    const root = screen.getByTestId('entry-tree-item-1')
    const child = screen.getByTestId('entry-tree-item-2')
    const grandchild = screen.getByTestId('entry-tree-item-3')

    // Check padding-left increases with depth
    expect(root).toHaveStyle({ paddingLeft: '0px' })
    expect(child).toHaveStyle({ paddingLeft: '16px' })
    expect(grandchild).toHaveStyle({ paddingLeft: '32px' })
  })
})
