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

The TUI has **8 main views**, **54 distinct actions**, **9 entry types**, and **20+ modal dialogs**. The CLI adds **65+ commands** with additional features like import/export, backup, and version history. The Frontend currently implements only **5 views** with many actions not yet available through any UI mechanism.

| Category | TUI/CLI | Frontend | Gap |
|----------|---------|----------|-----|
| Main Views | 8+ | 5 | **3+ missing** |
| TUI Actions | 54 | ~15 | **~39 missing** |
| CLI-only Features | 40+ | 0 | **40+ missing** |
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

## 10. CLI-ONLY FEATURES (Not in TUI Either)

The CLI has additional features not available in the TUI that should be considered for the Frontend.

### Data Import/Export

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo export` | Export data to JSON | ❌ **MISSING** |
| `bujo export --format csv` | Export to CSV files | ❌ **MISSING** |
| `bujo export <id> -o file.md` | Export entry subtree to Markdown | ❌ **MISSING** |
| `bujo import <file>` | Import from JSON backup | ❌ **MISSING** |
| `bujo add -f <file>` | Add entries from file | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- Export button (JSON/CSV download)
- Import dialog (file upload + merge/replace option)
- Bulk entry creation from text file or paste

### Backup Management

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo backup` | List all backups | ❌ **MISSING** |
| `bujo backup create` | Create new backup | ❌ **MISSING** |
| `bujo backup verify <path>` | Verify backup integrity | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- Backup list view in Settings
- Create backup button
- Verify/restore options

### Version History & Restore

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo deleted` | List deleted entries | ❌ **MISSING** |
| `bujo restore <entity-id>` | Restore deleted entry | ❌ **MISSING** |
| `bujo history show <id>` | View entry version history | ❌ **MISSING** |
| `bujo history restore <id> <ver>` | Restore to previous version | ❌ **MISSING** |
| `bujo archive` | Archive old data versions | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- "Trash" view showing deleted entries
- Restore button on deleted items
- Version history panel for entries
- Archive management in Settings

### Entry Operations

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo add --parent <id>` | Add entry as child of specific entry | ❌ **MISSING** |
| `bujo move <id> --parent <id>` | Reparent an entry | ❌ **MISSING** |
| `bujo move <id> --root` | Move entry to root level | ❌ **MISSING** |
| `bujo move <id> --logged <date>` | Change entry's logged date | ❌ **MISSING** |
| `bujo view <id> -u 3` | View entry with ancestor context | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- Drag-and-drop to reparent entries
- Context menu "Move to root"
- Date change option in edit modal
- Entry detail view with breadcrumb

### Question Management

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo questions` | List all unanswered questions | ❌ **MISSING** |
| `bujo questions --all` | List all questions (including answered) | ❌ **MISSING** |
| `bujo answer <id> <text>` | Answer a question | ❌ **MISSING** |
| `bujo reopen <id>` | Reopen answered question | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- Questions filter/view in sidebar or search
- Answer dialog on question entries
- Reopen button on answered questions

### Habit Management (Extended)

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo habit log <name> -d <date>` | Log habit for specific date | ❌ **MISSING** |
| `bujo habit rename <id> <name>` | Rename a habit | ❌ **MISSING** |
| `bujo habit set-goal <id> <n>` | Set daily goal | ❌ **MISSING** |
| `bujo habit set-weekly-goal <id> <n>` | Set weekly goal | ❌ **MISSING** |
| `bujo habit set-monthly-goal <id> <n>` | Set monthly goal | ❌ **MISSING** |
| `bujo habit log-delete <id> <date>` | Delete habit log for date | ❌ **MISSING** |
| `bujo habit undo <id>` | Undo last habit log | ❌ **MISSING** |
| `bujo habit show <id>` | Show habit details & history | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- Habit detail/edit modal
- Goal setting inputs (daily/weekly/monthly)
- Log history view with delete option
- Date picker for logging past days

### List Management (Extended)

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo list create <name>` | Create new list | ❌ **MISSING** |
| `bujo list delete <id>` | Delete entire list | ❌ **MISSING** |
| `bujo list rename <id> <name>` | Rename list | ❌ **MISSING** |
| `bujo list add <list> <content>` | Add item to list | ❌ **MISSING** |
| `bujo list remove <list> <item>` | Remove item from list | ❌ **MISSING** |
| `bujo list move <list> <item> <pos>` | Reorder item in list | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- "Create List" button
- List rename/delete in context menu
- Add item form per list
- Drag-and-drop reordering

### Day Context (Extended)

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo mood set <mood>` | Set mood for day | ❌ **MISSING** |
| `bujo mood show` | View mood history | ❌ **MISSING** |
| `bujo mood clear` | Clear mood for day | ❌ **MISSING** |
| `bujo weather set <weather>` | Set weather for day | ❌ **MISSING** |
| `bujo weather show` | View weather history | ❌ **MISSING** |
| `bujo work set <location>` | Set work location | ❌ **MISSING** |
| `bujo work show` | View location history | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- Day context editor in day header
- Mood/weather/location pickers
- History view in Stats or dedicated view

### Outstanding Tasks View

| CLI Command | Purpose | Frontend Status |
|-------------|---------|-----------------|
| `bujo tasks` | Show all outstanding tasks | ❌ **MISSING** |
| `bujo tasks --from <date>` | Filter by date range | ❌ **MISSING** |

**Frontend Equivalent Needed:**
- "Outstanding Tasks" view or filter
- Date range filter

---

## 11. BACKEND BINDINGS NEEDED (Updated)

### Missing Wails Bindings for Full Feature Parity:

> This list now includes bindings needed for CLI-only features.

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

### Additional Bindings for CLI-only Features:

| Function | Purpose |
|----------|---------|
| `ExportData(from, to, format)` | Export data (JSON/CSV) |
| `ImportData(data, mode)` | Import data (merge/replace) |
| `AddEntriesFromText(text, date)` | Bulk add from text/file content |
| `GetDeletedEntries()` | List deleted entries |
| `RestoreEntry(entityId)` | Restore deleted entry |
| `GetEntryHistory(entityId)` | Get version history |
| `RestoreEntryVersion(entityId, version)` | Restore specific version |
| `ListBackups()` | List available backups |
| `CreateBackup()` | Create new backup |
| `VerifyBackup(path)` | Verify backup integrity |
| `MoveEntry(id, parentId, date)` | Move/reparent entry |
| `GetQuestions(includeAnswered)` | List questions |
| `ReopenQuestion(id)` | Reopen answered question |
| `RenameHabit(id, name)` | Rename habit |
| `SetHabitGoal(id, daily, weekly, monthly)` | Set habit goals |
| `DeleteHabitLog(id, date)` | Delete specific habit log |
| `CreateList(name)` | Create new list |
| `DeleteList(id)` | Delete list |
| `RenameList(id, name)` | Rename list |
| `ReorderListItem(listId, itemId, position)` | Move item position |
| `SetMood(date, mood)` | Set daily mood |
| `SetWeather(date, weather)` | Set daily weather |
| `GetOutstandingTasks(from, to)` | Get incomplete tasks |

---

## 12. PRIORITY RECOMMENDATIONS

### Critical (Must Have)
1. **Capture Mode** - Core bullet journal workflow
2. **Question/Answer System** - Key entry types missing
3. **Search View** - Full-screen dedicated search
4. **Stats View** - Analytics and insights
5. **AI Summary** - Past date reflections

### High Priority
6. **Add/Delete Habits** - Cannot manage habits
7. **Date Navigation** - Navigate between dates
8. **Cancel/Uncancel Entries** - Entry lifecycle management
9. **Migrate Entries** - Move tasks between dates
10. **Priority System** - Set/change priority
11. **List CRUD** - Create, rename, delete lists
12. **Add/Delete List Items** - Full list item management
13. **Habit Goals** - Set daily/weekly/monthly goals
14. **Import/Export** - Data portability (JSON/CSV)

### Medium Priority
15. **Settings View** - Configuration UI
16. **Move to List** - List assignment
17. **Convert to Goal** - Task → Goal workflow
18. **Type Changing** - Retype entries
19. **Habit Period Views** - Month/Quarter views
20. **Backup Management** - Create/list/verify backups
21. **Deleted Items View** - Trash with restore
22. **Outstanding Tasks View** - Filter incomplete tasks

### Lower Priority
23. **Location/Mood/Weather** - Day context management
24. **URL Opening** - External links
25. **Undo System** - Single-level undo
26. **Expand/Collapse All** - Bulk tree management
27. **Version History** - View and restore versions
28. **Archive Management** - Clean up old versions
29. **Entry Reparenting** - Drag-and-drop tree restructuring

---

## 13. SUMMARY TABLE

| Category | Items Missing | Severity |
|----------|--------------|----------|
| Views/Screens | 3+ (Search, Stats, Settings, Trash) | 🔴 Critical |
| TUI Actions | ~39 | 🔴 Critical |
| CLI-only Features | 40+ commands | 🟠 High |
| Entry Types | 3 (Question system) | 🔴 Critical |
| UI Dialogs/Flows | 20+ | 🟠 High |
| Capture Mode | 1 | 🔴 Critical |
| AI Summary | 1 | 🟠 High |
| Habit Management | 10+ features | 🟠 High |
| List Management | 6 features | 🟠 High |
| Data Import/Export | 5 features | 🟠 High |
| Backup/Restore | 5 features | 🟡 Medium |
| Version History | 4 features | 🟡 Medium |
| Backend Bindings | 45+ | 🟠 High |

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
