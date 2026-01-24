# Writing Acceptance Tests

This guide documents patterns for writing acceptance tests in the bujo frontend using Vitest and React Testing Library.

## Test Infrastructure

### Mocking Wails Functions

**ALWAYS use Wails function mocks, NEVER fetch mocks:**

```typescript
import { GetAgenda, MigrateEntry } from './wailsjs/go/wails/App'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

// Mock the Wails module
vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn(),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  // ... other Wails functions
}))

// In beforeEach
beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
  // Default: empty agenda
  vi.mocked(GetAgenda).mockResolvedValue(createMockAgenda({ Days: [] }))
})
```

### Creating Mock Data

Use test helpers that match the Go backend structure:

```typescript
// Create a single entry
const entry = createMockEntry({
  ID: 1,
  Content: 'Test task',
  Type: 'Task',
  Priority: 'None',
  ParentID: null,
  CreatedAt: '2026-01-24T10:00:00Z',
})

// Create day with entries
const day = createMockDayEntries({
  Date: '2026-01-24T00:00:00Z',
  Entries: [entry]
})

// Create full agenda
const agenda = createMockAgenda({
  Days: [day]
})

// Use in test
vi.mocked(GetAgenda).mockResolvedValue(agenda)
```

### Parent-Child Relationships

Go entries use ParentID, not nested children:

```typescript
// ✓ Correct: Flat structure with ParentID
const agenda = createMockAgenda({
  Days: [
    createMockDayEntries({
      Entries: [
        createMockEntry({ ID: 1, Content: 'Parent', ParentID: null }),
        createMockEntry({ ID: 2, Content: 'Child', ParentID: 1 })
      ]
    })
  ]
})

// ✗ Wrong: Nested children (fetch-style mock)
const agenda = {
  days: [{
    entries: [{
      id: 1,
      children: [{ id: 2, parentId: 1 }]
    }]
  }]
}
```

## Test Structure

### Basic Test Pattern

```typescript
describe('Feature: Description', () => {
  it('should do expected behavior', async () => {
    // 1. Setup mock data
    const agendaData = createMockAgenda({
      Days: [
        createMockDayEntries({
          Date: '2026-01-24T00:00:00Z',
          Entries: [
            createMockEntry({
              ID: 1,
              Content: 'Test entry',
              Type: 'Task',
            })
          ]
        })
      ]
    })
    vi.mocked(GetAgenda).mockResolvedValue(agendaData)

    // 2. Render component
    render(<App />)
    const user = userEvent.setup()

    // 3. Wait for component to load
    await waitFor(() => {
      expect(screen.getByText(/Test entry/i)).toBeInTheDocument()
    })

    // 4. Interact with component
    const entry = screen.getByText(/Test entry/i)
    await user.click(entry)

    // 5. Assert expected outcome
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })
  })
})
```

### Testing Mutations

When testing Wails function calls:

```typescript
import { MigrateEntry } from './wailsjs/go/wails/App'

// In test
await user.keyboard('m')

await waitFor(() => {
  expect(vi.mocked(MigrateEntry)).toHaveBeenCalledWith(1, expect.any(String))
})
```

### Testing Navigation

Use sidebar buttons for navigation:

```typescript
// Navigate to Week view
const weekButton = screen.getByRole('button', { name: /Weekly Review/i })
await user.click(weekButton)

await waitFor(() => {
  expect(screen.getByText(/Meetings/i)).toBeInTheDocument()
})
```

## Common Patterns

### Testing Popovers

```typescript
// Open popover
const entry = screen.getByText(/Task name/i)
await user.click(entry)

await waitFor(() => {
  expect(screen.getByRole('dialog')).toBeInTheDocument()
})

// Close popover
await user.click(document.body)

await waitFor(() => {
  expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
})
```

### Testing Keyboard Shortcuts

```typescript
// Open popover first
const entry = screen.getByText(/Task name/i)
await user.click(entry)

await waitFor(() => {
  expect(screen.getByRole('dialog')).toBeInTheDocument()
})

// Press keyboard shortcut
await user.keyboard('m')

// Verify action
await waitFor(() => {
  expect(vi.mocked(MigrateEntry)).toHaveBeenCalled()
})
```

### Testing Lists

```typescript
// Check order
const items = screen.getAllByRole('button', { name: /task|question/i })
expect(items[0]).toHaveTextContent(/First item/i)
expect(items[1]).toHaveTextContent(/Second item/i)

// Check item not in list
expect(screen.queryByText(/Not shown/i)).not.toBeInTheDocument()
```

## Anti-Patterns

### ✗ Using fetch mocks

```typescript
// WRONG - App doesn't use fetch
global.fetch = vi.fn()
;(global.fetch as any).mockImplementation(...)
```

### ✗ Nested children in mock data

```typescript
// WRONG - Go backend doesn't have Children field
{
  id: 1,
  children: [{ id: 2, parentId: 1 }]
}
```

### ✗ Testing implementation details

```typescript
// WRONG - testing mock behavior
expect(global.fetch).toHaveBeenCalled()

// RIGHT - testing component behavior
expect(screen.getByRole('dialog')).toBeInTheDocument()
```

## Debugging

### Check if data is reaching component

```typescript
// Add debug output
screen.debug()

// Check specific element
screen.debug(screen.getByTestId('entry-tree'))
```

### Verify mock was called

```typescript
console.log(vi.mocked(GetAgenda).mock.calls)
```

### Check rendered HTML

```typescript
const element = screen.getByText(/Test/i)
console.log(element.outerHTML)
```

## Entry Types

Valid entry types (case-sensitive in mocks):

- `Task` - Unfinished task
- `Done` - Completed task
- `Event` - Calendar event
- `Note` - Plain note
- `Question` - Question/inquiry
- `Migrated` - Task migrated to another day

## Date Format

Always use ISO 8601 format with timezone:

```typescript
CreatedAt: '2026-01-24T10:00:00Z'
Date: '2026-01-24T00:00:00Z'
```

## Priority Values

- `None` (default)
- `Low`
- `Medium`
- `High`

## Navigation History Testing

### Multi-Level Navigation Stack

The app supports multi-level navigation history. Tests must verify that:

1. **Sidebar navigation** pushes to history (except when navigating TO 'today' which clears)
2. **Programmatic navigation** (e.g., popover "Go to") always pushes to history
3. **Back button** remains visible while history exists
4. **Multiple back operations** are supported

```typescript
// Correct multi-level navigation test
it('supports multi-level navigation history', async () => {
  render(<App />)
  const user = userEvent.setup()

  // Navigate: today → week (pushes today to history)
  await user.click(screen.getByRole('button', { name: /Weekly Review/i }))

  // Navigate: week → today via popover (pushes week to history)
  await user.click(screen.getByText(/Task needing attention/i))
  await user.click(screen.getByRole('button', { name: /go to/i }))

  // Back: today → week (history still has today)
  await user.click(screen.getByRole('button', { name: /go back/i }))
  expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()

  // Back again: week → today (history now empty)
  await user.click(screen.getByRole('button', { name: /go back/i }))
  expect(screen.queryByRole('button', { name: /go back/i })).not.toBeInTheDocument()
})
```

### Navigation to 'Today' Clears History

Sidebar navigation TO the 'today' view (labeled "Journal") clears all history:

```typescript
it('clears history when navigating to today via sidebar', async () => {
  render(<App />)
  const user = userEvent.setup()

  // Create history by navigating away
  await user.click(screen.getByRole('button', { name: /Weekly Review/i }))
  expect(screen.getByRole('button', { name: /go back/i })).toBeInTheDocument()

  // Navigate back to today via sidebar - clears history
  await user.click(screen.getByRole('button', { name: /journal/i }))
  expect(screen.queryByRole('button', { name: /go back/i })).not.toBeInTheDocument()
})
```

### Finding Sidebar Buttons by Label

Sidebar buttons use specific labels, not view names:

- 'today' view → "Journal" button
- 'week' view → "Weekly Review" button
- 'habits' view → "Habit Tracker" button
- 'overview' view → "Pending Tasks" button

```typescript
// ✓ Correct
const journalButton = screen.getByRole('button', { name: /journal/i })

// ✗ Wrong
const todayButton = screen.getByRole('button', { name: /today/i })
```

## Custom Hook Testing

### Return Values from Hooks with setState

When a custom hook needs to return a value derived from state, read the state BEFORE calling setState:

```typescript
// ✗ Wrong: Variable assignment inside setState callback returns null
const goBack = useCallback(() => {
  let popped: NavigationState | null = null
  setHistory((prev) => {
    popped = prev[prev.length - 1]
    return prev.slice(0, -1)
  })
  return popped  // Always null!
}, [])

// ✓ Correct: Read state before setState
const goBack = useCallback(() => {
  const current = history[history.length - 1] || null
  setHistory((prev) => prev.slice(0, -1))
  return current
}, [history])
```

**Why:** Variable assignments inside setState callbacks execute asynchronously and don't capture values for the return statement.

### Testing Hook Return Values

Always verify hook return values match expected behavior:

```typescript
it('returns history state on goBack', () => {
  const { result } = renderHook(() => useNavigationHistory())

  act(() => {
    result.current.pushHistory({ view: 'week', scrollPosition: 100 })
  })

  const returned = act(() => result.current.goBack())

  // Verify the returned value is correct
  expect(returned).toEqual({ view: 'week', scrollPosition: 100 })
})
```

## Test Maintenance

### Intentionally Skipped Tests

Some tests are intentionally skipped due to known limitations:

```typescript
it.skip('changing date picker navigates to selected date', async () => {
  // Skipped: jsdom doesn't reliably handle HTML date input change events
  // Same behavior tested via next/prev buttons and keyboard shortcuts
})
```

**Document why tests are skipped** in commit messages and comments. Check git history:

```bash
git log --oneline --all --grep="skip"
```

### Conflicting Test Requirements

When tests fail, verify they don't conflict with other test requirements:

**Red flag:** Same user flow, different expectations

```typescript
// Test A expects: back button disappears after one back
// Bug #9 expects: back button persists (multi-level navigation)
// Flow in both: today → week → today → back → week
```

**Resolution strategy:**
1. Check for explicit bug/feature numbers (e.g., "Bug #9")
2. Named bugs/features take precedence over generic tests
3. Update generic tests to match named requirements
4. Look for commit history explaining the intent

### Test Naming Best Practices

Test names should describe actual behavior, not incorrect assumptions:

```typescript
// ✗ Bad: Name describes wrong behavior
it('back button disappears after going back', async () => {
  // Test expects single-level navigation (conflicts with Bug #9)
})

// ✓ Good: Name describes actual multi-level behavior
it('supports multi-level navigation history', async () => {
  // Test verifies back button persists through multiple backs
})
```

## Lessons Learned

### React State Closures in useCallback

When using `useCallback` with state dependencies, the callback captures current state:

```typescript
// This is correct - history in dependency array means callback sees current history
const goBack = useCallback(() => {
  const current = history[history.length - 1] || null
  setHistory((prev) => prev.slice(0, -1))
  return current
}, [history])  // ← history dependency is necessary
```

Previously attempted to remove `[history]` dependency to avoid "stale closures," but this prevents reading current state. The correct pattern is to keep the dependency and read state before mutating it.

### Test Fixture Data Reuse

Tests with identical flows may use the same mock data but test different aspects:

```typescript
// Both tests use weekAgendaWithAttentionItems but test different navigation expectations
const weekAgendaWithAttentionItems = createMockAgenda({
  Days: [createMockDayEntries({
    Entries: [
      createMockEntry({ Type: 'Task', Priority: 'high', Content: 'Task needing attention' })
    ]
  })]
})
```

Shared mock data is fine as long as test expectations are consistent with actual behavior.

### scrollIntoView Mocking

JSDOM doesn't implement `scrollIntoView`. Mock it globally in test setup:

```typescript
// In test file setup
Element.prototype.scrollIntoView = vi.fn()
```

Missing this mock causes "Unhandled Error" messages but doesn't fail tests. Add the mock for cleaner test output.

## See Also

- `src/test/mocks.ts` - Mock creation helpers
- `src/test/setup.ts` - Global test configuration
- `src/App.bugfixes.test.tsx` - Example of correct Wails mocking and multi-level navigation
- `src/App.historyNavigation.test.tsx` - Navigation history test patterns
- `src/components/bujo/WeekSummary.popover.test.tsx` - Component-level tests
- `src/hooks/__tests__/useNavigationHistory.test.ts` - Custom hook testing patterns
