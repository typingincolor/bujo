import { useEffect, useState, useCallback } from 'react'
import { GetAgenda, GetHabits, GetLists, GetGoals } from './wailsjs/go/wails/App'
import { service, domain, wails } from './wailsjs/go/models'
import { Sidebar, ViewType } from '@/components/bujo/Sidebar'
import { DayView } from '@/components/bujo/DayView'
import { HabitTracker } from '@/components/bujo/HabitTracker'
import { ListsView } from '@/components/bujo/ListsView'
import { GoalsView } from '@/components/bujo/GoalsView'
import { Header } from '@/components/bujo/Header'
import { AddEntryBar } from '@/components/bujo/AddEntryBar'
import { KeyboardShortcuts } from '@/components/bujo/KeyboardShortcuts'
import { DayEntries, Entry, Habit, BujoList, Goal, EntryType, Priority } from '@/types/bujo'
import { format } from 'date-fns'
import './index.css'

function transformEntry(e: domain.Entry): Entry {
  return {
    id: e.ID,
    content: e.Content,
    type: e.Type.toLowerCase() as EntryType,
    priority: (e.Priority?.toLowerCase() || 'none') as Priority,
    parentId: e.ParentID ?? null,
    loggedDate: e.CreatedAt ? String(e.CreatedAt) : new Date().toISOString(),
  }
}

function transformDayEntries(d: service.DayEntries): DayEntries {
  const dateStr = d.Date ? String(d.Date).split('T')[0] : format(new Date(), 'yyyy-MM-dd')
  return {
    date: dateStr,
    location: d.Location,
    mood: d.Mood,
    weather: d.Weather,
    entries: (d.Entries || []).map(transformEntry),
  }
}

function transformHabit(h: service.HabitStatus): Habit {
  return {
    id: h.ID,
    name: h.Name,
    streak: h.CurrentStreak,
    completionRate: h.CompletionPercent,
    goal: h.GoalPerDay,
    history: (h.DayHistory || []).map(d => d.Completed),
    todayLogged: h.TodayCount > 0,
  }
}

function transformList(l: wails.ListWithItems): BujoList {
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

function transformGoal(g: domain.Goal): Goal {
  const monthStr = g.Month ? String(g.Month).slice(0, 7) : format(new Date(), 'yyyy-MM')
  return {
    id: g.ID,
    content: g.Content,
    month: monthStr,
    completed: g.Status === 'done',
  }
}

function App() {
  const [view, setView] = useState<ViewType>('today')
  const [days, setDays] = useState<DayEntries[]>([])
  const [habits, setHabits] = useState<Habit[]>([])
  const [lists, setLists] = useState<BujoList[]>([])
  const [goals, setGoals] = useState<Goal[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const now = new Date()
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      const weekLater = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000)
      const monthStart = new Date(now.getFullYear(), now.getMonth(), 1)

      const [agendaData, habitsData, listsData, goalsData] = await Promise.all([
        GetAgenda(today.toISOString() as unknown as Date, weekLater.toISOString() as unknown as Date),
        GetHabits(30),
        GetLists(),
        GetGoals(monthStart.toISOString() as unknown as Date),
      ])

      const transformedDays = (agendaData?.Days || []).map(transformDayEntries)
      setDays(transformedDays)
      setHabits((habitsData?.Habits || []).map(transformHabit))
      setLists((listsData || []).map(transformList))
      setGoals((goalsData || []).map(transformGoal))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handleViewChange = (newView: ViewType) => {
    setView(newView)
  }

  const viewTitles: Record<ViewType, string> = {
    today: 'Today',
    week: 'This Week',
    habits: 'Habits',
    lists: 'Lists',
    goals: 'Goals',
  }

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4" />
          <p className="text-muted-foreground">Loading your journal...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="text-center space-y-4">
          <h1 className="text-2xl font-display text-destructive">Error</h1>
          <p className="text-muted-foreground">{error}</p>
          <button
            onClick={loadData}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  const today = days[0]
  const weekDays = days.slice(0, 7)

  return (
    <div className="flex h-screen bg-background">
      <Sidebar currentView={view} onViewChange={handleViewChange} />

      <div className="flex-1 flex flex-col overflow-hidden">
        <Header title={viewTitles[view]} />

        <main className="flex-1 overflow-y-auto p-6">
          {view === 'today' && today && (
            <div className="max-w-3xl mx-auto space-y-6">
              <AddEntryBar />
              <DayView day={today} />
            </div>
          )}

          {view === 'week' && (
            <div className="max-w-4xl mx-auto space-y-8">
              {weekDays.map((day, i) => (
                <DayView key={i} day={day} />
              ))}
              {weekDays.length === 0 && (
                <p className="text-muted-foreground text-center py-8">No entries this week</p>
              )}
            </div>
          )}

          {view === 'habits' && (
            <div className="max-w-4xl mx-auto">
              <HabitTracker habits={habits} />
            </div>
          )}

          {view === 'lists' && (
            <div className="max-w-4xl mx-auto">
              <ListsView lists={lists} />
            </div>
          )}

          {view === 'goals' && (
            <div className="max-w-3xl mx-auto">
              <GoalsView goals={goals} />
            </div>
          )}
        </main>

        {/* Keyboard shortcuts hint */}
        <div className="hidden lg:block fixed bottom-4 right-4 w-72">
          <KeyboardShortcuts />
        </div>
      </div>
    </div>
  )
}

export default App
