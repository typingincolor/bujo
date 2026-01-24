import { useState, useEffect, useCallback } from 'react';
import { Entry, EntryType } from '@/types/bujo';
import { EntrySymbol } from './EntrySymbol';
import { EntryActionBar } from './EntryActions';
import { cn } from '@/lib/utils';
import { calculateMenuPosition } from '@/lib/menuPosition';
import { ChevronRight, ChevronDown } from 'lucide-react';

interface EntryItemProps {
  entry: Entry;
  depth?: number;
  isCollapsed?: boolean;
  hasChildren?: boolean;
  hasParent?: boolean;
  childCount?: number;
  isSelected?: boolean;
  disableClick?: boolean;
  onToggleCollapse?: () => void;
  onToggleDone?: () => void;
  onSelect?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  onAnswer?: () => void;
  onCancel?: () => void;
  onUncancel?: () => void;
  onCyclePriority?: () => void;
  onMigrate?: () => void;
  onAddChild?: () => void;
  onCycleType?: () => void;
  onMoveToRoot?: () => void;
  onMoveToList?: () => void;
}

const CONTEXT_MENU_ESTIMATED_SIZE = { width: 150, height: 300 };

const contentStyles: Record<EntryType, string> = {
  task: '',
  note: 'text-muted-foreground italic',
  event: 'font-medium',
  done: 'text-bujo-done',
  migrated: 'text-muted-foreground',
  cancelled: 'line-through text-muted-foreground opacity-60',
  question: 'text-bujo-question font-medium',
  answered: 'text-bujo-answered',
  answer: 'text-muted-foreground pl-4',
};

export function EntryItem({
  entry,
  depth = 0,
  isCollapsed = false,
  hasChildren = false,
  hasParent = false,
  childCount = 0,
  isSelected = false,
  disableClick = false,
  onToggleCollapse,
  onToggleDone,
  onSelect,
  onEdit,
  onDelete,
  onAnswer,
  onCancel,
  onUncancel,
  onCyclePriority,
  onMigrate,
  onAddChild,
  onCycleType,
  onMoveToRoot,
  onMoveToList,
}: EntryItemProps) {
  const isToggleable = entry.type === 'task' || entry.type === 'done';
  const canChangeType = entry.type === 'task' || entry.type === 'note' || entry.type === 'event' || entry.type === 'question';
  const canEdit = entry.type !== 'cancelled';
  const [contextMenuPos, setContextMenuPos] = useState<{ x: number; y: number } | null>(null);

  const closeContextMenu = useCallback(() => {
    setContextMenuPos(null);
  }, []);

  useEffect(() => {
    if (!contextMenuPos) return;

    const handleClickOutside = () => closeContextMenu();
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeContextMenu();
    };

    document.addEventListener('click', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('click', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [contextMenuPos, closeContextMenu]);

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    const adjusted = calculateMenuPosition(
      { x: e.clientX, y: e.clientY },
      CONTEXT_MENU_ESTIMATED_SIZE,
      { width: window.innerWidth, height: window.innerHeight }
    );
    setContextMenuPos(adjusted);
  };

  const handleClick = () => {
    onSelect?.();
  };

  return (
    <div
      data-testid="entry-item"
      data-entry-id={entry.id}
      data-selected={isSelected}
      className={cn(
        'group flex items-center gap-2 py-1.5 px-2 rounded-md transition-colors relative font-mono',
        'animate-fade-in',
        !isSelected && 'hover:bg-secondary/50',
        isToggleable && !disableClick && 'cursor-pointer',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      style={{ paddingLeft: `${depth * 20 + 8}px`, fontFamily: 'monospace' }}
      onClick={disableClick ? undefined : handleClick}
      onContextMenu={handleContextMenu}
    >
      {/* Collapse indicator */}
      {hasChildren ? (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onToggleCollapse?.();
          }}
          className="w-4 h-4 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
        >
          {isCollapsed ? (
            <ChevronRight className="w-3.5 h-3.5" />
          ) : (
            <ChevronDown className="w-3.5 h-3.5" />
          )}
        </button>
      ) : (
        <span className="w-4" />
      )}

      {/* Symbol - clickable for task/done entries */}
      {isToggleable && onToggleDone ? (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onToggleDone();
          }}
          className="cursor-pointer hover:opacity-70 transition-opacity"
          title={entry.type === 'task' ? 'Mark as done' : 'Mark as not done'}
        >
          <EntrySymbol type={entry.type} priority={entry.priority} />
        </button>
      ) : (
        <EntrySymbol type={entry.type} priority={entry.priority} />
      )}

      {/* Content */}
      <span className={cn('flex-1 text-sm', contentStyles[entry.type])}>
        {entry.content}
      </span>

      {/* Hidden child count when collapsed */}
      {isCollapsed && childCount > 0 && (
        <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
          {childCount} hidden
        </span>
      )}

      {/* Action buttons (shown on hover) */}
      <EntryActionBar
        entry={entry}
        callbacks={{
          onAnswer,
          onCancel,
          onUncancel,
          onCyclePriority,
          onCycleType,
          onMigrate,
          onMoveToList,
          onEdit,
          onDelete,
        }}
        variant="hover-reveal"
        size="sm"
      />

      {/* Entry ID (shown on hover) */}
      <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
        #{entry.id}
      </span>

      {/* Context menu */}
      {contextMenuPos && (
        <div
          role="menu"
          className="fixed z-50 bg-popover border rounded-md shadow-lg py-1 min-w-[120px]"
          style={{ left: contextMenuPos.x, top: contextMenuPos.y }}
          onClick={(e) => e.stopPropagation()}
        >
          {onToggleDone && entry.type === 'task' && (
            <button
              role="menuitem"
              onClick={() => {
                onToggleDone();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Mark done
            </button>
          )}
          {onToggleDone && entry.type === 'done' && (
            <button
              role="menuitem"
              onClick={() => {
                onToggleDone();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Mark not done
            </button>
          )}
          {onCancel && entry.type !== 'cancelled' && (
            <button
              role="menuitem"
              onClick={() => {
                onCancel();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Cancel
            </button>
          )}
          {onUncancel && entry.type === 'cancelled' && (
            <button
              role="menuitem"
              onClick={() => {
                onUncancel();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Uncancel
            </button>
          )}
          {onMigrate && entry.type === 'task' && (
            <button
              role="menuitem"
              onClick={() => {
                onMigrate();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Migrate
            </button>
          )}
          {onMoveToList && entry.type === 'task' && (
            <button
              role="menuitem"
              onClick={() => {
                onMoveToList();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Move to list
            </button>
          )}
          {onCycleType && canChangeType && (
            <button
              role="menuitem"
              onClick={() => {
                onCycleType();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Change type
            </button>
          )}
          {onCyclePriority && (
            <button
              role="menuitem"
              onClick={() => {
                onCyclePriority();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Cycle priority
            </button>
          )}
          {onMoveToRoot && hasParent && (
            <button
              role="menuitem"
              onClick={() => {
                onMoveToRoot();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Move to root
            </button>
          )}
          {onAddChild && entry.type !== 'question' && (
            <button
              role="menuitem"
              onClick={() => {
                onAddChild();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Add child
            </button>
          )}
          {onEdit && canEdit && (
            <button
              role="menuitem"
              onClick={() => {
                onEdit();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors"
            >
              Edit
            </button>
          )}
          {onDelete && (
            <button
              role="menuitem"
              onClick={() => {
                onDelete();
                closeContextMenu();
              }}
              className="w-full text-left px-3 py-1.5 text-sm hover:bg-destructive/20 text-destructive transition-colors"
            >
              Delete
            </button>
          )}
        </div>
      )}
    </div>
  );
}
