# bujo

[![CI](https://github.com/typingincolor/bujo/actions/workflows/ci.yml/badge.svg)](https://github.com/typingincolor/bujo/actions/workflows/ci.yml)
[![Release](https://github.com/typingincolor/bujo/actions/workflows/release.yml/badge.svg)](https://github.com/typingincolor/bujo/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/typingincolor/bujo/graph/badge.svg)](https://codecov.io/gh/typingincolor/bujo)
[![Go Report Card](https://goreportcard.com/badge/github.com/typingincolor/bujo?v=1)](https://goreportcard.com/report/github.com/typingincolor/bujo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/v/release/typingincolor/bujo)](https://github.com/typingincolor/bujo/releases/latest)

A command-line Bullet Journal for rapid task capture, habit tracking, and AI-powered reflections.

## Features

- **Local AI Summaries** - Generate journal summaries with complete privacy (runs offline on your machine)
- **Rapid Entry** - Add tasks, notes, and events with simple symbols
- **Hierarchical Notes** - Indent entries to create parent-child relationships
- **Habit Tracking** - Track daily habits with streaks and completion rates (7, 30, or 90-day views)
- **List Management** - Organize items in separate lists (shopping, projects, etc.)
- **Location Context** - Set your work location for the day via CLI or command palette
- **URL Detection** - Automatically open URLs from entries with a keypress
- **Mood Tracking** - Track your daily mood with history
- **Weather Tracking** - Record daily weather conditions
- **Weekly View** - See entries from the last 7 days at a glance
- **Entry Management** - Edit, delete, migrate, and reorganize entries
- **Interactive TUI** - Navigate and manage entries with keyboard shortcuts
- **Backup & Restore** - Built-in database backups with verification
- **Version History** - View and restore previous versions of list items

## Installation

### Homebrew (macOS)

```bash
brew tap typingincolor/tap
brew install bujo
```

### Go Install

```bash
go install github.com/typingincolor/bujo/cmd/bujo@latest
```

### Download Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/typingincolor/bujo/releases).

Available platforms:
- macOS (Intel and Apple Silicon) - includes local AI support
- Linux (amd64 and arm64) - includes local AI support
- Windows (amd64) - cloud AI only (Gemini API required)

### Building from Source

See [BUILD.md](BUILD.md) for detailed build instructions, including local AI setup.

## AI Setup

bujo supports two AI backends for generating journal summaries:

### Local AI (Recommended)

Run AI models entirely on your machine - complete privacy, works offline, no API costs:

```bash
# Download a model (one-time setup)
ollama pull llama3.2:1b

# Generate summaries
bujo summary daily
bujo summary weekly
```

**Available Models:**
- `tinyllama` (637 MB) - Fast, good for testing
- `llama3.2:1b` (1.3 GB) - Recommended for most users
- `llama3.2:3b` (2.0 GB) - Higher quality summaries
- `mistral:7b` (4.1 GB) - Best quality, needs more RAM

### Cloud AI (Gemini)

Use Google's Gemini API for fast cloud-based summaries:

```bash
# Set API key (see docs/AI_SETUP.md for details)
export GEMINI_API_KEY="your-key-here"

# Generate summaries
bujo summary daily
```

**Configuration:**
- AI features are disabled by default. Set `BUJO_AI_ENABLED=true` to enable
- Local AI is used by default if a model is downloaded
- Set `BUJO_AI_PROVIDER=gemini` to force Gemini
- Set `BUJO_AI_PROVIDER=local` to force local AI

See [docs/AI_SETUP.md](docs/AI_SETUP.md) for complete configuration details.

## Quick Start

```bash
# Add tasks for today
bujo add ". Buy groceries"
bujo add ". Finish report" "- Remember to include Q4 data"

# View last 7 days
bujo ls

# View today only
bujo today

# Set your work location
bujo work set "Home Office"

# Mark a task complete
bujo done 1

# Log a habit
bujo habit log Gym

# View habit tracker
bujo habit
```

## Entry Types

| Symbol | Type | Description |
|--------|------|-------------|
| `.` | Task | A todo item to be completed |
| `-` | Note | Information or observation |
| `o` | Event | A scheduled occurrence |
| `x` | Done | A completed task |
| `>` | Migrated | A task moved to another day |

## Commands

### Viewing Entries

#### `bujo ls`

Display entries for the last 7 days, including overdue tasks.

```bash
bujo ls                              # Last 7 days
bujo ls --from yesterday             # From yesterday to today
bujo ls --from "last monday" --to today
bujo ls --from 2026-01-01 --to 2026-01-07
```

**Note:** `--from` must be before or equal to `--to`.

#### `bujo today`

Display today's entries with overdue tasks and location.

```
📅 Tuesday, Jan 6, 2026 | 📍 Home Office
---------------------------------------------------------
TODAY
• Buy groceries (1)
• Finish report (2)
  └── – Remember to include Q4 data (3)
---------------------------------------------------------
```

#### `bujo tomorrow`

Display tomorrow's entries.

#### `bujo next`

Display entries for the next 7 days (today through 6 days ahead).

#### `bujo tasks`

Show outstanding tasks only (incomplete tasks, excluding notes, events, done, and migrated).

```bash
bujo tasks                             # Last 30 days
bujo tasks --from "last week"          # Custom range
bujo tasks --from 2026-01-01 --to 2026-01-31
```

#### `bujo view <id>`

View an entry with its parent and siblings for context.

```bash
bujo view 42           # Show parent context
bujo view 42 --up 1    # Show grandparent context
```

### Adding Entries

#### `bujo add [entries...]`

Add entries to today's journal.

```bash
bujo add ". Call mom"                    # Single entry
bujo add ". Task one" ". Task two"       # Multiple entries
echo ". Task from pipe" | bujo add       # From stdin
bujo add --file tasks.txt                # From file
bujo add -f tasks.txt --at "Home"        # File with location
bujo add --at "Coffee Shop" ". Write"    # With location
bujo add --date yesterday ". Backfill"   # Add to specific date
bujo add -d "last monday" ". Forgot"     # Natural language dates
```

### Completing Tasks

#### `bujo done <id>`

Mark a task as complete.

```bash
bujo done 42
```

#### `bujo undo <id>`

Mark a completed task as incomplete.

```bash
bujo undo 42
```

#### `bujo cancel <id>`

Cancel a task (mark as not doing). Shows with strikethrough styling.

```bash
bujo cancel 42
```

#### `bujo uncancel <id>`

Restore a cancelled task back to active.

```bash
bujo uncancel 42
```

### Editing Entries

#### `bujo edit <id> <new-content>`

Edit an entry's content.

```bash
bujo edit 42 "Buy milk instead"
```

#### `bujo delete <id>`

Delete an entry. Prompts if entry has children.

```bash
bujo delete 42
bujo delete 42 --force    # Skip prompt, delete with children
```

#### `bujo migrate <id> --to <date>`

Migrate a task to a future date. Original is marked as migrated.

```bash
bujo migrate 42 --to tomorrow
bujo migrate 42 --to "next monday"
bujo migrate 42 --to 2026-01-15
```

#### `bujo move <id>`

Reorganize entries (change parent or logged date).

```bash
bujo move 42 --parent 10         # Make child of entry 10
bujo move 42 --root              # Make root entry (no parent)
bujo move 42 --logged yesterday  # Change logged date
```

### Work Location

#### `bujo work`

Show today's work location.

#### `bujo work set <location>`

Set location for today (or a specific date).

```bash
bujo work set "Home Office"
bujo work set "Manchester" --date yesterday
```

#### `bujo work show`

Show location history.

```bash
bujo work show
bujo work show --from "last week"
```

#### `bujo work clear`

Clear location for a day.

```bash
bujo work clear
bujo work clear --date yesterday
```

### Mood Tracking

#### `bujo mood`

Show today's mood.

#### `bujo mood set <mood>`

Set mood for today (or a specific date).

```bash
bujo mood set happy
bujo mood set "tired but productive"
bujo mood set energetic --date yesterday
```

#### `bujo mood show`

Show mood history.

```bash
bujo mood show
bujo mood show --from "last week"
```

#### `bujo mood clear`

Clear mood for a day.

```bash
bujo mood clear
bujo mood clear --date yesterday
```

### Weather Tracking

#### `bujo weather`

Show today's weather.

#### `bujo weather set <weather>`

Set weather for today (or a specific date).

```bash
bujo weather set sunny
bujo weather set "Rainy, 15°C"
bujo weather set cloudy --date yesterday
```

#### `bujo weather show`

Show weather history.

```bash
bujo weather show
bujo weather show --from "last week"
```

#### `bujo weather clear`

Clear weather for a day.

```bash
bujo weather clear
bujo weather clear --date yesterday
```

### Habit Tracking

#### `bujo habit`

Display habit tracker with streaks and completion rates in multiple views.

```bash
bujo habit             # 7-day sparkline view
bujo habit --month     # 30-day calendar view
bujo habit --quarter   # 90-day quarterly view
```

```
🔥 Habit Tracker

Gym (3 day streak)
  ○ ○ ○ ○ ● ● ●
  W T F S S M T
  1/1 today | 43% completion
```

**Note:** In the TUI, press `w` to cycle through Week → Month → Quarter views.

#### `bujo habit log <name|#id> [count]`

Log a habit completion. If the habit doesn't exist, you'll be prompted to create it.

**Important:** To reference a habit by ID, use `#` prefix (e.g., `#1`). Without `#`, the argument is treated as a habit name.

```bash
bujo habit log Gym
bujo habit log Water 8
bujo habit log Gym --date yesterday
bujo habit log "#1" 5            # By ID with count (quote the #)
bujo habit log NewHabit --yes    # Create without prompting
```

#### `bujo habit set-goal <name|#id> <goal>`

Set the daily goal for a habit. Goals are shown in the tracker display.

```bash
bujo habit set-goal Water 8
bujo habit set-goal #1 10
```

#### `bujo habit show <name|#id>`

Show habit details and log history.

```bash
bujo habit show Gym
bujo habit show #1
bujo habit show Gym --from "last month"
```

#### `bujo habit undo <name|#id>`

Delete the most recent log for a habit.

```bash
bujo habit undo Gym
bujo habit undo #1
```

#### `bujo habit rename <old> <new>`

Rename a habit (logs are preserved).

```bash
bujo habit rename Gym Workout
bujo habit rename #1 "Morning Workout"
```

#### `bujo habit log delete <log-id>`

Delete a specific log entry by ID (use `habit show` to see IDs).

```bash
bujo habit log delete 42
```

#### `bujo habit delete <name|#id>`

Delete a habit and all its logs. Requires confirmation unless `--force` is used.

```bash
bujo habit delete Gym
bujo habit delete "#1"
bujo habit delete OldHabit --force    # Skip confirmation
```

### List Management

Lists are separate from your daily journal - useful for shopping lists, project backlogs, or any collection of items.

#### `bujo list`

Show all lists with progress.

```
Lists
---------------------------------------------------------
#1 Shopping List 1/4 done
#2 Work
```

#### `bujo list create <name>`

Create a new list. Names can include spaces if quoted.

```bash
bujo list create Shopping
bujo list create "Shopping List"
```

#### `bujo list show <list>`

Show items in a list. Reference by name or ID (`#1`). Always quote `#` IDs.

```bash
bujo list show Shopping      # By name
bujo list show "#1"          # By ID (must quote)
```

```
#1 Shopping List
---------------------------------------------------------
(1) . Buy milk
(2) . Buy bread
(3) . Remember eggs
---------------------------------------------------------
0/3 done
```

#### `bujo list add <list> <content>`

Add an item to a list. Prefix with symbol for type (default: task).

**Important:** When referencing lists by ID (`#1`), always quote the ID to prevent shell interpretation.

```bash
bujo list add Shopping "Buy milk"          # By name
bujo list add "#1" "Buy bread"             # By ID (must quote #1)
bujo list add "#1" ". Buy eggs"            # Task with explicit symbol
bujo list add "#1" -- "- Remember eggs"    # Note type (use -- before dash)
```

#### `bujo list done <item-id>`

Mark a list item as complete.

```bash
bujo list done 42
```

#### `bujo list undo <item-id>`

Mark a completed item as incomplete.

```bash
bujo list undo 42
```

#### `bujo list remove <item-id>`

Remove an item from a list.

```bash
bujo list remove 42
```

#### `bujo list move <item-id> <target-list>`

Move an item to another list.

```bash
bujo list move 42 Work
bujo list move 42 "#2"
```

#### `bujo list rename <list> <new-name>`

Rename a list.

```bash
bujo list rename Shopping Groceries
bujo list rename "#1" "New Name"
```

#### `bujo list delete <list>`

Delete a list. Requires `--force` if list has items.

```bash
bujo list delete "#1"
bujo list delete Shopping --force    # Delete with items
```

### Monthly Goals

Goals are higher-level objectives tracked by month. They help you plan longer-term and can be converted from tasks.

#### `bujo goal`

Show goals for the current month.

```bash
bujo goal                     # Current month's goals
bujo goal --month 2026-02     # February 2026 goals
```

#### `bujo goal add <content>`

Add a new goal to a month.

```bash
bujo goal add "Learn Go"
bujo goal add "Read 12 books" --month 2026-02
```

#### `bujo goal done <#id>`

Mark a goal as completed.

```bash
bujo goal done #1
bujo goal done 1
```

#### `bujo goal undo <#id>`

Mark a completed goal as active again.

```bash
bujo goal undo #1
```

#### `bujo goal move <#id> <YYYY-MM>`

Move a goal to a different month.

```bash
bujo goal move #1 2026-02
```

#### `bujo goal delete <#id>`

Delete a goal permanently.

```bash
bujo goal delete #1
```

### Interactive TUI

#### `bujo tui`

Launch an interactive terminal UI for viewing and managing entries.

```
OVERDUE
  • Urgent task (1)

Tuesday, Jan 7 | Home Office
▸ • Buy groceries (2)             ← selected
  • Finish report (3)
    – Remember Q4 data (4)

j/k: move  space: done  x: cancel  d: delete  q: quit  ?: help
```

**Keyboard shortcuts:**

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `Space` | Toggle done/undone |
| `x` | Cancel/uncancel task |
| `e` | Edit entry content |
| `a` | Add new entry (sibling) |
| `A` | Add child entry (under selected) |
| `r` | Add root entry |
| `c` | Enter capture mode (multi-entry) |
| `m` | Migrate task to future date |
| `d` | Delete entry |
| `u` | Undo last mark done/undone operation |
| `o` | Open URL from selected entry |
| `w` | Toggle day/week view (habits: cycle week/month/quarter) |
| `/` | Go to date |
| `Ctrl+P` | Open command palette |
| `Ctrl+S` | Search forward |
| `Ctrl+R` | Search reverse |
| `@` | Set location for current day |
| `s` | Toggle AI summary collapse/expand |
| `Esc` | Navigate back through view stack |
| `?` | Toggle help |
| `q` | Quit (shows confirmation dialog) |

**Habits View shortcuts:**

Press `2` to view habit tracker:

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down (select habit) |
| `k` / `↑` | Move up (select habit) |
| `h` / `←` | Move left (select earlier day) |
| `l` / `→` | Move right (select later day) |
| `Space` | Log habit for selected day |
| `w` | Cycle view mode (Week → Month → Quarter) |
| `a` | Add new habit |
| `d` | Delete selected habit |
| `h` / `←` | Navigate to previous week/month/quarter |
| `l` / `→` | Navigate to next week/month/quarter |

**List Items View shortcuts:**

When viewing items in a list (press `3` for lists, then `Enter` to view items):

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Space` | Toggle done/undone |
| `e` | Edit item content |
| `a` | Add new item |
| `M` | Move item to another list |
| `d` | Delete item |
| `Esc` | Return to lists view |

**Goals View shortcuts:**

Press `6` to view monthly goals:

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Space` | Toggle done/undone |
| `e` | Edit goal content |
| `a` | Add new goal |
| `m` | Move goal to different month |
| `M` | Convert selected task to goal (in journal view) |
| `d` | Delete goal |
| `h` / `←` | Previous month |
| `l` / `→` | Next month |

#### Capture Mode

Press `c` to enter capture mode for rapid multi-entry input. Type entries with symbols at the start of each line, using indentation for hierarchy:

```
. Task one
. Task two
  - Note under task two
  . Subtask
- General note
o Event happening
```

**Capture mode shortcuts:**

| Key | Action |
|-----|--------|
| `Ctrl+X` | Save entries and exit |
| `Esc` | Cancel (prompts if content exists) |
| `Tab` | Indent current line |
| `Shift+Tab` | Unindent current line |
| `Ctrl+S` | Search forward in content |
| `Ctrl+R` | Search reverse in content |
| Arrow keys | Navigate cursor |

**Draft persistence:** If you exit the app unexpectedly while in capture mode, your draft is saved to `~/.bujo/capture_draft.txt`. On re-entering capture mode, you'll be prompted to restore or discard the draft.

#### Undo

Press `u` to undo the last mark done/undone operation. This allows you to quickly revert an accidental toggle. Currently supports one level of undo for marking tasks complete or incomplete.

#### URL Opening

When an entry contains a URL, press `o` to automatically open it in your default browser. Works across platforms (macOS, Linux, Windows) and extracts the first URL found in the selected entry's content.

#### Search

Press `Ctrl+S` for forward search or `Ctrl+R` for reverse search. Type your query to incrementally search through entries. Press `Enter` to jump to the match, or `Esc` to cancel. Search is case-insensitive and highlights matches in the view.

**Ancestry Context:** When a search result is selected, the UI displays the parent hierarchy (e.g., "↳ Project A > Phase 1") to provide context. Deep hierarchies are automatically truncated for readability.

#### Command Palette

Press `Ctrl+P` to open the command palette for quick access to actions like:
- Set Location: Change your work location for the current day
- Quit: Exit the application
- Other commands accessible from the main view

Type to filter commands and press `Enter` to execute, or `Esc` to cancel.

### Other

#### `bujo version`

Display version information.

#### `bujo completion <shell>`

Generate shell completion scripts (bash, zsh, fish, powershell).

## Data Storage

bujo stores all data in a SQLite database at `~/.bujo/bujo.db`.

To use a different location:

```bash
bujo --db-path /path/to/custom.db ls
```

### Backup

bujo includes built-in backup functionality using SQLite's VACUUM INTO for consistent snapshots.

#### `bujo backup`

List existing backups.

```bash
bujo backup
```

#### `bujo backup create`

Create a new backup. Backups are stored in `~/.bujo/backups/` with timestamps.

```bash
bujo backup create
# Output: Backup created: /Users/you/.bujo/backups/bujo-2026-01-08-143052.db
```

#### `bujo backup verify <path>`

Verify the integrity of a backup file.

```bash
bujo backup verify ~/.bujo/backups/bujo-2026-01-08-143052.db
```

### Archive

Clean up old data versions to reduce database size. bujo uses event sourcing which keeps historical versions of changed records.

#### `bujo archive`

Show how many old versions can be archived (dry run).

```bash
bujo archive                           # Check archivable count
bujo archive --older-than 2025-01-01   # Only versions before date
```

#### `bujo archive --execute`

Actually perform the archive operation.

```bash
bujo archive --execute
bujo archive --older-than 2025-06-01 --execute
```

### History

View and restore previous versions of list items.

#### `bujo history show <entity-id>`

Display all versions of an item.

```bash
bujo history show abc123-def456-...
```

#### `bujo history restore <entity-id> <version>`

Restore an item to a previous version. Creates a new version with the old content.

```bash
bujo history restore abc123-def456-... 1
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--db-path` | Path to database file (default: `~/.bujo/bujo.db`) |
| `-v, --verbose` | Enable verbose output |

## Shell Completions

bujo supports shell completions for tab-completion of commands and flags.

### Bash

```bash
# Add to ~/.bashrc
source <(bujo completion bash)
```

### Zsh

```bash
# Add to ~/.zshrc
source <(bujo completion zsh)

# Or install to fpath
bujo completion zsh > "${fpath[1]}/_bujo"
```

### Fish

```bash
bujo completion fish | source

# Or install permanently
bujo completion fish > ~/.config/fish/completions/bujo.fish
```

### PowerShell

```powershell
bujo completion powershell | Out-String | Invoke-Expression
```

## Building from Source

```bash
git clone https://github.com/typingincolor/bujo.git
cd bujo
go build -o bujo ./cmd/bujo
go test ./...
```

## License

MIT
