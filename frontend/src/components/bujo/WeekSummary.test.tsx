import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { WeekSummary } from './WeekSummary'
import { DayEntries, Entry } from '@/types/bujo'

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

    it('shows events with children', () => {
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

  describe('Needs Attention', () => {
    it('renders needs attention section', () => {
      render(<WeekSummary days={[]} />)
      expect(screen.getByText('Needs Attention')).toBeInTheDocument()
    })

    it('shows open tasks sorted by attention score', () => {
      const fourDaysAgo = new Date()
      fourDaysAgo.setDate(fourDaysAgo.getDate() - 4)

      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'task', content: 'New task' }),
            createEntry({
              id: 2,
              type: 'task',
              content: 'Old urgent task',
              priority: 'high',
              loggedDate: fourDaysAgo.toISOString(),
            }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      // Old urgent task should appear first (higher attention score)
      const items = screen.getAllByTestId('attention-item')
      expect(items[0]).toHaveTextContent('Old urgent task')
    })

    it('shows unanswered questions', () => {
      const days = [
        createDay({
          entries: [
            createEntry({ id: 1, type: 'question', content: 'What is the deadline?' }),
          ],
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText('What is the deadline?')).toBeInTheDocument()
    })

    it('shows attention indicators', () => {
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

      expect(screen.getByText(/priority/i)).toBeInTheDocument()
    })

    it('limits to top 5 items', () => {
      const days = [
        createDay({
          entries: Array.from({ length: 10 }, (_, i) =>
            createEntry({ id: i + 1, type: 'task', content: `Task ${i + 1}` })
          ),
        }),
      ]
      render(<WeekSummary days={days} />)

      const items = screen.getAllByTestId('attention-item')
      expect(items.length).toBe(5)
    })

    it('shows "Show all" link when more than 5 items', () => {
      const days = [
        createDay({
          entries: Array.from({ length: 10 }, (_, i) =>
            createEntry({ id: i + 1, type: 'task', content: `Task ${i + 1}` })
          ),
        }),
      ]
      render(<WeekSummary days={days} />)

      expect(screen.getByText(/show all/i)).toBeInTheDocument()
    })
  })
})
