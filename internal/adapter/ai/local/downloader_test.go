package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/typingincolor/bujo/internal/domain"
)

func TestModelDownloader_Download(t *testing.T) {
	tmpDir := t.TempDir()

	mockFetcher := &mockHTTPFetcher{
		content: []byte("fake model data for testing"),
	}

	downloader := NewModelDownloader(tmpDir, mockFetcher)

	model := domain.ModelInfo{
		Spec:    domain.ModelSpec{Name: "tinyllama", Variant: ""},
		Version: domain.ModelVersion{Major: 1, Minor: 0, Patch: 0},
		Size:    27,
		HFRepo:  "test/repo",
		HFFile:  "model.gguf",
	}

	var progressCalls int
	progressCallback := func(downloaded, total int64) {
		progressCalls++
		if downloaded > total {
			t.Errorf("Progress callback: downloaded (%d) > total (%d)", downloaded, total)
		}
	}

	result, err := downloader.Download(context.Background(), model, progressCallback)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	if result.Path == "" {
		t.Error("Download() returned empty path")
	}

	if !filepath.IsAbs(result.Path) {
		t.Errorf("Download() returned relative path: %s", result.Path)
	}

	if result.Size != model.Size {
		t.Errorf("Download() size = %d, want %d", result.Size, model.Size)
	}

	if result.SHA256 == "" {
		t.Error("Download() returned empty SHA256")
	}

	if progressCalls == 0 {
		t.Error("Download() did not call progress callback")
	}

	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "fake model data for testing" {
		t.Errorf("Downloaded content mismatch")
	}
}

func TestModelDownloader_Download_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "nested", "models")

	mockFetcher := &mockHTTPFetcher{
		content: []byte("test"),
	}

	downloader := NewModelDownloader(destDir, mockFetcher)

	model := domain.ModelInfo{
		Spec:   domain.ModelSpec{Name: "test", Variant: ""},
		Size:   4,
		HFFile: "test.gguf",
	}

	result, err := downloader.Download(context.Background(), model, nil)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	if _, err := os.Stat(filepath.Dir(result.Path)); os.IsNotExist(err) {
		t.Error("Download() did not create destination directory")
	}
}

func TestModelDownloader_Download_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	mockFetcher := &mockHTTPFetcher{
		content:      make([]byte, 1000),
		simulateHang: true,
	}

	downloader := NewModelDownloader(tmpDir, mockFetcher)

	model := domain.ModelInfo{
		Spec:   domain.ModelSpec{Name: "test", Variant: ""},
		Size:   1000,
		HFFile: "test.gguf",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := downloader.Download(ctx, model, nil)
	if err == nil {
		t.Error("Download() expected error on cancelled context, got nil")
	}
}

func TestModelDownloader_Download_InvalidModel(t *testing.T) {
	tmpDir := t.TempDir()

	mockFetcher := &mockHTTPFetcher{
		content: []byte("test"),
	}

	downloader := NewModelDownloader(tmpDir, mockFetcher)

	model := domain.ModelInfo{
		Spec:   domain.ModelSpec{Name: "", Variant: ""},
		Size:   4,
		HFFile: "",
	}

	_, err := downloader.Download(context.Background(), model, nil)
	if err == nil {
		t.Error("Download() expected error for invalid model, got nil")
	}
}

type mockHTTPFetcher struct {
	content      []byte
	simulateHang bool
	shouldError  bool
}

func (m *mockHTTPFetcher) Fetch(ctx context.Context, url string, progress func(int64, int64)) (io.ReadCloser, int64, error) {
	if m.shouldError {
		return nil, 0, io.ErrUnexpectedEOF
	}

	if m.simulateHang {
		<-ctx.Done()
		return nil, 0, ctx.Err()
	}

	reader := io.NopCloser(strings.NewReader(string(m.content)))
	size := int64(len(m.content))

	if progress != nil {
		progress(size, size)
	}

	return reader, size, nil
}
