import { format } from 'date-fns'
import { service, domain, wails } from '../wailsjs/go/models'
import { DayEntries, Entry, Habit, BujoList, Goal, EntryType, Priority } from '@/types/bujo'

export function transformEntry(e: domain.Entry): Entry {
  const loggedDate = e.ScheduledDate
    ? String(e.ScheduledDate)
    : e.CreatedAt
      ? String(e.CreatedAt)
      : new Date().toISOString()
  const scheduledDate = e.ScheduledDate ? String(e.ScheduledDate).split('T')[0] : undefined
  return {
    id: e.ID,
    content: e.Content,
    type: e.Type.toLowerCase() as EntryType,
    priority: (e.Priority?.toLowerCase() || 'none') as Priority,
    parentId: e.ParentID ?? null,
    loggedDate,
    scheduledDate,
    migrationCount: e.MigrationCount,
    tags: e.Tags || [],
  }
}

export function transformDayEntries(d: service.DayEntries): DayEntries {
  const dateStr = d.Date ? String(d.Date).split('T')[0] : format(new Date(), 'yyyy-MM-dd')
  return {
    date: dateStr,
    location: d.Location,
    mood: d.Mood,
    weather: d.Weather,
    entries: (d.Entries || []).map(transformEntry),
  }
}

export function transformHabit(h: service.HabitStatus): Habit {
  return {
    id: h.ID,
    name: h.Name,
    streak: h.CurrentStreak,
    completionRate: Math.round(h.CompletionPercent * 10) / 10,
    goal: h.GoalPerDay,
    goalPerWeek: h.GoalPerWeek > 0 ? h.GoalPerWeek : undefined,
    goalPerMonth: h.GoalPerMonth > 0 ? h.GoalPerMonth : undefined,
    weeklyProgress: h.WeeklyProgress > 0 ? Math.round(h.WeeklyProgress * 10) / 10 : undefined,
    monthlyProgress: h.MonthlyProgress > 0 ? Math.round(h.MonthlyProgress * 10) / 10 : undefined,
    dayHistory: (h.DayHistory || []).map(d => ({
      date: d.Date ? String(d.Date).split('T')[0] : '',
      completed: d.Completed,
      count: d.Count,
    })),
    todayLogged: h.TodayCount > 0,
    todayCount: h.TodayCount,
  }
}

export function transformList(l: wails.ListWithItems): BujoList {
  const items = (l.Items || []).map(item => ({
    id: item.RowID,
    content: item.Content,
    type: item.Type.toLowerCase() as EntryType,
    done: item.Type.toLowerCase() === 'done',
  }))
  return {
    id: l.ID,
    name: l.Name,
    items,
    doneCount: items.filter(i => i.done).length,
    totalCount: items.length,
  }
}

export function transformGoal(g: domain.Goal): Goal {
  const monthStr = g.Month ? String(g.Month).slice(0, 7) : format(new Date(), 'yyyy-MM')
  return {
    id: g.ID,
    content: g.Content,
    month: monthStr,
    status: (g.Status?.toLowerCase() || 'active') as Goal['status'],
    migratedTo: g.MigratedTo ? String(g.MigratedTo).slice(0, 7) : undefined,
  }
}
