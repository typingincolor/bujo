CREATE TABLE entry_tags (
    entry_id INTEGER NOT NULL,
    tag TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (entry_id, tag),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE
);

CREATE INDEX idx_entry_tags_tag ON entry_tags(tag);
