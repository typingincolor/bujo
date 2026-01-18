package ai

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/typingincolor/bujo/internal/domain"
)

//go:embed prompts/*.txt
var defaultPrompts embed.FS

type PromptLoader struct {
	userDir string
}

func NewPromptLoader(userDir string) *PromptLoader {
	return &PromptLoader{
		userDir: userDir,
	}
}

func (l *PromptLoader) Load(ctx context.Context, promptType domain.PromptType) (domain.PromptTemplate, error) {
	if !promptType.IsValid() {
		return domain.PromptTemplate{}, fmt.Errorf("invalid prompt type: %s", promptType)
	}

	filename := promptType.String() + ".txt"

	content, err := l.loadContent(filename)
	if err != nil {
		return domain.PromptTemplate{}, fmt.Errorf("failed to load prompt: %w", err)
	}

	tmpl := domain.PromptTemplate{
		Type:     promptType,
		Content:  content,
		Filename: filename,
	}

	if err := tmpl.Validate(); err != nil {
		return domain.PromptTemplate{}, fmt.Errorf("invalid prompt template: %w", err)
	}

	return tmpl, nil
}

func (l *PromptLoader) loadContent(filename string) (string, error) {
	if l.userDir != "" {
		userPath := filepath.Join(l.userDir, filename)
		if content, err := os.ReadFile(userPath); err == nil {
			return string(content), nil
		}
	}

	embeddedPath := path.Join("prompts", filename)
	content, err := defaultPrompts.ReadFile(embeddedPath)
	if err != nil {
		return "", fmt.Errorf("prompt file not found: %s", filename)
	}

	return string(content), nil
}

func (l *PromptLoader) EnsureDefaults(ctx context.Context) error {
	if l.userDir == "" {
		return nil
	}

	if err := os.MkdirAll(l.userDir, 0755); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	promptTypes := []domain.PromptType{
		domain.PromptTypeSummaryDaily,
		domain.PromptTypeSummaryWeekly,
		domain.PromptTypeAsk,
	}

	for _, pt := range promptTypes {
		filename := pt.String() + ".txt"
		userPath := filepath.Join(l.userDir, filename)

		if _, err := os.Stat(userPath); err == nil {
			continue
		}

		embeddedPath := path.Join("prompts", filename)
		content, err := defaultPrompts.ReadFile(embeddedPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded prompt %s: %w", filename, err)
		}

		if err := os.WriteFile(userPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write prompt %s: %w", filename, err)
		}
	}

	return nil
}
