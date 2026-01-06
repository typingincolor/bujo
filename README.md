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
- **Daily Agenda** - View today's tasks, overdue items, and location at a glance

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
# Add some tasks for today
bujo add ". Buy groceries"
bujo add ". Finish report" "- Remember to include Q4 data"

# Or pipe multiple entries
echo ". Call dentist
- Note: morning appointments preferred
o Team standup at 10am" | bujo add

# Set your work location
bujo work "Home Office"

# View today's agenda (shows entry IDs)
bujo ls

# Mark a task complete
bujo done 1

# Log a habit
bujo habit log Gym
bujo habit log Water 8

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

### `bujo ls`

Display today's agenda including overdue tasks and current location. Each entry shows its ID for use with `bujo done`.

```
📅 Tuesday, Jan 6, 2026 | 📍 Home Office
---------------------------------------------------------
TODAY
  1 . Buy groceries
  2 . Finish report
  └──   3 - Remember to include Q4 data
  4 o Team standup at 10am
---------------------------------------------------------
```

### `bujo add [entries...]`

Add one or more entries to today's journal. Returns the ID of each entry added (one per line).

```bash
# Single entry (returns ID)
bujo add ". Call mom"
# Output: 1

# Multiple entries (each argument is one entry)
bujo add ". Task one" ". Task two"
# Output:
# 2
# 3

# Hierarchical entries via pipe
echo ". Project planning
  - Review requirements
  - Estimate timeline
  . Schedule kickoff meeting" | bujo add

# From a file
cat daily-tasks.txt | bujo add

# With location override
bujo add --at "Coffee Shop" ". Write blog post"

# Use in scripts (mark task done immediately after adding)
id=$(bujo add ". Quick task" 2>/dev/null) && bujo done $id
```

### `bujo work <location>`

Set the location context for today. This appears in your daily agenda.

```bash
bujo work "Home Office"
bujo work "Manchester Office"
bujo work "Coffee Shop"
```

Run again to change the location if you set it incorrectly.

### `bujo done <id>`

Mark a task as complete by its ID. Use `bujo ls` to see entry IDs.

```bash
bujo done 42
bujo done 15
```

### `bujo undo <id>`

Mark a completed task as incomplete. Reverses `bujo done`.

```bash
bujo undo 42
```

### `bujo view <id>`

View an entry with its parent and siblings for context. The requested entry is highlighted.

```bash
bujo view 42           # Show parent context
bujo view 42 --up 1    # Show grandparent context
bujo view 42 -u 2      # Show great-grandparent context
```

```
  2 - Requirements
  └──   3 . Review specs
  └──   4 . Write tests      <- highlighted
```

### `bujo habit`

Display the habit tracker with 7-day history, streaks, and completion rates.

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

Water (7 day streak)
  ● ● ● ● ● ● ●
  W T F S S M T
  100% completion
```

### `bujo habit log <name> [count]`

Log a habit completion. Creates the habit if it doesn't exist.

```bash
# Log once (count defaults to 1)
bujo habit log Gym

# Log with count
bujo habit log Water 8
bujo habit log "Morning Pages" 3

# Log for a different date
bujo habit log Gym --date yesterday
bujo habit log Gym -d 2026-01-05
```

### `bujo version`

Display version information.

```
bujo v0.1.0
  commit: abc1234
  built:  2026-01-06T12:00:00Z
```

## Data Storage

bujo stores all data in a SQLite database at:

```
~/.bujo/bujo.db
```

This location is the same regardless of how you install bujo (Homebrew, go install, or direct download).

To use a different database location:

```bash
bujo --db-path /path/to/custom.db ls
```

### Backup

To backup your data, simply copy the database file:

```bash
cp ~/.bujo/bujo.db ~/.bujo/bujo.db.backup
```

### Database Migrations

Migrations run automatically when bujo starts. Your data is preserved across updates.

## Configuration

### Global Flags

| Flag | Description |
|------|-------------|
| `--db-path` | Path to database file (default: `~/.bujo/bujo.db`) |
| `-v, --verbose` | Enable verbose output |

## Shell Completion

Generate completion scripts for your shell:

```bash
# Bash
bujo completion bash > /etc/bash_completion.d/bujo

# Zsh
bujo completion zsh > "${fpath[1]}/_bujo"

# Fish
bujo completion fish > ~/.config/fish/completions/bujo.fish

# PowerShell
bujo completion powershell > bujo.ps1
```

## Building from Source

```bash
git clone https://github.com/typingincolor/bujo.git
cd bujo
go build -o bujo ./cmd/bujo
```

### Running Tests

```bash
go test ./...
```

## License

MIT
