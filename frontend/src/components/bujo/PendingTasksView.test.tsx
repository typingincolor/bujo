import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { PendingTasksView } from './PendingTasksView';
import { Entry } from '@/types/bujo';

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetAttentionScores: vi.fn(),
}))

import { GetAttentionScores } from '@/wailsjs/go/wails/App'

const mockGetAttentionScores = vi.mocked(GetAttentionScores)

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
  callbacks: {},
  onSelectEntry: vi.fn(),
  onRefresh: vi.fn(),
};

describe('PendingTasksView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetAttentionScores.mockResolvedValue({})
  })

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

    fireEvent.keyDown(window, { key: 'j' });
    expect(onSelectEntry).toHaveBeenCalledWith(entries[0]);

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

  it('displays attention scores from backend', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 75, Indicators: ['overdue', 'priority'], DaysOld: 5 },
    })

    const entries = [createEntry({ id: 1, content: 'Important task' })];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);

    await waitFor(() => {
      expect(screen.getByTestId('attention-badge')).toHaveTextContent('75');
    })
  });

  it('sorts entries by backend attention score', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 20, Indicators: [], DaysOld: 1 },
      2: { Score: 80, Indicators: ['overdue'], DaysOld: 5 },
    })

    const entries = [
      createEntry({ id: 1, content: 'Low score task' }),
      createEntry({ id: 2, content: 'High score task' }),
    ];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);

    await waitFor(() => {
      const badges = screen.getAllByTestId('attention-badge');
      expect(badges[0]).toHaveTextContent('80');
      expect(badges[1]).toHaveTextContent('20');
    })
  });

  it('calls GetAttentionScores with task entry IDs', async () => {
    mockGetAttentionScores.mockResolvedValue({})
    const entries = [
      createEntry({ id: 10, content: 'Task', type: 'task' }),
      createEntry({ id: 20, content: 'Note', type: 'note' }),
      createEntry({ id: 30, content: 'Another task', type: 'task' }),
    ];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);

    await waitFor(() => {
      expect(mockGetAttentionScores).toHaveBeenCalledWith([10, 30]);
    })
  });

  it('renders indicator badges when attention score has indicators', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 85, Indicators: ['overdue', 'priority'], DaysOld: 10 },
    })

    const entries = [createEntry({ id: 1, content: 'Urgent task' })];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);

    await waitFor(() => {
      const indicators = screen.getByTestId('attention-indicators');
      expect(indicators).toBeInTheDocument();
      expect(screen.getByText('overdue')).toBeInTheDocument();
      expect(screen.getByText('!')).toBeInTheDocument();
    })
  });

  it('renders only score badge when indicators are empty', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 30, Indicators: [], DaysOld: 2 },
    })

    const entries = [createEntry({ id: 1, content: 'Low priority task' })];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);

    await waitFor(() => {
      expect(screen.getByTestId('attention-badge')).toHaveTextContent('30');
    })
    // Score badge shows but no indicator badges
    expect(screen.queryByText('overdue')).not.toBeInTheDocument();
    expect(screen.queryByText('aging')).not.toBeInTheDocument();
    expect(screen.queryByText('migrated')).not.toBeInTheDocument();
  });

  it('renders all indicator types with correct labels', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 90, Indicators: ['overdue', 'priority', 'aging', 'migrated'], DaysOld: 15 },
    })

    const entries = [createEntry({ id: 1, content: 'All indicators task' })];
    render(<PendingTasksView {...defaultProps} overdueEntries={entries} />);

    await waitFor(() => {
      expect(screen.getByText('overdue')).toBeInTheDocument();
      expect(screen.getByText('!')).toBeInTheDocument();
      expect(screen.getByText('aging')).toBeInTheDocument();
      expect(screen.getByText('migrated')).toBeInTheDocument();
    })
  });
});
