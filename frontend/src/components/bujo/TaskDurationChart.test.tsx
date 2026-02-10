import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { TaskDurationChart } from './TaskDurationChart'
import { DayEntries } from '@/types/bujo'

vi.mock('recharts', async () => {
  const OriginalModule = await vi.importActual('recharts')
  return {
    ...OriginalModule,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="responsive-container" style={{ width: 400, height: 200 }}>
        {children}
      </div>
    ),
  }
})

const createDoneEntry = (
  id: number,
  date: string,
  completedAt: string,
  originalCreatedAt?: string,
) => ({
  id,
  content: `Task ${id}`,
  type: 'done' as const,
  priority: 'none' as const,
  parentId: null,
  loggedDate: date,
  completedAt,
  originalCreatedAt,
})

describe('TaskDurationChart', () => {
  it('renders title', () => {
    render(<TaskDurationChart days={[]} />)
    expect(screen.getByText(/time to complete/i)).toBeInTheDocument()
  })

  it('renders empty state when no completed tasks have timestamps', () => {
    const days: DayEntries[] = [{
      date: '2026-02-10',
      entries: [{
        id: 1,
        content: 'Task without completedAt',
        type: 'done',
        priority: 'none',
        parentId: null,
        loggedDate: '2026-02-10',
      }],
    }]
    render(<TaskDurationChart days={days} />)
    expect(screen.getByText(/no data/i)).toBeInTheDocument()
  })

  it('renders a BarChart when completed tasks have timestamps', () => {
    const days: DayEntries[] = [{
      date: '2026-02-10',
      entries: [
        createDoneEntry(1, '2026-02-10', '2026-02-10T18:00:00Z', '2026-02-07T09:00:00Z'),
      ],
    }]
    render(<TaskDurationChart days={days} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.queryByText(/no data/i)).not.toBeInTheDocument()
  })

  it('calculates average days from originalCreatedAt to completedAt', () => {
    const days: DayEntries[] = [{
      date: '2026-02-10',
      entries: [
        createDoneEntry(1, '2026-02-10', '2026-02-10T12:00:00Z', '2026-02-08T12:00:00Z'),
        createDoneEntry(2, '2026-02-10', '2026-02-10T12:00:00Z', '2026-02-06T12:00:00Z'),
      ],
    }]
    render(<TaskDurationChart days={days} />)
    expect(screen.getByText(/avg.*3\.0/i)).toBeInTheDocument()
  })

  it('falls back to loggedDate when originalCreatedAt is missing', () => {
    const days: DayEntries[] = [{
      date: '2026-02-10',
      entries: [
        createDoneEntry(1, '2026-02-09', '2026-02-10T12:00:00Z'),
      ],
    }]
    render(<TaskDurationChart days={days} />)
    expect(screen.getByText(/avg.*1\.0/i)).toBeInTheDocument()
  })

  it('ignores entries without completedAt', () => {
    const days: DayEntries[] = [{
      date: '2026-02-10',
      entries: [
        {
          id: 1,
          content: 'No timestamp',
          type: 'done',
          priority: 'none',
          parentId: null,
          loggedDate: '2026-02-10',
        },
        createDoneEntry(2, '2026-02-10', '2026-02-10T12:00:00Z', '2026-02-08T12:00:00Z'),
      ],
    }]
    render(<TaskDurationChart days={days} />)
    expect(screen.getByText(/avg.*2\.0/i)).toBeInTheDocument()
  })

  it('groups by completion week', () => {
    const days: DayEntries[] = [
      {
        date: '2026-02-03',
        entries: [
          createDoneEntry(1, '2026-02-03', '2026-02-03T12:00:00Z', '2026-02-01T12:00:00Z'),
        ],
      },
      {
        date: '2026-02-10',
        entries: [
          createDoneEntry(2, '2026-02-10', '2026-02-10T12:00:00Z', '2026-02-05T12:00:00Z'),
        ],
      },
    ]
    render(<TaskDurationChart days={days} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
  })
})
