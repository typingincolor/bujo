import { useEffect, useState } from 'react';
import { GetInsightsActionsForWeek } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';
import { cn } from '@/lib/utils';

interface InsightsActionsProps {
  weekStart: string;
}

export function InsightsActions({ weekStart }: InsightsActionsProps) {
  const [actions, setActions] = useState<domain.InsightsAction[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setError(null);
    GetInsightsActionsForWeek(weekStart)
      .then((data) => setActions(data))
      .catch((err: Error) => setError(err.message));
  }, [weekStart]);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load actions: {error}</div>;
  }

  if (actions.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No actions for this week.</p>
      </div>
    );
  }

  const today = new Date().toISOString().split('T')[0];

  const priorityBadge = (priority: string) => {
    const colors: Record<string, string> = {
      high: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
      medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      low: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
    };
    return (
      <span className={cn('px-1.5 py-0.5 rounded text-xs', colors[priority] || colors.low)}>
        {priority}
      </span>
    );
  };

  return (
    <div className="space-y-2">
      {actions.map((a) => {
        const isOverdue = a.DueDate && a.DueDate < today;
        return (
          <div
            key={a.ID}
            className={cn(
              'border border-border rounded-lg p-3 flex items-start justify-between gap-3',
              isOverdue && 'border-red-300 dark:border-red-800'
            )}
          >
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                {priorityBadge(a.Priority)}
                <span className="text-xs text-muted-foreground">
                  from {a.WeekStart}
                </span>
              </div>
              <p className="text-sm">{a.ActionText}</p>
            </div>
            {a.DueDate && (
              <div className={cn(
                'text-xs whitespace-nowrap',
                isOverdue ? 'text-red-600 dark:text-red-400 font-medium' : 'text-muted-foreground'
              )}>
                {isOverdue ? 'Overdue: ' : 'Due: '}{a.DueDate}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
