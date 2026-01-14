package ai

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/typingincolor/bujo/internal/domain"
)

func TestPromptLoader_Load_EmbeddedDefaults(t *testing.T) {
	loader := NewPromptLoader("")

	tests := []struct {
		name       string
		promptType domain.PromptType
	}{
		{
			name:       "summary daily",
			promptType: domain.PromptTypeSummaryDaily,
		},
		{
			name:       "summary weekly",
			promptType: domain.PromptTypeSummaryWeekly,
		},
		{
			name:       "summary quarterly",
			promptType: domain.PromptTypeSummaryQuarterly,
		},
		{
			name:       "summary annual",
			promptType: domain.PromptTypeSummaryAnnual,
		},
		{
			name:       "ask",
			promptType: domain.PromptTypeAsk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := loader.Load(context.Background(), tt.promptType)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if tmpl.Type != tt.promptType {
				t.Errorf("Load() type = %v, want %v", tmpl.Type, tt.promptType)
			}

			if tmpl.Content == "" {
				t.Error("Load() returned empty content")
			}

			if err := tmpl.Validate(); err != nil {
				t.Errorf("Load() returned invalid template: %v", err)
			}
		})
	}
}

func TestPromptLoader_Load_UserOverride(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewPromptLoader(tmpDir)

	customContent := "Custom daily prompt: {{.Entries}}"
	customPath := filepath.Join(tmpDir, "summary-daily.txt")
	if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tmpl, err := loader.Load(context.Background(), domain.PromptTypeSummaryDaily)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if tmpl.Content != customContent {
		t.Errorf("Load() did not use user override. got = %v, want = %v", tmpl.Content, customContent)
	}
}

func TestPromptLoader_Load_InvalidType(t *testing.T) {
	loader := NewPromptLoader("")

	_, err := loader.Load(context.Background(), domain.PromptType("invalid"))
	if err == nil {
		t.Error("Load() expected error for invalid type, got nil")
	}
}

func TestPromptLoader_EnsureDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewPromptLoader(tmpDir)

	if err := loader.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("EnsureDefaults() error = %v", err)
	}

	expectedFiles := []string{
		"summary-daily.txt",
		"summary-weekly.txt",
		"summary-quarterly.txt",
		"summary-annual.txt",
		"ask.txt",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("EnsureDefaults() did not create %s", filename)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read created file %s: %v", filename, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("EnsureDefaults() created empty file %s", filename)
		}
	}
}

func TestPromptLoader_EnsureDefaults_PreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewPromptLoader(tmpDir)

	customContent := "My custom prompt that should not be overwritten"
	customPath := filepath.Join(tmpDir, "summary-daily.txt")
	if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := loader.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("EnsureDefaults() error = %v", err)
	}

	content, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != customContent {
		t.Error("EnsureDefaults() overwrote existing custom prompt")
	}
}
