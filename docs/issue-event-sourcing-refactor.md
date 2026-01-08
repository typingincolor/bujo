# Refactor Data Layer: Event Sourcing with Table Separation

## Summary

Implement an immutable append-only data persistence pattern (event sourcing) while restructuring the schema to separate list items from journal entries. This architectural change provides complete audit trails, enables temporal queries, and enforces cleaner domain boundaries.

## Parent Issue

This issue encompasses:
- [ ] #50 - Implement Immutable Append-Only Data Persistence Pattern
- [ ] #54 - Removing a list item does not check whether the item being removed is on a list

## Motivation

1. **Issue #54 reveals a domain boundary problem**: `ListService.RemoveItem` can delete any entry, not just list items. Rather than adding validation checks, separating tables enforces this at the schema level.

2. **Issue #50 requires restructuring all tables**: Adding versioning columns to every table is the right time to also reconsider table structure.

3. **Combined approach avoids two migrations**: Doing both changes together minimizes disruption.

## Proposed Schema

### Core Versioning Columns (all tables)

```sql
row_id INTEGER PRIMARY KEY AUTOINCREMENT,  -- Unique version identifier
entity_id TEXT NOT NULL,                    -- UUID, persistent across versions
version INTEGER NOT NULL DEFAULT 1,         -- Incremental counter
valid_from TEXT NOT NULL,                   -- ISO8601 timestamp
valid_to TEXT,                              -- NULL = current, set on supersession
op_type TEXT NOT NULL CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'))
```

### Table: `entries` (daily journal items)

```sql
CREATE TABLE entries (
    row_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    op_type TEXT NOT NULL CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE')),

    -- Business columns
    type TEXT NOT NULL CHECK (type IN ('task', 'note', 'event', 'done', 'migrated')),
    content TEXT NOT NULL,
    parent_entity_id TEXT,  -- References entity_id (not row_id)
    depth INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    scheduled_date TEXT,

    created_at TEXT NOT NULL
);

CREATE INDEX idx_entries_current ON entries(entity_id, valid_to) WHERE valid_to IS NULL;
CREATE INDEX idx_entries_scheduled ON entries(scheduled_date) WHERE valid_to IS NULL;
CREATE INDEX idx_entries_parent ON entries(parent_entity_id) WHERE valid_to IS NULL;
```

### Table: `list_items` (separate from entries)

```sql
CREATE TABLE list_items (
    row_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    op_type TEXT NOT NULL CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE')),

    -- Business columns
    list_entity_id TEXT NOT NULL,  -- References lists.entity_id
    type TEXT NOT NULL CHECK (type IN ('task', 'done')),
    content TEXT NOT NULL,

    created_at TEXT NOT NULL
);

CREATE INDEX idx_list_items_current ON list_items(entity_id, valid_to) WHERE valid_to IS NULL;
CREATE INDEX idx_list_items_list ON list_items(list_entity_id) WHERE valid_to IS NULL;
```

### Table: `lists`

```sql
CREATE TABLE lists (
    row_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    op_type TEXT NOT NULL CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE')),

    -- Business columns
    name TEXT NOT NULL,

    created_at TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_lists_name_current ON lists(name) WHERE valid_to IS NULL;
CREATE INDEX idx_lists_current ON lists(entity_id, valid_to) WHERE valid_to IS NULL;
```

### Other tables (habits, habit_logs, day_context, summaries)

Apply same versioning pattern to each.

## Implementation Plan

### Phase 1: Domain Layer Updates

1. **Add `EntityID` field to all domain types**
   - Use UUID for new entities
   - Generate on creation, immutable thereafter

2. **Create versioning types**
   ```go
   type VersionInfo struct {
       RowID     int64
       EntityID  string
       Version   int
       ValidFrom time.Time
       ValidTo   *time.Time
       OpType    string // "INSERT", "UPDATE", "DELETE"
   }
   ```

3. **Update Entry domain type**
   - Remove `ListID` field (list items now separate)
   - Add `EntityID` field
   - Change `ParentID` to `ParentEntityID`

4. **Create ListItem domain type**
   ```go
   type ListItem struct {
       VersionInfo
       ListEntityID string
       Type         EntryType  // task or done only
       Content      string
       CreatedAt    time.Time
   }
   ```

### Phase 2: Repository Layer

1. **Create base versioned repository**
   - `insertVersion()` - creates new version, sets valid_to on previous
   - `getCurrentByEntityID()` - WHERE valid_to IS NULL
   - `getHistoryByEntityID()` - all versions ordered by version
   - `getAsOf(timestamp)` - point-in-time query

2. **Refactor EntryRepository**
   - All writes create new versions
   - `Delete()` creates record with op_type='DELETE'
   - Remove `list_id` handling

3. **Create ListItemRepository** (new)
   - Dedicated repository for list items
   - Impossible to accidentally affect entries table

4. **Update ListRepository**
   - Add versioning support
   - Cascade soft-deletes to list_items

### Phase 3: Service Layer

1. **Update BujoService**
   - Use entity_id for references
   - Parent-child relationships use entity_id

2. **Update ListService**
   - Use new ListItemRepository
   - `RemoveItem` automatically scoped to list_items table
   - Issue #54 resolved by design

3. **Add HistoryService** (new)
   - Query historical states
   - Point-in-time reconstruction
   - Audit trail queries

### Phase 4: Migration

Migration is complex due to SQLite limitations (no `DROP COLUMN`) and the need to generate UUIDs in Go code. This is a **breaking migration** requiring the app to be offline during execution.

#### Migration Steps (Migration 007)

```
007_event_sourcing_refactor
├── 007a_add_entity_id_columns.up.sql
├── 007b_generate_uuids.go              (Go migration - generates UUIDs)
├── 007c_create_id_mapping.up.sql
├── 007d_add_parent_entity_id.up.sql
├── 007e_create_list_items_table.up.sql
├── 007f_migrate_list_items.up.sql
├── 007g_recreate_entries_table.up.sql  (removes list_id column)
├── 007h_add_versioning_columns.up.sql
├── 007i_initialize_versions.up.sql
└── 007j_cleanup.up.sql
```

#### Step-by-Step Details

**Step 1: Add entity_id columns (SQL)**
```sql
ALTER TABLE entries ADD COLUMN entity_id TEXT;
ALTER TABLE lists ADD COLUMN entity_id TEXT;
ALTER TABLE habits ADD COLUMN entity_id TEXT;
ALTER TABLE habit_logs ADD COLUMN entity_id TEXT;
ALTER TABLE day_context ADD COLUMN entity_id TEXT;
ALTER TABLE summaries ADD COLUMN entity_id TEXT;
```

**Step 2: Generate UUIDs (Go code required)**
```go
// SQLite has no UUID function - must use Go
rows, _ := db.Query("SELECT id FROM entries WHERE entity_id IS NULL")
for rows.Next() {
    var id int64
    rows.Scan(&id)
    uuid := uuid.New().String()
    db.Exec("UPDATE entries SET entity_id = ? WHERE id = ?", uuid, id)
}
// Repeat for all tables
```

**Step 3: Create ID mapping table (SQL)**
```sql
-- Temporary table to resolve parent_id → parent_entity_id
CREATE TABLE _id_mapping (
    table_name TEXT NOT NULL,
    old_id INTEGER NOT NULL,
    entity_id TEXT NOT NULL,
    PRIMARY KEY (table_name, old_id)
);

INSERT INTO _id_mapping (table_name, old_id, entity_id)
SELECT 'entries', id, entity_id FROM entries;

INSERT INTO _id_mapping (table_name, old_id, entity_id)
SELECT 'lists', id, entity_id FROM lists;
```

**Step 4: Add and populate parent_entity_id (SQL)**
```sql
ALTER TABLE entries ADD COLUMN parent_entity_id TEXT;

UPDATE entries
SET parent_entity_id = (
    SELECT m.entity_id
    FROM _id_mapping m
    WHERE m.table_name = 'entries' AND m.old_id = entries.parent_id
)
WHERE parent_id IS NOT NULL;
```

**Step 5: Create list_items table (SQL)**
```sql
CREATE TABLE list_items (
    row_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT',
    list_entity_id TEXT NOT NULL,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TEXT NOT NULL
);
```

**Step 6: Migrate list items (SQL)**
```sql
INSERT INTO list_items (entity_id, version, valid_from, op_type, list_entity_id, type, content, created_at)
SELECT
    e.entity_id,
    1,
    e.created_at,
    'INSERT',
    (SELECT m.entity_id FROM _id_mapping m WHERE m.table_name = 'lists' AND m.old_id = e.list_id),
    e.type,
    e.content,
    e.created_at
FROM entries e
WHERE e.list_id IS NOT NULL;

-- Remove migrated entries
DELETE FROM entries WHERE list_id IS NOT NULL;
```

**Step 7: Recreate entries table without list_id (SQL)**
```sql
-- SQLite can't DROP COLUMN, must recreate table
CREATE TABLE entries_new (
    row_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT',
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    parent_entity_id TEXT,
    depth INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    scheduled_date TEXT,
    created_at TEXT NOT NULL
);

INSERT INTO entries_new (entity_id, version, valid_from, op_type, type, content, parent_entity_id, depth, location, scheduled_date, created_at)
SELECT entity_id, 1, created_at, 'INSERT', type, content, parent_entity_id, depth, location, scheduled_date, created_at
FROM entries;

DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;

-- Recreate indexes
CREATE INDEX idx_entries_current ON entries(entity_id, valid_to) WHERE valid_to IS NULL;
CREATE INDEX idx_entries_scheduled ON entries(scheduled_date) WHERE valid_to IS NULL;
CREATE INDEX idx_entries_parent ON entries(parent_entity_id) WHERE valid_to IS NULL;
```

**Step 8: Apply versioning to remaining tables (SQL)**

Repeat the recreate-table pattern for: `lists`, `habits`, `habit_logs`, `day_context`, `summaries`.

**Step 9: Cleanup (SQL)**
```sql
DROP TABLE _id_mapping;
```

#### Migration Testing Strategy

1. **Backup before migration**: `cp bujo.db bujo.db.backup`
2. **Test on copy first**: Run migration against a copy of production data
3. **Verify counts**: Entry counts before/after should match (minus list items moved)
4. **Verify relationships**: Parent-child hierarchies intact
5. **Verify list items**: All items with list_id now in list_items table
6. **Smoke test CLI**: Basic commands work after migration

#### Rollback Strategy

If migration fails partway through:
1. Stop migration
2. Restore from backup: `cp bujo.db.backup bujo.db`
3. Fix migration script
4. Retry

There is no automatic rollback - the migration is destructive. Always backup first.

### Phase 5: CLI Updates

#### User-Facing ID Strategy

**Decision: Keep row_id for CLI, use entity_id internally**

Users will continue using integer IDs for commands:
```bash
bujo done 42        # Still works - row_id
bujo list remove 7  # Still works - row_id
```

The `entity_id` (UUID) is internal for:
- Version tracking across mutations
- Parent-child relationships (stable across table recreations)
- Future sync/merge capabilities

**Why not expose entity_id to users?**
- UUIDs are hard to type and remember
- row_id is auto-incrementing and user-friendly
- No practical benefit for CLI users to see UUIDs

#### Display Changes

```
# Before
ID  Type  Content
42  .     Buy groceries

# After (row_id shown, entity_id hidden)
ID  Type  Content
42  .     Buy groceries
```

#### Optional: History Commands

For users who want to see version history:
```bash
bujo history 42                    # Show all versions of entry 42
bujo history 42 --at "2024-01-15"  # Show state as of date
```

Output:
```
Version  Op      Date                 Content
1        INSERT  2024-01-10 09:00:00  Buy groceries
2        UPDATE  2024-01-12 14:30:00  Buy groceries and milk
3        DELETE  2024-01-15 10:00:00  (deleted)
```

## Query Patterns

### Get current state
```sql
SELECT * FROM entries WHERE valid_to IS NULL;
```

### Get entity history
```sql
SELECT * FROM entries WHERE entity_id = ? ORDER BY version;
```

### Point-in-time query
```sql
SELECT * FROM entries
WHERE entity_id = ?
  AND valid_from <= ?
  AND (valid_to IS NULL OR valid_to > ?);
```

### Soft delete
```sql
-- Step 1: Supersede current version
UPDATE entries SET valid_to = ? WHERE entity_id = ? AND valid_to IS NULL;

-- Step 2: Insert delete marker
INSERT INTO entries (entity_id, version, valid_from, valid_to, op_type, ...)
VALUES (?, (SELECT MAX(version)+1 FROM entries WHERE entity_id = ?), ?, NULL, 'DELETE', ...);
```

## Testing Strategy

1. **Unit tests for versioning logic**
   - Version incrementing
   - valid_to setting on supersession
   - Current state queries

2. **Integration tests for migration**
   - Data integrity after migration
   - list_id entries correctly split

3. **Property-based tests**
   - Any sequence of operations produces valid history
   - Point-in-time queries return consistent state

## Benefits

1. **Complete audit trail** - Every change preserved
2. **Temporal queries** - "What did this look like last week?"
3. **Safe "undo"** - Restore any previous state
4. **Domain enforcement** - List items physically separate from entries
5. **Issue #54 resolved** - Cannot delete wrong entity type by design

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Storage growth | Add archival strategy for old versions |
| Query complexity | Repository layer abstracts versioning |
| Migration data loss | Backup before migration, test thoroughly |
| Performance | Partial indexes on `valid_to IS NULL` |

## Labels

`enhancement`, `architecture`, `breaking-change`
