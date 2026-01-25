import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';
import { format } from 'date-fns';

interface DayBoxProps {
  date: Date;
  entries: Entry[];
  selectedEntryId?: number;
  onEntrySelect?: (entryId: number) => void;
}

export function DayBox({ date, entries, selectedEntryId, onEntrySelect }: DayBoxProps) {
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
            isSelected={entry.id === selectedEntryId}
            onSelect={() => onEntrySelect?.(entry.id)}
          />
        ))}
      </div>
    </div>
  );
}
