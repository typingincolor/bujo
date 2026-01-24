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
})
