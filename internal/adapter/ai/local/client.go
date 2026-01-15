package local

import (
	"context"
	"errors"
	"fmt"

	"github.com/ollama/ollama/api"
)

type LocalClient struct {
	client    *api.Client
	modelName string
}

func NewLocalClient(modelName string) (*LocalClient, error) {
	if modelName == "" {
		return nil, errors.New("model name cannot be empty")
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w (is Ollama running?)", err)
	}

	return &LocalClient{
		client:    client,
		modelName: modelName,
	}, nil
}

func (c *LocalClient) Generate(ctx context.Context, prompt string) (string, error) {
	var result string
	req := &api.GenerateRequest{
		Model:  c.modelName,
		Prompt: prompt,
		Stream: new(bool),
		Options: map[string]interface{}{
			"temperature": 0.4,
			"top_p":       0.9,
		},
	}

	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		result += resp.Response
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return result, nil
}

func (c *LocalClient) GenerateStream(ctx context.Context, prompt string, callback func(token string)) error {
	req := &api.GenerateRequest{
		Model:  c.modelName,
		Prompt: prompt,
		Options: map[string]interface{}{
			"temperature": 0.4,
			"top_p":       0.9,
		},
	}

	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		callback(resp.Response)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to generate streaming response: %w", err)
	}

	return nil
}

func (c *LocalClient) Close() error {
	return nil
}
