import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { ContextPill } from './ContextPill'

describe('ContextPill', () => {
  it('renders with ancestor count', () => {
    render(<ContextPill count={2} onClick={() => {}} />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toBeInTheDocument()
    expect(pill).toHaveTextContent('2')
  })

  it('shows singular parent in title for count of 1', () => {
    render(<ContextPill count={1} onClick={() => {}} />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toHaveAttribute('title', 'Show 1 parent')
  })

  it('shows plural parents in title for count > 1', () => {
    render(<ContextPill count={3} onClick={() => {}} />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toHaveAttribute('title', 'Show 3 parents')
  })

  it('has accessible aria-label', () => {
    render(<ContextPill count={2} onClick={() => {}} />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toHaveAttribute('aria-label', 'Show 2 parents')
  })

  it('calls onClick when clicked', async () => {
    const handleClick = vi.fn()
    const user = userEvent.setup()

    render(<ContextPill count={1} onClick={handleClick} />)

    await user.click(screen.getByTestId('context-pill'))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('stops event propagation when clicked', async () => {
    const handleClick = vi.fn()
    const handleParentClick = vi.fn()
    const user = userEvent.setup()

    render(
      <div onClick={handleParentClick}>
        <ContextPill count={1} onClick={handleClick} />
      </div>
    )

    await user.click(screen.getByTestId('context-pill'))
    expect(handleClick).toHaveBeenCalledTimes(1)
    expect(handleParentClick).not.toHaveBeenCalled()
  })

  it('shows loading state when isLoading is true', () => {
    render(<ContextPill count={0} onClick={() => {}} isLoading />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toHaveTextContent('⋯')
    expect(pill).toHaveAttribute('title', 'Loading parent context')
  })

  it('shows count when isLoading is false', () => {
    render(<ContextPill count={2} onClick={() => {}} isLoading={false} />)

    const pill = screen.getByTestId('context-pill')
    expect(pill).toHaveTextContent('2')
  })
})
