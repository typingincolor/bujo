import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Clock, Check, ChevronDown, ChevronRight, X, RotateCcw, Trash2, Pencil, ArrowRight, Flag, RefreshCw } from 'lucide-react';
import { ContextPill } from './ContextPill';
import { format, parseISO } from 'date-fns';
import { useState, useEffect } from 'react';
import { MarkEntryDone, MarkEntryUndone, CancelEntry, UncancelEntry, DeleteEntry, CyclePriority, RetypeEntry } from '@/wailsjs/go/wails/App';

function ActionPlaceholder() {
  return <span data-action-slot className="p-1 w-6 h-6" aria-hidden="true" />;
}

interface OverviewViewProps {
  overdueEntries: Entry[];
  onEntryChanged?: () => void;
  onError?: (message: string) => void;
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

export function OverviewView({ overdueEntries, onEntryChanged, onError }: OverviewViewProps) {
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
  const flatEntries: Entry[] = [];
  for (const dateStr of sortedDates) {
    flatEntries.push(...grouped.get(dateStr)!);
  }

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

  const handleMarkDone = async (entry: Entry) => {
    try {
      await MarkEntryDone(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to mark entry done:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to mark entry done');
    }
  };

  const handleMarkUndone = async (entry: Entry) => {
    try {
      await MarkEntryUndone(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to mark entry undone:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to mark entry undone');
    }
  };

  const handleCancel = async (entry: Entry) => {
    try {
      await CancelEntry(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to cancel entry:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to cancel entry');
    }
  };

  const handleUncancel = async (entry: Entry) => {
    try {
      await UncancelEntry(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to uncancel entry:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to uncancel entry');
    }
  };

  const handleDelete = async (entry: Entry) => {
    try {
      await DeleteEntry(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to delete entry:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to delete entry');
    }
  };

  const handleCyclePriority = async (entry: Entry) => {
    try {
      await CyclePriority(entry.id);
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to cycle priority:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to cycle priority');
    }
  };

  const handleCycleType = async (entry: Entry) => {
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
  };

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
  }, [flatEntries, selectedIndex]);

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
              No outstanding tasks. You're all caught up!
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
                            <span
                              data-testid="entry-symbol"
                              className="w-5 text-center text-muted-foreground font-mono"
                            >
                              {ENTRY_SYMBOLS[entry.type]}
                            </span>
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
                            {entry.type === 'done' ? (
                              <button
                                data-action-slot
                                onClick={(e) => { e.stopPropagation(); handleMarkUndone(entry); }}
                                title="Mark undone"
                                className="p-1 rounded hover:bg-orange-500/20 text-muted-foreground hover:text-orange-600 transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <span className="text-sm font-bold leading-none">•</span>
                              </button>
                            ) : entry.type !== 'cancelled' ? (
                              <button
                                data-action-slot
                                onClick={(e) => { e.stopPropagation(); handleMarkDone(entry); }}
                                title="Mark done"
                                className="p-1 rounded hover:bg-bujo-done/20 text-muted-foreground hover:text-bujo-done transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <Check className="w-4 h-4" />
                              </button>
                            ) : (
                              <ActionPlaceholder />
                            )}
                            {entry.type !== 'cancelled' ? (
                              <button
                                data-action-slot
                                onClick={(e) => { e.stopPropagation(); handleCancel(entry); }}
                                title="Cancel entry"
                                className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <X className="w-4 h-4" />
                              </button>
                            ) : (
                              <button
                                data-action-slot
                                onClick={(e) => { e.stopPropagation(); handleUncancel(entry); }}
                                title="Uncancel entry"
                                className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <RotateCcw className="w-4 h-4" />
                              </button>
                            )}
                            <button
                              data-action-slot
                              onClick={(e) => { e.stopPropagation(); handleCyclePriority(entry); }}
                              title="Cycle priority"
                              className="p-1 rounded hover:bg-warning/20 text-muted-foreground hover:text-warning transition-colors opacity-0 group-hover:opacity-100"
                            >
                              <Flag className="w-4 h-4" />
                            </button>
                            {entry.type === 'task' ? (
                              <button
                                data-action-slot
                                onClick={(e) => { e.stopPropagation(); handleCycleType(entry); }}
                                title="Change type"
                                className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <RefreshCw className="w-4 h-4" />
                              </button>
                            ) : (
                              <ActionPlaceholder />
                            )}
                            {entry.type === 'task' ? (
                              <button
                                data-action-slot
                                onClick={(e) => e.stopPropagation()}
                                title="Migrate entry"
                                className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                              >
                                <ArrowRight className="w-4 h-4" />
                              </button>
                            ) : (
                              <ActionPlaceholder />
                            )}
                            <button
                              data-action-slot
                              onClick={(e) => e.stopPropagation()}
                              title="Edit entry"
                              className="p-1 rounded hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors opacity-0 group-hover:opacity-100"
                            >
                              <Pencil className="w-4 h-4" />
                            </button>
                            <button
                              data-action-slot
                              onClick={(e) => { e.stopPropagation(); handleDelete(entry); }}
                              title="Delete entry"
                              className="p-1 rounded hover:bg-destructive/20 text-muted-foreground hover:text-destructive transition-colors opacity-0 group-hover:opacity-100"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
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
