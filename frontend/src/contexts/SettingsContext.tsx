/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import type { Settings, Theme, DefaultView } from '../types/settings'
import { DEFAULT_SETTINGS } from '../types/settings'

interface SettingsContextValue extends Settings {
  setTheme: (theme: Theme) => void
  setDefaultView: (view: DefaultView) => void
}

const SettingsContext = createContext<SettingsContextValue | undefined>(undefined)

interface SettingsProviderProps {
  children: ReactNode
}

function loadSettings(): Settings {
  try {
    const stored = localStorage.getItem('bujo-settings')
    if (stored) {
      const parsed = JSON.parse(stored)
      return { ...DEFAULT_SETTINGS, ...parsed }
    }
  } catch (error) {
    console.warn('Failed to load settings from localStorage:', error)
  }
  return DEFAULT_SETTINGS
}

export function SettingsProvider({ children }: SettingsProviderProps) {
  const [settings, setSettings] = useState<Settings>(loadSettings)

  useEffect(() => {
    localStorage.setItem('bujo-settings', JSON.stringify(settings))
  }, [settings])

  useEffect(() => {
    document.documentElement.classList.toggle('dark', settings.theme === 'dark')
  }, [settings.theme])

  const setTheme = (theme: Theme) => {
    setSettings(prev => ({ ...prev, theme }))
  }

  const setDefaultView = (view: DefaultView) => {
    setSettings(prev => ({ ...prev, defaultView: view }))
  }

  const value: SettingsContextValue = {
    ...settings,
    setTheme,
    setDefaultView,
  }

  return <SettingsContext.Provider value={value}>{children}</SettingsContext.Provider>
}

export function useSettings(): SettingsContextValue {
  const context = useContext(SettingsContext)
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider')
  }
  return context
}
