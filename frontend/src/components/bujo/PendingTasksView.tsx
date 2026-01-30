import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { JournalSidebarCallbacks } from './JournalSidebar';
import { EntryActionBar } from './EntryActions/EntryActionBar';
import { cn } from '@/lib/utils';
import { calculateAttentionScore } from '@/lib/attentionScore';
import { RefreshCw } from 'lucide-react';

interface PendingTasksViewProps {
  overdueEntries: Entry[];
  now: Date;
  callbacks: JournalSidebarCallbacks;
  selectedEntry?: Entry;
  onSelectEntry: (entry: Entry) => void;
  onRefresh: () => void;
}

export function PendingTasksView({
  overdueEntries,
  now,
  callbacks,
  selectedEntry,
  onSelectEntry,
  onRefresh,
}: PendingTasksViewProps) {
  const [localStatusOverrides, setLocalStatusOverrides] = useState<Map<number, Entry['type']>>(new Map());
  const selectedIndexRef = useRef(-1);

  const taskEntries = useMemo(() => {
    const filtered = overdueEntries.filter(e => e.type === 'task');
    const withOverrides = filtered.map(entry => {
      const override = localStatusOverrides.get(entry.id);
      return override ? { ...entry, type: override } : entry;
    });
    return withOverrides.sort((a, b) => {
      const scoreA = calculateAttentionScore(a, now).score;
      const scoreB = calculateAttentionScore(b, now).score;
      return scoreB - scoreA;
    });
  }, [overdueEntries, localStatusOverrides, now]);

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
        if (taskEntries.length === 0) return;
        const nextIndex = Math.min(selectedIndexRef.current + 1, taskEntries.length - 1);
        selectedIndexRef.current = nextIndex;
        onSelectEntry(taskEntries[nextIndex]);
      }

      if (e.key === 'k' || e.key === 'ArrowUp') {
        e.preventDefault();
        if (taskEntries.length === 0) return;
        const prevIndex = Math.max(selectedIndexRef.current - 1, 0);
        selectedIndexRef.current = prevIndex;
        onSelectEntry(taskEntries[prevIndex]);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [taskEntries, onSelectEntry]);

  // Sync selectedIndexRef when selectedEntry changes
  useEffect(() => {
    if (selectedEntry) {
      const idx = taskEntries.findIndex(e => e.id === selectedEntry.id);
      if (idx !== -1) selectedIndexRef.current = idx;
    } else {
      selectedIndexRef.current = -1;
    }
  }, [selectedEntry, taskEntries]);

  return (
    <div className="flex flex-col h-full max-w-3xl mx-auto">
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
          taskEntries.map((entry) => (
            <PendingTaskItem
              key={entry.id}
              entry={entry}
              now={now}
              isSelected={selectedEntry?.id === entry.id}
              onSelect={() => onSelectEntry(entry)}
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
  now: Date;
  isSelected: boolean;
  onSelect: () => void;
  callbacks: Record<string, (() => void) | undefined>;
}

function PendingTaskItem({ entry, now, isSelected, onSelect, callbacks }: PendingTaskItemProps) {
  const [isHovered, setIsHovered] = useState(false);
  const attentionResult = calculateAttentionScore(entry, now);
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
        className="flex items-center gap-2 text-left min-w-0 flex-1"
      >
        <span data-testid="entry-symbol" className="text-muted-foreground flex-shrink-0">
          {symbol}
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
