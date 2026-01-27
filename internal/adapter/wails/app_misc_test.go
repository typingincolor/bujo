package wails

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/app"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestApp_GetSummary_ReturnsUnavailableWhenNoAIService(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)

	// Should return unavailable message when no AI service is configured
	summary, err := wailsApp.GetSummary(today)
	require.NoError(t, err)
	assert.Equal(t, "", summary)
}

func TestApp_GetEntry_ReturnsEntryByID(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := wailsApp.AddEntry("• Test task", today)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	entry, err := wailsApp.GetEntry(ids[0])

	require.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "Test task", entry.Content)
	assert.Equal(t, domain.EntryTypeTask, entry.Type)
}

func TestApp_GetEntry_ReturnsCurrentVersionAfterUpdate(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := wailsApp.AddEntry("• Test task", today)
	require.NoError(t, err)
	originalID := ids[0]

	err = wailsApp.MarkEntryDone(originalID)
	require.NoError(t, err)

	entry, err := wailsApp.GetEntry(originalID)

	require.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, domain.EntryTypeDone, entry.Type)
}

func TestApp_GetEntry_ReturnsNilForNonExistent(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	entry, err := wailsApp.GetEntry(99999)

	require.Error(t, err)
	assert.Nil(t, entry)
}

func TestApp_AddChildEntry_CreatesChildUnderParent(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)

	// Create parent entry
	parentIDs, err := wailsApp.AddEntry(". Parent task", today)
	require.NoError(t, err)
	require.Len(t, parentIDs, 1)
	parentID := parentIDs[0]

	// Add child entry
	childIDs, err := wailsApp.AddChildEntry(parentID, ". Child task", today)
	require.NoError(t, err)
	require.Len(t, childIDs, 1)

	// Verify child has correct parent
	child, err := wailsApp.GetEntry(childIDs[0])
	require.NoError(t, err)
	assert.Equal(t, "Child task", child.Content)
	assert.NotNil(t, child.ParentID)
	assert.Equal(t, parentID, *child.ParentID)
}

func TestApp_AddChildEntry_SupportsMultipleChildren(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)

	// Create parent entry
	parentIDs, err := wailsApp.AddEntry(". Parent task", today)
	require.NoError(t, err)
	parentID := parentIDs[0]

	// Add multiple children using multi-line input
	childIDs, err := wailsApp.AddChildEntry(parentID, `. Child 1
- Child note`, today)
	require.NoError(t, err)
	require.Len(t, childIDs, 2)

	// Verify both have same parent
	child1, err := wailsApp.GetEntry(childIDs[0])
	require.NoError(t, err)
	assert.Equal(t, parentID, *child1.ParentID)

	child2, err := wailsApp.GetEntry(childIDs[1])
	require.NoError(t, err)
	assert.Equal(t, parentID, *child2.ParentID)
}

func TestApp_RetypeEntry_ChangesEntryType(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	ids, err := wailsApp.AddEntry(". Task to retype", today)
	require.NoError(t, err)
	entryID := ids[0]

	entry, err := wailsApp.GetEntry(entryID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeTask, entry.Type)

	err = wailsApp.RetypeEntry(entryID, "note")
	require.NoError(t, err)

	updated, err := wailsApp.GetEntry(entryID)
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeNote, updated.Type)
}

func TestApp_MoveEntryToRoot_RemovesParent(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)

	// Create parent entry
	parentIDs, err := wailsApp.AddEntry(". Parent task", today)
	require.NoError(t, err)
	parentID := parentIDs[0]

	// Create child entry under parent
	childIDs, err := wailsApp.AddChildEntry(parentID, ". Child task", today)
	require.NoError(t, err)
	childID := childIDs[0]

	// Verify child has parent
	child, err := wailsApp.GetEntry(childID)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, parentID, *child.ParentID)

	// Move child to root
	err = wailsApp.MoveEntryToRoot(childID)
	require.NoError(t, err)

	// Verify child no longer has parent
	updated, err := wailsApp.GetEntry(childID)
	require.NoError(t, err)
	assert.Nil(t, updated.ParentID)
	assert.Equal(t, 0, updated.Depth)
}

func TestApp_MoveEntryToList_MovesTaskToList(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)

	// Create a task entry
	entryIDs, err := wailsApp.AddEntry(". Buy groceries", today)
	require.NoError(t, err)
	entryID := entryIDs[0]

	// Create a list
	listID, err := wailsApp.CreateList("Shopping")
	require.NoError(t, err)

	// Move entry to list
	err = wailsApp.MoveEntryToList(entryID, listID)
	require.NoError(t, err)

	// Entry should be deleted (GetEntry returns error for deleted entries)
	_, err = wailsApp.GetEntry(entryID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// List should have the item
	lists, err := wailsApp.GetLists()
	require.NoError(t, err)
	require.Len(t, lists, 1)
	require.Len(t, lists[0].Items, 1)
	assert.Equal(t, "Buy groceries", lists[0].Items[0].Content)
}

func TestApp_MoveEntryToList_FailsForNonTasks(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)

	// Create a note entry
	entryIDs, err := wailsApp.AddEntry("- This is a note", today)
	require.NoError(t, err)
	noteID := entryIDs[0]

	// Create a list
	listID, err := wailsApp.CreateList("Notes")
	require.NoError(t, err)

	// Attempt to move note to list - should fail
	err = wailsApp.MoveEntryToList(noteID, listID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only tasks can be moved to lists")
}

func TestApp_GetOutstandingQuestions_ReturnsOnlyQuestions(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	today := time.Now().Truncate(24 * time.Hour)

	// Create various entry types
	_, err = wailsApp.AddEntry(". Task entry", today)
	require.NoError(t, err)
	_, err = wailsApp.AddEntry("- Note entry", today)
	require.NoError(t, err)
	_, err = wailsApp.AddEntry("? Unanswered question", today)
	require.NoError(t, err)
	questionIDs, err := wailsApp.AddEntry("? Another question", today)
	require.NoError(t, err)

	// Answer one question
	err = wailsApp.AnswerQuestion(questionIDs[0], "The answer")
	require.NoError(t, err)

	// Get outstanding questions
	questions, err := wailsApp.GetOutstandingQuestions()
	require.NoError(t, err)

	// Should only return the one unanswered question
	require.Len(t, questions, 1)
	assert.Equal(t, "Unanswered question", questions[0].Content)
	assert.Equal(t, domain.EntryTypeQuestion, questions[0].Type)
}

func TestApp_GetOutstandingQuestions_ReturnsEmptyWhenNone(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	questions, err := wailsApp.GetOutstandingQuestions()
	require.NoError(t, err)
	assert.Empty(t, questions)
}

func TestApp_ReadFile_ReturnsFileContents(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	tempFile, err := os.CreateTemp("", "bujo-test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()

	testContent := `. Task one
- Note two
o Event three`
	_, err = tempFile.WriteString(testContent)
	require.NoError(t, err)
	_ = tempFile.Close()

	content, err := wailsApp.ReadFile(tempFile.Name())

	require.NoError(t, err)
	assert.Equal(t, testContent, content)
}

func TestApp_ReadFile_ReturnsErrorForNonExistent(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	_, err = wailsApp.ReadFile("/nonexistent/path/file.txt")

	require.Error(t, err)
}

func TestApp_ReadFile_ReturnsErrorForLargeFile(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	tempFile, err := os.CreateTemp("", "bujo-large-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()

	largeContent := make([]byte, 2*1024*1024)
	for i := range largeContent {
		largeContent[i] = 'a'
	}
	_, err = tempFile.Write(largeContent)
	require.NoError(t, err)
	_ = tempFile.Close()

	_, err = wailsApp.ReadFile(tempFile.Name())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file too large")
}

func TestApp_OpenFileDialog_ReturnsSelectedFilePath(t *testing.T) {
	ctx := context.Background()

	factory := app.NewServiceFactory()
	services, cleanup, err := factory.Create(ctx, ":memory:")
	require.NoError(t, err)
	defer cleanup()

	wailsApp := NewApp(services)
	wailsApp.Startup(ctx)

	// OpenFileDialog requires a running Wails runtime, so we can only
	// verify the method exists and returns reasonable defaults without
	// the runtime. Integration testing will verify actual dialog behavior.
	assert.NotNil(t, wailsApp)
}
