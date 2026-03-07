# reMarkable Cloud Integration Test Harness — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a CLI test harness that proves bujo can authenticate with the reMarkable cloud API, list documents, download ZIPs, extract typed text, and parse entries via TreeParser — all printing to stdout with no DB writes.

**Architecture:** New `internal/adapter/remarkable/` package with HTTP client (auth, list, download) and document processor (ZIP extraction, text parsing). Three Cobra subcommands under `bujo remarkable` (register, list, import). Test harness bypasses `rootCmd.PersistentPreRunE` since no DB is needed.

**Tech Stack:** Go 1.24, net/http, archive/zip, encoding/json, Cobra CLI, testify

---

### Task 1: Token Storage — Config File Read/Write

**Files:**
- Create: `internal/adapter/remarkable/config.go`
- Create: `internal/adapter/remarkable/config_test.go`

**Step 1: Write the failing test for saving config**

```go
// internal/adapter/remarkable/config_test.go
package remarkable

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "remarkable.json")

	cfg := Config{
		DeviceToken: "test-device-token",
		DeviceID:    "test-device-id",
	}

	err := SaveConfig(path, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test-device-token")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/remarkable/... -run TestSaveConfig -v`
Expected: FAIL — package/types don't exist

**Step 3: Write minimal implementation**

```go
// internal/adapter/remarkable/config.go
package remarkable

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DeviceToken string `json:"device_token"`
	DeviceID    string `json:"device_id"`
}

func SaveConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/remarkable/... -run TestSaveConfig -v`
Expected: PASS

**Step 5: Write failing test for loading config**

```go
func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "remarkable.json")

	original := Config{
		DeviceToken: "saved-token",
		DeviceID:    "saved-id",
	}
	err := SaveConfig(path, original)
	require.NoError(t, err)

	loaded, err := LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/remarkable.json")
	assert.Error(t, err)
}
```

**Step 6: Run test to verify it fails**

Run: `go test ./internal/adapter/remarkable/... -run TestLoadConfig -v`
Expected: FAIL — `LoadConfig` undefined

**Step 7: Implement LoadConfig**

```go
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
```

**Step 8: Run test to verify it passes**

Run: `go test ./internal/adapter/remarkable/... -run TestLoadConfig -v`
Expected: PASS

**Step 9: Add DefaultConfigPath helper**

```go
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "bujo", "remarkable.json")
}
```

**Step 10: Commit**

```bash
git add internal/adapter/remarkable/
git commit -m "feat(remarkable): add config file read/write for token storage"
```

---

### Task 2: Device Registration — HTTP Client

**Files:**
- Create: `internal/adapter/remarkable/client.go`
- Create: `internal/adapter/remarkable/client_test.go`

**Step 1: Write the failing test for device registration**

```go
// internal/adapter/remarkable/client_test.go
package remarkable

import (
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/remarkable/... -run TestRegisterDevice -v`
Expected: FAIL — `NewClient`, `RegisterDevice` undefined

**Step 3: Implement Client and RegisterDevice**

```go
// internal/adapter/remarkable/client.go
package remarkable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

const (
	DefaultAuthHost = "https://webapp-prod.cloud.remarkable.engineering"
	DefaultSyncHost = "https://eu.tectonic.remarkable.com"
)

type Client struct {
	authHost   string
	syncHost   string
	httpClient *http.Client
}

func NewClient(authHost string) *Client {
	return &Client{
		authHost:   authHost,
		syncHost:   DefaultSyncHost,
		httpClient: &http.Client{},
	}
}

func (c *Client) RegisterDevice(code string) (string, error) {
	body := map[string]string{
		"code":       code,
		"deviceDesc": "desktop-macos",
		"deviceID":   uuid.New().String(),
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(
		c.authHost+"/token/json/2/device/new",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed: status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/remarkable/... -run TestRegisterDevice -v`
Expected: PASS

**Step 5: Write failing test for registration failure**

```go
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
```

**Step 6: Run test to verify it fails then passes**

Run: `go test ./internal/adapter/remarkable/... -run TestRegisterDevice -v`
Expected: PASS (error handling already in implementation)

**Step 7: Commit**

```bash
git add internal/adapter/remarkable/
git commit -m "feat(remarkable): add device registration HTTP client"
```

---

### Task 3: User Token Refresh

**Files:**
- Modify: `internal/adapter/remarkable/client.go`
- Modify: `internal/adapter/remarkable/client_test.go`

**Step 1: Write failing test for user token refresh**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/remarkable/... -run TestRefreshUserToken -v`
Expected: FAIL — `RefreshUserToken` undefined

**Step 3: Implement RefreshUserToken**

```go
func (c *Client) RefreshUserToken(deviceToken string) (string, error) {
	req, err := http.NewRequest("POST", c.authHost+"/token/json/2/user/new", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh failed: status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/remarkable/... -run TestRefreshUserToken -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/
git commit -m "feat(remarkable): add user token refresh"
```

---

### Task 4: Document Listing

**Files:**
- Create: `internal/adapter/remarkable/document.go`
- Modify: `internal/adapter/remarkable/client.go`
- Modify: `internal/adapter/remarkable/client_test.go`

**Step 1: Write the document type**

```go
// internal/adapter/remarkable/document.go
package remarkable

type Document struct {
	ID           string `json:"ID"`
	Version      int    `json:"Version"`
	VisibleName  string `json:"VissibleName"` // typo is in the API
	Type         string `json:"Type"`
	Parent       string `json:"Parent"`
	ModifiedAt   string `json:"ModifiedClient"`
	BlobURLGet   string `json:"BlobURLGet"`
}
```

**Step 2: Write failing test for listing documents**

```go
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
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/adapter/remarkable/... -run TestListDocuments -v`
Expected: FAIL — `ListDocuments` undefined

**Step 4: Implement ListDocuments**

```go
func (c *Client) ListDocuments(deviceToken string) ([]Document, error) {
	userToken, err := c.RefreshUserToken(deviceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	req, err := http.NewRequest("GET", c.syncHost+"/doc/v2/files", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list documents failed: status %d", resp.StatusCode)
	}

	var docs []Document
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, err
	}
	return docs, nil
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/adapter/remarkable/... -run TestListDocuments -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/adapter/remarkable/
git commit -m "feat(remarkable): add document listing from cloud API"
```

---

### Task 5: Document Download

**Files:**
- Modify: `internal/adapter/remarkable/client.go`
- Modify: `internal/adapter/remarkable/client_test.go`

**Step 1: Write failing test for downloading a document**

```go
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

	// Main API server returns doc with blob URL when queried by ID
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/remarkable/... -run TestDownloadDocument -v`
Expected: FAIL — `DownloadDocument` undefined

**Step 3: Implement DownloadDocument**

```go
func (c *Client) DownloadDocument(deviceToken string, docID string) ([]byte, error) {
	userToken, err := c.RefreshUserToken(deviceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	req, err := http.NewRequest("GET", c.syncHost+"/doc/v2/files", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)
	q := req.URL.Query()
	q.Set("doc", docID)
	q.Set("withBlob", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch document failed: status %d", resp.StatusCode)
	}

	var docs []Document
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("document %s not found", docID)
	}
	if docs[0].BlobURLGet == "" {
		return nil, fmt.Errorf("no blob URL for document %s", docID)
	}

	blobResp, err := c.httpClient.Get(docs[0].BlobURLGet)
	if err != nil {
		return nil, err
	}
	defer blobResp.Body.Close()

	return io.ReadAll(blobResp.Body)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/remarkable/... -run TestDownloadDocument -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/
git commit -m "feat(remarkable): add document download with blob URL fetch"
```

---

### Task 6: ZIP Extraction — Typed Text Content

**Files:**
- Modify: `internal/adapter/remarkable/document.go`
- Create: `internal/adapter/remarkable/document_test.go`
- Create: `internal/adapter/remarkable/testdata/` (test fixture directory)

**Step 1: Write failing test for text extraction from ZIP**

We don't yet know the exact format of "Convert to text" content inside reMarkable ZIPs. The test harness is designed to discover this. For now, we'll search for common text file extensions (`.txt`, `.json` with text content) inside the ZIP.

```go
// internal/adapter/remarkable/document_test.go
package remarkable

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return buf.Bytes()
}

func TestExtractTextFromZIP(t *testing.T) {
	zipData := createTestZIP(t, map[string]string{
		"doc-id.content":  `{"fileType": "notebook"}`,
		"doc-id/0.rm":     "binary-stroke-data",
		"doc-id/0-metadata.json": `{"layers": [{"name": "Layer 1"}]}`,
		"doc-id/0.txt":    ". Buy groceries\n- Remember to call dentist",
	})

	texts, err := ExtractTextFromZIP(zipData)
	require.NoError(t, err)
	require.Len(t, texts, 1)
	assert.Contains(t, texts[0], "Buy groceries")
}

func TestExtractTextFromZIPNoTextFiles(t *testing.T) {
	zipData := createTestZIP(t, map[string]string{
		"doc-id.content": `{"fileType": "notebook"}`,
		"doc-id/0.rm":    "binary-stroke-data",
	})

	texts, err := ExtractTextFromZIP(zipData)
	require.NoError(t, err)
	assert.Empty(t, texts)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/remarkable/... -run TestExtractText -v`
Expected: FAIL — `ExtractTextFromZIP` undefined

**Step 3: Implement ExtractTextFromZIP**

This is deliberately broad — it looks for any `.txt` files and also dumps the ZIP manifest so the test harness user can inspect what's actually in a real reMarkable document bundle.

```go
// Add to internal/adapter/remarkable/document.go

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func ExtractTextFromZIP(data []byte) ([]string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}

	var texts []string
	for _, f := range r.File {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".txt" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open %s: %w", f.Name, err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", f.Name, err)
			}
			texts = append(texts, string(content))
		}
	}
	return texts, nil
}

func ListZIPContents(data []byte) ([]string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}
	var names []string
	for _, f := range r.File {
		names = append(names, fmt.Sprintf("%s (%d bytes)", f.Name, f.UncompressedSize64))
	}
	return names, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/remarkable/... -run TestExtractText -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/remarkable/
git commit -m "feat(remarkable): add ZIP text extraction and manifest listing"
```

---

### Task 7: Cobra Commands — `remarkable` Parent + `register`

**Files:**
- Create: `cmd/bujo/cmd/remarkable.go`

**Step 1: Create the remarkable command group and register subcommand**

This is a CLI adapter — the logic is tested through the client tests. The remarkable command group bypasses `rootCmd.PersistentPreRunE` since it doesn't need a database.

```go
// cmd/bujo/cmd/remarkable.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/remarkable"
)

var remarkableCmd = &cobra.Command{
	Use:   "remarkable",
	Short: "reMarkable cloud integration (test harness)",
	Long:  `Commands for testing reMarkable cloud API integration. No database required.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Bypass rootCmd.PersistentPreRunE — no DB needed
		return nil
	},
}

var remarkableRegisterCmd = &cobra.Command{
	Use:   "register <code>",
	Short: "Register device with reMarkable cloud using one-time code",
	Long: `Register this device with the reMarkable cloud API.

Get a code from: my.remarkable.com/connect/desktop
Then run: bujo remarkable register <8-char-code>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code := args[0]
		client := remarkable.NewClient(remarkable.DefaultAuthHost)

		fmt.Println("Registering with reMarkable cloud...")
		deviceToken, err := client.RegisterDevice(code)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}

		configPath := remarkable.DefaultConfigPath()
		cfg := remarkable.Config{
			DeviceToken: deviceToken,
			DeviceID:    "", // UUID was generated inside RegisterDevice
		}
		if err := remarkable.SaveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Device registered. Token saved to %s\n", configPath)
		return nil
	},
}

func init() {
	remarkableCmd.AddCommand(remarkableRegisterCmd)
	rootCmd.AddCommand(remarkableCmd)
}
```

**Step 2: Verify it compiles**

Run: `go build ./cmd/bujo/...`
Expected: Success

**Step 3: Verify help output**

Run: `go run ./cmd/bujo remarkable --help`
Expected: Shows `register` subcommand in help

**Step 4: Commit**

```bash
git add cmd/bujo/cmd/remarkable.go
git commit -m "feat(remarkable): add CLI commands — remarkable register"
```

---

### Task 8: Cobra Commands — `list` and `import`

**Files:**
- Modify: `cmd/bujo/cmd/remarkable.go`

**Step 1: Add the list subcommand**

```go
var remarkableListCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents from reMarkable cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := remarkable.DefaultConfigPath()
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		docs, err := client.ListDocuments(cfg.DeviceToken)
		if err != nil {
			return fmt.Errorf("failed to list documents: %w", err)
		}

		if len(docs) == 0 {
			fmt.Println("No documents found.")
			return nil
		}

		fmt.Printf("%-40s %-20s %s\n", "NAME", "MODIFIED", "ID")
		for _, doc := range docs {
			if doc.Type != "DocumentType" {
				continue
			}
			fmt.Printf("%-40s %-20s %s\n", doc.VisibleName, doc.ModifiedAt, doc.ID)
		}
		return nil
	},
}
```

**Step 2: Add the import subcommand**

```go
var remarkableImportCmd = &cobra.Command{
	Use:   "import <doc-id>",
	Short: "Download document, extract text, parse entries, print to stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]

		configPath := remarkable.DefaultConfigPath()
		cfg, err := remarkable.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("not registered — run 'bujo remarkable register <code>' first: %w", err)
		}

		client := remarkable.NewClient(remarkable.DefaultAuthHost)
		client.SetSyncHost(remarkable.DefaultSyncHost)

		fmt.Printf("Downloading document %s...\n", docID)
		data, err := client.DownloadDocument(cfg.DeviceToken, docID)
		if err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		fmt.Printf("Downloaded %d bytes\n", len(data))

		// Show ZIP manifest for debugging
		manifest, err := remarkable.ListZIPContents(data)
		if err != nil {
			return fmt.Errorf("failed to read ZIP: %w", err)
		}
		fmt.Println("\nZIP contents:")
		for _, entry := range manifest {
			fmt.Printf("  %s\n", entry)
		}

		// Extract text
		texts, err := remarkable.ExtractTextFromZIP(data)
		if err != nil {
			return fmt.Errorf("failed to extract text: %w", err)
		}

		if len(texts) == 0 {
			fmt.Println("\nNo text files found in ZIP. The document may not have been converted to text.")
			fmt.Println("On your reMarkable, select the page → Convert to text, then sync.")
			return nil
		}

		// Parse each text through TreeParser and print
		parser := domain.NewTreeParser()
		for i, text := range texts {
			fmt.Printf("\n--- Page %d ---\n", i+1)
			fmt.Printf("Raw text:\n%s\n", text)

			entries, err := parser.Parse(text)
			if err != nil {
				fmt.Printf("Parse error: %v\n", err)
				continue
			}

			fmt.Printf("\nParsed %d entries:\n", len(entries))
			for _, e := range entries {
				indent := strings.Repeat("  ", e.Depth)
				fmt.Printf("%s%s %s", indent, e.Type, e.Content)
				if e.Priority != domain.PriorityNone {
					fmt.Printf(" [%s]", e.Priority)
				}
				if len(e.Tags) > 0 {
					fmt.Printf(" tags:%v", e.Tags)
				}
				fmt.Println()
			}
		}
		return nil
	},
}
```

**Step 3: Add SetSyncHost method to Client and register commands in init()**

Add to `client.go`:
```go
func (c *Client) SetSyncHost(host string) {
	c.syncHost = host
}
```

Update `init()` in `remarkable.go`:
```go
func init() {
	remarkableCmd.AddCommand(remarkableRegisterCmd)
	remarkableCmd.AddCommand(remarkableListCmd)
	remarkableCmd.AddCommand(remarkableImportCmd)
	rootCmd.AddCommand(remarkableCmd)
}
```

Add required imports to `remarkable.go`:
```go
import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/typingincolor/bujo/internal/adapter/remarkable"
	"github.com/typingincolor/bujo/internal/domain"
)
```

**Step 4: Verify it compiles**

Run: `go build ./cmd/bujo/...`
Expected: Success

**Step 5: Verify help output shows all subcommands**

Run: `go run ./cmd/bujo remarkable --help`
Expected: Shows `register`, `list`, `import` subcommands

**Step 6: Commit**

```bash
git add cmd/bujo/cmd/remarkable.go internal/adapter/remarkable/client.go
git commit -m "feat(remarkable): add list and import CLI commands"
```

---

### Task 9: Run All Tests + Final Verification

**Step 1: Run the full test suite**

Run: `go test ./... -v`
Expected: All tests pass, including new remarkable package tests

**Step 2: Run vet**

Run: `go vet ./...`
Expected: No issues

**Step 3: Build the binary**

Run: `go build -o bujo ./cmd/bujo`
Expected: Success

**Step 4: Verify remarkable subcommand**

Run: `./bujo remarkable --help`
Expected: Help text showing register, list, import

**Step 5: Final commit (if any cleanup needed)**

```bash
git add -A
git commit -m "feat(remarkable): test harness complete — register, list, import commands"
```

---

## Manual Testing Guide

After completing the implementation, test with a real reMarkable account:

```bash
# 1. Get a code from my.remarkable.com/connect/desktop
# 2. Register
./bujo remarkable register <code>

# 3. List documents
./bujo remarkable list

# 4. Import a document (use ID from list output)
./bujo remarkable import <doc-id>
```

The import command will show the ZIP manifest — this is how we discover the actual format of "Convert to text" content. Based on findings, we may need to adjust `ExtractTextFromZIP` to look for different file patterns.
