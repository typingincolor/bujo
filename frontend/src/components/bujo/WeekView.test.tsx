import React from 'react';
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

  it('shows all parent entries and excludes children', () => {
    const withChildren: DayEntries[] = [
      {
        date: '2026-01-19',
        entries: [
          { id: 1, content: 'Meeting', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-19', children: [] },
          { id: 2, content: 'Task no priority', type: 'task', priority: 'none', parentId: null, loggedDate: '2026-01-19', children: [] },
          { id: 3, content: 'Child task', type: 'task', priority: 'high', parentId: 1, loggedDate: '2026-01-19', children: [] },
        ],
      },
      ...mockWeekData.slice(1),
    ];

    render(<WeekView days={withChildren} />);
    expect(screen.getByText('Meeting')).toBeInTheDocument();
    expect(screen.getByText('Task no priority')).toBeInTheDocument();
    expect(screen.queryByText('Child task')).not.toBeInTheDocument();
  });

  it('shows context panel when entry selected', async () => {
    const user = userEvent.setup();
    render(<WeekView days={mockWeekData} isContextCollapsed={false} />);

    await user.click(screen.getByText('Mon meeting'));
    expect(screen.getByText('Context')).toBeInTheDocument();
  });

  it('shows context tree for root-level entries', async () => {
    const withRootEntry: DayEntries[] = [
      {
        date: '2026-01-19',
        entries: [
          { id: 1, content: 'Root task', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-19', children: [] },
          { id: 2, content: 'Another root', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-19', children: [] },
        ],
      },
      ...mockWeekData.slice(1),
    ];

    const user = userEvent.setup();
    render(<WeekView days={withRootEntry} isContextCollapsed={false} />);

    // Click root-level entry
    await user.click(screen.getByText('Root task'));

    // Should show context tree, not "No context"
    expect(screen.queryByText('No context')).not.toBeInTheDocument();
    // Both entries should be visible in context tree
    expect(screen.getAllByText('Root task').length).toBeGreaterThan(1); // Once in day view, once in context
    expect(screen.getAllByText('Another root').length).toBeGreaterThan(0); // Appears in context tree
  });

  it('shows "No entry selected" initially', () => {
    render(<WeekView days={mockWeekData} isContextCollapsed={false} />);
    expect(screen.getByText('No entry selected')).toBeInTheDocument();
  });

  it('accepts callbacks prop without errors', () => {
    const callbacks = {
      onNavigateToEntry: vi.fn(),
    };

    // This test verifies that WeekView accepts callbacks prop
    // and doesn't throw during render
    expect(() => {
      render(<WeekView days={mockWeekData} callbacks={callbacks} />);
    }).not.toThrow();
  });

  it('displays correct day numbers matching the dates', () => {
    const { container } = render(<WeekView days={mockWeekData} />);
    const dayBoxes = container.querySelectorAll('.rounded-lg.border');

    // Mon Jan 19 should show "19"
    expect(dayBoxes[0]).toHaveTextContent('19');

    // Tue Jan 20 should show "20"
    expect(dayBoxes[1]).toHaveTextContent('20');

    // Wed Jan 21 should show "21"
    expect(dayBoxes[2]).toHaveTextContent('21');

    // Thu Jan 22 should show "22"
    expect(dayBoxes[3]).toHaveTextContent('22');

    // Fri Jan 23 should show "23"
    expect(dayBoxes[4]).toHaveTextContent('23');
  });

  it('handles backend UTC dates correctly without timezone shift', () => {
    // Backend returns dates as UTC strings like "2026-01-20T00:00:00Z"
    // Frontend transforms to just date part "2026-01-20"
    // This test verifies parseISO interprets correctly as local date
    const backendFormat: DayEntries[] = [
      {
        date: '2026-01-20', // Transformed from backend's "2026-01-20T00:00:00Z"
        entries: [],
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
        entries: [],
      },
      {
        date: '2026-01-24',
        entries: [],
      },
      {
        date: '2026-01-25',
        entries: [],
      },
      {
        date: '2026-01-26',
        entries: [],
      },
    ];

    const { container } = render(<WeekView days={backendFormat} />);
    const dayBoxes = container.querySelectorAll('.rounded-lg.border');

    // Day numbers should match the dates exactly (20, 21, 22, 23, 24)
    expect(dayBoxes[0]).toHaveTextContent('20');
    expect(dayBoxes[1]).toHaveTextContent('21');
    expect(dayBoxes[2]).toHaveTextContent('22');
    expect(dayBoxes[3]).toHaveTextContent('23');
    expect(dayBoxes[4]).toHaveTextContent('24');
  });

  it('displays context tree from prop when entry selected', async () => {
    const contextEntries = [
      { id: 10, content: 'Parent task', type: 'task' as const, priority: 'high' as const, parentId: null, loggedDate: '2026-01-19', children: [] },
      { id: 11, content: 'Child task', type: 'task' as const, priority: 'high' as const, parentId: 10, loggedDate: '2026-01-19', children: [] },
    ];

    const user = userEvent.setup();
    render(
      <WeekView
        days={mockWeekData}
        contextTree={contextEntries}
        isContextCollapsed={false}
      />
    );

    await user.click(screen.getByText('Mon meeting'));

    expect(screen.getByText('Parent task')).toBeInTheDocument();
    expect(screen.getByText('Child task')).toBeInTheDocument();
  });

  it('shows "No context" when contextTree prop is empty', async () => {
    const user = userEvent.setup();
    render(
      <WeekView
        days={mockWeekData}
        contextTree={[]}
        isContextCollapsed={false}
      />
    );

    await user.click(screen.getByText('Mon meeting'));

    expect(screen.getByText('No context')).toBeInTheDocument();
  });

  describe('Collapsible Context Panel', () => {
    it('context panel is collapsed by default', () => {
      render(<WeekView days={mockWeekData} />);

      // When collapsed, context content should not be visible
      expect(screen.queryByText('Context')).not.toBeInTheDocument();
    });

    it('shows collapse toggle button', () => {
      render(<WeekView days={mockWeekData} />);

      const toggleButton = screen.getByLabelText('Toggle context panel');
      expect(toggleButton).toBeInTheDocument();
    });

    it('expands context panel when toggle clicked', async () => {
      const user = userEvent.setup();

      // Create a wrapper component that manages state
      function WrapperComponent() {
        const [isCollapsed, setIsCollapsed] = React.useState(true);
        return (
          <WeekView
            days={mockWeekData}
            isContextCollapsed={isCollapsed}
            onToggleContextCollapse={() => setIsCollapsed(!isCollapsed)}
          />
        );
      }

      render(<WrapperComponent />);

      // Initially collapsed - Context should not be visible
      expect(screen.queryByText('Context')).not.toBeInTheDocument();

      const toggleButton = screen.getByLabelText('Toggle context panel');
      await user.click(toggleButton);

      // After expanding, context section should be visible
      expect(screen.getByText('Context')).toBeInTheDocument();
    });

    it('calls onToggleContextCollapse callback when toggle clicked', async () => {
      const user = userEvent.setup();
      const onToggleContextCollapse = vi.fn();

      render(
        <WeekView
          days={mockWeekData}
          onToggleContextCollapse={onToggleContextCollapse}
        />
      );

      const toggleButton = screen.getByLabelText('Toggle context panel');
      await user.click(toggleButton);

      expect(onToggleContextCollapse).toHaveBeenCalledOnce();
    });

    it('shows ChevronLeft icon when collapsed', () => {
      render(
        <WeekView
          days={mockWeekData}
          isContextCollapsed={true}
        />
      );

      const toggleButton = screen.getByLabelText('Toggle context panel');
      // ChevronLeft points left, indicating expand action
      expect(toggleButton.querySelector('svg')).toBeInTheDocument();
    });

    it('shows ChevronRight icon when expanded', () => {
      render(
        <WeekView
          days={mockWeekData}
          isContextCollapsed={false}
        />
      );

      const toggleButton = screen.getByLabelText('Toggle context panel');
      // ChevronRight points right, indicating collapse action
      expect(toggleButton.querySelector('svg')).toBeInTheDocument();
    });
  });

  describe('Habit Display', () => {
    it('displays habits logged on each day', () => {
      const habitsForWeek = [
        {
          id: 1,
          name: 'Exercise',
          streak: 5,
          completionRate: 80,
          dayHistory: [
            { date: '2026-01-19', completed: true, count: 1 },
            { date: '2026-01-20', completed: true, count: 2 },
          ],
          todayLogged: false,
          todayCount: 0,
        },
        {
          id: 2,
          name: 'Meditation',
          streak: 3,
          completionRate: 70,
          dayHistory: [
            { date: '2026-01-19', completed: true, count: 1 },
          ],
          todayLogged: false,
          todayCount: 0,
        },
      ];

      render(<WeekView days={mockWeekData} habits={habitsForWeek} />);

      // Monday (Jan 19) should show both habits
      expect(screen.getByText('Exercise')).toBeInTheDocument();
      expect(screen.getByText('Meditation')).toBeInTheDocument();

      // Tuesday (Jan 20) should show Exercise with count
      expect(screen.getByText('Exercise (2)')).toBeInTheDocument();
    });

    it('does not display habits with zero count', () => {
      const habitsForWeek = [
        {
          id: 1,
          name: 'Exercise',
          streak: 5,
          completionRate: 80,
          dayHistory: [
            { date: '2026-01-19', completed: false, count: 0 },
          ],
          todayLogged: false,
          todayCount: 0,
        },
      ];

      render(<WeekView days={mockWeekData} habits={habitsForWeek} />);

      expect(screen.queryByText('Exercise')).not.toBeInTheDocument();
    });

    it('displays habits in weekend box for Saturday and Sunday', () => {
      const habitsForWeek = [
        {
          id: 1,
          name: 'Reading',
          streak: 10,
          completionRate: 90,
          dayHistory: [
            { date: '2026-01-24', completed: true, count: 1 },
            { date: '2026-01-25', completed: true, count: 3 },
          ],
          todayLogged: false,
          todayCount: 0,
        },
      ];

      render(<WeekView days={mockWeekData} habits={habitsForWeek} />);

      expect(screen.getByText('Reading')).toBeInTheDocument();
      expect(screen.getByText('Reading (3)')).toBeInTheDocument();
    });
  });
});
