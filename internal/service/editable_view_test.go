package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupEditableViewService(t *testing.T) (*EditableViewService, *BujoService, *sqlite.EntryRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	parser := domain.NewTreeParser()

	bujoService := NewBujoService(entryRepo, dayCtxRepo, parser)
	editableViewService := NewEditableViewService(entryRepo, nil, nil)
	return editableViewService, bujoService, entryRepo
}

func setupEditableViewServiceWithLists(t *testing.T) (*EditableViewService, *sqlite.EntryRepository, *sqlite.ListRepository, *sqlite.ListItemRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	listRepo := sqlite.NewListRepository(db)
	listItemRepo := sqlite.NewListItemRepository(db)
	entryToListMover := sqlite.NewEntryToListMover(db)

	editableViewService := NewEditableViewService(entryRepo, entryToListMover, listRepo)
	return editableViewService, entryRepo, listRepo, listItemRepo
}

func TestGetEditableDocument_EmptyDay(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Empty(t, doc)
}

func TestGetEditableDocument_SingleEntry(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(context.Background(), ". Buy groceries", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". Buy groceries")
}

func TestGetEditableDocument_MultipleEntries(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(context.Background(), ". Task one\n- Note two\no Event three", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". Task one")
	require.Contains(t, doc, "- Note two")
	require.Contains(t, doc, "o Event three")
}

func TestGetEditableDocument_WithHierarchy(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(context.Background(), ". Parent task\n  - Child note", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". Parent task")
	require.Contains(t, doc, "- Child note")
}

func TestGetEditableDocument_WithPriority(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	ids, err := bujoSvc.LogEntries(context.Background(), ". Important task", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(context.Background(), ids[0])
	require.NoError(t, err)
	entry.Priority = domain.PriorityHigh
	err = entryRepo.Update(context.Background(), *entry)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(context.Background(), date)

	require.NoError(t, err)
	require.Contains(t, doc, ". !!! Important task")
}

func TestValidateDocument_ValidDocument(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument(". Task one\n- Note two")

	require.True(t, result.IsValid)
	require.Empty(t, result.Errors)
	require.Len(t, result.ParsedLines, 2)
}

func TestValidateDocument_InvalidLine(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument("invalid line without symbol")

	require.False(t, result.IsValid)
	require.Len(t, result.Errors, 1)
	require.Equal(t, 1, result.Errors[0].LineNumber)
}

func TestValidateDocument_OrphanChild(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument("  . Orphan child at depth 1")

	require.False(t, result.IsValid)
	require.Len(t, result.Errors, 1)
}

func TestValidateDocument_MixedValidInvalid(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument(". Valid task\ninvalid line\n- Valid note")

	require.False(t, result.IsValid)
	require.Len(t, result.Errors, 1)
	require.Equal(t, 2, result.Errors[0].LineNumber)
	require.Len(t, result.ParsedLines, 3)
}

func TestValidateDocument_EmptyDocument(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)

	result := svc.ValidateDocument("")

	require.True(t, result.IsValid)
	require.Empty(t, result.Errors)
	require.Empty(t, result.ParsedLines)
}

func TestApplyChanges_InsertEntries(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	result, err := svc.ApplyChanges(ctx, ". Task one\n- Note two", date)

	require.NoError(t, err)
	require.Equal(t, 2, result.Inserted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "Task one", entries[0].Content)
	require.Equal(t, "Note two", entries[1].Content)
}

func TestApplyChanges_ReplacesExistingEntries(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Old task\n- Old note", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	result, err := svc.ApplyChanges(ctx, ". New task\no New event", date)

	require.NoError(t, err)
	require.Equal(t, 2, result.Inserted)
	require.Equal(t, 2, result.Deleted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "New task", entries[0].Content)
	require.Equal(t, "New event", entries[1].Content)
}

func TestApplyChanges_ChildEntriesGetParentID(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	result, err := svc.ApplyChanges(ctx, ". Parent task\n  - Child note\n    - Grandchild note", date)

	require.NoError(t, err)
	require.Equal(t, 3, result.Inserted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	parent := entries[0]
	child := entries[1]
	grandchild := entries[2]

	require.Equal(t, "Parent task", parent.Content)
	require.Equal(t, "Child note", child.Content)
	require.Equal(t, "Grandchild note", grandchild.Content)

	require.Nil(t, parent.ParentID)
	require.NotNil(t, child.ParentID)
	require.Equal(t, parent.ID, *child.ParentID)
	require.NotNil(t, grandchild.ParentID)
	require.Equal(t, child.ID, *grandchild.ParentID)
}

func TestApplyChanges_SiblingsShareParent(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". Parent\n  - Sibling one\n  - Sibling two", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	require.NotNil(t, entries[1].ParentID)
	require.NotNil(t, entries[2].ParentID)
	require.Equal(t, entries[0].ID, *entries[1].ParentID)
	require.Equal(t, entries[0].ID, *entries[2].ParentID)
}

func TestApplyChanges_PreservesDocumentOrder(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". First\n- Second\no Third", date)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	lines := splitNonEmpty(doc)
	require.Len(t, lines, 3)
	require.Contains(t, lines[0], "First")
	require.Contains(t, lines[1], "Second")
	require.Contains(t, lines[2], "Third")
}

func TestApplyChanges_ReorderPreserved(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". First\n- Second\no Third", date)
	require.NoError(t, err)

	_, err = svc.ApplyChanges(ctx, "o Third\n. First\n- Second", date)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	lines := splitNonEmpty(doc)
	require.Len(t, lines, 3)
	require.Contains(t, lines[0], "Third")
	require.Contains(t, lines[1], "First")
	require.Contains(t, lines[2], "Second")
}

func TestApplyChanges_ValidationErrors(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, "invalid line", date)
	require.Error(t, err)
}

func TestApplyChanges_EmptyDocument(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Task to remove", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	result, err := svc.ApplyChanges(ctx, "", date)
	require.NoError(t, err)
	require.Equal(t, 0, result.Inserted)
	require.Equal(t, 1, result.Deleted)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestApplyChanges_WithPriority(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". !!! High priority task", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, domain.PriorityHigh, entries[0].Priority)
	require.Equal(t, "High priority task", entries[0].Content)
}

func TestApplyChanges_RoundTrip(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	original := ". Parent task\n  - Child one\n  - Child two\no Standalone event"
	_, err := svc.ApplyChanges(ctx, original, date)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)

	_, err = svc.ApplyChanges(ctx, doc, date)
	require.NoError(t, err)

	doc2, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Equal(t, doc, doc2)
}

func TestApplyChanges_QuestionWithChildBecomesAnswered(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, "? How does auth work\n  - It uses JWT tokens", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, domain.EntryTypeAnswered, entries[0].Type)
	require.Equal(t, domain.EntryTypeAnswer, entries[1].Type)
}

func TestApplyChanges_QuestionWithoutChildStaysQuestion(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, "? How does auth work", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, domain.EntryTypeQuestion, entries[0].Type)
}

func TestApplyChanges_QuestionWithMultipleChildrenAllBecomeAnswers(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, "? What tools do we use\n  - Go for backend\n  - React for frontend", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 3)
	require.Equal(t, domain.EntryTypeAnswered, entries[0].Type)
	require.Equal(t, domain.EntryTypeAnswer, entries[1].Type)
	require.Equal(t, domain.EntryTypeAnswer, entries[2].Type)
}

func TestApplyChanges_TaskWithChildrenTypesUnchanged(t *testing.T) {
	svc, _, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, ". Parent task\n  - Child note", date)
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, domain.EntryTypeTask, entries[0].Type)
	require.Equal(t, domain.EntryTypeNote, entries[1].Type)
}

func TestApplyChanges_AnsweredQuestionRoundTrips(t *testing.T) {
	svc, _, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChanges(ctx, "? How does auth work\n  - It uses JWT tokens", date)
	require.NoError(t, err)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Equal(t, "* How does auth work\n  - It uses JWT tokens", doc)

	_, err = svc.ApplyChanges(ctx, doc, date)
	require.NoError(t, err)

	doc2, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Equal(t, doc, doc2)
}

func TestApplyChangesWithActions_Migration(t *testing.T) {
	svc, entryRepo, _, _ := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	result, err := svc.ApplyChangesWithActions(ctx, ". Keep this\n> Migrate this", today, ApplyActions{
		MigrateDate: &tomorrow,
	})

	require.NoError(t, err)
	require.Equal(t, 2, result.Inserted)

	todayEntries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, todayEntries, 2)
	require.Equal(t, "Keep this", todayEntries[0].Content)
	require.Equal(t, domain.EntryTypeTask, todayEntries[0].Type)
	require.Equal(t, "Migrate this", todayEntries[1].Content)
	require.Equal(t, domain.EntryTypeMigrated, todayEntries[1].Type)

	tomorrowEntries, err := entryRepo.GetByDate(ctx, tomorrow)
	require.NoError(t, err)
	require.Len(t, tomorrowEntries, 1)
	require.Equal(t, "Migrate this", tomorrowEntries[0].Content)
	require.Equal(t, domain.EntryTypeTask, tomorrowEntries[0].Type)
}

func TestApplyChangesWithActions_MoveToList(t *testing.T) {
	svc, entryRepo, listRepo, listItemRepo := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	result, err := svc.ApplyChangesWithActions(ctx, ". Keep this\n^ Move this to list", today, ApplyActions{
		ListID: &list.ID,
	})

	require.NoError(t, err)
	require.Equal(t, 2, result.Inserted)

	todayEntries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, todayEntries, 1)
	require.Equal(t, "Keep this", todayEntries[0].Content)

	listItems, err := listItemRepo.GetByListEntityID(ctx, list.EntityID)
	require.NoError(t, err)
	require.Len(t, listItems, 1)
	require.Equal(t, "Move this to list", listItems[0].Content)
}

func TestApplyChangesWithActions_BothMigrationAndMoveToList(t *testing.T) {
	svc, entryRepo, listRepo, listItemRepo := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.ApplyChangesWithActions(ctx, "> Migrate this\n^ Move this", today, ApplyActions{
		MigrateDate: &tomorrow,
		ListID:      &list.ID,
	})
	require.NoError(t, err)

	todayEntries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, todayEntries, 1)
	require.Equal(t, "Migrate this", todayEntries[0].Content)
	require.Equal(t, domain.EntryTypeMigrated, todayEntries[0].Type)

	tomorrowEntries, err := entryRepo.GetByDate(ctx, tomorrow)
	require.NoError(t, err)
	require.Len(t, tomorrowEntries, 1)
	require.Equal(t, "Migrate this", tomorrowEntries[0].Content)
	require.Equal(t, domain.EntryTypeTask, tomorrowEntries[0].Type)

	listItems, err := listItemRepo.GetByListEntityID(ctx, list.EntityID)
	require.NoError(t, err)
	require.Len(t, listItems, 1)
	require.Equal(t, "Move this", listItems[0].Content)
}

func TestApplyChangesWithActions_NoActions(t *testing.T) {
	svc, entryRepo, _, _ := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	result, err := svc.ApplyChangesWithActions(ctx, ". Task one\n- Note two", today, ApplyActions{})

	require.NoError(t, err)
	require.Equal(t, 2, result.Inserted)

	entries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, entries, 2)
}

func TestApplyChangesWithActions_MigrationPreservesPriority(t *testing.T) {
	svc, entryRepo, _, _ := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChangesWithActions(ctx, "> !!! Important migrate", today, ApplyActions{
		MigrateDate: &tomorrow,
	})
	require.NoError(t, err)

	tomorrowEntries, err := entryRepo.GetByDate(ctx, tomorrow)
	require.NoError(t, err)
	require.Len(t, tomorrowEntries, 1)
	require.Equal(t, "Important migrate", tomorrowEntries[0].Content)
	require.Equal(t, domain.PriorityHigh, tomorrowEntries[0].Priority)
}

func TestApplyChangesWithActions_MigrationWithChildren(t *testing.T) {
	svc, entryRepo, _, _ := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChangesWithActions(ctx, "> Migrate parent\n  . Child task\n  - Child note", today, ApplyActions{
		MigrateDate: &tomorrow,
	})
	require.NoError(t, err)

	todayEntries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, todayEntries, 3)
	require.Equal(t, domain.EntryTypeMigrated, todayEntries[0].Type)
	require.Equal(t, domain.EntryTypeTask, todayEntries[1].Type)
	require.Equal(t, domain.EntryTypeNote, todayEntries[2].Type)

	tomorrowEntries, err := entryRepo.GetByDate(ctx, tomorrow)
	require.NoError(t, err)
	require.Len(t, tomorrowEntries, 1)
	require.Equal(t, "Migrate parent", tomorrowEntries[0].Content)
}

func TestApplyChangesWithActions_MultipleMigratedEntries(t *testing.T) {
	svc, entryRepo, _, _ := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChangesWithActions(ctx, "> First migrate\n. Stay here\n> Second migrate", today, ApplyActions{
		MigrateDate: &tomorrow,
	})
	require.NoError(t, err)

	todayEntries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, todayEntries, 3)

	tomorrowEntries, err := entryRepo.GetByDate(ctx, tomorrow)
	require.NoError(t, err)
	require.Len(t, tomorrowEntries, 2)
	require.Equal(t, "First migrate", tomorrowEntries[0].Content)
	require.Equal(t, "Second migrate", tomorrowEntries[1].Content)
}

func TestApplyChangesWithActions_MultipleMovedToListEntries(t *testing.T) {
	svc, entryRepo, listRepo, listItemRepo := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.ApplyChangesWithActions(ctx, "^ First item\n. Stay here\n^ Second item", today, ApplyActions{
		ListID: &list.ID,
	})
	require.NoError(t, err)

	todayEntries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, todayEntries, 1)
	require.Equal(t, "Stay here", todayEntries[0].Content)

	listItems, err := listItemRepo.GetByListEntityID(ctx, list.EntityID)
	require.NoError(t, err)
	require.Len(t, listItems, 2)
	require.Equal(t, "First item", listItems[0].Content)
	require.Equal(t, "Second item", listItems[1].Content)
}

func TestApplyChangesWithActions_MoveToListWithChildren_FailsOnFKConstraint(t *testing.T) {
	svc, _, listRepo, _ := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	list, err := listRepo.Create(ctx, "Tasks")
	require.NoError(t, err)

	_, err = svc.ApplyChangesWithActions(ctx, "^ Move parent\n  . Child task", today, ApplyActions{
		ListID: &list.ID,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "FOREIGN KEY")
}

func TestApplyChangesWithActions_MoveToListPreservesPriority(t *testing.T) {
	svc, _, listRepo, listItemRepo := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	list, err := listRepo.Create(ctx, "Important")
	require.NoError(t, err)

	_, err = svc.ApplyChangesWithActions(ctx, "^ !! Medium priority item", today, ApplyActions{
		ListID: &list.ID,
	})
	require.NoError(t, err)

	listItems, err := listItemRepo.GetByListEntityID(ctx, list.EntityID)
	require.NoError(t, err)
	require.Len(t, listItems, 1)
	require.Equal(t, "Medium priority item", listItems[0].Content)
}

func TestApplyChangesWithActions_MigrationWithNoMigratedEntries(t *testing.T) {
	svc, entryRepo, _, _ := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	_, err := svc.ApplyChangesWithActions(ctx, ". Task one\n- Note two", today, ApplyActions{
		MigrateDate: &tomorrow,
	})
	require.NoError(t, err)

	tomorrowEntries, err := entryRepo.GetByDate(ctx, tomorrow)
	require.NoError(t, err)
	require.Empty(t, tomorrowEntries)
}

func TestApplyChangesWithActions_MoveToListWithNoMovedEntries(t *testing.T) {
	svc, entryRepo, listRepo, listItemRepo := setupEditableViewServiceWithLists(t)
	ctx := context.Background()
	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	_, err = svc.ApplyChangesWithActions(ctx, ". Task one\n- Note two", today, ApplyActions{
		ListID: &list.ID,
	})
	require.NoError(t, err)

	entries, err := entryRepo.GetByDate(ctx, today)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	listItems, err := listItemRepo.GetByListEntityID(ctx, list.EntityID)
	require.NoError(t, err)
	require.Empty(t, listItems)
}

func splitNonEmpty(s string) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}
