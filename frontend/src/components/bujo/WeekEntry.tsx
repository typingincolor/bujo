import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';

interface WeekEntryProps {
  entry: Entry;
  isSelected?: boolean;
  onSelect?: () => void;
  datePrefix?: string;
}

export function WeekEntry({ entry, isSelected, onSelect, datePrefix }: WeekEntryProps) {
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];

  return (
    <div className={cn(isSelected && 'bg-primary/10')}>
      <button onClick={onSelect}>
        {datePrefix && <span>{datePrefix}</span>}
        <span>{symbol}</span>
        {prioritySymbol && <span>{prioritySymbol}</span>}
        <span>{entry.content}</span>
      </button>
    </div>
  );
}
