import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { WeekView } from './WeekView';
import { DayEntries } from '@/types/bujo';

describe('WeekView', () => {
  const mockWeekData: DayEntries[] = [
    {
      date: '2026-01-19',
      entries: [
        { id: 1, content: 'Mon meeting', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-19', children: [] },
      ],
    },
    {
      date: '2026-01-20',
      entries: [
        { id: 2, content: 'Tue task', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-20', children: [] },
      ],
    },
    {
      date: '2026-01-21',
      entries: [],
    },
    {
      date: '2026-01-22',
      entries: [],
    },
    {
      date: '2026-01-23',
      entries: [
        { id: 3, content: 'Fri event', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-23', children: [] },
      ],
    },
    {
      date: '2026-01-24',
      entries: [
        { id: 4, content: 'Sat lunch', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-24', children: [] },
      ],
    },
    {
      date: '2026-01-25',
      entries: [
        { id: 5, content: 'Sun task', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-25', children: [] },
      ],
    },
  ];

  it('renders 5 day boxes plus weekend box', () => {
    const { container } = render(<WeekView days={mockWeekData} />);
    const boxes = container.querySelectorAll('.rounded-lg.border');
    expect(boxes).toHaveLength(6);
  });

  it('renders week date range header', () => {
    render(<WeekView days={mockWeekData} />);
    expect(screen.getByText(/Jan 19.*Jan 25, 2026/)).toBeInTheDocument();
  });

  it('filters to events and priority entries only', () => {
    const withNonPriority: DayEntries[] = [
      {
        date: '2026-01-19',
        entries: [
          { id: 1, content: 'Meeting', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-19', children: [] },
          { id: 2, content: 'Task no priority', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-19', children: [] },
          { id: 3, content: 'Task with priority', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-19', children: [] },
        ],
      },
      ...mockWeekData.slice(1),
    ];

    render(<WeekView days={withNonPriority} />);
    expect(screen.getByText('Meeting')).toBeInTheDocument();
    expect(screen.getByText('Task with priority')).toBeInTheDocument();
    expect(screen.queryByText('Task no priority')).not.toBeInTheDocument();
  });

  it('shows context panel when entry selected', async () => {
    const user = userEvent.setup();
    render(<WeekView days={mockWeekData} />);

    await user.click(screen.getByText('Mon meeting'));
    expect(screen.getByText('Context')).toBeInTheDocument();
  });

  it('shows "No entry selected" initially', () => {
    render(<WeekView days={mockWeekData} />);
    expect(screen.getByText('No entry selected')).toBeInTheDocument();
  });

  it('accepts callbacks prop without errors', () => {
    const callbacks = {
      onMarkDone: vi.fn(),
      onMigrate: vi.fn(),
      onEdit: vi.fn(),
      onDelete: vi.fn(),
      onCyclePriority: vi.fn(),
      onMoveToList: vi.fn(),
    };

    // This test verifies that WeekView accepts callbacks prop
    // and doesn't throw during render
    expect(() => {
      render(<WeekView days={mockWeekData} callbacks={callbacks} />);
    }).not.toThrow();
  });
});
