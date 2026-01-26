import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';

interface DayBoxProps {
  dayNumber: number;
  dayName: string;
  entries: Entry[];
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
}

export function DayBox({ dayNumber, dayName, entries, selectedEntry, onSelectEntry }: DayBoxProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3 flex items-baseline gap-2">
        <span className="text-2xl font-semibold">{dayNumber}</span>
        <span className="text-sm text-muted-foreground">{dayName}</span>
      </div>

      <div className="space-y-1 max-h-64 overflow-y-auto">
        {entries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No events</p>
        ) : (
          entries.map(entry => (
            <WeekEntry
              key={entry.id}
              entry={entry}
              isSelected={selectedEntry?.id === entry.id}
              onSelect={onSelectEntry}
            />
          ))
        )}
      </div>
    </div>
  );
}
