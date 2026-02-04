import { useState } from 'react';
import { cn } from '@/lib/utils';
import { InsightsDashboard } from './InsightsDashboard';
import { InsightsSummaries } from './InsightsSummaries';
import { InsightsActions } from './InsightsActions';

type InsightsTab = 'dashboard' | 'summaries' | 'actions';

const tabs: { id: InsightsTab; label: string }[] = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'summaries', label: 'Summaries' },
  { id: 'actions', label: 'Actions' },
];

export function InsightsView() {
  const [activeTab, setActiveTab] = useState<InsightsTab>('dashboard');

  return (
    <div className="flex flex-col h-full">
      <div className="border-b border-border px-4">
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
      </div>
      <div className="flex-1 overflow-auto p-4">
        {activeTab === 'dashboard' && <InsightsDashboard />}
        {activeTab === 'summaries' && <InsightsSummaries />}
        {activeTab === 'actions' && <InsightsActions />}
      </div>
    </div>
  );
}
