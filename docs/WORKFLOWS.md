# Common Workflows

Practical patterns for using bujo effectively.

## Daily Review

Start each day by reviewing what needs attention:

```bash
# See today's entries plus overdue tasks
bujo today

# Or use the TUI for interactive review
bujo tui
```

In the TUI, press `d` to toggle the day view which shows today with overdue items highlighted.

## Weekly Planning

At the start of each week:

```bash
# Review outstanding tasks
bujo tasks

# See the upcoming week
bujo next

# Review last week
bujo ls --from "last monday" --to "last sunday"
```

Consider migrating tasks that have been sitting too long:

```bash
bujo migrate 42 --to "next monday"
```

## Capture Mode (TUI)

For rapid multi-entry capture, use the TUI's capture mode:

```bash
bujo tui
# Press 'c' to enter capture mode
```

In capture mode:
- Type entries one per line with prefixes (`. task`, `- note`)
- Use indentation for hierarchy
- Press `Esc` to submit all entries at once

## Project Tracking

Use hierarchical entries to track project tasks:

```bash
# Add a project with tasks
bujo add ". Project Alpha"
# Note the ID returned (e.g., 42)

bujo add --parent 42 ". Design phase"
bujo add --parent 42 ". Implementation"
bujo add --parent 42 ". Testing"
```

View the project tree:

```bash
bujo view 42
```

## Habit Streaks

Build consistent habits:

```bash
# Log habits daily
bujo habit log Gym
bujo habit log Reading
bujo habit log Meditation

# Check your progress
bujo habit

# View monthly calendar
bujo habit --month

# Set goals for motivation
bujo habit set-goal Gym 1           # Daily goal
bujo habit set-weekly-goal Gym 5    # Weekly goal
```

## Shopping Lists

Keep a persistent shopping list:

```bash
# Create the list once
bujo list create "Groceries"

# Add items as you think of them
bujo list add Groceries "Milk"
bujo list add Groceries "Bread"
bujo list add Groceries "Eggs"

# At the store, view and check off
bujo list show Groceries
bujo list done 1
bujo list done 2
```

## Monthly Goals

Track monthly objectives:

```bash
# At month start
bujo goal add "Read 2 books"
bujo goal add "Complete Go course"
bujo goal add "Run 50km total"

# Throughout the month
bujo goal                  # View progress
bujo goal done #1          # Mark complete

# At month end, carry over incomplete goals
bujo goal move #2 2026-02  # Move to next month
```

## Question Tracking

Capture questions as you work and answer them later:

```bash
# Add a question
bujo add "? What's the API rate limit"

# Later, when you find the answer
bujo answer 15 "100 requests per minute, documented in API.md"

# View unanswered questions
bujo questions
```

## Searching History

Find past entries efficiently:

```bash
# General search
bujo search "meeting"

# Search by type
bujo search "project" --type task
bujo search "decision" --type note

# Search within date range
bujo search "bug" --from "last month" --to "last week"

# Limit results
bujo search "todo" -n 20
```

## Daily Context

Record context to enrich your journal:

```bash
# Set location (useful for reviewing patterns)
bujo work set "Home Office"
bujo work set "Coffee Shop"

# Record mood (track well-being over time)
bujo mood set focused
bujo mood set tired

# Note weather
bujo weather set "Rainy, 15C"

# Review patterns
bujo mood show --from "last week"
bujo work show --from "last month"
```

## Backup Strategy

Protect your journal data:

```bash
# Create regular backups
bujo backup create

# List existing backups
bujo backup

# Verify a backup's integrity
bujo backup verify ~/.bujo/backups/bujo_2026-01-20.db
```

For automated backups, add to your crontab:

```bash
0 9 * * * /usr/local/bin/bujo backup create
```

## Export for Analysis

Export your data for external analysis:

```bash
# Export all data as JSON
bujo export > journal_backup.json

# Export specific date range
bujo export --from 2026-01-01 --to 2026-01-31 > january.json

# Export as CSV for spreadsheets
bujo export --format csv > journal.csv

# Export a specific entry tree as markdown
bujo export 42 -o project_notes.md
```

## AI Reflections

Generate AI-powered summaries (requires AI setup):

```bash
# Daily summary
bujo summary

# Weekly reflection
bujo summary --weekly

# Regenerate for past date
bujo summary --date yesterday --refresh
```

See [AI Setup](AI_SETUP.md) for configuration.
