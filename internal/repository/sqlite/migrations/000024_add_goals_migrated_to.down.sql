-- Recreate original table without migrated_to column and with original CHECK constraint
CREATE TABLE goals_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id TEXT NOT NULL,
    content TEXT NOT NULL,
    month TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'done')),
    created_at TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT,
    valid_to TEXT,
    op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'))
);

-- Copy data (excluding migrated goals which would fail the CHECK constraint)
INSERT INTO goals_old (id, entity_id, content, month, status, created_at, version, valid_from, valid_to, op_type)
SELECT id, entity_id, content, month,
    CASE WHEN status = 'migrated' THEN 'active' ELSE status END,
    created_at, version, valid_from, valid_to, op_type
FROM goals;

DROP TABLE goals;
ALTER TABLE goals_old RENAME TO goals;

CREATE INDEX idx_goals_entity_id ON goals(entity_id);
CREATE INDEX idx_goals_month ON goals(month);
CREATE INDEX idx_goals_status ON goals(status);
CREATE INDEX idx_goals_valid_to ON goals(valid_to);
