import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { WeekSummary } from '../WeekSummary'
import { DayEntries } from '@/types/bujo'

describe('WeekSummary', () => {
  describe('entry interaction', () => {
    it('opens context popover when attention item clicked', async () => {
      const mockDays: DayEntries[] = [
        {
          date: '2026-01-15',
          entries: [
            { id: 1, content: 'Test task', type: 'task', loggedDate: '2026-01-15', priority: 'high', parentId: null, children: [] }
          ]
        }
      ]

      render(
        <WeekSummary
          days={mockDays}
          onAction={vi.fn()}
          onNavigate={vi.fn()}
        />
      )

      await userEvent.click(screen.getByText('Test task'))

      expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
    })

    it('calls onAction when popover action triggered', async () => {
      const mockDays: DayEntries[] = [
        {
          date: '2026-01-15',
          entries: [
            { id: 1, content: 'Test task', type: 'task', loggedDate: '2026-01-15', priority: 'none', parentId: null, children: [] }
          ]
        }
      ]
      const onAction = vi.fn()

      render(
        <WeekSummary
          days={mockDays}
          onAction={onAction}
          onNavigate={vi.fn()}
        />
      )

      await userEvent.click(screen.getByText('Test task'))
      await userEvent.click(screen.getByRole('button', { name: /done/i }))

      expect(onAction).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }), 'done')
    })

    it('shows "Show all" button that calls onShowAllAttention when clicked', async () => {
      const mockDays: DayEntries[] = [
        {
          date: '2026-01-15',
          entries: Array.from({ length: 10 }, (_, i) => ({
            id: i + 1,
            content: `Task ${i + 1}`,
            type: 'task' as const,
            loggedDate: '2026-01-15',
            priority: 'none' as const,
            parentId: null,
            children: []
          }))
        }
      ]
      const onShowAll = vi.fn()

      render(
        <WeekSummary
          days={mockDays}
          onAction={vi.fn()}
          onNavigate={vi.fn()}
          onShowAllAttention={onShowAll}
        />
      )

      await userEvent.click(screen.getByText('Show all'))

      expect(onShowAll).toHaveBeenCalled()
    })
  })
})
