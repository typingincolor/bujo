-- Populate entity_id for existing habits using SQLite's randomblob to generate UUID-like strings
-- Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx (v4 UUID format)
UPDATE habits
SET entity_id = lower(hex(randomblob(4))) || '-' ||
                lower(hex(randomblob(2))) || '-4' ||
                substr(lower(hex(randomblob(2))),2) || '-' ||
                substr('89ab',abs(random()) % 4 + 1, 1) ||
                substr(lower(hex(randomblob(2))),2) || '-' ||
                lower(hex(randomblob(6))),
    valid_from = created_at
WHERE entity_id IS NULL;
