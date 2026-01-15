//go:build cgo
// +build cgo

package local

import (
	"context"
	"errors"
	"fmt"
	"os"

	llama "github.com/go-skynet/go-llama.cpp"
)

type LocalClient struct {
	model     *llama.LLama
	modelPath string
	closed    bool
}

func NewLocalClient(modelPath string) (*LocalClient, error) {
	if modelPath == "" {
		return nil, errors.New("model path cannot be empty")
	}

	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("model file not found: %w", err)
	}

	model, err := llama.New(modelPath, llama.EnableF16Memory, llama.SetContext(2048))
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	return &LocalClient{
		model:     model,
		modelPath: modelPath,
		closed:    false,
	}, nil
}

func (c *LocalClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.closed {
		return "", errors.New("client is closed")
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	result, err := c.model.Predict(
		prompt,
		llama.SetTemperature(0.7),
		llama.SetTopP(0.9),
		llama.SetTokens(512),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return result, nil
}

func (c *LocalClient) GenerateStream(ctx context.Context, prompt string, callback func(token string)) error {
	if c.closed {
		return errors.New("client is closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tokenCallback := func(token string) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			callback(token)
			return true
		}
	}

	_, err := c.model.Predict(
		prompt,
		llama.SetTemperature(0.7),
		llama.SetTopP(0.9),
		llama.SetTokens(512),
		llama.SetTokenCallback(tokenCallback),
	)

	if err != nil {
		return fmt.Errorf("failed to generate streaming response: %w", err)
	}

	return nil
}

func (c *LocalClient) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	c.model.Free()
	return nil
}
