-- Add versioning columns to entries table for event sourcing
ALTER TABLE entries ADD COLUMN entity_id TEXT;
ALTER TABLE entries ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE entries ADD COLUMN valid_from TEXT;
ALTER TABLE entries ADD COLUMN valid_to TEXT;
ALTER TABLE entries ADD COLUMN op_type TEXT NOT NULL DEFAULT 'INSERT' CHECK (op_type IN ('INSERT', 'UPDATE', 'DELETE'));

-- Index for entity lookups
CREATE INDEX idx_entries_entity_id ON entries(entity_id);
CREATE INDEX idx_entries_valid_to ON entries(valid_to);
