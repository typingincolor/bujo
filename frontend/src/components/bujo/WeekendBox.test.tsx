import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { WeekendBox } from './WeekendBox';
import { Entry } from '@/types/bujo';

describe('WeekendBox', () => {
  const satEntry: Entry = {
    id: 1,
    content: 'Saturday event',
    type: 'event',
    priority: 'none',
    parentId: null,
    loggedDate: '2026-01-24',
    children: [],
  };

  const sunEntry: Entry = {
    id: 2,
    content: 'Sunday event',
    type: 'event',
    priority: 'none',
    parentId: null,
    loggedDate: '2026-01-25',
    children: [],
  };

  it('renders "Weekend" label', () => {
    render(<WeekendBox saturdayEntries={[]} sundayEntries={[]} />);
    expect(screen.getByText('Weekend')).toBeInTheDocument();
  });

  it('renders Saturday entries with "Sat:" prefix', () => {
    render(<WeekendBox saturdayEntries={[satEntry]} sundayEntries={[]} />);
    expect(screen.getByText('Sat:')).toBeInTheDocument();
    expect(screen.getByText('Saturday event')).toBeInTheDocument();
  });

  it('renders Sunday entries with "Sun:" prefix', () => {
    render(<WeekendBox saturdayEntries={[]} sundayEntries={[sunEntry]} />);
    expect(screen.getByText('Sun:')).toBeInTheDocument();
    expect(screen.getByText('Sunday event')).toBeInTheDocument();
  });

  it('renders both Saturday and Sunday entries together', () => {
    render(
      <WeekendBox saturdayEntries={[satEntry]} sundayEntries={[sunEntry]} />
    );
    expect(screen.getByText('Saturday event')).toBeInTheDocument();
    expect(screen.getByText('Sunday event')).toBeInTheDocument();
  });

  it('passes selectedEntryId to WeekEntry', () => {
    render(
      <WeekendBox
        saturdayEntries={[satEntry]}
        sundayEntries={[]}
        selectedEntryId={1}
      />
    );
    const container = screen.getByText('Saturday event').closest('div');
    expect(container).toHaveClass('bg-primary/10');
  });

  it('calls onEntrySelect when entry clicked', async () => {
    const user = userEvent.setup();
    const onEntrySelect = vi.fn();
    render(
      <WeekendBox
        saturdayEntries={[satEntry]}
        sundayEntries={[]}
        onEntrySelect={onEntrySelect}
      />
    );

    const buttons = screen.getAllByRole('button');
    await user.click(buttons[0]);
    expect(onEntrySelect).toHaveBeenCalledWith(1);
  });
});
