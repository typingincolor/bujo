-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
CREATE TABLE habits_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    goal_per_day INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL
);

INSERT INTO habits_new (id, name, goal_per_day, created_at)
SELECT id, name, goal_per_day, created_at FROM habits WHERE valid_to IS NULL OR valid_to = '';

DROP TABLE habits;
ALTER TABLE habits_new RENAME TO habits;

CREATE INDEX idx_habits_name ON habits(name);
