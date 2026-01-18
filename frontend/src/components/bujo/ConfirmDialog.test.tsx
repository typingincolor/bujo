import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ConfirmDialog } from './ConfirmDialog'

describe('ConfirmDialog', () => {
  it('renders title and message', () => {
    render(
      <ConfirmDialog
        isOpen={true}
        title="Delete Entry"
        message="Are you sure you want to delete this entry?"
        onConfirm={() => {}}
        onCancel={() => {}}
      />
    )

    expect(screen.getByText('Delete Entry')).toBeInTheDocument()
    expect(screen.getByText('Are you sure you want to delete this entry?')).toBeInTheDocument()
  })

  it('renders confirm and cancel buttons', () => {
    render(
      <ConfirmDialog
        isOpen={true}
        title="Delete Entry"
        message="Are you sure?"
        onConfirm={() => {}}
        onCancel={() => {}}
      />
    )

    expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('calls onConfirm when confirm button is clicked', () => {
    const onConfirm = vi.fn()
    render(
      <ConfirmDialog
        isOpen={true}
        title="Delete Entry"
        message="Are you sure?"
        onConfirm={onConfirm}
        onCancel={() => {}}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: /confirm/i }))
    expect(onConfirm).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel when cancel button is clicked', () => {
    const onCancel = vi.fn()
    render(
      <ConfirmDialog
        isOpen={true}
        title="Delete Entry"
        message="Are you sure?"
        onConfirm={() => {}}
        onCancel={onCancel}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('does not render when isOpen is false', () => {
    render(
      <ConfirmDialog
        isOpen={false}
        title="Delete Entry"
        message="Are you sure?"
        onConfirm={() => {}}
        onCancel={() => {}}
      />
    )

    expect(screen.queryByText('Delete Entry')).not.toBeInTheDocument()
  })

  it('renders custom confirm button text', () => {
    render(
      <ConfirmDialog
        isOpen={true}
        title="Delete Entry"
        message="Are you sure?"
        confirmText="Delete"
        onConfirm={() => {}}
        onCancel={() => {}}
      />
    )

    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
  })

  it('applies destructive variant styling', () => {
    render(
      <ConfirmDialog
        isOpen={true}
        title="Delete Entry"
        message="Are you sure?"
        variant="destructive"
        onConfirm={() => {}}
        onCancel={() => {}}
      />
    )

    const confirmButton = screen.getByRole('button', { name: /confirm/i })
    expect(confirmButton).toHaveClass('bg-destructive')
  })
})
