-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
CREATE TABLE habit_logs_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    habit_id INTEGER NOT NULL,
    count INTEGER NOT NULL DEFAULT 1,
    logged_at TEXT NOT NULL,
    FOREIGN KEY (habit_id) REFERENCES habits(id) ON DELETE CASCADE
);

INSERT INTO habit_logs_new (id, habit_id, count, logged_at)
SELECT id, habit_id, count, logged_at FROM habit_logs WHERE valid_to IS NULL OR valid_to = '';

DROP TABLE habit_logs;
ALTER TABLE habit_logs_new RENAME TO habit_logs;

CREATE INDEX idx_habit_logs_habit_id ON habit_logs(habit_id);
CREATE INDEX idx_habit_logs_logged_at ON habit_logs(logged_at);
