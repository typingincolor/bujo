# Entry Tags Design

## Issue

#363: Tags/labels for entries

## Scope (this PR)

- Domain: Tag parsing from content (`#tag` extraction)
- Database: `entry_tags` table + migration
- Repository: Tag storage, query by tags
- Service: Wire tag extraction into entry save flow
- Search: `--tag` flag with OR semantics
- Wails adapter: Expose tag search to frontend

Follow-on work tracked in #458: frontend display, autocomplete, tag management.

## Design Decisions

1. **Tags extracted from content** - Tags are written inline (`#shopping`) and parsed automatically. The `#tag` text stays in the content string. Tags are derived data.

2. **Normalize to lowercase** - `#Shopping` and `#shopping` are the same tag, stored as `shopping`.

3. **Allowed characters** - Alphanumeric + hyphens, must start with a letter. Regex: `#([a-zA-Z][a-zA-Z0-9-]*)`.

4. **No event sourcing for tags** - Entries don't use event sourcing (delete-by-date + reinsert). Tags follow the same pattern: cascade delete + reinsert from parsed content.

5. **OR semantics for multi-tag search** - `--tag shopping,errands` returns entries with either tag.

## Domain Layer

New file `domain/tag.go`:

```go
func ExtractTags(content string) []string
```

Extracts `#tag` tokens from content, normalizes to lowercase, deduplicates, returns sorted.

Entry struct gets:

```go
Tags []string
```

TreeParser calls `ExtractTags` on each entry's content after parsing.

## Database

Migration creates `entry_tags`:

```sql
CREATE TABLE entry_tags (
    entry_id INTEGER NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (entry_id, tag),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE
);
CREATE INDEX idx_entry_tags_tag ON entry_tags(tag);
```

## Repository

```go
type TagRepository interface {
    InsertEntryTags(ctx context.Context, entryID int64, tags []string) error
    GetTagsForEntries(ctx context.Context, entryIDs []int64) (map[int64][]string, error)
    GetAllTags(ctx context.Context) ([]string, error)
    DeleteByEntryID(ctx context.Context, entryID int64) error
}
```

## Service Integration

- After inserting entries, extract tags and store via TagRepository
- When loading entries, hydrate Tags field via GetTagsForEntries
- SaveDayEntries: cascade handles cleanup, reinsert handles fresh tags

## Search

SearchOptions gets `Tags []string`. When non-empty:

```sql
SELECT DISTINCT e.* FROM entries e
JOIN entry_tags et ON e.id = et.entry_id
WHERE et.tag IN (?, ?)
```

Combined with existing content/type/date filters.

## CLI

`--tag` flag on search command, comma-separated values.

## Wails Adapter

- SearchEntries passes Tags through
- New GetAllTags() method for future autocomplete
