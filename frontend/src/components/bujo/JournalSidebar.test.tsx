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
    it('renders Overdue Items section', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      expect(screen.getByRole('button', { name: 'Overdue' })).toBeInTheDocument()
    })

    it('renders overdue entries with attention scores', async () => {
      const user = userEvent.setup()
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)

      // Expand the collapsed section
      await user.click(screen.getByRole('button', { name: 'Overdue' }))

      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
      expect(screen.getByText('Overdue task 2')).toBeInTheDocument()
    })

    it('shows entry count in section header', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      expect(screen.getByText(/overdue.*\(2\)/i)).toBeInTheDocument()
    })

    it('starts collapsed by default', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      // Radix Collapsible removes content from DOM when collapsed
      expect(screen.queryByText('Overdue task 1')).not.toBeInTheDocument()
    })

    it('expands Overdue Items section when clicking header', async () => {
      const user = userEvent.setup()
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)

      const triggerButton = screen.getByRole('button', { name: 'Overdue' })
      await user.click(triggerButton)
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()
    })

    it('collapses Overdue Items section when clicking header again', async () => {
      const user = userEvent.setup()
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)

      const triggerButton = screen.getByRole('button', { name: 'Overdue' })

      // Expand
      await user.click(triggerButton)
      expect(screen.getByText('Overdue task 1')).toBeInTheDocument()

      // Collapse
      await user.click(triggerButton)
      expect(screen.queryByText('Overdue task 1')).not.toBeInTheDocument()
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

      // Expand the collapsed section
      await user.click(screen.getByRole('button', { name: 'Overdue' }))

      await user.click(screen.getByText('Overdue task 1'))
      expect(onSelectEntry).toHaveBeenCalledWith(mockOverdueEntries[0])
    })
  })

  describe('Context section', () => {
    it('always shows Context section', () => {
      render(<JournalSidebar overdueEntries={mockOverdueEntries} now={now} />)
      expect(screen.getByText('Context')).toBeInTheDocument()
    })

    it('shows ancestor hierarchy in Context when entry selected', () => {
      const selectedEntry = createTestEntry({ id: 3, content: 'Selected' })
      const ancestors = [
        createTestEntry({ id: 2, content: 'Parent' }),
        createTestEntry({ id: 1, content: 'Grandparent' }),
      ]

      render(
        <JournalSidebar
          overdueEntries={[]}
          now={now}
          selectedEntry={selectedEntry}
          ancestors={ancestors}
        />
      )

      expect(screen.getByText('Grandparent')).toBeInTheDocument()
      expect(screen.getByText('Parent')).toBeInTheDocument()
      expect(screen.getByText('Selected')).toBeInTheDocument()
    })

    it('shows "No context" when no entry selected', () => {
      render(<JournalSidebar overdueEntries={[]} now={now} />)
      expect(screen.getByText(/no entry selected/i)).toBeInTheDocument()
    })

    it('shows "No context" when entry has no ancestors', () => {
      const selectedEntry = createTestEntry({ id: 1, content: 'Orphan' })
      render(
        <JournalSidebar
          overdueEntries={[]}
          now={now}
          selectedEntry={selectedEntry}
          ancestors={[]}
        />
      )
      expect(screen.getByText(/no context/i)).toBeInTheDocument()
    })

    it('shows entry symbols for each ancestor level', () => {
      const selectedEntry = createTestEntry({ id: 2, content: 'Child', type: 'task' })
      const ancestors = [
        createTestEntry({ id: 1, content: 'Parent', type: 'note' }),
      ]

      render(
        <JournalSidebar
          overdueEntries={[]}
          now={now}
          selectedEntry={selectedEntry}
          ancestors={ancestors}
        />
      )

      const contextSection = screen.getByTestId('context-section')
      expect(contextSection).toHaveTextContent('–') // note symbol
      expect(contextSection).toHaveTextContent('•') // task symbol
    })
  })

  describe('empty state', () => {
    it('shows empty message when no overdue entries', async () => {
      const user = userEvent.setup()
      render(<JournalSidebar overdueEntries={[]} now={now} />)

      // Expand the collapsed section
      await user.click(screen.getByRole('button', { name: 'Overdue' }))

      expect(screen.getByText(/no overdue items/i)).toBeInTheDocument()
    })
  })

  describe('highlight selected entry', () => {
    it('highlights selected entry in overdue list', async () => {
      const user = userEvent.setup()
      const selectedEntry = mockOverdueEntries[0]
      render(
        <JournalSidebar
          overdueEntries={mockOverdueEntries}
          now={now}
          selectedEntry={selectedEntry}
        />
      )

      // Expand the collapsed section
      await user.click(screen.getByRole('button', { name: 'Overdue' }))

      // Find the button containing the text (within the overdue items, not the header)
      const allButtons = screen.getAllByRole('button')
      const overdueItem = allButtons.find(
        (btn) => btn.textContent?.includes('Overdue task 1') && btn.getAttribute('aria-label') !== 'Overdue'
      )
      expect(overdueItem).toHaveClass('bg-accent')
    })
  })
})
