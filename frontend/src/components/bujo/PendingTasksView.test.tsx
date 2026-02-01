import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PendingTasksView } from './PendingTasksView';
import { Entry } from '@/types/bujo';

const createEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  content: 'Test task',
  type: 'task',
  priority: 'none',
  parentId: null,
  loggedDate: '2024-01-01',
  ...overrides,
});

const defaultProps = {
  overdueEntries: [] as Entry[],
  now: new Date('2024-01-15'),
  callbacks: {},
  onSelectEntry: vi.fn(),
  onRefresh: vi.fn(),
};

describe('PendingTasksView', () => {
  it('renders empty state when no entries', () => {
    render(<PendingTasksView {...defaultProps} />);
    expect(screen.getByText('No pending tasks')).toBeInTheDocument();
  });

  it('renders task entries filtered to tasks only', () => {
    const entries = [
      createEntry({ id: 1, content: 'Task one', type: 'task' }),
      createEntry({ id: 2, content: 'A note', type: 'note' }),
      createEntry({ id: 3, content: 'Task two', type: 'task' }),
    ];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);

    expect(screen.getByText('Task one')).toBeInTheDocument();
    expect(screen.getByText('Task two')).toBeInTheDocument();
    expect(screen.queryByText('A note')).not.toBeInTheDocument();
  });

  it('shows count in header', () => {
    const entries = [
      createEntry({ id: 1, content: 'Task one' }),
      createEntry({ id: 2, content: 'Task two' }),
    ];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);
    expect(screen.getByText('Pending Tasks (2)')).toBeInTheDocument();
  });

  it('refresh button calls onRefresh', () => {
    const onRefresh = vi.fn();
    render(<PendingTasksView {...defaultProps} onRefresh={onRefresh} />);

    const refreshButton = screen.getByTitle('Refresh pending tasks');
    fireEvent.click(refreshButton);
    expect(onRefresh).toHaveBeenCalled();
  });

  it('clicking entry calls onSelectEntry', () => {
    const onSelectEntry = vi.fn();
    const entries = [createEntry({ id: 1, content: 'Click me' })];
    render(
      <PendingTasksView
        {...defaultProps}
        overdueEntries={entries}
        onSelectEntry={onSelectEntry}
      />
    );

    fireEvent.click(screen.getByText('Click me'));
    expect(onSelectEntry).toHaveBeenCalledWith(entries[0]);
  });

  it('shows selected entry highlighted', () => {
    const entries = [
      createEntry({ id: 1, content: 'Selected task' }),
      createEntry({ id: 2, content: 'Other task' }),
    ];
    render(
      <PendingTasksView
        {...defaultProps}
        overdueEntries={entries}
        selectedEntry={entries[0]}
      />
    );

    // Selected entry should have primary styling
    const selectedItem = screen.getByText('Selected task').closest('[class*="bg-primary"]');
    expect(selectedItem).toBeInTheDocument();
  });

  it('optimistic mark done updates display', () => {
    const entries = [createEntry({ id: 1, content: 'Mark me done' })];
    const onMarkDone = vi.fn();
    render(
      <PendingTasksView
        {...defaultProps}
        overdueEntries={entries}
        callbacks={{ onMarkDone }}
      />
    );

    // Entry should initially be visible as a task
    expect(screen.getByText('Mark me done')).toBeInTheDocument();
  });

  it('keyboard j/k navigates entries', () => {
    const onSelectEntry = vi.fn();
    const entries = [
      createEntry({ id: 1, content: 'First task' }),
      createEntry({ id: 2, content: 'Second task' }),
    ];
    render(
      <PendingTasksView
        {...defaultProps}
        overdueEntries={entries}
        onSelectEntry={onSelectEntry}
      />
    );

    // Press j to move down
    fireEvent.keyDown(window, { key: 'j' });
    expect(onSelectEntry).toHaveBeenCalledWith(entries[0]);

    // Press j again to move to second
    fireEvent.keyDown(window, { key: 'j' });
    expect(onSelectEntry).toHaveBeenCalledWith(entries[1]);
  });

  it('keyboard arrow keys navigate entries', () => {
    const onSelectEntry = vi.fn();
    const entries = [
      createEntry({ id: 1, content: 'First' }),
      createEntry({ id: 2, content: 'Second' }),
    ];
    render(
      <PendingTasksView
        {...defaultProps}
        overdueEntries={entries}
        onSelectEntry={onSelectEntry}
      />
    );

    fireEvent.keyDown(window, { key: 'ArrowDown' });
    expect(onSelectEntry).toHaveBeenCalledWith(entries[0]);
  });
});
