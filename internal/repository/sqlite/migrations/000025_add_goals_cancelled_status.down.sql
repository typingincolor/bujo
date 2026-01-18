-- Remove 'cancelled' from status CHECK constraint
-- SQLite requires table recreation to modify CHECK constraints
-- WARNING: This will change any 'cancelled' goals to 'active'

-- Create new table with original schema
CREATE TABLE goals_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    content TEXT NOT NULL,
    month TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'done', 'migrated')),
    migrated_to TEXT,
    created_at TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'))
);

-- Copy existing data, converting 'cancelled' to 'active'
INSERT INTO goals_new (id, entity_id, content, month, status, migrated_to, created_at, version, valid_from, valid_to, op_type)
SELECT id, entity_id, content, month,
    CASE WHEN status = 'cancelled' THEN 'active' ELSE status END,
    migrated_to, created_at, version, valid_from, valid_to, op_type FROM goals;

-- Drop old table
DROP TABLE goals;

-- Rename new table
ALTER TABLE goals_new RENAME TO goals;

-- Recreate indexes
CREATE INDEX idx_goals_entity_id ON goals(entity_id);
CREATE INDEX idx_goals_month ON goals(month);
CREATE INDEX idx_goals_status ON goals(status);
CREATE INDEX idx_goals_valid_to ON goals(valid_to);
