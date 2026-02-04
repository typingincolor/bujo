import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';
import { HabitItem, HabitDisplay } from './HabitItem';

interface WeekendBoxProps {
  saturdayDay: number;
  sundayDay: number;
  saturdayEntries: Entry[];
  sundayEntries: Entry[];
  saturdayHabits?: HabitDisplay[];
  sundayHabits?: HabitDisplay[];
  saturdayLocation?: string;
  sundayLocation?: string;
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
  onNavigateToEntry?: (entry: Entry) => void;
}

export function WeekendBox({
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
  onNavigateToEntry,
}: WeekendBoxProps) {
  const hasContent = saturdayEntries.length > 0 || sundayEntries.length > 0 || saturdayHabits.length > 0 || sundayHabits.length > 0;

  return (
    <div className="border rounded-lg p-3 bg-card">
      <div className="mb-3 flex items-baseline gap-2">
        <span className="text-lg font-semibold">{saturdayDay} - {sundayDay}</span>
        <span className="text-sm text-muted-foreground">Weekend</span>
        {(saturdayLocation || sundayLocation) && (
          <span className="text-sm text-muted-foreground">
            ({saturdayLocation || 'not set'} / {sundayLocation || 'not set'})
          </span>
        )}
      </div>
      <div className="space-y-1">
        {!hasContent ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          <>
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
                onNavigateToEntry={onNavigateToEntry}
              />
            ))}
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
                onNavigateToEntry={onNavigateToEntry}
              />
            ))}
          </>
        )}
      </div>
    </div>
  );
}
