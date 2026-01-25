export type Theme = 'light' | 'dark' | 'system'
export type DefaultView = 'today' | 'week' | 'overview' | 'search'

export interface Settings {
  theme: Theme
  defaultView: DefaultView
}

export const DEFAULT_SETTINGS: Settings = {
  theme: 'dark',
  defaultView: 'today',
}
