import { BujoList } from '@/types/bujo'
import { cn } from '@/lib/utils'
import { List, CheckCircle2, Circle, ChevronRight, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { MarkListItemDone, MarkListItemUndone, AddListItem, RemoveListItem } from '@/wailsjs/go/wails/App'

interface ListsViewProps {
  lists: BujoList[]
  onListChanged?: () => void
}

interface ListCardProps {
  list: BujoList
  isExpanded: boolean
  onToggle: () => void
  onToggleItem: (itemId: number, done: boolean) => void
  onAddItem: (listId: number, content: string) => void
  onDeleteItem: (itemId: number) => void
}

function ListCard({ list, isExpanded, onToggle, onToggleItem, onAddItem, onDeleteItem }: ListCardProps) {
  const [newItemContent, setNewItemContent] = useState('')
  const progress = list.totalCount > 0
    ? Math.round((list.doneCount / list.totalCount) * 100)
    : 0

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      const trimmed = newItemContent.trim()
      if (trimmed) {
        onAddItem(list.id, trimmed)
        setNewItemContent('')
      }
    }
  }

  const handleDeleteItem = (e: React.MouseEvent, itemId: number) => {
    e.stopPropagation()
    onDeleteItem(itemId)
  }

  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden animate-fade-in">
      {/* Header */}
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-3 p-4 hover:bg-secondary/30 transition-colors"
      >
        <ChevronRight
          className={cn(
            'w-4 h-4 text-muted-foreground transition-transform',
            isExpanded && 'rotate-90'
          )}
        />
        <span className="font-medium flex-1 text-left">{list.name}</span>
        <span className="text-sm text-muted-foreground">
          {list.doneCount}/{list.totalCount}
        </span>
        {/* Progress bar */}
        <div className="w-16 h-1.5 bg-muted rounded-full overflow-hidden">
          <div
            className="h-full bg-bujo-done transition-all"
            style={{ width: `${progress}%` }}
          />
        </div>
      </button>

      {/* Items */}
      {isExpanded && (
        <div className="border-t border-border px-4 py-2 space-y-1">
          {list.items.map((item) => (
            <div
              key={item.id}
              onClick={() => onToggleItem(item.id, item.done)}
              className="flex items-center gap-3 py-1.5 group hover:bg-secondary/20 rounded px-2 -mx-2 cursor-pointer"
            >
              {item.done ? (
                <CheckCircle2 className="w-4 h-4 text-bujo-done flex-shrink-0" />
              ) : (
                <Circle className="w-4 h-4 text-muted-foreground flex-shrink-0" />
              )}
              <span className={cn(
                'text-sm flex-1',
                item.done && 'line-through text-muted-foreground'
              )}>
                {item.content}
              </span>
              <button
                onClick={(e) => handleDeleteItem(e, item.id)}
                title="Delete item"
                className="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors opacity-0 group-hover:opacity-100"
              >
                <Trash2 className="w-3.5 h-3.5" />
              </button>
            </div>
          ))}
          {/* Add item input */}
          <div className="flex items-center gap-2 py-1.5 px-2 -mx-2">
            <Plus className="w-4 h-4 text-muted-foreground flex-shrink-0" />
            <input
              type="text"
              value={newItemContent}
              onChange={(e) => setNewItemContent(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Add item..."
              className="flex-1 text-sm bg-transparent border-none focus:outline-none placeholder:text-muted-foreground"
            />
          </div>
        </div>
      )}
    </div>
  )
}

export function ListsView({ lists, onListChanged }: ListsViewProps) {
  const [expandedIds, setExpandedIds] = useState<Set<number>>(
    () => lists[0]?.id !== undefined ? new Set([lists[0].id]) : new Set()
  )

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

  const handleToggleItem = async (itemId: number, currentlyDone: boolean) => {
    try {
      if (currentlyDone) {
        await MarkListItemUndone(itemId)
      } else {
        await MarkListItemDone(itemId)
      }
      onListChanged?.()
    } catch (error) {
      console.error('Failed to toggle list item:', error)
    }
  }

  const handleAddItem = async (listId: number, content: string) => {
    try {
      await AddListItem(listId, content)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to add list item:', error)
    }
  }

  const handleDeleteItem = async (itemId: number) => {
    try {
      await RemoveListItem(itemId)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to delete list item:', error)
    }
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 mb-4">
        <List className="w-5 h-5 text-primary" />
        <h2 className="font-display text-xl font-semibold">Lists</h2>
      </div>

      <div className="space-y-2">
        {lists.map((list) => (
          <ListCard
            key={list.id}
            list={list}
            isExpanded={expandedIds.has(list.id)}
            onToggle={() => toggleExpanded(list.id)}
            onToggleItem={handleToggleItem}
            onAddItem={handleAddItem}
            onDeleteItem={handleDeleteItem}
          />
        ))}
      </div>
    </div>
  )
}
