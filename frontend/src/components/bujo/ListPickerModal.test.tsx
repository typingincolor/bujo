import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ListPickerModal } from './ListPickerModal';

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetLists: vi.fn(),
}));

import { GetLists } from '@/wailsjs/go/wails/App';

const mockLists = [
  { ID: 1, Name: 'Shopping', Items: [{ ID: 101 }, { ID: 102 }] },
  { ID: 2, Name: 'Work Tasks', Items: [{ ID: 201 }] },
  { ID: 3, Name: 'Empty List', Items: [] },
];

describe('ListPickerModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (GetLists as ReturnType<typeof vi.fn>).mockResolvedValue(mockLists);
  });

  describe('visibility', () => {
    it('renders nothing when isOpen is false', () => {
      render(
        <ListPickerModal
          isOpen={false}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      expect(screen.queryByText('Move to List')).not.toBeInTheDocument();
    });

    it('renders modal when isOpen is true', async () => {
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Move to List')).toBeInTheDocument();
      });
    });
  });

  describe('content display', () => {
    it('displays entry content in the modal', async () => {
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Buy groceries"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/Buy groceries/)).toBeInTheDocument();
      });
    });

    it('displays all available lists', async () => {
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Shopping')).toBeInTheDocument();
        expect(screen.getByText('Work Tasks')).toBeInTheDocument();
        expect(screen.getByText('Empty List')).toBeInTheDocument();
      });
    });

    it('displays item count for each list', async () => {
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('2 items')).toBeInTheDocument();
        expect(screen.getByText('1 items')).toBeInTheDocument();
        expect(screen.getByText('0 items')).toBeInTheDocument();
      });
    });
  });

  describe('selection', () => {
    it('selects first list by default', async () => {
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        const shoppingLabel = screen.getByText('Shopping').closest('label');
        expect(shoppingLabel).toHaveClass('border-primary');
      });
    });

    it('allows selecting a different list', async () => {
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Work Tasks')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Work Tasks'));

      const workTasksLabel = screen.getByText('Work Tasks').closest('label');
      expect(workTasksLabel).toHaveClass('border-primary');
    });
  });

  describe('callbacks', () => {
    it('calls onSelect with list ID when Move button clicked', async () => {
      const onSelect = vi.fn();
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={onSelect}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Shopping')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Move' }));

      expect(onSelect).toHaveBeenCalledWith(1);
    });

    it('calls onSelect with selected list ID', async () => {
      const onSelect = vi.fn();
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={onSelect}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Work Tasks')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Work Tasks'));
      fireEvent.click(screen.getByRole('button', { name: 'Move' }));

      expect(onSelect).toHaveBeenCalledWith(2);
    });

    it('calls onCancel when Cancel button clicked', async () => {
      const onCancel = vi.fn();
      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={onCancel}
        />
      );

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

      expect(onCancel).toHaveBeenCalled();
    });
  });

  describe('empty state', () => {
    it('shows message when no lists exist', async () => {
      (GetLists as ReturnType<typeof vi.fn>).mockResolvedValue([]);

      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/No lists found/)).toBeInTheDocument();
      });
    });

    it('shows Close button instead of Move when no lists exist', async () => {
      (GetLists as ReturnType<typeof vi.fn>).mockResolvedValue([]);

      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Move' })).not.toBeInTheDocument();
      });
    });
  });

  describe('loading state', () => {
    it('shows loading message while fetching lists', async () => {
      (GetLists as ReturnType<typeof vi.fn>).mockImplementation(
        () => new Promise(() => {})
      );

      render(
        <ListPickerModal
          isOpen={true}
          entryContent="Test entry"
          onSelect={vi.fn()}
          onCancel={vi.fn()}
        />
      );

      expect(screen.getByText(/Loading lists/)).toBeInTheDocument();
    });
  });
});
