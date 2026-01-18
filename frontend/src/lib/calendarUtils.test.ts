import { describe, it, expect } from 'vitest';
import {
  getWeekDates,
  getMonthCalendar,
  getQuarterMonths,
  navigatePeriod,
  mapDayHistoryToCalendar,
  formatPeriodLabel,
  CalendarDay,
} from './calendarUtils';
import { HabitDayStatus } from '@/types/bujo';

describe('calendarUtils', () => {
  describe('getWeekDates', () => {
    it('returns 7 days starting from Sunday', () => {
      // Wednesday Jan 15, 2025
      const anchor = new Date(2025, 0, 15);
      const week = getWeekDates(anchor);

      expect(week).toHaveLength(7);
      // Should start on Sunday Jan 12
      expect(week[0].date).toBe('2025-01-12');
      expect(week[0].dayOfWeek).toBe(0); // Sunday
      // Wednesday should be at index 3
      expect(week[3].date).toBe('2025-01-15');
      expect(week[3].dayOfWeek).toBe(3);
      // Should end on Saturday Jan 18
      expect(week[6].date).toBe('2025-01-18');
      expect(week[6].dayOfWeek).toBe(6); // Saturday
    });

    it('handles week spanning month boundary', () => {
      // Thursday Jan 30, 2025
      const anchor = new Date(2025, 0, 30);
      const week = getWeekDates(anchor);

      expect(week).toHaveLength(7);
      // Sunday Jan 26
      expect(week[0].date).toBe('2025-01-26');
      // Saturday Feb 1
      expect(week[6].date).toBe('2025-02-01');
    });

    it('marks today correctly', () => {
      const today = new Date();
      const week = getWeekDates(today);

      const todayStr = today.toISOString().split('T')[0];
      const todayDay = week.find(d => d.date === todayStr);
      expect(todayDay?.isToday).toBe(true);

      // Other days should not be today
      const otherDays = week.filter(d => d.date !== todayStr);
      otherDays.forEach(d => expect(d.isToday).toBe(false));
    });
  });

  describe('getMonthCalendar', () => {
    it('returns calendar grid with padding days', () => {
      // January 2025 starts on Wednesday
      const anchor = new Date(2025, 0, 15);
      const calendar = getMonthCalendar(anchor);

      // Should have rows for the full calendar grid
      expect(calendar.length).toBeGreaterThanOrEqual(4);
      expect(calendar.length).toBeLessThanOrEqual(6);

      // Each row should have 7 days
      calendar.forEach(row => {
        expect(row).toHaveLength(7);
      });

      // First row starts with Sunday
      expect(calendar[0][0].dayOfWeek).toBe(0);

      // Find Jan 1 - it should be on Wednesday (index 3 of first row)
      const jan1 = calendar.flat().find(d => d.date === '2025-01-01');
      expect(jan1).toBeDefined();
      expect(jan1?.dayOfWeek).toBe(3); // Wednesday
      expect(jan1?.isPadding).toBe(false);
      expect(jan1?.dayOfMonth).toBe(1);
    });

    it('marks padding days from previous/next month', () => {
      // January 2025 starts on Wednesday
      const anchor = new Date(2025, 0, 15);
      const calendar = getMonthCalendar(anchor);

      // First 3 days of first row should be padding (Dec 29, 30, 31)
      expect(calendar[0][0].isPadding).toBe(true);
      expect(calendar[0][0].date).toBe('2024-12-29');
      expect(calendar[0][1].isPadding).toBe(true);
      expect(calendar[0][2].isPadding).toBe(true);
      expect(calendar[0][3].isPadding).toBe(false);
      expect(calendar[0][3].date).toBe('2025-01-01');
    });

    it('handles February with 28 days', () => {
      // February 2025 (non-leap year) has 28 days
      const anchor = new Date(2025, 1, 15);
      const calendar = getMonthCalendar(anchor);

      // Find all non-padding days
      const febDays = calendar.flat().filter(d => !d.isPadding);
      expect(febDays).toHaveLength(28);

      // Feb 1 is Saturday
      const feb1 = calendar.flat().find(d => d.date === '2025-02-01');
      expect(feb1?.dayOfWeek).toBe(6);
    });

    it('handles month with 6 rows needed', () => {
      // March 2025 - starts on Saturday and has 31 days
      const anchor = new Date(2025, 2, 15);
      const calendar = getMonthCalendar(anchor);

      // March 1 is Saturday (index 6)
      // 31 days spanning 6 weeks
      expect(calendar).toHaveLength(6);
    });
  });

  describe('getQuarterMonths', () => {
    it('returns 3 months ending with anchor month (shows past)', () => {
      const anchor = new Date(2025, 0, 15); // January
      const quarters = getQuarterMonths(anchor);

      expect(quarters).toHaveLength(3);

      // First month: November 2024 (2 months ago)
      expect(quarters[0].month).toBe(10);
      expect(quarters[0].year).toBe(2024);
      expect(quarters[0].name).toBe('November');

      // Second month: December 2024 (1 month ago)
      expect(quarters[1].month).toBe(11);
      expect(quarters[1].year).toBe(2024);
      expect(quarters[1].name).toBe('December');

      // Third month: January 2025 (current)
      expect(quarters[2].month).toBe(0);
      expect(quarters[2].year).toBe(2025);
      expect(quarters[2].name).toBe('January');
    });

    it('handles year boundary (March shows Jan-Mar)', () => {
      // March 2025 - quarter shows Jan, Feb, Mar
      const anchor = new Date(2025, 2, 15);
      const quarters = getQuarterMonths(anchor);

      expect(quarters[0].month).toBe(0); // January
      expect(quarters[0].year).toBe(2025);

      expect(quarters[1].month).toBe(1); // February
      expect(quarters[1].year).toBe(2025);

      expect(quarters[2].month).toBe(2); // March
      expect(quarters[2].year).toBe(2025);
    });

    it('includes calendar grid for each month', () => {
      const anchor = new Date(2025, 0, 15);
      const quarters = getQuarterMonths(anchor);

      quarters.forEach(q => {
        expect(q.calendar.length).toBeGreaterThanOrEqual(4);
        expect(q.calendar.length).toBeLessThanOrEqual(6);
        q.calendar.forEach(row => {
          expect(row).toHaveLength(7);
        });
      });
    });
  });

  describe('navigatePeriod', () => {
    it('navigates week forward by 7 days', () => {
      const anchor = new Date(2025, 0, 15);
      const next = navigatePeriod(anchor, 'week', 'next');

      expect(next.getDate()).toBe(22);
      expect(next.getMonth()).toBe(0);
    });

    it('navigates week backward by 7 days', () => {
      const anchor = new Date(2025, 0, 15);
      const prev = navigatePeriod(anchor, 'week', 'prev');

      expect(prev.getDate()).toBe(8);
      expect(prev.getMonth()).toBe(0);
    });

    it('navigates month forward to first of next month', () => {
      const anchor = new Date(2025, 0, 15);
      const next = navigatePeriod(anchor, 'month', 'next');

      expect(next.getDate()).toBe(1);
      expect(next.getMonth()).toBe(1); // February
    });

    it('navigates month backward to first of previous month', () => {
      const anchor = new Date(2025, 1, 15);
      const prev = navigatePeriod(anchor, 'month', 'prev');

      expect(prev.getDate()).toBe(1);
      expect(prev.getMonth()).toBe(0); // January
    });

    it('navigates quarter forward by 3 months', () => {
      const anchor = new Date(2025, 0, 15);
      const next = navigatePeriod(anchor, 'quarter', 'next');

      expect(next.getDate()).toBe(1);
      expect(next.getMonth()).toBe(3); // April
    });

    it('navigates quarter backward by 3 months', () => {
      const anchor = new Date(2025, 3, 15);
      const prev = navigatePeriod(anchor, 'quarter', 'prev');

      expect(prev.getDate()).toBe(1);
      expect(prev.getMonth()).toBe(0); // January
    });

    it('handles year boundary when navigating', () => {
      const anchor = new Date(2025, 0, 15);
      const prev = navigatePeriod(anchor, 'month', 'prev');

      expect(prev.getMonth()).toBe(11); // December
      expect(prev.getFullYear()).toBe(2024);
    });
  });

  describe('mapDayHistoryToCalendar', () => {
    it('creates lookup map by date string', () => {
      const history: HabitDayStatus[] = [
        { date: '2025-01-15', completed: true, count: 2 },
        { date: '2025-01-14', completed: true, count: 1 },
        { date: '2025-01-13', completed: false, count: 0 },
      ];

      const map = mapDayHistoryToCalendar(history);

      expect(map.get('2025-01-15')).toEqual({ completed: true, count: 2 });
      expect(map.get('2025-01-14')).toEqual({ completed: true, count: 1 });
      expect(map.get('2025-01-13')).toEqual({ completed: false, count: 0 });
      expect(map.get('2025-01-12')).toBeUndefined();
    });

    it('handles empty history', () => {
      const map = mapDayHistoryToCalendar([]);
      expect(map.size).toBe(0);
    });
  });

  describe('formatPeriodLabel', () => {
    it('formats week as date range', () => {
      const anchor = new Date(2025, 0, 15);
      const label = formatPeriodLabel(anchor, 'week');

      expect(label).toBe('Jan 12 - Jan 18, 2025');
    });

    it('formats week spanning months', () => {
      const anchor = new Date(2025, 0, 30);
      const label = formatPeriodLabel(anchor, 'week');

      expect(label).toBe('Jan 26 - Feb 1, 2025');
    });

    it('formats month as month name and year', () => {
      const anchor = new Date(2025, 0, 15);
      const label = formatPeriodLabel(anchor, 'month');

      expect(label).toBe('January 2025');
    });

    it('formats quarter as month range (past to current)', () => {
      const anchor = new Date(2025, 0, 15);
      const label = formatPeriodLabel(anchor, 'quarter');

      expect(label).toBe('Nov 2024 - Jan 2025');
    });

    it('formats quarter within same year', () => {
      const anchor = new Date(2025, 5, 15); // June
      const label = formatPeriodLabel(anchor, 'quarter');

      expect(label).toBe('Apr - Jun 2025');
    });
  });
});
