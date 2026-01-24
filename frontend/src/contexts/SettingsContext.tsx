import { createContext, useContext, useState, ReactNode } from 'react'
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

export function SettingsProvider({ children }: SettingsProviderProps) {
  const [settings, setSettings] = useState<Settings>(DEFAULT_SETTINGS)

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
