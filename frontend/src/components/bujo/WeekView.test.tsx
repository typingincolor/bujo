import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { WeekView } from './WeekView';
import * as api from '@/api/bujo';
import { Entry } from '@/types/bujo';

vi.mock('@/api/bujo');

describe('WeekView', () => {
  const mockEntries: Entry[] = [
    {
      id: 1,
      content: 'Monday event',
      type: 'event',
      priority: 'none',
      parentId: null,
      loggedDate: '2026-01-19',
      children: [],
    },
    {
      id: 2,
      content: 'Tuesday priority task',
      type: 'task',
      priority: 'high',
      parentId: null,
      loggedDate: '2026-01-20',
      children: [],
    },
    {
      id: 3,
      content: 'Saturday event',
      type: 'event',
      priority: 'none',
      parentId: null,
      loggedDate: '2026-01-24',
      children: [],
    },
  ];

  beforeEach(() => {
    vi.mocked(api.getEntriesForDateRange).mockResolvedValue(mockEntries);
  });

  it('renders 2x3 grid layout', async () => {
    render(<WeekView startDate={new Date('2026-01-19')} />);

    await waitFor(() => {
      expect(screen.getByText('Mon 1/19')).toBeInTheDocument();
      expect(screen.getByText('Tue 1/20')).toBeInTheDocument();
      expect(screen.getByText('Wed 1/21')).toBeInTheDocument();
      expect(screen.getByText('Thu 1/22')).toBeInTheDocument();
      expect(screen.getByText('Fri 1/23')).toBeInTheDocument();
      expect(screen.getByText(/24-25/)).toBeInTheDocument();
    });
  });

  it('fetches entries for the week', async () => {
    render(<WeekView startDate={new Date('2026-01-19')} />);

    await waitFor(() => {
      expect(api.getEntriesForDateRange).toHaveBeenCalledWith(
        '2026-01-19',
        '2026-01-25'
      );
    });
  });

  it('filters and distributes entries to correct days', async () => {
    render(<WeekView startDate={new Date('2026-01-19')} />);

    await waitFor(() => {
      expect(screen.getByText('Monday event')).toBeInTheDocument();
      expect(screen.getByText('Tuesday priority task')).toBeInTheDocument();
      expect(screen.getByText('Saturday event')).toBeInTheDocument();
    });
  });

  it('handles entry selection', async () => {
    const user = userEvent.setup();
    render(<WeekView startDate={new Date('2026-01-19')} />);

    await waitFor(() => {
      expect(screen.getByText('Monday event')).toBeInTheDocument();
    });

    const buttons = screen.getAllByRole('button');
    await user.click(buttons[0]);

    const container = screen.getByText('Monday event').closest('div');
    expect(container).toHaveClass('bg-primary/10');
  });

  it('shows context panel when entry selected', async () => {
    const user = userEvent.setup();
    render(<WeekView startDate={new Date('2026-01-19')} />);

    await waitFor(() => {
      expect(screen.getByText('Monday event')).toBeInTheDocument();
    });

    const buttons = screen.getAllByRole('button');
    await user.click(buttons[0]);

    await waitFor(() => {
      expect(screen.getByText('Context')).toBeInTheDocument();
    });
  });
});
