import { DayEntries, Entry } from '@/types/bujo';
import { EntryItem } from './EntryItem';
import { Calendar, MapPin, Cloud, Heart } from 'lucide-react';
import { format, isToday, isTomorrow, isYesterday } from 'date-fns';
import { cn } from '@/lib/utils';
import { useState } from 'react';
import { MarkEntryDone, MarkEntryUndone } from '@/wailsjs/go/wails/App';

interface DayViewProps {
  day: DayEntries;
  selectedEntryId?: number | null;
  onEntryChanged?: () => void;
}

function buildTree(entries: Entry[]): Entry[] {
  const map = new Map<number, Entry>();
  const roots: Entry[] = [];
  
  entries.forEach(e => {
    map.set(e.id, { ...e, children: [] });
  });
  
  entries.forEach(e => {
    const entry = map.get(e.id)!;
    if (e.parentId && map.has(e.parentId)) {
      map.get(e.parentId)!.children!.push(entry);
    } else {
      roots.push(entry);
    }
  });
  
  return roots;
}

function formatDayLabel(dateStr: string): string {
  const date = new Date(dateStr + 'T00:00:00');
  if (isToday(date)) return 'Today';
  if (isTomorrow(date)) return 'Tomorrow';
  if (isYesterday(date)) return 'Yesterday';
  return format(date, 'EEEE, MMM d');
}

interface EntryTreeProps {
  entries: Entry[];
  depth?: number;
  collapsedIds: Set<number>;
  selectedEntryId?: number | null;
  onToggleCollapse: (id: number) => void;
  onToggleDone: (id: number) => void;
}

function EntryTree({ entries, depth = 0, collapsedIds, selectedEntryId, onToggleCollapse, onToggleDone }: EntryTreeProps) {
  return (
    <>
      {entries.map((entry) => {
        const hasChildren = entry.children && entry.children.length > 0;
        const isCollapsed = collapsedIds.has(entry.id);

        return (
          <div key={entry.id}>
            <EntryItem
              entry={entry}
              depth={depth}
              isCollapsed={isCollapsed}
              hasChildren={hasChildren}
              childCount={entry.children?.length || 0}
              isSelected={entry.id === selectedEntryId}
              onToggleCollapse={() => onToggleCollapse(entry.id)}
              onToggleDone={() => onToggleDone(entry.id)}
            />
            {hasChildren && !isCollapsed && (
              <EntryTree
                entries={entry.children!}
                depth={depth + 1}
                collapsedIds={collapsedIds}
                selectedEntryId={selectedEntryId}
                onToggleCollapse={onToggleCollapse}
                onToggleDone={onToggleDone}
              />
            )}
          </div>
        );
      })}
    </>
  );
}

export function DayView({ day, selectedEntryId, onEntryChanged }: DayViewProps) {
  const [collapsedIds, setCollapsedIds] = useState<Set<number>>(new Set());
  const tree = buildTree(day.entries);
  const dateObj = new Date(day.date + 'T00:00:00');

  const toggleCollapse = (id: number) => {
    setCollapsedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleToggleDone = async (id: number) => {
    const entry = day.entries.find(e => e.id === id);
    if (!entry || (entry.type !== 'task' && entry.type !== 'done')) return;

    try {
      if (entry.type === 'done') {
        await MarkEntryUndone(id);
      } else {
        await MarkEntryDone(id);
      }
      onEntryChanged?.();
    } catch (error) {
      console.error('Failed to toggle entry:', error);
    }
  };
  
  return (
    <div className="animate-fade-in">
      {/* Day Header */}
      <div className="flex items-center gap-4 mb-3 pb-2 border-b border-border">
        <div className="flex items-center gap-2">
          <Calendar className="w-4 h-4 text-primary" />
          <h3 className={cn(
            'font-display text-lg font-semibold',
            isToday(dateObj) && 'text-primary'
          )}>
            {formatDayLabel(day.date)}
          </h3>
          <span className="text-xs text-muted-foreground">
            {format(dateObj, 'MMM d, yyyy')}
          </span>
        </div>
        
        {/* Context indicators */}
        <div className="flex items-center gap-3 ml-auto text-xs text-muted-foreground">
          {day.location && (
            <span className="flex items-center gap-1">
              <MapPin className="w-3.5 h-3.5" />
              {day.location}
            </span>
          )}
          {day.weather && (
            <span className="flex items-center gap-1">
              <Cloud className="w-3.5 h-3.5" />
              {day.weather}
            </span>
          )}
          {day.mood && (
            <span className="flex items-center gap-1">
              <Heart className="w-3.5 h-3.5" />
              {day.mood}
            </span>
          )}
        </div>
      </div>
      
      {/* Entries */}
      <div className="space-y-0.5">
        {tree.length > 0 ? (
          <EntryTree
            entries={tree}
            collapsedIds={collapsedIds}
            selectedEntryId={selectedEntryId}
            onToggleCollapse={toggleCollapse}
            onToggleDone={handleToggleDone}
          />
        ) : (
          <p className="text-sm text-muted-foreground italic py-4 text-center">
            No entries yet. Start journaling!
          </p>
        )}
      </div>
    </div>
  );
}
