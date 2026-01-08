-- SQLite doesn't support DROP COLUMN directly, so we recreate the table
-- This is a destructive migration - use with caution

DROP INDEX IF EXISTS idx_entries_valid_to;
DROP INDEX IF EXISTS idx_entries_entity_id;

CREATE TABLE entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    parent_id INTEGER,
    depth INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    scheduled_date TEXT,
    created_at TEXT NOT NULL,
    list_id INTEGER REFERENCES lists(id),
    FOREIGN KEY (parent_id) REFERENCES entries_new(id) ON DELETE CASCADE
);

INSERT INTO entries_new (id, type, content, parent_id, depth, location, scheduled_date, created_at, list_id)
SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, list_id FROM entries;

DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;

CREATE INDEX idx_entries_scheduled_date ON entries(scheduled_date);
CREATE INDEX idx_entries_parent_id ON entries(parent_id);
CREATE INDEX idx_entries_created_at ON entries(created_at);
