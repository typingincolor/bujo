import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';
import { format } from 'date-fns';

interface DayBoxProps {
  date: Date;
  entries: Entry[];
  selectedEntry?: Entry;
  onSelectEntry?: (entry: Entry) => void;
}

export function DayBox({ date, entries, selectedEntry, onSelectEntry }: DayBoxProps) {
  const label = format(date, 'EEE M/d');

  return (
    <div className="border rounded-lg p-3 bg-card">
      <h3 className="text-sm font-semibold mb-2 text-muted-foreground">
        {label}
      </h3>
      <div className="space-y-1">
        {entries.map(entry => (
          <WeekEntry
            key={entry.id}
            entry={entry}
            isSelected={selectedEntry?.id === entry.id}
            onSelect={() => onSelectEntry?.(entry)}
          />
        ))}
      </div>
    </div>
  );
}
