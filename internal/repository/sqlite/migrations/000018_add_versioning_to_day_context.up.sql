-- Add versioning columns to day_context table for event sourcing
-- Need to recreate table since date is currently PRIMARY KEY

CREATE TABLE day_context_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    location TEXT,
    mood TEXT,
    weather TEXT,
    entity_id TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'))
);

INSERT INTO day_context_new (date, location, mood, weather)
SELECT date, location, mood, weather FROM day_context;

DROP TABLE day_context;
ALTER TABLE day_context_new RENAME TO day_context;

-- Generate entity IDs for existing rows
UPDATE day_context SET entity_id = lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))) WHERE entity_id IS NULL;

-- Set valid_from for existing rows
UPDATE day_context SET valid_from = datetime('now') WHERE valid_from IS NULL;

-- Recreate indexes
CREATE INDEX idx_day_context_date ON day_context(date);
CREATE INDEX idx_day_context_entity_id ON day_context(entity_id);
CREATE INDEX idx_day_context_valid_to ON day_context(valid_to);
