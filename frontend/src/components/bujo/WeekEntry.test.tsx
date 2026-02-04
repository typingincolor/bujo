import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { WeekEntry } from './WeekEntry';
import { Entry } from '@/types/bujo';

describe('WeekEntry', () => {
  const mockEntry: Entry = {
    id: 1,
    content: 'Test entry',
    type: 'task',
    priority: 'high',
    parentId: null,
    loggedDate: '2026-01-20',
    children: [],
  };

  it('renders entry with symbol and content', () => {
    render(<WeekEntry entry={mockEntry} />);
    expect(screen.getByText('•')).toBeInTheDocument();
    expect(screen.getByText('Test entry')).toBeInTheDocument();
  });

  it('shows priority indicator for high priority', () => {
    render(<WeekEntry entry={mockEntry} />);
    expect(screen.getByText('!!!')).toBeInTheDocument();
  });

  it('calls onSelect when clicked', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<WeekEntry entry={mockEntry} onSelect={onSelect} />);

    await user.click(screen.getByRole('button'));
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it('shows selected state', () => {
    const { container } = render(<WeekEntry entry={mockEntry} isSelected={true} />);
    expect(container.firstChild).toHaveClass('bg-primary/10');
  });

  it('shows date prefix when provided', () => {
    render(<WeekEntry entry={mockEntry} datePrefix="Sat:" />);
    expect(screen.getByText('Sat:')).toBeInTheDocument();
  });

  it('calls onNavigateToEntry on double click', () => {
    const onNavigateToEntry = vi.fn();
    render(<WeekEntry entry={mockEntry} onNavigateToEntry={onNavigateToEntry} />);

    fireEvent.doubleClick(screen.getByRole('button'));
    expect(onNavigateToEntry).toHaveBeenCalledWith(mockEntry);
  });
});
