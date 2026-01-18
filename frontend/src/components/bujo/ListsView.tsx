import { BujoList } from '@/types/bujo'
import { cn } from '@/lib/utils'
import { List, CheckCircle2, Circle, ChevronRight, Plus, Trash2, Pencil, X, Ban, RotateCcw } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'
import { MarkListItemDone, MarkListItemUndone, AddListItem, RemoveListItem, CreateList, DeleteList, RenameList, EditListItem, CancelListItem, UncancelListItem } from '@/wailsjs/go/wails/App'
import { ConfirmDialog } from './ConfirmDialog'

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
  onDeleteList: (listId: number) => void
  onRenameList: (listId: number, newName: string) => void
  onEditItem: (itemId: number, content: string) => void
  onCancelItem: (itemId: number) => void
  onUncancelItem: (itemId: number) => void
}

function ListCard({ list, isExpanded, onToggle, onToggleItem, onAddItem, onDeleteItem, onDeleteList, onRenameList, onEditItem, onCancelItem, onUncancelItem }: ListCardProps) {
  const [newItemContent, setNewItemContent] = useState('')
  const [isRenaming, setIsRenaming] = useState(false)
  const [renameName, setRenameName] = useState(list.name)
  const [editingItemId, setEditingItemId] = useState<number | null>(null)
  const [editingContent, setEditingContent] = useState('')
  const renameInputRef = useRef<HTMLInputElement>(null)
  const editInputRef = useRef<HTMLInputElement>(null)

  const progress = list.totalCount > 0
    ? Math.round((list.doneCount / list.totalCount) * 100)
    : 0

  useEffect(() => {
    if (isRenaming) {
      renameInputRef.current?.focus()
      renameInputRef.current?.select()
    }
  }, [isRenaming])

  useEffect(() => {
    if (editingItemId !== null) {
      editInputRef.current?.focus()
      editInputRef.current?.select()
    }
  }, [editingItemId])

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

  const handleRenameClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    setRenameName(list.name)
    setIsRenaming(true)
  }

  const handleRenameKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      const trimmed = renameName.trim()
      if (trimmed && trimmed !== list.name) {
        onRenameList(list.id, trimmed)
      }
      setIsRenaming(false)
    } else if (e.key === 'Escape') {
      setIsRenaming(false)
    }
  }

  const handleDeleteListClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onDeleteList(list.id)
  }

  const handleEditItemClick = (e: React.MouseEvent, itemId: number, content: string) => {
    e.stopPropagation()
    setEditingItemId(itemId)
    setEditingContent(content)
  }

  const handleEditKeyDown = (e: React.KeyboardEvent, itemId: number) => {
    if (e.key === 'Enter') {
      const trimmed = editingContent.trim()
      if (trimmed) {
        onEditItem(itemId, trimmed)
      }
      setEditingItemId(null)
    } else if (e.key === 'Escape') {
      setEditingItemId(null)
    }
  }

  const handleCancelItem = (e: React.MouseEvent, itemId: number) => {
    e.stopPropagation()
    onCancelItem(itemId)
  }

  const handleUncancelItem = (e: React.MouseEvent, itemId: number) => {
    e.stopPropagation()
    onUncancelItem(itemId)
  }

  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden animate-fade-in">
      {/* Header */}
      <div className="flex items-center gap-3 p-4 hover:bg-secondary/30 transition-colors group">
        <button onClick={onToggle} className="flex-1 flex items-center gap-3">
          <ChevronRight
            className={cn(
              'w-4 h-4 text-muted-foreground transition-transform',
              isExpanded && 'rotate-90'
            )}
          />
          {isRenaming ? (
            <input
              ref={renameInputRef}
              type="text"
              value={renameName}
              onChange={(e) => setRenameName(e.target.value)}
              onKeyDown={handleRenameKeyDown}
              onBlur={() => setIsRenaming(false)}
              onClick={(e) => e.stopPropagation()}
              className="font-medium flex-1 text-left bg-transparent border-b border-primary focus:outline-none"
            />
          ) : (
            <span className="font-medium flex-1 text-left">{list.name}</span>
          )}
        </button>
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
        {/* Action buttons */}
        <button
          onClick={handleRenameClick}
          title="Rename list"
          className="p-1 rounded text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors opacity-0 group-hover:opacity-100"
        >
          <Pencil className="w-3.5 h-3.5" />
        </button>
        <button
          onClick={handleDeleteListClick}
          title="Delete list"
          className="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors opacity-0 group-hover:opacity-100"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Items */}
      {isExpanded && (
        <div className="border-t border-border px-4 py-2 space-y-1">
          {list.items.map((item) => (
            <div
              key={item.id}
              onClick={() => editingItemId !== item.id && onToggleItem(item.id, item.done)}
              className="flex items-center gap-3 py-1.5 group hover:bg-secondary/20 rounded px-2 -mx-2 cursor-pointer"
            >
              {item.type === 'done' ? (
                <CheckCircle2 className="w-4 h-4 text-bujo-done flex-shrink-0" />
              ) : item.type === 'cancelled' ? (
                <Ban className="w-4 h-4 text-muted-foreground flex-shrink-0" />
              ) : (
                <Circle className="w-4 h-4 text-muted-foreground flex-shrink-0" />
              )}
              {editingItemId === item.id ? (
                <input
                  ref={editInputRef}
                  type="text"
                  value={editingContent}
                  onChange={(e) => setEditingContent(e.target.value)}
                  onKeyDown={(e) => handleEditKeyDown(e, item.id)}
                  onBlur={() => setEditingItemId(null)}
                  onClick={(e) => e.stopPropagation()}
                  className="text-sm flex-1 bg-transparent border-b border-primary focus:outline-none"
                />
              ) : (
                <span className={cn(
                  'text-sm flex-1',
                  (item.done || item.type === 'cancelled') && 'line-through text-muted-foreground'
                )}>
                  {item.content}
                </span>
              )}
              {item.type === 'task' && (
                <button
                  onClick={(e) => handleCancelItem(e, item.id)}
                  title="Cancel item"
                  className="p-1 rounded text-muted-foreground hover:text-orange-500 hover:bg-orange-500/10 transition-colors opacity-0 group-hover:opacity-100"
                >
                  <Ban className="w-3.5 h-3.5" />
                </button>
              )}
              {item.type === 'cancelled' && (
                <button
                  onClick={(e) => handleUncancelItem(e, item.id)}
                  title="Uncancel item"
                  className="p-1 rounded text-muted-foreground hover:text-green-500 hover:bg-green-500/10 transition-colors opacity-0 group-hover:opacity-100"
                >
                  <RotateCcw className="w-3.5 h-3.5" />
                </button>
              )}
              <button
                onClick={(e) => handleEditItemClick(e, item.id, item.content)}
                title="Edit item"
                className="p-1 rounded text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors opacity-0 group-hover:opacity-100"
              >
                <Pencil className="w-3.5 h-3.5" />
              </button>
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
  const [isCreatingList, setIsCreatingList] = useState(false)
  const [newListName, setNewListName] = useState('')
  const [listToDelete, setListToDelete] = useState<BujoList | null>(null)
  const createInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (isCreatingList) {
      createInputRef.current?.focus()
    }
  }, [isCreatingList])

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

  const handleCreateList = async () => {
    const trimmed = newListName.trim()
    if (!trimmed) return

    try {
      await CreateList(trimmed)
      setNewListName('')
      setIsCreatingList(false)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to create list:', error)
    }
  }

  const handleCreateKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleCreateList()
    } else if (e.key === 'Escape') {
      setIsCreatingList(false)
      setNewListName('')
    }
  }

  const handleDeleteList = async () => {
    if (!listToDelete) return

    try {
      await DeleteList(listToDelete.id, true)
      setListToDelete(null)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to delete list:', error)
    }
  }

  const handleRenameList = async (listId: number, newName: string) => {
    try {
      await RenameList(listId, newName)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to rename list:', error)
    }
  }

  const handleEditItem = async (itemId: number, content: string) => {
    try {
      await EditListItem(itemId, content)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to edit list item:', error)
    }
  }

  const handleCancelItem = async (itemId: number) => {
    try {
      await CancelListItem(itemId)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to cancel list item:', error)
    }
  }

  const handleUncancelItem = async (itemId: number) => {
    try {
      await UncancelListItem(itemId)
      onListChanged?.()
    } catch (error) {
      console.error('Failed to uncancel list item:', error)
    }
  }

  const handleRequestDeleteList = (listId: number) => {
    const list = lists.find(l => l.id === listId)
    if (list) {
      setListToDelete(list)
    }
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 mb-4">
        <List className="w-5 h-5 text-primary" />
        <h2 className="font-display text-xl font-semibold">Lists</h2>
        <button
          onClick={() => setIsCreatingList(true)}
          className="ml-auto px-2 py-1 text-xs rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors flex items-center gap-1"
          aria-label="New list"
        >
          <Plus className="w-3 h-3" />
          New List
        </button>
      </div>

      {isCreatingList && (
        <div className="flex items-center gap-2 py-2 px-4 rounded-lg bg-card border border-border animate-fade-in">
          <input
            ref={createInputRef}
            type="text"
            value={newListName}
            onChange={(e) => setNewListName(e.target.value)}
            onKeyDown={handleCreateKeyDown}
            placeholder="List name"
            className="flex-1 px-2 py-1.5 text-sm rounded-md border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
          <button
            onClick={() => { setIsCreatingList(false); setNewListName('') }}
            className="p-1.5 rounded-md hover:bg-secondary transition-colors"
            aria-label="Cancel"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

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
            onDeleteList={handleRequestDeleteList}
            onRenameList={handleRenameList}
            onEditItem={handleEditItem}
            onCancelItem={handleCancelItem}
            onUncancelItem={handleUncancelItem}
          />
        ))}
      </div>

      <ConfirmDialog
        isOpen={!!listToDelete}
        title="Delete List"
        message={`Are you sure you want to delete "${listToDelete?.name}"? This will also delete all items in this list.`}
        confirmText="Delete"
        variant="destructive"
        onConfirm={handleDeleteList}
        onCancel={() => setListToDelete(null)}
      />
    </div>
  )
}
