-- Reparent current entries whose parent_id points to a superseded history row.
-- Update parent_id to the current version (valid_to IS NULL) of the same entity.
UPDATE entries SET parent_id = (
    SELECT cur.id
    FROM entries old
    JOIN entries cur ON old.entity_id = cur.entity_id AND cur.valid_to IS NULL AND cur.op_type != 'DELETE'
    WHERE old.id = entries.parent_id
)
WHERE parent_id IN (
    SELECT id FROM entries WHERE valid_to IS NOT NULL OR op_type = 'DELETE'
)
AND (valid_to IS NULL AND op_type != 'DELETE');

-- Break FK chains between history rows so they can be deleted in any order
UPDATE entries SET parent_id = NULL
WHERE (valid_to IS NOT NULL OR op_type = 'DELETE')
AND parent_id IS NOT NULL;

-- Remove superseded history rows (old versions replaced by newer ones)
DELETE FROM entries WHERE valid_to IS NOT NULL;

-- Remove soft-delete tombstone rows (entities that were deleted)
DELETE FROM entries WHERE op_type = 'DELETE';
