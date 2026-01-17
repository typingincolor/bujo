# TUI vs Frontend Gap Analysis Report

**Date:** 2026-01-17
**Purpose:** Comprehensive audit of TUI functionality missing from the Frontend

---

## Important Note on Scope

This report focuses on **functional parity** - ensuring every action achievable in the TUI is achievable in the Frontend through appropriate UI mechanisms (buttons, menus, context actions, modals, dropdowns, etc.).

**Keyboard shortcuts are NOT required** - the Frontend is a web application and will use different interaction patterns. The TUI key bindings are listed only to document what *actions* exist, not to mandate keyboard shortcuts.

At implementation time, each feature needs discussion on the appropriate UI pattern for the web context.

---

## Executive Summary

The TUI has **8 main views**, **54 distinct actions**, **9 entry types**, and **20+ modal dialogs**. The Frontend currently implements only **5 views** with many actions not yet available through any UI mechanism.

| Category | TUI | Frontend | Gap |
|----------|-----|----------|-----|
| Main Views | 8 | 5 | **3 missing** |
| Distinct Actions | 54 | ~15 | **~39 missing** |
| Entry Types | 9 | 6 | **3 missing** |
| UI Dialogs/Flows | 20+ | 3 | **17+ missing** |

---

## 1. MISSING SCREENS/VIEWS

### ❌ Search View (MISSING)
**TUI:** Dedicated full-screen search view activated by `4` key
- Real-time search input
- Direction indicator (forward/reverse)
- Results with date, type symbol, content, ID
- Color-coded results (done, migrated, cancelled)
- Shows ancestry chain (up to 3 ancestors)
- Navigate results with j/k
- Select result to jump to context

**Frontend:** Only has a search bar in the header with dropdown results
- No dedicated search view
- No sidebar navigation entry
- Cannot navigate search results with keyboard
- No ancestry context shown
- Cannot jump to entry in context

---

### ❌ Stats View (MISSING)
**TUI:** Dedicated statistics view activated by `5` key
- Entry counts: total, tasks (%), notes (%), events (%), completed (%)
- Task completion rate: X% (Y completed/Z total)
- Average entries per day
- Most productive day of week
- Least productive day of week
- Habit stats: active count, best streak, most logged

**Frontend:** `QuickStats.tsx` component exists but is **NOT integrated** into any view
- No sidebar navigation entry for Stats
- Component shows basic stats but unused

---

### ❌ Settings View (MISSING)
**TUI:** Settings view activated by `7` key
- Shows current theme
- Shows default view
- Shows date format
- Instructions to edit config file

**Frontend:** Settings button exists in sidebar but is a **non-functional placeholder**
- No settings screen
- No configuration options

---

## 2. MISSING SIDEBAR NAVIGATION ENTRIES

| Screen | TUI Key | Sidebar Entry | Status |
|--------|---------|---------------|--------|
| Journal/Today | `1` | ✅ "Today" | Present |
| Weekly View | `1` + `w` | ✅ "This Week" | Present |
| Habits | `2` | ✅ "Habits" | Present |
| Lists | `3` | ✅ "Lists" | Present |
| **Search** | `4` | ❌ **MISSING** | **Add** |
| **Stats** | `5` | ❌ **MISSING** | **Add** |
| Goals | `6` | ✅ "Goals" | Present |
| **Settings** | `7` | ⚠️ Button only, no view | **Add view** |

---

## 3. MISSING ACTIONS BY CATEGORY

> **Note:** TUI key bindings shown for reference only. Frontend needs equivalent UI mechanisms (buttons, menus, etc.), not necessarily keyboard shortcuts.

### Navigation & Date Control
| TUI Key | Action | Frontend Status | Suggested UI |
|---------|--------|-----------------|--------------|
| `h`/`←` | Previous day/week | ❌ **MISSING** | Arrow buttons, date picker |
| `l`/`→` | Next day/week | ❌ **MISSING** | Arrow buttons, date picker |
| `/` | Go to specific date | ❌ **MISSING** | Date picker, calendar widget |
| `[` | Previous habit period | ❌ **MISSING** | Period selector dropdown |
| `]` | Next habit period | ❌ **MISSING** | Period selector dropdown |

### Entry Management
| TUI Key | Action | Frontend Status | Suggested UI |
|---------|--------|-----------------|--------------|
| `x` | Cancel entry (⊗) | ❌ **MISSING** | Context menu or button |
| `X` | Uncancel entry | ❌ **MISSING** | Context menu or button |
| `a` | Add sibling entry | ❌ **MISSING** | + button, inline form |
| `A` | Add child entry | ❌ **MISSING** | Indent button, context menu |
| `m` | Migrate task to date | ❌ **MISSING** | Context menu + date picker |
| `M` | Convert task to goal | ❌ **MISSING** | Context menu + month picker |
| `L` | Move entry to list | ❌ **MISSING** | Context menu + list picker |
| `!` | Set priority | ❌ **MISSING** | Priority dropdown/selector |
| `t` | Change entry type | ❌ **MISSING** | Type dropdown/selector |
| `u` | Undo last action | ❌ **MISSING** | Undo button, toast action |
| `R` | Answer question | ❌ **MISSING** | Reply button on questions |
| `c` | **Capture mode** | ❌ **CRITICAL** | Multi-line modal/editor |
| `o` | Open URL in entry | ❌ **MISSING** | Clickable links |

### Collapse/Expand
| TUI Key | Action | Frontend Status | Suggested UI |
|---------|--------|-----------------|--------------|
| `Enter` | Toggle single entry | ✅ Click chevron | Already works |
| `ctrl+e` | Expand all | ❌ **MISSING** | "Expand All" button |
| `ctrl+c` | Collapse all | ❌ **MISSING** | "Collapse All" button |

### Search
| Action | Frontend Status | Suggested UI |
|--------|-----------------|--------------|
| Full-text search | ⚠️ Header bar only | Dedicated search view |
| Jump to search result | ❌ **MISSING** | Clickable results |
| Show ancestry context | ❌ **MISSING** | Breadcrumb in results |

### Day Context & AI
| TUI Key | Action | Frontend Status | Suggested UI |
|---------|--------|-----------------|--------------|
| `@` | Set location | ❌ **MISSING** | Location input in header |
| `s` | Toggle AI summary | ❌ **MISSING** | Collapsible summary card |
| - | Set mood | ❌ **MISSING** | Mood selector |
| - | Set weather | ❌ **MISSING** | Weather input |

### View Switching
| Action | Frontend Status | Notes |
|--------|-----------------|-------|
| Switch views | ✅ Sidebar nav | Already works |
| Search view | ❌ **View missing** | Need sidebar entry |
| Stats view | ❌ **View missing** | Need sidebar entry |
| Settings view | ❌ **View missing** | Need sidebar entry |

---

## 4. MISSING ENTRY TYPES

| Symbol | Type | TUI | Frontend |
|--------|------|-----|----------|
| `•` | Task | ✅ | ✅ |
| `–` | Note | ✅ | ✅ |
| `○` | Event | ✅ | ✅ |
| `✓` | Done | ✅ | ✅ |
| `→` | Migrated | ✅ | ✅ |
| `✗` | Cancelled | ✅ | ⚠️ Listed but **not functional** |
| `?` | **Question** | ✅ | ❌ **MISSING** |
| `★` | **Answered** | ✅ | ❌ **MISSING** |
| `↳` | **Answer** | ✅ | ❌ **MISSING** |

### Question/Answer System (CRITICAL GAP)
The TUI has a full question/answer workflow:
1. Create question entry with `?` prefix
2. Answer questions with `R` key
3. Toggle answered state with `space`
4. Child answers shown with `↳` symbol

**Frontend has NONE of this functionality.**

---

## 5. MISSING FEATURES BY VIEW

### Journal View Gaps

| Feature | TUI | Frontend |
|---------|-----|----------|
| Day/Week toggle | `w` key | ❌ Separate views, no toggle |
| Overdue section | ⚠️ OVERDUE header | ❌ **MISSING** |
| AI Summary | Collapsible, markdown | ❌ **MISSING** |
| Monthly goals progress | Progress bar in view | ❌ **MISSING** |
| Capture mode | `c` → external editor | ❌ **CRITICAL MISSING** |
| Natural date navigation | `/` → "tomorrow", "next week" | ❌ **MISSING** |
| Location history picker | `@` with suggestions | ❌ **MISSING** |
| Entry migration | `m` → date picker | ❌ **MISSING** |
| Move to list | `L` → list picker | ❌ **MISSING** |
| Convert to goal | `M` → month picker | ❌ **MISSING** |
| Priority cycling | `!` key | ❌ **MISSING** |
| Type changing | `t` key → menu | ❌ **MISSING** |
| URL opening | `o` key | ❌ **MISSING** |
| Undo | `u` key | ❌ **MISSING** |

### Habits View Gaps

| Feature | TUI | Frontend |
|---------|-----|----------|
| View period toggle | Week/Month/Quarter | ❌ **Only 7-day view** |
| Period navigation | `[`/`]` keys | ❌ **MISSING** |
| Day navigation | `h`/`l` in sparkline | ❌ **MISSING** |
| Log for specific day | Select day + space | ❌ **Only logs today** |
| Remove log | Backspace/Delete | ❌ **MISSING** |
| Add habit | `a` key | ❌ **MISSING** |
| Delete habit | `d` key with confirm | ❌ **MISSING** |
| Keyboard navigation | j/k through habits | ❌ **MISSING** |
| Day labels | S M T W T F S | ❌ **MISSING** |
| Month labels | Quarter view with separators | ❌ **MISSING** |
| Progress stats | "Week: X%, Month: Y%" | ❌ **MISSING** |

### Lists View Gaps

| Feature | TUI | Frontend |
|---------|-----|----------|
| Create new list | CLI command | ❌ **No UI for creation** |
| Delete list | - | ❌ **MISSING** |
| Edit item | `e` key | ❌ **MISSING** |
| Add item | `a` key | ❌ **MISSING** |
| Delete item | `d` key | ❌ **MISSING** (only backend binding exists) |
| Move to another list | `M` key | ❌ **MISSING** |
| Keyboard navigation | j/k through items | ❌ **MISSING** |

### Goals View Gaps

| Feature | TUI | Frontend |
|---------|-----|----------|
| Edit goal | `e` key | ❌ **MISSING** |
| Move goal to month | `m` key → month picker | ❌ **MISSING** |
| Keyboard navigation | j/k through goals | ❌ **MISSING** |
| Goal ID display | `#1`, `#2` format | ⚠️ Shows on hover only |

---

## 6. MISSING MODAL DIALOGS

| Dialog | TUI Trigger | Frontend Status |
|--------|-------------|-----------------|
| Edit Entry | `e` | ✅ Implemented |
| Add Entry | `a`/`A`/`r` | ⚠️ Only bar, no modal |
| Delete Confirm | `d` | ✅ Implemented |
| **Answer Question** | `R` | ❌ **MISSING** |
| **Migrate Entry** | `m` | ❌ **MISSING** |
| **Go to Date** | `/` | ❌ **MISSING** |
| **Set Location** | `@` | ❌ **MISSING** |
| **Add Habit** | `a` in habits | ❌ **MISSING** |
| **Delete Habit Confirm** | `d` in habits | ❌ **MISSING** |
| **Move to List Picker** | `L` | ❌ **MISSING** |
| **Convert to Goal** | `M` | ❌ **MISSING** |
| **Edit Goal** | `e` in goals | ❌ **MISSING** |
| **Move Goal** | `m` in goals | ❌ **MISSING** |
| **Retype Entry** | `t` | ❌ **MISSING** |
| **Command Palette** | `ctrl+p`/`:` | ❌ **MISSING** |
| **Help Panel** | `?` | ❌ **MISSING** |

---

## 7. MISSING UI ELEMENTS & BEHAVIORS

### Styling Gaps
| Element | TUI | Frontend |
|---------|-----|----------|
| Strikethrough for cancelled | ✅ | ⚠️ Styled but action missing |
| Overdue entries in red | ✅ | ❌ **MISSING** |
| Search highlight (yellow) | ✅ | ❌ **MISSING** |
| Habit sparkline day selection | ✅ Inverted | ❌ **MISSING** |

### UI Indicators
| Element | TUI | Frontend |
|---------|-----|----------|
| "↑ N more above" scroll indicator | ✅ | ❌ **MISSING** |
| "↓ N more below" scroll indicator | ✅ | ❌ **MISSING** |
| `[N hidden]` collapsed count | ✅ | ❌ **MISSING** |
| Ancestry chain in search results | ✅ | ❌ **MISSING** |
| Monthly goals progress in journal | ✅ | ❌ **MISSING** |

---

## 8. CAPTURE MODE (CRITICAL GAP)

**This is identified as a CRITICAL missing feature.**

### TUI Capture Mode Features:
1. **Trigger:** `c` key opens `$EDITOR` (or `$VISUAL`, defaults to `vi`)
2. **Multi-line input:** Full editor experience for composing entries
3. **Draft recovery:** Auto-saves to `~/.bujo/capture_draft.txt`
4. **TreeParser support:** Hierarchical entries via indentation
5. **Type prefixes:** `• task`, `– note`, `o event`, `? question`
6. **Date inheritance:** Entries scheduled for current viewing date
7. **Bulk entry creation:** Multiple entries in one capture session

### Frontend Equivalent Needed:
- Multi-line text area or rich text editor
- Support for TreeParser syntax
- All entry type prefixes
- Date selector for scheduling
- Draft auto-save (localStorage)
- Preview of parsed entries before save

---

## 9. AI SUMMARY (CRITICAL GAP)

### TUI AI Summary Features:
1. Only shows when viewing past dates
2. Daily or Weekly summary based on view mode
3. Collapsible with `s` key
4. Streams tokens from Gemini API
5. Markdown rendered with glamour
6. Cached results (won't regenerate)

### Frontend Equivalent Needed:
- Summary component for past day/week views
- Collapse/expand toggle
- Loading state with streaming text
- Markdown rendering
- Cache integration

---

## 10. BACKEND BINDINGS NEEDED

### Missing Wails Bindings for Full Feature Parity:

| Function | Purpose |
|----------|---------|
| `CancelEntry(id)` | Mark entry as cancelled |
| `UncancelEntry(id)` | Restore cancelled entry |
| `MigrateEntry(id, date)` | Move entry to different date |
| `ConvertToGoal(entryId, month)` | Transform task to monthly goal |
| `MoveToList(entryId, listId)` | Move entry to a list |
| `SetPriority(id, priority)` | Set entry priority level |
| `ChangeType(id, type)` | Change entry type |
| `CreateHabit(name, goalPerDay)` | Create new habit |
| `DeleteHabit(id)` | Delete habit |
| `LogHabitForDate(id, date, count)` | Log habit for specific date |
| `RemoveHabitLog(id, date)` | Remove habit log for date |
| `GetHabitsForPeriod(period)` | Get habits with extended history |
| `CreateList(name)` | Create new list |
| `DeleteList(id)` | Delete list |
| `EditListItem(id, content)` | Edit list item content |
| `MoveListItem(id, listId)` | Move item to another list |
| `EditGoal(id, content)` | Edit goal content |
| `MoveGoal(id, month)` | Move goal to different month |
| `SetDayContext(date, location, mood, weather)` | Set daily context |
| `GetDayContext(date)` | Get daily context |
| `GetSummary(date, type)` | Get AI summary for date |
| `GetStats(startDate, endDate)` | Get statistics for date range |
| `AnswerQuestion(id, answer)` | Answer a question entry |
| `CreateQuestion(content, date)` | Create question entry |

---

## 11. PRIORITY RECOMMENDATIONS

### Critical (Must Have)
1. **Capture Mode** - Core bullet journal workflow
2. **Question/Answer System** - Key entry types missing
3. **Search View** - Full-screen dedicated search
4. **Stats View** - Analytics and insights
5. **AI Summary** - Past date reflections

### High Priority
6. **Add/Delete Habits** - Cannot manage habits
7. **Date Navigation** - h/l/`/` keys for date jumping
8. **Cancel/Uncancel Entries** - Entry lifecycle management
9. **Migrate Entries** - Move tasks between dates
10. **Priority System** - `!` cycling

### Medium Priority
11. **Settings View** - Configuration UI
12. **Move to List** - List assignment
13. **Convert to Goal** - Task → Goal workflow
14. **Type Changing** - Retype entries
15. **Habit Period Views** - Month/Quarter views
16. **Command Palette** - Power user feature

### Lower Priority
17. **Location Picker** - Day context
18. **URL Opening** - External links
19. **Undo System** - Single-level undo
20. **Expand/Collapse All** - Bulk tree management

---

## 12. SUMMARY TABLE

| Category | Items Missing | Severity |
|----------|--------------|----------|
| Views/Screens | 3 | 🔴 Critical |
| Distinct Actions | ~39 | 🔴 Critical |
| Entry Types | 3 (Question system) | 🔴 Critical |
| UI Dialogs/Flows | 17+ | 🟠 High |
| Capture Mode | 1 | 🔴 Critical |
| AI Summary | 1 | 🟠 High |
| Habit Management | 5 features | 🟠 High |
| Backend Bindings | 25+ | 🟠 High |

---

## Appendix: Complete Action Reference

> This appendix lists all TUI actions with their key bindings for reference. The "Frontend" column indicates whether the action is achievable through ANY UI mechanism (not necessarily keyboard).

| Key | TUI Function | Frontend |
|-----|--------------|----------|
| `j`/`↓` | Move down | ✅ |
| `k`/`↑` | Move up | ✅ |
| `g` | Jump to top | ❌ |
| `G` | Jump to bottom | ❌ |
| `h`/`←` | Previous day | ❌ |
| `l`/`→` | Next day | ❌ |
| `w` | Toggle day/week | ❌ |
| `[` | Previous period | ❌ |
| `]` | Next period | ❌ |
| `space` | Toggle done | ✅ |
| `x` | Cancel | ❌ |
| `X` | Uncancel | ❌ |
| `e` | Edit | ✅ |
| `a` | Add sibling | ❌ |
| `A` | Add child | ❌ |
| `r` | Add root | ❌ |
| `d` | Delete | ✅ |
| `m` | Migrate | ❌ |
| `M` | To goal | ❌ |
| `L` | To list | ❌ |
| `!` | Priority | ❌ |
| `t` | Retype | ❌ |
| `u` | Undo | ❌ |
| `R` | Answer | ❌ |
| `Enter` | Toggle collapse | ✅ (click) |
| `ctrl+e` | Expand all | ❌ |
| `ctrl+c` | Collapse all | ❌ |
| `C` | Overdue context | ❌ |
| `ctrl+s` | Search forward | ❌ |
| `ctrl+r` | Search reverse | ❌ |
| `/` | Go to date | ❌ |
| `o` | Open URL | ❌ |
| `@` | Set location | ❌ |
| `c` | Capture | ❌ |
| `s` | Toggle summary | ❌ |
| `1` | Journal view | ❌ |
| `2` | Habits view | ❌ |
| `3` | Lists view | ❌ |
| `4` | Search view | ❌ |
| `5` | Stats view | ❌ |
| `6` | Goals view | ❌ |
| `7` | Settings view | ❌ |
| `ctrl+p`/`:` | Command palette | ❌ |
| `?` | Help | ❌ |
| `esc` | Back/cancel | ⚠️ Modal only |
| `q` | Quit | N/A |

---

*Report generated by comprehensive TUI/Frontend audit*
