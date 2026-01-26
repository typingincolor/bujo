import { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { useNavigationHistory } from '@/hooks/useNavigationHistory'
import { useSettings } from '@/contexts/SettingsContext'
import { EventsOn } from './wailsjs/runtime/runtime'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { DateNavigator } from '@/components/bujo/DateNavigator'
import { GetAgenda, GetHabits, GetLists, GetGoals, GetOutstandingQuestions, AddEntry, AddChildEntry, MarkEntryDone, MarkEntryUndone, EditEntry, DeleteEntry, HasChildren, MigrateEntry, MoveEntryToList, MoveEntryToRoot, OpenFileDialog, GetEntryContext, CyclePriority, RetypeEntry } from './wailsjs/go/wails/App'
import { Sidebar, ViewType } from '@/components/bujo/Sidebar'
import { DayView } from '@/components/bujo/DayView'
import { HabitTracker } from '@/components/bujo/HabitTracker'
import { ListsView } from '@/components/bujo/ListsView'
import { GoalsView } from '@/components/bujo/GoalsView'
import { QuestionsView } from '@/components/bujo/QuestionsView'
import { SearchView, SearchResult } from '@/components/bujo/SearchView'
import { StatsView } from '@/components/bujo/StatsView'
import { SettingsView } from '@/components/bujo/SettingsView'
import { Header } from '@/components/bujo/Header'
import { CaptureModal } from '@/components/bujo/CaptureModal'
import { KeyboardShortcuts } from '@/components/bujo/KeyboardShortcuts'
import { EditEntryModal } from '@/components/bujo/EditEntryModal'
import { ConfirmDialog } from '@/components/bujo/ConfirmDialog'
import { MigrateModal } from '@/components/bujo/MigrateModal'
import { ListPickerModal } from '@/components/bujo/ListPickerModal'
import { AnswerQuestionModal } from '@/components/bujo/AnswerQuestionModal'
import { QuickStats } from '@/components/bujo/QuickStats'
import { CaptureBar } from '@/components/bujo/CaptureBar'
import { WeekView } from '@/components/bujo/WeekView'
import { JournalSidebar } from '@/components/bujo/JournalSidebar'
import { DayEntries, Habit, BujoList, Goal, Entry } from '@/types/bujo'
import { transformDayEntries, transformEntry, transformHabit, transformList, transformGoal } from '@/lib/transforms'
import { startOfDay } from '@/lib/utils'
import { toWailsTime } from '@/lib/wailsTime'
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

const validViews: ViewType[] = ['today', 'week', 'questions', 'habits', 'lists', 'goals', 'search', 'stats', 'settings']

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
  const [overdueCount, setOverdueCount] = useState(0)
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
  const [showCaptureModal, setShowCaptureModal] = useState(false)
  const [, setSelectedEntry] = useState<Entry | null>(null)
  const [sidebarSelectedEntry, setSidebarSelectedEntry] = useState<Entry | null>(null)
  const [sidebarSelectedIndex, setSidebarSelectedIndex] = useState(0)
  const [focusedPanel, setFocusedPanel] = useState<'main' | 'sidebar'>('main')
  const [sidebarContextTree, setSidebarContextTree] = useState<Entry[]>([])
  const [captureParentEntry, setCaptureParentEntry] = useState<Entry | null>(null)
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [journalSidebarWidth, setJournalSidebarWidth] = useState(512)
  const initialLoadCompleteRef = useRef(false)
  const captureBarRef = useRef<HTMLTextAreaElement>(null)
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
      setOverdueCount(transformedOverdue.filter(e => e.type === 'task').length)
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

  // Fetch full context tree for sidebar-selected entry from backend
  useEffect(() => {
    if (!sidebarSelectedEntry) {
      setSidebarContextTree([])
      return
    }

    GetEntryContext(sidebarSelectedEntry.id)
      .then((entries) => {
        setSidebarContextTree(entries.map(transformEntry))
      })
      .catch((err) => {
        console.error('Failed to fetch entry context:', err)
        setSidebarContextTree([])
      })
  }, [sidebarSelectedEntry])

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
        if (e.key === '[') {
          e.preventDefault()
          setIsSidebarCollapsed(prev => !prev)
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

      // Entry creation shortcuts (c, i, r, a, A) - work in today view
      if (view === 'today') {
        if (e.key === 'c') {
          e.preventDefault()
          setShowCaptureModal(true)
          return
        }
        if (e.key === 'i') {
          e.preventDefault()
          captureBarRef.current?.focus()
          return
        }
        if (e.key === 'r') {
          e.preventDefault()
          setCaptureParentEntry(null)
          captureBarRef.current?.focus()
          return
        }
        if (e.key === 'a') {
          e.preventDefault()
          // If selected entry is a question, open answer modal instead
          const selectedEntry = flatEntries[selectedIndex]
          if (selectedEntry?.type === 'question') {
            setAnswerModalEntry(selectedEntry)
          } else {
            setCaptureParentEntry(null)
            captureBarRef.current?.focus()
          }
          return
        }
        if (e.key === 'A') {
          e.preventDefault()
          const selectedEntry = flatEntries[selectedIndex]
          if (selectedEntry) {
            setCaptureParentEntry(selectedEntry)
            captureBarRef.current?.focus()
          }
          return
        }
      }

      if (view !== 'today') return

      // Get sidebar task entries (filtered the same way as JournalSidebar)
      const sidebarTaskEntries = overdueEntries.filter(e => e.type === 'task')

      // Tab key switches focus between main panel and sidebar
      if (e.key === 'Tab') {
        e.preventDefault()
        if (focusedPanel === 'main') {
          // Switch to sidebar
          setFocusedPanel('sidebar')
          setSelectedIndex(-1) // Clear main panel selection
          setSelectedEntry(null)
          if (sidebarTaskEntries.length > 0) {
            setSidebarSelectedIndex(0)
            setSidebarSelectedEntry(sidebarTaskEntries[0])
          }
        } else {
          // Switch to main panel
          setFocusedPanel('main')
          setSidebarSelectedIndex(-1)
          setSidebarSelectedEntry(null)
          if (flatEntries.length > 0) {
            setSelectedIndex(0)
            setSelectedEntry(flatEntries[0])
          }
        }
        return
      }

      // Navigation keys depend on which panel is focused
      if (focusedPanel === 'sidebar') {
        switch (e.key) {
          case 'j':
          case 'ArrowDown': {
            e.preventDefault()
            if (sidebarTaskEntries.length === 0) return
            const nextIndex = Math.min(sidebarSelectedIndex + 1, sidebarTaskEntries.length - 1)
            setSidebarSelectedIndex(nextIndex)
            setSidebarSelectedEntry(sidebarTaskEntries[nextIndex])
            break
          }
          case 'k':
          case 'ArrowUp': {
            e.preventDefault()
            if (sidebarTaskEntries.length === 0) return
            const prevIndex = Math.max(sidebarSelectedIndex - 1, 0)
            setSidebarSelectedIndex(prevIndex)
            setSidebarSelectedEntry(sidebarTaskEntries[prevIndex])
            break
          }
        }
        return
      }

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
  }, [view, flatEntries, selectedIndex, overdueEntries, focusedPanel, sidebarSelectedIndex, loadData, handleDeleteEntryRequest, handlePrevDay, handleNextDay, handleGoToToday, cycleHabitPeriod])

  useEffect(() => {
    // Only reset main panel selection if main panel is focused
    // This prevents dual highlighting when sidebar has selection and data refreshes
    if (focusedPanel === 'main') {
      setSelectedIndex(0)
      const entries = flattenEntries(days[0]?.entries || [])
      setSelectedEntry(entries[0] ?? null)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [days]) // Only run when days changes, focusedPanel is just a guard condition

  const handleViewChange = (newView: ViewType) => {
    if (newView === 'today') {
      clearHistory()
    } else {
      pushHistory({
        view: view,
        scrollPosition: window.scrollY,
      })
    }
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

  const handleSelectEntry = useCallback((id: number) => {
    const index = flatEntries.findIndex(e => e.id === id)
    if (index !== -1) {
      setSelectedIndex(index)
      setSelectedEntry(flatEntries[index])
      // Clear sidebar selection when main panel is selected
      setSidebarSelectedEntry(null)
      setSidebarSelectedIndex(-1)
      setFocusedPanel('main')
    }
  }, [flatEntries])

  const handleAddChild = useCallback((entry: Entry) => {
    const index = flatEntries.findIndex(e => e.id === entry.id)
    if (index !== -1) {
      setSelectedIndex(index)
      setCaptureParentEntry(entry)
      captureBarRef.current?.focus()
    }
  }, [flatEntries])

  const handleSidebarSelectEntry = useCallback((entry: Entry) => {
    // Find the index of this entry in the filtered task entries
    const sidebarTaskEntries = overdueEntries.filter(e => e.type === 'task')
    const index = sidebarTaskEntries.findIndex(e => e.id === entry.id)
    setSidebarSelectedEntry(entry)
    setSidebarSelectedIndex(index)
    // Clear main panel selection when sidebar is selected
    setSelectedIndex(-1)
    setSelectedEntry(null)
    setFocusedPanel('sidebar')
  }, [overdueEntries])

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

  const handleSidebarCyclePriority = useCallback(async (entry: Entry) => {
    try {
      await CyclePriority(entry.id)
      loadData()
    } catch (err) {
      console.error('Failed to cycle priority:', err)
      setError(err instanceof Error ? err.message : 'Failed to cycle priority')
    }
  }, [loadData])

  const sidebarCallbacks = useMemo(() => ({
    onMarkDone: handleSidebarMarkDone,
    onMigrate: (entry: Entry) => setMigrateModalEntry(entry),
    onEdit: (entry: Entry) => setEditModalEntry(entry),
    onDelete: handleDeleteEntryRequest,
    onCyclePriority: handleSidebarCyclePriority,
    onMoveToList: (entry: Entry) => setMoveToListEntry(entry),
  }), [handleSidebarMarkDone, handleDeleteEntryRequest, handleSidebarCyclePriority])


  const handleCaptureBarSubmit = useCallback(async (content: string) => {
    try {
      await AddEntry(content, toWailsTime(currentDate))
      loadData()
    } catch (err) {
      console.error('Failed to add entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to add entry')
    }
  }, [loadData, currentDate])

  const handleCaptureBarSubmitChild = useCallback(async (parentId: number, content: string) => {
    try {
      await AddChildEntry(parentId, content, toWailsTime(currentDate))
      setCaptureParentEntry(null)
      loadData()
    } catch (err) {
      console.error('Failed to add child entry:', err)
      setError(err instanceof Error ? err.message : 'Failed to add child entry')
    }
  }, [loadData, currentDate])

  const handleCaptureBarClearParent = useCallback(() => {
    setCaptureParentEntry(null)
  }, [])

  const handleCaptureBarFileImport = useCallback(async () => {
    try {
      const fileContent = await OpenFileDialog()
      if (fileContent && captureBarRef.current) {
        const currentValue = captureBarRef.current.value
        captureBarRef.current.value = currentValue + fileContent
        captureBarRef.current.dispatchEvent(new Event('input', { bubbles: true }))
      }
    } catch (err) {
      console.error('Failed to import file:', err)
      setError(err instanceof Error ? err.message : 'Failed to import file')
    }
  }, [])

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
    week: 'Weekly Review',
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
          onUpload={view === 'today' ? handleCaptureBarFileImport : undefined}
        />

        <main className="flex-1 overflow-y-auto p-6">
          {view === 'today' && (
            <div className="space-y-6">
              {/* Day Navigation */}
              <div className="flex items-center justify-center">
                <DateNavigator
                  date={currentDate}
                  onDateChange={handleDateNavigatorChange}
                />
              </div>
              <QuickStats days={days} habits={habits} goals={goals} overdueCount={overdueCount} />
              {today && (
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
                  onMoveToList={(entry) => setMoveToListEntry(entry)}
                />
              )}
              <CaptureBar
                ref={captureBarRef}
                onSubmit={handleCaptureBarSubmit}
                onSubmitChild={handleCaptureBarSubmitChild}
                onClearParent={handleCaptureBarClearParent}
                parentEntry={captureParentEntry}
                sidebarWidth={journalSidebarWidth}
                isSidebarCollapsed={isSidebarCollapsed}
              />
            </div>
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
              </div>
              <WeekView
                days={reviewDays}
                callbacks={{
                  onMarkDone: async (entry) => {
                    try {
                      await MarkEntryDone(entry.id)
                      loadData()
                    } catch (err) {
                      console.error('Failed to mark entry done:', err)
                      setError(err instanceof Error ? err.message : 'Failed to mark entry done')
                    }
                  },
                  onMigrate: (entry) => setMigrateModalEntry(entry),
                  onEdit: (entry) => setEditModalEntry(entry),
                  onDelete: handleDeleteEntryRequest,
                  onCyclePriority: async (entry) => {
                    try {
                      await CyclePriority(entry.id)
                      loadData()
                    } catch (err) {
                      console.error('Failed to cycle priority:', err)
                      setError(err instanceof Error ? err.message : 'Failed to cycle priority')
                    }
                  },
                  onMoveToList: (entry) => setMoveToListEntry(entry),
                }}
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

      {/* Journal Sidebar - Overdue + Context - always visible in journal view */}
      {view === 'today' && (
        <aside
          className="h-screen border-l border-border bg-background overflow-y-auto p-2 transition-all duration-300 ease-in-out"
          style={{
            width: isSidebarCollapsed ? '2.5rem' : `${journalSidebarWidth}px`
          }}
        >
          <JournalSidebar
            overdueEntries={overdueEntries}
            now={currentDate}
            selectedEntry={sidebarSelectedEntry ?? undefined}
            contextTree={sidebarContextTree}
            onSelectEntry={handleSidebarSelectEntry}
            callbacks={sidebarCallbacks}
            isCollapsed={isSidebarCollapsed}
            onToggleCollapse={() => setIsSidebarCollapsed(prev => !prev)}
            onWidthChange={setJournalSidebarWidth}
          />
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
        entryContent={moveToListEntry?.content || ''}
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
