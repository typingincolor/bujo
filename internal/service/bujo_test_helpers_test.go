package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
)

func setupBujoService(t *testing.T) (*BujoService, *sqlite.EntryRepository, *sqlite.DayContextRepository) {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	parser := domain.NewTreeParser()

	service := NewBujoService(entryRepo, dayCtxRepo, parser)
	return service, entryRepo, dayCtxRepo
}
