package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

func TestModelManifest_Save_And_Load(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewModelManifest(manifestPath)

	spec := domain.ModelSpec{Name: "llama3.2", Variant: "3b"}
	version := domain.ModelVersion{Major: 1, Minor: 0, Patch: 0}

	record := ModelRecord{
		Spec:         spec,
		Version:      version,
		File:         "llama3.2-3b-q4.gguf",
		Size:         2147483648,
		SHA256:       "abc123def456",
		DownloadedAt: time.Now(),
		LastUsed:     time.Now(),
		HFRepo:       "TheBloke/Llama-3.2-3B-GGUF",
		HFCommit:     "commit123",
	}

	if err := manifest.AddModel(context.Background(), record); err != nil {
		t.Fatalf("AddModel() error = %v", err)
	}

	loaded, err := NewModelManifest(manifestPath).Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Load() returned %d models, want 1", len(loaded))
	}

	got := loaded[0]
	if got.Spec.String() != spec.String() {
		t.Errorf("Load() spec = %v, want %v", got.Spec, spec)
	}

	if got.File != record.File {
		t.Errorf("Load() file = %v, want %v", got.File, record.File)
	}

	if got.SHA256 != record.SHA256 {
		t.Errorf("Load() SHA256 = %v, want %v", got.SHA256, record.SHA256)
	}
}

func TestModelManifest_GetModel(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewModelManifest(manifestPath)

	spec := domain.ModelSpec{Name: "tinyllama", Variant: ""}
	record := ModelRecord{
		Spec:    spec,
		Version: domain.ModelVersion{Major: 1, Minor: 0, Patch: 0},
		File:    "tinyllama.gguf",
		Size:    637000000,
	}

	if err := manifest.AddModel(context.Background(), record); err != nil {
		t.Fatalf("AddModel() error = %v", err)
	}

	got, err := manifest.GetModel(context.Background(), spec)
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	if got.Spec.String() != spec.String() {
		t.Errorf("GetModel() spec = %v, want %v", got.Spec, spec)
	}

	if got.File != record.File {
		t.Errorf("GetModel() file = %v, want %v", got.File, record.File)
	}
}

func TestModelManifest_GetModel_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewModelManifest(manifestPath)

	spec := domain.ModelSpec{Name: "nonexistent", Variant: ""}
	_, err := manifest.GetModel(context.Background(), spec)
	if err == nil {
		t.Error("GetModel() expected error for nonexistent model, got nil")
	}
}

func TestModelManifest_RemoveModel(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewModelManifest(manifestPath)

	spec := domain.ModelSpec{Name: "llama3.2", Variant: "3b"}
	record := ModelRecord{
		Spec:    spec,
		Version: domain.ModelVersion{Major: 1, Minor: 0, Patch: 0},
		File:    "llama3.2-3b.gguf",
		Size:    2000000000,
	}

	if err := manifest.AddModel(context.Background(), record); err != nil {
		t.Fatalf("AddModel() error = %v", err)
	}

	if err := manifest.RemoveModel(context.Background(), spec); err != nil {
		t.Fatalf("RemoveModel() error = %v", err)
	}

	_, err := manifest.GetModel(context.Background(), spec)
	if err == nil {
		t.Error("GetModel() expected error after removal, got nil")
	}
}

func TestModelManifest_UpdateLastUsed(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewModelManifest(manifestPath)

	spec := domain.ModelSpec{Name: "tinyllama", Variant: ""}
	originalTime := time.Now().Add(-24 * time.Hour)
	record := ModelRecord{
		Spec:     spec,
		Version:  domain.ModelVersion{Major: 1, Minor: 0, Patch: 0},
		File:     "tinyllama.gguf",
		Size:     637000000,
		LastUsed: originalTime,
	}

	if err := manifest.AddModel(context.Background(), record); err != nil {
		t.Fatalf("AddModel() error = %v", err)
	}

	newTime := time.Now()
	if err := manifest.UpdateLastUsed(context.Background(), spec, newTime); err != nil {
		t.Fatalf("UpdateLastUsed() error = %v", err)
	}

	got, err := manifest.GetModel(context.Background(), spec)
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	if got.LastUsed.Before(originalTime) {
		t.Error("UpdateLastUsed() did not update timestamp")
	}
}

func TestModelManifest_EmptyManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewModelManifest(manifestPath)
	models, err := manifest.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(models) != 0 {
		t.Errorf("Load() on empty manifest returned %d models, want 0", len(models))
	}
}

func TestModelManifest_MultipleModels(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	manifest := NewModelManifest(manifestPath)

	models := []ModelRecord{
		{
			Spec:    domain.ModelSpec{Name: "tinyllama", Variant: ""},
			Version: domain.ModelVersion{Major: 1, Minor: 0, Patch: 0},
			File:    "tinyllama.gguf",
			Size:    637000000,
		},
		{
			Spec:    domain.ModelSpec{Name: "llama3.2", Variant: "3b"},
			Version: domain.ModelVersion{Major: 1, Minor: 0, Patch: 0},
			File:    "llama3.2-3b.gguf",
			Size:    2000000000,
		},
		{
			Spec:    domain.ModelSpec{Name: "mistral", Variant: "7b"},
			Version: domain.ModelVersion{Major: 1, Minor: 0, Patch: 0},
			File:    "mistral-7b.gguf",
			Size:    4100000000,
		},
	}

	for _, record := range models {
		if err := manifest.AddModel(context.Background(), record); err != nil {
			t.Fatalf("AddModel() error = %v", err)
		}
	}

	loaded, err := NewModelManifest(manifestPath).Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != len(models) {
		t.Errorf("Load() returned %d models, want %d", len(loaded), len(models))
	}
}

func TestModelManifest_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	if err := os.WriteFile(manifestPath, []byte("invalid json{{{"), 0644); err != nil {
		t.Fatalf("Failed to create corrupt file: %v", err)
	}

	manifest := NewModelManifest(manifestPath)
	_, err := manifest.Load(context.Background())
	if err == nil {
		t.Error("Load() expected error for corrupt JSON, got nil")
	}
}
