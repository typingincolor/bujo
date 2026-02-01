import { EntryType, ENTRY_SYMBOLS, Priority, PRIORITY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';

interface EntrySymbolProps {
  type: EntryType;
  priority?: Priority;
  className?: string;
}

const typeStyles: Record<EntryType, string> = {
  task: 'text-bujo-task',
  note: 'text-bujo-note',
  event: '',
  done: 'text-bujo-done',
  migrated: 'text-bujo-migrated',
  cancelled: 'text-bujo-cancelled',
  question: 'text-bujo-question',
  answered: 'text-bujo-answered',
  answer: 'text-bujo-answer',
  movedToList: 'text-bujo-migrated',
};

const priorityStyles: Record<Priority, string> = {
  none: '',
  low: 'text-priority-low',
  medium: 'text-priority-medium',
  high: 'text-priority-high',
};

export function EntrySymbol({ type, priority = 'none', className }: EntrySymbolProps) {
  return (
    <span className={cn('inline-flex items-center gap-1 font-body', className)}>
      <span className={cn('text-lg font-medium w-5 text-center', typeStyles[type])}>
        {ENTRY_SYMBOLS[type]}
      </span>
      {priority !== 'none' && (
        <span className={cn('text-xs font-bold', priorityStyles[priority])}>
          {PRIORITY_SYMBOLS[priority]}
        </span>
      )}
    </span>
  );
}
