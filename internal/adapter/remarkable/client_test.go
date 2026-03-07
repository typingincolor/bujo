package remarkable

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token/json/2/device/new", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("fake-device-token-jwt"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	token, err := client.RegisterDevice("abcd1234")
	require.NoError(t, err)
	assert.Equal(t, "fake-device-token-jwt", token)
}

func TestRefreshUserToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token/json/2/user/new", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer fake-device-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("fake-user-token-jwt"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	token, err := client.RefreshUserToken("fake-device-token")
	require.NoError(t, err)
	assert.Equal(t, "fake-user-token-jwt", token)
}

func TestRegisterDeviceFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.RegisterDevice("badcode1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestDownloadDocument(t *testing.T) {
	blobContent := []byte("fake-zip-content")

	blobServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(blobContent)
	}))
	defer blobServer.Close()

	doc := Document{
		ID:         "doc-1",
		BlobURLGet: blobServer.URL + "/blob",
	}

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			w.Write([]byte("user-token"))
			return
		}
		assert.Equal(t, "/doc/v2/files", r.URL.Path)
		assert.Equal(t, "doc-1", r.URL.Query().Get("doc"))
		assert.Equal(t, "true", r.URL.Query().Get("withBlob"))
		docs, _ := json.Marshal([]Document{doc})
		w.Write(docs)
	}))
	defer apiServer.Close()

	client := NewClient(apiServer.URL)
	client.syncHost = apiServer.URL

	data, err := client.DownloadDocument("fake-device-token", "doc-1")
	require.NoError(t, err)
	assert.Equal(t, blobContent, data)
}

func TestListDocuments(t *testing.T) {
	docs := []Document{
		{ID: "doc-1", VisibleName: "Meeting Notes", Type: "DocumentType", ModifiedAt: "2026-03-01"},
		{ID: "doc-2", VisibleName: "Journal", Type: "DocumentType", ModifiedAt: "2026-03-02"},
	}
	docsJSON, _ := json.Marshal(docs)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			w.Write([]byte("user-token"))
			return
		}
		assert.Equal(t, "/doc/v2/files", r.URL.Path)
		assert.Equal(t, "Bearer user-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.Write(docsJSON)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.syncHost = server.URL

	result, err := client.ListDocuments("fake-device-token")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Meeting Notes", result[0].VisibleName)
}
