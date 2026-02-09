package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMentionRepository_InsertEntryMentions(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	mentionRepo := NewMentionRepository(db)
	ctx := context.Background()

	entryID := insertTestEntry(t, entryRepo, "Call @john.smith")

	err := mentionRepo.InsertEntryMentions(ctx, entryID, []string{"john.smith"})

	require.NoError(t, err)
}

func TestMentionRepository_InsertEntryMentions_Empty(t *testing.T) {
	db := setupTestDB(t)
	mentionRepo := NewMentionRepository(db)
	ctx := context.Background()

	err := mentionRepo.InsertEntryMentions(ctx, 1, nil)

	require.NoError(t, err)
}

func TestMentionRepository_GetMentionsForEntries(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	mentionRepo := NewMentionRepository(db)
	ctx := context.Background()

	id1 := insertTestEntry(t, entryRepo, "Call @john")
	id2 := insertTestEntry(t, entryRepo, "Meet @alice.smith")

	require.NoError(t, mentionRepo.InsertEntryMentions(ctx, id1, []string{"john"}))
	require.NoError(t, mentionRepo.InsertEntryMentions(ctx, id2, []string{"alice.smith"}))

	mentionsMap, err := mentionRepo.GetMentionsForEntries(ctx, []int64{id1, id2})

	require.NoError(t, err)
	assert.Equal(t, []string{"john"}, mentionsMap[id1])
	assert.Equal(t, []string{"alice.smith"}, mentionsMap[id2])
}

func TestMentionRepository_GetMentionsForEntries_Empty(t *testing.T) {
	db := setupTestDB(t)
	mentionRepo := NewMentionRepository(db)
	ctx := context.Background()

	mentionsMap, err := mentionRepo.GetMentionsForEntries(ctx, nil)

	require.NoError(t, err)
	assert.Empty(t, mentionsMap)
}

func TestMentionRepository_GetAllMentions(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	mentionRepo := NewMentionRepository(db)
	ctx := context.Background()

	id1 := insertTestEntry(t, entryRepo, "Entry 1")
	id2 := insertTestEntry(t, entryRepo, "Entry 2")

	require.NoError(t, mentionRepo.InsertEntryMentions(ctx, id1, []string{"john", "alice"}))
	require.NoError(t, mentionRepo.InsertEntryMentions(ctx, id2, []string{"bob", "john"}))

	mentions, err := mentionRepo.GetAllMentions(ctx)

	require.NoError(t, err)
	assert.Equal(t, []string{"alice", "bob", "john"}, mentions)
}

func TestMentionRepository_GetAllMentions_Empty(t *testing.T) {
	db := setupTestDB(t)
	mentionRepo := NewMentionRepository(db)
	ctx := context.Background()

	mentions, err := mentionRepo.GetAllMentions(ctx)

	require.NoError(t, err)
	assert.Empty(t, mentions)
}

func TestMentionRepository_DeleteByEntryID(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	mentionRepo := NewMentionRepository(db)
	ctx := context.Background()

	id1 := insertTestEntry(t, entryRepo, "Entry 1")
	id2 := insertTestEntry(t, entryRepo, "Entry 2")

	require.NoError(t, mentionRepo.InsertEntryMentions(ctx, id1, []string{"john"}))
	require.NoError(t, mentionRepo.InsertEntryMentions(ctx, id2, []string{"alice"}))

	err := mentionRepo.DeleteByEntryID(ctx, id1)
	require.NoError(t, err)

	mentionsMap, err := mentionRepo.GetMentionsForEntries(ctx, []int64{id1, id2})
	require.NoError(t, err)
	assert.Empty(t, mentionsMap[id1])
	assert.Equal(t, []string{"alice"}, mentionsMap[id2])
}
