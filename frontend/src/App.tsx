import { useEffect, useState, useCallback } from 'react'
import { GetAgenda, GetHabits, GetLists, GetGoals, AddEntry, MarkEntryDone, MarkEntryUndone, Search } from './wailsjs/go/wails/App'
import { time } from './wailsjs/go/models'
import { Sidebar, ViewType } from '@/components/bujo/Sidebar'
import { DayView } from '@/components/bujo/DayView'
import { HabitTracker } from '@/components/bujo/HabitTracker'
import { ListsView } from '@/components/bujo/ListsView'
import { GoalsView } from '@/components/bujo/GoalsView'
import { Header } from '@/components/bujo/Header'
import { AddEntryBar } from '@/components/bujo/AddEntryBar'
import { KeyboardShortcuts } from '@/components/bujo/KeyboardShortcuts'
import { DayEntries, Habit, BujoList, Goal, EntryType, ENTRY_SYMBOLS, Entry } from '@/types/bujo'
import { transformDayEntries, transformHabit, transformList, transformGoal } from '@/lib/transforms'
import { startOfDay } from '@/lib/utils'
import './index.css'

// Wails serializes Go time.Time as ISO strings over JSON.
// This helper provides type-safe conversion from Date to the expected binding type.
function toWailsTime(date: Date): time.Time {
  return date.toISOString() as unknown as time.Time
}

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

function App() {
  const [view, setView] = useState<ViewType>('today')
  const [days, setDays] = useState<DayEntries[]>([])
  const [habits, setHabits] = useState<Habit[]>([])
  const [lists, setLists] = useState<BujoList[]>([])
  const [goals, setGoals] = useState<Goal[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [searchResults, setSearchResults] = useState<Array<{ id: number; content: string; type: string; date: string }>>([])

  const loadData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const now = new Date()
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      const weekLater = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000)
      const monthStart = new Date(now.getFullYear(), now.getMonth(), 1)

      const [agendaData, habitsData, listsData, goalsData] = await Promise.all([
        GetAgenda(toWailsTime(today), toWailsTime(weekLater)),
        GetHabits(30),
        GetLists(),
        GetGoals(toWailsTime(monthStart)),
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

  const todayEntries = days[0]?.entries || []
  const flatEntries = flattenEntries(todayEntries)

  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      if (view !== 'today' || flatEntries.length === 0) return

      const target = e.target as HTMLElement
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return

      switch (e.key) {
        case 'j':
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex(prev => Math.min(prev + 1, flatEntries.length - 1))
          break
        case 'k':
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex(prev => Math.max(prev - 1, 0))
          break
        case ' ': {
          e.preventDefault()
          const entry = flatEntries[selectedIndex]
          if (entry && (entry.type === 'task' || entry.type === 'done')) {
            if (entry.type === 'done') {
              await MarkEntryUndone(entry.id)
            } else {
              await MarkEntryDone(entry.id)
            }
            loadData()
          }
          break
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [view, flatEntries, selectedIndex, loadData])

  useEffect(() => {
    setSelectedIndex(0)
  }, [days])

  const handleViewChange = (newView: ViewType) => {
    setView(newView)
    setSelectedIndex(0)
  }

  const handleSearch = useCallback(async (query: string) => {
    if (!query) {
      setSearchResults([])
      return
    }
    try {
      const results = await Search(query)
      setSearchResults((results || []).map(entry => ({
        id: entry.ID,
        content: entry.Content,
        type: entry.Type,
        date: (entry.CreatedAt as unknown as string)?.split('T')[0] || '',
      })))
    } catch (err) {
      console.error('Search failed:', err)
      setSearchResults([])
    }
  }, [])

  const handleAddEntry = useCallback(async (content: string, type: EntryType) => {
    const symbol = ENTRY_SYMBOLS[type]
    const formattedContent = `${symbol} ${content}`
    const today = startOfDay(new Date())
    try {
      await AddEntry(formattedContent, toWailsTime(today))
      loadData()
    } catch (err) {
      console.error('Failed to add entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to add entry')
    }
  }, [loadData])

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
  const selectedEntryId = flatEntries[selectedIndex]?.id ?? null
  const weekDays = days.slice(0, 7)

  return (
    <div className="flex h-screen bg-background">
      <Sidebar currentView={view} onViewChange={handleViewChange} />

      <div className="flex-1 flex flex-col overflow-hidden">
        <Header title={viewTitles[view]} searchResults={searchResults} onSearch={handleSearch} />

        <main className="flex-1 overflow-y-auto p-6">
          {view === 'today' && today && (
            <div className="max-w-3xl mx-auto space-y-6">
              <AddEntryBar onAdd={handleAddEntry} />
              <DayView day={today} selectedEntryId={selectedEntryId} onEntryChanged={loadData} />
            </div>
          )}

          {view === 'week' && (
            <div className="max-w-4xl mx-auto space-y-8">
              {weekDays.map((day, i) => (
                <DayView key={i} day={day} onEntryChanged={loadData} />
              ))}
              {weekDays.length === 0 && (
                <p className="text-muted-foreground text-center py-8">No entries this week</p>
              )}
            </div>
          )}

          {view === 'habits' && (
            <div className="max-w-4xl mx-auto">
              <HabitTracker habits={habits} onHabitChanged={loadData} />
            </div>
          )}

          {view === 'lists' && (
            <div className="max-w-4xl mx-auto">
              <ListsView lists={lists} onListChanged={loadData} />
            </div>
          )}

          {view === 'goals' && (
            <div className="max-w-3xl mx-auto">
              <GoalsView goals={goals} onGoalChanged={loadData} />
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
