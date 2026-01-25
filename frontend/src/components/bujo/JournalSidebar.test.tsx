import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { JournalSidebar } from './JournalSidebar'
import { Entry } from '@/types/bujo'

function createTestEntry(overrides: Partial<Entry> = {}): Entry {
  return {
    id: 1,
    content: 'Test entry',
    type: 'task',
    priority: 'none',
    parentId: null,
    loggedDate: '2026-01-25T10:00:00Z',
    ...overrides,
  }
}

describe('JournalSidebar', () => {
  const now = new Date('2026-01-25T12:00:00Z')

  const mockOverdueEntries = [
    createTestEntry({ id: 1, content: 'Overdue task 1', type: 'task' }),
    createTestEntry({ id: 2, content: 'Overdue task 2', type: 'task' }),
  ]

  describe('Overdue section', () => {
    it('renders Pending Tasks section header', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      expect(screen.getByText(/pending tasks.*\(2\)/i)).toBeInTheDocument()
    })

    it('only shows task entries, filtering out notes and events', () => {
      const mixedEntries = [
        createTestEntry({ id: 1, content: 'Task entry', type: 'task' }),
        createTestEntry({ id: 2, content: 'Note entry', type: 'note' }),
        createTestEntry({ id: 3, content: 'Event entry', type: 'event' }),
        createTestEntry({ id: 4, content: 'Another task', type: 'task' }),
        createTestEntry({ id: 5, content: 'Question entry', type: 'question' }),
      ]

      render(<JournalSidebar overdueEntries={mixedEntries} now={now} />)

      // Only tasks should appear
      expect(screen.getByText('Task entry')).toBeInTheDocument()
      expect(screen.getByText('Another task')).toBeInTheDocument()

      // Non-tasks should not appear
      expect(screen.queryByText('Note entry')).not.toBeInTheDocument()
      expect(screen.queryByText('Event entry')).not.toBeInTheDocument()
      expect(screen.queryByText('Question entry')).not.toBeInTheDocument()

      // Count should only include tasks (2, not 5)
      expect(screen.getByText(/pending tasks.*\(2\)/i)).toBeInTheDocument()
    })

    it('is not collapsible - no collapse button exists', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      // Should not have a button role for section header
      expect(screen.queryByRole('button', { name: /pending tasks/i })).not.toBeInTheDocument()
    })

    it('renders overdue entries', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)

      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      expect(screen.getByText('Overdue task 2')).toBeInTheDocument()
    })

    it('shows entry count in section header', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      expect(screen.getByText(/pending tasks.*\(2\)/i)).toBeInTheDocument()
    })

    it('always shows entries without collapse functionality', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      // Entries always visible - no interaction needed
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      expect(screen.getByText('Overdue task 2')).toBeInTheDocument()
    })

    it('calls onSelectEntry when clicking overdue item', async () => {
      const onSelectEntry = vi.fn()
      const user = userEvent.setup()

      render(
        <JournalSidebar
          overdueEntries={mockOverdueEntries}
          now={now}
          onSelectEntry={onSelectEntry}
        />
      )

      await user.click(screen.getByText('Overdue task 1'))
      expect(onSelectEntry).toHaveBeenCalledWith(mockOverdueEntries[0])
    })
  })

  describe('Context section', () => {
    it('always shows Context section', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      expect(screen.getByText('Context')).toBeInTheDocument()
    })

    it('shows full tree structure in Context when entry selected', () => {
      const selectedEntry = createTestEntry({ id: 3, content: 'Selected', parentId: 2 })
      // Context tree is a flat list that gets built into a tree
      // Root -> Child1 (selected entry's sibling) and Child2 (parent of selected) -> Selected
      const contextTree = [
        createTestEntry({ id: 1, content: 'Root', parentId: null }),
        createTestEntry({ id: 2, content: 'Parent', parentId: 1 }),
        createTestEntry({ id: 3, content: 'Selected', parentId: 2 }),
        createTestEntry({ id: 4, content: 'Sibling', parentId: 2 }),
      ]

      render(
        <JournalSidebar
          overdueEntries={[]}
          now={now}
          selectedEntry={selectedEntry}
          contextTree={contextTree}
        />
      )

      expect(screen.getByText('Root')).toBeInTheDocument()
      expect(screen.getByText('Parent')).toBeInTheDocument()
      expect(screen.getByText('Selected')).toBeInTheDocument()
      expect(screen.getByText('Sibling')).toBeInTheDocument()
    })

    it('shows "No entry selected" when no entry selected', () => {
      render(<JournalSidebar overdueEntries={[]} now={now} />)
      expect(screen.getByText(/no entry selected/i)).toBeInTheDocument()
    })

    it('shows "No context" when entry is selected but tree is empty', () => {
      const selectedEntry = createTestEntry({ id: 1, content: 'Orphan' })
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={now}
          selectedEntry={selectedEntry}
          contextTree={[]}
        />
      )
      expect(screen.getByText(/no context/i)).toBeInTheDocument()
    })

    it('shows entry symbols for each tree level', () => {
      const selectedEntry = createTestEntry({ id: 2, content: 'Child', type: 'task', parentId: 1 })
      const contextTree = [
        createTestEntry({ id: 1, content: 'Parent', type: 'note', parentId: null }),
        createTestEntry({ id: 2, content: 'Child', type: 'task', parentId: 1 }),
      ]

      render(
        <JournalSidebar
          overdueEntries={[]}
          now={now}
          selectedEntry={selectedEntry}
          contextTree={contextTree}
        />
      )

      const contextSection = screen.getByTestId('context-section')
      expect(contextSection).toHaveTextContent('–') // note symbol
      expect(contextSection).toHaveTextContent('•') // task symbol
    })

    it('highlights the selected entry in the tree', () => {
      const selectedEntry = createTestEntry({ id: 2, content: 'Selected Child', type: 'task', parentId: 1 })
      const contextTree = [
        createTestEntry({ id: 1, content: 'Parent', type: 'note', parentId: null }),
        createTestEntry({ id: 2, content: 'Selected Child', type: 'task', parentId: 1 }),
      ]

      render(
        <JournalSidebar
          overdueEntries={[]}
          now={now}
          selectedEntry={selectedEntry}
          contextTree={contextTree}
        />
      )

      // The selected entry should have font-medium class
      const selectedText = screen.getByText('Selected Child')
      expect(selectedText.closest('div')).toHaveClass('font-medium')
    })
  })

  describe('empty state', () => {
    it('shows empty message when no pending tasks', () => {
      render(<JournalSidebar overdueEntries={[]} now={now} />)

      // Empty message always visible since section is not collapsible
      expect(screen.getByText(/no pending tasks/i)).toBeInTheDocument()
    })
  })

  describe('highlight selected entry', () => {
    it('highlights selected entry in overdue list', () => {
      const selectedEntry = mockOverdueEntries[0]
      render(
        <JournalSidebar
          overdueEntries={mockOverdueEntries}
          now={now}
          selectedEntry={selectedEntry}
        />
      )

      // Find the button containing the text
      const allButtons = screen.getAllByRole('button')
      const overdueItem = allButtons.find(
        (btn) => btn.textContent?.includes('Overdue task 1')
      )
      expect(overdueItem).toHaveClass('bg-accent')
    })
  })
})
