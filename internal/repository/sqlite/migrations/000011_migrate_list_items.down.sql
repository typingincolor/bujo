-- Remove migrated list items (keep original entries intact)
DELETE FROM list_items WHERE list_id IS NOT NULL;
