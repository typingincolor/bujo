package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func insertTestEntry(t *testing.T, repo *EntryRepository, content string) int64 {
	t.Helper()
	id, err := repo.Insert(context.Background(), domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   content,
		Depth:     0,
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	return id
}

func TestTagRepository_InsertEntryTags(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	entryID := insertTestEntry(t, entryRepo, "Buy groceries")

	err := tagRepo.InsertEntryTags(ctx, entryID, []string{"shopping", "errands"})

	require.NoError(t, err)
}

func TestTagRepository_InsertEntryTags_Empty(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	err := tagRepo.InsertEntryTags(ctx, 1, nil)

	require.NoError(t, err)
}

func TestTagRepository_GetTagsForEntries(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	id1 := insertTestEntry(t, entryRepo, "Buy groceries")
	id2 := insertTestEntry(t, entryRepo, "Fix build")

	require.NoError(t, tagRepo.InsertEntryTags(ctx, id1, []string{"shopping", "errands"}))
	require.NoError(t, tagRepo.InsertEntryTags(ctx, id2, []string{"work", "urgent"}))

	tagsMap, err := tagRepo.GetTagsForEntries(ctx, []int64{id1, id2})

	require.NoError(t, err)
	assert.Equal(t, []string{"errands", "shopping"}, tagsMap[id1])
	assert.Equal(t, []string{"urgent", "work"}, tagsMap[id2])
}

func TestTagRepository_GetTagsForEntries_Empty(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	tagsMap, err := tagRepo.GetTagsForEntries(ctx, nil)

	require.NoError(t, err)
	assert.Empty(t, tagsMap)
}

func TestTagRepository_GetAllTags(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	id1 := insertTestEntry(t, entryRepo, "Entry 1")
	id2 := insertTestEntry(t, entryRepo, "Entry 2")

	require.NoError(t, tagRepo.InsertEntryTags(ctx, id1, []string{"shopping", "errands"}))
	require.NoError(t, tagRepo.InsertEntryTags(ctx, id2, []string{"work", "shopping"}))

	tags, err := tagRepo.GetAllTags(ctx)

	require.NoError(t, err)
	assert.Equal(t, []string{"errands", "shopping", "work"}, tags)
}

func TestTagRepository_GetAllTags_Empty(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	tags, err := tagRepo.GetAllTags(ctx)

	require.NoError(t, err)
	assert.Empty(t, tags)
}

func TestTagRepository_DeleteByEntryID(t *testing.T) {
	db := setupTestDB(t)
	entryRepo := NewEntryRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	id1 := insertTestEntry(t, entryRepo, "Entry 1")
	id2 := insertTestEntry(t, entryRepo, "Entry 2")

	require.NoError(t, tagRepo.InsertEntryTags(ctx, id1, []string{"shopping"}))
	require.NoError(t, tagRepo.InsertEntryTags(ctx, id2, []string{"work"}))

	err := tagRepo.DeleteByEntryID(ctx, id1)
	require.NoError(t, err)

	tagsMap, err := tagRepo.GetTagsForEntries(ctx, []int64{id1, id2})
	require.NoError(t, err)
	assert.Empty(t, tagsMap[id1])
	assert.Equal(t, []string{"work"}, tagsMap[id2])
}
