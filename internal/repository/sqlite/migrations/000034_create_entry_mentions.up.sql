CREATE TABLE entry_mentions (
    entry_id INTEGER NOT NULL,
    mention TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (entry_id, mention),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE
);

CREATE INDEX idx_entry_mentions_mention ON entry_mentions(mention);
