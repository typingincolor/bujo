package local

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 0,
		},
	}
}

func (h *HTTPClient) Fetch(ctx context.Context, url string, progress func(int64, int64)) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	size := resp.ContentLength
	if size < 0 {
		size = 0
	}

	if progress != nil {
		reader := &progressReader{
			reader:   resp.Body,
			size:     size,
			progress: progress,
		}
		return reader, size, nil
	}

	return resp.Body, size, nil
}

type progressReader struct {
	reader   io.ReadCloser
	size     int64
	read     int64
	progress func(int64, int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.read += int64(n)

	if pr.progress != nil {
		pr.progress(pr.read, pr.size)
	}

	return n, err
}

func (pr *progressReader) Close() error {
	return pr.reader.Close()
}
