package remarkable

import (
	"encoding/json"
	"fmt"
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
		_, _ = w.Write([]byte("fake-device-token-jwt"))
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
		_, _ = w.Write([]byte("fake-user-token-jwt"))
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
	rootEntries := "3\ndocHash1:80000000:doc-id-1:3:1024\n"
	docEntries := "3\nmetaHash:0:doc-id-1.metadata:0:100\ncontentHash:0:doc-id-1.content:0:50\npdfHash:0:doc-id-1.pdf:0:500\n"
	meta := `{"visibleName":"Test Doc","lastModified":"1709251200000","parent":"","pinned":false}`
	pdfContent := []byte("fake-pdf-content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			_, _ = w.Write([]byte("user-token"))
			return
		}
		switch r.URL.Path {
		case "/sync/v4/root":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"hash": "rootHash", "generation": 1, "schemaVersion": 3,
			})
		case "/sync/v3/files/rootHash":
			_, _ = w.Write([]byte(rootEntries))
		case "/sync/v3/files/docHash1":
			_, _ = w.Write([]byte(docEntries))
		case "/sync/v3/files/metaHash":
			_, _ = w.Write([]byte(meta))
		case "/sync/v3/files/pdfHash":
			_, _ = w.Write(pdfContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetSyncHost(server.URL)

	data, err := client.DownloadDocument("fake-device-token", "doc-id-1")
	require.NoError(t, err)
	assert.Equal(t, pdfContent, data)
}

func TestGetRootHash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			_, _ = w.Write([]byte("user-token"))
			return
		}
		assert.Equal(t, "/sync/v4/root", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer user-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"hash":          "abc123def456",
			"generation":    42,
			"schemaVersion": 3,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetSyncHost(server.URL)

	hash, gen, err := client.GetRootHash("fake-device-token")
	require.NoError(t, err)
	assert.Equal(t, "abc123def456", hash)
	assert.Equal(t, 42, gen)
}

func TestGetEntries(t *testing.T) {
	entriesContent := "3\nhashA:80000000:doc-id-1:5:1024\nhashB:80000000:doc-id-2:3:512\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			_, _ = w.Write([]byte("user-token"))
			return
		}
		assert.Equal(t, "/sync/v3/files/root-hash-abc", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		_, _ = w.Write([]byte(entriesContent))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetSyncHost(server.URL)

	entries, err := client.GetEntries("user-token", "root-hash-abc")
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "doc-id-1", entries[0].ID)
	assert.Equal(t, "hashA", entries[0].Hash)
	assert.Equal(t, "doc-id-2", entries[1].ID)
	assert.Equal(t, "hashB", entries[1].Hash)
}

func TestListDocuments(t *testing.T) {
	rootEntries := "3\ndocHash1:80000000:doc-id-1:5:1024\ndocHash2:80000000:doc-id-2:3:512\n"
	doc1Entries := "3\nmetaHash1:0:doc-id-1.metadata:0:100\ncontentHash1:0:doc-id-1.content:0:50\n"
	doc2Entries := "3\nmetaHash2:0:doc-id-2.metadata:0:100\ncontentHash2:0:doc-id-2.content:0:50\n"
	meta1 := `{"visibleName":"Meeting Notes","lastModified":"1709251200000","parent":"","pinned":false}`
	meta2 := `{"visibleName":"Journal","lastModified":"1709337600000","parent":"folder-1","pinned":false}`
	content1 := `{"fileType":"notebook"}`
	content2 := `{"fileType":"pdf"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			_, _ = w.Write([]byte("user-token"))
			return
		}
		switch r.URL.Path {
		case "/sync/v4/root":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"hash": "rootHash", "generation": 1, "schemaVersion": 3,
			})
		case "/sync/v3/files/rootHash":
			_, _ = w.Write([]byte(rootEntries))
		case "/sync/v3/files/docHash1":
			_, _ = w.Write([]byte(doc1Entries))
		case "/sync/v3/files/docHash2":
			_, _ = w.Write([]byte(doc2Entries))
		case "/sync/v3/files/metaHash1":
			_, _ = w.Write([]byte(meta1))
		case "/sync/v3/files/metaHash2":
			_, _ = w.Write([]byte(meta2))
		case "/sync/v3/files/contentHash1":
			_, _ = w.Write([]byte(content1))
		case "/sync/v3/files/contentHash2":
			_, _ = w.Write([]byte(content2))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetSyncHost(server.URL)

	docs, err := client.ListDocuments("fake-device-token")
	require.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.Equal(t, "doc-id-1", docs[0].ID)
	assert.Equal(t, "Meeting Notes", docs[0].VisibleName)
	assert.Equal(t, "notebook", docs[0].FileType)
	assert.Equal(t, "doc-id-2", docs[1].ID)
	assert.Equal(t, "Journal", docs[1].VisibleName)
	assert.Equal(t, "folder-1", docs[1].Parent)
	assert.Equal(t, "pdf", docs[1].FileType)
}

func TestDownloadPages(t *testing.T) {
	page1Content := []byte("rm-page-1-binary")
	page2Content := []byte("rm-page-2-binary")
	docID := "doc-uuid-1"

	contentJSON := `{"cPages":{"pages":[{"id":"page-a"},{"id":"page-b"}]},"fileType":"notebook"}`

	rootEntries := fmt.Sprintf("3\ndocHash:%s:%s:5:1024\n", "80000000", docID)
	docEntries := fmt.Sprintf("3\ncontentHash:0:%s.content:0:100\nmetaHash:0:%s.metadata:0:50\npageAHash:0:%s/page-a.rm:0:200\npageBHash:0:%s/page-b.rm:0:300\n",
		docID, docID, docID, docID)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token/json/2/user/new" {
			_, _ = w.Write([]byte("user-token"))
			return
		}
		switch r.URL.Path {
		case "/sync/v4/root":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"hash": "rootHash", "generation": 1, "schemaVersion": 3,
			})
		case "/sync/v3/files/rootHash":
			_, _ = w.Write([]byte(rootEntries))
		case "/sync/v3/files/docHash":
			_, _ = w.Write([]byte(docEntries))
		case "/sync/v3/files/contentHash":
			_, _ = w.Write([]byte(contentJSON))
		case "/sync/v3/files/pageAHash":
			_, _ = w.Write(page1Content)
		case "/sync/v3/files/pageBHash":
			_, _ = w.Write(page2Content)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetSyncHost(server.URL)

	pages, err := client.DownloadPages("fake-device-token", docID)
	require.NoError(t, err)
	require.Len(t, pages, 2)
	assert.Equal(t, "page-a", pages[0].PageID)
	assert.Equal(t, page1Content, pages[0].Data)
	assert.Equal(t, "page-b", pages[1].PageID)
	assert.Equal(t, page2Content, pages[1].Data)
}
