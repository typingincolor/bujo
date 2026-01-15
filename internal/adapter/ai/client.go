package ai

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-2.0-flash",
	}, nil
}

func (c *GeminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(result.Candidates) == 0 {
		return "", errors.New("no response generated")
	}

	return result.Text(), nil
}

func (c *GeminiClient) GenerateStream(ctx context.Context, prompt string, callback func(token string)) error {
	iter := c.client.Models.GenerateContentStream(ctx, c.model, genai.Text(prompt), nil)

	for chunk, err := range iter {
		if err != nil {
			return fmt.Errorf("failed to stream content: %w", err)
		}

		if chunk == nil {
			continue
		}

		text := chunk.Text()
		if text != "" {
			callback(text)
		}
	}

	return nil
}
