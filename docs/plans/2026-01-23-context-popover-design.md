# Context Popover Design

> **For Claude:** This is a design document. Use superpowers:writing-plans to create an implementation plan from this design.

**Goal:** Add context popovers to WeekSummary and Pending Tasks views so users can see the full entry tree and take quick actions without leaving the view.

**Applies to:** WeekSummary (Needs Attention, Meetings sections), Pending Tasks (OverviewView)

---

## User Experience

### Opening the Popover

Click any entry in:
- WeekSummary → Needs Attention section
- WeekSummary → Meetings section
- Pending Tasks view

A popover appears anchored to the clicked item.

### Popover Content

Shows the full tree from root to the clicked entry using nested list style:

```
o Finance meeting
  - Attendees: Alice, Bob, Carol
  . Action items
    x Send slides to team
    . Review Q3 budget  ← highlighted
```

- Minimal styling: indents + bullet symbols only
- No separated boxes per entry
- Clicked entry is visually highlighted (background color)
- Shows siblings at each level, not just direct ancestors

### Popover Constraints

- Max height: 400px
- Scrollable when content exceeds max height
- Smart repositioning: flips above if no room below
- Click outside or `Escape` to dismiss

### Quick Actions

Actions appear at bottom of popover. Available actions depend on entry type:

| Entry Type | Actions |
|------------|---------|
| Task (`.`) | Done, Priority, Migrate |
| Question (`?`) | Answer, Priority |
| Done (`x`) | Undo |
| Event (`o`) | Priority |
| Note (`-`) | Priority |

Button layout:
```
───────────────────────────────────
[✓] [!] [>]           [Go to entry →]
```

Icons with tooltips: ✓ Done, ! Priority, > Migrate

After action:
- Entry updates in tree
- If entry no longer qualifies for list (e.g., marked done), popover closes and list refreshes
- Otherwise popover stays open

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` | Toggle done |
| `x` | Cancel/uncancel |
| `p` | Cycle priority |
| `m` | Migrate |
| `Enter` | Go to entry |
| `Escape` | Close popover |

These are consistent with shortcuts used elsewhere in the app (App.tsx, SearchView, QuestionsView, OverviewView).

---

## Navigation History

### "Go to Entry" Flow

1. User clicks "Go to entry" (or presses `Enter`)
2. App remembers current view and scroll position
3. Navigates to journal view for entry's date
4. Scrolls to and highlights the entry
5. Back button appears in header

### Back Button

```
┌─────────────────────────────────────────────────────┐
│  [← Back]   January 15, 2026        [+ Add Entry]   │
└─────────────────────────────────────────────────────┘
```

- Only appears when there's navigation history
- Click returns to exact scroll position in previous view
- `Escape` key also triggers back (when no popover/modal open)

### History Behavior

- Shallow history (one level only, not a stack)
- Navigating manually via sidebar clears back state
- If entry no longer exists on return, just return to view (no error)

---

## Component Structure

### New Components

**`EntryContextPopover`**
- Props: `entry`, `allEntries`, `onAction`, `onNavigate`, `onClose`
- Manages popover open/close state
- Uses Radix UI Popover for anchoring/positioning

**`EntryTree`**
- Props: `rootEntry`, `highlightedEntryId`, `entries`
- Pure presentation component
- Renders nested list with symbols and indentation

**`useNavigationHistory` hook**
- State: `{ view: string, scrollPosition: number } | null`
- Methods: `pushHistory(view, position)`, `goBack()`, `clearHistory()`

### Modified Components

**`WeekSummary`**
- Add `onEntryClick` prop
- Wrap attention items and meeting items with popover trigger

**`OverviewView`**
- Remove current inline expand (ContextPill + indented boxes)
- Replace with popover on entry click

**`App`**
- Add NavigationHistoryProvider context
- Handle `onNavigate` from popovers

**`Header`**
- Accept `onBack` prop
- Conditionally render back button

---

## Visual Style

### Tree Rendering

Nested list with minimal styling:
- 16px indent per level
- Bullet journal symbols: `.` task, `-` note, `o` event, `?` question, `x` done, `~` cancelled, `>` migrated
- Muted text color for ancestors
- Highlighted entry: subtle background (e.g., `bg-primary/10`)

### Popover Styling

- Border and shadow consistent with existing cards
- Separator line above action buttons
- "Go to entry" as text link on right side
- Icon buttons on left for quick actions
