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

  it('renders date range with "Weekend" label when no locations', () => {
    render(<WeekendBox saturdayDay={24} sundayDay={25} saturdayEntries={[]} sundayEntries={[]} />);
    expect(screen.getByText(/24 - 25/)).toBeInTheDocument();
    expect(screen.getByText('Weekend')).toBeInTheDocument();
  });

  it('renders Saturday entries with "Sat:" prefix', () => {
    render(<WeekendBox saturdayDay={24} sundayDay={25} saturdayEntries={[satEntry]} sundayEntries={[]} />);
    expect(screen.getByText('Sat:')).toBeInTheDocument();
    expect(screen.getByText('Saturday event')).toBeInTheDocument();
  });

  it('renders Sunday entries with "Sun:" prefix', () => {
    render(<WeekendBox saturdayDay={24} sundayDay={25} saturdayEntries={[]} sundayEntries={[sunEntry]} />);
    expect(screen.getByText('Sun:')).toBeInTheDocument();
    expect(screen.getByText('Sunday event')).toBeInTheDocument();
  });

  it('renders both Saturday and Sunday entries together', () => {
    render(
      <WeekendBox saturdayDay={24} sundayDay={25} saturdayEntries={[satEntry]} sundayEntries={[sunEntry]} />
    );
    expect(screen.getByText('Saturday event')).toBeInTheDocument();
    expect(screen.getByText('Sunday event')).toBeInTheDocument();
  });

  it('passes selectedEntry to WeekEntry', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[satEntry]}
        sundayEntries={[]}
        selectedEntry={satEntry}
      />
    );
    const container = screen.getByText('Saturday event').closest('div');
    expect(container).toHaveClass('bg-primary/10');
  });

  it('calls onSelectEntry when entry clicked', async () => {
    const user = userEvent.setup();
    const onSelectEntry = vi.fn();
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[satEntry]}
        sundayEntries={[]}
        onSelectEntry={onSelectEntry}
      />
    );

    const buttons = screen.getAllByRole('button');
    await user.click(buttons[0]);
    expect(onSelectEntry).toHaveBeenCalledWith(satEntry);
  });

  it('shows "No events" when both days are empty', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[]}
        sundayEntries={[]}
      />
    );
    expect(screen.getByText('No events')).toBeInTheDocument();
  });

  it('does not show "No events" when Saturday has entries', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[satEntry]}
        sundayEntries={[]}
      />
    );
    expect(screen.queryByText('No events')).not.toBeInTheDocument();
  });

  it('does not show "No events" when Sunday has entries', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[]}
        sundayEntries={[sunEntry]}
      />
    );
    expect(screen.queryByText('No events')).not.toBeInTheDocument();
  });

  it('handles month boundary correctly (Jan 31 - Feb 1)', () => {
    render(<WeekendBox saturdayDay={31} sundayDay={1} saturdayEntries={[]} sundayEntries={[]} />);
    expect(screen.getByText(/31 - 1/)).toBeInTheDocument();
    expect(screen.getByText('Weekend')).toBeInTheDocument();
  });

  it('displays Saturday location with "not set" for Sunday', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[]}
        sundayEntries={[]}
        saturdayLocation="Office"
      />
    );
    expect(screen.getByText(/24 - 25/)).toBeInTheDocument();
    expect(screen.getByText('Weekend')).toBeInTheDocument();
    expect(screen.getByText(/Office \/ not set/)).toBeInTheDocument();
  });

  it('displays Sunday location with "not set" for Saturday', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[]}
        sundayEntries={[]}
        sundayLocation="Home"
      />
    );
    expect(screen.getByText(/24 - 25/)).toBeInTheDocument();
    expect(screen.getByText('Weekend')).toBeInTheDocument();
    expect(screen.getByText(/not set \/ Home/)).toBeInTheDocument();
  });

  it('displays both Saturday and Sunday locations when provided', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[]}
        sundayEntries={[]}
        saturdayLocation="Office"
        sundayLocation="Home"
      />
    );
    expect(screen.getByText(/24 - 25/)).toBeInTheDocument();
    expect(screen.getByText('Weekend')).toBeInTheDocument();
    expect(screen.getByText(/Office \/ Home/)).toBeInTheDocument();
  });

  it('displays header without location info when not provided', () => {
    render(
      <WeekendBox
        saturdayDay={24}
        sundayDay={25}
        saturdayEntries={[]}
        sundayEntries={[]}
      />
    );
    expect(screen.getByText(/24 - 25/)).toBeInTheDocument();
    expect(screen.getByText('Weekend')).toBeInTheDocument();
  });
});
