import { QuarterMonth } from '@/lib/calendarUtils';
import { CalendarGrid } from './CalendarGrid';

export interface QuarterGridProps {
  quarters: QuarterMonth[];
  dayHistory: Map<string, { completed: boolean; count: number }>;
  onLog: (date: string) => void;
  onDecrement: (date: string) => void;
}

export function QuarterGrid({
  quarters,
  dayHistory,
  onLog,
  onDecrement,
}: QuarterGridProps) {
  return (
    <div className="flex gap-4">
      {quarters.map((quarter) => (
        <div key={`${quarter.year}-${quarter.month}`} className="flex-1">
          <div className="text-xs font-medium text-center mb-1">{quarter.name}</div>
          <CalendarGrid
            calendar={quarter.calendar}
            dayHistory={dayHistory}
            onLog={onLog}
            onDecrement={onDecrement}
            showHeader={false}
            compact
          />
        </div>
      ))}
    </div>
  );
}
