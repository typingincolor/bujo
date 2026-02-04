import { useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { GetInsightsSummaryForWeek } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';
import { cn } from '@/lib/utils';

interface InsightsSummariesProps {
  weekStart: string;
}

export function InsightsSummaries({ weekStart }: InsightsSummariesProps) {
  const [summary, setSummary] = useState<domain.InsightsSummary | null>(null);
  const [topics, setTopics] = useState<domain.InsightsTopic[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setError(null);
    GetInsightsSummaryForWeek(weekStart)
      .then((detail) => {
        setSummary(detail.Summary ?? null);
        setTopics(detail.Topics ?? []);
      })
      .catch((err: Error) => setError(err.message));
  }, [weekStart]);

  if (error) {
    return <div className="text-destructive text-sm">Failed to load summary: {error}</div>;
  }

  if (!summary) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No summary for this week.</p>
      </div>
    );
  }

  const importanceBadge = (importance: string) => {
    const colors: Record<string, string> = {
      high: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
      medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      low: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
    };
    return (
      <span className={cn('px-1.5 py-0.5 rounded text-xs', colors[importance] || colors.low)}>
        {importance}
      </span>
    );
  };

  return (
    <div className="space-y-4">
      <div className="border border-border rounded-lg p-4">
        <div className="flex justify-between items-center mb-3">
          <h3 className="text-sm font-medium">
            {summary.WeekStart} — {summary.WeekEnd}
          </h3>
          <span className="text-xs text-muted-foreground">{summary.CreatedAt.split(' ')[0]}</span>
        </div>
        <div className="text-sm prose prose-sm dark:prose-invert max-w-none">
          <ReactMarkdown>{summary.SummaryText}</ReactMarkdown>
        </div>
      </div>
      {topics.length > 0 && (
        <div className="border border-border rounded-lg p-4">
          <h4 className="text-xs font-medium text-muted-foreground mb-2">Topics</h4>
          <div className="space-y-2">
            {topics.map((t) => (
              <div key={t.ID} className="flex items-start gap-2">
                {importanceBadge(t.Importance)}
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
    </div>
  );
}
