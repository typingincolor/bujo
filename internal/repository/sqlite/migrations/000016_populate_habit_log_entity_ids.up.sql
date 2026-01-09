-- Populate entity_id for existing habit_logs using SQLite's randomblob to generate UUID-like strings
UPDATE habit_logs
SET entity_id = lower(hex(randomblob(4))) || '-' ||
                lower(hex(randomblob(2))) || '-4' ||
                substr(lower(hex(randomblob(2))),2) || '-' ||
                substr('89ab',abs(random()) % 4 + 1, 1) ||
                substr(lower(hex(randomblob(2))),2) || '-' ||
                lower(hex(randomblob(6))),
    valid_from = logged_at
WHERE entity_id IS NULL;

-- Link habit_entity_id to the habit's entity_id
UPDATE habit_logs
SET habit_entity_id = (SELECT entity_id FROM habits WHERE habits.id = habit_logs.habit_id LIMIT 1)
WHERE habit_entity_id IS NULL;
