package ai

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/typingincolor/bujo/internal/adapter/ai/local"
)

var ErrAIDisabled = errors.New("AI features are disabled (set BUJO_AI_ENABLED=true to enable)")

func NewAIClient(ctx context.Context) (GenAIClient, error) {
	enabled := os.Getenv("BUJO_AI_ENABLED")
	if enabled != "true" {
		return nil, ErrAIDisabled
	}

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
	modelName := os.Getenv("BUJO_MODEL")
	if modelName == "" {
		modelName = "llama3.2:3b"
	}

	return local.NewLocalClient(modelName)
}
