import { useEffect, useState, useCallback, useRef } from 'react'
import { ChevronLeft, ChevronRight, PenLine, Plus } from 'lucide-react'
import { GetAgenda, GetHabits, GetLists, GetGoals, GetOutstandingQuestions, AddEntry, AddChildEntry, MarkEntryDone, MarkEntryUndone, EditEntry, DeleteEntry, HasChildren, MigrateEntry } from './wailsjs/go/wails/App'
import { Sidebar, ViewType } from '@/components/bujo/Sidebar'
import { DayView } from '@/components/bujo/DayView'
import { HabitTracker } from '@/components/bujo/HabitTracker'
import { ListsView } from '@/components/bujo/ListsView'
import { GoalsView } from '@/components/bujo/GoalsView'
import { OverviewView } from '@/components/bujo/OverviewView'
import { QuestionsView } from '@/components/bujo/QuestionsView'
import { SearchView } from '@/components/bujo/SearchView'
import { StatsView } from '@/components/bujo/StatsView'
import { SettingsView } from '@/components/bujo/SettingsView'
import { Header } from '@/components/bujo/Header'
import { CaptureModal } from '@/components/bujo/CaptureModal'
import { InlineEntryInput } from '@/components/bujo/InlineEntryInput'
import { KeyboardShortcuts } from '@/components/bujo/KeyboardShortcuts'
import { EditEntryModal } from '@/components/bujo/EditEntryModal'
import { ConfirmDialog } from '@/components/bujo/ConfirmDialog'
import { MigrateModal } from '@/components/bujo/MigrateModal'
import { AnswerQuestionModal } from '@/components/bujo/AnswerQuestionModal'
import { QuickStats } from '@/components/bujo/QuickStats'
import { DayEntries, Habit, BujoList, Goal, Entry } from '@/types/bujo'
import { transformDayEntries, transformEntry, transformHabit, transformList, transformGoal } from '@/lib/transforms'
import { startOfDay } from '@/lib/utils'
import { toWailsTime } from '@/lib/wailsTime'
import './index.css'

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
  const [overdueEntries, setOverdueEntries] = useState<Entry[]>([])
  const [overdueCount, setOverdueCount] = useState(0)
  const [outstandingQuestions, setOutstandingQuestions] = useState<Entry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [editModalEntry, setEditModalEntry] = useState<Entry | null>(null)
  const [deleteDialogEntry, setDeleteDialogEntry] = useState<Entry | null>(null)
  const [deleteHasChildren, setDeleteHasChildren] = useState(false)
  const [migrateModalEntry, setMigrateModalEntry] = useState<Entry | null>(null)
  const [answerModalEntry, setAnswerModalEntry] = useState<Entry | null>(null)
  const [currentDate, setCurrentDate] = useState(() => startOfDay(new Date()))
  const [habitDays, setHabitDays] = useState(14)
  const [habitPeriod, setHabitPeriod] = useState<'week' | 'month' | 'quarter'>('week')
  const [habitAnchorDate, setHabitAnchorDate] = useState(() => new Date())
  const [reviewAnchorDate, setReviewAnchorDate] = useState(() => startOfDay(new Date()))
  const [reviewDays, setReviewDays] = useState<DayEntries[]>([])
  const [showKeyboardShortcuts, setShowKeyboardShortcuts] = useState(false)
  const [showCaptureModal, setShowCaptureModal] = useState(false)
  const [inlineInputMode, setInlineInputMode] = useState<'root' | 'sibling' | 'child' | null>(null)
  const initialLoadCompleteRef = useRef(false)

  const loadData = useCallback(async () => {
    // Only show loading spinner on initial load, not on refresh
    if (!initialLoadCompleteRef.current) {
      setLoading(true)
    }
    setError(null)
    try {
      const now = new Date()
      const weekLater = new Date(currentDate.getTime() + 7 * 24 * 60 * 60 * 1000)
      const monthStart = new Date(now.getFullYear(), now.getMonth(), 1)

      // Review view: past 7 days ending at reviewAnchorDate
      const reviewEnd = new Date(reviewAnchorDate.getTime() + 24 * 60 * 60 * 1000) // Include anchor date
      const reviewStart = new Date(reviewAnchorDate.getTime() - 6 * 24 * 60 * 60 * 1000)

      const [agendaData, reviewData, habitsData, listsData, goalsData, questionsData] = await Promise.all([
        GetAgenda(toWailsTime(currentDate), toWailsTime(weekLater)),
        GetAgenda(toWailsTime(reviewStart), toWailsTime(reviewEnd)),
        GetHabits(habitDays),
        GetLists(),
        GetGoals(toWailsTime(monthStart)),
        GetOutstandingQuestions(),
      ])

      const transformedDays = (agendaData?.Days || []).map(transformDayEntries)
      setDays(transformedDays)
      const transformedReviewDays = (reviewData?.Days || []).map(transformDayEntries)
      setReviewDays(transformedReviewDays)
      const transformedOverdue = (agendaData?.Overdue || []).map(transformEntry)
      setOverdueEntries(transformedOverdue)
      setOverdueCount(transformedOverdue.length)
      setHabits((habitsData?.Habits || []).map(transformHabit))
      setLists((listsData || []).map(transformList))
      setGoals((goalsData || []).map(transformGoal))
      setOutstandingQuestions((questionsData || []).map(transformEntry))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data')
    } finally {
      setLoading(false)
      initialLoadCompleteRef.current = true
    }
  }, [currentDate, habitDays, reviewAnchorDate])

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

  const handleGoToToday = useCallback(() => {
    setCurrentDate(startOfDay(new Date()))
  }, [])

  const handlePrevWeek = useCallback(() => {
    setReviewAnchorDate(prev => {
      const newDate = new Date(prev)
      newDate.setDate(newDate.getDate() - 7)
      return newDate
    })
  }, [])

  const handleNextWeek = useCallback(() => {
    setReviewAnchorDate(prev => {
      const newDate = new Date(prev)
      newDate.setDate(newDate.getDate() + 7)
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
      const target = e.target as HTMLElement
      const isInputFocused = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA'

      // ? toggles keyboard shortcuts panel (works even when not in input)
      if (e.key === '?' && !isInputFocused) {
        e.preventDefault()
        setShowKeyboardShortcuts(prev => !prev)
        return
      }

      if (isInputFocused) return

      // Day navigation shortcuts (h/l/T) - always available in today view
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
        if (e.key === 'T') {
          e.preventDefault()
          handleGoToToday()
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

      // Entry creation shortcuts (c, r, a, A) - work in today view
      if (view === 'today') {
        if (e.key === 'c') {
          e.preventDefault()
          setShowCaptureModal(true)
          return
        }
        if (e.key === 'r') {
          e.preventDefault()
          setInlineInputMode('root')
          return
        }
        if (e.key === 'a') {
          e.preventDefault()
          // If selected entry is a question, open answer modal instead
          const selectedEntry = flatEntries[selectedIndex]
          if (selectedEntry?.type === 'question') {
            setAnswerModalEntry(selectedEntry)
          } else {
            setInlineInputMode('sibling')
          }
          return
        }
        if (e.key === 'A') {
          e.preventDefault()
          setInlineInputMode('child')
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
  }, [view, flatEntries, selectedIndex, loadData, handleDeleteEntryRequest, handlePrevDay, handleNextDay, handleGoToToday, cycleHabitPeriod])

  useEffect(() => {
    setSelectedIndex(0)
  }, [days])

  const handleViewChange = (newView: ViewType) => {
    setView(newView)
    setSelectedIndex(0)
  }

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

  const handleSelectEntry = useCallback((id: number) => {
    const index = flatEntries.findIndex(e => e.id === id)
    if (index !== -1) {
      setSelectedIndex(index)
    }
  }, [flatEntries])

  const handleAddChild = useCallback((entry: Entry) => {
    const index = flatEntries.findIndex(e => e.id === entry.id)
    if (index !== -1) {
      setSelectedIndex(index)
      setInlineInputMode('child')
    }
  }, [flatEntries])

  const handleInlineEntrySubmit = useCallback(async (content: string) => {
    const today = startOfDay(new Date())
    try {
      const selectedEntry = flatEntries[selectedIndex]

      if (inlineInputMode === 'child' && selectedEntry) {
        await AddChildEntry(selectedEntry.id, content, toWailsTime(today))
      } else {
        await AddEntry(content, toWailsTime(today))
      }

      setInlineInputMode(null)
      loadData()
    } catch (err) {
      console.error('Failed to add entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to add entry')
    }
  }, [flatEntries, selectedIndex, inlineInputMode, loadData])

  const handleInlineEntryCancel = useCallback(() => {
    setInlineInputMode(null)
  }, [])

  const viewTitles: Record<ViewType, string> = {
    today: 'Journal',
    week: 'Weekly Review',
    overview: 'Pending Tasks',
    questions: 'Open Questions',
    habits: 'Habit Tracker',
    lists: 'Lists',
    goals: 'Monthly Goals',
    search: 'Search',
    stats: 'Insights',
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
  const isViewingToday = currentDate.toDateString() === new Date().toDateString()

  return (
    <div className="flex h-screen bg-background">
      <Sidebar currentView={view} onViewChange={handleViewChange} />

      <div className="flex-1 flex flex-col overflow-hidden">
        <Header
          title={viewTitles[view]}
          currentMood={today?.mood}
          currentWeather={today?.weather}
          currentLocation={today?.location}
          onMoodChanged={loadData}
          onWeatherChanged={loadData}
          onLocationChanged={loadData}
        />

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
                {!isViewingToday && (
                  <button
                    onClick={handleGoToToday}
                    aria-label="Go to today"
                    className="px-3 py-2 text-sm rounded-lg bg-primary/10 hover:bg-primary/20 text-primary transition-colors"
                  >
                    Today
                  </button>
                )}
                <button
                  onClick={() => setShowCaptureModal(true)}
                  title="Open capture modal"
                  className="p-2 rounded-lg bg-primary/10 hover:bg-primary/20 text-primary transition-colors"
                >
                  <PenLine className="w-5 h-5" />
                </button>
              </div>
              <QuickStats days={days} habits={habits} goals={goals} overdueCount={overdueCount} />
              <DayView
                day={today}
                selectedEntryId={selectedEntryId}
                onEntryChanged={loadData}
                onSelectEntry={handleSelectEntry}
                onEditEntry={(entry) => setEditModalEntry(entry)}
                onDeleteEntry={handleDeleteEntryRequest}
                onMigrateEntry={(entry) => setMigrateModalEntry(entry)}
                onAddChild={handleAddChild}
                onAnswerEntry={(entry) => setAnswerModalEntry(entry)}
              />
              {inlineInputMode ? (
                <InlineEntryInput
                  mode={inlineInputMode}
                  onSubmit={handleInlineEntrySubmit}
                  onCancel={handleInlineEntryCancel}
                />
              ) : (
                <button
                  onClick={() => setInlineInputMode('root')}
                  title="Add new entry"
                  className="w-full flex items-center justify-center gap-2 py-3 text-muted-foreground hover:text-foreground hover:bg-secondary/50 rounded-lg transition-colors"
                >
                  <Plus className="w-4 h-4" />
                  <span className="text-sm">Add entry</span>
                </button>
              )}
            </div>
          )}

          {view === 'week' && (
            <div className="max-w-4xl mx-auto space-y-8">
              {/* Week Navigation */}
              <div className="flex items-center justify-center gap-4">
                <button
                  onClick={handlePrevWeek}
                  title="Previous week"
                  className="p-2 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <span className="text-sm text-muted-foreground">
                  {reviewDays.length > 0 && reviewDays[0]?.date} - {reviewDays.length > 0 && reviewDays[reviewDays.length - 1]?.date}
                </span>
                <button
                  onClick={handleNextWeek}
                  title="Next week"
                  disabled={reviewAnchorDate >= startOfDay(new Date())}
                  className="p-2 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
              </div>
              {reviewDays.map((day, i) => (
                <DayView
                  key={i}
                  day={day}
                  onEntryChanged={loadData}
                  onEditEntry={(entry) => setEditModalEntry(entry)}
                  onDeleteEntry={handleDeleteEntryRequest}
                  onMigrateEntry={(entry) => setMigrateModalEntry(entry)}
                  onAddChild={handleAddChild}
                  onAnswerEntry={(entry) => setAnswerModalEntry(entry)}
                />
              ))}
              {reviewDays.length === 0 && (
                <p className="text-muted-foreground text-center py-8">No entries this week</p>
              )}
            </div>
          )}

          {view === 'overview' && (
            <div className="max-w-3xl mx-auto">
              <OverviewView
                overdueEntries={overdueEntries}
                onEntryChanged={loadData}
                onError={setError}
              />
            </div>
          )}

          {view === 'questions' && (
            <div className="max-w-3xl mx-auto">
              <QuestionsView
                questions={outstandingQuestions}
                onEntryChanged={loadData}
                onError={setError}
              />
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

        {/* Keyboard shortcuts hint - toggle with ? */}
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

      {/* Answer Question Modal */}
      {answerModalEntry && (
        <AnswerQuestionModal
          isOpen={answerModalEntry !== null}
          questionId={answerModalEntry.id}
          questionContent={answerModalEntry.content}
          onClose={() => setAnswerModalEntry(null)}
          onAnswered={() => {
            setAnswerModalEntry(null)
            loadData()
          }}
        />
      )}

      {/* Capture Modal */}
      <CaptureModal
        isOpen={showCaptureModal}
        onClose={() => setShowCaptureModal(false)}
        onEntriesCreated={() => {
          setShowCaptureModal(false)
          loadData()
        }}
      />
    </div>
  )
}

export default App
