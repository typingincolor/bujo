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

func TestMockLLMClient_GenerateStream(t *testing.T) {
	mock := NewMockLLMClient()
	mock.AddResponse("test", "Hello world from AI")

	ctx := context.Background()
	var tokens []string

	err := mock.GenerateStream(ctx, "test prompt", func(token string) {
		tokens = append(tokens, token)
	})

	if err != nil {
		t.Fatalf("GenerateStream() unexpected error: %v", err)
	}

	expected := []string{"Hello ", "world ", "from ", "AI "}
	if len(tokens) != len(expected) {
		t.Errorf("GenerateStream() got %d tokens, want %d", len(tokens), len(expected))
	}

	for i, token := range tokens {
		if i >= len(expected) {
			break
		}
		if token != expected[i] {
			t.Errorf("GenerateStream() token[%d] = %q, want %q", i, token, expected[i])
		}
	}
}

func TestMockLLMClient_GenerateStream_ContextCancellation(t *testing.T) {
	mock := NewMockLLMClient()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mock.GenerateStream(ctx, "test prompt", func(token string) {})
	if err == nil {
		t.Error("GenerateStream() expected error with canceled context, got nil")
	}
}

func TestMockLLMClient_GenerateStream_Error(t *testing.T) {
	mock := NewMockLLMClient()
	mock.SetError(true)

	err := mock.GenerateStream(context.Background(), "test prompt", func(token string) {})
	if err == nil {
		t.Error("GenerateStream() expected error, got nil")
	}
}
