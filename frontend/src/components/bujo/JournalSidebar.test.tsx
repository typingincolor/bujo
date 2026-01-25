import { describe, it, expect, vi } from 'vitest'
import { render, screen, act } from '@testing-library/react'
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

    it('has scrollable container for context tree', () => {
      const selectedEntry = createTestEntry({ id: 2, content: 'Selected', parentId: 1 })
      const contextTree = [
        createTestEntry({ id: 1, content: 'Root', parentId: null }),
        createTestEntry({ id: 2, content: 'Selected', parentId: 1 }),
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
      const scrollableContainer = contextSection.querySelector('.overflow-y-auto')
      expect(scrollableContainer).toBeInTheDocument()
    })

    it('context section uses flex-1 to fill available space', () => {
      const selectedEntry = createTestEntry({ id: 2, content: 'Selected', parentId: 1 })
      const contextTree = [
        createTestEntry({ id: 1, content: 'Root', parentId: null }),
        createTestEntry({ id: 2, content: 'Selected', parentId: 1 }),
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
      expect(contextSection).toHaveClass('flex-1')
    })
  })

  describe('Resize functionality', () => {
    it('renders with default width of 512px', () => {
      render(<JournalSidebar overdueEntries={[]} now={new Date()} />)
      const sidebar = screen.getByTestId('overdue-sidebar')
      expect(sidebar).toHaveStyle({ width: '512px' })
    })

    it('renders resize handle when not collapsed', () => {
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          isCollapsed={false}
        />
      )
      expect(screen.getByTestId('resize-handle')).toBeInTheDocument()
    })

    it('does not render resize handle when collapsed', () => {
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          isCollapsed={true}
        />
      )
      expect(screen.queryByTestId('resize-handle')).not.toBeInTheDocument()
    })

    it('calls onWidthChange on mount with default width', () => {
      const onWidthChange = vi.fn()
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          onWidthChange={onWidthChange}
        />
      )
      expect(onWidthChange).toHaveBeenCalledWith(512)
    })

    it('updates width on resize drag', async () => {
      const onWidthChange = vi.fn()
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          onWidthChange={onWidthChange}
        />
      )

      const resizeHandle = screen.getByTestId('resize-handle')

      // Clear the initial mount call
      onWidthChange.mockClear()

      // Simulate resize: mousedown, mousemove, mouseup
      Object.defineProperty(window, 'innerWidth', { value: 1920, writable: true })

      await act(async () => {
        const mouseDownEvent = new MouseEvent('mousedown', { bubbles: true, clientX: 0 })
        resizeHandle.dispatchEvent(mouseDownEvent)

        // Simulate mouse move to resize (window width 1920, clientX 1320 = 600px sidebar)
        const mouseMoveEvent = new MouseEvent('mousemove', { bubbles: true, clientX: 1320 })
        document.dispatchEvent(mouseMoveEvent)

        // Complete resize
        const mouseUpEvent = new MouseEvent('mouseup', { bubbles: true })
        document.dispatchEvent(mouseUpEvent)
      })

      expect(onWidthChange).toHaveBeenCalledWith(600)
    })

    it('clamps width to minimum 384px', async () => {
      const onWidthChange = vi.fn()
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          onWidthChange={onWidthChange}
        />
      )

      const resizeHandle = screen.getByTestId('resize-handle')

      // Clear the initial mount call
      onWidthChange.mockClear()

      // Simulate resize to very wide (should clamp to min)
      Object.defineProperty(window, 'innerWidth', { value: 1920, writable: true })

      await act(async () => {
        const mouseDownEvent = new MouseEvent('mousedown', { bubbles: true, clientX: 0 })
        resizeHandle.dispatchEvent(mouseDownEvent)

        const mouseMoveEvent = new MouseEvent('mousemove', { bubbles: true, clientX: 1700 })
        document.dispatchEvent(mouseMoveEvent)

        // Complete resize
        const mouseUpEvent = new MouseEvent('mouseup', { bubbles: true })
        document.dispatchEvent(mouseUpEvent)
      })

      // 1920 - 1700 = 220, should clamp to 384
      expect(onWidthChange).toHaveBeenCalledWith(384)
    })

    it('clamps width to maximum 960px', async () => {
      const onWidthChange = vi.fn()
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          onWidthChange={onWidthChange}
        />
      )

      const resizeHandle = screen.getByTestId('resize-handle')

      // Clear the initial mount call
      onWidthChange.mockClear()

      // Simulate resize to very narrow (should clamp to max)
      Object.defineProperty(window, 'innerWidth', { value: 1920, writable: true })

      await act(async () => {
        const mouseDownEvent = new MouseEvent('mousedown', { bubbles: true, clientX: 0 })
        resizeHandle.dispatchEvent(mouseDownEvent)

        const mouseMoveEvent = new MouseEvent('mousemove', { bubbles: true, clientX: 800 })
        document.dispatchEvent(mouseMoveEvent)

        // Complete resize
        const mouseUpEvent = new MouseEvent('mouseup', { bubbles: true })
        document.dispatchEvent(mouseUpEvent)
      })

      // 1920 - 800 = 1120, should clamp to 960
      expect(onWidthChange).toHaveBeenCalledWith(960)
    })

    it('stops resizing on mouse up', async () => {
      const onWidthChange = vi.fn()
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={new Date()}
          onWidthChange={onWidthChange}
        />
      )

      const resizeHandle = screen.getByTestId('resize-handle')

      Object.defineProperty(window, 'innerWidth', { value: 1920, writable: true })

      await act(async () => {
        // Start resize
        const mouseDownEvent = new MouseEvent('mousedown', { bubbles: true, clientX: 0 })
        resizeHandle.dispatchEvent(mouseDownEvent)

        // Move once
        const mouseMoveEvent1 = new MouseEvent('mousemove', { bubbles: true, clientX: 1320 })
        document.dispatchEvent(mouseMoveEvent1)

        // End resize
        const mouseUpEvent = new MouseEvent('mouseup', { bubbles: true })
        document.dispatchEvent(mouseUpEvent)
      })

      // Clear previous calls
      onWidthChange.mockClear()

      await act(async () => {
        // Move again - should not trigger width change
        const mouseMoveEvent2 = new MouseEvent('mousemove', { bubbles: true, clientX: 1200 })
        document.dispatchEvent(mouseMoveEvent2)
      })

      expect(onWidthChange).not.toHaveBeenCalled()
    })
  })
})
