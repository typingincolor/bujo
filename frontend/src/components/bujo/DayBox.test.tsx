import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { DayBox } from './DayBox';
import { Entry } from '@/types/bujo';

describe('DayBox', () => {
  const mockEntry: Entry = {
    id: 1,
    content: 'Test entry',
    type: 'event',
    priority: 'none',
    parentId: null,
    loggedDate: '2026-01-20',
    children: [],
  };

  it('renders day label', () => {
    render(<DayBox date={new Date('2026-01-20')} entries={[]} />);
    expect(screen.getByText('Tue 1/20')).toBeInTheDocument();
  });

  it('renders entries using WeekEntry', () => {
    render(<DayBox date={new Date('2026-01-20')} entries={[mockEntry]} />);
    expect(screen.getByText('Test entry')).toBeInTheDocument();
  });

  it('passes selectedEntry to WeekEntry', () => {
    render(
      <DayBox
        date={new Date('2026-01-20')}
        entries={[mockEntry]}
        selectedEntry={mockEntry}
      />
    );
    const container = screen.getByText('Test entry').closest('div');
    expect(container).toHaveClass('bg-primary/10');
  });

  it('calls onSelectEntry when entry clicked', async () => {
    const user = userEvent.setup();
    const onSelectEntry = vi.fn();
    render(
      <DayBox
        date={new Date('2026-01-20')}
        entries={[mockEntry]}
        onSelectEntry={onSelectEntry}
      />
    );

    await user.click(screen.getByRole('button'));
    expect(onSelectEntry).toHaveBeenCalledWith(mockEntry);
  });
});
