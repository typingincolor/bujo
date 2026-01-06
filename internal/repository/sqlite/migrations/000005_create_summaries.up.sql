CREATE TABLE summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    horizon TEXT NOT NULL,
    content TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_summaries_horizon ON summaries(horizon);
CREATE INDEX idx_summaries_dates ON summaries(start_date, end_date);
