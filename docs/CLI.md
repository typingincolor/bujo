# CLI Reference

Complete command reference for the bujo command-line interface.

For practical usage patterns and workflows, see [Common Workflows](WORKFLOWS.md).

## Global Flags

| Flag | Description |
|------|-------------|
| `--db-path` | Path to the database file (default: `~/.bujo/bujo.db`) |
| `-v, --verbose` | Enable verbose output |

## Entry Commands

### add

Add entries to your journal.

```bash
bujo add [entries...]
bujo add ". Buy groceries"
bujo add ". Task one" "- Note one"
echo ". Task from pipe" | bujo add
bujo add --file tasks.txt
bujo add --at "Home Office" ". Work on project"
bujo add --date yesterday ". Forgot to log this"
bujo add --parent 123 ". Add as child of entry 123"
```

| Flag | Description |
|------|-------------|
| `-a, --at` | Set location for entries |
| `-d, --date` | Date to add entries (e.g., 'yesterday', '2026-01-01') |
| `-f, --file` | Read entries from file |
| `-p, --parent` | Add entries as children of specified entry ID |
| `-y, --yes` | Skip date confirmation prompt |

**Entry types:**
- `. ` Task (todo item)
- `- ` Note (information)
- `o ` Event (scheduled occurrence)

### today

Display today's entries with overdue tasks and monthly goals.

```bash
bujo today
```

### ls

Display entries for the last 7 days.

```bash
bujo ls
bujo ls --from yesterday
bujo ls --from "last monday" --to today
bujo ls --from 2026-01-01 --to 2026-01-07
```

| Flag | Description |
|------|-------------|
| `--from` | Start date |
| `--to` | End date |

### tomorrow

Show entries scheduled for tomorrow.

```bash
bujo tomorrow
```

### next

Show entries for the upcoming 7 days (starting from tomorrow).

```bash
bujo next
```

### tasks

Show outstanding (incomplete) tasks.

```bash
bujo tasks
bujo tasks --from "last week"
bujo tasks --from 2026-01-01 --to 2026-01-31
```

| Flag | Description |
|------|-------------|
| `--from` | Start date |
| `--to` | End date |

By default shows tasks from the last 30 days.

### done

Mark an entry as complete.

```bash
bujo done <id>
bujo done 42
```

### undo

Mark a completed entry as incomplete.

```bash
bujo undo <id>
bujo undo 42
```

### edit

Edit an entry's content or priority.

```bash
bujo edit <id> [new-content]
bujo edit 42 "Buy milk instead"
bujo edit 1 --priority high
bujo edit 5 "New content" --priority medium
```

| Flag | Description |
|------|-------------|
| `-p, --priority` | Set priority (none, low, medium, high) |

### retype

Change an entry's type.

```bash
bujo retype <id> <type>
bujo retype 42 note     # Change to note
bujo retype 15 task     # Change to task
bujo retype 23 event    # Change to event
```

Valid types: `task`, `note`, `event`

### move

Move an entry to a different parent or logged date.

```bash
bujo move <id> [flags]
bujo move 42 --parent 10           # Make entry 42 a child of entry 10
bujo move 42 --root                # Make entry 42 a root entry (no parent)
bujo move 42 --logged yesterday    # Change logged date to yesterday
bujo move 42 --parent 10 --logged "last monday"
```

| Flag | Description |
|------|-------------|
| `--parent` | New parent entry ID |
| `--root` | Move entry to root (no parent) |
| `--logged` | New logged date |
| `-y, --yes` | Skip date confirmation prompt |

Unlike `migrate` (which reschedules tasks to future dates), `move` reorganizes entries within the journal.

### view

View an entry with parent and sibling context.

```bash
bujo view <id>
bujo view 42
bujo view 42 --up 1    # Show grandparent context
```

| Flag | Description |
|------|-------------|
| `-u, --up` | Number of additional ancestor levels to show |

### migrate

Migrate a task to a future date.

```bash
bujo migrate <id> --to <date>
bujo migrate 42 --to tomorrow
bujo migrate 1 --to "next monday"
bujo migrate 5 --to 2026-01-15
```

The original entry is marked as migrated (→) and a new task is created on the target date.

| Flag | Description |
|------|-------------|
| `--to` | Target date (required) |
| `-y, --yes` | Skip date confirmation prompt |

### delete

Delete an entry.

```bash
bujo delete <id>
bujo delete 42
bujo delete 1 --force
```

If the entry has children, you'll be prompted to choose how to handle them.

| Flag | Description |
|------|-------------|
| `-f, --force` | Delete without prompting (includes children) |

### cancel

Cancel an entry (mark as no longer relevant).

```bash
bujo cancel <id>
bujo cancel 42
```

Cancelled entries remain visible with strikethrough styling but are clearly marked as not active. Use this when a task becomes irrelevant rather than completed.

### uncancel

Restore a cancelled entry back to a task.

```bash
bujo uncancel <id>
bujo uncancel 42
```

### deleted

List entries that have been deleted but can still be restored.

```bash
bujo deleted
```

Each entry shows its entity ID which can be used with `bujo restore` to bring it back.

### restore

Restore a previously deleted entry.

```bash
bujo deleted              # See deleted entries and their entity IDs
bujo restore <entity-id>  # Restore entry by entity ID
```

### search

Search through entries.

```bash
bujo search <query>
bujo search "groceries"
bujo search "meeting" --from "last month"
bujo search "project" --type task
bujo search "report" -n 10
```

| Flag | Description |
|------|-------------|
| `-f, --from` | Start date for search |
| `-t, --to` | End date for search |
| `--type` | Filter by type (task, note, event, done, migrated, cancelled) |
| `-n, --limit` | Maximum number of results (default: 50) |

## Question Commands

### questions

List unanswered questions.

```bash
bujo questions
bujo questions --all
bujo questions --limit 50
```

| Flag | Description |
|------|-------------|
| `--all` | Show both answered and unanswered questions |
| `--limit` | Maximum number to show (default: 100) |

### answer

Answer a question entry.

```bash
bujo answer <id> <answer-text>
```

### reopen

Reopen an answered question.

```bash
bujo reopen <id>
```

## Habit Commands

### habit

Display habit tracker with streaks and completion rates.

```bash
bujo habit
bujo habit --month    # Calendar view
```

| Flag | Description |
|------|-------------|
| `-m, --month` | Show month calendar view |

### habit log

Log a habit completion.

```bash
bujo habit log <habit-name|#id> [count]
bujo habit log Gym
bujo habit log Water 8
bujo habit log #1              # Log by ID
bujo habit log Gym --date yesterday
bujo habit log NewHabit --yes  # Create without prompting
```

| Flag | Description |
|------|-------------|
| `-d, --date` | Date to log for |
| `-y, --yes` | Create habit without prompting if new |

### habit show

Show detailed habit statistics.

```bash
bujo habit show <habit-name|#id>
```

### habit rename

Rename a habit.

```bash
bujo habit rename <habit-name|#id> <new-name>
```

### habit delete

Delete a habit.

```bash
bujo habit delete <habit-name|#id>
```

### habit undo

Undo the last habit log for today.

```bash
bujo habit undo <habit-name|#id>
```

### habit set-goal

Set daily goal for a habit.

```bash
bujo habit set-goal <habit-name|#id> <count>
```

### habit set-weekly-goal

Set weekly goal for a habit.

```bash
bujo habit set-weekly-goal <habit-name|#id> <count>
```

### habit set-monthly-goal

Set monthly goal for a habit.

```bash
bujo habit set-monthly-goal <habit-name|#id> <count>
```

## List Commands

### list

Show all lists with progress.

```bash
bujo list
```

### list create

Create a new list.

```bash
bujo list create <name>
```

### list show

Show items in a list.

```bash
bujo list show <list-name|#id>
```

### list add

Add an item to a list.

```bash
bujo list add <list-name|#id> <content>
```

### list done

Mark a list item as done.

```bash
bujo list done <item-id>
```

### list undo

Mark a list item as not done.

```bash
bujo list undo <item-id>
```

### list move

Move an item to a different list.

```bash
bujo list move <item-id> <target-list>
```

### list remove

Remove an item from a list.

```bash
bujo list remove <item-id>
```

### list rename

Rename a list.

```bash
bujo list rename <list-name|#id> <new-name>
```

### list delete

Delete a list.

```bash
bujo list delete <list-name|#id>
```

## Goal Commands

### goal

Show goals for current month.

```bash
bujo goal
bujo goal --month 2026-02    # Specific month
```

| Flag | Description |
|------|-------------|
| `--month` | Month in YYYY-MM format |

### goal add

Add a goal to current month.

```bash
bujo goal add <content>
bujo goal add "Learn Go"
```

### goal done

Mark a goal as done.

```bash
bujo goal done #<id>
bujo goal done #1
```

### goal undo

Mark a goal as active again.

```bash
bujo goal undo #<id>
```

### goal move

Move a goal to a different month.

```bash
bujo goal move #<id> <month>
bujo goal move #1 2026-02
```

### goal delete

Delete a goal.

```bash
bujo goal delete #<id>
```

## Day Context Commands

### work

Manage work locations.

```bash
bujo work                         # Show today's location
bujo work set "Home Office"       # Set today's location
bujo work set "Office" -d monday  # Set for a past date
bujo work show --from "last week" # View location history
bujo work clear -d yesterday      # Clear a day's location
```

### mood

Manage daily mood.

```bash
bujo mood                         # Show today's mood
bujo mood set happy               # Set today's mood
bujo mood set tired -d yesterday  # Set for a past date
bujo mood show --from "last week" # View mood history
bujo mood clear -d yesterday      # Clear a day's mood
```

### weather

Manage daily weather.

```bash
bujo weather                              # Show today's weather
bujo weather set sunny                    # Set today's weather
bujo weather set "Rainy, 15°C" -d yesterday
bujo weather show --from "last week"      # View weather history
bujo weather clear -d yesterday           # Clear a day's weather
```

## AI Commands

### summary

Generate AI-powered summaries. Requires AI configuration (see [AI Setup](AI_SETUP.md)).

```bash
bujo summary           # Daily summary
bujo summary --weekly  # Weekly reflection
bujo summary --date yesterday
```

| Flag | Description |
|------|-------------|
| `--weekly` | Generate weekly reflection |
| `-d, --date` | Reference date |
| `--refresh` | Force regenerate even for completed periods |

## Statistics

### stats

Show summary statistics about journal usage.

```bash
bujo stats                       # Last 30 days
bujo stats --from "last month"
bujo stats --from "2026-01-01" --to "2026-01-31"
```

| Flag | Description |
|------|-------------|
| `-f, --from` | Start date |
| `-t, --to` | End date |

## Backup Commands

### backup

List available backups.

```bash
bujo backup
```

### backup create

Create a new backup.

```bash
bujo backup create
```

### backup verify

Verify backup integrity.

```bash
bujo backup verify <path>
```

## History Commands

### history show

Show version history for a list item.

```bash
bujo history show <entity-id>
```

### history restore

Restore a list item to a previous version.

```bash
bujo history restore <entity-id> <version>
```

## Archive Commands

### archive

Archive old data versions to reduce database size.

```bash
bujo archive                              # Dry run
bujo archive --execute                    # Actually archive
bujo archive --older-than 2026-01-01      # Archive before date
```

| Flag | Description |
|------|-------------|
| `--older-than` | Archive versions older than date (YYYY-MM-DD) |
| `--execute` | Actually perform the archive |

## Data Commands

### export

Export bujo data to JSON format for backup or migration.

```bash
bujo export > backup.json              # Export all data
bujo export --from 2026-01-01          # Export from date
bujo export --from 2026-01-01 --to 2026-01-31  # Export date range
```

| Flag | Description |
|------|-------------|
| `--from` | Start date for export (YYYY-MM-DD) |
| `--to` | End date for export (YYYY-MM-DD) |
| `--format` | Export format: `json` or `csv` (default: json) |

Export a specific entry tree to markdown:

```bash
bujo export 42                 # Export entry 42 and children as markdown
bujo export 42 -o entry.md     # Export to file
```

| Flag | Description |
|------|-------------|
| `-o, --output` | Output file (for markdown export) |

### import

Import bujo data from a JSON backup file.

```bash
bujo import backup.json                    # Merge with existing data
bujo import backup.json --mode replace     # Replace all data (destructive)
```

| Flag | Description |
|------|-------------|
| `--mode` | Import mode: `merge` (default) or `replace` |

Modes:
- `merge` - Add new records, skip if entity_id already exists
- `replace` - Clear all existing data and import fresh (destructive)

## Other Commands

### tui

Launch interactive terminal UI.

```bash
bujo tui
```

See [TUI Guide](TUI.md) for keyboard shortcuts.

### version

Show version information.

```bash
bujo version
```

### completion

Generate shell completion scripts.

```bash
source <(bujo completion bash)
source <(bujo completion zsh)
bujo completion fish | source
```

## Date Formats

Most commands accept natural language dates:

- `today`, `yesterday`, `tomorrow`
- `last monday`, `next friday`
- `2 days ago`, `3 weeks ago`
- `2026-01-15` (ISO format)

## Environment Variables

| Variable | Description |
|----------|-------------|
| `BUJO_AI_ENABLED` | Enable AI features: `true` or `false` (default: false) |
| `BUJO_AI_PROVIDER` | AI provider: `local` or `gemini` |
| `BUJO_MODEL` | Model name for local AI (default: llama3.2:3b) |
| `GEMINI_API_KEY` | API key for Google Gemini |
| `DB_PATH` | Default database path |

See [AI Setup](AI_SETUP.md) for detailed AI configuration.
