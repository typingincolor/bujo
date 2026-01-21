import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Clock, ChevronDown, ChevronRight } from 'lucide-react';
import { ContextPill } from './ContextPill';
import { EntryActionBar } from './EntryActions';
import { format, parseISO } from 'date-fns';
import { useState, useEffect, useCallback, useMemo } from 'react';
import { MarkEntryDone, MarkEntryUndone, CancelEntry, UncancelEntry, DeleteEntry, CyclePriority, RetypeEntry } from '@/wailsjs/go/wails/App';

interface OverviewViewProps {
  overdueEntries: Entry[];
  onEntryChanged?: () => void;
  onError?: (message: string) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
  onNavigateToEntry?: (entry: Entry) => void;
}

function groupByDate(entries: Entry[]): Map<string, Entry[]> {
  const groups = new Map<string, Entry[]>();
  for (const entry of entries) {
    const date = entry.loggedDate.split('T')[0];
    if (!groups.has(date)) {
      groups.set(date, []);
    }
    groups.get(date)!.push(entry);
  }
  return groups;
}

function formatDateHeader(dateStr: string): string {
  try {
    const date = parseISO(dateStr);
    return format(date, 'MMM d');
  } catch {
    return dateStr;
  }
}

function buildParentChain(entry: Entry, entriesById: Map<number, Entry>): Entry[] {
  const chain: Entry[] = [];
  let current = entry;
  while (current.parentId !== null) {
    const parent = entriesById.get(current.parentId);
    if (!parent) break;
    chain.unshift(parent);
    current = parent;
  }
  return chain;
}

export function OverviewView({ overdueEntries, onEntryChanged, onError, onMigrate, onEdit, onMoveToList, onNavigateToEntry }: OverviewViewProps) {
  const [collapsed, setCollapsed] = useState(false);
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());
  const [selectedIndex, setSelectedIndex] = useState(-1);

  // Build a lookup map for all entries by ID
  const entriesById = new Map<number, Entry>();
  for (const entry of overdueEntries) {
    entriesById.set(entry.id, entry);
  }

  // Filter to only show task-related entries (task, done, or cancelled)
  const taskEntries = overdueEntries.filter(e => e.type === 'task' || e.type === 'done' || e.type === 'cancelled');
  const grouped = groupByDate(taskEntries);
  const sortedDates = Array.from(grouped.keys()).sort();

  // Build flat list of entries in display order for keyboard navigation
  const flatEntries = useMemo(() => {
    const taskEntriesFiltered = overdueEntries.filter(e => e.type === 'task' || e.type === 'done' || e.type === 'cancelled');
    const groupedEntries = groupByDate(taskEntriesFiltered);
    const dates = Array.from(groupedEntries.keys()).sort();
    const entries: Entry[] = [];
    for (const dateStr of dates) {
      entries.push(...groupedEntries.get(dateStr)!);
    }
    return entries;
  }, [overdueEntries]);

  // Map entry ID to flat index for selection
  const entryToFlatIndex = new Map<number, number>();
  flatEntries.forEach((entry, index) => {
    entryToFlatIndex.set(entry.id, index);
  });

  const toggleExpanded = (id: number) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleMarkDone = useCallback(async (entry: Entry) => {
    try {
      await MarkEntryDone(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to mark entry done:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to mark entry done');
    }
  }, [onEntryChanged, onError]);

  const handleMarkUndone = useCallback(async (entry: Entry) => {
    try {
      await MarkEntryUndone(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to mark entry undone:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to mark entry undone');
    }
  }, [onEntryChanged, onError]);

  const handleCancel = useCallback(async (entry: Entry) => {
    try {
      await CancelEntry(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to cancel entry:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to cancel entry');
    }
  }, [onEntryChanged, onError]);

  const handleUncancel = useCallback(async (entry: Entry) => {
    try {
      await UncancelEntry(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to uncancel entry:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to uncancel entry');
    }
  }, [onEntryChanged, onError]);

  const handleDelete = useCallback(async (entry: Entry) => {
    try {
      await DeleteEntry(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to delete entry:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to delete entry');
    }
  }, [onEntryChanged, onError]);

  const handleCyclePriority = useCallback(async (entry: Entry) => {
    try {
      await CyclePriority(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to cycle priority:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to cycle priority');
    }
  }, [onEntryChanged, onError]);

  const handleCycleType = useCallback(async (entry: Entry) => {
    const cycleOrder = ['task', 'note', 'event', 'question'] as const;
    const currentIndex = cycleOrder.indexOf(entry.type as typeof cycleOrder[number]);
    if (currentIndex === -1) return;
    const nextType = cycleOrder[(currentIndex + 1) % cycleOrder.length];
    try {
      await RetypeEntry(entry.id, nextType);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to cycle type:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to cycle type');
    }
  }, [onEntryChanged, onError]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isInputFocused = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA';
      if (isInputFocused) return;
      if (flatEntries.length === 0) return;

      switch (e.key) {
        case 'j':
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev => Math.min(prev + 1, flatEntries.length - 1));
          break;
        case 'k':
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev => Math.max(prev - 1, 0));
          break;
        case ' ':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex];
            if (selected.type === 'task') {
              await handleMarkDone(selected);
            } else if (selected.type === 'done') {
              await handleMarkUndone(selected);
            }
          }
          break;
        case 'x':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex];
            if (selected.type === 'cancelled') {
              await handleUncancel(selected);
            } else {
              await handleCancel(selected);
            }
          }
          break;
        case 'p':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex];
            await handleCyclePriority(selected);
          }
          break;
        case 't':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex];
            await handleCycleType(selected);
          }
          break;
        case 'Enter':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < flatEntries.length) {
            const selected = flatEntries[selectedIndex];
            toggleExpanded(selected.id);
          }
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [flatEntries, selectedIndex, handleCancel, handleCyclePriority, handleCycleType, handleMarkDone, handleMarkUndone, handleUncancel]);

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center gap-2">
        <button
          onClick={() => setCollapsed(!collapsed)}
          title={collapsed ? 'Expand' : 'Collapse'}
          className="p-1 rounded hover:bg-secondary transition-colors"
        >
          {collapsed ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <ChevronDown className="w-4 h-4" />
          )}
        </button>
        <Clock className="w-5 h-5 text-warning" data-testid="outstanding-icon" />
        <h2 className="font-display text-xl font-semibold flex-1">Pending Tasks</h2>
        <span className="px-2 py-0.5 text-sm font-medium bg-warning/20 text-warning rounded-full">
          {taskEntries.length}
        </span>
      </div>

      {/* Content */}
      {!collapsed && (
        <>
          {taskEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground italic py-6 text-center">
              No pending tasks. You're all caught up!
            </p>
          ) : (
            <div className="space-y-4">
              {sortedDates.map((dateStr) => (
                <div key={dateStr} className="space-y-2">
                  <h3 className="text-sm font-medium text-muted-foreground">
                    {formatDateHeader(dateStr)}
                  </h3>
                  <div className="space-y-1">
                    {grouped.get(dateStr)!.map((entry) => {
                      const isExpanded = expandedIds.has(entry.id);
                      const parentChain = buildParentChain(entry, entriesById);
                      const ancestorCount = parentChain.length;
                      const isSelected = entryToFlatIndex.get(entry.id) === selectedIndex;
                      return (
                        <div key={entry.id} className="space-y-1">
                          {/* Parent context entries (shown when expanded) */}
                          {isExpanded && parentChain.map((parent, index) => (
                            <div
                              key={parent.id}
                              className={cn(
                                'flex items-center gap-3 p-2 rounded-lg border border-border',
                                'bg-muted/30 text-muted-foreground'
                              )}
                              style={{ marginLeft: `${index * 16}px` }}
                            >
                              <span className="w-5 text-center font-mono">
                                {ENTRY_SYMBOLS[parent.type]}
                              </span>
                              <span className="flex-1 text-sm">{parent.content}</span>
                            </div>
                          ))}
                          {/* Task entry */}
                          <div
                            data-entry-id={entry.id}
                            onClick={() => toggleExpanded(entry.id)}
                            className={cn(
                              'flex items-center gap-3 p-2 rounded-lg border border-border cursor-pointer',
                              'bg-card transition-colors group',
                              !isSelected && 'hover:bg-secondary/30',
                              isSelected && 'ring-2 ring-primary'
                            )}
                            style={{ marginLeft: isExpanded ? `${parentChain.length * 16}px` : undefined }}
                          >
                            {/* Context pill - shows ancestor count when entry has parent and isn't expanded */}
                            {ancestorCount > 0 && !isExpanded && (
                              <ContextPill
                                count={ancestorCount}
                                onClick={() => toggleExpanded(entry.id)}
                              />
                            )}
                            {/* Symbol - clickable for task/done entries */}
                            {entry.type === 'task' || entry.type === 'done' ? (
                              <button
                                data-testid="entry-symbol"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  if (entry.type === 'task') {
                                    handleMarkDone(entry);
                                  } else {
                                    handleMarkUndone(entry);
                                  }
                                }}
                                title={entry.type === 'task' ? 'Mark done' : 'Mark undone'}
                                className="w-5 text-center text-muted-foreground font-mono cursor-pointer hover:opacity-70 transition-opacity"
                              >
                                {ENTRY_SYMBOLS[entry.type]}
                              </button>
                            ) : (
                              <span
                                data-testid="entry-symbol"
                                className="w-5 text-center text-muted-foreground font-mono"
                              >
                                {ENTRY_SYMBOLS[entry.type]}
                              </span>
                            )}
                            <span className={cn(
                              'flex-1 text-sm',
                              entry.type === 'done' && 'text-bujo-done',
                              entry.type === 'cancelled' && 'line-through text-muted-foreground'
                            )}>
                              {entry.content}
                            </span>
                            {entry.priority !== 'none' && (
                              <span className="text-xs text-warning font-medium">
                                {PRIORITY_SYMBOLS[entry.priority]}
                              </span>
                            )}
                            <EntryActionBar
                              entry={entry}
                              callbacks={{
                                onCancel: () => handleCancel(entry),
                                onUncancel: () => handleUncancel(entry),
                                onCyclePriority: () => handleCyclePriority(entry),
                                onCycleType: () => handleCycleType(entry),
                                onMigrate: onMigrate ? () => onMigrate(entry) : undefined,
                                onMoveToList: onMoveToList ? () => onMoveToList(entry) : undefined,
                                onNavigateToEntry: onNavigateToEntry ? () => onNavigateToEntry(entry) : undefined,
                                onEdit: onEdit ? () => onEdit(entry) : undefined,
                                onDelete: () => handleDelete(entry),
                              }}
                              variant="hover-reveal"
                              usePlaceholders
                            />
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
