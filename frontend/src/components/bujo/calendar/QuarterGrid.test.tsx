import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { QuarterGrid } from './QuarterGrid';
import { QuarterMonth, CalendarDay } from '@/lib/calendarUtils';

describe('QuarterGrid', () => {
  // Simple mock month calendar (just first week for brevity)
  const mockCalendar: CalendarDay[][] = [
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

  const mockQuarters: QuarterMonth[] = [
    { month: 0, year: 2025, name: 'January', calendar: mockCalendar },
    { month: 1, year: 2025, name: 'February', calendar: mockCalendar },
    { month: 2, year: 2025, name: 'March', calendar: mockCalendar },
  ];

  const defaultProps = {
    quarters: mockQuarters,
    dayHistory: new Map<string, { completed: boolean; count: number }>([
      ['2025-01-08', { completed: true, count: 1 }],
    ]),
    onLog: vi.fn(),
    onDecrement: vi.fn(),
  };

  it('renders all three month names', () => {
    render(<QuarterGrid {...defaultProps} />);

    expect(screen.getByText('January')).toBeInTheDocument();
    expect(screen.getByText('February')).toBeInTheDocument();
    expect(screen.getByText('March')).toBeInTheDocument();
  });

  it('renders three CalendarGrid components', () => {
    render(<QuarterGrid {...defaultProps} />);

    // Each month has 7 days, so 21 total buttons
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(21);
  });

  it('passes compact prop to CalendarGrids', () => {
    render(<QuarterGrid {...defaultProps} />);

    // In quarter view, cells should be compact
    const buttons = screen.getAllByRole('button');
    buttons.forEach(button => {
      expect(button).toHaveClass('w-5', 'h-5');
    });
  });

  it('shows day history data across all months', () => {
    render(<QuarterGrid {...defaultProps} />);

    // Day 8 appears 3 times (once per mock month), and shows count=1
    // Since same date is used in each mock, we get 3 buttons with count 1
    const countButtons = screen.getAllByText('1');
    expect(countButtons.length).toBeGreaterThan(0);
  });

  it('calls onLog with correct date', () => {
    const onLog = vi.fn();
    render(<QuarterGrid {...defaultProps} onLog={onLog} />);

    // Click on first instance of day 5 using aria-label
    const day5Buttons = screen.getAllByLabelText(/Log for 2025-01-05/);
    fireEvent.click(day5Buttons[0]);

    expect(onLog).toHaveBeenCalledWith('2025-01-05');
  });

  it('calls onDecrement with correct date', () => {
    const onDecrement = vi.fn();
    render(<QuarterGrid {...defaultProps} onDecrement={onDecrement} />);

    // Cmd-click on first logged day (showing count 1)
    const countButtons = screen.getAllByText('1');
    fireEvent.click(countButtons[0], { metaKey: true });

    expect(onDecrement).toHaveBeenCalledWith('2025-01-08');
  });

  it('does not show header row for individual months', () => {
    render(<QuarterGrid {...defaultProps} />);

    // Day labels should not appear (no S, M, T, etc. in header)
    // The month names contain letters, but the header should be hidden
    expect(screen.queryAllByText('S')).toHaveLength(0);
    expect(screen.queryAllByText('T')).toHaveLength(0);
  });

  it('renders months in horizontal layout', () => {
    const { container } = render(<QuarterGrid {...defaultProps} />);

    // Check that the container uses flex layout
    const grid = container.firstChild;
    expect(grid).toHaveClass('flex');
  });
});
