import { useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { GetInsightsDashboard } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';

export function InsightsDashboard() {
  const [dashboard, setDashboard] = useState<domain.InsightsDashboard | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    GetInsightsDashboard()
      .then((data) => setDashboard(data))
      .catch((err: Error) => setError(err.message));
  }, []);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load insights: {error}</div>;
  }

  if (!dashboard) {
    return <div className="text-muted-foreground text-sm">Loading...</div>;
  }

  if (dashboard.Status === 'not_initialized') {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Insights not available.</p>
        <p className="text-sm text-muted-foreground mt-2">
          Generate weekly summaries with Claude to see insights here.
        </p>
      </div>
    );
  }

  if (dashboard.Status === 'empty') {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No insights yet.</p>
        <p className="text-sm text-muted-foreground mt-2">
          Generate your first weekly summary to get started.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Staleness indicator */}
      {dashboard.DaysSinceLastSummary > 0 && (
        <div className="text-xs text-muted-foreground">
          Last summary: {dashboard.LatestSummary?.WeekStart} — {dashboard.LatestSummary?.WeekEnd}
          {' '}({dashboard.DaysSinceLastSummary} days ago)
        </div>
      )}

      {/* Latest Summary */}
      {dashboard.LatestSummary && (
        <div className="border border-border rounded-lg p-4 max-h-80 overflow-y-auto">
          <h3 className="text-sm font-medium mb-2">
            Latest Summary ({dashboard.LatestSummary.WeekStart} — {dashboard.LatestSummary.WeekEnd})
          </h3>
          <div className="text-sm text-muted-foreground prose prose-sm dark:prose-invert max-w-none">
            <ReactMarkdown>
              {dashboard.LatestSummary.SummaryText}
            </ReactMarkdown>
          </div>
        </div>
      )}

      {/* Active Initiatives */}
      {dashboard.ActiveInitiatives?.length > 0 && (
        <div className="border border-border rounded-lg p-4 max-h-64 overflow-y-auto">
          <h3 className="text-sm font-medium mb-2">Active Initiatives</h3>
          <ul className="space-y-1">
            {dashboard.ActiveInitiatives.map((i) => (
              <li key={i.ID} className="text-sm flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-green-500" />
                {i.Name}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* High Priority Actions */}
      {dashboard.HighPriorityActions?.length > 0 && (
        <div className="border border-border rounded-lg p-4 max-h-64 overflow-y-auto">
          <h3 className="text-sm font-medium mb-2">High Priority Actions</h3>
          <ul className="list-disc list-inside space-y-1">
            {dashboard.HighPriorityActions.map((a) => (
              <li key={a.ID} className="text-sm">
                <span>{a.ActionText}</span>
                {a.DueDate && (
                  <span className="text-xs text-muted-foreground ml-2">{a.DueDate}</span>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Recent Decisions */}
      {dashboard.RecentDecisions?.length > 0 && (
        <div className="border border-border rounded-lg p-4 max-h-64 overflow-y-auto">
          <h3 className="text-sm font-medium mb-2">Recent Decisions</h3>
          <ul className="space-y-2">
            {dashboard.RecentDecisions.map((d) => (
              <li key={d.ID} className="text-sm">
                <div className="font-medium">{d.DecisionText}</div>
                <div className="text-xs text-muted-foreground">{d.DecisionDate}</div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
