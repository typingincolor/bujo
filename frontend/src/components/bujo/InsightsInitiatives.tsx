import { useEffect, useState } from 'react';
import { GetInsightsInitiativePortfolio } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';
import { cn } from '@/lib/utils';
import { InsightsInitiativeDetailView } from './InsightsInitiativeDetail';

export function InsightsInitiatives() {
  const [initiatives, setInitiatives] = useState<domain.InsightsInitiativePortfolio[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    GetInsightsInitiativePortfolio()
      .then((data) => { if (!cancelled) { setError(null); setInitiatives(data); } })
      .catch((err: Error) => { if (!cancelled) setError(err.message); });
    return () => { cancelled = true; };
  }, []);

  if (selectedId !== null) {
    return <InsightsInitiativeDetailView initiativeId={selectedId} onBack={() => setSelectedId(null)} />;
  }

  if (error) {
    return <div className="text-destructive text-sm">Failed to load initiatives: {error}</div>;
  }

  if (initiatives.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No initiatives found.</p>
      </div>
    );
  }

  const grouped = initiatives.reduce<Record<string, domain.InsightsInitiativePortfolio[]>>((acc, init) => {
    const status = init.Status || 'unknown';
    if (!acc[status]) acc[status] = [];
    acc[status].push(init);
    return acc;
  }, {});

  const statusOrder = ['active', 'planned', 'completed', 'on_hold'];
  const sortedStatuses = Object.keys(grouped).sort(
    (a, b) => (statusOrder.indexOf(a) === -1 ? 99 : statusOrder.indexOf(a)) -
              (statusOrder.indexOf(b) === -1 ? 99 : statusOrder.indexOf(b))
  );

  const statusColors: Record<string, string> = {
    active: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    planned: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    completed: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
    on_hold: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  };

  return (
    <div className="space-y-6">
      {sortedStatuses.map((status) => (
        <div key={status}>
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-3">
            {status.replace('_', ' ')}
          </h3>
          <div className="space-y-2">
            {grouped[status].map((init) => (
              <button
                key={init.ID}
                onClick={() => setSelectedId(init.ID)}
                className="border border-border rounded-lg p-3 w-full text-left hover:bg-secondary/30 transition-colors cursor-pointer"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium text-sm">{init.Name}</span>
                      <span className={cn('px-1.5 py-0.5 rounded text-xs', statusColors[status] || statusColors.completed)}>
                        {status.replace('_', ' ')}
                      </span>
                    </div>
                    {init.Description && (
                      <p className="text-sm text-muted-foreground">{init.Description}</p>
                    )}
                  </div>
                  <div className="text-right text-xs text-muted-foreground whitespace-nowrap">
                    <div>{init.MentionCount} mention{init.MentionCount !== 1 ? 's' : ''}</div>
                    {init.LastMentionWeek && <div>Last: {init.LastMentionWeek}</div>}
                  </div>
                </div>
              </button>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
