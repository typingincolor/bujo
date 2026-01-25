import { Entry } from '@/types/bujo';
import { WeekEntry } from './WeekEntry';

interface WeekendBoxProps {
  saturdayEntries: Entry[];
  sundayEntries: Entry[];
  selectedEntryId?: number;
  onEntrySelect?: (entryId: number) => void;
}

export function WeekendBox({
  saturdayEntries,
  sundayEntries,
  selectedEntryId,
  onEntrySelect,
}: WeekendBoxProps) {
  return (
    <div className="border rounded-lg p-3 bg-card">
      <h3 className="text-sm font-semibold mb-2 text-muted-foreground">
        Weekend
      </h3>
      <div className="space-y-1">
        {saturdayEntries.map(entry => (
          <WeekEntry
            key={entry.id}
            entry={entry}
            datePrefix="Sat:"
            isSelected={entry.id === selectedEntryId}
            onSelect={() => onEntrySelect?.(entry.id)}
          />
        ))}
        {sundayEntries.map(entry => (
          <WeekEntry
            key={entry.id}
            entry={entry}
            datePrefix="Sun:"
            isSelected={entry.id === selectedEntryId}
            onSelect={() => onEntrySelect?.(entry.id)}
          />
        ))}
      </div>
    </div>
  );
}
