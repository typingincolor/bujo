package sqlite

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAndMigrate(t *testing.T) {
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	tables := []string{"entries", "habits", "habit_logs", "day_context", "summaries", "lists"}
	for _, table := range tables {
		t.Run("table_"+table+"_exists", func(t *testing.T) {
			var name string
			err := db.QueryRow(
				"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
				table,
			).Scan(&name)
			require.NoError(t, err)
			assert.Equal(t, table, name)
		})
	}
}

func TestOpen_EnablesForeignKeys(t *testing.T) {
	db, err := Open(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	require.NoError(t, err)
	assert.Equal(t, 1, fkEnabled)
}

func TestOpen_EnablesWAL(t *testing.T) {
	db, err := Open(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "memory", journalMode)
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}
