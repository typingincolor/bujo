import { BujoList } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { List, CheckCircle2, Circle, ChevronRight } from 'lucide-react';
import { useState } from 'react';

interface ListsViewProps {
  lists: BujoList[];
}

interface ListCardProps {
  list: BujoList;
  isExpanded: boolean;
  onToggle: () => void;
}

function ListCard({ list, isExpanded, onToggle }: ListCardProps) {
  const progress = list.totalCount > 0 
    ? Math.round((list.doneCount / list.totalCount) * 100) 
    : 0;
  
  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden animate-fade-in">
      {/* Header */}
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-3 p-4 hover:bg-secondary/30 transition-colors"
      >
        <ChevronRight
          className={cn(
            'w-4 h-4 text-muted-foreground transition-transform',
            isExpanded && 'rotate-90'
          )}
        />
        <span className="font-medium flex-1 text-left">{list.name}</span>
        <span className="text-sm text-muted-foreground">
          {list.doneCount}/{list.totalCount}
        </span>
        {/* Progress bar */}
        <div className="w-16 h-1.5 bg-muted rounded-full overflow-hidden">
          <div
            className="h-full bg-bujo-done transition-all"
            style={{ width: `${progress}%` }}
          />
        </div>
      </button>
      
      {/* Items */}
      {isExpanded && (
        <div className="border-t border-border px-4 py-2 space-y-1">
          {list.items.map((item) => (
            <div
              key={item.id}
              className="flex items-center gap-3 py-1.5 group hover:bg-secondary/20 rounded px-2 -mx-2 cursor-pointer"
            >
              {item.done ? (
                <CheckCircle2 className="w-4 h-4 text-bujo-done flex-shrink-0" />
              ) : (
                <Circle className="w-4 h-4 text-muted-foreground flex-shrink-0" />
              )}
              <span className={cn(
                'text-sm flex-1',
                item.done && 'line-through text-muted-foreground'
              )}>
                {item.content}
              </span>
              <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100">
                ({item.id})
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function ListsView({ lists }: ListsViewProps) {
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set([lists[0]?.id]));
  
  const toggleExpanded = (id: number) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };
  
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 mb-4">
        <List className="w-5 h-5 text-primary" />
        <h2 className="font-display text-xl font-semibold">Lists</h2>
      </div>
      
      <div className="space-y-2">
        {lists.map((list) => (
          <ListCard
            key={list.id}
            list={list}
            isExpanded={expandedIds.has(list.id)}
            onToggle={() => toggleExpanded(list.id)}
          />
        ))}
      </div>
    </div>
  );
}
