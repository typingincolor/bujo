-- Migrate existing list items from entries table to list_items table
-- This copies items with list_id set to the new list_items table

INSERT INTO list_items (entity_id, version, valid_from, op_type, list_entity_id, list_id, type, content, created_at)
SELECT
    e.entity_id,
    1,
    COALESCE(e.valid_from, e.created_at),
    'INSERT',
    l.entity_id,
    e.list_id,
    CASE e.type
        WHEN 'done' THEN 'done'
        ELSE 'task'
    END,
    e.content,
    e.created_at
FROM entries e
JOIN lists l ON e.list_id = l.id
WHERE e.list_id IS NOT NULL
  AND (e.valid_to IS NULL OR e.valid_to = '')
  AND e.entity_id IS NOT NULL;
