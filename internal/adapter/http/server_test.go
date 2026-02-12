package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

func TestServerStartStop(t *testing.T) {
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	mentionRepo := sqlite.NewMentionRepository(db)
	parser := domain.NewTreeParser()
	bujoService := service.NewBujoServiceWithLists(entryRepo, dayCtxRepo, parser, nil, nil, nil, tagRepo, mentionRepo)

	srv := NewServer(bujoService, -1)
	addr, err := srv.Start()
	require.NoError(t, err)
	require.NotEmpty(t, addr)

	resp, err := http.Get("http://" + addr + "/api/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])

	err = srv.Stop()
	assert.NoError(t, err)

	_, err = http.Get("http://" + addr + "/api/health")
	assert.Error(t, err)
}
