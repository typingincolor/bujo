import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OverdueItem } from './OverdueItem'
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

describe('OverdueItem', () => {
  const now = new Date('2026-01-25T12:00:00Z')

  describe('basic rendering', () => {
    it('renders entry content', () => {
      const entry = createTestEntry({ content: 'Buy groceries' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByText('Buy groceries')).toBeInTheDocument()
    })

    it('renders entry symbol for task', () => {
      const entry = createTestEntry({ type: 'task' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('•')
    })

    it('renders entry symbol for note', () => {
      const entry = createTestEntry({ type: 'note' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('–')
    })

    it('renders entry symbol for event', () => {
      const entry = createTestEntry({ type: 'event' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('⚬')
    })

    it('renders entry symbol for question', () => {
      const entry = createTestEntry({ type: 'question' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('?')
    })
  })

  describe('context dot', () => {
    it('shows context dot when entry has parent', () => {
      const entry = createTestEntry({ parentId: 123 })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('context-dot')).toBeInTheDocument()
    })

    it('does not show context dot when entry has no parent', () => {
      const entry = createTestEntry({ parentId: null })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.queryByTestId('context-dot')).not.toBeInTheDocument()
    })
  })

  describe('priority indicator', () => {
    it('shows priority indicator for low priority', () => {
      const entry = createTestEntry({ priority: 'low' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('priority-indicator')).toHaveTextContent('!')
    })

    it('shows priority indicator for medium priority', () => {
      const entry = createTestEntry({ priority: 'medium' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('priority-indicator')).toHaveTextContent('!!')
    })

    it('shows priority indicator for high priority', () => {
      const entry = createTestEntry({ priority: 'high' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('priority-indicator')).toHaveTextContent('!!!')
    })

    it('does not show priority indicator for none priority', () => {
      const entry = createTestEntry({ priority: 'none' })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.queryByTestId('priority-indicator')).not.toBeInTheDocument()
    })
  })

  describe('breadcrumb', () => {
    it('shows breadcrumb when provided', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} now={now} breadcrumb="Project > Phase 1" />)
      expect(screen.getByText('Project > Phase 1')).toBeInTheDocument()
    })

    it('does not show breadcrumb when not provided', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.queryByTestId('breadcrumb')).not.toBeInTheDocument()
    })
  })

  describe('attention score badge', () => {
    it('shows attention badge', () => {
      const entry = createTestEntry({
        scheduledDate: '2026-01-20T00:00:00Z', // 5 days overdue
      })
      render(<OverdueItem entry={entry} now={now} />)
      expect(screen.getByTestId('attention-badge')).toBeInTheDocument()
    })

    it('shows red badge for high attention score (overdue + high priority)', () => {
      const entry = createTestEntry({
        priority: 'high',
        scheduledDate: '2026-01-20T00:00:00Z', // overdue
      })
      render(<OverdueItem entry={entry} now={now} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveClass('bg-red-500')
    })

    it('shows orange badge for medium attention score', () => {
      const entry = createTestEntry({
        scheduledDate: '2026-01-23T00:00:00Z', // 2 days overdue
      })
      render(<OverdueItem entry={entry} now={now} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveClass('bg-orange-500')
    })

    it('shows yellow badge for low attention score', () => {
      const entry = createTestEntry({
        priority: 'low',
      })
      render(<OverdueItem entry={entry} now={now} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveClass('bg-yellow-500')
    })

    it('displays score value in badge', () => {
      const entry = createTestEntry({
        priority: 'high',
        scheduledDate: '2026-01-20T00:00:00Z',
      })
      render(<OverdueItem entry={entry} now={now} />)
      const badge = screen.getByTestId('attention-badge')
      // High priority (30+20) + overdue (50) = 100
      expect(badge).toHaveTextContent('100')
    })
  })

  describe('tooltip', () => {
    it('shows tooltip with score breakdown on hover', async () => {
      const user = userEvent.setup()
      const entry = createTestEntry({
        priority: 'high',
        scheduledDate: '2026-01-20T00:00:00Z',
      })
      render(<OverdueItem entry={entry} now={now} />)

      await user.hover(screen.getByTestId('attention-badge'))

      // Radix Tooltip renders two elements with role="tooltip" (visible + sr-only)
      // Get all and check the visible one
      const tooltips = await screen.findAllByRole('tooltip')
      const visibleTooltip = tooltips.find(t => !t.style.clip?.includes('rect(0px'))
      expect(visibleTooltip).toHaveTextContent(/overdue/i)
      expect(visibleTooltip).toHaveTextContent(/priority/i)
    })
  })

  describe('click interaction', () => {
    it('is clickable and calls onSelect', async () => {
      const onSelect = vi.fn()
      const user = userEvent.setup()
      const entry = createTestEntry({ id: 42 })

      render(<OverdueItem entry={entry} now={now} onSelect={onSelect} />)
      await user.click(screen.getByRole('button'))

      expect(onSelect).toHaveBeenCalledWith(entry)
    })

    it('does not crash when onSelect is not provided', async () => {
      const user = userEvent.setup()
      const entry = createTestEntry()

      render(<OverdueItem entry={entry} now={now} />)
      await user.click(screen.getByRole('button'))
      // No error thrown
    })
  })

  describe('selected state', () => {
    it('shows selected styling when isSelected is true', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} now={now} isSelected />)
      const button = screen.getByRole('button')
      expect(button).toHaveClass('bg-accent')
    })

    it('does not show selected styling when isSelected is false', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} now={now} isSelected={false} />)
      const button = screen.getByRole('button')
      expect(button).not.toHaveClass('bg-accent')
    })
  })
})
