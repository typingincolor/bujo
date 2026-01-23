import { ReactNode, useState, useCallback, useEffect } from 'react'
import * as Popover from '@radix-ui/react-popover'
import { Entry } from '@/types/bujo'
import { EntryTree } from './EntryTree'

type ActionType = 'done' | 'cancel' | 'priority' | 'migrate'

interface EntryContextPopoverProps {
  entry: Entry
  entries: Entry[]
  onAction: (entry: Entry, action: ActionType) => void
  onNavigate: (entry: Entry) => void
  children: ReactNode
}

function getAvailableActions(entry: Entry): ActionType[] {
  switch (entry.type) {
    case 'task':
      return ['done', 'priority', 'migrate']
    case 'question':
      return ['done', 'priority']
    case 'done':
      return ['cancel'] // undo
    case 'event':
    case 'note':
      return ['priority']
    default:
      return []
  }
}

function findRootId(entry: Entry, entries: Entry[]): number {
  const entriesById = new Map(entries.map(e => [e.id, e]))
  let current = entry
  while (current.parentId) {
    const parent = entriesById.get(current.parentId)
    if (!parent) break
    current = parent
  }
  return current.id
}

export function EntryContextPopover({
  entry,
  entries,
  onAction,
  onNavigate,
  children,
}: EntryContextPopoverProps) {
  const [open, setOpen] = useState(false)
  const availableActions = getAvailableActions(entry)
  const rootId = findRootId(entry, entries)

  const handleAction = useCallback((action: ActionType) => {
    onAction(entry, action)
    if (action === 'done' || action === 'cancel') {
      setOpen(false)
    }
  }, [entry, onAction])

  const handleNavigate = useCallback(() => {
    onNavigate(entry)
    setOpen(false)
  }, [entry, onNavigate])

  useEffect(() => {
    if (!open) return

    function handleKeyDown(e: KeyboardEvent) {
      switch (e.key) {
        case ' ':
          e.preventDefault()
          if (availableActions.includes('done')) handleAction('done')
          break
        case 'x':
          if (availableActions.includes('cancel')) handleAction('cancel')
          break
        case 'p':
          if (availableActions.includes('priority')) handleAction('priority')
          break
        case 'm':
          if (availableActions.includes('migrate')) handleAction('migrate')
          break
        case 'Enter':
          e.preventDefault()
          handleNavigate()
          break
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [open, availableActions, handleAction, handleNavigate])

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger asChild>
        {children}
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          data-testid="entry-context-popover"
          className="z-50 w-80 max-h-[400px] overflow-auto rounded-lg border border-border bg-card p-3 shadow-lg"
          sideOffset={4}
          collisionPadding={16}
        >
          <EntryTree
            entries={entries}
            highlightedEntryId={entry.id}
            rootEntryId={rootId}
          />

          <div className="mt-3 pt-3 border-t border-border flex items-center justify-between">
            <div className="flex gap-1">
              {availableActions.includes('done') && (
                <button
                  onClick={() => handleAction('done')}
                  aria-label="Mark done"
                  className="p-2 rounded hover:bg-muted"
                  title="Done (Space)"
                >
                  ✓
                </button>
              )}
              {availableActions.includes('cancel') && (
                <button
                  onClick={() => handleAction('cancel')}
                  aria-label="Cancel"
                  className="p-2 rounded hover:bg-muted"
                  title="Cancel (x)"
                >
                  ✕
                </button>
              )}
              {availableActions.includes('priority') && (
                <button
                  onClick={() => handleAction('priority')}
                  aria-label="Cycle priority"
                  className="p-2 rounded hover:bg-muted"
                  title="Priority (p)"
                >
                  !
                </button>
              )}
              {availableActions.includes('migrate') && (
                <button
                  onClick={() => handleAction('migrate')}
                  aria-label="Migrate"
                  className="p-2 rounded hover:bg-muted"
                  title="Migrate (m)"
                >
                  &gt;
                </button>
              )}
            </div>

            <button
              onClick={handleNavigate}
              className="text-sm text-primary hover:underline"
            >
              Go to entry
            </button>
          </div>

          <Popover.Arrow className="fill-border" />
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  )
}
