-- Clear entity_ids (reversible)
UPDATE entries SET entity_id = NULL, valid_from = NULL WHERE entity_id IS NOT NULL;
