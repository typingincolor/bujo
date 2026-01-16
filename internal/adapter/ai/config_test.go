package ai

import (
	"context"
	"os"
	"testing"
)

func TestNewAIClient_Gemini(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("GEMINI_API_KEY", originalKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Setenv("BUJO_AI_ENABLED", "true")
	_ = os.Setenv("GEMINI_API_KEY", "test-key")
	_ = os.Setenv("BUJO_AI_PROVIDER", "gemini")

	ctx := context.Background()
	client, err := NewAIClient(ctx)
	if err != nil {
		t.Fatalf("NewAIClient() unexpected error: %v", err)
	}

	if client == nil {
		t.Error("NewAIClient() returned nil client")
	}
}

func TestNewAIClient_Local(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	originalModel := os.Getenv("BUJO_MODEL")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
		_ = os.Setenv("BUJO_MODEL", originalModel)
	}()

	_ = os.Setenv("BUJO_AI_ENABLED", "true")
	_ = os.Setenv("BUJO_AI_PROVIDER", "local")
	_ = os.Setenv("BUJO_MODEL", "llama3.2:1b")

	ctx := context.Background()
	client, err := NewAIClient(ctx)

	// With Ollama, client creation succeeds even if model isn't pulled yet
	// Errors occur at generation time, not at client creation
	if err != nil {
		t.Logf("NewAIClient() returned error (Ollama may not be running): %v", err)
		return
	}

	if client == nil {
		t.Error("NewAIClient() returned nil client")
	}
}

func TestNewAIClient_DefaultToLocal(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	originalModel := os.Getenv("BUJO_MODEL")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("GEMINI_API_KEY", originalKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
		_ = os.Setenv("BUJO_MODEL", originalModel)
	}()

	_ = os.Setenv("BUJO_AI_ENABLED", "true")
	_ = os.Unsetenv("GEMINI_API_KEY")
	_ = os.Unsetenv("BUJO_AI_PROVIDER")
	_ = os.Unsetenv("BUJO_MODEL")

	ctx := context.Background()
	client, err := NewAIClient(ctx)

	// With Ollama, defaults to local with llama3.2:1b
	// Client creation succeeds, errors occur at generation if Ollama not running
	if err != nil {
		t.Logf("NewAIClient() returned error (Ollama may not be running): %v", err)
		return
	}

	if client == nil {
		t.Error("NewAIClient() returned nil client")
	}
}

func TestNewAIClient_UnknownProvider(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Setenv("BUJO_AI_ENABLED", "true")
	_ = os.Setenv("BUJO_AI_PROVIDER", "unknown")

	ctx := context.Background()
	_, err := NewAIClient(ctx)

	if err == nil {
		t.Error("NewAIClient() expected error for unknown provider")
	}
}

func TestNewAIClient_GeminiWithoutKey(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("GEMINI_API_KEY", originalKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Setenv("BUJO_AI_ENABLED", "true")
	_ = os.Unsetenv("GEMINI_API_KEY")
	_ = os.Setenv("BUJO_AI_PROVIDER", "gemini")

	ctx := context.Background()
	_, err := NewAIClient(ctx)

	if err == nil {
		t.Error("NewAIClient() expected error when gemini provider set without API key")
	}
}

func TestNewAIClient_DisabledByDefault(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("GEMINI_API_KEY", originalKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Unsetenv("BUJO_AI_ENABLED")
	_ = os.Setenv("GEMINI_API_KEY", "test-key")
	_ = os.Setenv("BUJO_AI_PROVIDER", "gemini")

	ctx := context.Background()
	_, err := NewAIClient(ctx)

	if err != ErrAIDisabled {
		t.Errorf("NewAIClient() expected ErrAIDisabled when BUJO_AI_ENABLED not set, got: %v", err)
	}
}

func TestNewAIClient_ExplicitlyEnabled(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("GEMINI_API_KEY", originalKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Setenv("BUJO_AI_ENABLED", "true")
	_ = os.Setenv("GEMINI_API_KEY", "test-key")
	_ = os.Setenv("BUJO_AI_PROVIDER", "gemini")

	ctx := context.Background()
	client, err := NewAIClient(ctx)

	if err != nil {
		t.Fatalf("NewAIClient() unexpected error: %v", err)
	}
	if client == nil {
		t.Error("NewAIClient() returned nil client")
	}
}

func TestNewAIClient_ExplicitlyDisabled(t *testing.T) {
	originalEnabled := os.Getenv("BUJO_AI_ENABLED")
	originalKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("BUJO_AI_ENABLED", originalEnabled)
		_ = os.Setenv("GEMINI_API_KEY", originalKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Setenv("BUJO_AI_ENABLED", "false")
	_ = os.Setenv("GEMINI_API_KEY", "test-key")
	_ = os.Setenv("BUJO_AI_PROVIDER", "gemini")

	ctx := context.Background()
	_, err := NewAIClient(ctx)

	if err != ErrAIDisabled {
		t.Errorf("NewAIClient() expected ErrAIDisabled when BUJO_AI_ENABLED=false, got: %v", err)
	}
}
