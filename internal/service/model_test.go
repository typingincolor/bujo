package service

import (
	"context"
	"testing"

	"github.com/typingincolor/bujo/internal/domain"
)

func TestModelService_List(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewModelService(tmpDir)

	models, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(models) == 0 {
		t.Fatal("List() returned empty list, expected available models")
	}

	expectedModels := map[string]bool{
		"tinyllama":   true,
		"llama3.2:1b": true,
		"llama3.2:3b": true,
		"phi-3-mini":  true,
		"mistral:7b":  true,
	}

	for _, model := range models {
		modelName := model.Spec.String()
		if !expectedModels[modelName] {
			t.Errorf("List() contains unexpected model: %s", modelName)
		}
		delete(expectedModels, modelName)
	}

	for missing := range expectedModels {
		t.Errorf("List() missing expected model: %s", missing)
	}
}

func TestModelService_List_ShowsDownloadStatus(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewModelService(tmpDir)

	models, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	for _, model := range models {
		if model.IsDownloaded() {
			t.Errorf("List() model %s shows as downloaded in fresh environment", model.Spec)
		}

		if model.LocalPath != "" {
			t.Errorf("List() model %s has non-empty LocalPath without download", model.Spec)
		}

		if model.LocalVersion != nil {
			t.Errorf("List() model %s has LocalVersion without download", model.Spec)
		}
	}
}

func TestModelService_GetDefaultModel(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewModelService(tmpDir)

	model, err := svc.GetDefaultModel(context.Background())
	if err != nil {
		t.Fatalf("GetDefaultModel() error = %v", err)
	}

	expected := domain.ModelSpec{Name: "llama3.2", Variant: "1b"}
	if model.Spec.String() != expected.String() {
		t.Errorf("GetDefaultModel() = %v, want %v", model.Spec, expected)
	}
}

func TestModelService_FindModel(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewModelService(tmpDir)

	spec := domain.ModelSpec{Name: "tinyllama", Variant: ""}
	model, err := svc.FindModel(context.Background(), spec)
	if err != nil {
		t.Fatalf("FindModel() error = %v", err)
	}

	if model.Spec.String() != spec.String() {
		t.Errorf("FindModel() = %v, want %v", model.Spec, spec)
	}
}

func TestModelService_FindModel_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewModelService(tmpDir)

	spec := domain.ModelSpec{Name: "nonexistent", Variant: ""}
	_, err := svc.FindModel(context.Background(), spec)
	if err == nil {
		t.Error("FindModel() expected error for nonexistent model, got nil")
	}
}
