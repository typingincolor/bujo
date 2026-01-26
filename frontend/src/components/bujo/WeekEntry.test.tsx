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

  it('shows action bar on hover', async () => {
    const callbacks = {
      onCancel: vi.fn(),
      onEdit: vi.fn(),
      onDelete: vi.fn(),
    };

    const { container } = render(
      <WeekEntry entry={mockEntry} callbacks={callbacks} />
    );

    const entryContainer = container.firstChild as HTMLElement;

    // Action bar should not be visible initially
    const actionBar = screen.queryByTestId('entry-action-bar');
    expect(actionBar).not.toBeInTheDocument();

    // Hover over the entry
    fireEvent.mouseEnter(entryContainer);

    // Action bar should now be visible
    expect(screen.getByTestId('entry-action-bar')).toBeInTheDocument();
  });

  it('calls callbacks when action buttons clicked', () => {
    const callbacks = {
      onCancel: vi.fn(),
      onEdit: vi.fn(),
      onDelete: vi.fn(),
    };

    const { container } = render(
      <WeekEntry entry={mockEntry} callbacks={callbacks} />
    );

    const entryContainer = container.firstChild as HTMLElement;

    // Hover to reveal action bar
    fireEvent.mouseEnter(entryContainer);

    // Find and click action buttons using fireEvent instead of userEvent
    const cancelButton = screen.getByTitle('Cancel entry');
    fireEvent.click(cancelButton);
    expect(callbacks.onCancel).toHaveBeenCalledTimes(1);

    const editButton = screen.getByTitle('Edit entry');
    fireEvent.click(editButton);
    expect(callbacks.onEdit).toHaveBeenCalledTimes(1);

    const deleteButton = screen.getByTitle('Delete entry');
    fireEvent.click(deleteButton);
    expect(callbacks.onDelete).toHaveBeenCalledTimes(1);
  });
});
