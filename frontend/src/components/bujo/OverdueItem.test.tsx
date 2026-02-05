import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OverdueItem } from './OverdueItem'
import { Entry } from '@/types/bujo'
import { AttentionScore } from '@/hooks/useAttentionScores'

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

function createScore(overrides: Partial<AttentionScore> = {}): AttentionScore {
  return {
    score: 50,
    indicators: [],
    daysOld: 0,
    ...overrides,
  }
}

describe('OverdueItem', () => {

  describe('basic rendering', () => {
    it('renders entry content', () => {
      const entry = createTestEntry({ content: 'Buy groceries' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByText('Buy groceries')).toBeInTheDocument()
    })

    it('renders entry symbol for task', () => {
      const entry = createTestEntry({ type: 'task' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('•')
    })

    it('renders entry symbol for note', () => {
      const entry = createTestEntry({ type: 'note' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('–')
    })

    it('renders entry symbol for event', () => {
      const entry = createTestEntry({ type: 'event' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('⚬')
    })

    it('renders entry symbol for question', () => {
      const entry = createTestEntry({ type: 'question' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('entry-symbol')).toHaveTextContent('?')
    })
  })

  describe('context dot', () => {
    it('shows context dot when entry has parent', () => {
      const entry = createTestEntry({ parentId: 123 })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('context-dot')).toBeInTheDocument()
    })

    it('does not show context dot when entry has no parent', () => {
      const entry = createTestEntry({ parentId: null })
      render(<OverdueItem entry={entry} />)
      expect(screen.queryByTestId('context-dot')).not.toBeInTheDocument()
    })

    it('always reserves space for context dot to keep symbols aligned', () => {
      const entry = createTestEntry({ parentId: null })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('context-dot-container')).toBeInTheDocument()
    })
  })

  describe('priority indicator', () => {
    it('shows priority indicator for low priority', () => {
      const entry = createTestEntry({ priority: 'low' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('priority-indicator')).toHaveTextContent('!')
    })

    it('shows priority indicator for medium priority', () => {
      const entry = createTestEntry({ priority: 'medium' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('priority-indicator')).toHaveTextContent('!!')
    })

    it('shows priority indicator for high priority', () => {
      const entry = createTestEntry({ priority: 'high' })
      render(<OverdueItem entry={entry} />)
      expect(screen.getByTestId('priority-indicator')).toHaveTextContent('!!!')
    })

    it('does not show priority indicator for none priority', () => {
      const entry = createTestEntry({ priority: 'none' })
      render(<OverdueItem entry={entry} />)
      expect(screen.queryByTestId('priority-indicator')).not.toBeInTheDocument()
    })
  })

  describe('breadcrumb', () => {
    it('shows breadcrumb when provided', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} breadcrumb="Project > Phase 1" />)
      expect(screen.getByText('Project > Phase 1')).toBeInTheDocument()
    })

    it('does not show breadcrumb when not provided', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} />)
      expect(screen.queryByTestId('breadcrumb')).not.toBeInTheDocument()
    })
  })

  describe('attention score badge', () => {
    it('shows attention badge with score from prop', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} attentionScore={createScore({ score: 65 })} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveTextContent('65')
    })

    it('shows red badge for score >= 80', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} attentionScore={createScore({ score: 85 })} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveClass('bg-red-500')
    })

    it('shows orange badge for score >= 50', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} attentionScore={createScore({ score: 55 })} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveClass('bg-orange-500')
    })

    it('shows yellow badge for score < 50', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} attentionScore={createScore({ score: 30 })} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveClass('bg-yellow-500')
    })

    it('defaults to score 0 when no attentionScore provided', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} />)
      const badge = screen.getByTestId('attention-badge')
      expect(badge).toHaveTextContent('0')
    })
  })

  describe('tooltip', () => {
    it('shows tooltip with indicator labels on hover', async () => {
      const user = userEvent.setup()
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} attentionScore={createScore({ score: 90, indicators: ['overdue', 'priority'] })} />)

      await user.hover(screen.getByTestId('attention-badge'))

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

      render(<OverdueItem entry={entry} onSelect={onSelect} />)
      await user.click(screen.getByRole('button'))

      expect(onSelect).toHaveBeenCalledWith(entry)
    })

    it('does not crash when onSelect is not provided', async () => {
      const user = userEvent.setup()
      const entry = createTestEntry()

      render(<OverdueItem entry={entry} />)
      await user.click(screen.getByRole('button'))
    })
  })

  describe('selected state', () => {
    it('shows selected styling when isSelected is true', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} isSelected />)
      const button = screen.getByRole('button')
      expect(button).toHaveClass('ring-primary/30')
    })

    it('does not show selected styling when isSelected is false', () => {
      const entry = createTestEntry()
      render(<OverdueItem entry={entry} isSelected={false} />)
      const button = screen.getByRole('button')
      expect(button).not.toHaveClass('ring-primary/30')
    })
  })
})
