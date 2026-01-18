import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Header } from './Header'

vi.mock('@/wailsjs/go/wails/App', () => ({
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
}))

import { SetMood, SetWeather } from '@/wailsjs/go/wails/App'

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

describe('Header - Day Context', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders mood button', () => {
    render(<Header title="Today" />)
    expect(screen.getByTitle('Set mood')).toBeInTheDocument()
  })

  it('shows mood picker when mood button is clicked', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set mood'))

    expect(screen.getByText('😊')).toBeInTheDocument()
    expect(screen.getByText('😐')).toBeInTheDocument()
    expect(screen.getByText('😢')).toBeInTheDocument()
  })

  it('calls SetMood binding when selecting mood', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set mood'))
    await user.click(screen.getByText('😊'))

    await waitFor(() => {
      expect(SetMood).toHaveBeenCalled()
      const call = vi.mocked(SetMood).mock.calls[0]
      expect(call[1]).toBe('happy')
    })
  })

  it('renders weather button', () => {
    render(<Header title="Today" />)
    expect(screen.getByTitle('Set weather')).toBeInTheDocument()
  })

  it('shows weather picker when weather button is clicked', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set weather'))

    expect(screen.getByText('☀️')).toBeInTheDocument()
    expect(screen.getByText('☁️')).toBeInTheDocument()
    expect(screen.getByText('🌧️')).toBeInTheDocument()
  })

  it('calls SetWeather binding when selecting weather', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set weather'))
    await user.click(screen.getByText('☀️'))

    await waitFor(() => {
      expect(SetWeather).toHaveBeenCalled()
      const call = vi.mocked(SetWeather).mock.calls[0]
      expect(call[1]).toBe('sunny')
    })
  })

  it('displays current mood when set', () => {
    render(<Header title="Today" currentMood="happy" />)
    expect(screen.getByText('😊')).toBeInTheDocument()
  })

  it('displays current weather when set', () => {
    render(<Header title="Today" currentWeather="sunny" />)
    expect(screen.getByText('☀️')).toBeInTheDocument()
  })
})
