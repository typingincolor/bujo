# bujo

[![CI](https://github.com/typingincolor/bujo/actions/workflows/ci.yml/badge.svg)](https://github.com/typingincolor/bujo/actions/workflows/ci.yml)
[![Release](https://github.com/typingincolor/bujo/actions/workflows/release.yml/badge.svg)](https://github.com/typingincolor/bujo/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/typingincolor/bujo)](https://goreportcard.com/report/github.com/typingincolor/bujo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/v/release/typingincolor/bujo)](https://github.com/typingincolor/bujo/releases/latest)

A command-line Bullet Journal for rapid task capture, habit tracking, and daily planning.

## Features

- **Rapid Entry** - Add tasks, notes, and events with simple symbols
- **Hierarchical Notes** - Indent entries to create parent-child relationships
- **Habit Tracking** - Track daily habits with streaks and completion rates
- **Location Context** - Set your work location for the day
- **Weekly View** - See entries from the last 7 days at a glance
- **Entry Management** - Edit, delete, migrate, and reorganize entries

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
- macOS (Intel and Apple Silicon)
- Linux (amd64 and arm64)
- Windows (amd64)

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

#### `bujo today`

Display today's entries with overdue tasks and location.

```
📅 Tuesday, Jan 6, 2026 | 📍 Home Office
---------------------------------------------------------
TODAY
  1 . Buy groceries
  2 . Finish report
  └──   3 - Remember to include Q4 data
---------------------------------------------------------
```

#### `bujo view <id>`

View an entry with its parent and siblings for context.

```bash
bujo view 42           # Show parent context
bujo view 42 --up 1    # Show grandparent context
```

### Adding Entries

#### `bujo add [entries...]`

Add entries to today's journal. Returns the ID of each entry.

```bash
bujo add ". Call mom"                    # Single entry
bujo add ". Task one" ". Task two"       # Multiple entries
echo ". Task from pipe" | bujo add       # From stdin
cat tasks.txt | bujo add                 # From file
bujo add --at "Coffee Shop" ". Write"    # With location
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

#### `bujo work inspect`

Show location history.

```bash
bujo work inspect
bujo work inspect --from "last week"
```

#### `bujo work clear`

Clear location for a day.

```bash
bujo work clear
bujo work clear --date yesterday
```

### Habit Tracking

#### `bujo habit`

Display habit tracker with streaks and completion rates.

```bash
bujo habit          # 7-day sparkline view
bujo habit --month  # 30-day calendar view
```

```
🔥 Habit Tracker

Gym (3 day streak)
  ○ ○ ○ ○ ● ● ●
  W T F S S M T
  43% completion
```

#### `bujo habit log <name> [count]`

Log a habit completion. Creates the habit if it doesn't exist.

```bash
bujo habit log Gym
bujo habit log Water 8
bujo habit log Gym --date yesterday
bujo habit log #1 5              # By ID with count
```

#### `bujo habit inspect <name|#id>`

Show habit details and log history.

```bash
bujo habit inspect Gym
bujo habit inspect #1
bujo habit inspect Gym --from "last month"
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

#### `bujo habit delete-log <log-id>`

Delete a specific log entry by ID (use `habit inspect` to see IDs).

```bash
bujo habit delete-log 42
```

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

```bash
cp ~/.bujo/bujo.db ~/.bujo/bujo.db.backup
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--db-path` | Path to database file (default: `~/.bujo/bujo.db`) |
| `-v, --verbose` | Enable verbose output |

## Building from Source

```bash
git clone https://github.com/typingincolor/bujo.git
cd bujo
go build -o bujo ./cmd/bujo
go test ./...
```

## License

MIT
