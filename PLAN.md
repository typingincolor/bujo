# TUI Phase 4: Habits, Lists, Command Palette & Theming

Issue: #43

## Overview

Add habit tracking view, lists view, command palette for discoverability, and theme customization to the TUI.

## Implementation Order

### 1. Multi-View Architecture

**Goal:** Establish the foundation for switching between views.

- Add `ViewType` enum: `ViewTypeJournal`, `ViewTypeHabits`, `ViewTypeLists`, `ViewTypeListItems`
- Add `currentView ViewType` to Model
- Update keymap with view switching keys (`1`, `2`, `3`, `h`, `L`)
- Add status bar showing current view
- Refactor `View()` to dispatch to view-specific renderers
- Refactor `Update()` to dispatch to view-specific handlers

### 2. Habits View

**Goal:** Display and interact with habits from the TUI.

- Create `habitState` struct (selected habit index, view mode)
- Add `HabitService` to Model
- Implement `renderHabitsView()` - 7-day sparkline display
- Implement habit navigation (j/k)
- Add `l` to log selected habit
- Add `--month` toggle for 30-day view
- Show streak and completion stats

### 3. Lists View

**Goal:** Display and manage lists from the TUI.

- Create `listState` struct (selected list index, selected item index, viewing items bool)
- Add `ListService` to Model
- Implement `renderListsView()` - show all lists with progress
- Implement `renderListItemsView()` - show items in selected list
- Navigation:
  - `j/k` - navigate lists or items
  - `Enter` - view list items
  - `Esc` - return to lists overview
- Item management:
  - `a` - add new item to list
  - `Space` or `x` - mark item done/undone
  - `d` - remove item from list
  - `m` - move item to different list (show list picker)

### 4. Command Palette

**Goal:** Fuzzy-searchable command interface.

- Create `commandPaletteState` struct (active, query, filtered commands, selected index)
- Define `Command` struct (name, description, keybinding, action)
- Build command registry from keymap
- Implement fuzzy matching algorithm
- Implement `renderCommandPalette()` - overlay on current view
- Trigger with `Ctrl+P` or `:`
- Execute command on Enter
- Track recent commands

### 5. Theming

**Goal:** Customizable color themes.

- Define `Theme` struct with color definitions
- Create built-in themes: default, dark, light, solarized
- Add `--theme` flag to `bujo tui`
- Update all styles to use theme colors
- Theme preview (cycle with key in palette)

### 6. Configuration File

**Goal:** Persistent user preferences.

- Define config file locations: `~/.config/bujo/config.yaml`, `~/.bujo/config.yaml`
- Define `Config` struct:
  - `defaultView`: journal/habits/lists
  - `dateFormat`: string
  - `theme`: string
  - `keybindings`: vim/emacs preset or custom
  - `showHelp`: bool
- Load config on TUI start
- Apply config to Model

## Key Bindings Summary

| Key | Action |
|-----|--------|
| `1` | Switch to Journal view |
| `2` | Switch to Habits view |
| `3` | Switch to Lists view |
| `h` | Switch to Habits view |
| `L` | Switch to Lists view |
| `Ctrl+P` / `:` | Open command palette |
| `l` (habits) | Log selected habit |
| `Enter` (lists) | View list items |
| `Esc` (list items) | Return to lists |
| `a` (list items) | Add item |
| `x` (list items) | Toggle done |
| `d` (list items) | Delete item |
| `m` (list items) | Move item to another list |

## Services Required

- `BujoService` - existing, for journal entries
- `HabitService` - existing, for habit operations
- `ListService` - existing, for list and item management

## Acceptance Criteria - Behavioral Tests

All acceptance criteria must have corresponding automated tests before implementation (TDD).

### Multi-View Architecture Tests

- [ ] `TestModel_ViewSwitch_Key1_SwitchesToJournal` - pressing '1' sets currentView to Journal
- [ ] `TestModel_ViewSwitch_Key2_SwitchesToHabits` - pressing '2' sets currentView to Habits
- [ ] `TestModel_ViewSwitch_Key3_SwitchesToLists` - pressing '3' sets currentView to Lists
- [ ] `TestModel_ViewSwitch_KeyH_SwitchesToHabits` - pressing 'h' sets currentView to Habits
- [ ] `TestModel_ViewSwitch_KeyL_SwitchesToLists` - pressing 'L' sets currentView to Lists
- [ ] `TestModel_View_JournalView_RendersEntries` - View() in journal mode renders entry list
- [ ] `TestModel_View_HabitsView_RendersHabits` - View() in habits mode renders habit tracker
- [ ] `TestModel_View_ListView_RendersLists` - View() in lists mode renders list overview
- [ ] `TestModel_View_StatusBar_ShowsCurrentView` - status bar displays current view name

### Habits View Tests

- [ ] `TestModel_HabitsView_RendersHabitList` - displays all habits with sparklines
- [ ] `TestModel_HabitsView_ShowsStreak` - displays streak count for each habit
- [ ] `TestModel_HabitsView_ShowsCompletionRate` - displays completion percentage
- [ ] `TestModel_HabitsView_Navigation_J_MovesDown` - 'j' increments selected habit index
- [ ] `TestModel_HabitsView_Navigation_K_MovesUp` - 'k' decrements selected habit index
- [ ] `TestModel_HabitsView_Navigation_BoundsCheck` - navigation respects list bounds
- [ ] `TestModel_HabitsView_LogHabit_L_LogsSelectedHabit` - 'l' logs the selected habit
- [ ] `TestModel_HabitsView_LogHabit_UpdatesDisplay` - logging updates sparkline immediately
- [ ] `TestModel_HabitsView_ToggleMonthView` - toggle between 7-day and 30-day view
- [ ] `TestModel_HabitsView_EmptyState` - shows helpful message when no habits exist

### Lists View Tests

- [ ] `TestModel_ListsView_RendersAllLists` - displays all lists with names
- [ ] `TestModel_ListsView_ShowsItemCount` - displays item count for each list
- [ ] `TestModel_ListsView_ShowsCompletionProgress` - displays done/total for each list
- [ ] `TestModel_ListsView_Navigation_J_MovesDown` - 'j' increments selected list index
- [ ] `TestModel_ListsView_Navigation_K_MovesUp` - 'k' decrements selected list index
- [ ] `TestModel_ListsView_Enter_ViewsListItems` - Enter opens selected list items view
- [ ] `TestModel_ListsView_EmptyState` - shows helpful message when no lists exist

### List Items View Tests

- [ ] `TestModel_ListItemsView_RendersItems` - displays all items in selected list
- [ ] `TestModel_ListItemsView_ShowsItemSymbol` - displays . for task, x for done
- [ ] `TestModel_ListItemsView_Navigation_J_MovesDown` - 'j' increments selected item
- [ ] `TestModel_ListItemsView_Navigation_K_MovesUp` - 'k' decrements selected item
- [ ] `TestModel_ListItemsView_Escape_ReturnsToLists` - Esc returns to lists overview
- [ ] `TestModel_ListItemsView_AddItem_A_OpensInput` - 'a' opens add item input
- [ ] `TestModel_ListItemsView_AddItem_Enter_AddsItem` - Enter in add mode adds item
- [ ] `TestModel_ListItemsView_AddItem_Escape_Cancels` - Esc in add mode cancels
- [ ] `TestModel_ListItemsView_ToggleDone_Space_TogglesItem` - Space toggles done state
- [ ] `TestModel_ListItemsView_ToggleDone_X_TogglesItem` - 'x' toggles done state
- [ ] `TestModel_ListItemsView_Delete_D_DeletesItem` - 'd' deletes selected item
- [ ] `TestModel_ListItemsView_Delete_WithConfirmation` - delete prompts for confirmation
- [ ] `TestModel_ListItemsView_Move_M_OpensListPicker` - 'm' opens list picker
- [ ] `TestModel_ListItemsView_Move_SelectList_MovesItem` - selecting list moves item
- [ ] `TestModel_ListItemsView_Move_Escape_Cancels` - Esc cancels move operation
- [ ] `TestModel_ListItemsView_EmptyState` - shows message when list has no items

### Command Palette Tests

- [ ] `TestModel_CommandPalette_CtrlP_Opens` - Ctrl+P opens command palette
- [ ] `TestModel_CommandPalette_Colon_Opens` - ':' opens command palette
- [ ] `TestModel_CommandPalette_Escape_Closes` - Esc closes without executing
- [ ] `TestModel_CommandPalette_RendersOverlay` - palette renders on top of current view
- [ ] `TestModel_CommandPalette_ShowsAllCommands` - initially shows all available commands
- [ ] `TestModel_CommandPalette_ShowsKeybindings` - displays keybinding for each command
- [ ] `TestModel_CommandPalette_FiltersByQuery` - typing filters command list
- [ ] `TestModel_CommandPalette_FuzzyMatch` - matches partial/fuzzy input (e.g., "dl" matches "delete")
- [ ] `TestModel_CommandPalette_Navigation_J_MovesDown` - 'j' or down arrow selects next
- [ ] `TestModel_CommandPalette_Navigation_K_MovesUp` - 'k' or up arrow selects previous
- [ ] `TestModel_CommandPalette_Enter_ExecutesCommand` - Enter executes selected command
- [ ] `TestModel_CommandPalette_ExecutionClosessPalette` - executing command closes palette
- [ ] `TestModel_CommandPalette_RecentCommandsFirst` - recently used commands appear first
- [ ] `TestModel_CommandPalette_EmptyQuery_ShowsRecent` - empty query prioritizes recent

### Theming Tests

- [ ] `TestTheme_Default_HasAllColors` - default theme defines all required colors
- [ ] `TestTheme_Dark_HasAllColors` - dark theme defines all required colors
- [ ] `TestTheme_Light_HasAllColors` - light theme defines all required colors
- [ ] `TestTheme_Solarized_HasAllColors` - solarized theme defines all required colors
- [ ] `TestTheme_ApplyTheme_UpdatesStyles` - applying theme updates lipgloss styles
- [ ] `TestModel_ThemeFlag_LoadsTheme` - --theme flag loads specified theme
- [ ] `TestModel_ThemeFlag_InvalidTheme_UsesDefault` - invalid theme falls back to default
- [ ] `TestModel_View_UsesThemeColors` - rendered output uses theme colors

### Configuration Tests

- [ ] `TestConfig_Load_FromConfigDir` - loads from ~/.config/bujo/config.yaml
- [ ] `TestConfig_Load_FromBujoDir` - loads from ~/.bujo/config.yaml
- [ ] `TestConfig_Load_PrecedenceOrder` - ~/.config takes precedence over ~/.bujo
- [ ] `TestConfig_Load_NoFile_UsesDefaults` - missing config uses default values
- [ ] `TestConfig_Load_PartialFile_MergesDefaults` - partial config merges with defaults
- [ ] `TestConfig_DefaultView_AppliedOnStart` - defaultView config sets initial view
- [ ] `TestConfig_Theme_AppliedOnStart` - theme config applies on TUI start
- [ ] `TestConfig_DateFormat_AppliedToDisplay` - dateFormat config changes date rendering
- [ ] `TestConfig_InvalidYAML_UsesDefaults` - malformed config falls back to defaults

### Integration Tests

- [ ] `TestIntegration_JournalToHabits_AndBack` - switch to habits and back preserves state
- [ ] `TestIntegration_LogHabit_PersistsToDatabase` - logging habit writes to DB
- [ ] `TestIntegration_AddListItem_PersistsToDatabase` - adding item writes to DB
- [ ] `TestIntegration_ToggleListItem_PersistsToDatabase` - toggling item writes to DB
- [ ] `TestIntegration_MoveListItem_PersistsToDatabase` - moving item writes to DB
- [ ] `TestIntegration_DeleteListItem_PersistsToDatabase` - deleting item writes to DB
- [ ] `TestIntegration_CommandPalette_ExecutesRealCommands` - palette commands work end-to-end
