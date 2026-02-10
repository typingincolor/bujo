import { useMemo } from 'react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { DayEntries, Entry } from '@/types/bujo'

interface TaskDurationChartProps {
  days: DayEntries[]
}

const MS_PER_DAY = 1000 * 60 * 60 * 24

function flattenEntries(entries: Entry[]): Entry[] {
  const result: Entry[] = []
  function traverse(items: Entry[]) {
    for (const entry of items) {
      result.push(entry)
      if (entry.children && entry.children.length > 0) {
        traverse(entry.children)
      }
    }
  }
  traverse(entries)
  return result
}

function getWeekKey(date: string): string {
  const d = new Date(date + 'T00:00:00')
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1)
  const monday = new Date(d)
  monday.setDate(diff)
  const mm = String(monday.getMonth() + 1).padStart(2, '0')
  const dd = String(monday.getDate()).padStart(2, '0')
  return `${monday.getFullYear()}-${mm}-${dd}`
}

function formatWeekLabel(week: string): string {
  const d = new Date(week + 'T00:00:00')
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  return `${months[d.getMonth()]} ${d.getDate()}`
}

interface DurationEntry {
  completedAt: string
  durationDays: number
}

function toDateOnly(isoString: string): number {
  const dateStr = isoString.slice(0, 10)
  return new Date(dateStr + 'T00:00:00Z').getTime()
}

function computeDurations(days: DayEntries[]): DurationEntry[] {
  const results: DurationEntry[] = []

  for (const day of days) {
    const flat = flattenEntries(day.entries)
    for (const entry of flat) {
      if (entry.type !== 'done' || !entry.completedAt) continue

      const completedDate = toDateOnly(entry.completedAt)
      const createdDate = entry.originalCreatedAt
        ? toDateOnly(entry.originalCreatedAt)
        : new Date(entry.loggedDate + 'T00:00:00Z').getTime()

      const durationDays = (completedDate - createdDate) / MS_PER_DAY

      results.push({
        completedAt: entry.completedAt,
        durationDays,
      })
    }
  }

  return results
}

interface WeekDuration {
  week: string
  avgDays: number
}

function aggregateByWeek(durations: DurationEntry[]): WeekDuration[] {
  const weekMap = new Map<string, number[]>()

  for (const d of durations) {
    const dateStr = d.completedAt.slice(0, 10)
    const weekKey = getWeekKey(dateStr)
    const existing = weekMap.get(weekKey) || []
    existing.push(d.durationDays)
    weekMap.set(weekKey, existing)
  }

  return Array.from(weekMap.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([week, days]) => ({
      week,
      avgDays: Math.round((days.reduce((sum, d) => sum + d, 0) / days.length) * 10) / 10,
    }))
}

export function TaskDurationChart({ days }: TaskDurationChartProps) {
  const durations = useMemo(() => computeDurations(days), [days])
  const weeklyData = useMemo(() => aggregateByWeek(durations), [durations])

  const overallAvg = durations.length > 0
    ? Math.round((durations.reduce((sum, d) => sum + d.durationDays, 0) / durations.length) * 10) / 10
    : 0

  const hasData = durations.length > 0

  return (
    <div>
      <h3 className="text-sm font-medium text-muted-foreground mb-2">Time to Complete</h3>
      {!hasData ? (
        <p className="text-sm text-muted-foreground">No data yet</p>
      ) : (
        <>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={weeklyData}>
              <XAxis dataKey="week" tickFormatter={formatWeekLabel} fontSize={10} />
              <YAxis fontSize={10} />
              <Tooltip
                labelFormatter={formatWeekLabel}
                contentStyle={{ fontSize: 12 }}
              />
              <Bar
                dataKey="avgDays"
                fill="hsl(var(--primary))"
                fillOpacity={0.7}
                name="Avg Days"
              />
            </BarChart>
          </ResponsiveContainer>
          <p className="text-xs text-muted-foreground mt-2">
            Avg: {overallAvg.toFixed(1)} days
          </p>
        </>
      )}
    </div>
  )
}
