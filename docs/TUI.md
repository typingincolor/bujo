# TUI Guide

The Terminal User Interface provides an interactive way to manage your Bullet Journal entries.

## Launching

```bash
bujo tui
```

## Getting Started

When you first launch the TUI, you'll see the Journal view showing today's entries.

**Your first session:**

1. Press `a` to add a new entry
2. Type `. My first task` and press Enter
3. Use `j`/`k` to navigate to your new entry
4. Press `Space` to mark it done
5. Press `?` to see all available shortcuts

**Essential shortcuts to remember:**

| Key | What it does |
|-----|--------------|
| `a` | Add new entry |
| `Space` | Mark done |
| `j`/`k` | Move up/down |
| `1-7` | Switch views |
| `?` | Show help |
| `q` | Quit |

## Views

Navigate between views using number keys:

| Key | View | Description |
|-----|------|-------------|
| `1` | Journal | Daily entries with overdue tasks |
| `2` | Habits | Habit tracker with streaks |
| `3` | Lists | Collection lists with progress |
| `4` | Search | Search through entries |
| `5` | Stats | Usage statistics |
| `6` | Goals | Monthly goals |
| `7` | Settings | Configuration options |

## Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Enter` | Select / Expand |
| `Esc` | Back / Cancel |

## Journal View

### Date Navigation

| Key | Action |
|-----|--------|
| `t` | Jump to today |
| `/` | Go to specific date |
| `v` | Toggle day/week view |

### Entry Operations

| Key | Action |
|-----|--------|
| `Space` | Mark task done |
| `x` | Cancel task |
| `e` | Edit entry |
| `a` | Add new entry |
| `A` | Add child entry |
| `d` | Delete entry |
| `m` | Migrate task to future date |
| `p` | Cycle priority (none → low → medium → high) |
| `Tab` | Toggle collapse/expand |

### Entry Types

When adding entries, prefix with:
- `. ` Task (todo item)
- `- ` Note (information)
- `o ` Event (scheduled occurrence)

## Habits View

| Key | Action |
|-----|--------|
| `Space` | Log habit for today |
| `h` / `←` | Previous day |
| `l` / `→` | Next day |
| `v` | Cycle view (week → month → quarter) |
| `n` | Add new habit |

## Lists View

| Key | Action |
|-----|--------|
| `Enter` | View list items |
| `Space` | Toggle item done |
| `a` | Add new item |
| `n` | Create new list |
| `e` | Edit item |
| `d` | Delete item |

## Goals View

| Key | Action |
|-----|--------|
| `Space` | Toggle goal done |
| `a` | Add new goal |
| `e` | Edit goal |
| `d` | Delete goal |

## Search View

| Key | Action |
|-----|--------|
| Type | Enter search query |
| `Enter` | Execute search |
| `j`/`k` | Navigate results |
| `Enter` | View selected entry |

## Global Shortcuts

| Key | Action |
|-----|--------|
| `?` | Show help |
| `q` | Quit |
| `c` | Quick capture (add entry) |
| `Ctrl+P` or `:` | Command palette |

## Command Palette

Press `Ctrl+P` or `:` to open the command palette. Type to filter available commands:

- Navigation commands (go to view)
- Entry operations
- System commands (quit, help)

## View Modes

### Journal View Modes
- **Day View**: Shows entries for a single day
- **Week View**: Shows entries for the past 7 days

Toggle with `v` key.

### Habit View Modes
- **Week**: Last 7 days with daily completion
- **Month**: Last 30 days calendar view
- **Quarter**: Last 90 days overview

Cycle with `v` key in habits view.

## Capture Mode

For rapid multi-entry input, use capture mode:

1. Press `c` to enter capture mode
2. Type entries one per line with prefixes:
   ```
   . Buy groceries
   . Call mom
   - Remember her birthday is next week
   o Dentist at 2pm
   ```
3. Use indentation (2 spaces) for hierarchy:
   ```
   . Project tasks
     . Write proposal
     . Review with team
   ```
4. Press `Esc` to submit all entries

**Draft Persistence:** If the app crashes during capture, your draft is saved. You'll be prompted to restore it next time.

## Tips

1. **Quick Entry**: Press `c` from any view to quickly add an entry
2. **Keyboard Navigation**: Use vim-style `j`/`k` for fast navigation
3. **Collapse Trees**: Use `Tab` to collapse/expand nested entries
4. **Priority Management**: Press `p` to cycle through priority levels
5. **Date Jumping**: Press `/` and enter a date like "yesterday" or "2026-01-15"
6. **Incremental Search**: Press `Ctrl+S` to search forward, `Ctrl+R` to search backward
