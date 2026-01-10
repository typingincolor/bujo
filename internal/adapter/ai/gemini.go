package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/typingincolor/bujo/internal/domain"
)

type GenAIClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type GeminiGenerator struct {
	client GenAIClient
}

func NewGeminiGenerator(client GenAIClient) *GeminiGenerator {
	return &GeminiGenerator{client: client}
}

func (g *GeminiGenerator) GenerateSummary(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon) (string, error) {
	prompt := g.buildPrompt(entries, horizon)
	return g.client.Generate(ctx, prompt)
}

func (g *GeminiGenerator) buildPrompt(entries []domain.Entry, horizon domain.SummaryHorizon) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are a helpful assistant analyzing a %s bullet journal summary.\n\n", horizon))

	if len(entries) == 0 {
		sb.WriteString("There are no entries for this period. Provide a brief encouraging message.\n")
		return sb.String()
	}

	sb.WriteString("Here are the journal entries:\n\n")

	for _, entry := range entries {
		symbol := entry.Type.Symbol()
		sb.WriteString(fmt.Sprintf("%s %s\n", symbol, entry.Content))
	}

	sb.WriteString("\n")
	sb.WriteString("Please provide a thoughtful reflection on these entries. ")
	sb.WriteString("Highlight accomplishments, note patterns, and offer constructive insights. ")
	sb.WriteString("Keep the response concise and actionable.")

	return sb.String()
}
