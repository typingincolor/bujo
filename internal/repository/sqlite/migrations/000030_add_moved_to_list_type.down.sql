UPDATE entries SET type = 'task' WHERE type = 'movedToList';

CREATE TABLE entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('task', 'note', 'event', 'done', 'migrated', 'cancelled', 'question', 'answered', 'answer')),
    content TEXT NOT NULL,
    parent_id INTEGER,
    depth INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    scheduled_date TEXT NOT NULL,
    created_at TEXT NOT NULL,
    entity_id TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE')),
    priority TEXT NOT NULL DEFAULT 'none' CHECK (priority IN ('none', 'low', 'medium', 'high')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (parent_id) REFERENCES entries_new(id)
);

INSERT INTO entries_new (id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, valid_to, op_type, priority, sort_order)
SELECT id, type, content, parent_id, depth, location, scheduled_date, created_at, entity_id, version, valid_from, valid_to, op_type, priority, sort_order
FROM entries;

DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;

CREATE INDEX idx_entries_scheduled_date ON entries(scheduled_date);
CREATE INDEX idx_entries_parent_id ON entries(parent_id);
CREATE INDEX idx_entries_created_at ON entries(created_at);
CREATE INDEX idx_entries_entity_id ON entries(entity_id);
CREATE INDEX idx_entries_valid_to ON entries(valid_to);
