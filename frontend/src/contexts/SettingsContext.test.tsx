import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
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
  it('should provide default settings when no localStorage data exists', () => {
    localStorage.clear()

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

    localStorage.clear()
  })
})
