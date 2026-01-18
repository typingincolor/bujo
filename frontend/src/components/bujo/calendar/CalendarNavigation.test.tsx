import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CalendarNavigation } from './CalendarNavigation';

describe('CalendarNavigation', () => {
  const defaultProps = {
    label: 'January 2025',
    onPrev: vi.fn(),
    onNext: vi.fn(),
  };

  it('renders the label', () => {
    render(<CalendarNavigation {...defaultProps} />);
    expect(screen.getByText('January 2025')).toBeInTheDocument();
  });

  it('renders prev button with arrow icon', () => {
    render(<CalendarNavigation {...defaultProps} />);
    const prevButton = screen.getByLabelText('Previous');
    expect(prevButton).toBeInTheDocument();
  });

  it('renders next button with arrow icon', () => {
    render(<CalendarNavigation {...defaultProps} />);
    const nextButton = screen.getByLabelText('Next');
    expect(nextButton).toBeInTheDocument();
  });

  it('calls onPrev when prev button clicked', () => {
    const onPrev = vi.fn();
    render(<CalendarNavigation {...defaultProps} onPrev={onPrev} />);

    fireEvent.click(screen.getByLabelText('Previous'));

    expect(onPrev).toHaveBeenCalledTimes(1);
  });

  it('calls onNext when next button clicked', () => {
    const onNext = vi.fn();
    render(<CalendarNavigation {...defaultProps} onNext={onNext} />);

    fireEvent.click(screen.getByLabelText('Next'));

    expect(onNext).toHaveBeenCalledTimes(1);
  });

  it('disables prev button when canGoPrev is false', () => {
    render(<CalendarNavigation {...defaultProps} canGoPrev={false} />);

    const prevButton = screen.getByLabelText('Previous');
    expect(prevButton).toBeDisabled();
  });

  it('disables next button when canGoNext is false', () => {
    render(<CalendarNavigation {...defaultProps} canGoNext={false} />);

    const nextButton = screen.getByLabelText('Next');
    expect(nextButton).toBeDisabled();
  });

  it('enables both buttons by default', () => {
    render(<CalendarNavigation {...defaultProps} />);

    expect(screen.getByLabelText('Previous')).toBeEnabled();
    expect(screen.getByLabelText('Next')).toBeEnabled();
  });

  it('renders week range label correctly', () => {
    render(<CalendarNavigation {...defaultProps} label="Jan 12 - Jan 18, 2025" />);
    expect(screen.getByText('Jan 12 - Jan 18, 2025')).toBeInTheDocument();
  });

  it('renders quarter label correctly', () => {
    render(<CalendarNavigation {...defaultProps} label="Jan - Mar 2025" />);
    expect(screen.getByText('Jan - Mar 2025')).toBeInTheDocument();
  });
});
