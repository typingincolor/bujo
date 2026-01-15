package ai

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type GenAIClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string, callback func(token string)) error
}

type GeminiGenerator struct {
	client GenAIClient
	loader *PromptLoader
}

func NewGeminiGenerator(client GenAIClient) *GeminiGenerator {
	return &GeminiGenerator{
		client: client,
		loader: NewPromptLoader(""),
	}
}

func NewGeminiGeneratorWithLoader(client GenAIClient, loader *PromptLoader) *GeminiGenerator {
	return &GeminiGenerator{
		client: client,
		loader: loader,
	}
}

func (g *GeminiGenerator) GenerateSummary(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon) (string, error) {
	prompt, err := g.buildPrompt(ctx, entries, horizon)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}
	return g.client.Generate(ctx, prompt)
}

func (g *GeminiGenerator) GenerateSummaryStream(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon, callback func(token string)) error {
	prompt, err := g.buildPrompt(ctx, entries, horizon)
	if err != nil {
		return fmt.Errorf("failed to build prompt: %w", err)
	}
	return g.client.GenerateStream(ctx, prompt, callback)
}

func (g *GeminiGenerator) buildPrompt(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon) (string, error) {
	promptType := domain.PromptTypeFromHorizon(horizon)
	tmpl, err := g.loader.Load(ctx, promptType)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt template: %w", err)
	}

	t, err := template.New("prompt").Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	startDate := time.Now()
	endDate := time.Now()
	if len(entries) > 0 {
		startDate = entries[0].CreatedAt
		endDate = entries[len(entries)-1].CreatedAt
	}

	data := map[string]interface{}{
		"Entries":   entries,
		"Horizon":   horizon,
		"StartDate": startDate,
		"EndDate":   endDate,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
