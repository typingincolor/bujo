package local

import (
	"context"
	"errors"
	"strings"
)

type MockLLMClient struct {
	responses      map[string]string
	generateCalled int
	shouldError    bool
}

func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		responses: make(map[string]string),
	}
}

func (m *MockLLMClient) AddResponse(keyword string, response string) {
	m.responses[keyword] = response
}

func (m *MockLLMClient) Generate(ctx context.Context, prompt string) (string, error) {
	m.generateCalled++

	if m.shouldError {
		return "", errors.New("mock error")
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	for keyword, response := range m.responses {
		if strings.Contains(prompt, keyword) {
			return response, nil
		}
	}

	return "Mock AI response", nil
}

func (m *MockLLMClient) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockLLMClient) CallCount() int {
	return m.generateCalled
}
