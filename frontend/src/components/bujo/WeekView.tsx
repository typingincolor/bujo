import { useState, useMemo, useEffect } from 'react';
import { DayEntries, Entry, Habit } from '@/types/bujo';
import { DayBox } from './DayBox';
import { WeekendBox } from './WeekendBox';
import { filterWeekEntries, flattenEntries } from '@/lib/weekView';
import { format, parseISO } from 'date-fns';
import { ActionCallbacks } from './EntryActions/types';
import { buildTree } from '@/lib/buildTree';
import { ContextTree } from './ContextTree';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface WeekViewCallbacks {
  onMarkDone?: (entry: Entry) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
  onCyclePriority?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
}

interface WeekViewProps {
  days: DayEntries[];
  habits?: Habit[];
  callbacks?: WeekViewCallbacks;
  contextTree?: Entry[];
  onSelectEntry?: (entry: Entry | undefined) => void;
  isContextCollapsed?: boolean;
  onToggleContextCollapse?: () => void;
}

export function WeekView({
  days,
  habits = [],
  callbacks = {},
  contextTree: contextTreeProp,
  onSelectEntry,
  isContextCollapsed = true, // Default collapsed per user request
  onToggleContextCollapse
}: WeekViewProps) {
  const [selectedEntry, setSelectedEntry] = useState<Entry | undefined>();

  const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'];

  useEffect(() => {
    if (onSelectEntry) {
      onSelectEntry(selectedEntry);
    }
  }, [selectedEntry, onSelectEntry]);

  const createEntryCallbacks = (entry: Entry): ActionCallbacks => ({
    onCancel: callbacks.onMarkDone ? () => callbacks.onMarkDone!(entry) : undefined,
    onMigrate: callbacks.onMigrate ? () => callbacks.onMigrate!(entry) : undefined,
    onEdit: callbacks.onEdit ? () => callbacks.onEdit!(entry) : undefined,
    onDelete: callbacks.onDelete ? () => callbacks.onDelete!(entry) : undefined,
    onCyclePriority: callbacks.onCyclePriority ? () => callbacks.onCyclePriority!(entry) : undefined,
    onMoveToList: callbacks.onMoveToList ? () => callbacks.onMoveToList!(entry) : undefined,
  });

  const weekDays = days.slice(0, 5).map((day, index) => ({
    ...day,
    dayName: dayNames[index],
    dayNumber: parseISO(day.date).getDate(),
  }));

  const saturday = days[5];
  const sunday = days[6];

  const filteredWeekDays = weekDays.map(day => ({
    ...day,
    entries: filterWeekEntries(day.entries),
  }));

  const filteredSaturday = saturday ? filterWeekEntries(saturday.entries) : [];
  const filteredSunday = sunday ? filterWeekEntries(sunday.entries) : [];

  const getHabitsForDate = (date: string) => {
    return habits
      .map(habit => {
        const dayStatus = habit.dayHistory.find(h => h.date === date);
        return dayStatus && dayStatus.count > 0
          ? { name: habit.name, count: dayStatus.count }
          : null;
      })
      .filter((h): h is { name: string; count: number } => h !== null);
  };

  const startDate = days[0] ? parseISO(days[0].date) : new Date();
  const endDate = days[6] ? parseISO(days[6].date) : new Date();
  const dateRange = `${format(startDate, 'MMM d')} – ${format(endDate, 'MMM d, yyyy')}`;

  const contextTree = useMemo(() => {
    if (contextTreeProp !== undefined) {
      return buildTree(contextTreeProp);
    }
    const allEntries = days.flatMap(day => flattenEntries(day.entries));
    return buildTree(allEntries);
  }, [contextTreeProp, days]);

  return (
    <div className="flex h-full gap-4">
      <div className="flex-1 overflow-y-auto">
        <div className="mb-4">
          <h2 className="text-lg font-semibold">Weekly Review</h2>
          <p className="text-sm text-muted-foreground">{dateRange}</p>
        </div>

        <div className="grid grid-cols-3 gap-4">
          {filteredWeekDays.map((day) => (
            <DayBox
              key={day.date}
              dayNumber={day.dayNumber}
              dayName={day.dayName}
              entries={day.entries}
              habits={getHabitsForDate(day.date)}
              selectedEntry={selectedEntry}
              onSelectEntry={setSelectedEntry}
              createEntryCallbacks={createEntryCallbacks}
            />
          ))}

          {saturday && sunday && (
            <WeekendBox
              saturdayDay={parseISO(saturday.date).getDate()}
              sundayDay={parseISO(sunday.date).getDate()}
              saturdayEntries={filteredSaturday}
              sundayEntries={filteredSunday}
              saturdayHabits={getHabitsForDate(saturday.date)}
              sundayHabits={getHabitsForDate(sunday.date)}
              selectedEntry={selectedEntry}
              onSelectEntry={setSelectedEntry}
              createEntryCallbacks={createEntryCallbacks}
            />
          )}
        </div>
      </div>

      <div className={cn("relative", isContextCollapsed ? "w-12" : "w-96")}>
        {/* Collapse Toggle Button - always shown */}
        <button
          onClick={onToggleContextCollapse}
          aria-label="Toggle context panel"
          className={cn(
            "absolute top-2 p-1.5 hover:bg-secondary rounded-md transition-colors z-10",
            isContextCollapsed ? "left-1/2 -translate-x-1/2" : "right-2"
          )}
        >
          {isContextCollapsed ? (
            <ChevronLeft className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>

        {/* Context Panel Content - hidden when collapsed */}
        {!isContextCollapsed && (
          <div className="border-l border-border pl-4 overflow-y-auto h-full">
            <div className="mb-3">
              <h3 className="text-sm font-medium">Context</h3>
            </div>

            {!selectedEntry ? (
              <p className="text-sm text-muted-foreground">No entry selected</p>
            ) : contextTree.length === 0 ? (
              <p className="text-sm text-muted-foreground">No context</p>
            ) : (
              <ContextTree nodes={contextTree} selectedEntryId={selectedEntry.id} />
            )}
          </div>
        )}
      </div>
    </div>
  );
}
