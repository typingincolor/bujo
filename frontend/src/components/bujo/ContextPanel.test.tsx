import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ContextPanel } from './ContextPanel'
import { Entry } from '@/types/bujo'

const mockEntries: Entry[] = [
  { id: 1, content: 'Root entry', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
  { id: 2, content: 'Child entry', type: 'note', priority: 'none', parentId: 1, loggedDate: '2026-01-25' },
  { id: 3, content: 'Grandchild entry', type: 'task', priority: 'high', parentId: 2, loggedDate: '2026-01-25' },
]

describe('ContextPanel', () => {
  it('renders selected entry with no parent context message when entry has no ancestors', () => {
    const rootEntry = mockEntries[0]
    render(<ContextPanel selectedEntry={rootEntry} entries={mockEntries} />)
    expect(screen.getByText('Root entry')).toBeInTheDocument()
    expect(screen.getByText('No parent context')).toBeInTheDocument()
  })

  it('renders hierarchy tree when selectedEntry has ancestors', () => {
    const grandchild = mockEntries[2]
    render(<ContextPanel selectedEntry={grandchild} entries={mockEntries} />)
    expect(screen.getByText('Root entry')).toBeInTheDocument()
    expect(screen.getByText('Child entry')).toBeInTheDocument()
    expect(screen.getByText('Grandchild entry')).toBeInTheDocument()
  })

  it('highlights the selected entry in the tree', () => {
    const grandchild = mockEntries[2]
    render(<ContextPanel selectedEntry={grandchild} entries={mockEntries} />)
    const highlighted = screen.getByTestId(`context-panel-item-${grandchild.id}`)
    expect(highlighted).toHaveAttribute('data-highlighted', 'true')
  })

  it('renders nothing when selectedEntry is null', () => {
    const { container } = render(<ContextPanel selectedEntry={null} entries={mockEntries} />)
    expect(container.firstChild).toBeNull()
  })
})
