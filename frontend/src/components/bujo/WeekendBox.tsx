import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';
import { HabitItem, HabitDisplay } from './HabitItem';
import { ActionCallbacks } from './EntryActions/types';

interface WeekendBoxProps {
  startDay?: number;
  saturdayDay?: number;
  sundayDay?: number;
  saturdayEntries: Entry[];
  sundayEntries: Entry[];
  saturdayHabits?: HabitDisplay[];
  sundayHabits?: HabitDisplay[];
  saturdayLocation?: string;
  sundayLocation?: string;
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
  createEntryCallbacks?: (entry: Entry) => ActionCallbacks;
}

export function WeekendBox({
  startDay,
  saturdayDay,
  sundayDay,
  saturdayEntries,
  sundayEntries,
  saturdayHabits = [],
  sundayHabits = [],
  saturdayLocation,
  sundayLocation,
  selectedEntry,
  onSelectEntry,
  createEntryCallbacks,
}: WeekendBoxProps) {
  const hasContent = saturdayEntries.length > 0 || sundayEntries.length > 0 || saturdayHabits.length > 0 || sundayHabits.length > 0 || saturdayLocation || sundayLocation;

  const satDay = saturdayDay ?? startDay ?? 0;
  const sunDay = sundayDay ?? (startDay ? startDay + 1 : 0);

  return (
    <div className="border rounded-lg p-3 bg-card">
      <div className="mb-3 flex items-baseline gap-2">
        <span className="text-2xl font-semibold">{satDay}-{sunDay}</span>
        <span className="text-sm text-muted-foreground">Weekend</span>
      </div>
      <div className="space-y-1">
        {!hasContent ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          <>
            {saturdayLocation && (
              <p className="text-sm text-muted-foreground">Sat: {saturdayLocation}</p>
            )}
            {saturdayHabits.map((habit, index) => (
              <HabitItem
                key={`habit-sat-${habit.name}-${index}`}
                name={habit.name}
                count={habit.count}
                datePrefix="Sat:"
              />
            ))}
            {saturdayEntries.map(entry => (
              <WeekEntry
                key={entry.id}
                entry={entry}
                datePrefix="Sat:"
                isSelected={selectedEntry?.id === entry.id}
                onSelect={onSelectEntry}
                callbacks={createEntryCallbacks?.(entry)}
              />
            ))}
            {sundayLocation && (
              <p className="text-sm text-muted-foreground">Sun: {sundayLocation}</p>
            )}
            {sundayHabits.map((habit, index) => (
              <HabitItem
                key={`habit-sun-${habit.name}-${index}`}
                name={habit.name}
                count={habit.count}
                datePrefix="Sun:"
              />
            ))}
            {sundayEntries.map(entry => (
              <WeekEntry
                key={entry.id}
                entry={entry}
                datePrefix="Sun:"
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
