import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MigrateBatchModal } from './MigrateBatchModal'

describe('MigrateBatchModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('visibility', () => {
    it('renders nothing when isOpen is false', () => {
      render(
        <MigrateBatchModal
          isOpen={false}
          entries={['Test entry']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.queryByText('Migrate Entries')).not.toBeInTheDocument()
    })

    it('renders modal when isOpen is true', () => {
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Test entry']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.getByText('Migrate Entries')).toBeInTheDocument()
    })
  })

  describe('content display', () => {
    it('displays entry text', () => {
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Buy groceries']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.getByText('Buy groceries')).toBeInTheDocument()
    })

    it('displays multiple entries', () => {
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Task one', 'Task two', 'Task three']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.getByText('Task one')).toBeInTheDocument()
      expect(screen.getByText('Task two')).toBeInTheDocument()
      expect(screen.getByText('Task three')).toBeInTheDocument()
    })

    it('shows singular message for one entry', () => {
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Single entry']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.getByText(/Migrate this entry/)).toBeInTheDocument()
    })

    it('shows plural message for multiple entries', () => {
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['One', 'Two']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.getByText(/Migrate 2 entries/)).toBeInTheDocument()
    })
  })

  describe('callbacks', () => {
    it('calls onMigrate with date on submit', () => {
      const onMigrate = vi.fn()
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Test entry']}
          onMigrate={onMigrate}
          onCancel={vi.fn()}
        />
      )

      fireEvent.click(screen.getByRole('button', { name: 'Migrate' }))

      expect(onMigrate).toHaveBeenCalledTimes(1)
      expect(onMigrate).toHaveBeenCalledWith(expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/))
    })

    it('calls onCancel when Cancel button clicked', () => {
      const onCancel = vi.fn()
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Test entry']}
          onMigrate={vi.fn()}
          onCancel={onCancel}
        />
      )

      fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))

      expect(onCancel).toHaveBeenCalledTimes(1)
    })
  })

  describe('date input', () => {
    it('has a date input with tomorrow as default', () => {
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Test entry']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const tomorrow = new Date()
      tomorrow.setDate(tomorrow.getDate() + 1)
      const expected = tomorrow.toISOString().split('T')[0]

      const dateInput = screen.getByDisplayValue(expected)
      expect(dateInput).toBeInTheDocument()
    })

    it('has min attribute set to today', () => {
      render(
        <MigrateBatchModal
          isOpen={true}
          entries={['Test entry']}
          onMigrate={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const today = new Date().toISOString().split('T')[0]
      const dateInput = document.querySelector('input[type="date"]') as HTMLInputElement
      expect(dateInput.min).toBe(today)
    })
  })
})
