import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { WeekSummary } from './WeekSummary'
import { DayEntries, Entry } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetAttentionScores: vi.fn(),
}))

import { GetAttentionScores } from '@/wailsjs/go/wails/App'

const mockGetAttentionScores = vi.mocked(GetAttentionScores)

const createEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  type: 'task',
  content: 'Test task',
  priority: 'none',
  parentId: null,
  loggedDate: new Date().toISOString(),
  ...overrides,
})

const createDay = (overrides: Partial<DayEntries> = {}): DayEntries => ({
  date: new Date().toISOString().split('T')[0],
  location: '',
  mood: '',
  weather: '',
  entries: [],
  ...overrides,
})

describe('WeekSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetAttentionScores.mockResolvedValue({})
  })

  describe('Task Flow', () => {
    it('renders task flow section', () => {
      render(<WeekSummary days={[]} />)
      expect(screen.getByText('Task Flow')).toBeInTheDocument()
    })

    it('shows created count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'task' }),
            createEntry({ id: 2, type: 'task' }),
            createEntry({ id: 3, type: 'note' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Created')).toBeInTheDocument()
      expect(screen.getByTestId('task-flow-created')).toHaveTextContent('2') // 2 tasks created
    })

    it('shows done count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'done' }),
            createEntry({ id: 2, type: 'task' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Done')).toBeInTheDocument()
    })

    it('shows migrated count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'migrated' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Migrated')).toBeInTheDocument()
    })

    it('shows open count', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'task' }),
            createEntry({ id: 2, type: 'task' }),
            createEntry({ id: 3, type: 'done' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Open')).toBeInTheDocument()
    })
  })

  describe('Meetings', () => {
    it('renders meetings section', () => {
      render(<WeekSummary days={[]} />)
      expect(screen.getByText('Meetings')).toBeInTheDocument()
    })

    it('shows events with children', async () => {
      mockGetAttentionScores.mockResolvedValue({
        3: { Score: 10, Indicators: [], DaysOld: 0 },
      })

      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'event', content: 'Team standup' }),
            createEntry({ id: 2, type: 'note', content: 'Note 1', parentId: 1 }),
            createEntry({ id: 3, type: 'task', content: 'Action', parentId: 1 }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('Team standup')).toBeInTheDocument()
      expect(screen.getByText(/2 items/i)).toBeInTheDocument()
    })

    it('does not show events without children', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'event', content: 'Solo event' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.queryByText('Solo event')).not.toBeInTheDocument()
    })
  })

  describe('Entry Symbols', () => {
    it('shows event symbol in meetings section', () => {
      mockGetAttentionScores.mockResolvedValue({
        2: { Score: 20, Indicators: [], DaysOld: 0 },
      })

      const days: DayEntries[] = [{
        date: '2026-01-25',
        entries: [
          { id: 1, content: 'Team standup', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
          { id: 2, content: 'Action item', type: 'task', priority: 'none', parentId: 1, loggedDate: '2026-01-25' },
        ],
      }]
      render(<WeekSummary days={days} />)

      const meetingSection = screen.getByTestId('week-summary-meetings')
      expect(meetingSection).toHaveTextContent('⚬') // event symbol
    })

    it('shows task symbol in attention section for tasks', async () => {
      mockGetAttentionScores.mockResolvedValue({
        1: { Score: 50, Indicators: ['priority'], DaysOld: 6 },
      })

      const days: DayEntries[] = [{
        date: '2026-01-19',
        entries: [
          { id: 1, content: 'Old task', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-19' },
        ],
      }]
      render(<WeekSummary days={days} />)

      await waitFor(() => {
        const attentionSection = screen.getByTestId('week-summary-attention')
        expect(attentionSection).toHaveTextContent('•') // task symbol
      })
    })
  })

  describe('Needs Attention', () => {
    it('renders needs attention section', () => {
      render(<WeekSummary days={[]} />)
      expect(screen.getByText('Needs Attention')).toBeInTheDocument()
    })

    it('shows open tasks sorted by attention score', async () => {
      mockGetAttentionScores.mockResolvedValue({
        1: { Score: 20, Indicators: [], DaysOld: 0 },
        2: { Score: 80, Indicators: ['priority'], DaysOld: 4 },
      })

      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'task', content: 'New task' }),
            createEntry({
              id: 2,
              type: 'task',
              content: 'Old urgent task',
              priority: 'high',
            }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      await waitFor(() => {
        const items = screen.getAllByTestId('attention-item')
        expect(items[0]).toHaveTextContent('Old urgent task')
      })
    })

    it('shows unanswered questions', async () => {
      mockGetAttentionScores.mockResolvedValue({
        1: { Score: 30, Indicators: [], DaysOld: 0 },
      })

      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'question', content: 'What is the deadline?' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      await waitFor(() => {
        expect(screen.getByText('What is the deadline?')).toBeInTheDocument()
      })
    })

    it('shows attention indicators', async () => {
      mockGetAttentionScores.mockResolvedValue({
        1: { Score: 50, Indicators: ['priority'], DaysOld: 0 },
      })

      const days = [
        createDay({
          entries: [
            createEntry({
              id: 1,
              type: 'task',
              content: 'High priority task',
              priority: 'high',
            }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      await waitFor(() => {
        expect(screen.getByText(/priority/i)).toBeInTheDocument()
      })
    })

    it('limits to top 5 items', async () => {
      const scores: Record<number, { Score: number; Indicators: string[]; DaysOld: number }> = {}
      Array.from({ length: 10 }, (_, i) => {
        scores[i + 1] = { Score: 50 + i, Indicators: [], DaysOld: i }
      })
      mockGetAttentionScores.mockResolvedValue(scores)

      const days = [
        createDay({
          entries: Array.from({ length: 10 }, (_, i) =>
            createEntry({
              id: i + 1,
              type: 'task',
              content: `Task ${i + 1}`,
            })
          ),
        }),
      ]
      render(<WeekSummary days={days} />)

      await waitFor(() => {
        const items = screen.getAllByTestId('attention-item')
        expect(items.length).toBe(5)
      })
    })

    it('shows "Show all" link when more than 5 items', async () => {
      const scores: Record<number, { Score: number; Indicators: string[]; DaysOld: number }> = {}
      Array.from({ length: 10 }, (_, i) => {
        scores[i + 1] = { Score: 50 + i, Indicators: [], DaysOld: i }
      })
      mockGetAttentionScores.mockResolvedValue(scores)

      const days = [
        createDay({
          entries: Array.from({ length: 10 }, (_, i) =>
            createEntry({
              id: i + 1,
              type: 'task',
              content: `Task ${i + 1}`,
            })
          ),
        }),
      ]
      render(<WeekSummary days={days} />)

      await waitFor(() => {
        expect(screen.getByText(/show all/i)).toBeInTheDocument()
      })
    })
  })

  describe('WeekSummary without popover', () => {
    it('does not render entry context popover wrapper', () => {
      mockGetAttentionScores.mockResolvedValue({
        2: { Score: 10, Indicators: [], DaysOld: 0 },
      })

      const days: DayEntries[] = [{
        date: '2026-01-25',
        entries: [
          { id: 1, content: 'Event', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-25' },
          { id: 2, content: 'Child', type: 'task', priority: 'none', parentId: 1, loggedDate: '2026-01-25' },
        ],
      }]
      render(<WeekSummary days={days} />)
      expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
    })
  })
})
