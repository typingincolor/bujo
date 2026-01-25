# Settings Dynamic Configuration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform Settings page from hardcoded values to fully functional configuration with theme switcher, default view selector, and backend version display.

**Architecture:** React Context pattern for centralized settings state, localStorage persistence, Tailwind dark mode integration, and inline script for flash prevention.

**Tech Stack:** React 19, TypeScript, Vitest, React Testing Library, Tailwind CSS, Wails (backend binding)

---

## Task 1: Settings Type Definitions

**Files:**
- Create: `frontend/src/types/settings.ts`

**Step 1: Write the failing test**

Create: `frontend/src/types/settings.test.ts`

```typescript
import { describe, it, expect } from 'vitest'
import type { Settings, Theme, DefaultView } from './settings'

describe('Settings types', () => {
  it('should accept valid theme values', () => {
    const validThemes: Theme[] = ['light', 'dark', 'system']
    validThemes.forEach(theme => {
      const settings: Settings = { theme, defaultView: 'today' }
      expect(settings.theme).toBe(theme)
    })
  })

  it('should accept valid defaultView values', () => {
    const validViews: DefaultView[] = ['today', 'week', 'overview', 'search']
    validViews.forEach(view => {
      const settings: Settings = { theme: 'dark', defaultView: view }
      expect(settings.defaultView).toBe(view)
    })
  })
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/types/settings.test.ts
```

Expected: FAIL with "Cannot find module './settings'"

**Step 3: Write minimal implementation**

Create: `frontend/src/types/settings.ts`

```typescript
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
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/types/settings.test.ts
```

Expected: PASS (2 tests)

**Step 5: Commit**

```bash
git add frontend/src/types/settings.ts frontend/src/types/settings.test.ts
git commit -m "feat: add Settings type definitions

Define Settings interface with Theme and DefaultView types.
Includes default values matching current hardcoded behavior.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Settings Context - Provider Initialization

**Files:**
- Create: `frontend/src/contexts/SettingsContext.tsx`
- Create: `frontend/src/contexts/SettingsContext.test.tsx`

**Step 1: Write the failing test**

Create: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SettingsProvider, useSettings } from './SettingsContext'

function TestComponent() {
  const { theme, defaultView } = useSettings()
  return (
    <div>
      <span data-testid="theme">{theme}</span>
      <span data-testid="defaultView">{defaultView}</span>
    </div>
  )
}

describe('SettingsProvider', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('should provide default settings when localStorage is empty', () => {
    render(
      <SettingsProvider>
        <TestComponent />
      </SettingsProvider>
    )

    expect(screen.getByTestId('theme')).toHaveTextContent('dark')
    expect(screen.getByTestId('defaultView')).toHaveTextContent('today')
  })
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: FAIL with "Cannot find module './SettingsContext'"

**Step 3: Write minimal implementation**

Create: `frontend/src/contexts/SettingsContext.tsx`

```typescript
import { createContext, useContext, ReactNode } from 'react'
import { Settings, DEFAULT_SETTINGS } from '@/types/settings'

interface SettingsContextValue extends Settings {
  setTheme: (theme: Settings['theme']) => void
  setDefaultView: (view: Settings['defaultView']) => void
}

const SettingsContext = createContext<SettingsContextValue | undefined>(undefined)

interface SettingsProviderProps {
  children: ReactNode
}

export function SettingsProvider({ children }: SettingsProviderProps) {
  const setTheme = () => {}
  const setDefaultView = () => {}

  const value: SettingsContextValue = {
    ...DEFAULT_SETTINGS,
    setTheme,
    setDefaultView,
  }

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  )
}

export function useSettings() {
  const context = useContext(SettingsContext)
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider')
  }
  return context
}
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (1 test)

**Step 5: Commit**

```bash
git add frontend/src/contexts/SettingsContext.tsx frontend/src/contexts/SettingsContext.test.tsx
git commit -m "feat: add SettingsContext with default values

Create SettingsProvider and useSettings hook.
Provider initializes with default settings.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Settings Context - Read from localStorage

**Files:**
- Modify: `frontend/src/contexts/SettingsContext.test.tsx`
- Modify: `frontend/src/contexts/SettingsContext.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
it('should read existing settings from localStorage', () => {
  const savedSettings = { theme: 'light', defaultView: 'week' }
  localStorage.setItem('bujo-settings', JSON.stringify(savedSettings))

  render(
    <SettingsProvider>
      <TestComponent />
    </SettingsProvider>
  )

  expect(screen.getByTestId('theme')).toHaveTextContent('light')
  expect(screen.getByTestId('defaultView')).toHaveTextContent('week')
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: FAIL - theme shows 'dark' instead of 'light'

**Step 3: Write minimal implementation**

Modify: `frontend/src/contexts/SettingsContext.tsx`

```typescript
import { createContext, useContext, ReactNode, useState, useEffect } from 'react'
import { Settings, DEFAULT_SETTINGS } from '@/types/settings'

interface SettingsContextValue extends Settings {
  setTheme: (theme: Settings['theme']) => void
  setDefaultView: (view: Settings['defaultView']) => void
}

const SettingsContext = createContext<SettingsContextValue | undefined>(undefined)

const STORAGE_KEY = 'bujo-settings'

function loadSettings(): Settings {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const parsed = JSON.parse(stored)
      return { ...DEFAULT_SETTINGS, ...parsed }
    }
  } catch (error) {
    console.warn('Failed to load settings from localStorage:', error)
  }
  return DEFAULT_SETTINGS
}

interface SettingsProviderProps {
  children: ReactNode
}

export function SettingsProvider({ children }: SettingsProviderProps) {
  const [settings, setSettings] = useState<Settings>(loadSettings)

  const setTheme = () => {}
  const setDefaultView = () => {}

  const value: SettingsContextValue = {
    ...settings,
    setTheme,
    setDefaultView,
  }

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  )
}

export function useSettings() {
  const context = useContext(SettingsContext)
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider')
  }
  return context
}
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (2 tests)

**Step 5: Commit**

```bash
git add frontend/src/contexts/SettingsContext.tsx frontend/src/contexts/SettingsContext.test.tsx
git commit -m "feat: load settings from localStorage

SettingsProvider now reads from localStorage on mount.
Falls back to defaults if not found or invalid.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Settings Context - Persist Theme Changes

**Files:**
- Modify: `frontend/src/contexts/SettingsContext.test.tsx`
- Modify: `frontend/src/contexts/SettingsContext.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
import { renderHook, act } from '@testing-library/react'

it('should persist theme changes to localStorage', () => {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <SettingsProvider>{children}</SettingsProvider>
  )

  const { result } = renderHook(() => useSettings(), { wrapper })

  act(() => {
    result.current.setTheme('light')
  })

  expect(result.current.theme).toBe('light')
  const stored = JSON.parse(localStorage.getItem('bujo-settings')!)
  expect(stored.theme).toBe('light')
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: FAIL - theme doesn't change

**Step 3: Write minimal implementation**

Modify: `frontend/src/contexts/SettingsContext.tsx`

```typescript
export function SettingsProvider({ children }: SettingsProviderProps) {
  const [settings, setSettings] = useState<Settings>(loadSettings)

  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
    } catch (error) {
      console.error('Failed to save settings to localStorage:', error)
    }
  }, [settings])

  const setTheme = (theme: Settings['theme']) => {
    setSettings(prev => ({ ...prev, theme }))
  }

  const setDefaultView = () => {}

  const value: SettingsContextValue = {
    ...settings,
    setTheme,
    setDefaultView,
  }

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  )
}
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add frontend/src/contexts/SettingsContext.tsx frontend/src/contexts/SettingsContext.test.tsx
git commit -m "feat: persist theme changes to localStorage

setTheme now updates state and saves to localStorage.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Settings Context - Persist Default View Changes

**Files:**
- Modify: `frontend/src/contexts/SettingsContext.test.tsx`
- Modify: `frontend/src/contexts/SettingsContext.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
it('should persist defaultView changes to localStorage', () => {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <SettingsProvider>{children}</SettingsProvider>
  )

  const { result } = renderHook(() => useSettings(), { wrapper })

  act(() => {
    result.current.setDefaultView('search')
  })

  expect(result.current.defaultView).toBe('search')
  const stored = JSON.parse(localStorage.getItem('bujo-settings')!)
  expect(stored.defaultView).toBe('search')
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: FAIL - defaultView doesn't change

**Step 3: Write minimal implementation**

Modify: `frontend/src/contexts/SettingsContext.tsx`

```typescript
export function SettingsProvider({ children }: SettingsProviderProps) {
  const [settings, setSettings] = useState<Settings>(loadSettings)

  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
    } catch (error) {
      console.error('Failed to save settings to localStorage:', error)
    }
  }, [settings])

  const setTheme = (theme: Settings['theme']) => {
    setSettings(prev => ({ ...prev, theme }))
  }

  const setDefaultView = (defaultView: Settings['defaultView']) => {
    setSettings(prev => ({ ...prev, defaultView }))
  }

  const value: SettingsContextValue = {
    ...settings,
    setTheme,
    setDefaultView,
  }

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  )
}
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add frontend/src/contexts/SettingsContext.tsx frontend/src/contexts/SettingsContext.test.tsx
git commit -m "feat: persist defaultView changes to localStorage

setDefaultView now updates state and saves to localStorage.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Settings Context - Invalid Data Handling

**Files:**
- Modify: `frontend/src/contexts/SettingsContext.test.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
it('should fall back to defaults for invalid localStorage data', () => {
  localStorage.setItem('bujo-settings', 'invalid json')

  render(
    <SettingsProvider>
      <TestComponent />
    </SettingsProvider>
  )

  expect(screen.getByTestId('theme')).toHaveTextContent('dark')
  expect(screen.getByTestId('defaultView')).toHaveTextContent('today')
})

it('should merge partial settings with defaults', () => {
  localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'system' }))

  render(
    <SettingsProvider>
      <TestComponent />
    </SettingsProvider>
  )

  expect(screen.getByTestId('theme')).toHaveTextContent('system')
  expect(screen.getByTestId('defaultView')).toHaveTextContent('today')
})
```

**Step 2: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (6 tests) - implementation already handles this

**Step 3: Commit**

```bash
git add frontend/src/contexts/SettingsContext.test.tsx
git commit -m "test: add tests for invalid settings data handling

Verify fallback to defaults for invalid JSON and partial settings.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Settings Context - Hook Error Handling

**Files:**
- Modify: `frontend/src/contexts/SettingsContext.test.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
it('should throw error when useSettings is used outside provider', () => {
  const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})

  expect(() => {
    renderHook(() => useSettings())
  }).toThrow('useSettings must be used within a SettingsProvider')

  consoleError.mockRestore()
})
```

**Step 2: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (7 tests) - implementation already handles this

**Step 3: Commit**

```bash
git add frontend/src/contexts/SettingsContext.test.tsx
git commit -m "test: verify useSettings throws outside provider

Add test for error when hook used without SettingsProvider.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Theme Integration - Apply Dark Class

**Files:**
- Modify: `frontend/src/contexts/SettingsContext.tsx`
- Modify: `frontend/src/contexts/SettingsContext.test.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
it('should apply dark class to html element when theme is dark', () => {
  render(
    <SettingsProvider>
      <TestComponent />
    </SettingsProvider>
  )

  expect(document.documentElement.classList.contains('dark')).toBe(true)
})

it('should remove dark class when theme is light', () => {
  localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'light', defaultView: 'today' }))

  render(
    <SettingsProvider>
      <TestComponent />
    </SettingsProvider>
  )

  expect(document.documentElement.classList.contains('dark')).toBe(false)
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: FAIL - dark class not applied

**Step 3: Write minimal implementation**

Modify: `frontend/src/contexts/SettingsContext.tsx`

Add after the first useEffect:

```typescript
useEffect(() => {
  const isDark = settings.theme === 'dark' ||
    (settings.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)

  document.documentElement.classList.toggle('dark', isDark)
}, [settings.theme])
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (9 tests)

**Step 5: Commit**

```bash
git add frontend/src/contexts/SettingsContext.tsx frontend/src/contexts/SettingsContext.test.tsx
git commit -m "feat: apply dark class based on theme setting

Add useEffect to toggle dark class on html element.
Supports dark, light, and system themes.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Theme Integration - System Preference Listener

**Files:**
- Modify: `frontend/src/contexts/SettingsContext.tsx`
- Modify: `frontend/src/contexts/SettingsContext.test.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/contexts/SettingsContext.test.tsx`

```typescript
it('should respect system preference when theme is system', () => {
  const matchMediaMock = vi.fn().mockImplementation((query) => ({
    matches: query === '(prefers-color-scheme: dark)',
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }))
  window.matchMedia = matchMediaMock

  localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'system', defaultView: 'today' }))

  render(
    <SettingsProvider>
      <TestComponent />
    </SettingsProvider>
  )

  expect(document.documentElement.classList.contains('dark')).toBe(true)
})

it('should update dark class when system preference changes', () => {
  let listener: ((e: MediaQueryListEvent) => void) | null = null
  const matchMediaMock = vi.fn().mockImplementation((query) => ({
    matches: query === '(prefers-color-scheme: dark)',
    addEventListener: vi.fn((_, l) => { listener = l }),
    removeEventListener: vi.fn(),
  }))
  window.matchMedia = matchMediaMock

  localStorage.setItem('bujo-settings', JSON.stringify({ theme: 'system', defaultView: 'today' }))

  render(
    <SettingsProvider>
      <TestComponent />
    </SettingsProvider>
  )

  expect(document.documentElement.classList.contains('dark')).toBe(true)

  // Simulate system preference change to light
  act(() => {
    listener?.({ matches: false } as MediaQueryListEvent)
  })

  expect(document.documentElement.classList.contains('dark')).toBe(false)
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: FAIL - addEventListener not called

**Step 3: Write minimal implementation**

Modify: `frontend/src/contexts/SettingsContext.tsx`

Replace the theme useEffect with:

```typescript
useEffect(() => {
  const applyTheme = () => {
    const isDark = settings.theme === 'dark' ||
      (settings.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)

    document.documentElement.classList.toggle('dark', isDark)
  }

  applyTheme()

  if (settings.theme === 'system') {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = () => applyTheme()

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }
}, [settings.theme])
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/contexts/SettingsContext.test.tsx
```

Expected: PASS (11 tests)

**Step 5: Commit**

```bash
git add frontend/src/contexts/SettingsContext.tsx frontend/src/contexts/SettingsContext.test.tsx
git commit -m "feat: listen to system theme preference changes

When theme is 'system', listen to matchMedia changes.
Update dark class when system preference changes.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 10: Wrap App in SettingsProvider

**Files:**
- Modify: `frontend/src/main.tsx`

**Step 1: No test needed (integration point)**

**Step 2: Write implementation**

Modify: `frontend/src/main.tsx`

Find the existing structure and wrap `<App />` with `<SettingsProvider>`:

```typescript
import { SettingsProvider } from './contexts/SettingsContext'

// ... existing code ...

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <SettingsProvider>
      <App />
    </SettingsProvider>
  </React.StrictMode>,
)
```

**Step 3: Verify app still runs**

```bash
cd frontend
npm run dev
```

Expected: App loads successfully, no console errors

**Step 4: Commit**

```bash
git add frontend/src/main.tsx
git commit -m "feat: wrap App in SettingsProvider

Make settings context available throughout app.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 11: Use Default View in App.tsx

**Files:**
- Modify: `frontend/src/App.tsx`
- Create: `frontend/src/App.test.tsx`

**Step 1: Write the failing test**

Create: `frontend/src/App.test.tsx`

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SettingsProvider } from '@/contexts/SettingsContext'
import App from './App'

// Mock Wails runtime
vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}))

// Mock all Wails Go bindings
vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn(() => Promise.resolve({ Days: [], Overdue: [] })),
  GetHabits: vi.fn(() => Promise.resolve({ Habits: [] })),
  GetLists: vi.fn(() => Promise.resolve([])),
  GetGoals: vi.fn(() => Promise.resolve([])),
  GetOutstandingQuestions: vi.fn(() => Promise.resolve([])),
  AddEntry: vi.fn(),
  AddChildEntry: vi.fn(),
  MarkEntryDone: vi.fn(),
  MarkEntryUndone: vi.fn(),
  EditEntry: vi.fn(),
  DeleteEntry: vi.fn(),
  HasChildren: vi.fn(() => Promise.resolve(false)),
  MigrateEntry: vi.fn(),
  MoveEntryToList: vi.fn(),
  MoveEntryToRoot: vi.fn(),
  OpenFileDialog: vi.fn(),
  CyclePriority: vi.fn(),
  CancelEntry: vi.fn(),
}))

describe('App default view', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('should initialize with default view from settings', async () => {
    localStorage.setItem('bujo-settings', JSON.stringify({
      theme: 'dark',
      defaultView: 'search'
    }))

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    // Wait for loading to complete
    await screen.findByText(/Search/i)

    // Should show search view based on default setting
    expect(screen.getByText(/Search/i)).toBeInTheDocument()
  })
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/App.test.tsx
```

Expected: FAIL - shows "Journal" (today view) instead of "Search"

**Step 3: Write minimal implementation**

Modify: `frontend/src/App.tsx`

Add import at top:

```typescript
import { useSettings } from '@/contexts/SettingsContext'
```

Inside App function, replace:

```typescript
const [view, setView] = useState<ViewType>('today')
```

With:

```typescript
const { defaultView } = useSettings()
const [view, setView] = useState<ViewType>(defaultView)
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/App.test.tsx
```

Expected: PASS (1 test)

**Step 5: Commit**

```bash
git add frontend/src/App.tsx frontend/src/App.test.tsx
git commit -m "feat: initialize app with user's default view

App now reads defaultView from settings context.
Initial view respects user preference.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 12: Settings View - Theme Selector UI

**Files:**
- Modify: `frontend/src/components/bujo/SettingsView.tsx`
- Modify: `frontend/src/components/bujo/SettingsView.test.tsx`

**Step 1: Write the failing test**

Modify: `frontend/src/components/bujo/SettingsView.test.tsx`

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SettingsProvider } from '@/contexts/SettingsContext'
import { SettingsView } from './SettingsView'

describe('SettingsView', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('should display current theme setting', () => {
    localStorage.setItem('bujo-settings', JSON.stringify({
      theme: 'light',
      defaultView: 'today'
    }))

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    expect(screen.getByRole('button', { name: /light/i })).toHaveAttribute('data-active', 'true')
  })

  it('should update theme when clicking theme option', async () => {
    const user = userEvent.setup()

    render(
      <SettingsProvider>
        <SettingsView />
      </SettingsProvider>
    )

    const systemButton = screen.getByRole('button', { name: /system/i })
    await user.click(systemButton)

    expect(systemButton).toHaveAttribute('data-active', 'true')

    const stored = JSON.parse(localStorage.getItem('bujo-settings')!)
    expect(stored.theme).toBe('system')
  })
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/components/bujo/SettingsView.test.tsx
```

Expected: FAIL - button not found

**Step 3: Write minimal implementation**

Modify: `frontend/src/components/bujo/SettingsView.tsx`

Add import at top:

```typescript
import { useSettings } from '@/contexts/SettingsContext'
import type { Theme } from '@/types/settings'
```

Replace the Theme SettingRow section:

```typescript
const { theme, setTheme, defaultView, setDefaultView } = useSettings()

// ... in the JSX, replace the Appearance section SettingRow:

<SettingRow
  label="Theme"
  description="Choose your preferred color theme"
>
  <div className="flex gap-1">
    {(['light', 'dark', 'system'] as const).map((option) => (
      <button
        key={option}
        onClick={() => setTheme(option)}
        data-active={theme === option}
        className="px-3 py-1.5 text-xs rounded-md transition-colors data-[active=true]:bg-primary data-[active=true]:text-primary-foreground hover:bg-secondary"
      >
        {option.charAt(0).toUpperCase() + option.slice(1)}
      </button>
    ))}
  </div>
</SettingRow>
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/components/bujo/SettingsView.test.tsx
```

Expected: PASS (2 tests)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/SettingsView.tsx frontend/src/components/bujo/SettingsView.test.tsx
git commit -m "feat: add interactive theme selector to SettingsView

Replace hardcoded 'Dark' with segmented control.
Shows current theme and updates on click.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 13: Settings View - Default View Selector UI

**Files:**
- Modify: `frontend/src/components/bujo/SettingsView.tsx`
- Modify: `frontend/src/components/bujo/SettingsView.test.tsx`

**Step 1: Write the failing test**

Add to: `frontend/src/components/bujo/SettingsView.test.tsx`

```typescript
it('should display current default view setting', () => {
  localStorage.setItem('bujo-settings', JSON.stringify({
    theme: 'dark',
    defaultView: 'week'
  }))

  render(
    <SettingsProvider>
      <SettingsView />
    </SettingsProvider>
  )

  expect(screen.getByRole('combobox')).toHaveValue('week')
})

it('should update default view when selecting option', async () => {
  const user = userEvent.setup()

  render(
    <SettingsProvider>
      <SettingsView />
    </SettingsProvider>
  )

  const select = screen.getByRole('combobox')
  await user.selectOptions(select, 'search')

  expect(select).toHaveValue('search')

  const stored = JSON.parse(localStorage.getItem('bujo-settings')!)
  expect(stored.defaultView).toBe('search')
})
```

**Step 2: Run test to verify it fails**

```bash
cd frontend
npm test src/components/bujo/SettingsView.test.tsx
```

Expected: FAIL - combobox not found

**Step 3: Write minimal implementation**

Modify: `frontend/src/components/bujo/SettingsView.tsx`

Add import:

```typescript
import type { DefaultView } from '@/types/settings'
```

Replace the "Default View" SettingRow:

```typescript
<SettingRow
  label="Default View"
  description="The view shown when you open the app"
>
  <select
    value={defaultView}
    onChange={(e) => setDefaultView(e.target.value as DefaultView)}
    className="px-3 py-1.5 text-sm rounded-md bg-secondary hover:bg-secondary/80 transition-colors border-none focus:outline-none focus:ring-2 focus:ring-primary/50"
  >
    <option value="today">Today</option>
    <option value="week">Week</option>
    <option value="overview">Overview</option>
    <option value="search">Search</option>
  </select>
</SettingRow>
```

**Step 4: Run test to verify it passes**

```bash
cd frontend
npm test src/components/bujo/SettingsView.test.tsx
```

Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add frontend/src/components/bujo/SettingsView.tsx frontend/src/components/bujo/SettingsView.test.tsx
git commit -m "feat: add interactive default view selector

Replace hardcoded 'Today' with dropdown.
Shows current default view and updates on change.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 14: Backend Version Display

**Files:**
- Check if Wails binding exists
- Modify: `frontend/src/components/bujo/SettingsView.tsx`
- Modify: `frontend/src/components/bujo/SettingsView.test.tsx`

**Step 1: Check for existing GetVersion binding**

```bash
grep -r "GetVersion" frontend/src/wailsjs/
```

If not found, we need to add it to the Go backend first.

**Step 2: Write the failing test**

Add to: `frontend/src/components/bujo/SettingsView.test.tsx`

```typescript
import { vi } from 'vitest'

// Add to top of file with other mocks
vi.mock('@/wailsjs/go/wails/App', () => ({
  GetVersion: vi.fn(() => Promise.resolve('1.2.3')),
}))

it('should fetch and display version from backend', async () => {
  const { GetVersion } = await import('@/wailsjs/go/wails/App')

  render(
    <SettingsProvider>
      <SettingsView />
    </SettingsProvider>
  )

  expect(await screen.findByText('1.2.3')).toBeInTheDocument()
  expect(GetVersion).toHaveBeenCalled()
})

it('should show loading state while fetching version', () => {
  const { GetVersion } = require('@/wailsjs/go/wails/App')
  GetVersion.mockImplementation(() => new Promise(() => {})) // Never resolves

  render(
    <SettingsProvider>
      <SettingsView />
    </SettingsProvider>
  )

  expect(screen.getByText(/loading/i)).toBeInTheDocument()
})

it('should show error state if version fetch fails', async () => {
  const { GetVersion } = require('@/wailsjs/go/wails/App')
  GetVersion.mockRejectedValue(new Error('Failed'))

  render(
    <SettingsProvider>
      <SettingsView />
    </SettingsProvider>
  )

  expect(await screen.findByText(/unknown/i)).toBeInTheDocument()
})
```

**Step 3: Run test to verify it fails**

```bash
cd frontend
npm test src/components/bujo/SettingsView.test.tsx
```

Expected: FAIL - GetVersion import fails or version not displayed

**Step 4: Add GetVersion to Wails bindings (if needed)**

Check: `frontend/src/wailsjs/go/wails/App.js`

If `GetVersion` doesn't exist, add to Go backend:

```bash
# In app.go or version.go
func (a *App) GetVersion() string {
    return "0.0.0" // Or read from build info
}
```

Then regenerate Wails bindings:

```bash
wails generate module
```

**Step 5: Write minimal implementation**

Modify: `frontend/src/components/bujo/SettingsView.tsx`

Add imports:

```typescript
import { useState, useEffect } from 'react'
import { GetVersion } from '@/wailsjs/go/wails/App'
```

Add state inside component:

```typescript
const [version, setVersion] = useState<string>('Loading...')

useEffect(() => {
  GetVersion()
    .then(setVersion)
    .catch(() => setVersion('Unknown'))
}, [])
```

Replace the Version SettingRow:

```typescript
<SettingRow
  label="Version"
  description="Current application version"
>
  <span className="text-sm text-muted-foreground">{version}</span>
</SettingRow>
```

**Step 6: Run test to verify it passes**

```bash
cd frontend
npm test src/components/bujo/SettingsView.test.tsx
```

Expected: PASS (7 tests)

**Step 7: Commit**

```bash
git add frontend/src/components/bujo/SettingsView.tsx frontend/src/components/bujo/SettingsView.test.tsx
git commit -m "feat: display version from backend API

Replace hardcoded version with GetVersion() call.
Shows loading and error states.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 15: Flash Prevention Script

**Files:**
- Modify: `frontend/index.html`

**Step 1: No test needed (performance optimization)**

**Step 2: Write implementation**

Modify: `frontend/index.html`

Add before the `<div id="root"></div>`:

```html
<script>
  // Prevent flash of unstyled content for dark mode
  (function() {
    try {
      const settings = JSON.parse(localStorage.getItem('bujo-settings') || '{}');
      const theme = settings.theme || 'dark';

      if (theme === 'dark') {
        document.documentElement.classList.add('dark');
      } else if (theme === 'system') {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        if (prefersDark) {
          document.documentElement.classList.add('dark');
        }
      }
    } catch (e) {
      // If anything fails, default to dark mode
      document.documentElement.classList.add('dark');
    }
  })();
</script>
```

**Step 3: Manual verification**

```bash
cd frontend
npm run dev
```

Test:
1. Set theme to light in settings
2. Refresh page
3. Verify no white flash appears

**Step 4: Commit**

```bash
git add frontend/index.html
git commit -m "feat: prevent flash of unstyled content

Add inline script to apply dark class before React hydrates.
Eliminates white flash on page load in dark mode.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 16: Final Integration Test

**Files:**
- Run full test suite

**Step 1: Run all tests**

```bash
cd frontend
npm test
```

Expected: All tests pass (including existing tests)

**Step 2: Manual integration testing**

```bash
npm run dev
```

Test flow:
1. Open app - should show default view (today)
2. Navigate to Settings
3. Change theme to Light - UI should update immediately
4. Change theme to System - should follow system preference
5. Change default view to Search
6. Refresh app - should open to Search view
7. Verify version displays correctly
8. Change theme to Dark
9. Refresh - no white flash should appear

**Step 3: If all tests pass and manual testing succeeds, commit**

```bash
git add -A
git commit -m "test: verify complete settings integration

All unit and integration tests passing.
Manual testing confirms expected behavior.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Implementation Complete

**Summary:**
- ✅ Settings type definitions with tests
- ✅ SettingsContext with full localStorage integration
- ✅ Theme switching with system preference support
- ✅ Default view integration in App.tsx
- ✅ Interactive Settings UI (theme selector, view selector)
- ✅ Backend version display
- ✅ Flash prevention for dark mode
- ✅ Comprehensive test coverage

**Next Steps:**
1. Create pull request
2. Review and merge
3. Consider future enhancements (font size, density, etc.)
