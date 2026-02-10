import { useMemo } from 'react'
import { DayEntries } from '@/types/bujo'

interface ActivityHeatmapProps {
  days: DayEntries[]
}

const CELL_SIZE = 12
const CELL_GAP = 2
const CELL_STEP = CELL_SIZE + CELL_GAP
const WEEKS = 13
const DAYS_IN_WEEK = 7
const LABEL_WIDTH = 30
const HEADER_HEIGHT = 16
const MONTH_LABELS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
const DAY_LABELS: Record<number, string> = { 1: 'Mon', 3: 'Wed', 5: 'Fri' }

function buildDateGrid(now: Date) {
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const dayOfWeek = (today.getDay() + 6) % 7
  const totalDays = WEEKS * DAYS_IN_WEEK
  const startDate = new Date(today)
  startDate.setDate(startDate.getDate() - (totalDays - 1) + (DAYS_IN_WEEK - 1 - dayOfWeek))

  const grid: { date: string; col: number; row: number }[] = []
  for (let i = 0; i < totalDays; i++) {
    const d = new Date(startDate)
    d.setDate(d.getDate() + i)
    const col = Math.floor(i / DAYS_IN_WEEK)
    const row = i % DAYS_IN_WEEK
    const yyyy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    grid.push({ date: `${yyyy}-${mm}-${dd}`, col, row })
  }
  return { grid, startDate }
}

function getMonthLabels(startDate: Date) {
  const labels: { label: string; col: number }[] = []
  const seen = new Set<string>()
  for (let col = 0; col < WEEKS; col++) {
    const d = new Date(startDate)
    d.setDate(d.getDate() + col * DAYS_IN_WEEK)
    const key = `${d.getFullYear()}-${d.getMonth()}`
    if (!seen.has(key)) {
      seen.add(key)
      labels.push({ label: MONTH_LABELS[d.getMonth()], col })
    }
  }
  return labels
}

export function ActivityHeatmap({ days }: ActivityHeatmapProps) {
  const entryCountByDate = useMemo(() => {
    const map = new Map<string, number>()
    for (const day of days) {
      map.set(day.date, day.entries.length)
    }
    return map
  }, [days])

  const { grid, monthLabels, maxCount } = useMemo(() => {
    const now = new Date()
    const { grid, startDate } = buildDateGrid(now)
    const monthLabels = getMonthLabels(startDate)
    let maxCount = 0
    for (const cell of grid) {
      const count = entryCountByDate.get(cell.date) || 0
      if (count > maxCount) maxCount = count
    }
    return { grid, monthLabels, maxCount }
  }, [entryCountByDate])

  const svgWidth = LABEL_WIDTH + WEEKS * CELL_STEP
  const svgHeight = HEADER_HEIGHT + DAYS_IN_WEEK * CELL_STEP

  return (
    <div>
      <h3 className="text-sm font-medium text-muted-foreground mb-2">Activity</h3>
      <svg width={svgWidth} height={svgHeight} role="img">
        {monthLabels.map(({ label, col }) => (
          <text
            key={`month-${col}`}
            x={LABEL_WIDTH + col * CELL_STEP}
            y={HEADER_HEIGHT - 4}
            fontSize={10}
            fill="currentColor"
            className="text-muted-foreground"
          >
            {label}
          </text>
        ))}

        {Object.entries(DAY_LABELS).map(([row, label]) => (
          <text
            key={`day-${row}`}
            x={0}
            y={HEADER_HEIGHT + Number(row) * CELL_STEP + CELL_SIZE - 2}
            fontSize={10}
            fill="currentColor"
            className="text-muted-foreground"
          >
            {label}
          </text>
        ))}

        {grid.map(({ date, col, row }) => {
          const count = entryCountByDate.get(date) || 0
          const opacity = maxCount > 0 && count > 0 ? 0.2 + (count / maxCount) * 0.8 : 0.06
          return (
            <rect
              key={date}
              data-date={date}
              x={LABEL_WIDTH + col * CELL_STEP}
              y={HEADER_HEIGHT + row * CELL_STEP}
              width={CELL_SIZE}
              height={CELL_SIZE}
              rx={2}
              fill="currentColor"
              opacity={opacity}
              className="text-primary"
            >
              <title>{count > 0 ? `${count} entries` : 'No entries'}</title>
            </rect>
          )
        })}
      </svg>
    </div>
  )
}
