package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/typingincolor/bujo/internal/adapter/ai/local"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/repository"
)

type ModelService struct {
	modelsDir string
	manifest  *repository.ModelManifest
}

func NewModelService(modelsDir string) *ModelService {
	manifestPath := filepath.Join(modelsDir, "manifest.json")
	return &ModelService{
		modelsDir: modelsDir,
		manifest:  repository.NewModelManifest(manifestPath),
	}
}

func (s *ModelService) List(ctx context.Context) ([]domain.ModelInfo, error) {
	available := domain.AvailableModels()

	downloaded, err := s.manifest.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	downloadedMap := make(map[string]repository.ModelRecord)
	for _, record := range downloaded {
		key := record.Spec.String()
		downloadedMap[key] = record
	}

	for i := range available {
		key := available[i].Spec.String()
		if record, ok := downloadedMap[key]; ok {
			available[i].LocalPath = filepath.Join(s.modelsDir, record.File)
			available[i].LocalVersion = &record.Version
		}
	}

	return available, nil
}

func (s *ModelService) GetDefaultModel(ctx context.Context) (domain.ModelInfo, error) {
	defaultSpec := domain.ModelSpec{Name: "llama3.2", Variant: "1b"}
	return s.FindModel(ctx, defaultSpec)
}

func (s *ModelService) FindModel(ctx context.Context, spec domain.ModelSpec) (domain.ModelInfo, error) {
	models, err := s.List(ctx)
	if err != nil {
		return domain.ModelInfo{}, err
	}

	for _, model := range models {
		if model.Spec.String() == spec.String() {
			return model, nil
		}
	}

	return domain.ModelInfo{}, fmt.Errorf("model not found: %s", spec)
}

func (s *ModelService) Pull(ctx context.Context, spec domain.ModelSpec, progress func(int64, int64)) error {
	model, err := s.FindModel(ctx, spec)
	if err != nil {
		return err
	}

	if model.IsDownloaded() {
		return fmt.Errorf("model %s is already downloaded", spec)
	}

	fetcher := local.NewHTTPClient()
	downloader := local.NewModelDownloader(s.modelsDir, fetcher)

	result, err := downloader.Download(ctx, model, progress)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	now := time.Now()
	record := repository.ModelRecord{
		Spec:         spec,
		Version:      model.Version,
		File:         filepath.Base(result.Path),
		Size:         result.Size,
		SHA256:       result.SHA256,
		DownloadedAt: now,
		LastUsed:     now,
		HFRepo:       model.HFRepo,
		HFCommit:     "",
	}

	if err := s.manifest.AddModel(ctx, record); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

func (s *ModelService) Remove(ctx context.Context, spec domain.ModelSpec) error {
	record, err := s.manifest.GetModel(ctx, spec)
	if err != nil {
		return fmt.Errorf("model not found in manifest: %w", err)
	}

	modelPath := filepath.Join(s.modelsDir, record.File)
	if err := os.Remove(modelPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove model file: %w", err)
	}

	if err := s.manifest.RemoveModel(ctx, spec); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}
