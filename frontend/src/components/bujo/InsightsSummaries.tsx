import { useEffect, useState } from 'react';
import { GetInsightsSummaries, GetInsightsSummaryDetail } from '../../../wailsjs/go/wails/App';
import { cn } from '@/lib/utils';

interface Summary {
  ID: number;
  WeekStart: string;
  WeekEnd: string;
  SummaryText: string;
  CreatedAt: string;
}

interface Topic {
  ID: number;
  SummaryID: number;
  Topic: string;
  Content: string;
  Importance: string;
}

export function InsightsSummaries() {
  const [summaries, setSummaries] = useState<Summary[]>([]);
  const [expandedID, setExpandedID] = useState<number | null>(null);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    GetInsightsSummaries(10)
      .then((data: Summary[]) => setSummaries(data))
      .catch((err: Error) => setError(err.message));
  }, []);

  const toggleSummary = async (id: number) => {
    if (expandedID === id) {
      setExpandedID(null);
      setTopics([]);
      return;
    }
    setExpandedID(id);
    try {
      const detail = await GetInsightsSummaryDetail(id);
      setTopics(detail);
    } catch {
      setTopics([]);
    }
  };

  if (error) {
    return <div className="text-destructive text-sm">Failed to load summaries: {error}</div>;
  }

  if (summaries.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No weekly summaries yet.</p>
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
    <div className="space-y-3">
      {summaries.map((s) => (
        <div key={s.ID} className="border border-border rounded-lg">
          <button
            onClick={() => toggleSummary(s.ID)}
            className="w-full text-left p-4 hover:bg-muted/50 transition-colors"
          >
            <div className="flex justify-between items-center">
              <h3 className="text-sm font-medium">
                {s.WeekStart} — {s.WeekEnd}
              </h3>
              <span className="text-xs text-muted-foreground">{s.CreatedAt.split(' ')[0]}</span>
            </div>
            {expandedID !== s.ID && (
              <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                {s.SummaryText}
              </p>
            )}
          </button>
          {expandedID === s.ID && (
            <div className="px-4 pb-4 space-y-3">
              <p className="text-sm whitespace-pre-wrap">{s.SummaryText}</p>
              {topics.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-muted-foreground mb-1">Topics</h4>
                  <div className="flex flex-wrap gap-1">
                    {topics.map((t) => (
                      <span key={t.ID} className="inline-flex items-center gap-1">
                        {importanceBadge(t.Importance)}
                        <span className="text-xs">{t.Topic}</span>
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
