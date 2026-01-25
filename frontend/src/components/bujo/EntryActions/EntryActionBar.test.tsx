import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EntryActionBar } from './EntryActionBar';
import { Entry } from '@/types/bujo';

const createEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  content: 'Test entry',
  type: 'task',
  priority: 'none',
  parentId: null,
  loggedDate: '2024-01-01',
  ...overrides,
});

describe('EntryActionBar', () => {
  describe('action visibility', () => {
    it('shows cancel button for task entry when onCancel provided', () => {
      const onCancel = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry({ type: 'task' })}
          callbacks={{ onCancel }}
        />
      );

      expect(screen.getByTitle('Cancel entry')).toBeInTheDocument();
    });

    it('shows uncancel button for cancelled entry when onUncancel provided', () => {
      const onUncancel = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry({ type: 'cancelled' })}
          callbacks={{ onUncancel }}
        />
      );

      expect(screen.getByTitle('Uncancel entry')).toBeInTheDocument();
    });

    it('does not show cancel for cancelled entry', () => {
      const onCancel = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry({ type: 'cancelled' })}
          callbacks={{ onCancel }}
        />
      );

      expect(screen.queryByTitle('Cancel entry')).not.toBeInTheDocument();
    });

    it('shows migrate button only for task entries', () => {
      const onMigrate = vi.fn();

      const { rerender } = render(
        <EntryActionBar
          entry={createEntry({ type: 'task' })}
          callbacks={{ onMigrate }}
        />
      );
      expect(screen.getByTitle('Migrate entry')).toBeInTheDocument();

      rerender(
        <EntryActionBar
          entry={createEntry({ type: 'note' })}
          callbacks={{ onMigrate }}
        />
      );
      expect(screen.queryByTitle('Migrate entry')).not.toBeInTheDocument();
    });

    it('shows answer button only for question entries', () => {
      const onAnswer = vi.fn();

      const { rerender } = render(
        <EntryActionBar
          entry={createEntry({ type: 'question' })}
          callbacks={{ onAnswer }}
        />
      );
      expect(screen.getByTitle('Answer question')).toBeInTheDocument();

      rerender(
        <EntryActionBar
          entry={createEntry({ type: 'task' })}
          callbacks={{ onAnswer }}
        />
      );
      expect(screen.queryByTitle('Answer question')).not.toBeInTheDocument();
    });

    it('does not show edit button for cancelled entries', () => {
      const onEdit = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry({ type: 'cancelled' })}
          callbacks={{ onEdit }}
        />
      );

      expect(screen.queryByTitle('Edit entry')).not.toBeInTheDocument();
    });

    it('does not show cycleType for cancelled entries', () => {
      const onCycleType = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry({ type: 'cancelled' })}
          callbacks={{ onCycleType }}
        />
      );

      expect(screen.queryByTitle('Change type')).not.toBeInTheDocument();
    });

    it('shows cycleType for task entries', () => {
      const onCycleType = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry({ type: 'task' })}
          callbacks={{ onCycleType }}
        />
      );

      expect(screen.getByTitle('Change type')).toBeInTheDocument();
    });
  });

  describe('callback invocation', () => {
    it('calls onCancel when cancel button clicked', () => {
      const onCancel = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry({ type: 'task' })}
          callbacks={{ onCancel }}
        />
      );

      fireEvent.click(screen.getByTitle('Cancel entry'));
      expect(onCancel).toHaveBeenCalledTimes(1);
    });

    it('calls onDelete when delete button clicked', () => {
      const onDelete = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry()}
          callbacks={{ onDelete }}
        />
      );

      fireEvent.click(screen.getByTitle('Delete entry'));
      expect(onDelete).toHaveBeenCalledTimes(1);
    });

    it('calls onEdit when edit button clicked', () => {
      const onEdit = vi.fn();
      render(
        <EntryActionBar
          entry={createEntry()}
          callbacks={{ onEdit }}
        />
      );

      fireEvent.click(screen.getByTitle('Edit entry'));
      expect(onEdit).toHaveBeenCalledTimes(1);
    });
  });

  describe('placeholders', () => {
    it('renders placeholders to maintain alignment when usePlaceholders is true', () => {
      render(
        <EntryActionBar
          entry={createEntry({ type: 'note' })}
          callbacks={{ onCancel: vi.fn(), onDelete: vi.fn() }}
          usePlaceholders
        />
      );

      const actionSlots = document.querySelectorAll('[data-action-slot]');
      expect(actionSlots.length).toBeGreaterThan(2);
    });
  });

  describe('variants', () => {
    it('applies opacity-0 when not hovered/selected for hover-reveal variant', () => {
      render(
        <EntryActionBar
          entry={createEntry()}
          callbacks={{ onDelete: vi.fn() }}
          variant="hover-reveal"
          isHovered={false}
          isSelected={false}
        />
      );

      const container = document.querySelector('[data-testid="entry-action-bar"]');
      expect(container).toHaveClass('opacity-0');
      expect(container).toHaveClass('focus-within:opacity-100');
    });

    it('applies opacity-100 when hovered for hover-reveal variant', () => {
      render(
        <EntryActionBar
          entry={createEntry()}
          callbacks={{ onDelete: vi.fn() }}
          variant="hover-reveal"
          isHovered={true}
          isSelected={false}
        />
      );

      const container = document.querySelector('[data-testid="entry-action-bar"]');
      expect(container).toHaveClass('opacity-100');
      expect(container).toHaveClass('focus-within:opacity-100');
    });

    it('applies opacity-100 when selected for hover-reveal variant', () => {
      render(
        <EntryActionBar
          entry={createEntry()}
          callbacks={{ onDelete: vi.fn() }}
          variant="hover-reveal"
          isHovered={false}
          isSelected={true}
        />
      );

      const container = document.querySelector('[data-testid="entry-action-bar"]');
      expect(container).toHaveClass('opacity-100');
      expect(container).toHaveClass('focus-within:opacity-100');
    });

    it('does not apply opacity classes for always-visible variant', () => {
      render(
        <EntryActionBar
          entry={createEntry()}
          callbacks={{ onDelete: vi.fn() }}
          variant="always-visible"
        />
      );

      const container = document.querySelector('[data-testid="entry-action-bar"]');
      expect(container).not.toHaveClass('opacity-0');
    });
  });
});
