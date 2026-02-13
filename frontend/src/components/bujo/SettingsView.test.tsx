import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SettingsView } from './SettingsView'
import { SettingsProvider } from '../../contexts/SettingsContext'
import * as WailsApp from '@/wailsjs/go/wails/App'
import * as WailsRuntime from '@/wailsjs/runtime/runtime'

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetVersion: vi.fn(() => Promise.resolve('1.0.0')),
}))

vi.mock('@/wailsjs/runtime/runtime', () => ({
  BrowserOpenURL: vi.fn(),
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

  it('setting rows wrap content to prevent overlap', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const themeRow = screen.getByText('Theme').closest('div[class*="flex"]')
    expect(themeRow?.className).toContain('flex-wrap')
    expect(themeRow?.className).toContain('gap-')
  })

  it('setting rows have minimum width for labels', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const themeLabel = screen.getByText('Theme').parentElement
    expect(themeLabel?.className).toContain('min-w-')
  })

  it('opens GitHub repo in system browser when GitHub link is clicked', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const githubLink = screen.getByRole('button', { name: 'GitHub' })
    await user.click(githubLink)

    expect(WailsRuntime.BrowserOpenURL).toHaveBeenCalledWith(
      'https://github.com/typingincolor/bujo'
    )
  })

  it('opens GitHub issues page in system browser when Support link is clicked', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const supportLink = screen.getByRole('button', { name: 'GitHub Issues' })
    await user.click(supportLink)

    expect(WailsRuntime.BrowserOpenURL).toHaveBeenCalledWith(
      'https://github.com/typingincolor/bujo/issues'
    )
  })

  it('displays integrations section', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText('Integrations')).toBeInTheDocument()
  })

  it('displays Gmail bookmarklet link in integrations section', () => {
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )
    expect(screen.getByText('Gmail Bookmarklet')).toBeInTheDocument()
    expect(screen.getByText('Capture emails as tasks directly from Gmail')).toBeInTheDocument()
  })

  it('opens bookmarklet install page when Install link is clicked', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const installLink = screen.getByRole('button', { name: 'Install' })
    await user.click(installLink)

    expect(WailsRuntime.BrowserOpenURL).toHaveBeenCalledWith(
      'http://127.0.0.1:8743/install'
    )
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
