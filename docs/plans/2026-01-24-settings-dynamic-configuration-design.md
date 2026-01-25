# Settings Dynamic Configuration Design

**Date:** 2026-01-24
**Status:** Approved

## Overview

Transform the Settings page from hardcoded display values to a fully functional configuration interface with persistent user preferences. This implementation focuses on theme selection and default view preferences stored in localStorage, plus displaying real version data from the backend.

## Scope

**In Scope:**
- Theme switcher (Light/Dark/System) with localStorage persistence
- Default view selector (Today/Week/Overview/Search) with localStorage persistence
- Version display from backend API
- Settings context for centralized state management

**Out of Scope:**
- Database location configuration (remains display-only)
- Backend settings API (using localStorage only for now)
- User profiles or cloud sync
- Other settings beyond theme and default view

## Data Model

### Settings Interface

```typescript
interface Settings {
  theme: 'light' | 'dark' | 'system'
  defaultView: 'today' | 'week' | 'overview' | 'search'
}
```

### Defaults

- `theme: 'dark'` (matches current hardcoded value)
- `defaultView: 'today'` (matches current hardcoded value)

### Storage

- **localStorage key:** `'bujo-settings'`
- **Format:** JSON string via `JSON.stringify(settings)`
- **Lifecycle:** Read on mount, write on change

### Type Safety

- Theme literal union prevents invalid values
- Default view constrained to the 4 primary navigation views
- TypeScript compile-time validation

### Migration Strategy

- First launch: Use defaults if localStorage is empty
- Version updates: Merge existing settings with new defaults
- Backward compatibility for future settings additions

## Architecture

### Settings Context

**File:** `src/contexts/SettingsContext.tsx`

**Components:**
- `SettingsContext` - React context
- `SettingsProvider` - Context provider component
- `useSettings()` - Consumer hook

**Provider Responsibilities:**

1. **Initialize:** Read from localStorage on mount, merge with defaults
2. **Expose:** Provide settings object and individual setters
3. **Persist:** Auto-save to localStorage on settings change
4. **Validate:** Ensure values match TypeScript types before saving

**Usage Pattern:**

```typescript
// In App.tsx
<SettingsProvider>
  <App />
</SettingsProvider>

// In any component
const { theme, setTheme, defaultView, setDefaultView } = useSettings()
```

**Error Handling:**
- localStorage read failure: Fall back to defaults, log warning
- localStorage write failure: Show error but don't crash
- Invalid stored values: Validate and merge with defaults

## Theme Implementation

### Strategy

Uses Tailwind CSS class-based dark mode with system preference detection.

### System Theme Detection

```typescript
const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
const effectiveTheme = theme === 'system' ? systemTheme : theme
```

### DOM Class Management

```typescript
// In SettingsProvider useEffect
document.documentElement.classList.toggle('dark', effectiveTheme === 'dark')
```

### System Preference Listener

- When theme is 'system', listen to `matchMedia` change events
- Update dark class when system preference changes
- Clean up listener on unmount or theme change

### Initial Flash Prevention

**Problem:** White flash on dark mode load before React hydrates

**Solution:** Inline script in `index.html` that runs before React
- Reads localStorage directly
- Applies `dark` class immediately
- Prevents flash of unstyled content

### Tailwind Configuration

```js
// tailwind.config.js (already configured)
darkMode: 'class'  // Uses .dark class on <html>
```

## Default View Integration

### App.tsx Changes

**1. Read Default View on Mount:**

```typescript
const { defaultView } = useSettings()
const [view, setView] = useState<ViewType>(defaultView)
```

**2. Behavior:**
- App launch shows saved default view
- Navigation works normally (sidebar, keyboard shortcuts)
- Default view only affects initial state

**3. No Automatic Updates:**
- Changing default view in settings doesn't immediately switch current view
- New default applies on next app launch
- Predictable UX - user not teleported mid-session

**Alternative Considered:**
- Immediately switch to new default view when changed
- **Rejected:** Disorienting UX if user is in middle of task

**Edge Cases:**
- Invalid view in localStorage: Fall back to 'today'
- Validation happens in SettingsProvider

## Settings View UI

### Theme Setting Row

**Current:** Hardcoded "Dark" text
**New:** Interactive segmented control

- Three options: "Light", "Dark", "System"
- Visual indicator for current selection
- On change: calls `setTheme(newValue)`
- Component: Segmented control (3 buttons, one active)

### Default View Setting Row

**Current:** Hardcoded "Today" text
**New:** Interactive dropdown

- Four options: "Today", "Week", "Overview", "Search"
- Shows current selection with friendly labels
- On change: calls `setDefaultView(newValue)`
- Component: Dropdown/select element

### Version Setting Row

**Current:** Hardcoded "1.0.0" text
**New:** Backend API call

- Create `GetVersion()` Wails binding (if doesn't exist)
- Fetch on component mount
- Show loading state while fetching
- Handle error state (show "Unknown" if fetch fails)

## Testing Strategy

### Settings Context Tests

**File:** `src/contexts/SettingsContext.test.tsx`

- Provider initializes with defaults when localStorage is empty
- Provider reads existing settings from localStorage
- `setTheme()` updates state and persists to localStorage
- `setDefaultView()` updates state and persists to localStorage
- Invalid localStorage data falls back to defaults
- Hook throws error when used outside provider

### Theme Integration Tests

**File:** `src/components/App.test.tsx`

- 'dark' theme adds `dark` class to `<html>`
- 'light' theme removes `dark` class from `<html>`
- 'system' theme respects `matchMedia` preference
- System theme change event updates class when theme is 'system'
- System theme change ignored when theme is 'light' or 'dark'

### Default View Tests

**File:** `src/components/App.test.tsx`

- App initializes with saved default view
- Changing default view in settings doesn't affect current view
- Invalid default view falls back to 'today'

### Settings View Component Tests

**File:** `src/components/bujo/SettingsView.test.tsx`

- Theme selector shows current theme
- Clicking theme option calls `setTheme`
- Default view selector shows current view
- Changing default view calls `setDefaultView`
- Version fetches from backend on mount
- Version shows loading state during fetch
- Version shows error state on fetch failure

### Mocking Strategy

- Mock localStorage (vitest provides this)
- Mock Wails `GetVersion()` call
- Mock `matchMedia` for system theme tests

## Implementation Order (TDD)

1. **Settings types** - Define Settings interface
2. **SettingsContext** - RED-GREEN-REFACTOR with full test coverage
3. **Theme integration** - Add to App.tsx with tests
4. **Default view integration** - Add to App.tsx with tests
5. **Settings UI components** - Theme selector, view selector, version display
6. **Flash prevention** - Inline script in index.html
7. **Integration testing** - End-to-end settings flow

## Files to Create

- `src/types/settings.ts` - Settings type definitions
- `src/contexts/SettingsContext.tsx` - Context provider and hook
- `src/contexts/SettingsContext.test.tsx` - Context tests

## Files to Modify

- `src/App.tsx` - Wrap in SettingsProvider, read defaultView
- `src/components/bujo/SettingsView.tsx` - Add interactive controls
- `src/components/bujo/SettingsView.test.tsx` - Add new tests
- `index.html` - Add flash prevention script (if needed)

## Future Extensibility

This design supports easy addition of new settings:

1. Add field to Settings interface
2. Add default value
3. Context automatically handles persistence
4. Add UI control in SettingsView

**Potential future settings:**
- Font size preference
- Compact/comfortable density
- Keyboard shortcut customization
- Auto-refresh interval
- Backend sync preferences
