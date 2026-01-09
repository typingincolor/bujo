-- Revert to original schema without versioning columns but keep UNIQUE constraint removal
-- Note: We're deliberately NOT restoring the UNIQUE constraint as it was not ideal
CREATE TABLE lists_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    entity_id TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO lists_new (id, name, entity_id, created_at)
SELECT id, name, entity_id, created_at FROM lists WHERE valid_to IS NULL OR valid_to = '';

DROP TABLE lists;
ALTER TABLE lists_new RENAME TO lists;

CREATE INDEX idx_lists_entity_id ON lists(entity_id);
