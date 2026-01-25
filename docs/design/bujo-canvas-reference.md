# bujo-canvas UI Reference

> Design specifications from the [bujo-canvas](https://github.com/typingincolor/bujo-canvas) Lovable.dev mockup.
> Local clone: `~/Development/bujo-canvas`

## Overview

The mockup implements a paper-inspired bullet journal UI with:
- Three-panel layout (sidebar + main content + context panel)
- Warm cream/ivory color palette
- Serif display headings + sans-serif body text
- Keyboard-first interaction patterns
- Hierarchical entry management

---

## Layout Structure

```
┌─────────────────────────────────────────────────────────────┐
│ Sidebar (w-56)  │ Main Content (flex-1)  │ Context Panel    │
│                 │                        │ (w-80)           │
│ ┌─────────────┐ │ ┌────────────────────┐ │                  │
│ │ Logo        │ │ │ Header             │ │ Entry hierarchy  │
│ │ "bujo"      │ │ │ Title + Date       │ │ or               │
│ └─────────────┘ │ └────────────────────┘ │ Overdue items    │
│                 │                        │ (view-dependent) │
│ Navigation      │ Scrollable Content     │                  │
│ - Journal       │ (max-w-4xl centered)   │                  │
│ - Weekly Review │                        │                  │
│ - Pending Tasks │                        │                  │
│ - Questions     │                        │                  │
│ - Habits        │                        │                  │
│ - Lists         │                        │                  │
│ - Goals         │                        │                  │
│ - Search        │                        │                  │
│ - Insights      │ ┌────────────────────┐ │                  │
│                 │ │ CaptureBar (fixed) │ │                  │
│ ┌─────────────┐ │ │ bottom, sticky     │ │                  │
│ │ Settings    │ │ └────────────────────┘ │                  │
│ └─────────────┘ │                        │                  │
└─────────────────────────────────────────────────────────────┘
```

### Responsive Behavior
- Context panel: `hidden lg:flex` (hidden on mobile/tablet)
- Sidebar: Collapsible on smaller screens
- Main content: `max-w-4xl xl:max-w-5xl 2xl:max-w-6xl`

---

## Design System

### Color Palette (HSL)

#### Foundation
| Token | HSL Value | Usage |
|-------|-----------|-------|
| `--background` | `43 30% 96%` | Cream/ivory page background |
| `--foreground` | `25 20% 18%` | Warm dark brown text |
| `--card` | `40 25% 98%` | Aged paper white for cards |
| `--primary` | `28 85% 45%` | Warm amber/ochre accent |
| `--secondary` | `45 15% 92%` | Soft sage for hover states |
| `--accent` | `16 65% 55%` | Terracotta for highlights |
| `--muted` | `40 10% 85%` | Muted backgrounds |
| `--muted-foreground` | `25 10% 45%` | Secondary text |

#### Entry Type Colors
| Token | HSL Value | Symbol |
|-------|-----------|--------|
| `--bujo-task` | `25 20% 25%` | `•` (bullet) |
| `--bujo-note` | `200 15% 40%` | `–` (dash) |
| `--bujo-event` | `28 70% 50%` | `○` (circle) |
| `--bujo-done` | `145 35% 42%` | `✓` (checkmark) |
| `--bujo-migrated` | `280 25% 50%` | `→` (arrow) |
| `--bujo-cancelled` | `0 40% 55%` | `✗` (cross) |
| `--bujo-question` | `260 60% 55%` | `?` (question) |

#### Priority Colors
| Token | HSL Value | Symbol |
|-------|-----------|--------|
| `--priority-high` | `0 70% 55%` | `!!!` |
| `--priority-medium` | `35 85% 55%` | `!!` |
| `--priority-low` | `200 50% 55%` | `!` |

### Typography

| Element | Font | Weight | Size |
|---------|------|--------|------|
| Display/Headings | Crimson Pro (serif) | 600 | text-2xl |
| Body/UI | Inter (sans-serif) | 400 | text-sm |
| Entry Content | Inter | 400 | text-sm |
| Section Headers | Crimson Pro | 500 | text-lg |

### Spacing & Sizing

| Element | Value |
|---------|-------|
| Border radius | `0.625rem` (--radius) |
| Entry vertical padding | `py-1.5` (compact) or `py-2.5` (comfortable) |
| Entry horizontal padding | `px-2` |
| Depth indentation | `depth * 20 + 8px` |
| Section gaps | `gap-4` or `space-y-4` |
| Card padding | `p-4` or `p-6` |

### Animations

| Name | Duration | Effect |
|------|----------|--------|
| `fade-in` | 0.3s | opacity 0→1, translateY(4px→0) |
| `slide-in` | 0.2s | Entry appearance |
| `check-bounce` | 0.3s | Done confirmation |
| `streak-glow` | 2s infinite | Habit streak highlight |

---

## Component Specifications

### Sidebar

**Width:** `w-56` (224px)
**Height:** `h-screen` (sticky)

**Structure:**
```
Logo Section
├─ BookOpen icon (w-8 h-8)
├─ "bujo" wordmark (Crimson Pro, text-2xl)
└─ Tagline: "Capture. Track. Reflect."

Navigation (space-y-1)
├─ Calendar → Journal (today)
├─ CalendarDays → Weekly Review (week)
├─ Clock → Pending Tasks (overview)
├─ HelpCircle → Open Questions (questions)
├─ Flame → Habit Tracker (habits)
├─ List → Lists (lists)
├─ Target → Monthly Goals (goals)
├─ Search → Search (search)
└─ BarChart3 → Insights (stats)

Footer
└─ Settings → Settings (settings)
```

**States:**
- Default: `text-sidebar-foreground`
- Hover: `bg-sidebar-accent/50`
- Active: `bg-sidebar-accent font-medium`

### Header

**Layout:** Fixed top, full width
**Background:** `bg-card/50` with `border-b`

**Content:**
- Left: View title (Crimson Pro, text-2xl, font-semibold)
- Right: Calendar icon + formatted date (EEEE, MMMM d, yyyy)

### Entry Item

**Structure:**
```
[Collapse ▼] [Context Dot] [Symbol] [Priority] [Content] [Hidden Count] [Actions] [#ID]
```

| Element | Specification |
|---------|---------------|
| Collapse arrow | `w-4 h-4`, ChevronRight/Down, only if has children |
| Context dot | `w-1.5 h-1.5`, rounded-full, if has parent |
| Symbol | Entry type symbol, colored per type |
| Priority | `!!!`/`!!`/`!` suffix, colored per priority |
| Content | Styled per entry type (see below) |
| Hidden count | Badge showing collapsed child count |
| Actions | Hover-reveal action buttons |
| ID | `#123` shown on hover, muted |

**Content Styling by Type:**
| Type | Classes |
|------|---------|
| task | (default) |
| note | `text-muted-foreground italic` |
| event | `font-medium` |
| done | `line-through text-muted-foreground` |
| migrated | `text-muted-foreground` |
| cancelled | `line-through text-muted-foreground opacity-60` |
| question | `text-bujo-question italic` |

**Selection State:** `bg-primary/10 ring-1 ring-primary/30`

### CaptureBar

**Position:** Fixed bottom of main content area
**Padding:** Main content has `pb-24` to accommodate

**Structure:**
```
┌─────────────────────────────────────────────────────────┐
│ Input textarea (monospace font)               │ [Submit] │
│ ". Buy groceries"                             │          │
│                                               │          │
│ Adding to: [parent content] [×]               │          │
└─────────────────────────────────────────────────────────┘
```

**Typography:** Monospaced font in textarea for alignment with hierarchical entries.

**Keyboard Shortcuts:**
| Key | Action |
|-----|--------|
| `Enter` | Submit entry |
| `Shift+Enter` | New line (don't submit) |
| `Escape` | Clear input or blur |
| `i` or `a` | Focus input (global) |
| `r` | Clear parent context |

**Type Prefixes (auto-detect):**

User specifies entry type by starting with a prefix character. **The prefix is kept in the content** - it is not consumed/removed. This matches traditional bullet journal notation where the symbol is part of the entry.

| Prefix | Entry Type | Example Input | Stored Content |
|--------|------------|---------------|----------------|
| `.` | task | `. Buy milk` | `. Buy milk` |
| `-` | note | `- Remember to call mom` | `- Remember to call mom` |
| `o` | event | `o Team meeting at 3pm` | `o Team meeting at 3pm` |
| `?` | question | `? Should we refactor?` | `? Should we refactor?` |

**Priority Prefixes:**

Priority can be indicated with `!` characters after the type prefix:

| Pattern | Priority | Example |
|---------|----------|---------|
| `!` | low | `. ! Low priority task` |
| `!!` | medium | `. !! Medium priority task` |
| `!!!` | high | `. !!! High priority task` |

**Multi-line Entries:**

- Use `Shift+Enter` to add new lines within an entry
- `Enter` always submits (even with multiple lines)
- **User is responsible for indentation** in multi-line content
- Monospaced font helps user align text visually

**Parent Context:**

When adding a child entry:
- Shows "Adding to: [parent content]" above the input
- Clear button (×) removes parent context
- No automatic indentation - user controls formatting

### DateNavigator

**Position:** In Header (actions slot)
**Purpose:** Navigate between days in the journal

**Structure:**
```
[◀ Prev] [📅 Today or Date] [Next ▶] [Today button]
```

| Element | Specification |
|---------|---------------|
| Previous | Ghost button, ChevronLeft icon, `h-8 w-8` |
| Date Picker | Outline button, Calendar icon + date text, `min-w-[180px]` |
| Next | Ghost button, ChevronRight icon, `h-8 w-8` |
| Today | Secondary button, `min-w-[60px]`, invisible when viewing today |

**Date Display:**
- When viewing today: Shows "Today"
- When viewing other dates: Shows formatted date (EEE, MMM d, yyyy)

**Calendar Popover:**
- Opens on clicking the date button
- Uses shadcn Calendar component with `mode="single"`
- Selecting a date closes popover and navigates

**Navigation:**
- Prev/Next buttons move by 1 day
- Today button jumps to current date
- Today button uses `invisible` class (not `hidden`) to preserve layout space

### FileUploadButton

**Position:** In Header (after custom actions)
**Purpose:** File attachment interface for journal entries

**Structure:**
```
[Upload button with badge]
    ↓ (popover)
┌─────────────────────────────┐
│ Upload Files                │
│ ┌─────────────────────────┐ │
│ │     ⬆️ Upload icon       │ │
│ │   Click to upload       │ │
│ │   or drag and drop      │ │
│ └─────────────────────────┘ │
│                             │
│ [File 1]        [×]  1.2 MB │
│ [File 2]        [×]  3.4 KB │
│                             │
│ Storage not connected —     │
│ files will not persist      │
└─────────────────────────────┘
```

**Button:**
- Outline variant, size "sm"
- Upload icon + "Upload" text
- Badge showing file count when files selected

**Drop Zone:**
- Dashed border, rounded-lg
- Visual feedback on drag (border-primary, bg-primary/10)
- Click or drag-and-drop to add files

**File List:**
- Scrollable (`max-h-48`)
- Each file shows: icon, truncated name, formatted size, remove button
- Size formatting: B → KB → MB

**State:**
- Files tracked in memory (not persisted)
- Warning message: "Storage not connected — files will not persist"

**Note:** This component replaces the previous CaptureModal batch capture pattern.

### Context Panel

**Width:** `w-80` (320px)
**Visibility:** `hidden lg:flex`

**Content (varies by view):**

**Today View:** Overdue Items
```
Overdue (AlertTriangle icon)
├─ ScoredEntryItem (overdue task 1)
├─ ScoredEntryItem (overdue task 2)
└─ ...
```

**Other Views:** Entry Hierarchy
```
Entry Context
├─ Ancestor 1 (depth 0, muted)
├─ Ancestor 2 (depth 1, muted)
├─ Selected Entry (depth 2, highlighted ◀)
└─ "No context" message if no ancestors
```

### ScoredEntryItem

Enhanced entry display with attention scoring:

```
[Context dot] [Symbol] [Priority] [Content]
                       [Breadcrumb: parent > parent]
                       [Calendar: logged date]
                       [AlertTriangle: Overdue badge]
                       [Arrow: Migration count]       [Score Badge]
```

**Score Badge Colors:**
- High attention: Red background
- Medium: Orange background
- Low: Blue background

**Tooltip shows score breakdown:**
- "X days overdue"
- "High priority"
- "Migrated X times"
- etc.

---

## View-Specific Layouts

### Today View (Journal)

```
QuickStats (optional cards)
├─ Tasks count
├─ Events count
└─ Questions count

DayView
├─ Day Header (Today, January 25, 2026)
│  ├─ Context icons (location/weather/mood)
│  └─ Summary toggle
├─ Entry Tree (hierarchical)
└─ Empty state: "No entries yet. Start journaling!"

CaptureBar (fixed bottom)
```

**Right Sidebar:** Overdue Items section

### Week View (Weekly Review)

```
Week Navigation
├─ < Previous button
├─ "This Week" / date range
└─ Next > button

Needs Attention Section
├─ ScoredEntryItem cards
└─ Scoring based on overdue, priority, age

Meetings Section
├─ Events for the week
└─ Grouped by day
```

**Right Sidebar:** ContextPanel (entry hierarchy)

### Overview View (Pending Tasks)

```
Header + Filter Buttons
├─ All | Tasks | Events | Questions

Entries Grouped by Date
├─ Overdue (red badge)
│  └─ Entry items
├─ Today
│  └─ Entry items
├─ Tomorrow
│  └─ Entry items
└─ Future dates...
```

**Right Sidebar:** ContextPanel

### Questions View

```
Open Questions
├─ Question entries
└─ Each with answer action

Answered Questions (collapsible)
├─ Question with answer children
└─ ...
```

---

## File Reference

### Key Component Files (in bujo-canvas/src/components/bujo/)

| File | Purpose |
|------|---------|
| `Sidebar.tsx` | Navigation, view selection |
| `Header.tsx` | Page title, date display, actions slot |
| `DateNavigator.tsx` | Day navigation with calendar popover |
| `FileUploadButton.tsx` | File attachment interface (replaces CaptureModal) |
| `DayView.tsx` | Hierarchical entry display, tree builder |
| `EntryItem.tsx` | Individual entry with collapse/actions |
| `EntrySymbol.tsx` | Symbol + priority rendering |
| `CaptureBar.tsx` | Quick entry input |
| `CaptureModal.tsx` | Batch entry modal (deprecated, replaced by FileUploadButton) |
| `ContextPanel.tsx` | Entry ancestor display |
| `ScoredEntryItem.tsx` | Entry with attention scoring |
| `TodayView.tsx` | Journal view layout |
| `WeekView.tsx` | Weekly review layout |
| `OverviewView.tsx` | Pending tasks layout |
| `QuestionsView.tsx` | Questions management |
| `HabitTracker.tsx` | Habit tracking grid |
| `ListsView.tsx` | Lists management |
| `GoalsView.tsx` | Monthly goals |
| `SearchView.tsx` | Search interface |
| `StatsView.tsx` | Insights/analytics |
| `SettingsView.tsx` | Settings panel |

### Styling Files

| File | Purpose |
|------|---------|
| `src/index.css` | CSS custom properties, Tailwind layers |
| `tailwind.config.ts` | Theme extensions, animations |
| `src/types/bujo.ts` | Type definitions, symbol maps |

---

## Design Principles

1. **Paper-Inspired Aesthetic** - Warm cream colors, serif headings, subtle textures
2. **Keyboard-First** - Extensive shortcuts, Tab cycling, prefix detection
3. **Visual Hierarchy** - Entry symbols + colors communicate type at a glance
4. **Context Preservation** - Parent dots, breadcrumbs, hierarchy indentation
5. **Compact Density** - Small spacing for viewing many entries
6. **Progressive Disclosure** - Collapsible hierarchies, hover-reveal actions
7. **Attention Scoring** - Automatic prioritization based on overdue/priority/age
