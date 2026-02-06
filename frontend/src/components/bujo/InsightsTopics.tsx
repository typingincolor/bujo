import { useEffect, useState } from 'react';
import { GetInsightsDistinctTopics, GetInsightsTopicTimeline } from '@/wailsjs/go/wails/App';
import { domain } from '@/wailsjs/go/models';
import { ArrowLeft } from 'lucide-react';

export function InsightsTopics() {
  const [topics, setTopics] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [selectedTopic, setSelectedTopic] = useState<string | null>(null);
  const [timeline, setTimeline] = useState<domain.InsightsTopicTimeline[]>([]);
  const [timelineError, setTimelineError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    GetInsightsDistinctTopics()
      .then((data) => { if (!cancelled) { setError(null); setTopics(data); } })
      .catch((err: Error) => { if (!cancelled) setError(err.message); });
    return () => { cancelled = true; };
  }, []);

  const handleSelectTopic = (topic: string) => {
    setSelectedTopic(topic);
    setTimelineError(null);
    let cancelled = false;
    GetInsightsTopicTimeline(topic)
      .then((data) => { if (!cancelled) setTimeline(data); })
      .catch((err: Error) => { if (!cancelled) setTimelineError(err.message); });
    return () => { cancelled = true; };
  };

  if (selectedTopic) {
    return (
      <div className="space-y-4">
        <button
          onClick={() => { setSelectedTopic(null); setTimeline([]); }}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Topics
        </button>

        <h2 className="text-lg font-semibold">{selectedTopic}</h2>

        {timelineError && (
          <div className="text-destructive text-sm">Failed to load timeline: {timelineError}</div>
        )}

        {timeline.length === 0 && !timelineError && (
          <p className="text-muted-foreground text-sm">No timeline entries found.</p>
        )}

        <div className="space-y-2">
          {timeline.map((entry, i) => (
            <div key={i} className="border border-border rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-xs text-muted-foreground">
                  {entry.WeekStart} — {entry.WeekEnd}
                </span>
                {entry.Importance === 'high' && (
                  <span className="px-1.5 py-0.5 rounded text-xs bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
                    high
                  </span>
                )}
              </div>
              {entry.Content && <p className="text-sm">{entry.Content}</p>}
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return <div className="text-destructive text-sm">Failed to load topics: {error}</div>;
  }

  if (topics.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No topics found.</p>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-3">All Topics</h3>
      {topics.map((topic) => (
        <button
          key={topic}
          onClick={() => handleSelectTopic(topic)}
          className="block w-full text-left px-3 py-2 rounded-lg text-sm hover:bg-secondary/30 transition-colors"
        >
          {topic}
        </button>
      ))}
    </div>
  );
}
