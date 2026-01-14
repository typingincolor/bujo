package ai

import (
	"context"
	"os"
	"testing"
)

func TestNewAIClient_Gemini(t *testing.T) {
	originalKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("GEMINI_API_KEY", originalKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

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
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	originalModel := os.Getenv("BUJO_MODEL")
	originalModelDir := os.Getenv("BUJO_MODEL_DIR")
	defer func() {
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
		_ = os.Setenv("BUJO_MODEL", originalModel)
		_ = os.Setenv("BUJO_MODEL_DIR", originalModelDir)
	}()

	_ = os.Setenv("BUJO_AI_PROVIDER", "local")
	_ = os.Setenv("BUJO_MODEL", "tinyllama")

	ctx := context.Background()
	_, err := NewAIClient(ctx)

	if err == nil {
		t.Error("NewAIClient() expected error for non-downloaded model, got nil")
	}

	expectedMsg := "not downloaded"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("NewAIClient() error should mention %q, got: %v", expectedMsg, err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewAIClient_DefaultToLocal(t *testing.T) {
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		os.Setenv("GEMINI_API_KEY", originalGeminiKey)
		os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Unsetenv("GEMINI_API_KEY")
	_ = os.Unsetenv("BUJO_AI_PROVIDER")

	ctx := context.Background()
	_, err := NewAIClient(ctx)

	if err == nil {
		t.Error("NewAIClient() expected error when no model configured, got nil")
	}
}

func TestNewAIClient_GeminiFallback(t *testing.T) {
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	originalProvider := os.Getenv("BUJO_AI_PROVIDER")
	defer func() {
		_ = os.Setenv("GEMINI_API_KEY", originalGeminiKey)
		_ = os.Setenv("BUJO_AI_PROVIDER", originalProvider)
	}()

	_ = os.Setenv("GEMINI_API_KEY", "test-key")
	_ = os.Unsetenv("BUJO_AI_PROVIDER")

	ctx := context.Background()
	client, err := NewAIClient(ctx)

	if err != nil {
		t.Fatalf("NewAIClient() unexpected error: %v", err)
	}

	if client == nil {
		t.Error("NewAIClient() returned nil client for Gemini fallback")
	}
}
