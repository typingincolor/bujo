//go:build !cgo
// +build !cgo

package local

import (
	"context"
	"errors"
)

type LocalClient struct{}

func NewLocalClient(modelPath string) (*LocalClient, error) {
	return nil, errors.New("local AI not available: bujo was built without CGO support. Use BUJO_AI_PROVIDER=gemini or rebuild with CGO_ENABLED=1")
}

func (c *LocalClient) Generate(ctx context.Context, prompt string) (string, error) {
	return "", errors.New("local AI not available")
}

func (c *LocalClient) GenerateStream(ctx context.Context, prompt string, callback func(token string)) error {
	return errors.New("local AI not available")
}

func (c *LocalClient) Close() error {
	return nil
}
