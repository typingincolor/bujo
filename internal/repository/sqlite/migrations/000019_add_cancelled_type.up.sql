-- Add 'cancelled' to the entry type CHECK constraint
-- SQLite doesn't support ALTER TABLE to modify CHECK constraints, so we recreate the table

CREATE TABLE entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('task', 'note', 'event', 'done', 'migrated', 'cancelled')),
    content TEXT NOT NULL,
    parent_id INTEGER,
    depth INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    scheduled_date TEXT,
    created_at TEXT NOT NULL,
    entity_id TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE')),
    FOREIGN KEY (parent_id) REFERENCES entries_new(id)
);

-- Copy data
INSERT INTO entries_new (id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, valid_to, op_type)
SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, valid_to, op_type
FROM entries;

-- Drop old table and rename new one
DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;

-- Recreate indexes
CREATE INDEX idx_entries_scheduled_date ON entries(scheduled_date);
CREATE INDEX idx_entries_parent_id ON entries(parent_id);
CREATE INDEX idx_entries_entity_id ON entries(entity_id);
CREATE INDEX idx_entries_valid_to ON entries(valid_to);
