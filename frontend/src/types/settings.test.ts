import { describe, it, expect } from 'vitest'
import { DEFAULT_SETTINGS } from './settings'
import type { Settings, Theme, DefaultView } from './settings'

describe('Settings types', () => {
  it('should have default settings with correct types', () => {
    expect(DEFAULT_SETTINGS).toBeDefined()
    expect(DEFAULT_SETTINGS.theme).toBe('dark')
    expect(DEFAULT_SETTINGS.defaultView).toBe('today')
  })

  it('should allow valid theme values', () => {
    const validThemes: Theme[] = ['light', 'dark', 'system']
    validThemes.forEach(theme => {
      const settings: Settings = { ...DEFAULT_SETTINGS, theme }
      expect(settings.theme).toBe(theme)
    })
  })

  it('should allow valid default view values', () => {
    const validViews: DefaultView[] = ['today', 'week', 'overview', 'search']
    validViews.forEach(view => {
      const settings: Settings = { ...DEFAULT_SETTINGS, defaultView: view }
      expect(settings.defaultView).toBe(view)
    })
  })
})
