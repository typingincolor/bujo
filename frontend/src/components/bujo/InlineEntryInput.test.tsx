import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { InlineEntryInput } from './InlineEntryInput'

describe('InlineEntryInput', () => {
  it('renders an input field', () => {
    render(<InlineEntryInput onSubmit={() => {}} onCancel={() => {}} />)
    expect(screen.getByRole('textbox')).toBeInTheDocument()
  })

  it('auto-focuses the input on mount', () => {
    render(<InlineEntryInput onSubmit={() => {}} onCancel={() => {}} />)
    expect(screen.getByRole('textbox')).toHaveFocus()
  })

  it('calls onSubmit with content when Enter is pressed', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<InlineEntryInput onSubmit={onSubmit} onCancel={() => {}} />)

    await user.type(screen.getByRole('textbox'), '. Buy groceries{enter}')

    expect(onSubmit).toHaveBeenCalledWith('. Buy groceries')
  })

  it('calls onCancel when Escape is pressed', async () => {
    const user = userEvent.setup()
    const onCancel = vi.fn()
    render(<InlineEntryInput onSubmit={() => {}} onCancel={onCancel} />)

    await user.type(screen.getByRole('textbox'), 'some text{escape}')

    expect(onCancel).toHaveBeenCalled()
  })

  it('does not submit empty content', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<InlineEntryInput onSubmit={onSubmit} onCancel={() => {}} />)

    await user.type(screen.getByRole('textbox'), '{enter}')

    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('does not submit whitespace-only content', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<InlineEntryInput onSubmit={onSubmit} onCancel={() => {}} />)

    await user.type(screen.getByRole('textbox'), '   {enter}')

    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('renders with indentation when depth is provided', () => {
    render(<InlineEntryInput onSubmit={() => {}} onCancel={() => {}} depth={2} />)
    const container = screen.getByRole('textbox').parentElement
    expect(container).toHaveStyle({ marginLeft: '32px' })
  })

  it('renders with placeholder text', () => {
    render(<InlineEntryInput onSubmit={() => {}} onCancel={() => {}} />)
    expect(screen.getByPlaceholderText(/type entry/i)).toBeInTheDocument()
  })

  it('clears input after submit', async () => {
    const user = userEvent.setup()
    render(<InlineEntryInput onSubmit={() => {}} onCancel={() => {}} />)

    await user.type(screen.getByRole('textbox'), '. Task{enter}')

    expect(screen.getByRole('textbox')).toHaveValue('')
  })

  it('calls onCancel when clicking outside', async () => {
    const user = userEvent.setup()
    const onCancel = vi.fn()
    render(
      <div>
        <InlineEntryInput onSubmit={() => {}} onCancel={onCancel} />
        <button>Outside</button>
      </div>
    )

    await user.click(screen.getByRole('button', { name: 'Outside' }))

    expect(onCancel).toHaveBeenCalled()
  })
})
