import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { EntryActionBar } from './EntryActions/EntryActionBar';
import { cn } from '@/lib/utils';
import { calculateAttentionScore } from '@/lib/attentionScore';
import { ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react';
import { buildTree } from '@/lib/buildTree';
import { ContextTree } from './ContextTree';

interface EntryCallbacks {
  onMarkDone?: () => void;
  onUnmarkDone?: () => void;
  onMigrate?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  onCyclePriority?: () => void;
  onCycleType?: () => void;
  onMoveToList?: () => void;
  onCancel?: () => void;
  onUncancel?: () => void;
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

        <span className={cn(
          "flex-1 truncate",
          entry.type === 'cancelled' && "line-through text-muted-foreground"
        )}>
          {entry.content}
        </span>

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

      {/* Action bar below entry - visible on hover or selection */}
      {(isHovered || isSelected) && (
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
      )}
    </div>
  );
}

export interface JournalSidebarCallbacks {
  onMarkDone?: (entry: Entry) => void;
  onUnmarkDone?: (entry: Entry) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
  onCyclePriority?: (entry: Entry) => void;
  onCycleType?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
  onCancel?: (entry: Entry) => void;
  onUncancel?: (entry: Entry) => void;
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
  onRefresh?: () => void;
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
  onRefresh,
}: JournalSidebarProps) {
  const [sidebarWidth, setSidebarWidth] = useState(512);
  const [isResizing, setIsResizing] = useState(false);
  const [snapshotEntries, setSnapshotEntries] = useState<Entry[]>([]);
  // Track local status changes (optimistic updates) - cleared on snapshot refresh
  const [localStatusOverrides, setLocalStatusOverrides] = useState<Map<number, Entry['type']>>(new Map());
  const wasCollapsedRef = useRef(isCollapsed);

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

  // Helper to get filtered task entries from overdue entries
  const getTaskEntries = useCallback(() => {
    return overdueEntries.filter(e => e.type === 'task');
  }, [overdueEntries]);

  // Refresh the snapshot manually
  const refreshSnapshot = useCallback(() => {
    setSnapshotEntries(getTaskEntries());
    setLocalStatusOverrides(new Map());
  }, [getTaskEntries]);

  // Track whether we have a valid snapshot
  const hasSnapshotRef = useRef(false);
  // Track whether we're waiting for fresh data after re-expand
  const awaitingFreshDataRef = useRef(false);
  // Track if sidebar has ever been expanded (to distinguish first expand from re-expand)
  const hasEverExpandedRef = useRef(false);
  // Track the overdueEntries reference to detect when fresh data arrives
  const lastEntriesRef = useRef(overdueEntries);

  // Handle sidebar expand/collapse transitions
  useEffect(() => {
    const wasCollapsed = wasCollapsedRef.current;
    wasCollapsedRef.current = isCollapsed;

    if (wasCollapsed && !isCollapsed) {
      // Expanding
      if (!hasEverExpandedRef.current) {
        // First expand: capture snapshot immediately with current data
        setSnapshotEntries(overdueEntries.filter(e => e.type === 'task'));
        hasSnapshotRef.current = true;
        hasEverExpandedRef.current = true;
        onRefresh?.();
      } else {
        // Re-expand after collapse: wait for fresh data
        awaitingFreshDataRef.current = true;
        hasSnapshotRef.current = false;
        onRefresh?.();
      }
    } else if (!wasCollapsed && isCollapsed) {
      // Collapsing: clear snapshot and overrides so next expand gets fresh data
      setSnapshotEntries([]);
      setLocalStatusOverrides(new Map());
      hasSnapshotRef.current = false;
      awaitingFreshDataRef.current = false;
    }
  }, [isCollapsed, onRefresh, overdueEntries]);

  // Capture snapshot when fresh data arrives after re-expand
  useEffect(() => {
    if (awaitingFreshDataRef.current && overdueEntries !== lastEntriesRef.current) {
      // Fresh data arrived, capture snapshot
      setSnapshotEntries(overdueEntries.filter(e => e.type === 'task'));
      hasSnapshotRef.current = true;
      awaitingFreshDataRef.current = false;
    }
    lastEntriesRef.current = overdueEntries;
  }, [overdueEntries]);

  // Handle initial state when sidebar starts expanded (not a transition)
  useEffect(() => {
    if (!isCollapsed && !hasSnapshotRef.current && !awaitingFreshDataRef.current && overdueEntries.length > 0) {
      setSnapshotEntries(overdueEntries.filter(e => e.type === 'task'));
      hasSnapshotRef.current = true;
      hasEverExpandedRef.current = true;
    }
  }, [isCollapsed, overdueEntries]);

  // Use snapshot if we have one, otherwise use live filtered entries
  // This ensures backwards compatibility and handles initial state
  const baseEntries = snapshotEntries.length > 0 ? snapshotEntries : getTaskEntries();
  // Apply local status overrides (optimistic updates)
  const taskEntries = baseEntries.map(entry => {
    const override = localStatusOverrides.get(entry.id);
    return override ? { ...entry, type: override } : entry;
  });

  const createEntryCallbacks = (entry: Entry) => ({
    onMarkDone: callbacks.onMarkDone ? () => {
      setLocalStatusOverrides(prev => new Map(prev).set(entry.id, 'done'));
      callbacks.onMarkDone!(entry);
    } : undefined,
    onUnmarkDone: callbacks.onUnmarkDone ? () => {
      setLocalStatusOverrides(prev => new Map(prev).set(entry.id, 'task'));
      callbacks.onUnmarkDone!(entry);
    } : undefined,
    onMigrate: callbacks.onMigrate ? () => callbacks.onMigrate!(entry) : undefined,
    onEdit: callbacks.onEdit ? () => callbacks.onEdit!(entry) : undefined,
    onDelete: callbacks.onDelete ? () => callbacks.onDelete!(entry) : undefined,
    onCyclePriority: callbacks.onCyclePriority ? () => callbacks.onCyclePriority!(entry) : undefined,
    onCycleType: callbacks.onCycleType ? () => callbacks.onCycleType!(entry) : undefined,
    onMoveToList: callbacks.onMoveToList ? () => callbacks.onMoveToList!(entry) : undefined,
    onCancel: callbacks.onCancel ? () => {
      setLocalStatusOverrides(prev => new Map(prev).set(entry.id, 'cancelled'));
      callbacks.onCancel!(entry);
    } : undefined,
    onUncancel: callbacks.onUncancel ? () => {
      setLocalStatusOverrides(prev => new Map(prev).set(entry.id, 'task'));
      callbacks.onUncancel!(entry);
    } : undefined,
  });

  return (
    <div
      data-testid="overdue-sidebar"
      className={cn(
        "flex flex-col h-full relative",
        isResizing && "select-none"
      )}
      style={{ width: isCollapsed ? '100%' : `${sidebarWidth}px` }}
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
          className={cn(
            "absolute top-2 p-1.5 hover:bg-secondary rounded-md transition-colors z-10",
            isCollapsed ? "left-1/2 -translate-x-1/2" : "right-2"
          )}
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
              <button
                onClick={refreshSnapshot}
                title="Refresh pending tasks"
                className="p-1 hover:bg-secondary rounded-md transition-colors ml-auto"
              >
                <RefreshCw className="h-3.5 w-3.5" />
              </button>
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
