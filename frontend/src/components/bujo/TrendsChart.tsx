import { useMemo } from 'react'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { DayEntries } from '@/types/bujo'
import { flattenEntries, getWeekKey, formatWeekLabel } from '@/lib/chart-utils'

interface TrendsChartProps {
  days: DayEntries[]
}

interface WeekData {
  week: string
  created: number
  completed: number
}

function aggregateByWeek(days: DayEntries[]): WeekData[] {
  const weekMap = new Map<string, { created: number; completed: number }>()

  for (const day of days) {
    const weekKey = getWeekKey(day.date)
    const flat = flattenEntries(day.entries)
    const taskCount = flat.filter(e => e.type === 'task' || e.type === 'done').length
    const doneCount = flat.filter(e => e.type === 'done').length

    const existing = weekMap.get(weekKey) || { created: 0, completed: 0 }
    existing.created += taskCount
    existing.completed += doneCount
    weekMap.set(weekKey, existing)
  }

  return Array.from(weekMap.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([week, data]) => ({
      week,
      created: data.created,
      completed: data.completed,
    }))
}

export function TrendsChart({ days }: TrendsChartProps) {
  const allEntries = useMemo(
    () => days.flatMap(day => flattenEntries(day.entries)),
    [days],
  )

  const totalTasks = allEntries.filter(e => e.type === 'task' || e.type === 'done').length
  const totalDone = allEntries.filter(e => e.type === 'done').length
  const completionRate = totalTasks > 0 ? Math.round((totalDone / totalTasks) * 1000) / 10 : 0
  const hasData = totalTasks > 0

  const weeklyData = useMemo(() => aggregateByWeek(days), [days])

  return (
    <div>
      <h3 className="text-sm font-medium text-muted-foreground mb-2">Trends</h3>
      {!hasData ? (
        <p className="text-sm text-muted-foreground">No data yet</p>
      ) : (
        <>
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={weeklyData}>
              <XAxis dataKey="week" tickFormatter={formatWeekLabel} fontSize={10} />
              <YAxis fontSize={10} />
              <Tooltip
                labelFormatter={formatWeekLabel}
                contentStyle={{ fontSize: 12 }}
              />
              <Area
                type="monotone"
                dataKey="created"
                stroke="hsl(var(--primary))"
                fill="hsl(var(--primary))"
                fillOpacity={0.15}
                name="Tasks Created"
              />
              <Area
                type="monotone"
                dataKey="completed"
                stroke="hsl(var(--bujo-done))"
                fill="hsl(var(--bujo-done))"
                fillOpacity={0.15}
                name="Completed"
              />
            </AreaChart>
          </ResponsiveContainer>
          <p className="text-xs text-muted-foreground mt-2">
            Completion rate: {completionRate}%
          </p>
        </>
      )}
    </div>
  )
}
