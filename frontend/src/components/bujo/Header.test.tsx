import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Header } from './Header'

vi.mock('@/wailsjs/go/wails/App', () => ({
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue(['Home', 'Office', 'Coffee Shop']),
}))

import { SetMood, SetWeather, SetLocation, GetLocationHistory } from '@/wailsjs/go/wails/App'

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

  it('shows mood picker with all mood options when clicked', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set mood'))

    expect(screen.getByText('😊')).toBeInTheDocument()
    expect(screen.getByText('😐')).toBeInTheDocument()
    expect(screen.getByText('😢')).toBeInTheDocument()
    expect(screen.getByText('😤')).toBeInTheDocument()
    expect(screen.getByText('😴')).toBeInTheDocument()
    expect(screen.getByText('🤒')).toBeInTheDocument()
    expect(screen.getByText('😰')).toBeInTheDocument()
    expect(screen.getByText('🤗')).toBeInTheDocument()
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

  it('shows weather picker with all weather options when clicked', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set weather'))

    expect(screen.getByText('☀️')).toBeInTheDocument()
    expect(screen.getByText('🌤️')).toBeInTheDocument()
    expect(screen.getByText('☁️')).toBeInTheDocument()
    expect(screen.getByText('🌧️')).toBeInTheDocument()
    expect(screen.getByText('⛈️')).toBeInTheDocument()
    expect(screen.getByText('❄️')).toBeInTheDocument()
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

describe('Header - Location', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders location button', () => {
    render(<Header title="Today" />)
    expect(screen.getByTitle('Set location')).toBeInTheDocument()
  })

  it('shows location input when location button is clicked', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set location'))

    expect(screen.getByPlaceholderText('Enter location...')).toBeInTheDocument()
  })

  it('loads location history for suggestions', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set location'))

    await waitFor(() => {
      expect(GetLocationHistory).toHaveBeenCalled()
    })
  })

  it('shows location suggestions from history', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set location'))

    await waitFor(() => {
      expect(screen.getByText('Home')).toBeInTheDocument()
      expect(screen.getByText('Office')).toBeInTheDocument()
      expect(screen.getByText('Coffee Shop')).toBeInTheDocument()
    })
  })

  it('calls SetLocation binding when selecting a suggestion', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set location'))

    await waitFor(() => {
      expect(screen.getByText('Office')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Office'))

    await waitFor(() => {
      expect(SetLocation).toHaveBeenCalled()
      const call = vi.mocked(SetLocation).mock.calls[0]
      expect(call[1]).toBe('Office')
    })
  })

  it('calls SetLocation binding when typing and pressing enter', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set location'))

    const input = screen.getByPlaceholderText('Enter location...')
    await user.type(input, 'New Location{Enter}')

    await waitFor(() => {
      expect(SetLocation).toHaveBeenCalled()
      const call = vi.mocked(SetLocation).mock.calls[0]
      expect(call[1]).toBe('New Location')
    })
  })

  it('displays current location when set', () => {
    render(<Header title="Today" currentLocation="Home Office" />)
    // The location should be shown in the button text
    expect(screen.getByText('Home Office')).toBeInTheDocument()
  })

  it('shows quick location options with emojis when location button is clicked', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set location'))

    // Should show predefined location options with emojis
    expect(screen.getByText('🏠')).toBeInTheDocument()
    expect(screen.getByText('🏢')).toBeInTheDocument()
    expect(screen.getByText('☕')).toBeInTheDocument()
    expect(screen.getByText('📚')).toBeInTheDocument()
    expect(screen.getByText('✈️')).toBeInTheDocument()
  })

  it('displays location emoji when predefined location is set', () => {
    render(<Header title="Today" currentLocation="home" />)
    // Home emoji should be displayed
    expect(screen.getByText('🏠')).toBeInTheDocument()
  })

  it('calls SetLocation with predefined value when clicking quick option', async () => {
    const user = userEvent.setup()
    render(<Header title="Today" />)

    await user.click(screen.getByTitle('Set location'))
    await user.click(screen.getByText('🏢'))

    await waitFor(() => {
      expect(SetLocation).toHaveBeenCalled()
      const call = vi.mocked(SetLocation).mock.calls[0]
      expect(call[1]).toBe('office')
    })
  })
})
