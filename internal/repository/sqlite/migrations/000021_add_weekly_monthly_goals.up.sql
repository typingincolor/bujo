-- Add weekly and monthly goal columns to habits table
-- Default 0 means "not set" - at least one goal type must be positive

ALTER TABLE habits ADD COLUMN goal_per_week INTEGER NOT NULL DEFAULT 0;
ALTER TABLE habits ADD COLUMN goal_per_month INTEGER NOT NULL DEFAULT 0;
