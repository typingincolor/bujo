import { useState } from 'react';
import { startOfWeek, addWeeks, isSameWeek, format } from 'date-fns';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { InsightsDashboard } from './InsightsDashboard';
import { InsightsSummaries } from './InsightsSummaries';
import { InsightsActions } from './InsightsActions';
import { InsightsInitiatives } from './InsightsInitiatives';
import { InsightsTopics } from './InsightsTopics';
import { InsightsDecisions } from './InsightsDecisions';

type InsightsTab = 'dashboard' | 'summaries' | 'actions' | 'initiatives' | 'topics' | 'decisions';

const tabs: { id: InsightsTab; label: string }[] = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'summaries', label: 'Summaries' },
  { id: 'actions', label: 'Actions' },
  { id: 'initiatives', label: 'Initiatives' },
  { id: 'topics', label: 'Topics' },
  { id: 'decisions', label: 'Decisions' },
];

function getWeekStart(date: Date): string {
  const monday = startOfWeek(date, { weekStartsOn: 1 });
  return format(monday, 'yyyy-MM-dd');
}

function getWeekEnd(date: Date): string {
  const monday = startOfWeek(date, { weekStartsOn: 1 });
  const sunday = addWeeks(monday, 1);
  sunday.setDate(sunday.getDate() - 1);
  return format(sunday, 'yyyy-MM-dd');
}

export function InsightsView() {
  const [activeTab, setActiveTab] = useState<InsightsTab>('dashboard');
  const [weekAnchor, setWeekAnchor] = useState<Date>(() =>
    startOfWeek(new Date(), { weekStartsOn: 1 })
  );

  const weekStart = getWeekStart(weekAnchor);
  const weekEnd = getWeekEnd(weekAnchor);

  const handlePrevWeek = () => setWeekAnchor((prev) => addWeeks(prev, -1));
  const handleNextWeek = () => setWeekAnchor((prev) => addWeeks(prev, 1));
  const handleGoToCurrentWeek = () =>
    setWeekAnchor(startOfWeek(new Date(), { weekStartsOn: 1 }));

  const showWeekNav = activeTab !== 'dashboard' && activeTab !== 'initiatives' && activeTab !== 'topics' && activeTab !== 'decisions';

  return (
    <div className="flex flex-col h-full">
      <div className="border-b border-border px-4">
        <div className="flex items-center justify-between">
          <div className="flex gap-4">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={cn(
                  'py-2 px-1 text-sm border-b-2 transition-colors',
                  activeTab === tab.id
                    ? 'border-primary text-primary font-medium'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                )}
              >
                {tab.label}
              </button>
            ))}
          </div>
          {showWeekNav && (
            <div className="flex items-center gap-3">
              <button
                onClick={handlePrevWeek}
                title="Previous week"
                className="p-1.5 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <span className="text-sm text-muted-foreground">
                {weekStart} — {weekEnd}
              </span>
              <button
                onClick={handleNextWeek}
                title="Next week"
                className="p-1.5 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
              {!isSameWeek(weekAnchor, new Date(), { weekStartsOn: 1 }) && (
                <button
                  onClick={handleGoToCurrentWeek}
                  className="px-2 py-1 text-xs rounded hover:bg-secondary/50 transition-colors"
                >
                  Today
                </button>
              )}
            </div>
          )}
        </div>
      </div>
      <div className="flex-1 overflow-auto p-4">
        {activeTab === 'dashboard' && <InsightsDashboard />}
        {activeTab === 'summaries' && <InsightsSummaries weekStart={weekStart} />}
        {activeTab === 'actions' && <InsightsActions weekStart={weekStart} />}
        {activeTab === 'initiatives' && <InsightsInitiatives />}
        {activeTab === 'topics' && <InsightsTopics />}
        {activeTab === 'decisions' && <InsightsDecisions />}
      </div>
    </div>
  );
}
