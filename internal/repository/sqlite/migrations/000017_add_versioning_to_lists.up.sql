-- Add versioning columns to lists table for event sourcing
-- First, we need to recreate the table without the UNIQUE constraint on name
-- because soft delete pattern requires multiple rows with same name

CREATE TABLE lists_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    entity_id TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'))
);

INSERT INTO lists_new (id, name, entity_id, created_at)
SELECT id, name, entity_id, created_at FROM lists;

DROP TABLE lists;
ALTER TABLE lists_new RENAME TO lists;

-- Update valid_from for existing lists
UPDATE lists SET valid_from = created_at WHERE valid_from IS NULL;

-- Recreate indexes
CREATE INDEX idx_lists_entity_id ON lists(entity_id);
CREATE INDEX idx_lists_valid_to ON lists(valid_to);
