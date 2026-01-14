package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/typingincolor/bujo/internal/adapter/ai/local"
	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

func NewAIClient(ctx context.Context) (GenAIClient, error) {
	provider := os.Getenv("BUJO_AI_PROVIDER")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if provider == "" {
		if geminiKey != "" {
			provider = "gemini"
		} else {
			provider = "local"
		}
	}

	switch provider {
	case "gemini":
		if geminiKey == "" {
			return nil, errors.New("GEMINI_API_KEY is required for gemini provider")
		}
		return NewGeminiClient(ctx, geminiKey)

	case "local":
		return newLocalClient(ctx)

	default:
		return nil, fmt.Errorf("unknown AI provider: %s (expected 'local' or 'gemini')", provider)
	}
}

func newLocalClient(ctx context.Context) (GenAIClient, error) {
	modelSpec := os.Getenv("BUJO_MODEL")
	if modelSpec == "" {
		modelSpec = "llama3.2:1b"
	}

	modelsDir := os.Getenv("BUJO_MODEL_DIR")
	if modelsDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		modelsDir = filepath.Join(home, ".bujo", "models")
	}

	spec, err := domain.ParseModelSpec(modelSpec)
	if err != nil {
		return nil, fmt.Errorf("invalid BUJO_MODEL: %w", err)
	}

	modelService := service.NewModelService(modelsDir)
	model, err := modelService.FindModel(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w. Download with: bujo model pull %s", err, spec)
	}

	if !model.IsDownloaded() {
		return nil, fmt.Errorf("model %s not downloaded. Download with: bujo model pull %s", spec, spec)
	}

	return local.NewLocalClient(model.LocalPath)
}
