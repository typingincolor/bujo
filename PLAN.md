# Event Sourcing Implementation Plan

TDD implementation plan for issues #47, #50, #54.

## Principles

- RED: Write failing test first
- GREEN: Minimum code to pass
- REFACTOR: Clean up if needed
- COMMIT: After each feature passes

---

## Phase 1: Domain Layer Foundation

### 1.1 Create EntityID type and generator

**Test:** `internal/domain/entity_id_test.go`
```go
func TestNewEntityID_ReturnsNonEmpty(t *testing.T)
func TestNewEntityID_ReturnsUniqueValues(t *testing.T)
func TestEntityID_String_ReturnsValue(t *testing.T)
func TestParseEntityID_ValidUUID_Succeeds(t *testing.T)
func TestParseEntityID_InvalidUUID_Fails(t *testing.T)
func TestParseEntityID_EmptyString_Fails(t *testing.T)
```

**Implementation:** `internal/domain/entity_id.go`
- `EntityID` type (string wrapper)
- `NewEntityID()` function (generates UUID)
- `ParseEntityID(s string)` function (validates and parses)

---

### 1.2 Create OpType constants

**Test:** `internal/domain/version_test.go`
```go
func TestOpType_Insert_IsValid(t *testing.T)
func TestOpType_Update_IsValid(t *testing.T)
func TestOpType_Delete_IsValid(t *testing.T)
func TestOpType_Invalid_IsNotValid(t *testing.T)
```

**Implementation:** `internal/domain/version.go`
- `OpType` type
- Constants: `OpTypeInsert`, `OpTypeUpdate`, `OpTypeDelete`
- `IsValid()` method

---

### 1.3 Create VersionInfo struct

**Test:** Add to `internal/domain/version_test.go`
```go
func TestVersionInfo_IsCurrent_WhenValidToNil_ReturnsTrue(t *testing.T)
func TestVersionInfo_IsCurrent_WhenValidToSet_ReturnsFalse(t *testing.T)
func TestVersionInfo_IsDeleted_WhenOpTypeDelete_ReturnsTrue(t *testing.T)
func TestVersionInfo_IsDeleted_WhenOpTypeInsert_ReturnsFalse(t *testing.T)
```

**Implementation:** Add to `internal/domain/version.go`
```go
type VersionInfo struct {
    RowID     int64
    EntityID  EntityID
    Version   int
    ValidFrom time.Time
    ValidTo   *time.Time
    OpType    OpType
}
```

---

### 1.4 Create ListItemType (subset of EntryType)

**Test:** `internal/domain/list_item_test.go`
```go
func TestListItemType_Task_IsValid(t *testing.T)
func TestListItemType_Done_IsValid(t *testing.T)
func TestListItemType_Note_IsNotValid(t *testing.T)
func TestListItemType_Event_IsNotValid(t *testing.T)
func TestListItemType_Symbol_ReturnsCorrect(t *testing.T)
```

**Implementation:** `internal/domain/list_item.go`
- `ListItemType` type
- Constants: `ListItemTypeTask`, `ListItemTypeDone`
- `IsValid()` and `Symbol()` methods

---

### 1.5 Create ListItem domain type

**Test:** Add to `internal/domain/list_item_test.go`
```go
func TestListItem_Validate_ValidItem_Succeeds(t *testing.T)
func TestListItem_Validate_EmptyContent_Fails(t *testing.T)
func TestListItem_Validate_EmptyListEntityID_Fails(t *testing.T)
func TestListItem_Validate_InvalidType_Fails(t *testing.T)
func TestListItem_IsComplete_WhenDone_ReturnsTrue(t *testing.T)
func TestListItem_IsComplete_WhenTask_ReturnsFalse(t *testing.T)
func TestNewListItem_SetsEntityIDAndCreatedAt(t *testing.T)
```

**Implementation:** Add to `internal/domain/list_item.go`
```go
type ListItem struct {
    VersionInfo
    ListEntityID EntityID
    Type         ListItemType
    Content      string
    CreatedAt    time.Time
}
```

---

### 1.6 Add EntityID to Entry (non-breaking)

**Test:** Add to `internal/domain/entry_test.go`
```go
func TestEntry_WithEntityID_Validates(t *testing.T)
func TestNewEntry_GeneratesEntityID(t *testing.T)
```

**Implementation:** Update `internal/domain/entry.go`
- Add `EntityID EntityID` field to Entry struct
- Add `NewEntry()` constructor that generates EntityID

---

### 1.7 Add ParentEntityID to Entry (non-breaking)

**Test:** Add to `internal/domain/entry_test.go`
```go
func TestEntry_WithParentEntityID_Validates(t *testing.T)
func TestEntry_HasParent_WhenParentEntityIDSet_ReturnsTrue(t *testing.T)
func TestEntry_HasParent_WhenParentEntityIDEmpty_ReturnsFalse(t *testing.T)
```

**Implementation:** Update `internal/domain/entry.go`
- Add `ParentEntityID *EntityID` field (nullable, for migration period)

---

### 1.8 Add VersionInfo fields to Entry (non-breaking)

**Test:** Add to `internal/domain/entry_test.go`
```go
func TestEntry_IsCurrent_DelegatesToVersionInfo(t *testing.T)
func TestEntry_IsDeleted_DelegatesToVersionInfo(t *testing.T)
```

**Implementation:** Update `internal/domain/entry.go`
- Embed `VersionInfo` in Entry struct OR add fields directly
- Keep backward compatibility with existing ID field

---

### 1.9 Add EntityID to List (non-breaking)

**Test:** Add to `internal/domain/list_test.go`
```go
func TestList_WithEntityID_Validates(t *testing.T)
func TestNewList_GeneratesEntityID(t *testing.T)
```

**Implementation:** Update `internal/domain/list.go`
- Add `EntityID EntityID` field
- Update `NewList()` to generate EntityID

---

### 1.10 Add EntityID to Habit (non-breaking)

**Test:** Add to `internal/domain/habit_test.go`
```go
func TestHabit_WithEntityID_Validates(t *testing.T)
```

**Implementation:** Update `internal/domain/habit.go`
- Add `EntityID EntityID` field

---

### 1.11 Add EntityID to HabitLog (non-breaking)

**Test:** Add to `internal/domain/habit_test.go`
```go
func TestHabitLog_WithEntityID_Validates(t *testing.T)
func TestHabitLog_WithHabitEntityID_Validates(t *testing.T)
```

**Implementation:** Update `internal/domain/habit.go`
- Add `EntityID EntityID` field to HabitLog
- Add `HabitEntityID EntityID` field to HabitLog (for migration)

---

### 1.12 Add EntityID to Summary (non-breaking)

**Test:** Add to `internal/domain/summary_test.go`
```go
func TestSummary_WithEntityID_Validates(t *testing.T)
```

**Implementation:** Update `internal/domain/summary.go`
- Add `EntityID EntityID` field

---

### 1.13 Add EntityID to DayContext (non-breaking)

**Test:** Add to `internal/domain/context_test.go`
```go
func TestDayContext_WithEntityID_Validates(t *testing.T)
```

**Implementation:** Update `internal/domain/context.go`
- Add `EntityID EntityID` field

---

## Phase 2: Repository Interfaces

### 2.1 Create ListItemRepository interface

**Test:** None needed for interface definition

**Implementation:** Add to `internal/domain/repository.go`
```go
type ListItemRepository interface {
    Insert(ctx context.Context, item ListItem) (int64, error)
    GetByID(ctx context.Context, id int64) (*ListItem, error)
    GetByEntityID(ctx context.Context, entityID EntityID) (*ListItem, error)
    GetByListEntityID(ctx context.Context, listEntityID EntityID) ([]ListItem, error)
    Update(ctx context.Context, item ListItem) error
    Delete(ctx context.Context, id int64) error
    GetHistory(ctx context.Context, entityID EntityID) ([]ListItem, error)
}
```

---

### 2.2 Add versioned methods to EntryRepository interface

**Implementation:** Update `internal/domain/repository.go`
```go
// Add to EntryRepository interface:
GetByEntityID(ctx context.Context, entityID EntityID) (*Entry, error)
GetHistory(ctx context.Context, entityID EntityID) ([]Entry, error)
GetAsOf(ctx context.Context, entityID EntityID, asOf time.Time) (*Entry, error)
```

---

## Phase 3: SQLite Repository Implementation

### 3.1 Create ListItemRepository SQLite implementation

**Test:** `internal/repository/sqlite/list_item_repository_test.go`
```go
func TestListItemRepository_Insert_Success(t *testing.T)
func TestListItemRepository_GetByID_Found(t *testing.T)
func TestListItemRepository_GetByID_NotFound(t *testing.T)
func TestListItemRepository_GetByEntityID_Found(t *testing.T)
func TestListItemRepository_GetByListEntityID_ReturnsItems(t *testing.T)
func TestListItemRepository_Update_Success(t *testing.T)
func TestListItemRepository_Delete_SoftDeletes(t *testing.T)
func TestListItemRepository_GetHistory_ReturnsAllVersions(t *testing.T)
```

**Implementation:** `internal/repository/sqlite/list_item_repository.go`

---

### 3.2 Update EntryRepository for versioning

**Test:** Add to `internal/repository/sqlite/entry_repository_test.go`
```go
func TestEntryRepository_Insert_SetsVersionInfo(t *testing.T)
func TestEntryRepository_Update_CreatesNewVersion(t *testing.T)
func TestEntryRepository_Delete_SoftDeletes(t *testing.T)
func TestEntryRepository_GetByEntityID_ReturnsCurrent(t *testing.T)
func TestEntryRepository_GetHistory_ReturnsAllVersions(t *testing.T)
func TestEntryRepository_GetAsOf_ReturnsCorrectVersion(t *testing.T)
```

**Implementation:** Update `internal/repository/sqlite/entry_repository.go`

---

### 3.3-3.6 Update other repositories for versioning

Repeat pattern for: ListRepository, HabitRepository, HabitLogRepository, SummaryRepository, DayContextRepository

---

## Phase 4: Service Layer Updates

### 4.1 Update ListService to use ListItemRepository

**Test:** Update `internal/service/list_test.go`
```go
func TestListService_RemoveItem_OnlyRemovesListItems(t *testing.T)
func TestListService_RemoveItem_CannotRemoveEntry(t *testing.T)  // Issue #54 fixed
```

**Implementation:** Update `internal/service/list.go`
- Inject `ListItemRepository` instead of using `EntryRepository` for items
- Update all item operations to use new repository

---

### 4.2 Create HistoryService

**Test:** `internal/service/history_test.go`
```go
func TestHistoryService_GetEntryHistory_ReturnsAllVersions(t *testing.T)
func TestHistoryService_GetEntryAsOf_ReturnsCorrectState(t *testing.T)
func TestHistoryService_RestoreEntry_CreatesNewVersion(t *testing.T)
```

**Implementation:** `internal/service/history.go`

---

### 4.3 Create BackupService

**Test:** `internal/service/backup_test.go`
```go
func TestBackupService_CreateBackup_CreatesFile(t *testing.T)
func TestBackupService_CreateBackup_ValidSQLite(t *testing.T)
func TestBackupService_ListBackups_ReturnsFiles(t *testing.T)
func TestBackupService_VerifyBackup_ValidFile_Succeeds(t *testing.T)
func TestBackupService_VerifyBackup_CorruptFile_Fails(t *testing.T)
```

**Implementation:** `internal/service/backup.go`

---

### 4.4 Create ArchiveService

**Test:** `internal/service/archive_test.go`
```go
func TestArchiveService_Archive_MovesOldVersions(t *testing.T)
func TestArchiveService_Archive_KeepsRecentVersions(t *testing.T)
func TestArchiveService_Archive_KeepsCurrentState(t *testing.T)
func TestArchiveService_DryRun_DoesNotModify(t *testing.T)
```

**Implementation:** `internal/service/archive.go`

---

## Phase 5: Database Migration

### 5.1 Create migration: add entity_id columns

**File:** `internal/repository/sqlite/migrations/000007_add_entity_id.up.sql`
```sql
ALTER TABLE entries ADD COLUMN entity_id TEXT;
ALTER TABLE lists ADD COLUMN entity_id TEXT;
-- etc.
```

---

### 5.2 Create Go migration: generate UUIDs

**Test:** `internal/repository/sqlite/migrations/uuid_migration_test.go`
```go
func TestUUIDMigration_GeneratesUUIDsForExistingRows(t *testing.T)
func TestUUIDMigration_DoesNotOverwriteExistingUUIDs(t *testing.T)
```

**Implementation:** `internal/repository/sqlite/migrations/uuid_migration.go`

---

### 5.3 Create migration: add versioning columns

**File:** `internal/repository/sqlite/migrations/000008_add_versioning.up.sql`

---

### 5.4 Create migration: create list_items table

**File:** `internal/repository/sqlite/migrations/000009_create_list_items.up.sql`

---

### 5.5 Create Go migration: move list items

**Test:** `internal/repository/sqlite/migrations/list_items_migration_test.go`
```go
func TestListItemsMigration_MovesEntriesWithListID(t *testing.T)
func TestListItemsMigration_PreservesEntityID(t *testing.T)
func TestListItemsMigration_DeletesFromEntries(t *testing.T)
```

**Implementation:** `internal/repository/sqlite/migrations/list_items_migration.go`

---

### 5.6 Create migration: remove list_id from entries

**File:** `internal/repository/sqlite/migrations/000010_remove_list_id.up.sql`
(Recreates entries table without list_id)

---

## Phase 6: CLI Updates

### 6.1 Add backup command

**Test:** `cmd/bujo/cmd/backup_test.go`
```go
func TestBackupCmd_CreatesBackup(t *testing.T)
func TestBackupCmd_List_ShowsBackups(t *testing.T)
func TestBackupCmd_Verify_ValidatesFile(t *testing.T)
```

**Implementation:** `cmd/bujo/cmd/backup.go`

---

### 6.2 Add archive command

**Test:** `cmd/bujo/cmd/archive_test.go`
```go
func TestArchiveCmd_ArchivesOldVersions(t *testing.T)
func TestArchiveCmd_DryRun_ShowsWhatWouldArchive(t *testing.T)
```

**Implementation:** `cmd/bujo/cmd/archive.go`

---

### 6.3 Add history command

**Test:** `cmd/bujo/cmd/history_test.go`
```go
func TestHistoryCmd_ShowsVersionHistory(t *testing.T)
func TestHistoryCmd_AsOf_ShowsPastState(t *testing.T)
```

**Implementation:** `cmd/bujo/cmd/history.go`

---

## Phase 7: Cleanup

### 7.1 Remove deprecated fields

After migration is complete and verified:
- Remove `ListID` from Entry
- Remove `ParentID` from Entry (keep `ParentEntityID`)
- Update all code to use EntityID-based references

---

## Execution Order

```
Phase 1.1  → Phase 1.2  → Phase 1.3  → Phase 1.4  → Phase 1.5
    ↓
Phase 1.6  → Phase 1.7  → Phase 1.8  → Phase 1.9  → Phase 1.10
    ↓
Phase 1.11 → Phase 1.12 → Phase 1.13
    ↓
Phase 2.1  → Phase 2.2
    ↓
Phase 3.1  → Phase 3.2  → Phase 3.3-3.6
    ↓
Phase 4.1  → Phase 4.2  → Phase 4.3  → Phase 4.4
    ↓
Phase 5.1  → Phase 5.2  → Phase 5.3  → Phase 5.4  → Phase 5.5  → Phase 5.6
    ↓
Phase 6.1  → Phase 6.2  → Phase 6.3
    ↓
Phase 7.1
```

---

## Progress

### Completed

- [x] Phase 1: Domain Layer Foundation (all 13 steps)
  - EntityID type and generator
  - OpType constants and VersionInfo struct
  - ListItemType and ListItem domain types
  - Added EntityID to Entry, List, Habit, HabitLog, Summary, DayContext
  - Added ParentEntityID to Entry

- [x] Phase 2: Repository Interfaces
  - Created ListItemRepository interface
  - Added versioned methods to EntryRepository

- [x] Phase 3: SQLite Repository Implementation
  - Created list_items migration (000007)
  - Implemented ListItemRepository with event sourcing
  - Added entity_id migration to lists (000008)
  - Updated ListRepository with GetByEntityID

- [x] Phase 4.1: Fix issue #54
  - Added validation to RemoveItem, MarkDone, MarkUndone, MoveItem
  - Non-list entries cannot be operated on by ListService

- [x] Phase 4.2: Create BackupService
  - CreateBackup using SQLite VACUUM INTO
  - ListBackups to enumerate existing backups
  - VerifyBackup to check backup integrity

- [x] Phase 6.1: CLI backup command
  - `bujo backup` lists backups
  - `bujo backup create` creates a new backup
  - `bujo backup verify <path>` verifies backup integrity

- [x] Phase 4.3: Create ArchiveService
  - GetArchivableCount for dry-run preview
  - Archive to delete old versions
  - DryRun mode for safe preview

- [x] Phase 4.4: Create HistoryService
  - GetItemHistory returns all versions
  - GetItemAtVersion returns specific version
  - RestoreItem creates new version from old content

- [x] Phase 6.2: CLI archive command
  - `bujo archive` shows archivable count (dry run)
  - `bujo archive --execute` performs actual archive
  - `--older-than` flag to specify cutoff date

- [x] Phase 6.3: CLI history command
  - `bujo history show <entity-id>` shows version history
  - `bujo history restore <entity-id> <version>` restores to previous version

- [x] Phase 5: Database migrations for entries versioning
  - Migration 000009: Add entity_id, version, valid_from, valid_to, op_type columns
  - Migration 000010: Populate entity_ids for existing entries
  - Updated EntryRepository with GetByEntityID, GetHistory, GetAsOf
  - All queries filter out soft-deleted/superseded rows

### Remaining Work

- [ ] Migrate ListService to use ListItemRepository instead of EntryRepository for list items
- [ ] Phase 7: Cleanup deprecated fields after full migration
