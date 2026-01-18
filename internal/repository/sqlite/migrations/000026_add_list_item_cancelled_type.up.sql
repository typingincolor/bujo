-- Add 'cancelled' as a valid list item type
-- SQLite doesn't support modifying CHECK constraints, so we recreate the table

PRAGMA foreign_keys=off;

-- Create new table with updated constraint
CREATE TABLE list_items_new (
    row_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE')),
    list_entity_id TEXT NOT NULL,
    list_id INTEGER,
    type TEXT NOT NULL CHECK (type IN ('task', 'done', 'cancelled')),
    content TEXT NOT NULL,
    created_at TEXT NOT NULL
);

-- Copy data
INSERT INTO list_items_new SELECT * FROM list_items;

-- Drop old table
DROP TABLE list_items;

-- Rename new table
ALTER TABLE list_items_new RENAME TO list_items;

-- Recreate indexes
CREATE INDEX idx_list_items_current ON list_items(entity_id) WHERE valid_to IS NULL;
CREATE INDEX idx_list_items_list ON list_items(list_entity_id) WHERE valid_to IS NULL;
CREATE INDEX idx_list_items_list_id ON list_items(list_id) WHERE valid_to IS NULL;

PRAGMA foreign_keys=on;
