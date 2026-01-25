import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { WeekSummary } from '../WeekSummary'
import { DayEntries } from '@/types/bujo'

describe('WeekSummary', () => {
  describe('entry interaction', () => {
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
          onShowAllAttention={onShowAll}
        />
      )

      await userEvent.click(screen.getByText('Show all'))

      expect(onShowAll).toHaveBeenCalled()
    })
  })
})
