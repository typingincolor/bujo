import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Header } from './Header'

describe('Header', () => {
  it('renders title', () => {
    render(<Header title="Today" />)
    expect(screen.getByText('Today')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<Header title="Today" />)
    expect(screen.getByPlaceholderText(/search entries/i)).toBeInTheDocument()
  })

  it('renders current date', () => {
    render(<Header title="Today" />)
    const formattedDate = new Date().toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
    expect(screen.getByText(formattedDate)).toBeInTheDocument()
  })

  it('renders capture button', () => {
    render(<Header title="Today" onCapture={() => {}} />)
    expect(screen.getByTitle('Capture entries')).toBeInTheDocument()
  })

  it('calls onCapture when capture button is clicked', () => {
    const onCapture = vi.fn()
    render(<Header title="Today" onCapture={onCapture} />)

    fireEvent.click(screen.getByTitle('Capture entries'))
    expect(onCapture).toHaveBeenCalledTimes(1)
  })

  it('does not render capture button when onCapture not provided', () => {
    render(<Header title="Today" />)
    expect(screen.queryByTitle('Capture entries')).not.toBeInTheDocument()
  })
})
