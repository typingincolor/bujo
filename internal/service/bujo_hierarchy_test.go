package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func TestBujoService_LogEntries_WithParentID(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	// Create a parent entry
	parentIDs, err := service.LogEntries(ctx, ". Parent task", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, parentIDs, 1)

	// Add a child entry using the ParentID option
	childIDs, err := service.LogEntries(ctx, ". Child task", LogEntriesOptions{
		Date:     today,
		ParentID: &parentIDs[0],
	})
	require.NoError(t, err)
	require.Len(t, childIDs, 1)

	// Verify the child has the correct parent
	child, err := entryRepo.GetByID(ctx, childIDs[0])
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, parentIDs[0], *child.ParentID)
	assert.Equal(t, 1, child.Depth)
}

func TestBujoService_LogEntries_WithParentID_MultipleEntries(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	// Create a parent entry
	parentIDs, err := service.LogEntries(ctx, ". Parent task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Add multiple child entries
	childIDs, err := service.LogEntries(ctx, `. Child task 1
- Child note`, LogEntriesOptions{
		Date:     today,
		ParentID: &parentIDs[0],
	})
	require.NoError(t, err)
	require.Len(t, childIDs, 2)

	// Both should have parent as their root parent
	child1, err := entryRepo.GetByID(ctx, childIDs[0])
	require.NoError(t, err)
	require.NotNil(t, child1.ParentID)
	assert.Equal(t, parentIDs[0], *child1.ParentID)
	assert.Equal(t, 1, child1.Depth)

	child2, err := entryRepo.GetByID(ctx, childIDs[1])
	require.NoError(t, err)
	require.NotNil(t, child2.ParentID)
	assert.Equal(t, parentIDs[0], *child2.ParentID)
	assert.Equal(t, 1, child2.Depth)
}

func TestBujoService_LogEntries_WithParentID_NestedInput(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	// Create a parent entry
	parentIDs, err := service.LogEntries(ctx, ". Parent task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	// Add nested entries with parent ID - the internal nesting should be relative
	childIDs, err := service.LogEntries(ctx, `. Child task
  - Grandchild note`, LogEntriesOptions{
		Date:     today,
		ParentID: &parentIDs[0],
	})
	require.NoError(t, err)
	require.Len(t, childIDs, 2)

	// Child should have depth 1 (parent's depth + 1)
	child, err := entryRepo.GetByID(ctx, childIDs[0])
	require.NoError(t, err)
	assert.Equal(t, 1, child.Depth)
	assert.Equal(t, parentIDs[0], *child.ParentID)

	// Grandchild should have depth 2
	grandchild, err := entryRepo.GetByID(ctx, childIDs[1])
	require.NoError(t, err)
	assert.Equal(t, 2, grandchild.Depth)
	assert.Equal(t, childIDs[0], *grandchild.ParentID)
}

func TestBujoService_LogEntries_WithParentID_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	invalidParent := int64(99999)

	_, err := service.LogEntries(ctx, ". Child task", LogEntriesOptions{
		Date:     today,
		ParentID: &invalidParent,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent")
}

func TestBujoService_LogEntries_CannotAddChildToQuestion(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	// Create a question entry
	questionIDs, err := service.LogEntries(ctx, "? What is the answer", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, questionIDs, 1)

	// Try to add a child to the question - should fail
	_, err = service.LogEntries(ctx, ". Child task", LogEntriesOptions{
		Date:     today,
		ParentID: &questionIDs[0],
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add children to questions")
}

func TestBujoService_GetEntryAncestors_ReturnsParentChain(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, `. Grandparent
  - Parent
    . Grandchild`, LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 3)

	ancestors, err := service.GetEntryAncestors(ctx, ids[2])
	require.NoError(t, err)

	require.Len(t, ancestors, 2)
	assert.Equal(t, "Grandparent", ancestors[0].Content)
	assert.Equal(t, "Parent", ancestors[1].Content)
}

func TestBujoService_GetEntryAncestors_RootEntry(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Root task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	ancestors, err := service.GetEntryAncestors(ctx, ids[0])
	require.NoError(t, err)
	assert.Empty(t, ancestors)
}

func TestBujoService_GetEntryAncestors_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	_, err := service.GetEntryAncestors(ctx, 99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_MarkAnswered_QuestionBecomesAnswered(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "? What is the meaning of life", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = service.MarkAnswered(ctx, ids[0], "42")
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeAnswered, entry.Type)
}

func TestBujoService_MarkAnswered_CreatesAnswerEntry(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "? What is the meaning of life", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkAnswered(ctx, ids[0], "42")
	require.NoError(t, err)

	// Get the question after marking as answered to get its current ID
	// (Update creates a new version with a new ID due to event sourcing)
	question, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	require.Equal(t, domain.EntryTypeAnswered, question.Type)

	// Get children using the current question ID
	children, err := entryRepo.GetChildren(ctx, question.ID)
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, "42", children[0].Content)
	assert.Equal(t, domain.EntryTypeAnswer, children[0].Type)
}

func TestBujoService_MarkAnswered_RequiresAnswerText(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "? What is the meaning of life", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkAnswered(ctx, ids[0], "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "answer text is required")
}

func TestBujoService_MarkAnswered_OnlyQuestions(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". This is a task", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkAnswered(ctx, ids[0], "Some answer")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only questions")
}

func TestBujoService_MarkAnswered_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.MarkAnswered(ctx, 99999, "Some answer")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_ReopenQuestion_AnsweredBecomesQuestion(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "? What is the meaning of life", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MarkAnswered(ctx, ids[0], "42")
	require.NoError(t, err)

	err = service.ReopenQuestion(ctx, ids[0])
	require.NoError(t, err)

	entry, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeQuestion, entry.Type)
}

func TestBujoService_ReopenQuestion_NotFound(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	err := service.ReopenQuestion(ctx, 99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_CancelAnswer_ReopensQuestion(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "? What is the meaning of life", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = service.MarkAnswered(ctx, ids[0], "42")
	require.NoError(t, err)

	question, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	require.Equal(t, domain.EntryTypeAnswered, question.Type)

	children, err := entryRepo.GetChildren(ctx, question.ID)
	require.NoError(t, err)
	require.Len(t, children, 1)
	answerID := children[0].ID

	err = service.CancelEntry(ctx, answerID)
	require.NoError(t, err)

	question, err = entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeQuestion, question.Type)
}

func TestBujoService_DeleteAnswer_ReopensQuestion(t *testing.T) {
	service, entryRepo, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "? What is the meaning of life", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	err = service.MarkAnswered(ctx, ids[0], "42")
	require.NoError(t, err)

	question, err := entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	require.Equal(t, domain.EntryTypeAnswered, question.Type)

	children, err := entryRepo.GetChildren(ctx, question.ID)
	require.NoError(t, err)
	require.Len(t, children, 1)
	answerID := children[0].ID

	err = service.DeleteEntry(ctx, answerID)
	require.NoError(t, err)

	question, err = entryRepo.GetByID(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, domain.EntryTypeQuestion, question.Type)
}

func TestBujoService_GetDailyAgenda_WithMoodAndWeather(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, today, "Home Office")
	require.NoError(t, err)

	err = service.SetMood(ctx, today, "Focused")
	require.NoError(t, err)

	err = service.SetWeather(ctx, today, "Sunny")
	require.NoError(t, err)

	agenda, err := service.GetDailyAgenda(ctx, today)

	require.NoError(t, err)
	require.NotNil(t, agenda.Location)
	assert.Equal(t, "Home Office", *agenda.Location)
	require.NotNil(t, agenda.Mood)
	assert.Equal(t, "Focused", *agenda.Mood)
	require.NotNil(t, agenda.Weather)
	assert.Equal(t, "Sunny", *agenda.Weather)
}

func TestBujoService_GetMultiDayAgenda_WithMoodAndWeather(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	day1 := time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)

	err := service.SetLocation(ctx, day1, "Home")
	require.NoError(t, err)
	err = service.SetMood(ctx, day1, "Energetic")
	require.NoError(t, err)
	err = service.SetWeather(ctx, day1, "Cloudy")
	require.NoError(t, err)

	err = service.SetLocation(ctx, day2, "Office")
	require.NoError(t, err)
	err = service.SetMood(ctx, day2, "Calm")
	require.NoError(t, err)
	err = service.SetWeather(ctx, day2, "Rainy")
	require.NoError(t, err)

	agenda, err := service.GetMultiDayAgenda(ctx, day1, day2)

	require.NoError(t, err)
	require.Len(t, agenda.Days, 2)

	assert.NotNil(t, agenda.Days[0].Location)
	assert.Equal(t, "Home", *agenda.Days[0].Location)
	assert.NotNil(t, agenda.Days[0].Mood)
	assert.Equal(t, "Energetic", *agenda.Days[0].Mood)
	assert.NotNil(t, agenda.Days[0].Weather)
	assert.Equal(t, "Cloudy", *agenda.Days[0].Weather)

	assert.NotNil(t, agenda.Days[1].Location)
	assert.Equal(t, "Office", *agenda.Days[1].Location)
	assert.NotNil(t, agenda.Days[1].Mood)
	assert.Equal(t, "Calm", *agenda.Days[1].Mood)
	assert.NotNil(t, agenda.Days[1].Weather)
	assert.Equal(t, "Rainy", *agenda.Days[1].Weather)
}

func TestBujoService_ExportEntryMarkdown_SingleEntry(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	date := time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Test task", LogEntriesOptions{Date: date})
	require.NoError(t, err)

	markdown, err := service.ExportEntryMarkdown(ctx, ids[0])

	require.NoError(t, err)
	assert.Contains(t, markdown, "Test task")
	assert.Contains(t, markdown, "•")
}

func TestBujoService_ExportEntryMarkdown_WithChildren(t *testing.T) {
	service, _, _ := setupBujoService(t)
	ctx := context.Background()

	date := time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC)
	input := `. Parent task
  - Child note
  . Child task`
	ids, err := service.LogEntries(ctx, input, LogEntriesOptions{Date: date})
	require.NoError(t, err)

	markdown, err := service.ExportEntryMarkdown(ctx, ids[0])

	require.NoError(t, err)
	assert.Contains(t, markdown, "Parent task")
	assert.Contains(t, markdown, "Child note")
	assert.Contains(t, markdown, "Child task")
	assert.Contains(t, markdown, "  –")
	assert.Contains(t, markdown, "  •")
}

func setupBujoServiceWithLists(t *testing.T) (*BujoService, *sqlite.EntryRepository, *sqlite.ListRepository, *sqlite.ListItemRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	listRepo := sqlite.NewListRepository(db)
	listItemRepo := sqlite.NewListItemRepository(db)
	entryToListMover := sqlite.NewEntryToListMover(db)
	parser := domain.NewTreeParser()

	service := NewBujoServiceWithLists(entryRepo, dayCtxRepo, parser, listRepo, listItemRepo, entryToListMover, nil)
	return service, entryRepo, listRepo, listItemRepo
}

func TestBujoService_MoveEntryToList_Success(t *testing.T) {
	service, entryRepo, listRepo, listItemRepo := setupBujoServiceWithLists(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	entryID := ids[0]

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	err = service.MoveEntryToList(ctx, entryID, list.ID)

	require.NoError(t, err)

	// Entry should be deleted
	entry, err := entryRepo.GetByID(ctx, entryID)
	require.NoError(t, err)
	assert.Nil(t, entry)

	// List item should exist with same content
	items, err := listItemRepo.GetByListID(ctx, list.ID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Buy groceries", items[0].Content)
	assert.Equal(t, domain.ListItemTypeTask, items[0].Type)
}

func TestBujoService_MoveEntryToList_OnlyTasks(t *testing.T) {
	service, _, listRepo, _ := setupBujoServiceWithLists(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, "- A note", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	noteID := ids[0]

	list, err := listRepo.Create(ctx, "Notes")
	require.NoError(t, err)

	err = service.MoveEntryToList(ctx, noteID, list.ID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only tasks can be moved to lists")
}

func TestBujoService_MoveEntryToList_EntryNotFound(t *testing.T) {
	service, _, listRepo, _ := setupBujoServiceWithLists(t)
	ctx := context.Background()

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	err = service.MoveEntryToList(ctx, 9999, list.ID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBujoService_MoveEntryToList_ListNotFound(t *testing.T) {
	service, _, _, _ := setupBujoServiceWithLists(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{Date: today})
	require.NoError(t, err)

	err = service.MoveEntryToList(ctx, ids[0], 9999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list not found")
}

func TestBujoService_MoveEntryToList_WithChildren_Fails(t *testing.T) {
	service, _, listRepo, _ := setupBujoServiceWithLists(t)
	ctx := context.Background()

	today := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	ids, err := service.LogEntries(ctx, ". Parent task\n  . Child task", LogEntriesOptions{Date: today})
	require.NoError(t, err)
	require.Len(t, ids, 2)

	parentID := ids[0]

	list, err := listRepo.Create(ctx, "Test List")
	require.NoError(t, err)

	err = service.MoveEntryToList(ctx, parentID, list.ID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot move entry with children")
}
