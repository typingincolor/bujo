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
  describe('Pending Tasks section', () => {
    it('shows pending tasks count in header', () => {
      const entries = [
        createTestEntry({ id: 1, content: 'Task 1' }),
        createTestEntry({ id: 2, content: 'Task 2' }),
      ]
      render(<JournalSidebar overdueEntries={entries} now={new Date()} />)
      expect(screen.getByText('Pending Tasks (2)')).toBeInTheDocument()
    })

    it('shows "No pending tasks" when list is empty', () => {
      render(<JournalSidebar overdueEntries={[]} now={new Date()} />)
      expect(screen.getByText(/no pending tasks/i)).toBeInTheDocument()
    })

    it('shows action buttons on each pending task', () => {
      const entries = [createTestEntry({ id: 1, content: 'Task 1' })]
      const callbacks = {
        onMarkDone: vi.fn(),
        onEdit: vi.fn(),
      }

      render(
        <JournalSidebar
          overdueEntries={entries}
          now={new Date()}
          callbacks={callbacks}
        />
      )

      expect(screen.getByTestId('entry-action-bar')).toBeInTheDocument()
    })

    it('calls onMarkDone when done button clicked on pending task', async () => {
      const user = userEvent.setup()
      const entries = [createTestEntry({ id: 1, content: 'Task 1' })]
      const callbacks = {
        onMarkDone: vi.fn(),
      }

      render(
        <JournalSidebar
          overdueEntries={entries}
          now={new Date()}
          callbacks={callbacks}
        />
      )

      await user.click(screen.getByTitle('Cancel entry'))
      expect(callbacks.onMarkDone).toHaveBeenCalledWith(entries[0])
    })

    it('calls onMigrate when migrate button clicked on pending task', async () => {
      const user = userEvent.setup()
      const entries = [createTestEntry({ id: 1, type: 'task', content: 'Task 1' })]
      const callbacks = {
        onMigrate: vi.fn(),
      }

      render(
        <JournalSidebar
          overdueEntries={entries}
          now={new Date()}
          callbacks={callbacks}
        />
      )

      await user.click(screen.getByTitle('Migrate entry'))
      expect(callbacks.onMigrate).toHaveBeenCalledWith(entries[0])
    })

    it('calls onEdit when edit button clicked on pending task', async () => {
      const user = userEvent.setup()
      const entries = [createTestEntry({ id: 1, content: 'Task 1' })]
      const callbacks = {
        onEdit: vi.fn(),
      }

      render(
        <JournalSidebar
          overdueEntries={entries}
          now={new Date()}
          callbacks={callbacks}
        />
      )

      await user.click(screen.getByTitle('Edit entry'))
      expect(callbacks.onEdit).toHaveBeenCalledWith(entries[0])
    })

    it('only shows tasks, not notes or events', () => {
      const entries = [
        createTestEntry({ id: 1, content: 'A task', type: 'task' }),
        createTestEntry({ id: 2, content: 'A note', type: 'note' }),
        createTestEntry({ id: 3, content: 'An event', type: 'event' }),
      ]

      render(<JournalSidebar overdueEntries={entries} now={new Date()} />)

      expect(screen.getByText('Pending Tasks (1)')).toBeInTheDocument()
      expect(screen.getByText('A task')).toBeInTheDocument()
      expect(screen.queryByText('A note')).not.toBeInTheDocument()
      expect(screen.queryByText('An event')).not.toBeInTheDocument()
    })
  })

  describe('Collapse functionality', () => {
    it('renders collapse toggle button when onToggleCollapse provided', () => {
      const onToggleCollapse = vi.fn()
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          onToggleCollapse={onToggleCollapse}
        />
      )
      expect(screen.getByRole('button', { name: /toggle sidebar/i })).toBeInTheDocument()
    })

    it('does not render toggle button when onToggleCollapse not provided', () => {
      render(<JournalSidebar overdueEntries={[]} now={new Date()} />)
      expect(screen.queryByRole('button', { name: /toggle sidebar/i })).not.toBeInTheDocument()
    })

    it('calls onToggleCollapse when collapse button clicked', async () => {
      const user = userEvent.setup()
      const onToggleCollapse = vi.fn()

      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          onToggleCollapse={onToggleCollapse}
        />
      )

      await user.click(screen.getByRole('button', { name: /toggle sidebar/i }))
      expect(onToggleCollapse).toHaveBeenCalledTimes(1)
    })

    it('hides content when collapsed', () => {
      const entries = [createTestEntry({ id: 1, content: 'Task 1' })]
      render(
        <JournalSidebar
          overdueEntries={entries}
          now={new Date()}
          isCollapsed={true}
        />
      )

      expect(screen.queryByText('Pending Tasks (1)')).not.toBeInTheDocument()
      expect(screen.queryByText('Task 1')).not.toBeInTheDocument()
    })

    it('shows content when expanded', () => {
      const entries = [createTestEntry({ id: 1, content: 'Task 1' })]
      render(
        <JournalSidebar
          overdueEntries={entries}
          now={new Date()}
          isCollapsed={false}
        />
      )

      expect(screen.getByText('Pending Tasks (1)')).toBeInTheDocument()
      expect(screen.getByText('Task 1')).toBeInTheDocument()
    })
  })

  describe('Context section', () => {
    it('shows Context section header', () => {
      render(<JournalSidebar overdueEntries={[]} now={new Date()} />)
      expect(screen.getByText('Context')).toBeInTheDocument()
    })

    it('shows full tree structure in Context when entry selected', () => {
      const selectedEntry = createTestEntry({ id: 3, content: 'Selected', parentId: 2 })
      const contextTree = [
        createTestEntry({ id: 1, content: 'Root', parentId: null }),
        createTestEntry({ id: 2, content: 'Parent', parentId: 1 }),
        createTestEntry({ id: 3, content: 'Selected', parentId: 2 }),
        createTestEntry({ id: 4, content: 'Sibling', parentId: 2 }),
      ]

      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          selectedEntry={selectedEntry}
          contextTree={contextTree}
        />
      )

      expect(screen.getByText('Root')).toBeInTheDocument()
      expect(screen.getByText('Parent')).toBeInTheDocument()
      expect(screen.getAllByText('Selected').length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('Sibling')).toBeInTheDocument()
    })

    it('shows "No context" when entry is root level', () => {
      const selectedEntry = createTestEntry({ id: 1, content: 'Root entry', parentId: null })
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
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
          now={new Date()}
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
          now={new Date()}
          selectedEntry={selectedEntry}
          contextTree={contextTree}
        />
      )

      const contextSection = screen.getByTestId('context-section')
      const selectedText = contextSection.querySelector('.font-medium')
      expect(selectedText).toBeInTheDocument()
    })
  })
})
