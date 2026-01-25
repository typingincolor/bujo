import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo'
import { cn } from '@/lib/utils'
import { HelpCircle, ChevronDown, ChevronRight, X, RotateCcw, Trash2, Flag, RefreshCw, MessageCircle } from 'lucide-react'
import { format, parseISO } from 'date-fns'
import { useState, useEffect, useCallback, useMemo } from 'react'
import { CancelEntry, UncancelEntry, DeleteEntry, CyclePriority, RetypeEntry } from '@/wailsjs/go/wails/App'
import { AnswerQuestionModal } from './AnswerQuestionModal'

interface QuestionsViewProps {
  questions: Entry[]
  onEntryChanged?: () => void
  onError?: (message: string) => void
}

function groupByDate(entries: Entry[]): Map<string, Entry[]> {
  const groups = new Map<string, Entry[]>()
  for (const entry of entries) {
    const date = entry.loggedDate.split('T')[0]
    if (!groups.has(date)) {
      groups.set(date, [])
    }
    groups.get(date)!.push(entry)
  }
  return groups
}

function formatDateHeader(dateStr: string): string {
  try {
    const date = parseISO(dateStr)
    return format(date, 'MMM d')
  } catch {
    return dateStr
  }
}

function buildParentChain(entry: Entry, entriesById: Map<number, Entry>): Entry[] {
  const chain: Entry[] = []
  let current = entry
  while (current.parentId !== null) {
    const parent = entriesById.get(current.parentId)
    if (!parent) break
    chain.unshift(parent)
    current = parent
  }
  return chain
}

export function QuestionsView({ questions, onEntryChanged, onError }: QuestionsViewProps) {
  const [collapsed, setCollapsed] = useState(false)
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const [answerModalOpen, setAnswerModalOpen] = useState(false)
  const [questionToAnswer, setQuestionToAnswer] = useState<Entry | null>(null)

  // Build a lookup map for all entries by ID
  const entriesById = new Map<number, Entry>()
  for (const entry of questions) {
    entriesById.set(entry.id, entry)
  }

  // Filter to only show question entries
  const questionEntries = questions.filter(e => e.type === 'question')
  const grouped = groupByDate(questionEntries)
  const sortedDates = Array.from(grouped.keys()).sort()

  // Build flat list of entries in display order for keyboard navigation
  const flatEntries = useMemo(() => {
    const questionEntriesFiltered = questions.filter(e => e.type === 'question')
    const groupedEntries = groupByDate(questionEntriesFiltered)
    const dates = Array.from(groupedEntries.keys()).sort()
    const entries: Entry[] = []
    for (const dateStr of dates) {
      entries.push(...groupedEntries.get(dateStr)!)
    }
    return entries
  }, [questions])

  // Map entry ID to flat index for selection
  const entryToFlatIndex = new Map<number, number>()
  flatEntries.forEach((entry, index) => {
    entryToFlatIndex.set(entry.id, index)
  })

  const toggleExpanded = (id: number) => {
    setExpandedIds(prev => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const handleAnswer = (entry: Entry) => {
    setQuestionToAnswer(entry)
    setAnswerModalOpen(true)
  }

  const handleAnswerSubmitted = () => {
    setAnswerModalOpen(false)
    setQuestionToAnswer(null)
    onEntryChanged?.()
  }

  const handleCancel = useCallback(async (entry: Entry) => {
    try {
      await CancelEntry(entry.id)
      onEntryChanged?.()
    } catch (error) {
      console.error('Failed to cancel entry:', error)
      onError?.(error instanceof Error ? error.message : 'Failed to cancel entry')
    }
  }, [onEntryChanged, onError])

  const handleUncancel = useCallback(async (entry: Entry) => {
    try {
      await UncancelEntry(entry.id)
      onEntryChanged?.()
    } catch (error) {
      console.error('Failed to uncancel entry:', error)
      onError?.(error instanceof Error ? error.message : 'Failed to uncancel entry')
    }
  }, [onEntryChanged, onError])

  const handleDelete = useCallback(async (entry: Entry) => {
    try {
      await DeleteEntry(entry.id)
      onEntryChanged?.()
    } catch (error) {
      console.error('Failed to delete entry:', error)
      onError?.(error instanceof Error ? error.message : 'Failed to delete entry')
    }
  }, [onEntryChanged, onError])

  const handleCyclePriority = useCallback(async (entry: Entry) => {
    try {
      await CyclePriority(entry.id)
      onEntryChanged?.()
    } catch (error) {
      console.error('Failed to cycle priority:', error)
      onError?.(error instanceof Error ? error.message : 'Failed to cycle priority')
    }
  }, [onEntryChanged, onError])

  const handleCycleType = useCallback(async (entry: Entry) => {
    const cycleOrder = ['task', 'note', 'event', 'question'] as const
    const currentIndex = cycleOrder.indexOf(entry.type as typeof cycleOrder[number])
    if (currentIndex === -1) return
    const nextType = cycleOrder[(currentIndex + 1) % cycleOrder.length]
    try {
      await RetypeEntry(entry.id, nextType)
      onEntryChanged?.()
    } catch (error) {
      console.error('Failed to cycle type:', error)
      onError?.(error instanceof Error ? error.message : 'Failed to cycle type')
    }
  }, [onEntryChanged, onError])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      const isInputFocused = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA'
      if (isInputFocused) return
      if (flatEntries.length === 0) return

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
        case 'x':
          e.preventDefault()
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex]
            if (selected.type === 'cancelled') {
              await handleUncancel(selected)
            } else {
              await handleCancel(selected)
            }
          }
          break
        case 'p':
          e.preventDefault()
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex]
            await handleCyclePriority(selected)
          }
          break
        case 't':
          e.preventDefault()
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex]
            await handleCycleType(selected)
          }
          break
        case 'a':
          e.preventDefault()
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex]
            handleAnswer(selected)
          }
          break
        case 'Enter':
          e.preventDefault()
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex]
            toggleExpanded(selected.id)
          }
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [flatEntries, selectedIndex, handleCancel, handleCyclePriority, handleCycleType, handleUncancel])

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center gap-2">
        <button
          onClick={() => setCollapsed(!collapsed)}
          title={collapsed ? 'Expand' : 'Collapse'}
          className="p-1 rounded hover:bg-secondary transition-colors"
        >
          {collapsed ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <ChevronDown className="w-4 h-4" />
          )}
        </button>
        <HelpCircle className="w-5 h-5 text-bujo-question" data-testid="questions-icon" />
        <h2 className="font-display text-xl font-semibold flex-1">Open Questions</h2>
        <span className="px-2 py-0.5 text-sm font-medium bg-bujo-question/20 text-bujo-question rounded-full">
          {questionEntries.length}
        </span>
      </div>

      {/* Content */}
      {!collapsed && (
        <>
          {questionEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground italic py-6 text-center">
              No open questions. All answered!
            </p>
          ) : (
            <div className="space-y-4">
              {sortedDates.map((dateStr) => (
                <div key={dateStr} className="space-y-2">
                  <h3 className="text-sm font-medium text-muted-foreground">
                    {formatDateHeader(dateStr)}
                  </h3>
                  <div className="space-y-1">
                    {grouped.get(dateStr)!.map((entry) => {
                      const isExpanded = expandedIds.has(entry.id)
                      const parentChain = buildParentChain(entry, entriesById)
                      const isSelected = entryToFlatIndex.get(entry.id) === selectedIndex
                      return (
                        <div key={entry.id} className="space-y-1">
                          {/* Parent context entries (shown when expanded) */}
                          {isExpanded && parentChain.map((parent, index) => (
                            <div
                              key={parent.id}
                              className={cn(
                                'flex items-center gap-3 p-2 rounded-lg border border-border',
                                'bg-muted/30 text-muted-foreground'
                              )}
                              style={{ marginLeft: `${index * 16}px` }}
                            >
                              <span className="w-5 text-center font-mono">
                                {ENTRY_SYMBOLS[parent.type]}
                              </span>
                              <span className="flex-1 text-sm">{parent.content}</span>
                            </div>
                          ))}
                          {/* Question entry */}
                          <div
                            onClick={() => toggleExpanded(entry.id)}
                            className={cn(
                              'flex items-center gap-3 p-2 rounded-lg border border-border cursor-pointer',
                              'bg-card transition-colors group',
                              !isSelected && 'hover:bg-secondary/30',
                              isSelected && 'ring-2 ring-primary'
                            )}
                            style={{ marginLeft: isExpanded ? `${parentChain.length * 16}px` : undefined }}
                          >
                            {/* Context dot - indicates entry has ancestors */}
                            {entry.parentId !== null && (
                              <span
                                data-testid="context-dot"
                                className="w-1.5 h-1.5 rounded-full bg-muted-foreground flex-shrink-0"
                                title="Has parent context"
                              />
                            )}
                            <span
                              data-testid="entry-symbol"
                              className="w-5 text-center text-bujo-question font-mono"
                            >
                              {ENTRY_SYMBOLS[entry.type]}
                            </span>
                            <span className={cn(
                              'flex-1 text-sm',
                              entry.type === 'cancelled' && 'line-through text-muted-foreground'
                            )}>
                              {entry.content}
                            </span>
                            {entry.priority !== 'none' && (
                              <span className="text-xs text-warning font-medium">
                                {PRIORITY_SYMBOLS[entry.priority]}
                              </span>
                            )}
                            <button
                              data-action-slot
                              onClick={(e) => { e.stopPropagation(); handleAnswer(entry); }}
                              title="Answer question"
                              className="p-1 rounded hover:bg-bujo-question/20 text-muted-foreground hover:text-bujo-question transition-colors opacity-0 group-hover:opacity-100"
                            >
                              <MessageCircle className="w-4 h-4" />
                            </button>
                            {entry.type !== 'cancelled' ? (
                              <button
                                data-action-slot
                                onClick={(e) => { e.stopPropagation(); handleCancel(entry); }}
                                title="Cancel entry"
                                className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <X className="w-4 h-4" />
                              </button>
                            ) : (
                              <button
                                data-action-slot
                                onClick={(e) => { e.stopPropagation(); handleUncancel(entry); }}
                                title="Uncancel entry"
                                className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <RotateCcw className="w-4 h-4" />
                              </button>
                            )}
                            <button
                              data-action-slot
                              onClick={(e) => { e.stopPropagation(); handleCyclePriority(entry); }}
                              title="Cycle priority"
                              className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning transition-colors opacity-0 group-hover:opacity-100"
                            >
                              <Flag className="w-4 h-4" />
                            </button>
                            <button
                              data-action-slot
                              onClick={(e) => { e.stopPropagation(); handleCycleType(entry); }}
                              title="Change type"
                              className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                            >
                              <RefreshCw className="w-4 h-4" />
                            </button>
                            <button
                              data-action-slot
                              onClick={(e) => { e.stopPropagation(); handleDelete(entry); }}
                              title="Delete entry"
                              className="p-1 rounded hover:bg-destructive/20 text-muted-foreground hover:text-destructive transition-colors opacity-0 group-hover:opacity-100"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {/* Answer Question Modal */}
      {questionToAnswer && (
        <AnswerQuestionModal
          isOpen={answerModalOpen}
          questionId={questionToAnswer.id}
          questionContent={questionToAnswer.content}
          onClose={() => { setAnswerModalOpen(false); setQuestionToAnswer(null); }}
          onAnswered={handleAnswerSubmitted}
        />
      )}
    </div>
  )
}
