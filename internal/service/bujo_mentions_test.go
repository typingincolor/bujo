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

func setupBujoServiceWithMentions(t *testing.T) (*BujoService, *sqlite.EntryRepository, *sqlite.MentionRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	mentionRepo := sqlite.NewMentionRepository(db)
	parser := domain.NewTreeParser()

	svc := NewBujoServiceWithLists(entryRepo, dayCtxRepo, parser, nil, nil, nil, tagRepo, mentionRepo)
	return svc, entryRepo, mentionRepo
}

func TestBujoService_LogEntries_StoresMentions(t *testing.T) {
	svc, _, mentionRepo := setupBujoServiceWithMentions(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Call @john.smith about project", LogEntriesOptions{
		Date: time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
	})

	require.NoError(t, err)
	require.Len(t, ids, 1)

	mentions, err := mentionRepo.GetMentionsForEntries(ctx, ids)
	require.NoError(t, err)
	assert.Equal(t, []string{"john.smith"}, mentions[ids[0]])
}

func TestBujoService_LogEntries_NoMentionsNoInsert(t *testing.T) {
	svc, _, mentionRepo := setupBujoServiceWithMentions(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Buy groceries", LogEntriesOptions{
		Date: time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
	})

	require.NoError(t, err)
	require.Len(t, ids, 1)

	mentions, err := mentionRepo.GetMentionsForEntries(ctx, ids)
	require.NoError(t, err)
	assert.Empty(t, mentions[ids[0]])
}

func TestBujoService_EditEntry_UpdatesMentions(t *testing.T) {
	svc, _, mentionRepo := setupBujoServiceWithMentions(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Call @john about project", LogEntriesOptions{
		Date: time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	err = svc.EditEntry(ctx, ids[0], "Call @alice.smith about project")
	require.NoError(t, err)

	mentions, err := mentionRepo.GetMentionsForEntries(ctx, []int64{ids[0]})
	require.NoError(t, err)
	assert.Equal(t, []string{"alice.smith"}, mentions[ids[0]])
}

func TestBujoService_DeleteEntry_DeletesMentions(t *testing.T) {
	svc, _, mentionRepo := setupBujoServiceWithMentions(t)
	ctx := context.Background()

	ids, err := svc.LogEntries(ctx, ". Call @john about project", LogEntriesOptions{
		Date: time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	err = svc.DeleteEntry(ctx, ids[0])
	require.NoError(t, err)

	mentions, err := mentionRepo.GetMentionsForEntries(ctx, ids)
	require.NoError(t, err)
	assert.Empty(t, mentions[ids[0]])
}

func TestBujoService_GetAllMentions(t *testing.T) {
	svc, _, _ := setupBujoServiceWithMentions(t)
	ctx := context.Background()

	_, err := svc.LogEntries(ctx, ". Call @john @alice", LogEntriesOptions{
		Date: time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	_, err = svc.LogEntries(ctx, ". Meet @bob", LogEntriesOptions{
		Date: time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	mentions, err := svc.GetAllMentions(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"alice", "bob", "john"}, mentions)
}

func TestBujoService_SearchEntries_HydratesMentions(t *testing.T) {
	svc, _, _ := setupBujoServiceWithMentions(t)
	ctx := context.Background()

	_, err := svc.LogEntries(ctx, ". Call @john.smith about project", LogEntriesOptions{
		Date: time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	opts := domain.NewSearchOptions("project")
	results, err := svc.SearchEntries(ctx, opts)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, []string{"john.smith"}, results[0].Mentions)
}
