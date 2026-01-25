import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';

interface WeekendBoxProps {
  startDay: number;
  saturdayEntries: Entry[];
  sundayEntries: Entry[];
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
}

export function WeekendBox({
  startDay,
  saturdayEntries,
  sundayEntries,
  selectedEntry,
  onSelectEntry,
}: WeekendBoxProps) {
  return (
    <div className="border rounded-lg p-3 bg-card">
      <div className="mb-3 flex items-baseline gap-2">
        <span className="text-2xl font-semibold">{startDay}-{startDay + 1}</span>
        <span className="text-sm text-muted-foreground">Weekend</span>
      </div>
      <div className="space-y-1">
        {saturdayEntries.length === 0 && sundayEntries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          <>
            {saturdayEntries.map(entry => (
              <WeekEntry
                key={entry.id}
                entry={entry}
                datePrefix="Sat:"
                isSelected={selectedEntry?.id === entry.id}
                onSelect={() => onSelectEntry?.(entry)}
              />
            ))}
            {sundayEntries.map(entry => (
              <WeekEntry
                key={entry.id}
                entry={entry}
                datePrefix="Sun:"
                isSelected={selectedEntry?.id === entry.id}
                onSelect={() => onSelectEntry?.(entry)}
              />
            ))}
          </>
        )}
      </div>
    </div>
  );
}
