import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { WeekSummary } from '../WeekSummary'
import { DayEntries } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetAttentionScores: vi.fn(),
}))

import { GetAttentionScores } from '@/wailsjs/go/wails/App'

const mockGetAttentionScores = vi.mocked(GetAttentionScores)

describe('WeekSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetAttentionScores.mockResolvedValue({})
  })

  describe('entry interaction', () => {
    it('shows "Show all" button that calls onShowAllAttention when clicked', async () => {
      const scores: Record<number, { Score: number; Indicators: string[]; DaysOld: number }> = {}
      const entries = Array.from({ length: 10 }, (_, i) => {
        scores[i + 1] = { Score: 50 + i, Indicators: [], DaysOld: i }
        return {
          id: i + 1,
          content: `Task ${i + 1}`,
          type: 'task' as const,
          loggedDate: '2026-01-15',
          priority: 'none' as const,
          parentId: null,
          children: []
        }
      })
      mockGetAttentionScores.mockResolvedValue(scores)

      const mockDays: DayEntries[] = [{ date: '2026-01-15', entries }]
      const onShowAll = vi.fn()

      render(<WeekSummary days={mockDays} onShowAllAttention={onShowAll} />)

      await waitFor(() => {
        expect(screen.getByText('Show all')).toBeInTheDocument()
      })

      await userEvent.click(screen.getByText('Show all'))
      expect(onShowAll).toHaveBeenCalled()
    })

    it('displays attention items sorted by backend score', async () => {
      mockGetAttentionScores.mockResolvedValue({
        1: { Score: 20, Indicators: ['aging'], DaysOld: 3 },
        2: { Score: 80, Indicators: ['overdue', 'priority'], DaysOld: 10 },
      })

      const mockDays: DayEntries[] = [{
        date: '2026-01-15',
        entries: [
          { id: 1, content: 'Low score task', type: 'task', loggedDate: '2026-01-15', priority: 'none', parentId: null },
          { id: 2, content: 'High score task', type: 'task', loggedDate: '2026-01-15', priority: 'high', parentId: null },
        ]
      }]

      render(<WeekSummary days={mockDays} />)

      await waitFor(() => {
        const items = screen.getAllByTestId('attention-item')
        expect(items[0]).toHaveTextContent('High score task')
        expect(items[1]).toHaveTextContent('Low score task')
      })
    })

    it('shows backend indicators on attention items', async () => {
      mockGetAttentionScores.mockResolvedValue({
        1: { Score: 60, Indicators: ['overdue', 'aging'], DaysOld: 5 },
      })

      const mockDays: DayEntries[] = [{
        date: '2026-01-15',
        entries: [
          { id: 1, content: 'Overdue task', type: 'task', loggedDate: '2026-01-15', priority: 'none', parentId: null },
        ]
      }]

      render(<WeekSummary days={mockDays} />)

      await waitFor(() => {
        expect(screen.getByTestId('attention-indicators')).toBeInTheDocument()
      })
    })
  })
})
