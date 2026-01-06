CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    parent_id INTEGER,
    depth INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    scheduled_date TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES entries(id) ON DELETE CASCADE
);

CREATE INDEX idx_entries_scheduled_date ON entries(scheduled_date);
CREATE INDEX idx_entries_parent_id ON entries(parent_id);
CREATE INDEX idx_entries_created_at ON entries(created_at);
