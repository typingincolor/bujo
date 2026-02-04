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

  it('does not render search input', () => {
    render(<Header title="Today" />)
    expect(screen.queryByPlaceholderText(/search entries/i)).not.toBeInTheDocument()
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

describe('Header - Styling', () => {
  it('renders with bg-card/50 background and border-b', () => {
    render(<Header title="Journal" />)
    const header = screen.getByRole('banner')
    expect(header).toHaveClass('bg-card/50', 'border-b')
  })

  it('renders title with font-display styling', () => {
    render(<Header title="Journal" />)
    const title = screen.getByRole('heading', { level: 2 })
    expect(title).toHaveClass('font-display', 'text-2xl', 'font-semibold')
  })
})

describe('Header - Actions Slot', () => {
  it('renders custom actions in the actions slot', () => {
    render(
      <Header title="Journal" actions={<button>Custom Action</button>} />
    )
    expect(screen.getByRole('button', { name: 'Custom Action' })).toBeInTheDocument()
  })

  it('renders multiple actions in the actions slot', () => {
    render(
      <Header
        title="Journal"
        actions={
          <>
            <button>Action 1</button>
            <button>Action 2</button>
          </>
        }
      />
    )
    expect(screen.getByRole('button', { name: 'Action 1' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Action 2' })).toBeInTheDocument()
  })
})

describe('Header - Show Context Pickers', () => {
  it('shows context pickers by default', () => {
    render(<Header title="Journal" />)
    expect(screen.getByTitle('Set mood')).toBeInTheDocument()
    expect(screen.getByTitle('Set weather')).toBeInTheDocument()
    expect(screen.getByTitle('Set location')).toBeInTheDocument()
  })

  it('hides context pickers when showContextPickers is false', () => {
    render(<Header title="Journal" showContextPickers={false} />)
    expect(screen.queryByTitle('Set mood')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Set weather')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Set location')).not.toBeInTheDocument()
  })

  it('shows context pickers when showContextPickers is true', () => {
    render(<Header title="Journal" showContextPickers={true} />)
    expect(screen.getByTitle('Set mood')).toBeInTheDocument()
    expect(screen.getByTitle('Set weather')).toBeInTheDocument()
    expect(screen.getByTitle('Set location')).toBeInTheDocument()
  })
})

describe('Header - currentDate prop', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('uses currentDate prop for SetMood API call instead of today', async () => {
    const user = userEvent.setup()
    const pastDate = new Date('2026-01-15')
    render(<Header title="Today" currentDate={pastDate} />)

    await user.click(screen.getByTitle('Set mood'))
    await user.click(screen.getByText('😊'))

    await waitFor(() => {
      expect(SetMood).toHaveBeenCalled()
      const call = vi.mocked(SetMood).mock.calls[0]
      const dateArg = new Date(call[0] as unknown as string)
      expect(dateArg.toDateString()).toBe(pastDate.toDateString())
    })
  })

  it('uses currentDate prop for SetWeather API call instead of today', async () => {
    const user = userEvent.setup()
    const pastDate = new Date('2026-01-15')
    render(<Header title="Today" currentDate={pastDate} />)

    await user.click(screen.getByTitle('Set weather'))
    await user.click(screen.getByText('☀️'))

    await waitFor(() => {
      expect(SetWeather).toHaveBeenCalled()
      const call = vi.mocked(SetWeather).mock.calls[0]
      const dateArg = new Date(call[0] as unknown as string)
      expect(dateArg.toDateString()).toBe(pastDate.toDateString())
    })
  })

  it('uses currentDate prop for SetLocation API call instead of today', async () => {
    const user = userEvent.setup()
    const pastDate = new Date('2026-01-15')
    render(<Header title="Today" currentDate={pastDate} />)

    await user.click(screen.getByTitle('Set location'))
    await user.click(screen.getByText('🏢'))

    await waitFor(() => {
      expect(SetLocation).toHaveBeenCalled()
      const call = vi.mocked(SetLocation).mock.calls[0]
      const dateArg = new Date(call[0] as unknown as string)
      expect(dateArg.toDateString()).toBe(pastDate.toDateString())
    })
  })

  it('displays the currentDate in header when provided', () => {
    const pastDate = new Date('2026-01-15')
    render(<Header title="Today" currentDate={pastDate} />)

    expect(screen.getByText('Thursday, January 15, 2026')).toBeInTheDocument()
  })

  it('passes date with timezone offset (not UTC) to SetMood', async () => {
    const user = userEvent.setup()
    const localDate = new Date(2026, 0, 15, 0, 0, 0)
    render(<Header title="Today" currentDate={localDate} />)

    await user.click(screen.getByTitle('Set mood'))
    await user.click(screen.getByText('😊'))

    await waitFor(() => {
      expect(SetMood).toHaveBeenCalled()
      const dateStr = vi.mocked(SetMood).mock.calls[0][0] as unknown as string
      // toISOString() produces UTC ending in 'Z' which shifts dates for non-UTC timezones
      // toWailsTime() preserves local date with timezone offset like +00:00 or -05:00
      expect(dateStr).not.toMatch(/Z$/)
      expect(dateStr).toMatch(/[+-]\d{2}:\d{2}$/)
    })
  })

  it('passes date with timezone offset (not UTC) to SetWeather', async () => {
    const user = userEvent.setup()
    const localDate = new Date(2026, 0, 15, 0, 0, 0)
    render(<Header title="Today" currentDate={localDate} />)

    await user.click(screen.getByTitle('Set weather'))
    await user.click(screen.getByText('☀️'))

    await waitFor(() => {
      expect(SetWeather).toHaveBeenCalled()
      const dateStr = vi.mocked(SetWeather).mock.calls[0][0] as unknown as string
      expect(dateStr).not.toMatch(/Z$/)
      expect(dateStr).toMatch(/[+-]\d{2}:\d{2}$/)
    })
  })

  it('passes date with timezone offset (not UTC) to SetLocation', async () => {
    const user = userEvent.setup()
    const localDate = new Date(2026, 0, 15, 0, 0, 0)
    render(<Header title="Today" currentDate={localDate} />)

    await user.click(screen.getByTitle('Set location'))
    await user.click(screen.getByText('🏢'))

    await waitFor(() => {
      expect(SetLocation).toHaveBeenCalled()
      const dateStr = vi.mocked(SetLocation).mock.calls[0][0] as unknown as string
      expect(dateStr).not.toMatch(/Z$/)
      expect(dateStr).toMatch(/[+-]\d{2}:\d{2}$/)
    })
  })
})
