ALTER TABLE lists ADD COLUMN entity_id TEXT;

-- Populate entity_id for existing lists using SQLite's randomblob for UUID generation
UPDATE lists
SET entity_id = (
    lower(hex(randomblob(4))) || '-' ||
    lower(hex(randomblob(2))) || '-4' ||
    substr(lower(hex(randomblob(2))), 2) || '-' ||
    substr('89ab', abs(random()) % 4 + 1, 1) ||
    substr(lower(hex(randomblob(2))), 2) || '-' ||
    lower(hex(randomblob(6)))
)
WHERE entity_id IS NULL;
