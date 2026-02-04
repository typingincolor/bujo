package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestBujoService_GetAttentionScores_EmptyIDs(t *testing.T) {
	svc, _, _ := setupBujoService(t)
	ctx := context.Background()

	scores, err := svc.GetAttentionScores(ctx, nil)

	require.NoError(t, err)
	assert.Empty(t, scores)
}

func TestBujoService_GetAttentionScores_FutureTask(t *testing.T) {
	svc, _, _ := setupBujoService(t)
	ctx := context.Background()

	tomorrow := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	ids, err := svc.LogEntries(ctx, ". Plain task", LogEntriesOptions{Date: tomorrow})
	require.NoError(t, err)
	require.Len(t, ids, 1)

	scores, err := svc.GetAttentionScores(ctx, ids)

	require.NoError(t, err)
	require.Contains(t, scores, ids[0])
	assert.Equal(t, 0, scores[ids[0]].Score)
}

func TestBujoService_GetAttentionScores_OverdueTask(t *testing.T) {
	svc, _, _ := setupBujoService(t)
	ctx := context.Background()

	yesterday := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -1)
	ids, err := svc.LogEntries(ctx, ". Overdue task", LogEntriesOptions{Date: yesterday})
	require.NoError(t, err)

	scores, err := svc.GetAttentionScores(ctx, ids)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, scores[ids[0]].Score, 50)
}

func TestBujoService_GetAttentionScores_HighPriorityTask(t *testing.T) {
	svc, _, _ := setupBujoService(t)
	ctx := context.Background()

	tomorrow := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	ids, err := svc.LogEntries(ctx, ". Important task", LogEntriesOptions{Date: tomorrow})
	require.NoError(t, err)

	err = svc.EditEntryPriority(ctx, ids[0], domain.PriorityHigh)
	require.NoError(t, err)

	scores, err := svc.GetAttentionScores(ctx, ids)

	require.NoError(t, err)
	assert.Equal(t, 50, scores[ids[0]].Score)
}

func TestBujoService_GetAttentionScores_ChildOfEvent(t *testing.T) {
	svc, _, _ := setupBujoService(t)
	ctx := context.Background()

	tomorrow := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	eventIDs, err := svc.LogEntries(ctx, "o Team meeting", LogEntriesOptions{Date: tomorrow})
	require.NoError(t, err)

	childIDs, err := svc.LogEntries(ctx, ". Follow up action", LogEntriesOptions{Date: tomorrow, ParentID: &eventIDs[0]})
	require.NoError(t, err)

	scores, err := svc.GetAttentionScores(ctx, childIDs)

	require.NoError(t, err)
	assert.Equal(t, 5, scores[childIDs[0]].Score)
}

func TestBujoService_GetAttentionScores_MultipleEntries(t *testing.T) {
	svc, _, _ := setupBujoService(t)
	ctx := context.Background()

	tomorrow := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	ids, err := svc.LogEntries(ctx, ". Task one\n. Task two\n. Task three", LogEntriesOptions{Date: tomorrow})
	require.NoError(t, err)
	require.Len(t, ids, 3)

	scores, err := svc.GetAttentionScores(ctx, ids)

	require.NoError(t, err)
	assert.Len(t, scores, 3)
	for _, id := range ids {
		assert.Contains(t, scores, id)
	}
}

func TestBujoService_GetAttentionScores_SkipsMissingEntries(t *testing.T) {
	svc, _, _ := setupBujoService(t)
	ctx := context.Background()

	scores, err := svc.GetAttentionScores(ctx, []int64{99999})

	require.NoError(t, err)
	assert.Empty(t, scores)
}
