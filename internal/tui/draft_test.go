package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDraftPath_ReturnsCorrectLocation(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(home, ".bujo", "capture_draft.txt")
	actual := DraftPath()

	if actual != expected {
		t.Errorf("expected draft path %s, got %s", expected, actual)
	}
}

func TestLoadDraft_ReturnsEmptyWhenNoFile(t *testing.T) {
	// Use a temp directory to ensure no draft exists
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "nonexistent_draft.txt")

	content, exists := LoadDraft(path)

	if exists {
		t.Error("expected exists to be false for nonexistent file")
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestLoadDraft_ReturnsContentWhenFileExists(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	// Create draft file
	draftContent := ". Task one\n- Note here"
	if err := os.WriteFile(path, []byte(draftContent), 0644); err != nil {
		t.Fatal(err)
	}

	content, exists := LoadDraft(path)

	if !exists {
		t.Error("expected exists to be true for existing file")
	}
	if content != draftContent {
		t.Errorf("expected content %q, got %q", draftContent, content)
	}
}

func TestSaveDraft_CreatesFileWithContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	content := ". My task"
	err := SaveDraft(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created with correct content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read draft file: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected file content %q, got %q", content, string(data))
	}
}

func TestSaveDraft_CreatesDirectoryIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "subdir", "draft.txt")

	content := ". My task"
	err := SaveDraft(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("draft file was not created")
	}
}

func TestSaveDraft_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	// Create initial draft
	if err := os.WriteFile(path, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	newContent := ". New task"
	err := SaveDraft(path, newContent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read draft file: %v", err)
	}
	if string(data) != newContent {
		t.Errorf("expected file content %q, got %q", newContent, string(data))
	}
}

func TestSaveDraft_DoesNotSaveEmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	err := SaveDraft(path, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should not exist for empty content
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("draft file should not be created for empty content")
	}
}

func TestDeleteDraft_RemovesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "draft.txt")

	// Create draft file
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	err := DeleteDraft(path)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("draft file should be deleted")
	}
}

func TestDeleteDraft_NoErrorWhenFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "nonexistent.txt")

	err := DeleteDraft(path)

	if err != nil {
		t.Errorf("expected no error for nonexistent file, got %v", err)
	}
}
