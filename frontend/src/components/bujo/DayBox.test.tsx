import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { DayBox } from './DayBox';
import { Entry } from '@/types/bujo';

describe('DayBox', () => {
  const mockEntries: Entry[] = [
    { id: 1, content: 'Meeting', type: 'event', priority: 'none', parentId: null, loggedDate: '2026-01-20', children: [] },
    { id: 2, content: 'Task', type: 'task', priority: 'high', parentId: null, loggedDate: '2026-01-20', children: [] },
  ];

  it('renders day number and name', () => {
    render(<DayBox dayNumber={20} dayName="Mon" entries={[]} />);
    expect(screen.getByText('20')).toBeInTheDocument();
    expect(screen.getByText('Mon')).toBeInTheDocument();
  });

  it('shows "No events" when empty', () => {
    render(<DayBox dayNumber={20} dayName="Mon" entries={[]} />);
    expect(screen.getByText('No events')).toBeInTheDocument();
  });

  it('renders all entries', () => {
    render(<DayBox dayNumber={20} dayName="Mon" entries={mockEntries} />);
    expect(screen.getByText('Meeting')).toBeInTheDocument();
    expect(screen.getByText('Task')).toBeInTheDocument();
  });

  it('calls onSelectEntry when entry clicked', async () => {
    const user = userEvent.setup();
    const onSelectEntry = vi.fn();
    render(<DayBox dayNumber={20} dayName="Mon" entries={mockEntries} onSelectEntry={onSelectEntry} />);

    await user.click(screen.getByText('Meeting'));
    expect(onSelectEntry).toHaveBeenCalledWith(mockEntries[0]);
  });

  it('highlights selected entry', () => {
    const { container } = render(
      <DayBox dayNumber={20} dayName="Mon" entries={mockEntries} selectedEntry={mockEntries[0]} />
    );
    const selectedItem = container.querySelector('.bg-primary\\/10');
    expect(selectedItem).toBeInTheDocument();
  });
});
