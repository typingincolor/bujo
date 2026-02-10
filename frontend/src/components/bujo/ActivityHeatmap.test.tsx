import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ActivityHeatmap } from './ActivityHeatmap'
import { DayEntries } from '@/types/bujo'

const createTestDay = (date: string, entryCount: number): DayEntries => ({
  date,
  entries: Array.from({ length: entryCount }, (_, i) => ({
    id: i + 1,
    content: `Entry ${i + 1}`,
    type: 'task' as const,
    priority: 'none' as const,
    parentId: null,
    loggedDate: date,
  })),
})

describe('ActivityHeatmap', () => {
  it('renders title', () => {
    render(<ActivityHeatmap days={[]} />)
    expect(screen.getByText(/activity/i)).toBeInTheDocument()
  })

  it('renders an SVG element', () => {
    const { container } = render(<ActivityHeatmap days={[]} />)
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('renders day cells as rects', () => {
    const days = [
      createTestDay('2026-02-10', 3),
      createTestDay('2026-02-09', 1),
    ]
    const { container } = render(<ActivityHeatmap days={days} />)
    const rects = container.querySelectorAll('svg rect[data-date]')
    expect(rects.length).toBeGreaterThan(0)
  })

  it('renders cells for days with zero entries', () => {
    const { container } = render(<ActivityHeatmap days={[]} />)
    const rects = container.querySelectorAll('svg rect[data-date]')
    expect(rects.length).toBeGreaterThan(0)
  })

  it('assigns higher opacity to days with more entries', () => {
    const days = [
      createTestDay('2026-02-10', 10),
      createTestDay('2026-02-09', 1),
    ]
    const { container } = render(<ActivityHeatmap days={days} />)
    const highRect = container.querySelector('rect[data-date="2026-02-10"]')
    const lowRect = container.querySelector('rect[data-date="2026-02-09"]')
    expect(highRect).toBeInTheDocument()
    expect(lowRect).toBeInTheDocument()
    const highOpacity = Number(highRect?.getAttribute('opacity') || highRect?.style.opacity)
    const lowOpacity = Number(lowRect?.getAttribute('opacity') || lowRect?.style.opacity)
    expect(highOpacity).toBeGreaterThan(lowOpacity)
  })

  it('shows tooltip text for cells with entries', () => {
    const days = [createTestDay('2026-02-10', 5)]
    render(<ActivityHeatmap days={days} />)
    const title = screen.getByText(/5 entries/i)
    expect(title).toBeInTheDocument()
  })

  it('renders month labels', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-02-15'))
    render(<ActivityHeatmap days={[]} />)
    expect(screen.getByText('Dec')).toBeInTheDocument()
    expect(screen.getByText('Jan')).toBeInTheDocument()
    expect(screen.getByText('Feb')).toBeInTheDocument()
    vi.useRealTimers()
  })

  it('renders day-of-week labels', () => {
    render(<ActivityHeatmap days={[]} />)
    expect(screen.getByText('Mon')).toBeInTheDocument()
    expect(screen.getByText('Wed')).toBeInTheDocument()
    expect(screen.getByText('Fri')).toBeInTheDocument()
  })
})
