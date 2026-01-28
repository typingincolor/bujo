import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { DeletionReviewDialog } from './DeletionReviewDialog'
import type { DeletedEntry } from '@/hooks/useEditableDocument'

describe('DeletionReviewDialog', () => {
  const mockOnConfirm = vi.fn()
  const mockOnCancel = vi.fn()
  const mockOnRestore = vi.fn()

  const defaultProps = {
    isOpen: true,
    deletedEntries: [
      { entityId: 'entity-1', content: '. Task to delete' },
      { entityId: 'entity-2', content: '- Note to delete' },
    ] as DeletedEntry[],
    onConfirm: mockOnConfirm,
    onCancel: mockOnCancel,
    onRestore: mockOnRestore,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('visibility', () => {
    it('renders when isOpen is true', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })

    it('does not render when isOpen is false', () => {
      render(<DeletionReviewDialog {...defaultProps} isOpen={false} />)

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  describe('content', () => {
    it('shows dialog title with date when provided', () => {
      render(<DeletionReviewDialog {...defaultProps} date="Monday, Jan 27" />)

      expect(screen.getByText(/save changes to monday, jan 27/i)).toBeInTheDocument()
    })

    it('shows fallback title when no date provided', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      expect(screen.getByText(/confirm deletions/i)).toBeInTheDocument()
    })

    it('shows warning message about permanent deletions', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      expect(screen.getByText(/will be permanently deleted/i)).toBeInTheDocument()
    })

    it('lists all deleted entries', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      expect(screen.getByText('. Task to delete')).toBeInTheDocument()
      expect(screen.getByText('- Note to delete')).toBeInTheDocument()
    })

    it('shows deletion count', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      expect(screen.getByText(/2 items/i)).toBeInTheDocument()
    })
  })

  describe('actions', () => {
    it('calls onConfirm when save button is clicked', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      fireEvent.click(screen.getByRole('button', { name: /save.*delete/i }))

      expect(mockOnConfirm).toHaveBeenCalled()
    })

    it('calls onCancel when cancel button is clicked', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      fireEvent.click(screen.getByRole('button', { name: /cancel/i }))

      expect(mockOnCancel).toHaveBeenCalled()
    })

    it('calls onRestore with entityId when restore button is clicked', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      const restoreButtons = screen.getAllByRole('button', { name: /restore/i })
      fireEvent.click(restoreButtons[0])

      expect(mockOnRestore).toHaveBeenCalledWith('entity-1')
    })

    it('calls onDiscardAll when Discard All Changes button is clicked', () => {
      const mockOnDiscardAll = vi.fn()
      render(<DeletionReviewDialog {...defaultProps} onDiscardAll={mockOnDiscardAll} />)

      fireEvent.click(screen.getByRole('button', { name: /discard all changes/i }))

      expect(mockOnDiscardAll).toHaveBeenCalled()
    })
  })

  describe('empty state', () => {
    it('shows message when no deletions', () => {
      render(<DeletionReviewDialog {...defaultProps} deletedEntries={[]} />)

      expect(screen.getByText(/no deletions/i)).toBeInTheDocument()
    })
  })

  describe('accessibility', () => {
    it('has accessible dialog role', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      expect(screen.getByRole('dialog')).toHaveAttribute('aria-modal', 'true')
    })

    it('has accessible title', () => {
      render(<DeletionReviewDialog {...defaultProps} />)

      const dialog = screen.getByRole('dialog')
      expect(dialog).toHaveAttribute('aria-labelledby')
    })
  })
})
