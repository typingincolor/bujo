-- Add versioning columns to habit_logs table for event sourcing
ALTER TABLE habit_logs ADD COLUMN entity_id TEXT;
ALTER TABLE habit_logs ADD COLUMN habit_entity_id TEXT;
ALTER TABLE habit_logs ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE habit_logs ADD COLUMN valid_from TEXT;
ALTER TABLE habit_logs ADD COLUMN valid_to TEXT;
ALTER TABLE habit_logs ADD COLUMN op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'));

-- Index for entity lookups
CREATE INDEX idx_habit_logs_entity_id ON habit_logs(entity_id);
CREATE INDEX idx_habit_logs_valid_to ON habit_logs(valid_to);
