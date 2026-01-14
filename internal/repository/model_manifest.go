package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type ModelRecord struct {
	Spec         domain.ModelSpec
	Version      domain.ModelVersion
	File         string
	Size         int64
	SHA256       string
	DownloadedAt time.Time
	LastUsed     time.Time
	HFRepo       string
	HFCommit     string
}

type manifestData struct {
	Models map[string]ModelRecord `json:"models"`
}

type ModelManifest struct {
	path string
}

func NewModelManifest(path string) *ModelManifest {
	return &ModelManifest{path: path}
}

func (m *ModelManifest) Load(ctx context.Context) ([]ModelRecord, error) {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		return []ModelRecord{}, nil
	}

	data, err := os.ReadFile(m.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest manifestData
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	records := make([]ModelRecord, 0, len(manifest.Models))
	for _, record := range manifest.Models {
		records = append(records, record)
	}

	return records, nil
}

func (m *ModelManifest) save(ctx context.Context, models map[string]ModelRecord) error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	manifest := manifestData{
		Models: models,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

func (m *ModelManifest) load(ctx context.Context) (map[string]ModelRecord, error) {
	records, err := m.Load(ctx)
	if err != nil {
		return nil, err
	}

	models := make(map[string]ModelRecord)
	for _, record := range records {
		key := record.Spec.String()
		models[key] = record
	}

	return models, nil
}

func (m *ModelManifest) GetModel(ctx context.Context, spec domain.ModelSpec) (ModelRecord, error) {
	models, err := m.load(ctx)
	if err != nil {
		return ModelRecord{}, err
	}

	key := spec.String()
	record, ok := models[key]
	if !ok {
		return ModelRecord{}, fmt.Errorf("model not found: %s", key)
	}

	return record, nil
}

func (m *ModelManifest) AddModel(ctx context.Context, record ModelRecord) error {
	models, err := m.load(ctx)
	if err != nil {
		return err
	}

	key := record.Spec.String()
	models[key] = record

	return m.save(ctx, models)
}

func (m *ModelManifest) RemoveModel(ctx context.Context, spec domain.ModelSpec) error {
	models, err := m.load(ctx)
	if err != nil {
		return err
	}

	key := spec.String()
	if _, ok := models[key]; !ok {
		return fmt.Errorf("model not found: %s", key)
	}

	delete(models, key)

	return m.save(ctx, models)
}

func (m *ModelManifest) UpdateLastUsed(ctx context.Context, spec domain.ModelSpec, lastUsed time.Time) error {
	models, err := m.load(ctx)
	if err != nil {
		return err
	}

	key := spec.String()
	record, ok := models[key]
	if !ok {
		return fmt.Errorf("model not found: %s", key)
	}

	record.LastUsed = lastUsed
	models[key] = record

	return m.save(ctx, models)
}

var ErrModelNotFound = errors.New("model not found")
