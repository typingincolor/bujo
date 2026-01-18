import { HabitDayStatus } from '@/types/bujo';

export interface CalendarDay {
  date: string;
  dayOfWeek: number;
  dayOfMonth: number;
  isToday: boolean;
  isPadding: boolean;
}

export interface QuarterMonth {
  month: number;
  year: number;
  name: string;
  calendar: CalendarDay[][];
}

type PeriodType = 'week' | 'month' | 'quarter';
type Direction = 'prev' | 'next';

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
];

const MONTH_NAMES_SHORT = [
  'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
];

function formatDateString(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function getTodayString(): string {
  return formatDateString(new Date());
}

function startOfWeek(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  d.setDate(d.getDate() - day);
  d.setHours(0, 0, 0, 0);
  return d;
}

function endOfWeek(date: Date): Date {
  const d = startOfWeek(date);
  d.setDate(d.getDate() + 6);
  return d;
}

function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function endOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0);
}

function addDays(date: Date, days: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d;
}

function addMonths(date: Date, months: number): Date {
  const d = new Date(date);
  d.setMonth(d.getMonth() + months);
  return d;
}

export function getWeekDates(anchor: Date): CalendarDay[] {
  const today = getTodayString();
  const weekStart = startOfWeek(anchor);
  const days: CalendarDay[] = [];

  for (let i = 0; i < 7; i++) {
    const d = addDays(weekStart, i);
    const dateStr = formatDateString(d);
    days.push({
      date: dateStr,
      dayOfWeek: d.getDay(),
      dayOfMonth: d.getDate(),
      isToday: dateStr === today,
      isPadding: false,
    });
  }

  return days;
}

export function getMonthCalendar(anchor: Date): CalendarDay[][] {
  const today = getTodayString();
  const monthStart = startOfMonth(anchor);
  const monthEnd = endOfMonth(anchor);
  const calendarStart = startOfWeek(monthStart);
  const targetMonth = anchor.getMonth();

  const rows: CalendarDay[][] = [];
  let currentDate = new Date(calendarStart);

  while (currentDate <= monthEnd || currentDate.getDay() !== 0) {
    const row: CalendarDay[] = [];

    for (let i = 0; i < 7; i++) {
      const dateStr = formatDateString(currentDate);
      const isPadding = currentDate.getMonth() !== targetMonth;

      row.push({
        date: dateStr,
        dayOfWeek: currentDate.getDay(),
        dayOfMonth: currentDate.getDate(),
        isToday: dateStr === today,
        isPadding,
      });

      currentDate = addDays(currentDate, 1);
    }

    rows.push(row);

    // Stop if we've completed the month and are at the start of a new week
    if (currentDate.getMonth() !== targetMonth && currentDate.getDay() === 0) {
      break;
    }
  }

  return rows;
}

export function getQuarterMonths(anchor: Date): QuarterMonth[] {
  const months: QuarterMonth[] = [];

  for (let i = 0; i < 3; i++) {
    const monthDate = addMonths(anchor, i);
    months.push({
      month: monthDate.getMonth(),
      year: monthDate.getFullYear(),
      name: MONTH_NAMES[monthDate.getMonth()],
      calendar: getMonthCalendar(monthDate),
    });
  }

  return months;
}

export function navigatePeriod(anchor: Date, period: PeriodType, direction: Direction): Date {
  const delta = direction === 'next' ? 1 : -1;

  switch (period) {
    case 'week':
      return addDays(anchor, delta * 7);

    case 'month': {
      const newDate = addMonths(anchor, delta);
      return startOfMonth(newDate);
    }

    case 'quarter': {
      const newDate = addMonths(anchor, delta * 3);
      return startOfMonth(newDate);
    }
  }
}

export function mapDayHistoryToCalendar(
  history: HabitDayStatus[]
): Map<string, { completed: boolean; count: number }> {
  const map = new Map<string, { completed: boolean; count: number }>();

  for (const day of history) {
    map.set(day.date, { completed: day.completed, count: day.count });
  }

  return map;
}

export function formatPeriodLabel(anchor: Date, period: PeriodType): string {
  switch (period) {
    case 'week': {
      const weekStart = startOfWeek(anchor);
      const weekEnd = endOfWeek(anchor);
      const startMonth = MONTH_NAMES_SHORT[weekStart.getMonth()];
      const endMonth = MONTH_NAMES_SHORT[weekEnd.getMonth()];
      const year = weekEnd.getFullYear();

      if (startMonth === endMonth) {
        return `${startMonth} ${weekStart.getDate()} - ${endMonth} ${weekEnd.getDate()}, ${year}`;
      }
      return `${startMonth} ${weekStart.getDate()} - ${endMonth} ${weekEnd.getDate()}, ${year}`;
    }

    case 'month': {
      return `${MONTH_NAMES[anchor.getMonth()]} ${anchor.getFullYear()}`;
    }

    case 'quarter': {
      const monthStart = anchor;
      const monthEnd = addMonths(anchor, 2);
      const startMonth = MONTH_NAMES_SHORT[monthStart.getMonth()];
      const endMonth = MONTH_NAMES_SHORT[monthEnd.getMonth()];
      const startYear = monthStart.getFullYear();
      const endYear = monthEnd.getFullYear();

      if (startYear === endYear) {
        return `${startMonth} - ${endMonth} ${endYear}`;
      }
      return `${startMonth} ${startYear} - ${endMonth} ${endYear}`;
    }
  }
}
