import { useEffect, useState } from 'react';
import { GetInsightsInitiativeDetail } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';
import { cn } from '@/lib/utils';
import { ArrowLeft } from 'lucide-react';

interface InsightsInitiativeDetailProps {
  initiativeId: number;
  onBack: () => void;
}

export function InsightsInitiativeDetailView({ initiativeId, onBack }: InsightsInitiativeDetailProps) {
  const [detail, setDetail] = useState<domain.InsightsInitiativeDetail | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    GetInsightsInitiativeDetail(initiativeId)
      .then((data) => { if (!cancelled) { setError(null); setDetail(data); } })
      .catch((err: Error) => { if (!cancelled) setError(err.message); });
    return () => { cancelled = true; };
  }, [initiativeId]);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load initiative detail: {error}</div>;
  }

  if (!detail) {
    return <div className="text-muted-foreground text-sm">Loading...</div>;
  }

  const statusColors: Record<string, string> = {
    active: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    planned: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    completed: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
    on_hold: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  };

  const priorityColors: Record<string, string> = {
    high: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    low: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
  };

  const init = detail.Initiative;
  const status = init?.Status || 'unknown';

  return (
    <div className="space-y-6">
      <button
        onClick={onBack}
        className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Initiatives
      </button>

      <div>
        <div className="flex items-center gap-2 mb-2">
          <h2 className="text-lg font-semibold">{init?.Name}</h2>
          <span className={cn('px-1.5 py-0.5 rounded text-xs', statusColors[status] || statusColors.completed)}>
            {status.replace('_', ' ')}
          </span>
        </div>
        {init?.Description && (
          <p className="text-sm text-muted-foreground">{init.Description}</p>
        )}
      </div>

      {detail.Updates && detail.Updates.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-3">Timeline</h3>
          <div className="space-y-2">
            {detail.Updates.map((u, i) => (
              <div key={i} className="border border-border rounded-lg p-3">
                <div className="text-xs text-muted-foreground mb-1">
                  {u.WeekStart} — {u.WeekEnd}
                </div>
                <p className="text-sm">{u.UpdateText}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {detail.PendingActions && detail.PendingActions.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-3">Pending Actions</h3>
          <div className="space-y-2">
            {detail.PendingActions.map((a) => (
              <div key={a.ID} className="border border-border rounded-lg p-3 flex items-start justify-between gap-3">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={cn('px-1.5 py-0.5 rounded text-xs', priorityColors[a.Priority] || priorityColors.low)}>
                      {a.Priority}
                    </span>
                  </div>
                  <p className="text-sm">{a.ActionText}</p>
                </div>
                {a.DueDate && (
                  <div className="text-xs text-muted-foreground whitespace-nowrap">
                    Due: {a.DueDate}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {detail.Decisions && detail.Decisions.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-3">Decisions</h3>
          <div className="space-y-2">
            {detail.Decisions.map((d) => (
              <div key={d.ID} className="border border-border rounded-lg p-3">
                <p className="text-sm font-medium mb-1">{d.DecisionText}</p>
                {d.Rationale && (
                  <p className="text-sm text-muted-foreground">{d.Rationale}</p>
                )}
                <div className="flex gap-3 mt-2 text-xs text-muted-foreground">
                  {d.DecisionDate && <span>{d.DecisionDate}</span>}
                  {d.Participants && <span>{d.Participants}</span>}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
