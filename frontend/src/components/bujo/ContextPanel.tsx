import { useMemo } from 'react'
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo'
import { cn } from '@/lib/utils'

const INDENT_SIZE_PX = 16

interface ContextPanelProps {
  selectedEntry: Entry | null
  entries: Entry[]
}

function buildAncestorPath(entry: Entry, entries: Entry[]): Entry[] {
  const entriesById = new Map(entries.map(e => [e.id, e]))
  const path: Entry[] = []
  let current: Entry | undefined = entry

  while (current) {
    path.unshift(current)
    if (current.parentId === null) break
    current = entriesById.get(current.parentId)
  }

  return path
}

export function ContextPanel({ selectedEntry, entries }: ContextPanelProps) {
  const path = useMemo(
    () => selectedEntry ? buildAncestorPath(selectedEntry, entries) : [],
    [selectedEntry, entries]
  )

  if (selectedEntry === null) {
    return null
  }

  if (path.length === 1) {
    const entry = path[0]
    return (
      <div className="p-4">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
          Selected Entry
        </h3>
        <div className="py-1 text-sm font-medium text-foreground">
          <span className="mr-2">{ENTRY_SYMBOLS[entry.type]}</span>
          <span>{entry.content}</span>
        </div>
        <p className="mt-2 text-muted-foreground text-xs">
          No parent context
        </p>
      </div>
    )
  }

  return (
    <div className="p-4">
      <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
        Context
      </h3>
      {path.map((entry, index) => {
        const isSelected = entry.id === selectedEntry.id
        const indent = index * INDENT_SIZE_PX

        return (
          <div
            key={entry.id}
            data-testid={`context-panel-item-${entry.id}`}
            data-highlighted={isSelected ? 'true' : 'false'}
            style={{ paddingLeft: `${indent}px` }}
            className={cn(
              'py-1 text-sm',
              isSelected && 'font-medium text-foreground',
              !isSelected && 'text-muted-foreground'
            )}
          >
            <span className="mr-2">{ENTRY_SYMBOLS[entry.type]}</span>
            <span>{entry.content}</span>
          </div>
        )
      })}
    </div>
  )
}
