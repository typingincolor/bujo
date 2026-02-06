import { useEffect, useState } from 'react';
import { GetInsightsDecisionLog } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';

export function InsightsDecisions() {
  const [decisions, setDecisions] = useState<domain.InsightsDecisionWithInitiatives[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    GetInsightsDecisionLog()
      .then((data) => { if (!cancelled) { setError(null); setDecisions(data); } })
      .catch((err: Error) => { if (!cancelled) setError(err.message); });
    return () => { cancelled = true; };
  }, []);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load decisions: {error}</div>;
  }

  if (decisions.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No decisions found.</p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-3">
        Decision Log ({decisions.length})
      </h3>
      {decisions.map((d) => (
        <div key={d.ID} className="border border-border rounded-lg p-3">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-xs text-muted-foreground">{d.DecisionDate}</span>
          </div>
          <p className="text-sm font-medium">{d.DecisionText}</p>
          {d.Rationale && (
            <p className="text-sm text-muted-foreground mt-1">Rationale: {d.Rationale}</p>
          )}
          {d.Participants && (
            <p className="text-xs text-muted-foreground mt-1">Participants: {d.Participants}</p>
          )}
          {d.Initiatives && (
            <div className="mt-1 flex flex-wrap gap-1">
              {d.Initiatives.split(',').map((init) => (
                <span
                  key={init.trim()}
                  className="px-1.5 py-0.5 rounded text-xs bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                >
                  {init.trim()}
                </span>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
