import { useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { GetInsightsWeeklyReport } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';
import { ArrowLeft } from 'lucide-react';
import { LevelBadge } from './insights-styles';

interface InsightsWeeklyReportProps {
  weekStart: string;
  onBack: () => void;
}

export function InsightsWeeklyReport({ weekStart, onBack }: InsightsWeeklyReportProps) {
  const [report, setReport] = useState<domain.InsightsWeeklyReport | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    GetInsightsWeeklyReport(weekStart)
      .then((data) => { if (!cancelled) { setError(null); setReport(data); } })
      .catch((err: Error) => { if (!cancelled) setError(err.message); });
    return () => { cancelled = true; };
  }, [weekStart]);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load report: {error}</div>;
  }

  if (!report) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Loading report...</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <button
        onClick={onBack}
        className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Summaries
      </button>

      <h2 className="text-lg font-semibold">Weekly Report</h2>

      {report.Summary && (
        <div className="border border-border rounded-lg p-4">
          <h3 className="text-sm font-medium mb-2">
            {report.Summary.WeekStart} — {report.Summary.WeekEnd}
          </h3>
          <div className="text-sm prose prose-sm dark:prose-invert max-w-none">
            <ReactMarkdown>{report.Summary.SummaryText}</ReactMarkdown>
          </div>
        </div>
      )}

      {report.Topics && report.Topics.length > 0 && (
        <div className="border border-border rounded-lg p-4">
          <h4 className="text-xs font-medium text-muted-foreground mb-2">Topics</h4>
          <div className="space-y-2">
            {report.Topics.map((t) => (
              <div key={t.ID} className="flex items-start gap-2">
                <LevelBadge level={t.Importance} />
                <div>
                  <span className="text-sm font-medium">{t.Topic}</span>
                  {t.Content && (
                    <p className="text-sm text-muted-foreground mt-0.5">{t.Content}</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {report.InitiativeUpdates && report.InitiativeUpdates.length > 0 && (
        <div className="border border-border rounded-lg p-4">
          <h4 className="text-xs font-medium text-muted-foreground mb-2">Initiative Updates</h4>
          <div className="space-y-2">
            {report.InitiativeUpdates.map((u, i) => (
              <div key={i} className="border-l-2 border-primary/30 pl-3">
                <span className="text-sm font-medium">{u.InitiativeName}</span>
                <p className="text-sm text-muted-foreground mt-0.5">{u.UpdateText}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {report.Actions && report.Actions.length > 0 && (
        <div className="border border-border rounded-lg p-4">
          <h4 className="text-xs font-medium text-muted-foreground mb-2">Actions</h4>
          <div className="space-y-2">
            {report.Actions.map((a) => (
              <div key={a.ID} className="flex items-start gap-2">
                <LevelBadge level={a.Priority} />
                <div className="flex-1">
                  <p className="text-sm">{a.ActionText}</p>
                  {a.DueDate && (
                    <span className="text-xs text-muted-foreground">Due: {a.DueDate}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
