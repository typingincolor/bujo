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
    <div
      className={cn(
        'px-2 py-1.5 rounded-lg text-sm transition-colors',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
    >
      <button
        onClick={onSelect}
        className="flex items-center gap-2 text-left min-w-0 w-full"
      >
        {datePrefix && (
          <span className="text-muted-foreground text-xs flex-shrink-0">
            {datePrefix}
          </span>
        )}

        <span className="text-muted-foreground flex-shrink-0">
          {symbol}
        </span>

        {prioritySymbol && (
          <span className="text-orange-500 font-medium flex-shrink-0">
            {prioritySymbol}
          </span>
        )}

        <span className="flex-1 truncate">{entry.content}</span>
      </button>
    </div>
  );
}
