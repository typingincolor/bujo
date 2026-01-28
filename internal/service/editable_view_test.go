package service

import (
	"context"
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
	editableViewService := NewEditableViewService(entryRepo, bujoService)
	return editableViewService, bujoService, entryRepo
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
	require.Equal(t, ". Buy groceries", doc)
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
	require.Contains(t, doc, "  - Child note")
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

func TestApplyChanges_InsertNewEntry(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Existing task", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	newDoc := ". Existing task\n. New task"
	result, err := svc.ApplyChanges(ctx, newDoc, date, nil)

	require.NoError(t, err)
	require.Equal(t, 1, result.Inserted)
	require.Equal(t, 0, result.Updated)
	require.Equal(t, 0, result.Deleted)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Contains(t, doc, ". Existing task")
	require.Contains(t, doc, ". New task")
}

func TestApplyChanges_UpdateEntry(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	// Log entry - content-based matching requires content to stay the same
	_, err := bujoSvc.LogEntries(ctx, ". My task", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	// Update priority (keeps content same for matching, changes priority)
	newDoc := ". !!! My task"
	result, err := svc.ApplyChanges(ctx, newDoc, date, nil)

	require.NoError(t, err)
	require.Equal(t, 0, result.Inserted)
	require.Equal(t, 1, result.Updated)
	require.Equal(t, 0, result.Deleted)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Contains(t, doc, ". !!! My task")
}

func TestApplyChanges_DeleteEntry(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	ids, err := bujoSvc.LogEntries(ctx, ". Task to keep\n. Task to delete", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[1])
	require.NoError(t, err)
	deleteEntityID := entry.EntityID

	newDoc := ". Task to keep"
	result, err := svc.ApplyChanges(ctx, newDoc, date, []domain.EntityID{deleteEntityID})

	require.NoError(t, err)
	require.Equal(t, 0, result.Inserted)
	require.Equal(t, 0, result.Updated)
	require.Equal(t, 1, result.Deleted)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Contains(t, doc, ". Task to keep")
	require.NotContains(t, doc, ". Task to delete")
}

func TestApplyChanges_ChangeEntryType(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Incomplete task", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	newDoc := "x Incomplete task"
	result, err := svc.ApplyChanges(ctx, newDoc, date, nil)

	require.NoError(t, err)
	require.Equal(t, 1, result.Updated)

	doc, err := svc.GetEditableDocument(ctx, date)
	require.NoError(t, err)
	require.Contains(t, doc, "x Incomplete task")
}

func TestApplyChanges_MigrateEntry(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	ctx := context.Background()
	sourceDate := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	targetDate := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Task to migrate", LogEntriesOptions{Date: sourceDate})
	require.NoError(t, err)

	newDoc := ">[2026-01-29] . Task to migrate"
	result, err := svc.ApplyChanges(ctx, newDoc, sourceDate, nil)

	require.NoError(t, err)
	require.Equal(t, 1, result.Migrated)

	sourceDoc, err := svc.GetEditableDocument(ctx, sourceDate)
	require.NoError(t, err)
	require.Contains(t, sourceDoc, "> Task to migrate")

	targetDoc, err := svc.GetEditableDocument(ctx, targetDate)
	require.NoError(t, err)
	require.Contains(t, targetDoc, ". Task to migrate")
}

func TestApplyChanges_MultipleOperations(t *testing.T) {
	svc, bujoSvc, entryRepo := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	ids, err := bujoSvc.LogEntries(ctx, ". Keep unchanged\n. Will update priority\n. Will delete", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	deleteEntry, err := entryRepo.GetByID(ctx, ids[2])
	require.NoError(t, err)

	// Keep unchanged stays same, update priority on second (keeps content for matching), add new note
	newDoc := ". Keep unchanged\n. !!! Will update priority\n- New note"
	result, err := svc.ApplyChanges(ctx, newDoc, date, []domain.EntityID{deleteEntry.EntityID})

	require.NoError(t, err)
	require.Equal(t, 1, result.Inserted)
	require.Equal(t, 1, result.Updated)
	require.Equal(t, 1, result.Deleted)
}

func TestApplyChanges_ValidationErrors(t *testing.T) {
	svc, bujoSvc, _ := setupEditableViewService(t)
	ctx := context.Background()
	date := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	_, err := bujoSvc.LogEntries(ctx, ". Existing task", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	newDoc := "invalid line without symbol"
	_, err = svc.ApplyChanges(ctx, newDoc, date, nil)

	require.Error(t, err)
}
