# Quick Start Guide

Get started with bujo in 5 minutes.

## Your First Entries

Add entries using the prefix symbols:

```bash
# Add a task (.)
bujo add ". Buy groceries"

# Add a note (-)
bujo add "- Meeting moved to 3pm"

# Add an event (o)
bujo add "o Team standup at 10am"

# Add multiple entries at once
bujo add ". Call mom" "- Her birthday is next week" ". Order gift"
```

## View Your Journal

```bash
# See today's entries with overdue tasks
bujo today

# See the last 7 days
bujo ls

# See outstanding tasks
bujo tasks

# See what's coming up
bujo next
```

## Complete and Manage Tasks

```bash
# Mark task #5 as done
bujo done 5

# Oops, undo that
bujo undo 5

# Move task #5 to tomorrow
bujo migrate 5 --to tomorrow

# Task is no longer relevant
bujo cancel 5
```

## Hierarchical Entries

Create nested entries using indentation:

```bash
bujo add ". Project Alpha" "  - Review requirements" "  . Write proposal"
```

Or add children to an existing entry:

```bash
bujo add --parent 5 ". Subtask one" ". Subtask two"
```

## Habit Tracking

```bash
# Log a habit (creates if new)
bujo habit log Gym

# Log multiple times (like glasses of water)
bujo habit log Water 8

# View your habits
bujo habit

# See monthly calendar view
bujo habit --month

# Set a daily goal
bujo habit set-goal Water 8
```

## Lists

Keep separate lists for different purposes:

```bash
# Create a list
bujo list create "Shopping"

# Add items
bujo list add Shopping "Milk"
bujo list add Shopping "Eggs"

# View a list
bujo list show Shopping

# Mark item done
bujo list done 1
```

## Monthly Goals

```bash
# Add a goal for this month
bujo goal add "Read 2 books"

# View current goals
bujo goal

# Mark goal complete
bujo goal done #1
```

## Interactive Mode (TUI)

Launch the terminal UI for a richer experience:

```bash
bujo tui
```

Key shortcuts:
- `a` - Add entry
- `x` - Toggle done
- `m` - Migrate task
- `/` - Search
- `?` - Show all shortcuts

## Search

```bash
# Search entries
bujo search "groceries"

# Search only tasks
bujo search "project" --type task

# Search within date range
bujo search "meeting" --from "last week"
```

## Day Context

Record context about your day:

```bash
# Set today's location
bujo work set "Home Office"

# Set mood
bujo mood set productive

# Set weather
bujo weather set "Sunny, 22C"

# View history
bujo mood show --from "last week"
```

## Next Steps

- [CLI Reference](CLI.md) - Full command documentation
- [TUI Guide](TUI.md) - Interactive terminal UI
- [Desktop App](FRONTEND.md) - Native macOS application
- [AI Setup](AI_SETUP.md) - Enable AI-powered summaries
