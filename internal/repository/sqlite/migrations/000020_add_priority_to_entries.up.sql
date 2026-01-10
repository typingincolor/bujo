-- Add priority column to entries table
ALTER TABLE entries ADD COLUMN priority TEXT NOT NULL DEFAULT 'none' CHECK (priority IN ('none', 'low', 'medium', 'high'));
