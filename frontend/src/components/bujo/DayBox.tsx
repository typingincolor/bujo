import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';
import { HabitItem, HabitDisplay } from './HabitItem';
import { ActionCallbacks } from './EntryActions/types';

interface DayBoxProps {
  dayNumber: number;
  dayName: string;
  entries: Entry[];
  habits?: HabitDisplay[];
  location?: string;
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
  createEntryCallbacks?: (entry: Entry) => ActionCallbacks;
}

export function DayBox({ dayNumber, dayName, entries, habits = [], location, selectedEntry, onSelectEntry, createEntryCallbacks }: DayBoxProps) {
  const hasContent = entries.length > 0 || habits.length > 0;

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3 flex items-baseline gap-2">
        <span className="text-2xl font-semibold">{dayNumber}</span>
        <span className="text-sm text-muted-foreground">{dayName}</span>
        {location && <span className="text-sm text-muted-foreground">{location}</span>}
      </div>

      <div className="space-y-1 max-h-64 overflow-y-auto">
        {!hasContent ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          <>
            {habits.map((habit, index) => (
              <HabitItem
                key={`habit-${habit.name}-${index}`}
                name={habit.name}
                count={habit.count}
              />
            ))}
            {entries.map(entry => (
              <WeekEntry
                key={entry.id}
                entry={entry}
                isSelected={selectedEntry?.id === entry.id}
                onSelect={onSelectEntry}
                callbacks={createEntryCallbacks?.(entry)}
              />
            ))}
          </>
        )}
      </div>
    </div>
  );
}
