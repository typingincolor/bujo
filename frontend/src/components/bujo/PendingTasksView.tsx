import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { EntryActionBar } from './EntryActions/EntryActionBar';
import { cn } from '@/lib/utils';
import { useAttentionScores, AttentionScore } from '@/hooks/useAttentionScores';
import { RefreshCw } from 'lucide-react';

export interface EntryCallbacks {
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

interface PendingTasksViewProps {
  overdueEntries: Entry[];
  callbacks: EntryCallbacks;
  selectedEntry?: Entry;
  onSelectEntry: (entry: Entry) => void;
  onNavigateToEntry?: (entry: Entry) => void;
  onRefresh: () => void;
}

export function PendingTasksView({
  overdueEntries,
  callbacks,
  selectedEntry,
  onSelectEntry,
  onNavigateToEntry,
  onRefresh,
}: PendingTasksViewProps) {
  const [localStatusOverrides, setLocalStatusOverrides] = useState<Map<number, Entry['type']>>(new Map());
  const selectedIndexRef = useRef(-1);

  const taskEntries = useMemo(() => {
    const filtered = overdueEntries.filter(e => e.type === 'task');
    return filtered.map(entry => {
      const override = localStatusOverrides.get(entry.id);
      return override ? { ...entry, type: override } : entry;
    });
  }, [overdueEntries, localStatusOverrides]);

  const taskIds = useMemo(() => taskEntries.map(e => e.id), [taskEntries]);
  const { scores } = useAttentionScores(taskIds);

  const sortedTaskEntries = useMemo(() => {
    return [...taskEntries].sort((a, b) => {
      const scoreA = scores[a.id]?.score ?? 0;
      const scoreB = scores[b.id]?.score ?? 0;
      return scoreB - scoreA;
    });
  }, [taskEntries, scores]);

  const createEntryCallbacks = useCallback((entry: Entry) => ({
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
  }), [callbacks]);

  const handleRefresh = useCallback(() => {
    setLocalStatusOverrides(new Map());
    onRefresh();
  }, [onRefresh]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return;

      if (e.key === 'j' || e.key === 'ArrowDown') {
        e.preventDefault();
        if (sortedTaskEntries.length === 0) return;
        const nextIndex = Math.min(selectedIndexRef.current + 1, sortedTaskEntries.length - 1);
        selectedIndexRef.current = nextIndex;
        onSelectEntry(sortedTaskEntries[nextIndex]);
      }

      if (e.key === 'k' || e.key === 'ArrowUp') {
        e.preventDefault();
        if (sortedTaskEntries.length === 0) return;
        const prevIndex = Math.max(selectedIndexRef.current - 1, 0);
        selectedIndexRef.current = prevIndex;
        onSelectEntry(sortedTaskEntries[prevIndex]);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [sortedTaskEntries, onSelectEntry]);

  // Sync selectedIndexRef when selectedEntry changes
  useEffect(() => {
    if (selectedEntry) {
      const idx = sortedTaskEntries.findIndex(e => e.id === selectedEntry.id);
      if (idx !== -1) selectedIndexRef.current = idx;
    } else {
      selectedIndexRef.current = -1;
    }
  }, [selectedEntry, sortedTaskEntries]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-lg font-semibold">Pending Tasks ({taskEntries.length})</h2>
        <button
          onClick={handleRefresh}
          title="Refresh pending tasks"
          className="p-1.5 hover:bg-secondary rounded-md transition-colors"
        >
          <RefreshCw className="h-4 w-4" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-1">
        {taskEntries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No pending tasks</p>
        ) : (
          sortedTaskEntries.map((entry) => (
            <PendingTaskItem
              key={entry.id}
              entry={entry}
              attentionScore={scores[entry.id]}
              isSelected={selectedEntry?.id === entry.id}
              onSelect={() => onSelectEntry(entry)}
              onDoubleClick={() => onNavigateToEntry?.(entry)}
              callbacks={createEntryCallbacks(entry)}
            />
          ))
        )}
      </div>
    </div>
  );
}

interface PendingTaskItemProps {
  entry: Entry;
  attentionScore?: AttentionScore;
  isSelected: boolean;
  onSelect: () => void;
  onDoubleClick?: () => void;
  callbacks: Record<string, (() => void) | undefined>;
}

function PendingTaskItem({ entry, attentionScore, isSelected, onSelect, onDoubleClick, callbacks }: PendingTaskItemProps) {
  const [isHovered, setIsHovered] = useState(false);
  const score = attentionScore?.score ?? 0;
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];

  return (
    <div
      className={cn(
        'group flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors',
        !isSelected && isHovered && 'bg-secondary/50',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <button
        onClick={onSelect}
        onDoubleClick={onDoubleClick}
        className="flex items-center gap-2 text-left min-w-0 flex-1"
      >
        <span data-testid="entry-symbol" className="text-muted-foreground flex-shrink-0">
          {symbol}
        </span>

        <span
          data-testid="attention-badge"
          className={cn(
            'px-1.5 py-0.5 rounded text-xs font-medium text-white flex-shrink-0',
            score >= 80 ? 'bg-red-500' :
            score >= 50 ? 'bg-orange-500' : 'bg-yellow-500'
          )}
        >
          {score}
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
      </button>

      {/* Inline action bar - visible on hover or selection */}
      <div className="flex-shrink-0">
        <EntryActionBar
          entry={entry}
          callbacks={callbacks}
          variant="hover-reveal"
          size="sm"
          isSelected={isSelected}
          isHovered={isHovered}
        />
      </div>
    </div>
  );
}
