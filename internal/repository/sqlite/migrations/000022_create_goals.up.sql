-- Create goals table for monthly goals tracking
CREATE TABLE IF NOT EXISTS goals (
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

CREATE INDEX idx_goals_entity_id ON goals(entity_id);
CREATE INDEX idx_goals_month ON goals(month);
CREATE INDEX idx_goals_status ON goals(status);
CREATE INDEX idx_goals_valid_to ON goals(valid_to);
