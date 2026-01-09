-- Revert to original schema without versioning columns
CREATE TABLE day_context_new (
    date TEXT PRIMARY KEY,
    location TEXT,
    mood TEXT,
    weather TEXT
);

INSERT INTO day_context_new (date, location, mood, weather)
SELECT date, location, mood, weather FROM day_context WHERE valid_to IS NULL OR valid_to = '';

DROP TABLE day_context;
ALTER TABLE day_context_new RENAME TO day_context;
