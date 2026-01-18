import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CalendarGrid } from './CalendarGrid';
import { CalendarDay } from '@/lib/calendarUtils';

describe('CalendarGrid', () => {
  // Mock a simple week: Jan 12-18, 2025 (Sun-Sat)
  const mockWeek: CalendarDay[][] = [
    [
      { date: '2025-01-12', dayOfWeek: 0, dayOfMonth: 12, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-13', dayOfWeek: 1, dayOfMonth: 13, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-14', dayOfWeek: 2, dayOfMonth: 14, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-15', dayOfWeek: 3, dayOfMonth: 15, isToday: true, isPadding: false, isFuture: false },
      { date: '2025-01-16', dayOfWeek: 4, dayOfMonth: 16, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-17', dayOfWeek: 5, dayOfMonth: 17, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-18', dayOfWeek: 6, dayOfMonth: 18, isToday: false, isPadding: false, isFuture: false },
    ],
  ];

  // Mock month with padding (Jan 2025 - starts Wednesday)
  const mockMonth: CalendarDay[][] = [
    [
      { date: '2024-12-29', dayOfWeek: 0, dayOfMonth: 29, isToday: false, isPadding: true, isFuture: false },
      { date: '2024-12-30', dayOfWeek: 1, dayOfMonth: 30, isToday: false, isPadding: true, isFuture: false },
      { date: '2024-12-31', dayOfWeek: 2, dayOfMonth: 31, isToday: false, isPadding: true, isFuture: false },
      { date: '2025-01-01', dayOfWeek: 3, dayOfMonth: 1, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-02', dayOfWeek: 4, dayOfMonth: 2, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-03', dayOfWeek: 5, dayOfMonth: 3, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-04', dayOfWeek: 6, dayOfMonth: 4, isToday: false, isPadding: false, isFuture: false },
    ],
    [
      { date: '2025-01-05', dayOfWeek: 0, dayOfMonth: 5, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-06', dayOfWeek: 1, dayOfMonth: 6, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-07', dayOfWeek: 2, dayOfMonth: 7, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-08', dayOfWeek: 3, dayOfMonth: 8, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-09', dayOfWeek: 4, dayOfMonth: 9, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-10', dayOfWeek: 5, dayOfMonth: 10, isToday: false, isPadding: false, isFuture: false },
      { date: '2025-01-11', dayOfWeek: 6, dayOfMonth: 11, isToday: false, isPadding: false, isFuture: false },
    ],
  ];

  const defaultProps = {
    calendar: mockWeek,
    dayHistory: new Map<string, { completed: boolean; count: number }>([
      ['2025-01-15', { completed: true, count: 2 }],
    ]),
    onLog: vi.fn(),
    onDecrement: vi.fn(),
  };

  it('renders day-of-week header', () => {
    render(<CalendarGrid {...defaultProps} />);

    // Check that all expected labels are present
    expect(screen.getAllByText('S')).toHaveLength(2); // Sunday, Saturday
    expect(screen.getByText('M')).toBeInTheDocument();
    expect(screen.getAllByText('T')).toHaveLength(2); // Tuesday, Thursday
    expect(screen.getByText('W')).toBeInTheDocument();
    expect(screen.getByText('F')).toBeInTheDocument();
  });

  it('renders all days in a single-row week view', () => {
    render(<CalendarGrid {...defaultProps} />);

    // Should have 7 day buttons
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(7);
  });

  it('renders all days in a multi-row month view', () => {
    render(<CalendarGrid {...defaultProps} calendar={mockMonth} />);

    // Should have 14 day buttons (2 rows x 7)
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(14);
  });

  it('shows blank circles for uncompleted days', () => {
    render(<CalendarGrid {...defaultProps} />);

    // Days without completion show blank (empty circle)
    const day12Button = screen.getByLabelText(/Log for 2025-01-12/);
    const day13Button = screen.getByLabelText(/Log for 2025-01-13/);
    expect(day12Button.textContent).toBe('');
    expect(day13Button.textContent).toBe('');
  });

  it('shows count for completed days', () => {
    render(<CalendarGrid {...defaultProps} />);

    // Day with count=2 should show the count
    expect(screen.getByText('2')).toBeInTheDocument();
  });

  it('calls onLog with date when day is clicked', () => {
    const onLog = vi.fn();
    render(<CalendarGrid {...defaultProps} onLog={onLog} />);

    // Click on day 12 using aria-label
    fireEvent.click(screen.getByLabelText(/Log for 2025-01-12/));

    expect(onLog).toHaveBeenCalledWith('2025-01-12');
  });

  it('calls onDecrement with date when day is cmd-clicked', () => {
    const onDecrement = vi.fn();
    render(<CalendarGrid {...defaultProps} onDecrement={onDecrement} />);

    // Cmd-click on completed day (showing count 2)
    fireEvent.click(screen.getByText('2'), { metaKey: true });

    expect(onDecrement).toHaveBeenCalledWith('2025-01-15');
  });

  it('hides header when showHeader is false', () => {
    render(<CalendarGrid {...defaultProps} showHeader={false} />);

    // Header labels should not be present
    expect(screen.queryByText('S')).toBeNull();
    expect(screen.queryByText('M')).toBeNull();
  });

  describe('compact mode', () => {
    it('uses compact styling in quarter view', () => {
      render(<CalendarGrid {...defaultProps} compact />);

      // Buttons should have compact class
      const buttons = screen.getAllByRole('button');
      buttons.forEach(button => {
        expect(button).toHaveClass('w-5', 'h-5');
      });
    });
  });
});
