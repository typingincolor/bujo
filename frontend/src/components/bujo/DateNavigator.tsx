import { useState, useEffect } from 'react';
import { ChevronLeft, ChevronRight, Calendar } from 'lucide-react';
import { format, addDays, subDays, isSameDay, startOfMonth, endOfMonth, startOfWeek, endOfWeek, eachDayOfInterval, addMonths, subMonths } from 'date-fns';
import * as Popover from '@radix-ui/react-popover';
import { cn } from '@/lib/utils';

interface DateNavigatorProps {
  date: Date;
  onDateChange: (date: Date) => void;
}

export function DateNavigator({ date, onDateChange }: DateNavigatorProps) {
  const [isCalendarOpen, setIsCalendarOpen] = useState(false);
  const [calendarMonth, setCalendarMonth] = useState(date);
  const today = new Date();
  const isToday = isSameDay(date, today);

  useEffect(() => {
    setCalendarMonth(date);
  }, [date]);

  const handlePrevDay = () => {
    onDateChange(subDays(date, 1));
  };

  const handleNextDay = () => {
    onDateChange(addDays(date, 1));
  };

  const handleJumpToToday = () => {
    onDateChange(today);
  };

  const handleDateSelect = (selectedDate: Date) => {
    onDateChange(selectedDate);
    setIsCalendarOpen(false);
  };

  const handlePrevMonth = () => {
    setCalendarMonth(subMonths(calendarMonth, 1));
  };

  const handleNextMonth = () => {
    setCalendarMonth(addMonths(calendarMonth, 1));
  };

  const monthStart = startOfMonth(calendarMonth);
  const monthEnd = endOfMonth(calendarMonth);
  const calendarStart = startOfWeek(monthStart);
  const calendarEnd = endOfWeek(monthEnd);
  const calendarDays = eachDayOfInterval({ start: calendarStart, end: calendarEnd });

  return (
    <div className="flex items-center gap-2">
      <button
        onClick={handlePrevDay}
        aria-label="Previous day"
        className="h-8 w-8 flex items-center justify-center rounded-lg hover:bg-secondary/50 transition-colors"
      >
        <ChevronLeft className="h-4 w-4" />
      </button>

      <Popover.Root open={isCalendarOpen} onOpenChange={setIsCalendarOpen}>
        <Popover.Trigger asChild>
          <button
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-secondary/50 transition-colors min-w-[10rem]"
            aria-label={isToday ? 'Today' : format(date, 'EEE, MMM d, yyyy')}
            data-testid="date-picker"
          >
            <Calendar className="h-4 w-4" />
            <span className="text-sm font-medium">
              {isToday ? 'Today' : format(date, 'EEE, MMM d, yyyy')}
            </span>
          </button>
        </Popover.Trigger>
        <Popover.Portal>
          <Popover.Content
            role="dialog"
            className="z-50 bg-card border border-border rounded-lg shadow-lg p-3"
            sideOffset={5}
          >
            <div className="flex items-center justify-between mb-2">
              <button
                onClick={handlePrevMonth}
                className="h-7 w-7 flex items-center justify-center rounded hover:bg-secondary/50"
                aria-label="Previous month"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <span className="text-sm font-medium">
                {format(calendarMonth, 'MMMM yyyy')}
              </span>
              <button
                onClick={handleNextMonth}
                className="h-7 w-7 flex items-center justify-center rounded hover:bg-secondary/50"
                aria-label="Next month"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
            <div className="grid grid-cols-7 gap-1 text-center text-xs text-muted-foreground mb-1">
              {['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'].map(day => (
                <div key={day} className="h-7 flex items-center justify-center">
                  {day}
                </div>
              ))}
            </div>
            <div className="grid grid-cols-7 gap-1" role="grid">
              {calendarDays.map(day => {
                const isSelected = isSameDay(day, date);
                const isCurrentMonth = day.getMonth() === calendarMonth.getMonth();
                const isTodayCell = isSameDay(day, today);

                return (
                  <button
                    key={day.toISOString()}
                    role="gridcell"
                    onClick={() => handleDateSelect(day)}
                    className={cn(
                      'h-7 w-7 flex items-center justify-center rounded text-sm transition-colors',
                      isSelected && 'bg-primary text-primary-foreground',
                      !isSelected && isCurrentMonth && 'hover:bg-secondary/50',
                      !isSelected && !isCurrentMonth && 'text-muted-foreground/50 hover:bg-secondary/30',
                      isTodayCell && !isSelected && 'border border-primary'
                    )}
                  >
                    {format(day, 'd')}
                  </button>
                );
              })}
            </div>
          </Popover.Content>
        </Popover.Portal>
      </Popover.Root>

      <button
        onClick={handleNextDay}
        aria-label="Next day"
        className="h-8 w-8 flex items-center justify-center rounded-lg hover:bg-secondary/50 transition-colors"
      >
        <ChevronRight className="h-4 w-4" />
      </button>

      <button
        onClick={handleJumpToToday}
        aria-label="Jump to today"
        data-testid="jump-to-today"
        className={cn(
          'px-2 py-1 text-xs rounded hover:bg-secondary/50 transition-colors',
          isToday && 'invisible'
        )}
      >
        Today
      </button>
    </div>
  );
}
