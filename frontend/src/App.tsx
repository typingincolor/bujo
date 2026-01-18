import { useEffect, useState, useCallback } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { GetAgenda, GetHabits, GetLists, GetGoals, AddEntry, MarkEntryDone, MarkEntryUndone, Search, EditEntry, DeleteEntry, HasChildren, MigrateEntry } from './wailsjs/go/wails/App'
import { time } from './wailsjs/go/models'
import { Sidebar, ViewType } from '@/components/bujo/Sidebar'
import { DayView } from '@/components/bujo/DayView'
import { HabitTracker } from '@/components/bujo/HabitTracker'
import { ListsView } from '@/components/bujo/ListsView'
import { GoalsView } from '@/components/bujo/GoalsView'
import { SearchView } from '@/components/bujo/SearchView'
import { StatsView } from '@/components/bujo/StatsView'
import { SettingsView } from '@/components/bujo/SettingsView'
import { Header } from '@/components/bujo/Header'
import { AddEntryBar } from '@/components/bujo/AddEntryBar'
import { KeyboardShortcuts } from '@/components/bujo/KeyboardShortcuts'
import { EditEntryModal } from '@/components/bujo/EditEntryModal'
import { ConfirmDialog } from '@/components/bujo/ConfirmDialog'
import { MigrateModal } from '@/components/bujo/MigrateModal'
import { QuickStats } from '@/components/bujo/QuickStats'
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
  const [editModalEntry, setEditModalEntry] = useState<Entry | null>(null)
  const [deleteDialogEntry, setDeleteDialogEntry] = useState<Entry | null>(null)
  const [deleteHasChildren, setDeleteHasChildren] = useState(false)
  const [migrateModalEntry, setMigrateModalEntry] = useState<Entry | null>(null)
  const [currentDate, setCurrentDate] = useState(() => startOfDay(new Date()))
  const [habitDays, setHabitDays] = useState(14)
  const [habitPeriod, setHabitPeriod] = useState<'week' | 'month' | 'quarter'>('week')
  const [habitAnchorDate, setHabitAnchorDate] = useState(() => new Date())
  const [showKeyboardShortcuts, setShowKeyboardShortcuts] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const now = new Date()
      const weekLater = new Date(currentDate.getTime() + 7 * 24 * 60 * 60 * 1000)
      const monthStart = new Date(now.getFullYear(), now.getMonth(), 1)

      const [agendaData, habitsData, listsData, goalsData] = await Promise.all([
        GetAgenda(toWailsTime(currentDate), toWailsTime(weekLater)),
        GetHabits(habitDays),
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
  }, [currentDate, habitDays])

  useEffect(() => {
    loadData()
  }, [loadData])

  const todayEntries = days[0]?.entries || []
  const flatEntries = flattenEntries(todayEntries)

  const handleDeleteEntryRequest = useCallback(async (entry: Entry) => {
    try {
      const hasChildren = await HasChildren(entry.id)
      setDeleteHasChildren(hasChildren)
      setDeleteDialogEntry(entry)
    } catch (err) {
      console.error('Failed to check entry children:', err)
      setError(err instanceof Error ? err.message : 'Failed to check entry')
    }
  }, [])

  const handlePrevDay = useCallback(() => {
    setCurrentDate(prev => {
      const newDate = new Date(prev)
      newDate.setDate(newDate.getDate() - 1)
      return newDate
    })
  }, [])

  const handleNextDay = useCallback(() => {
    setCurrentDate(prev => {
      const newDate = new Date(prev)
      newDate.setDate(newDate.getDate() + 1)
      return newDate
    })
  }, [])

  const handleDateChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const dateValue = e.target.value
    if (dateValue) {
      const newDate = new Date(dateValue + 'T00:00:00')
      setCurrentDate(newDate)
    }
  }, [])

  const handleHabitPeriodChange = useCallback((period: 'week' | 'month' | 'quarter') => {
    const daysMap = { week: 14, month: 45, quarter: 120 }
    setHabitDays(daysMap[period])
    setHabitPeriod(period)
  }, [])

  const handleHabitNavigate = useCallback((newAnchor: Date) => {
    setHabitAnchorDate(newAnchor)
  }, [])

  const cycleHabitPeriod = useCallback(() => {
    const periods: Array<'week' | 'month' | 'quarter'> = ['week', 'month', 'quarter']
    const currentIndex = periods.indexOf(habitPeriod)
    const nextPeriod = periods[(currentIndex + 1) % periods.length]
    handleHabitPeriodChange(nextPeriod)
  }, [habitPeriod, handleHabitPeriodChange])

  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      // Cmd+? (Cmd+Shift+/) toggles keyboard shortcuts panel
      if (e.key === '?' && e.metaKey) {
        e.preventDefault()
        setShowKeyboardShortcuts(prev => !prev)
        return
      }

      const target = e.target as HTMLElement
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return

      // Day navigation shortcuts (h/l) - always available in today view
      if (view === 'today') {
        if (e.key === 'h') {
          e.preventDefault()
          handlePrevDay()
          return
        }
        if (e.key === 'l') {
          e.preventDefault()
          handleNextDay()
          return
        }
      }

      // Habit view shortcuts
      if (view === 'habits') {
        if (e.key === 'w') {
          e.preventDefault()
          cycleHabitPeriod()
          return
        }
      }

      if (view !== 'today' || flatEntries.length === 0) return

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
        case 'e': {
          e.preventDefault()
          const entry = flatEntries[selectedIndex]
          if (entry) {
            setEditModalEntry(entry)
          }
          break
        }
        case 'd': {
          e.preventDefault()
          const entry = flatEntries[selectedIndex]
          if (entry) {
            handleDeleteEntryRequest(entry)
          }
          break
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [view, flatEntries, selectedIndex, loadData, handleDeleteEntryRequest, handlePrevDay, handleNextDay, cycleHabitPeriod])

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

  const handleEditEntry = useCallback(async (newContent: string) => {
    if (!editModalEntry) return
    try {
      await EditEntry(editModalEntry.id, newContent)
      setEditModalEntry(null)
      loadData()
    } catch (err) {
      console.error('Failed to edit entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to edit entry')
    }
  }, [editModalEntry, loadData])

  const handleDeleteEntry = useCallback(async () => {
    if (!deleteDialogEntry) return
    try {
      await DeleteEntry(deleteDialogEntry.id)
      setDeleteDialogEntry(null)
      loadData()
    } catch (err) {
      console.error('Failed to delete entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to delete entry')
    }
  }, [deleteDialogEntry, loadData])

  const handleMigrateEntry = useCallback(async (dateStr: string) => {
    if (!migrateModalEntry) return
    try {
      const migrateDate = new Date(dateStr + 'T00:00:00')
      await MigrateEntry(migrateModalEntry.id, toWailsTime(migrateDate))
      setMigrateModalEntry(null)
      loadData()
    } catch (err) {
      console.error('Failed to migrate entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to migrate entry')
    }
  }, [migrateModalEntry, loadData])

  const viewTitles: Record<ViewType, string> = {
    today: 'Today',
    week: 'This Week',
    habits: 'Habits',
    lists: 'Lists',
    goals: 'Goals',
    search: 'Search',
    stats: 'Statistics',
    settings: 'Settings',
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
              {/* Day Navigation */}
              <div className="flex items-center justify-center gap-4">
                <button
                  onClick={handlePrevDay}
                  aria-label="Previous day"
                  className="p-2 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <input
                  type="date"
                  aria-label="Pick date"
                  value={currentDate.toISOString().split('T')[0]}
                  onChange={handleDateChange}
                  className="px-3 py-2 rounded-lg bg-secondary/50 hover:bg-secondary text-sm transition-colors border-none focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
                <button
                  onClick={handleNextDay}
                  aria-label="Next day"
                  className="p-2 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
              </div>
              <QuickStats days={days} habits={habits} goals={goals} />
              <AddEntryBar onAdd={handleAddEntry} />
              <DayView
                day={today}
                selectedEntryId={selectedEntryId}
                onEntryChanged={loadData}
                onEditEntry={(entry) => setEditModalEntry(entry)}
                onDeleteEntry={handleDeleteEntryRequest}
                onMigrateEntry={(entry) => setMigrateModalEntry(entry)}
              />
            </div>
          )}

          {view === 'week' && (
            <div className="max-w-4xl mx-auto space-y-8">
              {weekDays.map((day, i) => (
                <DayView
                  key={i}
                  day={day}
                  onEntryChanged={loadData}
                  onEditEntry={(entry) => setEditModalEntry(entry)}
                  onDeleteEntry={handleDeleteEntryRequest}
                  onMigrateEntry={(entry) => setMigrateModalEntry(entry)}
                />
              ))}
              {weekDays.length === 0 && (
                <p className="text-muted-foreground text-center py-8">No entries this week</p>
              )}
            </div>
          )}

          {view === 'habits' && (
            <div className="max-w-4xl mx-auto">
              <HabitTracker
                habits={habits}
                onHabitChanged={loadData}
                period={habitPeriod}
                onPeriodChange={handleHabitPeriodChange}
                anchorDate={habitAnchorDate}
                onNavigate={handleHabitNavigate}
              />
            </div>
          )}

          {view === 'lists' && (
            <div className="max-w-4xl mx-auto">
              <ListsView lists={lists} onListChanged={loadData} />
            </div>
          )}

          {view === 'goals' && (
            <div className="max-w-3xl mx-auto">
              <GoalsView goals={goals} onGoalChanged={loadData} onError={setError} />
            </div>
          )}

          {view === 'search' && (
            <div className="max-w-3xl mx-auto">
              <SearchView />
            </div>
          )}

          {view === 'stats' && (
            <div className="max-w-4xl mx-auto">
              <StatsView days={days} habits={habits} goals={goals} />
            </div>
          )}

          {view === 'settings' && (
            <div className="max-w-2xl mx-auto">
              <SettingsView />
            </div>
          )}
        </main>

        {/* Keyboard shortcuts hint - toggle with Cmd+? */}
        {showKeyboardShortcuts && (
          <div className="fixed bottom-4 right-4 w-72 z-50">
            <KeyboardShortcuts view={view} />
          </div>
        )}
      </div>

      {/* Edit Entry Modal */}
      <EditEntryModal
        isOpen={editModalEntry !== null}
        initialContent={editModalEntry?.content || ''}
        onSave={handleEditEntry}
        onCancel={() => setEditModalEntry(null)}
      />

      {/* Delete Entry Confirmation */}
      <ConfirmDialog
        isOpen={deleteDialogEntry !== null}
        title="Delete Entry"
        message={deleteHasChildren
          ? `Are you sure you want to delete "${deleteDialogEntry?.content}"? This will also delete all child entries.`
          : `Are you sure you want to delete "${deleteDialogEntry?.content}"?`}
        confirmText="Delete"
        variant="destructive"
        onConfirm={handleDeleteEntry}
        onCancel={() => setDeleteDialogEntry(null)}
      />

      {/* Migrate Entry Modal */}
      <MigrateModal
        isOpen={migrateModalEntry !== null}
        entryContent={migrateModalEntry?.content || ''}
        onMigrate={handleMigrateEntry}
        onCancel={() => setMigrateModalEntry(null)}
      />
    </div>
  )
}

export default App
