import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EditEntryModal } from './EditEntryModal'

describe('EditEntryModal', () => {
  it('renders with initial content', () => {
    render(
      <EditEntryModal
        isOpen={true}
        initialContent="Original content"
        onSave={() => {}}
        onCancel={() => {}}
      />
    )

    const input = screen.getByDisplayValue('Original content')
    expect(input).toBeInTheDocument()
  })

  it('renders save and cancel buttons', () => {
    render(
      <EditEntryModal
        isOpen={true}
        initialContent="Test"
        onSave={() => {}}
        onCancel={() => {}}
      />
    )

    expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('calls onSave with new content when save is clicked', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    render(
      <EditEntryModal
        isOpen={true}
        initialContent="Original"
        onSave={onSave}
        onCancel={() => {}}
      />
    )

    const input = screen.getByDisplayValue('Original')
    await user.clear(input)
    await user.type(input, 'Updated content')

    fireEvent.click(screen.getByRole('button', { name: /save/i }))
    expect(onSave).toHaveBeenCalledWith('Updated content')
  })

  it('calls onCancel when cancel button is clicked', () => {
    const onCancel = vi.fn()
    render(
      <EditEntryModal
        isOpen={true}
        initialContent="Test"
        onSave={() => {}}
        onCancel={onCancel}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('does not render when isOpen is false', () => {
    render(
      <EditEntryModal
        isOpen={false}
        initialContent="Test"
        onSave={() => {}}
        onCancel={() => {}}
      />
    )

    expect(screen.queryByDisplayValue('Test')).not.toBeInTheDocument()
  })

  it('calls onSave when Enter is pressed', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    render(
      <EditEntryModal
        isOpen={true}
        initialContent="Test"
        onSave={onSave}
        onCancel={() => {}}
      />
    )

    const input = screen.getByDisplayValue('Test')
    await user.type(input, '{Enter}')
    expect(onSave).toHaveBeenCalledWith('Test')
  })

  it('calls onCancel when Escape is pressed', async () => {
    const user = userEvent.setup()
    const onCancel = vi.fn()
    render(
      <EditEntryModal
        isOpen={true}
        initialContent="Test"
        onSave={() => {}}
        onCancel={onCancel}
      />
    )

    const input = screen.getByDisplayValue('Test')
    await user.type(input, '{Escape}')
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('disables save button when content is empty', async () => {
    const user = userEvent.setup()
    render(
      <EditEntryModal
        isOpen={true}
        initialContent="Test"
        onSave={() => {}}
        onCancel={() => {}}
      />
    )

    const input = screen.getByDisplayValue('Test')
    await user.clear(input)

    expect(screen.getByRole('button', { name: /save/i })).toBeDisabled()
  })
})
