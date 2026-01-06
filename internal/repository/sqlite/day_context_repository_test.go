package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func stringPtr(s string) *string {
	return &s
}

func TestDayContextRepository_Upsert_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDayContextRepository(db)
	ctx := context.Background()

	dayCtx := domain.DayContext{
		Date:     time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
		Location: stringPtr("Manchester Office"),
		Mood:     stringPtr("productive"),
	}

	err := repo.Upsert(ctx, dayCtx)

	require.NoError(t, err)

	result, err := repo.GetByDate(ctx, dayCtx.Date)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Manchester Office", *result.Location)
}

func TestDayContextRepository_Upsert_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDayContextRepository(db)
	ctx := context.Background()

	date := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	dayCtx1 := domain.DayContext{
		Date:     date,
		Location: stringPtr("Home"),
	}
	err := repo.Upsert(ctx, dayCtx1)
	require.NoError(t, err)

	dayCtx2 := domain.DayContext{
		Date:     date,
		Location: stringPtr("Office"),
		Mood:     stringPtr("focused"),
	}
	err = repo.Upsert(ctx, dayCtx2)
	require.NoError(t, err)

	result, err := repo.GetByDate(ctx, date)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Office", *result.Location)
	assert.Equal(t, "focused", *result.Mood)
}

func TestDayContextRepository_GetByDate_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDayContextRepository(db)
	ctx := context.Background()

	result, err := repo.GetByDate(ctx, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestDayContextRepository_GetRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDayContextRepository(db)
	ctx := context.Background()

	dates := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), // Outside range
	}

	for i, date := range dates {
		err := repo.Upsert(ctx, domain.DayContext{
			Date:     date,
			Location: stringPtr("Location " + string(rune('A'+i))),
		})
		require.NoError(t, err)
	}

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	results, err := repo.GetRange(ctx, start, end)

	require.NoError(t, err)
	assert.Len(t, results, 3)
}
