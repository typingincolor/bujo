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
  showContextDot?: boolean;
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
  movedToList: 'text-muted-foreground',
};

export function EntryItem({
  entry,
  depth = 0,
  isCollapsed = false,
  hasChildren = false,
  hasParent = false,
  childCount = 0,
  showContextDot = true,
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
  const [isHovered, setIsHovered] = useState(false);
  const [isKeyboardNavigating, setIsKeyboardNavigating] = useState(false);

  // Clear hover state when keyboard navigation occurs (arrow keys)
  // Also temporarily block mouse hover to prevent immediate re-hover
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown' || e.key === 'j' || e.key === 'k') {
        setIsHovered(false);
        setIsKeyboardNavigating(true);
      }
    };
    const handleMouseMove = () => {
      // Re-enable hover after any mouse movement following keyboard navigation
      setIsKeyboardNavigating(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('mousemove', handleMouseMove);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('mousemove', handleMouseMove);
    };
  }, []);

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
        'group flex items-center gap-2 py-1.5 px-2 rounded-md transition-colors relative',
        'animate-fade-in',
        !isSelected && isHovered && 'bg-secondary/50',
        isToggleable && !disableClick && 'cursor-pointer',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      style={{ paddingLeft: `${depth * 20 + 8}px` }}
      onClick={disableClick ? undefined : handleClick}
      onContextMenu={handleContextMenu}
      onMouseEnter={() => !isKeyboardNavigating && setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
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

      {/* Context dot - indicates entry has ancestors */}
      {showContextDot && (
        entry.parentId !== null ? (
          <span
            data-testid="context-dot"
            className="w-1.5 h-1.5 rounded-full bg-muted-foreground flex-shrink-0"
            title="Has parent context"
          />
        ) : (
          <span className="w-1.5" />
        )
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

      {/* Action buttons (shown on hover or when selected) */}
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
        isSelected={isSelected}
        isHovered={isHovered}
      />

      {/* Entry ID (shown on hover) */}
      <span className={cn(
        'text-xs text-muted-foreground transition-opacity',
        isHovered ? 'opacity-100' : 'opacity-0'
      )}>
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
