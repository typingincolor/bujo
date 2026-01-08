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

1. **Create migration script**
   - Add new tables with versioning columns
   - Migrate existing data:
     - Generate entity_ids for existing records
     - Split entries with list_id into list_items table
     - Set valid_from to created_at, valid_to to NULL
     - Set op_type to 'INSERT', version to 1

2. **Drop old columns**
   - Remove list_id from entries table

### Phase 5: CLI Updates

1. **Update commands to use entity_id**
   - Display short entity_id prefix for user reference
   - Accept entity_id or row_id for backwards compatibility

2. **Add history commands** (optional)
   - `bujo history <entity-id>` - show version history
   - `bujo snapshot --date <date>` - point-in-time view

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
