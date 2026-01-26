import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { EntryActionBar } from './EntryActions/EntryActionBar';
import { cn } from '@/lib/utils';
import { calculateAttentionScore } from '@/lib/attentionScore';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface EntryCallbacks {
  onCancel?: () => void;
  onMigrate?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  onCyclePriority?: () => void;
  onCycleType?: () => void;
  onMoveToList?: () => void;
}

interface OverdueEntryItemProps {
  entry: Entry;
  now: Date;
  isSelected: boolean;
  onSelect?: () => void;
  callbacks: EntryCallbacks;
}

function OverdueEntryItem({ entry, now, isSelected, onSelect, callbacks }: OverdueEntryItemProps) {
  const [isHovered, setIsHovered] = useState(false);
  const attentionResult = calculateAttentionScore(entry, now);
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];

  // Clear hover state when keyboard navigation occurs
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown' || e.key === 'j' || e.key === 'k') {
        setIsHovered(false);
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <div
      className={cn(
        'group px-2 py-1.5 rounded-lg text-sm transition-colors',
        !isSelected && isHovered && 'bg-secondary/50',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <button
        onClick={onSelect}
        className="flex items-center gap-2 text-left min-w-0 w-full"
      >
        <span data-testid="entry-symbol" className="text-muted-foreground flex-shrink-0">
          {symbol}
        </span>

        {prioritySymbol && (
          <span
            data-testid="priority-indicator"
            className="text-orange-500 font-medium flex-shrink-0"
          >
            {prioritySymbol}
          </span>
        )}

        <span className="flex-1 truncate">{entry.content}</span>

        <span
          data-testid="attention-badge"
          className={cn(
            'px-1.5 py-0.5 rounded text-xs font-medium text-white flex-shrink-0',
            attentionResult.score >= 80 ? 'bg-red-500' :
            attentionResult.score >= 50 ? 'bg-orange-500' : 'bg-yellow-500'
          )}
        >
          {attentionResult.score}
        </span>
      </button>

      {/* Action bar below entry - always visible */}
      <div
        className="pt-1"
        style={{ paddingLeft: 'calc(0.5rem + 0.5rem + 1ch)' }}
      >
        <EntryActionBar
          entry={entry}
          callbacks={callbacks}
          variant="always-visible"
          size="sm"
        />
      </div>
    </div>
  );
}

export interface JournalSidebarCallbacks {
  onMarkDone?: (entry: Entry) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
  onCyclePriority?: (entry: Entry) => void;
  onCycleType?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
}

interface JournalSidebarProps {
  overdueEntries?: Entry[];
  now?: Date;
  selectedEntry?: Entry;
  contextTree?: Entry[];
  onSelectEntry?: (entry: Entry) => void;
  callbacks?: JournalSidebarCallbacks;
  isCollapsed?: boolean;
  onToggleCollapse?: () => void;
  onWidthChange?: (width: number) => void;
  activelyCyclingEntry?: Entry;
  cyclingEntryPosition?: number;
}

interface TreeNode {
  entry: Entry;
  children: TreeNode[];
}

function buildTree(entries: Entry[]): TreeNode[] {
  if (entries.length === 0) return [];

  const entryMap = new Map<number, Entry>();
  const childrenMap = new Map<number | null, Entry[]>();

  for (const entry of entries) {
    entryMap.set(entry.id, entry);
    const parentId = entry.parentId;
    if (!childrenMap.has(parentId)) {
      childrenMap.set(parentId, []);
    }
    childrenMap.get(parentId)!.push(entry);
  }

  function buildNode(entry: Entry): TreeNode {
    const children = childrenMap.get(entry.id) || [];
    return {
      entry,
      children: children.map(buildNode),
    };
  }

  const roots = childrenMap.get(null) || [];
  return roots.map(buildNode);
}

interface ContextTreeProps {
  nodes: TreeNode[];
  selectedEntryId?: number;
  depth?: number;
}

function ContextTree({ nodes, selectedEntryId, depth = 0 }: ContextTreeProps) {
  return (
    <>
      {nodes.map((node) => (
        <div key={node.entry.id}>
          <div
            className={cn(
              'flex items-center gap-2 text-sm py-0.5 font-mono',
              node.entry.id === selectedEntryId
                ? 'font-medium'
                : 'text-muted-foreground'
            )}
            style={{ paddingLeft: `${depth * 12}px` }}
          >
            <span className="text-muted-foreground">
              {ENTRY_SYMBOLS[node.entry.type]}
            </span>
            <span className={cn(
              'truncate',
              node.entry.id === selectedEntryId && 'text-foreground'
            )}>
              {node.entry.content}
            </span>
          </div>
          {node.children.length > 0 && (
            <ContextTree
              nodes={node.children}
              selectedEntryId={selectedEntryId}
              depth={depth + 1}
            />
          )}
        </div>
      ))}
    </>
  );
}

export function JournalSidebar({
  overdueEntries = [],
  now = new Date(),
  selectedEntry,
  contextTree = [],
  onSelectEntry,
  callbacks = {},
  isCollapsed = false,
  onToggleCollapse,
  onWidthChange,
  activelyCyclingEntry,
  cyclingEntryPosition = -1,
}: JournalSidebarProps) {
  const [sidebarWidth, setSidebarWidth] = useState(512);
  const [isResizing, setIsResizing] = useState(false);

  const handleResizeMoveRef = useRef<(e: MouseEvent) => void>(() => {});
  const handleResizeEndRef = useRef<() => void>(() => {});

  const handleResizeMove = useCallback((e: MouseEvent) => {
    const newWidth = window.innerWidth - e.clientX;
    const clampedWidth = Math.max(384, Math.min(960, newWidth));
    setSidebarWidth(clampedWidth);
  }, []);

  const handleResizeEnd = useCallback(() => {
    setIsResizing(false);
    document.removeEventListener('mousemove', handleResizeMoveRef.current);
    document.removeEventListener('mouseup', handleResizeEndRef.current);
  }, []);

  useEffect(() => {
    handleResizeMoveRef.current = handleResizeMove;
    handleResizeEndRef.current = handleResizeEnd;
  }, [handleResizeMove, handleResizeEnd]);

  const handleResizeStart = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsResizing(true);
    document.addEventListener('mousemove', handleResizeMoveRef.current);
    document.addEventListener('mouseup', handleResizeEndRef.current);
  }, []);

  useEffect(() => {
    return () => {
      document.removeEventListener('mousemove', handleResizeMoveRef.current);
      document.removeEventListener('mouseup', handleResizeEndRef.current);
    };
  }, []);

  useEffect(() => {
    onWidthChange?.(sidebarWidth);
  }, [sidebarWidth, onWidthChange]);

  useEffect(() => {
    if (isResizing) {
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    } else {
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    }
  }, [isResizing]);

  const treeNodes = useMemo(() => buildTree(contextTree), [contextTree]);

  // Filter to only show task entries (not notes, events, questions, etc.)
  // Also include entry being actively cycled (tracked in parent) to give time to select the right type
  // IMPORTANT: Keep cycling entry in its original position using stored position index
  const taskEntries = useMemo(() => {
    // Filter entries, keeping cycling entry in its original position
    let result = overdueEntries
      .map((e) => {
        // If this is the cycling entry, use the updated version from state (preserves position)
        if (activelyCyclingEntry && e.id === activelyCyclingEntry.id) {
          return activelyCyclingEntry
        }
        return e
      })
      .filter((e) => {
        // Keep tasks, and keep the cycling entry even if it's not a task anymore
        const isTask = e.type === 'task'
        const isCycling = activelyCyclingEntry && e.id === activelyCyclingEntry.id
        return isTask || isCycling
      })

    // If cycling entry is not in result but should be, insert it at its stored position
    if (activelyCyclingEntry && cyclingEntryPosition >= 0 && !result.some(e => e.id === activelyCyclingEntry.id)) {
      // cyclingEntryPosition is now the position in the filtered task list, so use it directly
      const insertIndex = Math.min(cyclingEntryPosition, result.length)
      result = [...result.slice(0, insertIndex), activelyCyclingEntry, ...result.slice(insertIndex)]
    }

    return result
  }, [overdueEntries, activelyCyclingEntry, cyclingEntryPosition]);

  const createEntryCallbacks = (entry: Entry) => ({
    onCancel: callbacks.onMarkDone ? () => callbacks.onMarkDone!(entry) : undefined,
    onMigrate: callbacks.onMigrate ? () => callbacks.onMigrate!(entry) : undefined,
    onEdit: callbacks.onEdit ? () => callbacks.onEdit!(entry) : undefined,
    onDelete: callbacks.onDelete ? () => callbacks.onDelete!(entry) : undefined,
    onCyclePriority: callbacks.onCyclePriority ? () => callbacks.onCyclePriority!(entry) : undefined,
    onCycleType: callbacks.onCycleType ? () => callbacks.onCycleType!(entry) : undefined,
    onMoveToList: callbacks.onMoveToList ? () => callbacks.onMoveToList!(entry) : undefined,
  });

  return (
    <div
      data-testid="overdue-sidebar"
      className={cn(
        "flex flex-col h-full relative",
        isResizing && "select-none"
      )}
      style={{ width: `${sidebarWidth}px` }}
    >
      {/* Resize Handle */}
      {!isCollapsed && (
        <div
          data-testid="resize-handle"
          className="absolute left-0 top-0 h-full w-2 cursor-col-resize hover:bg-primary/10 transition-colors"
          onMouseDown={handleResizeStart}
        />
      )}

      {/* Collapse Toggle Button */}
      {onToggleCollapse && (
        <button
          onClick={onToggleCollapse}
          aria-label="Toggle sidebar"
          className="absolute top-2 right-2 p-1.5 hover:bg-secondary rounded-md transition-colors"
        >
          {isCollapsed ? (
            <ChevronLeft className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>
      )}

      {/* Content - hidden when collapsed */}
      {!isCollapsed && (
        <>
          {/* Pending Tasks Section */}
          <div>
            <div className="flex items-center gap-2 px-3 py-2 text-sm font-medium">
              <span>Pending Tasks ({taskEntries.length})</span>
            </div>

        <div className="px-1 py-1 max-h-80 overflow-y-auto">
          {taskEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground px-2 py-2">
              No pending tasks
            </p>
          ) : (
            taskEntries.map((entry) => (
              <OverdueEntryItem
                key={entry.id}
                entry={entry}
                now={now}
                isSelected={selectedEntry?.id === entry.id}
                onSelect={() => onSelectEntry?.(entry)}
                callbacks={createEntryCallbacks(entry)}
              />
            ))
          )}
        </div>
      </div>

          {/* Divider */}
          <hr className="my-4 border-border" />

          {/* Context Section */}
          <div data-testid="context-section" className="flex-1 flex flex-col min-h-0">
            <div className="flex items-center gap-2 px-3 py-2 text-sm font-medium">
              <span>Context</span>
            </div>

            <div className="px-3 py-2 flex-1 overflow-y-auto">
              {!selectedEntry ? (
                <p className="text-sm text-muted-foreground">No entry selected</p>
              ) : selectedEntry.parentId === null ? (
                <p className="text-sm text-muted-foreground">No context</p>
              ) : (
                <ContextTree
                  nodes={treeNodes}
                  selectedEntryId={selectedEntry.id}
                />
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
