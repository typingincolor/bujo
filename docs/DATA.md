# Data Management

How bujo stores, backs up, and manages your journal data.

## Storage Location

bujo stores all data in a SQLite database:

```
~/.bujo/bujo.db       # Main database
~/.bujo/backups/      # Automatic backups
```

Use `--db-path` to specify a different location:

```bash
bujo --db-path ~/Documents/journal.db today
```

Or set the `DB_PATH` environment variable:

```bash
export DB_PATH=~/Documents/journal.db
```

## Event Sourcing

bujo uses event sourcing for all data modifications. This means:

- **No data is ever overwritten** - updates create new versions
- **Full audit trail** - see the history of any item
- **Safe undo** - deleted items can be restored

### How It Works

When you edit an entry:
1. The current version is marked with an end timestamp
2. A new version is created with the changes
3. Both versions remain in the database

### Viewing History

```bash
# See version history for a list item
bujo history show <entity-id>

# Restore a previous version
bujo history restore <entity-id> <version>
```

## Backup

### Creating Backups

```bash
# Create a new backup
bujo backup create

# List existing backups
bujo backup

# Verify backup integrity
bujo backup verify ~/.bujo/backups/bujo_2026-01-20_143022.db
```

Backups are stored in `~/.bujo/backups/` with timestamps.

### Automated Backups

Add to your crontab for daily backups:

```bash
crontab -e
# Add: 0 9 * * * /usr/local/bin/bujo backup create
```

Or use a launchd plist on macOS for more reliability.

## Export

### Full Export

Export all data as JSON:

```bash
bujo export > backup.json
```

### Date Range Export

```bash
# Export a specific month
bujo export --from 2026-01-01 --to 2026-01-31 > january.json
```

### Format Options

```bash
# JSON (default)
bujo export --format json > data.json

# CSV for spreadsheets
bujo export --format csv > data.csv
```

### Entry Tree Export

Export a specific entry and its children as markdown:

```bash
bujo export 42 -o project_notes.md
```

This creates a hierarchical markdown document perfect for sharing.

## Import

### Merge Import

Add data from a backup without overwriting existing entries:

```bash
bujo import backup.json
```

This mode:
- Adds new records by entity_id
- Skips records that already exist
- Safe for restoring partial backups

### Replace Import

Replace all data with imported data:

```bash
bujo import backup.json --mode replace
```

**Warning:** This deletes all existing data before importing.

## Archive

Over time, the database grows with version history. Use archive to clean old versions:

### Dry Run

See what would be archived without making changes:

```bash
bujo archive
```

### Execute Archive

```bash
bujo archive --execute
```

### Archive by Date

Archive versions older than a specific date:

```bash
bujo archive --older-than 2025-01-01 --execute
```

Archiving:
- Keeps the current version of each item
- Removes old versions before the cutoff
- Reduces database size
- Cannot be undone

## Deleted Items

### View Deleted Items

```bash
bujo deleted
```

This shows entries that have been deleted but can still be restored.

### Restore Deleted Items

```bash
bujo restore <entity-id>
```

Items are restorable because of event sourcing - the delete operation creates a new version rather than removing data.

## Database Location Best Practices

### Cloud Sync

The database works well with cloud sync services:

```bash
export DB_PATH=~/Dropbox/bujo/bujo.db
```

Note: Avoid opening the database simultaneously from multiple machines.

### Multiple Databases

Use separate databases for different contexts:

```bash
# Work journal
bujo --db-path ~/.bujo/work.db add ". Review PR"

# Personal journal
bujo --db-path ~/.bujo/personal.db add ". Call dentist"
```

### Portable Database

Keep your journal on a USB drive:

```bash
bujo --db-path /Volumes/USB/bujo.db today
```

## Database Schema

bujo uses these main tables:

| Table | Purpose |
|-------|---------|
| `entries` | Journal entries (tasks, notes, events) |
| `lists` | Named lists |
| `list_items` | Items within lists |
| `habits` | Habit definitions |
| `habit_logs` | Habit completion records |
| `day_context` | Daily location, mood, weather |
| `summaries` | Cached AI summaries |

All tables (except summaries) include event sourcing columns for version tracking.

## Troubleshooting

### Database Locked

If you see "database is locked":
- Close other applications accessing the file
- Check for stale lock files
- Ensure only one bujo instance runs at a time

### Corrupted Database

If the database becomes corrupted:
1. Restore from the most recent backup
2. Or use SQLite's recovery tools:
   ```bash
   sqlite3 ~/.bujo/bujo.db ".recover" | sqlite3 recovered.db
   ```

### Large Database

If the database grows too large:
1. Run archive to remove old versions
2. Consider exporting and starting fresh
3. Use `VACUUM` to reclaim space:
   ```bash
   sqlite3 ~/.bujo/bujo.db "VACUUM"
   ```
