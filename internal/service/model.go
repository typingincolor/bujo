package service

import (
	"context"
	"fmt"
	"path/filepath"

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
	models, err := s.List(ctx)
	if err != nil {
		return domain.ModelInfo{}, err
	}

	defaultSpec := domain.ModelSpec{Name: "llama3.2", Variant: "1b"}

	for _, model := range models {
		if model.Spec.String() == defaultSpec.String() {
			return model, nil
		}
	}

	return domain.ModelInfo{}, fmt.Errorf("default model not found: %s", defaultSpec)
}
