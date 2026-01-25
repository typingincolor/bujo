import { useMemo, useState } from 'react'
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo'
import { cn } from '@/lib/utils'
import { ArrowLeft } from 'lucide-react'

interface EntryTreeProps {
  entries: Entry[]
  highlightedEntryId: number
  rootEntryId: number
}

const INDENT_PX = 16

export function EntryTree({ entries, highlightedEntryId, rootEntryId }: EntryTreeProps) {
  const entriesById = useMemo(() => new Map(entries.map(e => [e.id, e])), [entries])

  const { initialViewedId, initialStack } = useMemo(() => {
    // If highlighted entry has children, automatically navigate into it
    const hasChildren = entries.some(e => e.parentId === highlightedEntryId)
    if (hasChildren) {
      // Auto-navigate: view children, with parent as navigation root
      return {
        initialViewedId: highlightedEntryId,
        initialStack: [rootEntryId, highlightedEntryId]
      }
    } else {
      // No children: just show the highlighted entry
      return {
        initialViewedId: highlightedEntryId,
        initialStack: [highlightedEntryId]
      }
    }
  }, [highlightedEntryId, rootEntryId, entries])

  const [viewedEntryId, setViewedEntryId] = useState(initialViewedId)
  const [navigationStack, setNavigationStack] = useState<number[]>(initialStack)

  function buildPath(entryId: number): Entry[] {
    const path: Entry[] = []
    let current = entriesById.get(entryId)
    while (current) {
      path.unshift(current)
      current = current.parentId ? entriesById.get(current.parentId) : undefined
    }
    return path
  }

  function buildSubtree(parentId: number): Entry[] {
    const result: Entry[] = []
    const children = entries.filter(e => e.parentId === parentId)
    for (const child of children) {
      result.push(child)
      result.push(...buildSubtree(child.id))
    }
    return result
  }

  const pathToViewed = buildPath(viewedEntryId)
  const rootIndex = pathToViewed.findIndex(e => e.id === rootEntryId)
  const basePath = rootIndex >= 0 ? pathToViewed.slice(rootIndex) : pathToViewed

  // Determine what to show based on navigation state:
  // - If we've navigated (stack > 1): show viewed entry + its children (subtree view)
  // - If viewing root with no navigation: show root + all descendants
  // - Otherwise: show path from root to highlighted entry
  const visiblePath = navigationStack.length > 1
    ? [entriesById.get(viewedEntryId)!, ...buildSubtree(viewedEntryId)]
    : viewedEntryId === rootEntryId
      ? [entriesById.get(rootEntryId)!, ...buildSubtree(rootEntryId)]
      : basePath

  // Check if we can navigate back (navigation stack has more than one entry)
  const canNavigateBack = navigationStack.length > 1

  const handleNavigateToEntry = (entryId: number) => {
    const entry = entriesById.get(entryId)
    if (!entry) return
    // Only navigate if entry has children
    const hasChildren = entries.some(e => e.parentId === entryId)
    if (hasChildren) {
      setNavigationStack(prev => [...prev, entryId])
      setViewedEntryId(entryId)
    }
  }

  const handleNavigateBack = () => {
    if (navigationStack.length > 1) {
      const newStack = navigationStack.slice(0, -1)
      setNavigationStack(newStack)
      setViewedEntryId(newStack[newStack.length - 1])
    }
  }

  function calculateDepth(entry: Entry): number {
    let depth = 0
    let current = entry
    while (current.parentId && current.parentId !== rootEntryId) {
      const parent = entriesById.get(current.parentId)
      if (!parent) break
      depth++
      current = parent
    }
    return depth
  }

  return (
    <div data-testid="entry-tree" className="space-y-1">
      {canNavigateBack && (
        <button
          onClick={handleNavigateBack}
          aria-label="Back to parent"
          className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground mb-2 transition-colors"
        >
          <ArrowLeft className="w-3 h-3" />
          Back to parent
        </button>
      )}
      {visiblePath.map((entry, index) => {
        const isHighlighted = entry.id === highlightedEntryId
        const symbol = ENTRY_SYMBOLS[entry.type] || '-'
        const hasChildren = entries.some(e => e.parentId === entry.id)

        // Calculate indent based on display mode:
        // - Navigation subtree: depth from viewed entry (first item = 0)
        // - Root subtree: depth from root
        // - Path mode: index in path
        const indent = navigationStack.length > 1
          ? index * INDENT_PX
          : viewedEntryId === rootEntryId
            ? calculateDepth(entry) * INDENT_PX
            : index * INDENT_PX

        return (
          <div
            key={entry.id}
            data-testid={`entry-tree-item-${entry.id}`}
            data-highlighted={isHighlighted ? 'true' : undefined}
            style={{ paddingLeft: `${indent}px` }}
            className={cn(
              'py-1 px-2 rounded text-sm',
              isHighlighted ? 'bg-primary/10 font-medium' : 'text-muted-foreground',
              hasChildren && 'cursor-pointer hover:bg-muted/50'
            )}
            onClick={() => hasChildren && handleNavigateToEntry(entry.id)}
          >
            <span className="mr-2">{symbol}</span>
            {entry.content}
          </div>
        )
      })}
    </div>
  )
}
