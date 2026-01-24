import { useMemo } from 'react'
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo'
import { cn } from '@/lib/utils'

interface EntryTreeProps {
  entries: Entry[]
  highlightedEntryId: number
  rootEntryId: number
}

const INDENT_PX = 16

export function EntryTree({ entries, highlightedEntryId, rootEntryId }: EntryTreeProps) {
  const entriesById = useMemo(() => new Map(entries.map(e => [e.id, e])), [entries])

  function buildPath(entryId: number): Entry[] {
    const path: Entry[] = []
    let current = entriesById.get(entryId)
    while (current) {
      path.unshift(current)
      current = current.parentId ? entriesById.get(current.parentId) : undefined
    }
    return path
  }

  const pathToHighlighted = buildPath(highlightedEntryId)
  const rootIndex = pathToHighlighted.findIndex(e => e.id === rootEntryId)
  const visiblePath = rootIndex >= 0 ? pathToHighlighted.slice(rootIndex) : pathToHighlighted

  return (
    <div data-testid="entry-tree" className="space-y-1">
      {visiblePath.map((entry, index) => {
        const isHighlighted = entry.id === highlightedEntryId
        const symbol = ENTRY_SYMBOLS[entry.type] || '-'

        return (
          <div
            key={entry.id}
            data-testid={`entry-tree-item-${entry.id}`}
            style={{ paddingLeft: `${index * INDENT_PX}px` }}
            className={cn(
              'py-1 px-2 rounded text-sm',
              isHighlighted ? 'bg-primary/10 font-medium' : 'text-muted-foreground'
            )}
          >
            <span className="font-mono mr-2">{symbol}</span>
            {entry.content}
          </div>
        )
      })}
    </div>
  )
}
