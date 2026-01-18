import { cn } from '@/lib/utils';
import { CalendarDay } from '@/lib/calendarUtils';

export interface DayCellProps {
  day: CalendarDay;
  count: number;
  completed: boolean;
  onLog: (date: string) => void;
  onDecrement: (date: string) => void;
  compact?: boolean;
}

export function DayCell({
  day,
  count,
  completed,
  onLog,
  onDecrement,
  compact = false,
}: DayCellProps) {
  // Don't render future dates - return empty placeholder for grid layout
  if (day.isFuture) {
    return <div className={compact ? 'w-5 h-5' : 'w-6 h-6'} />;
  }

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();

    if (e.metaKey || e.ctrlKey) {
      if (count > 0) {
        onDecrement(day.date);
      } else {
        onLog(day.date);
      }
    } else {
      onLog(day.date);
    }
  };

  const ariaLabel = `${completed ? 'Logged' : 'Log'} for ${day.date}${count > 0 ? ` (${count})` : ''}`;

  return (
    <button
      onClick={handleClick}
      aria-label={ariaLabel}
      title={`Click to add, ${navigator.platform.includes('Mac') ? '⌘' : 'Ctrl'}+click to remove`}
      className={cn(
        'rounded-full flex items-center justify-center text-[10px] font-medium',
        compact ? 'w-5 h-5' : 'w-6 h-6',
        completed ? 'bg-bujo-habit-fill text-primary-foreground' : 'bg-bujo-habit-empty text-muted-foreground',
        day.isToday && 'ring-2 ring-bujo-today',
        day.isPadding && 'opacity-30',
        'cursor-pointer'
      )}
    >
      {count > 0 ? count : null}
    </button>
  );
}
