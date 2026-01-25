import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { renderHook, act } from '@testing-library/react'
import { SettingsProvider, useSettings } from './SettingsContext'
import { DEFAULT_SETTINGS } from '../types/settings'

function TestComponent() {
  const settings = useSettings()
  return (
    <div>
      <div data-testid="theme">{settings.theme}</div>
      <div data-testid="defaultView">{settings.defaultView}</div>
    </div>
  )
}

describe('SettingsContext', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('should provide default settings when no localStorage data exists', () => {

    render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    expect(screen.getByTestId('theme').textContent).toBe(DEFAULT_SETTINGS.theme)
    expect(screen.getByTestId('defaultView').textContent).toBe(DEFAULT_SETTINGS.defaultView)
  })

  it('should read existing settings from localStorage on initialization', () => {
    const savedSettings = { theme: 'light' as const, defaultView: 'week' as const }
    localStorage.setItem('bujo-settings', JSON.stringify(savedSettings))

    render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    expect(screen.getByTestId('theme').textContent).toBe('light')
    expect(screen.getByTestId('defaultView').textContent).toBe('week')
  })

  it('should persist theme changes to localStorage', () => {
    const { result } = renderHook(() => useSettings(), {
      wrapper: SettingsProvider,
    })

    act(() => {
      result.current.setTheme('light')
    })

    const stored = localStorage.getItem('bujo-settings')
    expect(stored).toBeDefined()
    const parsed = JSON.parse(stored!)
    expect(parsed.theme).toBe('light')
  })

  it('should persist default view changes to localStorage', () => {
    const { result } = renderHook(() => useSettings(), {
      wrapper: SettingsProvider,
    })

    act(() => {
      result.current.setDefaultView('week')
    })

    const stored = localStorage.getItem('bujo-settings')
    expect(stored).toBeDefined()
    const parsed = JSON.parse(stored!)
    expect(parsed.defaultView).toBe('week')
  })

  it('should fall back to defaults when localStorage contains invalid JSON', () => {
    const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
    localStorage.setItem('bujo-settings', 'invalid json{')

    render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    expect(screen.getByTestId('theme').textContent).toBe(DEFAULT_SETTINGS.theme)
    expect(screen.getByTestId('defaultView').textContent).toBe(DEFAULT_SETTINGS.defaultView)
    expect(consoleWarnSpy).toHaveBeenCalled()

    consoleWarnSpy.mockRestore()
  })

  it('should merge valid stored settings with defaults', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light' }))

    render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    expect(screen.getByTestId('theme').textContent).toBe('light')
    expect(screen.getByTestId('defaultView').textContent).toBe(DEFAULT_SETTINGS.defaultView)
  })

  it('should throw error when useSettings is used outside SettingsProvider', () => {
    expect(() => {
      renderHook(() => useSettings())
    }).toThrow('useSettings must be used within a SettingsProvider')
  })

  it('should apply dark class to html element when theme is dark', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'dark', defaultView: 'today' }))
    document.documentElement.classList.remove('dark')

    render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('should remove dark class from html element when theme is light', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'today' }))
    document.documentElement.classList.add('dark')

    render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('should apply dark class when theme is system and prefers-color-scheme is dark', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'system', defaultView: 'today' }))
    document.documentElement.classList.remove('dark')

    // Mock matchMedia to return dark preference
    const mockMatchMedia = vi.fn().mockImplementation((query) => ({
      matches: query === '(prefers-color-scheme: dark)',
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
        <TestComponent />
      </SettingsProvider>
    )

    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('should remove dark class when theme is system and prefers-color-scheme is light', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'system', defaultView: 'today' }))
    document.documentElement.classList.add('dark')

    // Mock matchMedia to return light preference
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
        <TestComponent />
      </SettingsProvider>
    )

    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('should remove matchMedia listener when theme changes from system to another value', () => {
    const removeEventListener = vi.fn()
    const addEventListener = vi.fn()
    const mockMatchMedia = vi.fn().mockImplementation(() => ({
      matches: true,
      addEventListener,
      removeEventListener,
    }))
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: mockMatchMedia,
    })

    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'system', defaultView: 'today' }))

    const { result } = renderHook(() => useSettings(), {
      wrapper: SettingsProvider,
    })

    // Verify listener was added for system theme
    expect(addEventListener).toHaveBeenCalledWith('change', expect.any(Function))

    // Change theme from system to dark
    act(() => {
      result.current.setTheme('dark')
    })

    // Verify listener was removed
    expect(removeEventListener).toHaveBeenCalledWith('change', expect.any(Function))
  })

  it('should remove matchMedia listener on unmount when using system theme', () => {
    const removeEventListener = vi.fn()
    const addEventListener = vi.fn()
    const mockMatchMedia = vi.fn().mockImplementation(() => ({
      matches: true,
      addEventListener,
      removeEventListener,
    }))
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: mockMatchMedia,
    })

    localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'system', defaultView: 'today' }))

    const { unmount } = render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    // Verify listener was added
    expect(addEventListener).toHaveBeenCalledWith('change', expect.any(Function))

    // Unmount the component
    unmount()

    // Verify listener was removed
    expect(removeEventListener).toHaveBeenCalledWith('change', expect.any(Function))
  })
})
