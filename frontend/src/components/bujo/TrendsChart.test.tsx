import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { TrendsChart } from './TrendsChart'
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

const createTestDay = (date: string, tasks: number, done: number, notes: number = 0): DayEntries => ({
  date,
  entries: [
    ...Array.from({ length: done }, (_, i) => ({
      id: i + 1,
      content: `Done ${i + 1}`,
      type: 'done' as const,
      priority: 'none' as const,
      parentId: null,
      loggedDate: date,
    })),
    ...Array.from({ length: tasks }, (_, i) => ({
      id: done + i + 1,
      content: `Task ${i + 1}`,
      type: 'task' as const,
      priority: 'none' as const,
      parentId: null,
      loggedDate: date,
    })),
    ...Array.from({ length: notes }, (_, i) => ({
      id: done + tasks + i + 1,
      content: `Note ${i + 1}`,
      type: 'note' as const,
      priority: 'none' as const,
      parentId: null,
      loggedDate: date,
    })),
  ],
})

describe('TrendsChart', () => {
  it('renders title', () => {
    render(<TrendsChart days={[]} />)
    expect(screen.getByText(/trends/i)).toBeInTheDocument()
  })

  it('renders a Recharts AreaChart when data exists', () => {
    const days = [createTestDay('2026-02-10', 1, 1)]
    render(<TrendsChart days={days} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.queryByText(/no data/i)).not.toBeInTheDocument()
  })

  it('renders empty state when no data', () => {
    render(<TrendsChart days={[]} />)
    expect(screen.getByText(/no data/i)).toBeInTheDocument()
  })

  it('shows completion rate insight', () => {
    const days = [
      createTestDay('2026-02-09', 2, 3),
      createTestDay('2026-02-10', 1, 2),
    ]
    render(<TrendsChart days={days} />)
    expect(screen.getByText(/completion/i)).toBeInTheDocument()
    expect(screen.getByText(/62\.5%/)).toBeInTheDocument()
  })

  it('aggregates entries by week', () => {
    const days = [
      createTestDay('2026-02-02', 2, 1),
      createTestDay('2026-02-03', 1, 2),
      createTestDay('2026-02-09', 3, 0),
      createTestDay('2026-02-10', 0, 4),
    ]
    render(<TrendsChart days={days} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.queryByText(/no data/i)).not.toBeInTheDocument()
  })

  it('counts tasks and done entries separately', () => {
    const days = [
      createTestDay('2026-02-10', 3, 2, 5),
    ]
    render(<TrendsChart days={days} />)
    expect(screen.getByText(/40%/)).toBeInTheDocument()
  })

  it('ignores non-task entry types in completion rate', () => {
    const days = [
      createTestDay('2026-02-10', 0, 0, 10),
    ]
    render(<TrendsChart days={days} />)
    expect(screen.getByText(/no data/i)).toBeInTheDocument()
  })
})
