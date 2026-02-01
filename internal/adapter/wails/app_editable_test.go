package wails

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/service"
)

func TestApp_GetEditableDocument_ReturnsSerializedEntries(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	_, err = services.Bujo.LogEntries(ctx, ". Task one\n- Note two\no Event three", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)

	doc, err := wailsApp.GetEditableDocument(today)

	require.NoError(t, err)
	assert.Contains(t, doc, ". Task one")
	assert.Contains(t, doc, "- Note two")
	assert.Contains(t, doc, "o Event three")
}

func TestApp_GetEditableDocument_EmptyDay(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	doc, err := wailsApp.GetEditableDocument(today)

	require.NoError(t, err)
	assert.Equal(t, "", doc)
}

func TestApp_ValidateEditableDocument_ValidDocument(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	result := wailsApp.ValidateEditableDocument(". Valid task\n- Valid note")

	assert.True(t, result.IsValid)
	assert.Empty(t, result.Errors)
}

func TestApp_ValidateEditableDocument_InvalidDocument(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	result := wailsApp.ValidateEditableDocument("Invalid line without symbol")

	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Errors)
}

func TestApp_ValidateEditableDocument_OrphanChild(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	result := wailsApp.ValidateEditableDocument("  . Orphan child without parent")

	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "Orphan")
}

func TestApp_ApplyEditableDocument_InsertsNewEntries(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	result, err := wailsApp.ApplyEditableDocument(". New task\n- New note", today)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Inserted)
	assert.Equal(t, 0, result.Deleted)

	doc, err := wailsApp.GetEditableDocument(today)
	require.NoError(t, err)
	assert.Contains(t, doc, ". New task")
	assert.Contains(t, doc, "- New note")
}

func TestApp_ApplyEditableDocument_DeletesEntries(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	_, err = services.Bujo.LogEntries(ctx, ". Task to keep\n. Task to delete", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)

	result, err := wailsApp.ApplyEditableDocument(". Task to keep", today)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Inserted)
	assert.Equal(t, 2, result.Deleted)

	doc, err := wailsApp.GetEditableDocument(today)
	require.NoError(t, err)
	assert.Contains(t, doc, ". Task to keep")
	assert.NotContains(t, doc, ". Task to delete")
}

func TestApp_ApplyEditableDocument_ValidationError(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	_, err = wailsApp.ApplyEditableDocument("Invalid document", today)

	require.Error(t, err)
}

func TestApp_ApplyEditableDocument_DeletesMissingEntries(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	_, err = services.Bujo.LogEntries(ctx, ". Task one\n. Task two", service.LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Verify both entries exist
	doc, err := wailsApp.GetEditableDocument(today)
	require.NoError(t, err)
	assert.Contains(t, doc, ". Task one")
	assert.Contains(t, doc, ". Task two")

	// Apply empty document WITHOUT explicit pendingDeletes
	// The backend should auto-detect that both entries are missing
	result, err := wailsApp.ApplyEditableDocument("", today)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Inserted)
	assert.Equal(t, 2, result.Deleted, "Expected 2 entries to be auto-deleted")

	// Verify entries are gone
	doc, err = wailsApp.GetEditableDocument(today)
	require.NoError(t, err)
	assert.Empty(t, doc)
}

func TestApp_ResolveDate_Tomorrow(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	result, err := wailsApp.ResolveDate("tomorrow")

	require.NoError(t, err)
	assert.NotEmpty(t, result.ISO)
	assert.NotEmpty(t, result.Display)

	tomorrow := time.Now().AddDate(0, 0, 1)
	expected := tomorrow.Format("2006-01-02")
	assert.Equal(t, expected, result.ISO)
}

func TestApp_ResolveDate_ISODate(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	result, err := wailsApp.ResolveDate("2026-02-15")

	require.NoError(t, err)
	assert.Equal(t, "2026-02-15", result.ISO)
	assert.Contains(t, result.Display, "Feb")
	assert.Contains(t, result.Display, "15")
}

func TestApp_ResolveDate_InvalidDate(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	_, err = wailsApp.ResolveDate("")

	require.Error(t, err)
}


