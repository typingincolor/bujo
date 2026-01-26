import { useState, useEffect } from 'react';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { ActionCallbacks } from './EntryActions/types';
import { EntryActionBar } from './EntryActions/EntryActionBar';

interface WeekEntryProps {
  entry: Entry;
  isSelected?: boolean;
  onSelect?: (entry: Entry) => void;
  datePrefix?: string;
  callbacks?: ActionCallbacks;
}

export function WeekEntry({ entry, isSelected, onSelect, datePrefix, callbacks }: WeekEntryProps) {
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];
  const [isHovered, setIsHovered] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (['ArrowUp', 'ArrowDown', 'j', 'k'].includes(e.key)) {
        setIsHovered(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <div
      className={cn(
        'px-2 py-1.5 rounded-lg text-sm transition-colors grid',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      style={{
        gridTemplateRows: callbacks && isHovered ? '1fr auto' : '1fr',
        transition: 'grid-template-rows 150ms ease-out',
      }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <button
        onClick={() => onSelect?.(entry)}
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

      {callbacks && isHovered && (
        <EntryActionBar
          entry={entry}
          callbacks={callbacks}
          isHovered={isHovered}
          variant="hover-reveal"
          size="sm"
          className="mt-1"
        />
      )}
    </div>
  );
}
