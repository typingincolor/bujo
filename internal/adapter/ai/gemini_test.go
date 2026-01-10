package ai

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typingincolor/bujo/internal/domain"
)

func TestGeminiGenerator_GenerateSummary(t *testing.T) {
	t.Run("formats entries correctly for daily summary", func(t *testing.T) {
		var capturedPrompt string
		mockClient := &mockGenAIClient{
			generateFunc: func(ctx context.Context, prompt string) (string, error) {
				capturedPrompt = prompt
				return "AI generated summary content", nil
			},
		}

		generator := NewGeminiGenerator(mockClient)

		entries := []domain.Entry{
			{ID: 1, Type: domain.EntryTypeTask, Content: "Buy groceries", CreatedAt: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
			{ID: 2, Type: domain.EntryTypeDone, Content: "Finish report", CreatedAt: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
			{ID: 3, Type: domain.EntryTypeNote, Content: "Remember to call mom", CreatedAt: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
		}

		result, err := generator.GenerateSummary(context.Background(), entries, domain.SummaryHorizonDaily)

		require.NoError(t, err)
		assert.Equal(t, "AI generated summary content", result)
		assert.Contains(t, capturedPrompt, "Buy groceries")
		assert.Contains(t, capturedPrompt, "Finish report")
		assert.Contains(t, capturedPrompt, "Remember to call mom")
		assert.Contains(t, capturedPrompt, "daily")
	})

	t.Run("includes horizon type in prompt", func(t *testing.T) {
		var capturedPrompt string
		mockClient := &mockGenAIClient{
			generateFunc: func(ctx context.Context, prompt string) (string, error) {
				capturedPrompt = prompt
				return "Weekly reflection", nil
			},
		}

		generator := NewGeminiGenerator(mockClient)

		entries := []domain.Entry{
			{ID: 1, Type: domain.EntryTypeTask, Content: "Task 1"},
		}

		_, err := generator.GenerateSummary(context.Background(), entries, domain.SummaryHorizonWeekly)

		require.NoError(t, err)
		assert.Contains(t, capturedPrompt, "weekly")
	})

	t.Run("handles empty entries", func(t *testing.T) {
		mockClient := &mockGenAIClient{
			generateFunc: func(ctx context.Context, prompt string) (string, error) {
				return "No entries to summarize", nil
			},
		}

		generator := NewGeminiGenerator(mockClient)

		result, err := generator.GenerateSummary(context.Background(), []domain.Entry{}, domain.SummaryHorizonDaily)

		require.NoError(t, err)
		assert.Equal(t, "No entries to summarize", result)
	})

	t.Run("propagates client errors", func(t *testing.T) {
		mockClient := &mockGenAIClient{
			generateFunc: func(ctx context.Context, prompt string) (string, error) {
				return "", assert.AnError
			},
		}

		generator := NewGeminiGenerator(mockClient)

		_, err := generator.GenerateSummary(context.Background(), []domain.Entry{}, domain.SummaryHorizonDaily)

		require.Error(t, err)
	})
}

type mockGenAIClient struct {
	generateFunc func(ctx context.Context, prompt string) (string, error)
}

func (m *mockGenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	return m.generateFunc(ctx, prompt)
}
