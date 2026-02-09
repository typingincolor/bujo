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

func setupBujoServiceWithTags(t *testing.T) (*BujoService, *sqlite.EntryRepository, *sqlite.TagRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	parser := domain.NewTreeParser()

	svc := NewBujoServiceWithLists(entryRepo, dayCtxRepo, parser, nil, nil, nil, tagRepo, nil)
	return svc, entryRepo, tagRepo
}

func TestBujoService_LogEntries_StoresTags(t *testing.T) {
	svc, _, tagRepo := setupBujoServiceWithTags(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Buy groceries #shopping #errands", LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	})

	require.NoError(t, err)
	require.Len(t, ids, 1)

	tags, err := tagRepo.GetTagsForEntries(ctx, ids)
	require.NoError(t, err)
	assert.Equal(t, []string{"errands", "shopping"}, tags[ids[0]])
}

func TestBujoService_LogEntries_NoTagsNoInsert(t *testing.T) {
	svc, _, tagRepo := setupBujoServiceWithTags(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	})

	require.NoError(t, err)
	require.Len(t, ids, 1)

	tags, err := tagRepo.GetTagsForEntries(ctx, ids)
	require.NoError(t, err)
	assert.Empty(t, tags[ids[0]])
}

func TestBujoService_EditEntry_UpdatesTags(t *testing.T) {
	svc, _, tagRepo := setupBujoServiceWithTags(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Buy groceries #shopping", LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	err = svc.EditEntry(ctx, ids[0], "Buy groceries #errands #food")
	require.NoError(t, err)

	tags, err := tagRepo.GetTagsForEntries(ctx, []int64{ids[0]})
	require.NoError(t, err)
	assert.Equal(t, []string{"errands", "food"}, tags[ids[0]])
}

func TestBujoService_DeleteEntry_DeletesTags(t *testing.T) {
	svc, _, tagRepo := setupBujoServiceWithTags(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Buy groceries #shopping", LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	err = svc.DeleteEntry(ctx, ids[0])
	require.NoError(t, err)

	tags, err := tagRepo.GetTagsForEntries(ctx, ids)
	require.NoError(t, err)
	assert.Empty(t, tags[ids[0]])
}

func TestBujoService_GetAllTags(t *testing.T) {
	svc, _, _ := setupBujoServiceWithTags(t)
	ctx := context.Background()

	_, err := svc.LogEntries(ctx, ". Buy groceries #shopping #errands", LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	_, err = svc.LogEntries(ctx, ". Fix build #work", LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	tags, err := svc.GetAllTags(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"errands", "shopping", "work"}, tags)
}

func TestBujoService_SearchEntries_HydratesTags(t *testing.T) {
	svc, _, _ := setupBujoServiceWithTags(t)
	ctx := context.Background()

	_, err := svc.LogEntries(ctx, ". Buy groceries #shopping #errands", LogEntriesOptions{
		Date: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	opts := domain.NewSearchOptions("groceries")
	results, err := svc.SearchEntries(ctx, opts)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, []string{"errands", "shopping"}, results[0].Tags)
}
