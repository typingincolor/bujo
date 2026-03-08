package remarkable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	DefaultAuthHost = "https://webapp-prod.cloud.remarkable.engineering"
	DefaultSyncHost = "https://eu.tectonic.remarkable.com"
)

const (
	maxTokenSize = 10 * 1024        // 10KB for JWT tokens
	maxMetaSize  = 1024 * 1024      // 1MB for metadata/entries
	maxFileSize  = 50 * 1024 * 1024 // 50MB for file content (.rm, pdf)
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

func (c *Client) SetSyncHost(host string) {
	c.syncHost = host
}

func (c *Client) RegisterDevice(ctx context.Context, code string) (string, error) {
	body := map[string]string{
		"code":       code,
		"deviceDesc": "browser-chrome",
		"deviceID":   uuid.New().String(),
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.authHost+"/token/json/2/device/new", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed: status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(io.LimitReader(resp.Body, maxTokenSize))
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (c *Client) RefreshUserToken(ctx context.Context, deviceToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.authHost+"/token/json/2/user/new", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh failed: status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(io.LimitReader(resp.Body, maxTokenSize))
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (c *Client) GetRootHash(ctx context.Context, deviceToken string) (string, int, error) {
	userToken, err := c.RefreshUserToken(ctx, deviceToken)
	if err != nil {
		return "", 0, fmt.Errorf("failed to refresh token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.syncHost+"/sync/v4/root", nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("get root hash failed: status %d", resp.StatusCode)
	}

	var root RootHashResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxMetaSize)).Decode(&root); err != nil {
		return "", 0, err
	}
	return root.Hash, root.Generation, nil
}

func (c *Client) GetEntries(ctx context.Context, userToken string, hash string) ([]SyncEntry, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.syncHost+"/sync/v3/files/"+hash, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get entries failed: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMetaSize))
	if err != nil {
		return nil, err
	}

	return ParseEntries(string(body))
}

func ParseEntries(content string) ([]SyncEntry, error) {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("empty entries file")
	}

	var entries []SyncEntry
	for _, line := range lines[1:] {
		parts := strings.SplitN(line, ":", 5)
		if len(parts) < 3 {
			log.Printf("remarkable: skipping malformed entry line: %q", line)
			continue
		}
		entries = append(entries, SyncEntry{
			Hash: parts[0],
			ID:   parts[2],
		})
	}
	return entries, nil
}

func (c *Client) GetFileContent(ctx context.Context, userToken string, hash string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.syncHost+"/sync/v3/files/"+hash, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get file failed: status %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxFileSize))
}

func (c *Client) getRootEntries(ctx context.Context, deviceToken string) (string, []SyncEntry, error) {
	userToken, err := c.RefreshUserToken(ctx, deviceToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.syncHost+"/sync/v4/root", nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var root RootHashResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxMetaSize)).Decode(&root); err != nil {
		return "", nil, err
	}

	entries, err := c.GetEntries(ctx, userToken, root.Hash)
	if err != nil {
		return "", nil, err
	}

	return userToken, entries, nil
}

func (c *Client) ListDocuments(ctx context.Context, deviceToken string) ([]Document, error) {
	userToken, rootEntries, err := c.getRootEntries(ctx, deviceToken)
	if err != nil {
		return nil, err
	}

	var docs []Document
	for _, entry := range rootEntries {
		subEntries, err := c.GetEntries(ctx, userToken, entry.Hash)
		if err != nil {
			log.Printf("remarkable: skipping document %s: failed to get entries: %v", entry.ID, err)
			continue
		}

		doc := Document{
			ID:   entry.ID,
			Hash: entry.Hash,
		}

		for _, sub := range subEntries {
			if strings.HasSuffix(sub.ID, ".metadata") {
				data, err := c.GetFileContent(ctx, userToken, sub.Hash)
				if err != nil {
					log.Printf("remarkable: skipping metadata for %s: %v", entry.ID, err)
					continue
				}
				var meta DocumentMetadata
				if err := json.Unmarshal(data, &meta); err != nil {
					log.Printf("remarkable: skipping metadata for %s: parse error: %v", entry.ID, err)
					continue
				}
				doc.VisibleName = meta.VisibleName
				doc.LastModified = meta.LastModified
				doc.Parent = meta.Parent
			}
			if strings.HasSuffix(sub.ID, ".content") {
				data, err := c.GetFileContent(ctx, userToken, sub.Hash)
				if err != nil {
					log.Printf("remarkable: skipping content for %s: %v", entry.ID, err)
					continue
				}
				var content DocumentContent
				if err := json.Unmarshal(data, &content); err != nil {
					log.Printf("remarkable: skipping content for %s: parse error: %v", entry.ID, err)
					continue
				}
				doc.FileType = content.FileType
			}
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func (c *Client) DownloadDocument(ctx context.Context, deviceToken string, docID string) ([]byte, error) {
	userToken, rootEntries, err := c.getRootEntries(ctx, deviceToken)
	if err != nil {
		return nil, err
	}

	for _, entry := range rootEntries {
		if entry.ID != docID {
			continue
		}

		subEntries, err := c.GetEntries(ctx, userToken, entry.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to get document entries: %w", err)
		}

		for _, sub := range subEntries {
			ext := sub.ID[strings.LastIndex(sub.ID, ".")+1:]
			switch ext {
			case "pdf", "epub", "zip", "rm":
				return c.GetFileContent(ctx, userToken, sub.Hash)
			}
		}

		return nil, fmt.Errorf("no downloadable file found for document %s", docID)
	}

	return nil, fmt.Errorf("document %s not found", docID)
}

func (c *Client) DownloadPages(ctx context.Context, deviceToken string, docID string) ([]PageData, error) {
	userToken, rootEntries, err := c.getRootEntries(ctx, deviceToken)
	if err != nil {
		return nil, err
	}

	for _, entry := range rootEntries {
		if entry.ID != docID {
			continue
		}

		subEntries, err := c.GetEntries(ctx, userToken, entry.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to get document entries: %w", err)
		}

		var pageOrder []string
		for _, sub := range subEntries {
			if strings.HasSuffix(sub.ID, ".content") {
				data, err := c.GetFileContent(ctx, userToken, sub.Hash)
				if err != nil {
					return nil, fmt.Errorf("failed to get content file: %w", err)
				}
				pageOrder, err = ParsePageOrder(data)
				if err != nil {
					return nil, fmt.Errorf("failed to parse page order: %w", err)
				}
				break
			}
		}

		rmHashes := make(map[string]string)
		for _, sub := range subEntries {
			if strings.HasSuffix(sub.ID, ".rm") {
				rmHashes[sub.ID] = sub.Hash
			}
		}

		var pages []PageData
		for _, pageID := range pageOrder {
			rmKey := docID + "/" + pageID + ".rm"
			hash, ok := rmHashes[rmKey]
			if !ok {
				continue
			}
			data, err := c.GetFileContent(ctx, userToken, hash)
			if err != nil {
				return nil, fmt.Errorf("failed to download page %s: %w", pageID, err)
			}
			pages = append(pages, PageData{PageID: pageID, Data: data})
		}

		return pages, nil
	}

	return nil, fmt.Errorf("document %s not found", docID)
}
