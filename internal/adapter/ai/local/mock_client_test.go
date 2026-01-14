package local

import (
	"context"
	"testing"
)

func TestMockLLMClient_Generate(t *testing.T) {
	mock := NewMockLLMClient()
	mock.AddResponse("habits", "You logged exercise 5 times this week.")

	ctx := context.Background()

	t.Run("matches keyword", func(t *testing.T) {
		response, err := mock.Generate(ctx, "What patterns do you see in my habits?")
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		expected := "You logged exercise 5 times this week."
		if response != expected {
			t.Errorf("Generate() = %q, want %q", response, expected)
		}
	})

	t.Run("no keyword match returns default", func(t *testing.T) {
		response, err := mock.Generate(ctx, "Some other question")
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		expected := "Mock AI response"
		if response != expected {
			t.Errorf("Generate() = %q, want %q", response, expected)
		}
	})

	t.Run("tracks call count", func(t *testing.T) {
		if mock.CallCount() != 2 {
			t.Errorf("CallCount() = %d, want 2", mock.CallCount())
		}
	})
}

func TestMockLLMClient_ContextCancellation(t *testing.T) {
	mock := NewMockLLMClient()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := mock.Generate(ctx, "test prompt")
	if err == nil {
		t.Error("Generate() expected error with canceled context, got nil")
	}
}

func TestMockLLMClient_Error(t *testing.T) {
	mock := NewMockLLMClient()
	mock.SetError(true)

	_, err := mock.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Error("Generate() expected error, got nil")
	}
}
