-- Add versioning columns to habits table for event sourcing
-- First, we need to recreate the table without the UNIQUE constraint on name
-- because soft delete pattern requires multiple rows with same name

CREATE TABLE habits_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    goal_per_day INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    entity_id TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'))
);

INSERT INTO habits_new (id, name, goal_per_day, created_at)
SELECT id, name, goal_per_day, created_at FROM habits;

DROP TABLE habits;
ALTER TABLE habits_new RENAME TO habits;

-- Recreate indexes
CREATE INDEX idx_habits_name ON habits(name);
CREATE INDEX idx_habits_entity_id ON habits(entity_id);
CREATE INDEX idx_habits_valid_to ON habits(valid_to);
