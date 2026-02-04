import { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { useNavigationHistory } from '@/hooks/useNavigationHistory'
import { useSettings } from '@/contexts/SettingsContext'
import { EventsOn } from './wailsjs/runtime/runtime'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { DateNavigator } from '@/components/bujo/DateNavigator'
import { GetDayEntries, GetOverdue, GetHabits, GetLists, GetGoals, GetOutstandingQuestions, MarkEntryDone, MarkEntryUndone, EditEntry, DeleteEntry, HasChildren, MigrateEntry, MoveEntryToList, GetEntryContext, CyclePriority, RetypeEntry, CancelEntry, UncancelEntry } from './wailsjs/go/wails/App'
import { Sidebar, ViewType } from '@/components/bujo/Sidebar'
import { HabitTracker } from '@/components/bujo/HabitTracker'
import { ListsView } from '@/components/bujo/ListsView'
import { GoalsView } from '@/components/bujo/GoalsView'
import { QuestionsView } from '@/components/bujo/QuestionsView'
import { SearchView, SearchResult } from '@/components/bujo/SearchView'
import { StatsView } from '@/components/bujo/StatsView'
import { SettingsView } from '@/components/bujo/SettingsView'
import { Header } from '@/components/bujo/Header'
import { KeyboardShortcuts } from '@/components/bujo/KeyboardShortcuts'
import { EditEntryModal } from '@/components/bujo/EditEntryModal'
import { ConfirmDialog } from '@/components/bujo/ConfirmDialog'
import { MigrateModal } from '@/components/bujo/MigrateModal'
import { ListPickerModal } from '@/components/bujo/ListPickerModal'
import { AnswerQuestionModal } from '@/components/bujo/AnswerQuestionModal'
import { WeekView } from '@/components/bujo/WeekView'
import { PendingTasksView } from '@/components/bujo/PendingTasksView'
import { ContextTree } from '@/components/bujo/ContextTree'
import { buildTree } from '@/lib/buildTree'
import { EditableJournalView } from '@/components/bujo/EditableJournalView'
import { DayEntries, Habit, BujoList, Goal, Entry } from '@/types/bujo'
import { transformDayEntries, transformEntry, transformHabit, transformList, transformGoal } from '@/lib/transforms'
import { startOfDay } from '@/lib/utils'
import { toWailsTime } from '@/lib/wailsTime'
import { startOfWeek, endOfWeek, isSameWeek } from 'date-fns'
import { scrollToPosition } from '@/lib/scrollUtils'
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

const validViews: ViewType[] = ['today', 'pending', 'week', 'questions', 'habits', 'lists', 'goals', 'search', 'stats', 'settings']

function isValidView(view: unknown): view is ViewType {
  return validViews.includes(view as ViewType)
}

function App() {
  const { defaultView } = useSettings()
  const [view, setView] = useState<ViewType>(defaultView)
  const [days, setDays] = useState<DayEntries[]>([])
  const [habits, setHabits] = useState<Habit[]>([])
  const [lists, setLists] = useState<BujoList[]>([])
  const [goals, setGoals] = useState<Goal[]>([])
  const [overdueEntries, setOverdueEntries] = useState<Entry[]>([])
  const [outstandingQuestions, setOutstandingQuestions] = useState<Entry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [editModalEntry, setEditModalEntry] = useState<Entry | null>(null)
  const [deleteDialogEntry, setDeleteDialogEntry] = useState<Entry | null>(null)
  const [deleteHasChildren, setDeleteHasChildren] = useState(false)
  const [migrateModalEntry, setMigrateModalEntry] = useState<Entry | null>(null)
  const [moveToListEntry, setMoveToListEntry] = useState<Entry | null>(null)
  const [answerModalEntry, setAnswerModalEntry] = useState<Entry | null>(null)
  const [currentDate, setCurrentDate] = useState(() => startOfDay(new Date()))
  const [habitDays, setHabitDays] = useState(14)
  const [habitPeriod, setHabitPeriod] = useState<'week' | 'month' | 'quarter'>('week')
  const [habitAnchorDate, setHabitAnchorDate] = useState(() => new Date())
  const [reviewAnchorDate, setReviewAnchorDate] = useState(() => startOfDay(new Date()))
  const [reviewDays, setReviewDays] = useState<DayEntries[]>([])
  const [showKeyboardShortcuts, setShowKeyboardShortcuts] = useState(false)
  const [, setSelectedEntry] = useState<Entry | null>(null)
  const [reviewSelectedEntry, setReviewSelectedEntry] = useState<Entry | null>(null)
  const [reviewContextTree, setReviewContextTree] = useState<Entry[]>([])
  const [isWeekContextCollapsed, setIsWeekContextCollapsed] = useState(true)
  const [pendingContextWidth, setPendingContextWidth] = useState(384)
  const [pendingSelectedEntry, setPendingSelectedEntry] = useState<Entry | null>(null)
  const [pendingContextTree, setPendingContextTree] = useState<Entry[]>([])
  const initialLoadCompleteRef = useRef(false)
  const [highlightText, setHighlightText] = useState<string | null>(null)
  const { canGoBack, pushHistory, goBack, clearHistory } = useNavigationHistory()

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

      // Review view: show the full week (Mon-Sun) containing reviewAnchorDate
      // Use date-fns to get proper week boundaries with Monday as first day
      const reviewStart = startOfWeek(reviewAnchorDate, { weekStartsOn: 1 })
      const reviewEnd = endOfWeek(reviewAnchorDate, { weekStartsOn: 1 })

      const [daysData, overdueData, reviewDaysData, habitsData, listsData, goalsData, questionsData] = await Promise.all([
        GetDayEntries(toWailsTime(currentDate), toWailsTime(weekLater)),
        GetOverdue(),
        GetDayEntries(toWailsTime(reviewStart), toWailsTime(reviewEnd)),
        GetHabits(habitDays),
        GetLists(),
        GetGoals(toWailsTime(monthStart)),
        GetOutstandingQuestions(),
      ])

      const transformedDays = (daysData || []).map(transformDayEntries)
      setDays(transformedDays)
      const transformedReviewDays = (reviewDaysData || []).map(transformDayEntries)
      setReviewDays(transformedReviewDays)
      const transformedOverdue = (overdueData || []).map(transformEntry)
      setOverdueEntries(transformedOverdue)
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

  useEffect(() => {
    const unsubscribe = EventsOn('data:changed', () => {
      loadData()
    })
    return unsubscribe
  }, [loadData])

  const todayEntries = days[0]?.entries || []
  const flatEntries = flattenEntries(todayEntries)

  // Fetch full context tree for review-selected entry from backend
  useEffect(() => {
    if (!reviewSelectedEntry) {
      setReviewContextTree([])
      return
    }

    GetEntryContext(reviewSelectedEntry.id)
      .then((entries) => {
        setReviewContextTree(entries.map(transformEntry))
      })
      .catch((err) => {
        console.error('Failed to fetch entry context:', err)
        setReviewContextTree([])
      })
  }, [reviewSelectedEntry])

  // Fetch full context tree for pending-selected entry from backend
  useEffect(() => {
    if (!pendingSelectedEntry) {
      setPendingContextTree([])
      return
    }

    GetEntryContext(pendingSelectedEntry.id)
      .then((entries) => {
        setPendingContextTree(entries.map(transformEntry))
      })
      .catch((err) => {
        console.error('Failed to fetch entry context:', err)
        setPendingContextTree([])
      })
  }, [pendingSelectedEntry])

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

  const handleDateNavigatorChange = useCallback((date: Date) => {
    setCurrentDate(startOfDay(date))
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

  const handleGoToCurrentWeek = useCallback(() => {
    setReviewAnchorDate(startOfDay(new Date()))
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

  const handleViewChange = useCallback((newView: ViewType) => {
    if (newView === 'today') {
      clearHistory()
    } else {
      pushHistory({
        view: view,
        scrollPosition: window.scrollY,
      })
    }
    setHighlightText(null)
    setView(newView)
    setSelectedIndex(0)
  }, [view, clearHistory, pushHistory])

  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      const isInputFocused = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA'
      // CodeMirror uses contenteditable, not INPUT/TEXTAREA
      const isCodeMirrorFocused = target.closest?.('.cm-editor') !== null

      // ? toggles keyboard shortcuts panel (works even when not in input)
      if (e.key === '?' && !isInputFocused && !isCodeMirrorFocused) {
        e.preventDefault()
        setShowKeyboardShortcuts(prev => !prev)
        return
      }

      if (isInputFocused || isCodeMirrorFocused) return

      // View navigation shortcuts (CMD+1 through CMD+9) - always available
      if ((e.metaKey || e.ctrlKey) && e.key >= '1' && e.key <= '9') {
        e.preventDefault()
        const viewMap: ViewType[] = ['today', 'pending', 'week', 'questions', 'habits', 'lists', 'goals', 'search', 'stats', 'settings']
        const viewIndex = parseInt(e.key) - 1
        if (viewIndex < viewMap.length) {
          handleViewChange(viewMap[viewIndex])
        }
        return
      }

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

      // Week view shortcuts
      if (view === 'week') {
        if (e.key === '[') {
          e.preventDefault()
          setIsWeekContextCollapsed(prev => !prev)
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

      if (view !== 'today') return

      // Main panel navigation
      if (flatEntries.length === 0) return

      switch (e.key) {
        case 'j':
        case 'ArrowDown': {
          e.preventDefault()
          const nextIndex = Math.min(selectedIndex + 1, flatEntries.length - 1)
          setSelectedIndex(nextIndex)
          setSelectedEntry(flatEntries[nextIndex])
          break
        }
        case 'k':
        case 'ArrowUp': {
          e.preventDefault()
          const prevIndex = Math.max(selectedIndex - 1, 0)
          setSelectedIndex(prevIndex)
          setSelectedEntry(flatEntries[prevIndex])
          break
        }
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
        case 'm': {
          e.preventDefault()
          const entry = flatEntries[selectedIndex]
          if (entry && (entry.type === 'task' || entry.type === 'question')) {
            setMigrateModalEntry(entry)
          }
          break
        }
        case 'p': {
          e.preventDefault()
          const entry = flatEntries[selectedIndex]
          if (entry) {
            CyclePriority(entry.id).then(() => loadData())
          }
          break
        }
        case 't': {
          e.preventDefault()
          const entry = flatEntries[selectedIndex]
          if (entry) {
            const cycleOrder = ['task', 'note', 'event', 'question'] as const
            const currentIndex = cycleOrder.indexOf(entry.type as typeof cycleOrder[number])
            if (currentIndex !== -1) {
              const nextType = cycleOrder[(currentIndex + 1) % cycleOrder.length]
              RetypeEntry(entry.id, nextType).then(() => loadData())
            }
          }
          break
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [view, flatEntries, selectedIndex, loadData, handleDeleteEntryRequest, handlePrevDay, handleNextDay, handleGoToToday, cycleHabitPeriod, handleViewChange])

  useEffect(() => {
    setSelectedIndex(0)
    const entries = flattenEntries(days[0]?.entries || [])
    setSelectedEntry(entries[0] ?? null)
  }, [days])

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

  const handleMoveToList = useCallback(async (listId: number) => {
    if (!moveToListEntry) return
    try {
      await MoveEntryToList(moveToListEntry.id, listId)
      setMoveToListEntry(null)
      loadData()
    } catch (err) {
      console.error('Failed to move entry to list:', err)
      setError(err instanceof Error ? err.message : 'Failed to move entry to list')
    }
  }, [moveToListEntry, loadData])

  // const handleMoveToRoot = useCallback(async (entry: Entry) => {
  //   try {
  //     await MoveEntryToRoot(entry.id)
  //     loadData()
  //   } catch (err) {
  //     console.error('Failed to move entry to root:', err)
  //     setError(err instanceof Error ? err.message : 'Failed to move entry to root')
  //   }
  // }, [loadData])

  const handleSidebarMarkDone = useCallback(async (entry: Entry) => {
    try {
      await MarkEntryDone(entry.id)
      loadData()
    } catch (err) {
      console.error('Failed to mark entry done:', err)
      setError(err instanceof Error ? err.message : 'Failed to mark entry done')
    }
  }, [loadData])

  const handleSidebarUnmarkDone = useCallback(async (entry: Entry) => {
    try {
      await MarkEntryUndone(entry.id)
      loadData()
    } catch (err) {
      console.error('Failed to unmark entry done:', err)
      setError(err instanceof Error ? err.message : 'Failed to unmark entry done')
    }
  }, [loadData])

  const handleSidebarCyclePriority = useCallback(async (entry: Entry) => {
    try {
      await CyclePriority(entry.id)
      loadData()
    } catch (err) {
      console.error('Failed to cycle priority:', err)
      setError(err instanceof Error ? err.message : 'Failed to cycle priority')
    }
  }, [loadData])

  const handleSidebarCycleType = useCallback(async (entry: Entry) => {
    const cycleOrder = ['task', 'note', 'event', 'question'] as const
    const currentIndex = cycleOrder.indexOf(entry.type as typeof cycleOrder[number])
    if (currentIndex === -1) return
    const nextType = cycleOrder[(currentIndex + 1) % cycleOrder.length]

    try {
      await RetypeEntry(entry.id, nextType)
      loadData()
    } catch (err) {
      console.error('Failed to cycle type:', err)
      setError(err instanceof Error ? err.message : 'Failed to cycle type')
    }
  }, [loadData])

  const handleSidebarCancel = useCallback(async (entry: Entry) => {
    try {
      await CancelEntry(entry.id)
      loadData()
    } catch (err) {
      console.error('Failed to cancel entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to cancel entry')
    }
  }, [loadData])

  const handleSidebarUncancel = useCallback(async (entry: Entry) => {
    try {
      await UncancelEntry(entry.id)
      loadData()
    } catch (err) {
      console.error('Failed to uncancel entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to uncancel entry')
    }
  }, [loadData])

  const sidebarCallbacks = useMemo(() => ({
    onMarkDone: handleSidebarMarkDone,
    onUnmarkDone: handleSidebarUnmarkDone,
    onMigrate: (entry: Entry) => setMigrateModalEntry(entry),
    onEdit: (entry: Entry) => setEditModalEntry(entry),
    onDelete: handleDeleteEntryRequest,
    onCyclePriority: handleSidebarCyclePriority,
    onCycleType: handleSidebarCycleType,
    onMoveToList: (entry: Entry) => setMoveToListEntry(entry),
    onCancel: handleSidebarCancel,
    onUncancel: handleSidebarUncancel,
  }), [handleSidebarMarkDone, handleSidebarUnmarkDone, handleDeleteEntryRequest, handleSidebarCyclePriority, handleSidebarCycleType, handleSidebarCancel, handleSidebarUncancel])


  const handleSearchMigrate = useCallback((result: SearchResult) => {
    setMigrateModalEntry({
      id: result.id,
      content: result.content,
      type: result.type,
      priority: result.priority,
      parentId: result.parentId,
      loggedDate: result.date
    })
  }, [])

  const handleNavigateToEntry = useCallback((entry: Entry) => {
    pushHistory({
      view: view,
      scrollPosition: window.scrollY,
    })
    const entryDate = new Date(entry.loggedDate)
    setCurrentDate(startOfDay(entryDate))
    setHighlightText(entry.content)
    setView('today')
    setSelectedIndex(0)
  }, [view, pushHistory])

  const handleSearchNavigate = useCallback((result: SearchResult) => {
    const entryDate = new Date(result.date)
    setReviewAnchorDate(startOfDay(entryDate))
    setView('week')
  }, [])

  const handleSearchSelectEntry = useCallback((result: SearchResult) => {
    setSelectedEntry({
      id: result.id,
      content: result.content,
      type: result.type,
      priority: result.priority,
      parentId: result.parentId,
      loggedDate: result.date,
    })
  }, [])

  const handleBack = useCallback(() => {
    const previousState = goBack()
    if (previousState && isValidView(previousState.view)) {
      setView(previousState.view)
      scrollToPosition(previousState.scrollPosition)
    }
  }, [goBack])

  const viewTitles: Record<ViewType, string> = {
    today: 'Journal',
    pending: 'Pending Tasks',
    week: 'Weekly Review',
    questions: 'Open Questions',
    habits: 'Habit Tracker',
    lists: 'Lists',
    goals: 'Monthly Goals',
    search: 'Search',
    stats: 'Insights',
    settings: 'Settings',
    editable: 'Edit Journal',
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

  return (
    <div className="flex h-screen bg-background">
      <Sidebar currentView={view} onViewChange={handleViewChange} />

      <div className="flex-1 flex flex-col overflow-hidden">
        <Header
          title={viewTitles[view]}
          currentMood={today?.mood}
          currentWeather={today?.weather}
          currentLocation={today?.location}
          currentDate={currentDate}
          onMoodChanged={loadData}
          onWeatherChanged={loadData}
          onLocationChanged={loadData}
          canGoBack={canGoBack}
          onBack={handleBack}
        />

        <main className={`flex-1 p-6 ${view === 'today' ? 'flex flex-col overflow-hidden pb-2' : 'overflow-y-auto pb-32'}`}>
          {view === 'today' && (
            <>
              <div className="flex items-center justify-center mb-6">
                <DateNavigator
                  date={currentDate}
                  onDateChange={handleDateNavigatorChange}
                />
              </div>
              <EditableJournalView
                date={currentDate}
                highlightText={highlightText}
                onHighlightDone={() => setHighlightText(null)}
              />
            </>
          )}

          {view === 'pending' && (
            <PendingTasksView
              overdueEntries={overdueEntries}
              now={new Date()}
              callbacks={sidebarCallbacks}
              selectedEntry={pendingSelectedEntry ?? undefined}
              onSelectEntry={(entry) => setPendingSelectedEntry(entry)}
              onNavigateToEntry={handleNavigateToEntry}
              onRefresh={loadData}
            />
          )}

          {view === 'week' && (
            <div className="max-w-full mx-auto space-y-8 h-full flex flex-col">
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
                  Week of {reviewDays.length > 0 && reviewDays[0]?.date} - {reviewDays.length > 0 && reviewDays[reviewDays.length - 1]?.date}
                </span>
                <button
                  onClick={handleNextWeek}
                  title="Next week"
                  disabled={reviewAnchorDate >= startOfDay(new Date())}
                  className="p-2 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
                {!isSameWeek(reviewAnchorDate, new Date(), { weekStartsOn: 1 }) && (
                  <button
                    onClick={handleGoToCurrentWeek}
                    className="px-2 py-1 text-xs rounded hover:bg-secondary/50 transition-colors"
                  >
                    Today
                  </button>
                )}
              </div>
              <WeekView
                days={reviewDays}
                habits={habits}
                callbacks={{
                  onNavigateToEntry: handleNavigateToEntry,
                }}
                onSelectEntry={(entry) => setReviewSelectedEntry(entry ?? null)}
                contextTree={reviewContextTree}
                isContextCollapsed={isWeekContextCollapsed}
                onToggleContextCollapse={() => setIsWeekContextCollapsed(prev => !prev)}
              />
            </div>
          )}

          {view === 'questions' && (
            <div className="max-w-3xl mx-auto">
              <QuestionsView
                questions={outstandingQuestions}
                onEntryChanged={loadData}
                onError={setError}
                onEdit={(entry) => setEditModalEntry(entry)}
                onNavigateToEntry={handleNavigateToEntry}
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
              <SearchView
                onMigrate={handleSearchMigrate}
                onNavigateToEntry={handleSearchNavigate}
                onSelectEntry={handleSearchSelectEntry}
              />
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

      {/* Pending Tasks Context Sidebar */}
      {view === 'pending' && (
        <aside
          className="flex flex-col self-stretch bg-background relative"
          style={{ width: `${pendingContextWidth}px` }}
        >
          <div
            data-testid="pending-context-resize-handle"
            className="absolute left-0 top-0 h-full w-4 cursor-col-resize group z-10"
            onMouseDown={(e) => {
              e.preventDefault();
              const handleMouseMove = (moveEvent: MouseEvent) => {
                const newWidth = window.innerWidth - moveEvent.clientX;
                const clampedWidth = Math.max(256, Math.min(960, newWidth));
                setPendingContextWidth(clampedWidth);
              };
              const handleMouseUp = () => {
                document.removeEventListener('mousemove', handleMouseMove);
                document.removeEventListener('mouseup', handleMouseUp);
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
              };
              document.addEventListener('mousemove', handleMouseMove);
              document.addEventListener('mouseup', handleMouseUp);
              document.body.style.cursor = 'col-resize';
              document.body.style.userSelect = 'none';
            }}
          >
            <div className="w-px h-full bg-border transition-colors group-hover:bg-primary/50" />
          </div>

          <div className="h-[73px] flex-shrink-0" />

          <div className="flex-1 overflow-y-auto p-4">
            <h3 className="text-sm font-medium mb-3">Context</h3>
            {!pendingSelectedEntry ? (
              <p className="text-sm text-muted-foreground">No entry selected</p>
            ) : pendingContextTree.length === 0 ? (
              <p className="text-sm text-muted-foreground">No context</p>
            ) : (
              <ContextTree nodes={buildTree(pendingContextTree)} selectedEntryId={pendingSelectedEntry.id} />
            )}
          </div>
        </aside>
      )}

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

      {/* Move to List Modal */}
      <ListPickerModal
        isOpen={moveToListEntry !== null}
        entries={moveToListEntry ? [moveToListEntry.content] : []}
        onSelect={handleMoveToList}
        onCancel={() => setMoveToListEntry(null)}
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

    </div>
  )
}

export default App
