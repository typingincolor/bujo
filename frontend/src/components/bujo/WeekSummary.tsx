import { DayEntries, Entry } from '@/types/bujo';
import { calculateAttentionScore, sortByAttentionScore } from '@/lib/attentionScore';
import { cn } from '@/lib/utils';
import { EntrySymbol } from './EntrySymbol';

interface WeekSummaryProps {
  days: DayEntries[];
  onShowAllAttention?: () => void;
}

const MAX_ATTENTION_ITEMS = 5;
const MIN_ATTENTION_SCORE = 10;

function flattenEntries(entries: Entry[]): Entry[] {
  const result: Entry[] = [];
  function traverse(items: Entry[]) {
    for (const entry of items) {
      result.push(entry);
      if (entry.children && entry.children.length > 0) {
        traverse(entry.children);
      }
    }
  }
  traverse(entries);
  return result;
}

export function WeekSummary({ days, onShowAllAttention }: WeekSummaryProps) {
  const allEntries = days.flatMap(day => flattenEntries(day.entries));
  const now = new Date();

  const createdCount = allEntries.filter(e => e.type === 'task').length;
  const doneCount = allEntries.filter(e => e.type === 'done').length;
  const migratedCount = allEntries.filter(e => e.type === 'migrated').length;
  const openCount = allEntries.filter(e => e.type === 'task').length;

  const eventsWithChildren = allEntries.filter(e => {
    if (e.type !== 'event') return false;
    const children = allEntries.filter(child => child.parentId === e.id);
    return children.length > 0;
  });

  const getChildCount = (eventId: number): number => {
    return allEntries.filter(e => e.parentId === eventId).length;
  };

  const needsAttentionEntries = allEntries.filter(
    e => e.type === 'task' || e.type === 'question'
  );
  const sortedAttentionEntries = sortByAttentionScore(needsAttentionEntries, now);
  const filteredAttentionEntries = sortedAttentionEntries.filter(
    e => calculateAttentionScore(e, now).score >= MIN_ATTENTION_SCORE
  );
  const hasMoreThanLimit = filteredAttentionEntries.length > MAX_ATTENTION_ITEMS;
  const topAttentionEntries = filteredAttentionEntries.slice(0, MAX_ATTENTION_ITEMS);

  return (
    <div data-testid="week-summary" className="space-y-6">
      <section className="space-y-3">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
          Task Flow
        </h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          <FlowStat label="Created" value={createdCount} testId="task-flow-created" />
          <FlowStat label="Done" value={doneCount} testId="task-flow-done" />
          <FlowStat label="Migrated" value={migratedCount} testId="task-flow-migrated" />
          <FlowStat label="Open" value={openCount} testId="task-flow-open" />
        </div>
      </section>

      <section data-testid="week-summary-meetings" className="space-y-3">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
          Meetings
        </h3>
        <div className="space-y-2">
          {eventsWithChildren.length === 0 ? (
            <p className="text-sm text-muted-foreground">No meetings this week</p>
          ) : (
            eventsWithChildren.map(event => (
              <button
                key={event.id}
                type="button"
                className="w-full flex items-center justify-between p-2 rounded-lg border border-border bg-card hover:bg-muted/50 cursor-pointer text-left"
              >
                <span className="flex items-center gap-2">
                  <EntrySymbol type={event.type} priority={event.priority} />
                  <span className="text-sm">{event.content}</span>
                </span>
                <span className="text-xs text-muted-foreground">
                  {getChildCount(event.id)} items
                </span>
              </button>
            ))
          )}
        </div>
      </section>

      <section data-testid="week-summary-attention" className="space-y-3">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
          Needs Attention
        </h3>
        <div className="space-y-2">
          {topAttentionEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground">All caught up!</p>
          ) : (
            topAttentionEntries.map(entry => {
              const { indicators } = calculateAttentionScore(entry, now);
              const isHighPriority = entry.priority === 'high';
              return (
                <button
                  key={entry.id}
                  type="button"
                  data-testid="attention-item"
                  data-attention-item
                  data-priority={isHighPriority ? 'high' : undefined}
                  className="w-full flex items-center justify-between p-2 rounded-lg border border-border bg-card hover:bg-muted/50 cursor-pointer text-left"
                >
                  <span className="flex items-center gap-2">
                    <EntrySymbol type={entry.type} priority={entry.priority} />
                    <span className="text-sm">{entry.content}</span>
                  </span>
                  {indicators.length > 0 && (
                    <div className="flex gap-1" data-testid="attention-indicators">
                      {indicators.map(indicator => (
                        <span
                          key={indicator}
                          data-indicator={indicator}
                          title={indicator}
                          className={cn(
                            'text-xs px-1.5 py-0.5 rounded',
                            indicator === 'priority' && 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
                            indicator === 'overdue' && 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
                            indicator === 'aging' && 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
                            indicator === 'migrated' && 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                          )}
                        >
                          {indicator === 'priority' ? '!' : indicator}
                        </span>
                      ))}
                    </div>
                  )}
                </button>
              );
            })
          )}
          {hasMoreThanLimit && (
            <button
              onClick={onShowAllAttention}
              className="text-sm text-primary hover:underline"
            >
              Show all
            </button>
          )}
        </div>
      </section>
    </div>
  );
}

interface FlowStatProps {
  label: string;
  value: number;
  testId?: string;
}

function FlowStat({ label, value, testId }: FlowStatProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-3 text-center">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="sr-only">: </span>
      <span data-testid={testId} className="font-display text-xl font-semibold block">{value}</span>
    </div>
  );
}
