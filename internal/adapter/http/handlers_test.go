package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository/sqlite"
	"github.com/typingincolor/bujo/internal/service"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := sqlite.OpenAndMigrate(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	entryRepo := sqlite.NewEntryRepository(db)
	dayCtxRepo := sqlite.NewDayContextRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	mentionRepo := sqlite.NewMentionRepository(db)
	parser := domain.NewTreeParser()

	bujoService := service.NewBujoServiceWithLists(entryRepo, dayCtxRepo, parser, nil, nil, nil, tagRepo, mentionRepo)

	handler := NewHandler(bujoService)
	server := httptest.NewServer(handler.Routes())
	t.Cleanup(server.Close)

	return server
}

func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer(t)

	resp, err := http.Get(server.URL + "/api/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}

func TestCreateEntries(t *testing.T) {
	server := setupTestServer(t)

	payload := createEntriesRequest{
		Entries: []entryInput{
			{Type: "task", Content: "Follow up: Q1 Planning @john #email"},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/entries", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result createEntriesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	require.Len(t, result.Entries, 1)
	assert.Greater(t, result.Entries[0].ID, int64(0))
}

func TestCreateEntriesWithChildren(t *testing.T) {
	server := setupTestServer(t)

	payload := createEntriesRequest{
		Entries: []entryInput{
			{
				Type:    "task",
				Content: "Follow up: Q1 Planning @john #email",
				Children: []entryInput{
					{Type: "note", Content: "Context: Thanks for the meeting yesterday"},
					{Type: "note", Content: "Email: https://mail.google.com/mail/u/0/#inbox/123"},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/entries", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result createEntriesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	require.Len(t, result.Entries, 1)

	parent := result.Entries[0]
	assert.Greater(t, parent.ID, int64(0))
	require.Len(t, parent.Children, 2)
	assert.Greater(t, parent.Children[0].ID, int64(0))
	assert.Greater(t, parent.Children[1].ID, int64(0))

	assert.NotEqual(t, parent.ID, parent.Children[0].ID)
	assert.NotEqual(t, parent.ID, parent.Children[1].ID)
	assert.NotEqual(t, parent.Children[0].ID, parent.Children[1].ID)
}

func TestCreateEntriesMissingContent(t *testing.T) {
	server := setupTestServer(t)

	payload := createEntriesRequest{
		Entries: []entryInput{
			{Type: "task", Content: ""},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/entries", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result createEntriesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "content")
}

func TestCreateEntriesInvalidType(t *testing.T) {
	server := setupTestServer(t)

	payload := createEntriesRequest{
		Entries: []entryInput{
			{Type: "invalid", Content: "Some content"},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/entries", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result createEntriesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "entry type")
}

func TestCORSPreflight(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest("OPTIONS", server.URL+"/api/entries", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://mail.google.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "https://mail.google.com", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestCORSHeadersOnResponse(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest("GET", server.URL+"/api/health", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://mail.google.com")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "https://mail.google.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest("GET", server.URL+"/api/health", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://evil.com")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestBuildInputStripsNewlines(t *testing.T) {
	entries := []entryInput{
		{
			Type:    "task",
			Content: "Follow up: Payment receipt @service #email",
			Children: []entryInput{
				{Type: "note", Content: "Context: Hi Andrew,\nYou paid £29.99\nMerchant: LinkedIn"},
				{Type: "note", Content: "Email: https://mail.google.com/mail/u/0/#inbox/123"},
			},
		},
	}

	input, childCounts, err := buildInput(entries)
	require.NoError(t, err)

	lines := strings.Split(input, "\n")
	assert.Equal(t, 3, len(lines), "should produce exactly 3 lines (1 parent + 2 children), got: %v", lines)
	assert.Equal(t, []int{2}, childCounts)
}

func TestInstallPage(t *testing.T) {
	server := setupTestServer(t)

	resp, err := http.Get(server.URL + "/install")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Gmail")
	assert.Contains(t, string(body), "Bujo")
}
