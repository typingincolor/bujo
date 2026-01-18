import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MigrateModal } from './MigrateModal'

describe('MigrateModal', () => {
  it('renders when open', () => {
    render(
      <MigrateModal
        isOpen={true}
        entryContent="Test task"
        onMigrate={() => {}}
        onCancel={() => {}}
      />
    )
    expect(screen.getByText('Migrate Entry')).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    render(
      <MigrateModal
        isOpen={false}
        entryContent="Test task"
        onMigrate={() => {}}
        onCancel={() => {}}
      />
    )
    expect(screen.queryByText('Migrate Entry')).not.toBeInTheDocument()
  })

  it('displays entry content', () => {
    render(
      <MigrateModal
        isOpen={true}
        entryContent="Buy groceries"
        onMigrate={() => {}}
        onCancel={() => {}}
      />
    )
    expect(screen.getByText(/Buy groceries/)).toBeInTheDocument()
  })

  it('calls onCancel when cancel button is clicked', () => {
    const onCancel = vi.fn()
    render(
      <MigrateModal
        isOpen={true}
        entryContent="Test task"
        onMigrate={() => {}}
        onCancel={onCancel}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('calls onMigrate with selected date when migrate button is clicked', () => {
    const onMigrate = vi.fn()
    render(
      <MigrateModal
        isOpen={true}
        entryContent="Test task"
        onMigrate={onMigrate}
        onCancel={() => {}}
      />
    )

    const dateInput = document.querySelector('input[type="date"]') as HTMLInputElement
    fireEvent.change(dateInput, { target: { value: '2026-01-25' } })
    fireEvent.click(screen.getByRole('button', { name: /^migrate$/i }))

    expect(onMigrate).toHaveBeenCalledWith('2026-01-25')
  })

  it('defaults to tomorrow date', () => {
    render(
      <MigrateModal
        isOpen={true}
        entryContent="Test task"
        onMigrate={() => {}}
        onCancel={() => {}}
      />
    )

    const dateInput = document.querySelector('input[type="date"]') as HTMLInputElement
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    expect(dateInput.value).toBe(tomorrow.toISOString().split('T')[0])
  })
})
