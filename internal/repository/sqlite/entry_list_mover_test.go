package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestEntryToListMover_MoveEntryToList_Atomic(t *testing.T) {
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	defer db.Close()

	entryRepo := NewEntryRepository(db)
	listRepo := NewListRepository(db)
	listItemRepo := NewListItemRepository(db)
	mover := NewEntryToListMover(db)
	ctx := context.Background()

	entry := domain.Entry{
		Type:      domain.EntryTypeTask,
		Content:   "Buy groceries",
		Priority:  domain.PriorityNone,
		CreatedAt: time.Now(),
	}
	entryID, err := entryRepo.Insert(ctx, entry)
	require.NoError(t, err)

	savedEntry, err := entryRepo.GetByID(ctx, entryID)
	require.NoError(t, err)
	require.NotNil(t, savedEntry)

	list, err := listRepo.Create(ctx, "Shopping")
	require.NoError(t, err)

	err = mover.MoveEntryToList(ctx, *savedEntry, list.EntityID)
	require.NoError(t, err)

	deletedEntry, err := entryRepo.GetByID(ctx, entryID)
	require.NoError(t, err)
	assert.Nil(t, deletedEntry, "entry should be deleted")

	items, err := listItemRepo.GetByListEntityID(ctx, list.EntityID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Buy groceries", items[0].Content)
}
