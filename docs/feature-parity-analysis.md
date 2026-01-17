# Phase 6: Feature Parity Analysis - CLI, TUI, and Frontend

## Executive Summary

This analysis compares functionality across three interfaces to identify gaps for achieving TUI-Frontend parity. The CLI provides administrative/power-user features that may not all translate to a GUI, while the TUI and Frontend should offer equivalent daily-use functionality.

## Feature Comparison Matrix

### Legend
- **CLI**: Command-line interface (power users, automation)
- **TUI**: Terminal UI (keyboard-driven interactive)
- **FE**: Wails Frontend (mouse/keyboard GUI)
- ✅ = Fully implemented
- ⚠️ = Partial/limited
- ❌ = Not implemented

---

## Entry Management

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **View entries (day)** | ✅ `today` | ✅ View 1 | ✅ DayView | All support single-day view |
| **View entries (week)** | ✅ `ls` | ✅ `w` toggle | ✅ Week view | |
| **View entries (range)** | ✅ `ls --from --to` | ❌ | ❌ | CLI-only custom date ranges |
| **Hierarchical display** | ✅ `view` cmd | ✅ collapse/expand | ✅ EntryTree | |
| **Add entry** | ✅ `add` | ✅ `a/A/r` keys | ⚠️ AddEntryBar exists but **NOT WIRED** | **HIGH PRIORITY** |
| **Add child entry** | ✅ `add --parent` | ✅ `A` key | ❌ | No parent selection in FE |
| **Edit entry** | ✅ `edit` | ✅ `e` key | ❌ | No edit UI in FE |
| **Delete entry** | ✅ `delete` | ✅ `d` key | ❌ | No delete UI in FE |
| **Mark done** | ✅ `done` | ✅ `Space` | ✅ Click entry | |
| **Mark undone** | ✅ `undo` | ✅ `Space` | ✅ Click entry | |
| **Cancel entry** | ✅ `cancel` | ✅ `x` key | ❌ | No cancel in FE |
| **Uncancel entry** | ✅ `uncancel` | ✅ `X` key | ❌ | |
| **Change entry type** | ✅ (via edit) | ✅ `t` key | ❌ | Can't change type after creation |
| **Set priority** | ✅ `edit --priority` | ✅ `!` key | ❌ | No priority editing in FE |
| **Migrate task** | ✅ `migrate --to` | ✅ `m` key | ❌ | No migration in FE |
| **Move entry** | ✅ `move --parent/--root/--logged` | ❌ | ❌ | CLI-only reparenting |
| **Restore deleted** | ✅ `restore` | ❌ | ❌ | CLI-only (event sourcing) |
| **View deleted** | ✅ `deleted` | ❌ | ❌ | CLI-only |
| **Undo last action** | ❌ | ✅ `u` key | ❌ | TUI-only undo |

---

## Navigation & Search

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **Go to specific date** | ✅ via flags | ✅ `/` key | ❌ | No date picker in FE |
| **Previous/next day** | ❌ | ✅ `h/l` keys | ❌ | No day navigation in FE |
| **Search entries** | ✅ `search` | ✅ `Ctrl+S/R` + View 4 | ⚠️ UI exists, **NOT WIRED** | |
| **Search with filters** | ✅ `--type --from --to` | ⚠️ limited | ❌ | |
| **Keyboard shortcuts** | N/A | ✅ 44+ bindings | ❌ Documented but **NOT IMPLEMENTED** | **HIGH PRIORITY** |
| **Command palette** | N/A | ✅ `Ctrl+P` | ❌ | Powerful fuzzy search |

---

## Habit Tracking

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **View habits** | ✅ `habit` | ✅ View 2 | ✅ HabitTracker | |
| **7-day sparkline** | ✅ default | ✅ default | ✅ | |
| **30-day view** | ✅ `habit --month` | ✅ `w` toggle | ❌ | FE only shows 7 days |
| **90-day view** | ❌ | ✅ `w` toggle | ❌ | TUI-only quarterly view |
| **Log habit** | ✅ `habit log` | ✅ `Space` | ✅ Log button | |
| **Log habit with count** | ✅ `habit log NAME COUNT` | ❌ | ❌ | Only logs count=1 |
| **Remove log** | ✅ `habit undo` | ✅ `Backspace` | ❌ | Can't remove logs in FE |
| **Create habit** | ✅ `habit log` (auto-creates) | ✅ `a` key | ❌ | No create habit UI |
| **Delete habit** | ✅ `habit delete` | ✅ `d` key | ❌ | No delete in FE |
| **Rename habit** | ✅ `habit rename` | ❌ | ❌ | CLI-only |
| **Set goal (daily)** | ✅ `habit set-goal` | ❌ | ❌ | CLI-only |
| **Set goal (weekly)** | ✅ `habit set-weekly-goal` | ❌ | ❌ | CLI-only |
| **Set goal (monthly)** | ✅ `habit set-monthly-goal` | ❌ | ❌ | CLI-only |
| **Navigate days** | N/A | ✅ `h/l` keys | ❌ | Can't select past days |
| **Navigate periods** | N/A | ✅ `[/]` keys | ❌ | |

---

## List Management

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **View lists** | ✅ `list` | ✅ View 3 | ✅ ListsView | |
| **View list items** | ✅ `list show` | ✅ Enter | ✅ Expand list | |
| **Create list** | ✅ `list create` | ✅ `a` key | ❌ | No create list UI |
| **Delete list** | ✅ `list delete` | ❌ | ❌ | CLI-only |
| **Rename list** | ✅ `list rename` | ❌ | ❌ | CLI-only |
| **Add item** | ✅ `list add` | ✅ `a` key | ❌ | Binding exists: `AddListItem` but no UI |
| **Mark item done** | ✅ `list done` | ✅ `Space` | ✅ Click item | |
| **Mark item undone** | (via toggle) | ✅ `Space` | ✅ Click item | |
| **Edit item** | ❌ | ✅ `e` key | ❌ | |
| **Delete item** | ✅ `list remove` | ✅ `d` key | ❌ | Binding exists: `RemoveListItem` but no UI |
| **Move item to list** | ✅ `list move` | ✅ `M` key | ❌ | |

---

## Goal Management

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **View goals** | ✅ `goal` | ✅ View 6 | ✅ GoalsView | |
| **Month navigation** | ✅ `--month` flag | ✅ `h/l` keys | ✅ Chevron buttons | |
| **Create goal** | ✅ `goal add` | ✅ `a` key | ❌ | Binding exists: `CreateGoal` but no UI |
| **Mark done** | ✅ `goal done` | ✅ `Space` | ✅ Click goal | |
| **Mark active** | ✅ `goal undo` | ✅ `Space` | ✅ Click goal | |
| **Edit goal** | ❌ | ✅ `e` key | ❌ | |
| **Delete goal** | ✅ `goal delete` | ✅ `d` key | ❌ | Binding exists: `DeleteGoal` but no UI |
| **Move goal** | ✅ `goal move` | ✅ `m` key | ❌ | |
| **Convert entry to goal** | ❌ | ✅ `M` key | ❌ | TUI-only feature |

---

## Day Context

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **View location** | ✅ in output | ✅ in header | ✅ MapPin icon | Display only |
| **Set location** | ✅ `work set` | ✅ `@` key | ❌ | No set location UI |
| **View mood** | ✅ `mood show` | ✅ in display | ✅ Heart icon | Display only |
| **Set mood** | ✅ `mood set` | ❌ | ❌ | CLI-only |
| **View weather** | ✅ `weather show` | ✅ in display | ✅ Cloud icon | Display only |
| **Set weather** | ✅ `weather set` | ❌ | ❌ | CLI-only |

---

## AI & Summaries

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **Daily summary** | ✅ `summary` | ✅ `s` toggle | ❌ | No AI integration in FE |
| **Weekly summary** | ✅ `summary --weekly` | ✅ `s` toggle | ❌ | |
| **Streaming display** | N/A | ✅ | N/A | TUI shows token-by-token |

---

## Questions System

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **View questions** | ✅ `questions` | ⚠️ Inline with entries | ❌ | Questions show as entry type |
| **Answer question** | ✅ `answer` | ✅ `R` key | ❌ | |
| **Reopen question** | ✅ `reopen` | ❌ | ❌ | CLI-only |

---

## Statistics & Analytics

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **View stats** | ✅ `stats` | ✅ View 5 | ⚠️ QuickStats exists but **NOT USED** | Component defined, not rendered |
| **Outstanding tasks** | ✅ `tasks` | ⚠️ Overdue section | ⚠️ In agenda | |

---

## Data Operations (Admin)

| Feature | CLI | TUI | FE | Notes |
|---------|-----|-----|-----|-------|
| **Export data** | ✅ `export` | ❌ | ❌ | CLI-only |
| **Create backup** | ✅ `backup create` | ❌ | ❌ | CLI-only |
| **List backups** | ✅ `backup` | ❌ | ❌ | CLI-only |
| **Archive old data** | ✅ `archive` | ❌ | ❌ | CLI-only |
| **View history** | ✅ `history show` | ❌ | ❌ | CLI-only (event sourcing) |
| **Restore version** | ✅ `history restore` | ❌ | ❌ | CLI-only |

---

## Priority Features for TUI-Frontend Parity

### P0 - Critical (Broken/Missing Core Features)

1. **Wire AddEntryBar** - Component exists, binding exists, just not connected
2. **Implement keyboard shortcuts** - Component displays hints but nothing works
3. **Wire search** - Header search input exists but not functional

### P1 - High Priority (Daily Use Features)

4. **Edit entry** - Add edit modal/inline editing
5. **Delete entry** - Add delete with confirmation
6. **Add habit** - Create new habit UI
7. **Add list item** - UI for adding items to lists (binding exists)
8. **Add goal** - UI for creating goals (binding exists)
9. **Delete goal** - UI for deleting goals (binding exists)
10. **30-day habit view** - TUI has week/month/quarter toggle

### P2 - Medium Priority (Power User Features)

11. **Cancel/uncancel entry**
12. **Change entry type**
13. **Set priority**
14. **Migrate task to future date**
15. **Remove habit log**
16. **Create list**
17. **Delete list item** (binding exists: `RemoveListItem`)
18. **Set location**
19. **Go to specific date**
20. **Navigate days (prev/next)**

### P3 - Low Priority (Nice to Have)

21. **Undo last action**
22. **AI summaries**
23. **Stats view** (QuickStats component exists but not rendered)
24. **Move entry/reparent**
25. **Questions system**

---

## Wails Bindings Status

### Existing Bindings (Working)
- `GetAgenda`, `GetHabits`, `GetLists`, `GetGoals` - Read operations
- `MarkEntryDone`, `MarkEntryUndone` - Entry completion
- `LogHabit` - Habit tracking
- `MarkListItemDone`, `MarkListItemUndone` - List items
- `MarkGoalDone`, `MarkGoalActive` - Goals

### Existing Bindings (No UI)
- `AddEntry` - Exists but AddEntryBar not wired
- `AddListItem` - Binding ready, no UI
- `RemoveListItem` - Binding ready, no UI
- `CreateGoal` - Binding ready, no UI
- `DeleteGoal` - Binding ready, no UI

### Bindings Needed
- `EditEntry(id, content, priority?)` - For edit functionality
- `DeleteEntry(id)` - For delete functionality
- `CancelEntry(id)` / `UncancelEntry(id)` - For cancel state
- `MigrateEntry(id, targetDate)` - For migration
- `ChangeEntryType(id, newType)` - For type changes
- `CreateHabit(name, goal?)` - For creating habits
- `DeleteHabit(id)` - For deleting habits
- `RemoveHabitLog(habitId, date?)` - For removing logs
- `CreateList(name)` - For creating lists
- `DeleteList(id)` - For deleting lists
- `SetLocation(date, location)` - For day context
- `GetSummary(date, weekly?)` - For AI summaries
- `Search(query, filters?)` - For search functionality

---

## Recommended Phase 6 Implementation Order

### Sprint 1: Wire Existing Components
1. Connect AddEntryBar to AddEntry binding
2. Implement basic keyboard shortcuts (j/k navigation, Space toggle)
3. Wire search functionality

### Sprint 2: CRUD Completion
4. Add edit entry modal + EditEntry binding
5. Add delete entry confirmation + DeleteEntry binding
6. Add create goal UI (binding exists)
7. Add delete goal confirmation (binding exists)

### Sprint 3: List & Habit Management
8. Add create habit UI + CreateHabit binding
9. Add delete habit confirmation + DeleteHabit binding
10. Add list item input + wire AddListItem
11. Add delete list item + wire RemoveListItem

### Sprint 4: Navigation & Polish
12. Add date picker for navigation
13. Implement day navigation (prev/next)
14. Add 30-day habit view toggle
15. Render QuickStats component

### Sprint 5: Advanced Features
16. Cancel/uncancel entries
17. Priority editing
18. Task migration
19. Set location UI
20. AI summaries integration

---

## Files to Modify

### Frontend (React)
- `frontend/src/App.tsx` - Wire AddEntryBar, add keyboard handlers
- `frontend/src/components/bujo/DayView.tsx` - Edit/delete modals
- `frontend/src/components/bujo/HabitTracker.tsx` - Create/delete UI
- `frontend/src/components/bujo/ListsView.tsx` - Add item UI
- `frontend/src/components/bujo/GoalsView.tsx` - Create/delete UI
- `frontend/src/components/bujo/Header.tsx` - Wire search

### Backend (Wails Adapter)
- `internal/adapter/wails/app.go` - New bindings
- `internal/adapter/wails/app_test.go` - Tests for new bindings

---

## Verification

After implementation:
1. Run `npm test` in frontend/ - Verify component tests pass
2. Run `go test ./internal/adapter/wails/...` - Verify binding tests pass
3. Run `wails dev` - Manual testing of all features
4. Compare TUI vs Frontend side-by-side for each feature
