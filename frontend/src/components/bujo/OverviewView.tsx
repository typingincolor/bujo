import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { AlertTriangle, Check, ChevronDown, ChevronRight, Undo2 } from 'lucide-react';
import { format, parseISO } from 'date-fns';
import { useState } from 'react';
import { MarkEntryDone, MarkEntryUndone } from '@/wailsjs/go/wails/App';

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

export function OverviewView({ overdueEntries, onEntryChanged, onError }: OverviewViewProps) {
  const [collapsed, setCollapsed] = useState(false);
  const grouped = groupByDate(overdueEntries);
  const sortedDates = Array.from(grouped.keys()).sort();

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
        <AlertTriangle className="w-5 h-5 text-warning" />
        <h2 className="font-display text-xl font-semibold flex-1">Overdue Tasks</h2>
        <span className="px-2 py-0.5 text-sm font-medium bg-warning/20 text-warning rounded-full">
          {overdueEntries.length}
        </span>
      </div>

      {/* Content */}
      {!collapsed && (
        <>
          {overdueEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground italic py-6 text-center">
              No overdue tasks. You're all caught up!
            </p>
          ) : (
            <div className="space-y-4">
              {sortedDates.map((dateStr) => (
                <div key={dateStr} className="space-y-2">
                  <h3 className="text-sm font-medium text-muted-foreground">
                    {formatDateHeader(dateStr)}
                  </h3>
                  <div className="space-y-1">
                    {grouped.get(dateStr)!.map((entry) => (
                      <div
                        key={entry.id}
                        className={cn(
                          'flex items-center gap-3 p-2 rounded-lg border border-border',
                          'bg-card hover:bg-secondary/30 transition-colors group'
                        )}
                      >
                        <span
                          data-testid="entry-symbol"
                          className="w-5 text-center text-muted-foreground font-mono"
                        >
                          {ENTRY_SYMBOLS[entry.type]}
                        </span>
                        <span className={cn(
                          'flex-1 text-sm',
                          entry.type === 'done' && 'line-through text-muted-foreground'
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
                            onClick={() => handleMarkUndone(entry)}
                            title="Mark undone"
                            className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                          >
                            <Undo2 className="w-4 h-4" />
                          </button>
                        ) : (
                          <button
                            onClick={() => handleMarkDone(entry)}
                            title="Mark done"
                            className="p-1 rounded hover:bg-bujo-done/20 text-muted-foreground hover:text-bujo-done transition-colors opacity-0 group-hover:opacity-100"
                          >
                            <Check className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    ))}
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
