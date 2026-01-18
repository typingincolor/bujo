import { CalendarDay } from '@/lib/calendarUtils';
import { DayCell } from './DayCell';

const DAY_LABELS = ['S', 'M', 'T', 'W', 'T', 'F', 'S'];

export interface CalendarGridProps {
  calendar: CalendarDay[][];
  dayHistory: Map<string, { completed: boolean; count: number }>;
  onLog: (date: string) => void;
  onDecrement: (date: string) => void;
  showHeader?: boolean;
  compact?: boolean;
}

export function CalendarGrid({
  calendar,
  dayHistory,
  onLog,
  onDecrement,
  showHeader = true,
  compact = false,
}: CalendarGridProps) {
  return (
    <div className="flex flex-col gap-px p-0.5">
      {showHeader && (
        <div className="grid grid-cols-7 gap-px mb-1">
          {DAY_LABELS.map((label, i) => (
            <div
              key={i}
              className="w-full flex justify-center text-[10px] text-muted-foreground font-medium"
            >
              {label}
            </div>
          ))}
        </div>
      )}

      {calendar.map((row, rowIndex) => (
        <div key={rowIndex} className="grid grid-cols-7 gap-px">
          {row.map((day) => {
            const dayData = dayHistory.get(day.date);
            const count = dayData?.count ?? 0;
            const completed = dayData?.completed ?? false;

            return (
              <div key={day.date} className="flex justify-center">
                <DayCell
                  day={day}
                  count={count}
                  completed={completed}
                  onLog={onLog}
                  onDecrement={onDecrement}
                  compact={compact}
                />
              </div>
            );
          })}
        </div>
      ))}
    </div>
  );
}
