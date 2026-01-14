package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/typingincolor/bujo/internal/domain"
)

type DownloadResult struct {
	Path   string
	Size   int64
	SHA256 string
}

type HTTPFetcher interface {
	Fetch(ctx context.Context, url string, progress func(int64, int64)) (io.ReadCloser, int64, error)
}

type ModelDownloader struct {
	destDir string
	fetcher HTTPFetcher
}

func NewModelDownloader(destDir string, fetcher HTTPFetcher) *ModelDownloader {
	return &ModelDownloader{
		destDir: destDir,
		fetcher: fetcher,
	}
}

func (d *ModelDownloader) Download(ctx context.Context, model domain.ModelInfo, progress func(int64, int64)) (DownloadResult, error) {
	if err := model.Validate(); err != nil {
		return DownloadResult{}, fmt.Errorf("invalid model: %w", err)
	}

	if model.HFFile == "" {
		return DownloadResult{}, fmt.Errorf("model file name is required")
	}

	if err := os.MkdirAll(d.destDir, 0755); err != nil {
		return DownloadResult{}, fmt.Errorf("failed to create destination directory: %w", err)
	}

	url := fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", model.HFRepo, model.HFFile)

	reader, _, err := d.fetcher.Fetch(ctx, url, progress)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("failed to fetch model: %w", err)
	}
	defer reader.Close()

	destPath := filepath.Join(d.destDir, model.HFFile)
	tempPath := destPath + ".tmp"

	outFile, err := os.Create(tempPath)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	hash := sha256.New()
	writer := io.MultiWriter(outFile, hash)

	written, err := io.Copy(writer, reader)
	if err != nil {
		os.Remove(tempPath)
		return DownloadResult{}, fmt.Errorf("failed to write model data: %w", err)
	}

	if err := outFile.Close(); err != nil {
		os.Remove(tempPath)
		return DownloadResult{}, fmt.Errorf("failed to close output file: %w", err)
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		return DownloadResult{}, fmt.Errorf("failed to finalize download: %w", err)
	}

	sha256sum := hex.EncodeToString(hash.Sum(nil))

	return DownloadResult{
		Path:   destPath,
		Size:   written,
		SHA256: sha256sum,
	}, nil
}
