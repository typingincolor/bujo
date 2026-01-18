import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DayCell, DayCellProps } from './DayCell';
import { CalendarDay } from '@/lib/calendarUtils';

describe('DayCell', () => {
  const defaultDay: CalendarDay = {
    date: '2025-01-15',
    dayOfWeek: 3,
    dayOfMonth: 15,
    isToday: false,
    isPadding: false,
    isFuture: false,
  };

  const defaultProps: DayCellProps = {
    day: defaultDay,
    count: 0,
    completed: false,
    onLog: vi.fn(),
    onDecrement: vi.fn(),
  };

  it('renders blank when count is 0', () => {
    render(<DayCell {...defaultProps} count={0} />);
    const button = screen.getByRole('button');
    expect(button.textContent).toBe('');
  });

  it('displays count when greater than 0', () => {
    render(<DayCell {...defaultProps} count={3} completed />);
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('applies completed styling when completed', () => {
    render(<DayCell {...defaultProps} completed count={1} />);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('bg-bujo-habit-fill');
  });

  it('applies empty styling when not completed', () => {
    render(<DayCell {...defaultProps} completed={false} count={0} />);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('bg-bujo-habit-empty');
  });

  it('applies today highlight when isToday', () => {
    const todayDay = { ...defaultDay, isToday: true };
    render(<DayCell {...defaultProps} day={todayDay} />);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('ring-bujo-today');
  });

  it('applies padding styling when isPadding', () => {
    const paddingDay = { ...defaultDay, isPadding: true };
    render(<DayCell {...defaultProps} day={paddingDay} />);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('opacity-30');
  });

  it('calls onLog when clicked', () => {
    const onLog = vi.fn();
    render(<DayCell {...defaultProps} onLog={onLog} />);

    fireEvent.click(screen.getByRole('button'));

    expect(onLog).toHaveBeenCalledWith('2025-01-15');
    expect(onLog).toHaveBeenCalledTimes(1);
  });

  it('calls onDecrement when cmd-clicked (Mac)', () => {
    const onDecrement = vi.fn();
    render(<DayCell {...defaultProps} count={1} completed onDecrement={onDecrement} />);

    fireEvent.click(screen.getByRole('button'), { metaKey: true });

    expect(onDecrement).toHaveBeenCalledWith('2025-01-15');
    expect(onDecrement).toHaveBeenCalledTimes(1);
  });

  it('calls onDecrement when ctrl-clicked (Windows/Linux)', () => {
    const onDecrement = vi.fn();
    render(<DayCell {...defaultProps} count={1} completed onDecrement={onDecrement} />);

    fireEvent.click(screen.getByRole('button'), { ctrlKey: true });

    expect(onDecrement).toHaveBeenCalledWith('2025-01-15');
    expect(onDecrement).toHaveBeenCalledTimes(1);
  });

  it('does not call onDecrement when count is 0', () => {
    const onDecrement = vi.fn();
    const onLog = vi.fn();
    render(<DayCell {...defaultProps} count={0} onDecrement={onDecrement} onLog={onLog} />);

    fireEvent.click(screen.getByRole('button'), { metaKey: true });

    expect(onDecrement).not.toHaveBeenCalled();
    // Still calls onLog since no decrement action
    expect(onLog).toHaveBeenCalled();
  });

  it('has accessible aria-label', () => {
    render(<DayCell {...defaultProps} />);
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('aria-label', expect.stringContaining('Log for'));
  });

  it('shows count in aria-label when logged', () => {
    render(<DayCell {...defaultProps} count={2} completed />);
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('aria-label', expect.stringContaining('(2)'));
  });

  describe('compact mode', () => {
    it('renders smaller when compact is true', () => {
      render(<DayCell {...defaultProps} compact />);
      const button = screen.getByRole('button');
      expect(button).toHaveClass('w-5');
      expect(button).toHaveClass('h-5');
    });

    it('renders normal size when compact is false', () => {
      render(<DayCell {...defaultProps} compact={false} />);
      const button = screen.getByRole('button');
      expect(button).toHaveClass('w-6');
      expect(button).toHaveClass('h-6');
    });
  });

  describe('future dates', () => {
    const futureDay: CalendarDay = {
      date: '2025-01-20',
      dayOfWeek: 1,
      dayOfMonth: 20,
      isToday: false,
      isPadding: false,
      isFuture: true,
    };

    it('renders nothing for future dates', () => {
      const { container } = render(<DayCell {...defaultProps} day={futureDay} />);
      // Should render an empty placeholder, not a button
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
      // Container should still have a div for layout
      expect(container.firstChild).toBeInTheDocument();
    });
  });
});
