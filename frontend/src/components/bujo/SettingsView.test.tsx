import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SettingsView } from './SettingsView'
import { SettingsProvider } from '../../contexts/SettingsContext'
import * as WailsApp from '@/wailsjs/go/wails/App'

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetVersion: vi.fn(() => Promise.resolve('1.0.0')),
}))

describe('SettingsView', () => {
  beforeEach(() => {
    vi.mocked(WailsApp.GetVersion).mockResolvedValue('1.0.0')
  })
  it('renders settings title', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText(/settings/i)).toBeInTheDocument()
  })

  it('displays appearance section', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText(/appearance/i)).toBeInTheDocument()
  })

  it('displays theme setting', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText('Theme')).toBeInTheDocument()
  })

  it('displays data section', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText('Data')).toBeInTheDocument()
  })

  it('displays database path info', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText(/database/i)).toBeInTheDocument()
  })

  it('displays about section', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText(/about/i)).toBeInTheDocument()
  })

  it('displays version info', async () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText('Version')).toBeInTheDocument()
    expect(await screen.findByText('1.0.0')).toBeInTheDocument()
  })

  it('displays default view setting', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText(/default view/i)).toBeInTheDocument()
  })

  it('displays current theme from settings context', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'today' }))

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    expect(screen.getByText('Light')).toBeInTheDocument()
  })

  it('allows changing theme to dark', async () => {
    const user = userEvent.setup()
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'today' }))

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const darkOption = screen.getByText('Dark')
    await user.click(darkOption)

    const stored = localStorage.getItem('bujo-settings')
    const parsed = JSON.parse(stored!)
    expect(parsed.theme).toBe('dark')
  })

  it('allows changing theme to system', async () => {
    const user = userEvent.setup()
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'today' }))

    const mockMatchMedia = vi.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }))
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: mockMatchMedia,
    })

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const systemOption = screen.getByText('System')
    await user.click(systemOption)

    const stored = localStorage.getItem('bujo-settings')
    const parsed = JSON.parse(stored!)
    expect(parsed.theme).toBe('system')
  })

  it('displays current default view from settings context', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'week' }))

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    expect(screen.getByText('Week')).toBeInTheDocument()
  })

  it('allows changing default view to overview', async () => {
    const user = userEvent.setup()
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'today' }))

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const overviewOption = screen.getByText('Overview')
    await user.click(overviewOption)

    const stored = localStorage.getItem('bujo-settings')
    const parsed = JSON.parse(stored!)
    expect(parsed.defaultView).toBe('overview')
  })

  it('allows changing default view to search', async () => {
    const user = userEvent.setup()
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'today' }))

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const searchOption = screen.getByText('Search')
    await user.click(searchOption)

    const stored = localStorage.getItem('bujo-settings')
    const parsed = JSON.parse(stored!)
    expect(parsed.defaultView).toBe('search')
  })

  it('displays backend version from API', async () => {
    vi.mocked(WailsApp.GetVersion).mockResolvedValue('v0.1.0-nightly+85d8787')

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('v0.1.0-nightly+85d8787')).toBeInTheDocument()
    })
  })
})
